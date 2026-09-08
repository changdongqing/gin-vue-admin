package service

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/model"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// Excel 点表导入导出（03 文档 §五）：三 sheet（通道/设备/测点）+ 名称外键 +
// 预览三分类（insert/update/error）+ 事务 upsert + 自动编译部署。
// 注：commit 由前端回传预览数据、服务端重新全量校验后落库（无服务端会话，状态自洽）。

// ImportService 导入导出服务。
type ImportService struct{}

// ImportChannelRow 通道导入行。
type ImportChannelRow struct {
	model.CollectChannel
	MqttServer string `json:"mqttServer"`
	MqttTopic  string `json:"mqttTopic"`
}

// ImportDeviceRow 设备导入行（ChannelName 为名称外键）。
type ImportDeviceRow struct {
	model.CollectDevice
	ChannelName string `json:"channelName"`
}

// ImportVariableRow 测点导入行（ChannelName/DeviceName 为名称外键）。
type ImportVariableRow struct {
	model.CollectVariable
	ChannelName string `json:"channelName"`
	DeviceName  string `json:"deviceName"`
}

// ImportPayload 预览/提交共用数据结构。
type ImportPayload struct {
	Channels  []ImportChannelRow  `json:"channels"`
	Devices   []ImportDeviceRow   `json:"devices"`
	Variables []ImportVariableRow `json:"variables"`
}

// ImportRowError 行级错误。
type ImportRowError struct {
	Sheet  string `json:"sheet"`
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}

// ImportPreviewResult 预览结果（三分类统计由前端按数据计算）。
type ImportPreviewResult struct {
	Data   ImportPayload    `json:"data"`
	Errors []ImportRowError `json:"errors"`
	Stats  map[string]int   `json:"stats"`
}

// ---------- 模板与导出 ----------

const (
	sheetChannel = "通道"
	sheetDevice  = "设备"
	sheetVar     = "测点"
	sheetReadme  = "说明"
)

var channelHeaders = []string{"通道名称*", "接入方式*", "驱动*", "连接配置JSON*", "轮询周期ms", "MQTT发布server", "MQTT发布topic", "使能", "备注"}
var deviceHeaders = []string{"所属通道名称*", "设备名称*", "设备类型*", "设备类型名", "站号", "采集周期ms", "使能", "备注"}
var varHeaders = []string{"所属设备(通道/设备)*", "测点名称*", "地址*", "数据类型*", "系数", "偏移", "字节序", "采集分组", "读写", "单位", "使能", "备注"}

// GenerateTemplate 生成导入模板（三 sheet + 说明页）。
func (ImportService) GenerateTemplate() (*excelize.File, error) {
	f := excelize.NewFile()
	defaultSheet := f.GetSheetName(0)
	_ = f.SetSheetName(defaultSheet, sheetChannel)
	for name, headers := range map[string][]string{
		sheetDevice: deviceHeaders,
		sheetVar:    varHeaders,
	} {
		if _, err := f.NewSheet(name); err != nil {
			return nil, err
		}
		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			_ = f.SetCellValue(name, cell, h)
		}
	}
	for i, h := range channelHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheetChannel, cell, h)
	}
	if _, err := f.NewSheet(sheetReadme); err == nil {
		_ = f.SetCellValue(sheetReadme, "A1", "采集点表导入模板")
		notes := []string{
			"1. 接入方式：poll=轮询采集（Modbus/BACnet/S7 等结构化点表），report=主动上报（MQTT 订阅报文）",
			"2. 连接配置JSON：随驱动而异，如 Modbus {\"server\":\"tcp://192.168.1.100:502\"}；MQTT {\"server\":\"127.0.0.1:1883\",\"deviceTypeId\":设备类型ID}",
			"3. 驱动：modbus/bacnet/s7/opcua/snmp/fins/mc/iec104/dlt645/eip/mqtt",
			"4. 设备类型：register=寄存器型（按点表直采），report=报文型（须填设备类型名，且该类型已绑定已发布子流程）",
			"5. 测点'所属设备'格式：通道名称/设备名称；地址按驱动格式（Modicon 如 40001；BACnet 如 ai:1）",
			"6. 数据类型：INT16/UINT16/INT32/UINT32/FLOAT32/FLOAT64/BOOL/STRING",
			"7. 通道可引用已有名称做增量更新（按名称匹配）；导入成功后自动编译部署变更通道",
		}
		for i, n := range notes {
			_ = f.SetCellValue(sheetReadme, fmt.Sprintf("A%d", i+3), n)
		}
	}
	return f, nil
}

// Export 全量配置导出（外键列输出名称，支撑"导出→修改→再导入"闭环）。
func (ImportService) Export() (*excelize.File, error) {
	db := global.GVA_DB
	f, err := (ImportService{}).GenerateTemplate()
	if err != nil {
		return nil, err
	}
	var channels []model.CollectChannel
	var devices []model.CollectDevice
	var variables []model.CollectVariable
	if err := db.Order("id").Find(&channels).Error; err != nil {
		return nil, err
	}
	if err := db.Order("id").Find(&devices).Error; err != nil {
		return nil, err
	}
	if err := db.Order("id").Find(&variables).Error; err != nil {
		return nil, err
	}
	chName := map[uint]string{}
	for _, c := range channels {
		chName[c.ID] = c.Name
	}
	devName := map[uint]string{}
	for _, d := range devices {
		devName[d.ID] = d.Name
	}

	r := 2
	for _, c := range channels {
		mqttServer, mqttTopic := outputMQTT(c.OutputConfig)
		enable := "true"
		if c.Enable != nil && !*c.Enable {
			enable = "false"
		}
		writeRow(f, sheetChannel, r, []interface{}{c.Name, c.AccessMode, c.Driver, string(c.ConnConfig),
			c.PollInterval, mqttServer, mqttTopic, enable, c.Remark})
		r++
	}
	r = 2
	for _, d := range devices {
		enable := "true"
		if d.Enable != nil && !*d.Enable {
			enable = "false"
		}
		var unitID interface{}
		if d.UnitID != nil {
			unitID = *d.UnitID
		}
		typeName := ""
		if d.DeviceTypeID != nil {
			var dt model.CollectDeviceType
			if db.First(&dt, *d.DeviceTypeID).Error == nil {
				typeName = dt.Name
			}
		}
		writeRow(f, sheetDevice, r, []interface{}{chName[d.ChannelID], d.Name, d.DeviceKind, typeName,
			unitID, d.PollInterval, enable, d.Remark})
		r++
	}
	r = 2
	for _, v := range variables {
		enable := "true"
		if v.Enable != nil && !*v.Enable {
			enable = "false"
		}
		writeRow(f, sheetVar, r, []interface{}{
			chNameOf(devName, v.DeviceID, chName) + "/" + devName[v.DeviceID], v.Name, v.Addr, v.DataType,
			v.Scale, v.Offset, v.Endian, v.CollectGroup, v.RW, v.Unit, enable, v.Remark})
		r++
	}
	return f, nil
}

func chNameOf(devName map[uint]string, deviceID uint, chName map[uint]string) string {
	// 由 deviceID 反查通道名（两段外键展示）
	for _, d := range func() []model.CollectDevice {
		var ds []model.CollectDevice
		global.GVA_DB.Where("id = ?", deviceID).Find(&ds)
		return ds
	}() {
		return chName[d.ChannelID]
	}
	return ""
}

func writeRow(f *excelize.File, sheet string, row int, vals []interface{}) {
	for i, v := range vals {
		cell, _ := excelize.CoordinatesToCellName(i+1, row)
		_ = f.SetCellValue(sheet, cell, v)
	}
}

func outputMQTT(raw []byte) (server, topic string) {
	var out struct {
		MqttPublish struct {
			Server string `json:"server"`
			Topic  string `json:"topic"`
		} `json:"mqttPublish"`
	}
	if json.Unmarshal(raw, &out) == nil {
		return out.MqttPublish.Server, out.MqttPublish.Topic
	}
	return "", ""
}

// ---------- 预览 ----------

// Preview 解析 Excel 并做全量校验（insert/update 分类 + 错误行收集）。
func (ImportService) Preview(fileBytes []byte) (ImportPreviewResult, error) {
	res := ImportPreviewResult{Stats: map[string]int{}}
	f, err := excelize.OpenReader(io.NopCloser(strings.NewReader(string(fileBytes))))
	if err != nil {
		return res, fmt.Errorf("Excel 打开失败: %w", err)
	}
	defer f.Close()

	db := global.GVA_DB
	// 库内现有名称索引
	var existChannels []model.CollectChannel
	db.Find(&existChannels)
	chByName := map[string]model.CollectChannel{}
	newChNames := map[string]bool{}
	for _, c := range existChannels {
		chByName[c.Name] = c
	}

	// 通道 sheet
	rows, err := f.GetRows(sheetChannel)
	if err != nil {
		return res, fmt.Errorf("缺少「%s」sheet", sheetChannel)
	}
	for i, row := range rows {
		if i == 0 || isEmptyRow(row) {
			continue
		}
		line := i + 1
		get := func(idx int) string {
			if idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}
		name := get(0)
		if name == "" {
			res.Errors = append(res.Errors, ImportRowError{sheetChannel, line, "通道名称为空"})
			continue
		}
		mode := get(1)
		if mode != "poll" && mode != "report" {
			res.Errors = append(res.Errors, ImportRowError{sheetChannel, line, "接入方式必须为 poll/report"})
			continue
		}
		driver := get(2)
		if driver == "" {
			res.Errors = append(res.Errors, ImportRowError{sheetChannel, line, "驱动为空"})
			continue
		}
		conn := get(3)
		if conn == "" {
			conn = "{}"
		}
		var jsCheck map[string]interface{}
		if json.Unmarshal([]byte(conn), &jsCheck) != nil {
			res.Errors = append(res.Errors, ImportRowError{sheetChannel, line, "连接配置JSON非法"})
			continue
		}
		interval, _ := strconv.Atoi(get(4))
		enable := get(7) != "false"
		ch := model.CollectChannel{
			Name: name, AccessMode: mode, Driver: driver,
			ConnConfig:   []byte(conn),
			PollInterval: interval,
			Enable:       &enable,
			Remark:       get(8),
		}
		rowData := ImportChannelRow{CollectChannel: ch, MqttServer: get(5), MqttTopic: get(6)}
		if old, ok := chByName[name]; ok && !newChNames[name] {
			rowData.ID = old.ID // 更新
		} else if newChNames[name] {
			res.Errors = append(res.Errors, ImportRowError{sheetChannel, line, "通道名称在文件内重复"})
			continue
		} else {
			newChNames[name] = true
		}
		res.Data.Channels = append(res.Data.Channels, rowData)
	}
	// 输出配置重算（MQTT 列 → output_config）
	for i := range res.Data.Channels {
		rowData := &res.Data.Channels[i]
		out := map[string]interface{}{"dbWrite": map[string]interface{}{"enable": true}}
		if rowData.MqttTopic != "" {
			out["mqttPublish"] = map[string]interface{}{
				"server": rowData.MqttServer, "topic": rowData.MqttTopic, "qos": 1,
			}
		}
		b, _ := json.Marshal(out)
		rowData.OutputConfig = b
	}

	// 设备 sheet
	devByName := map[string]model.CollectDevice{} // "通道/设备" -> 设备
	var existDevices []model.CollectDevice
	db.Find(&existDevices)
	var existChannelAll []model.CollectChannel
	db.Find(&existChannelAll)
	chIDByName := map[uint]string{}
	for _, c := range existChannelAll {
		chIDByName[c.ID] = c.Name
	}
	for _, d := range existDevices {
		if n, ok := chIDByName[d.ChannelID]; ok {
			devByName[n+"/"+d.Name] = d
		}
	}
	newDevKeys := map[string]bool{}
	var typeNameToID = map[string]uint{}
	var existTypes []model.CollectDeviceType
	db.Find(&existTypes)
	for _, t := range existTypes {
		typeNameToID[t.Name] = t.ID
	}

	drows, err := f.GetRows(sheetDevice)
	if err != nil {
		return res, fmt.Errorf("缺少「%s」sheet", sheetDevice)
	}
	for i, row := range drows {
		if i == 0 || isEmptyRow(row) {
			continue
		}
		line := i + 1
		get := func(idx int) string {
			if idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}
		chName := get(0)
		name := get(1)
		if chName == "" || name == "" {
			res.Errors = append(res.Errors, ImportRowError{sheetDevice, line, "所属通道/设备名称为空"})
			continue
		}
		kind := get(2)
		if kind != "register" && kind != "report" {
			res.Errors = append(res.Errors, ImportRowError{sheetDevice, line, "设备类型必须为 register/report"})
			continue
		}
		var typeID *uint
		if kind == "report" {
			typeName := get(3)
			id, ok := typeNameToID[typeName]
			if !ok {
				res.Errors = append(res.Errors, ImportRowError{sheetDevice, line, "设备类型名不存在或未创建: " + typeName})
				continue
			}
			typeID = &id
		}
		var unitID *int
		if s := get(4); s != "" {
			if v, err := strconv.Atoi(s); err == nil {
				unitID = &v
			}
		}
		interval, _ := strconv.Atoi(get(5))
		enable := get(6) != "false"
		key := chName + "/" + name
		d := model.CollectDevice{
			Name: name, DeviceKind: kind, DeviceTypeID: typeID,
			UnitID: unitID, PollInterval: interval, Enable: &enable, Remark: get(7),
		}
		rowData := ImportDeviceRow{CollectDevice: d, ChannelName: chName}
		if old, ok := devByName[key]; ok && !newDevKeys[key] {
			rowData.ID = old.ID
			rowData.ChannelID = old.ChannelID
		} else if newDevKeys[key] {
			res.Errors = append(res.Errors, ImportRowError{sheetDevice, line, "设备在文件内重复: " + key})
			continue
		} else {
			newDevKeys[key] = true
		}
		res.Data.Devices = append(res.Data.Devices, rowData)
	}

	// 测点 sheet
	varNameByKey := map[string]model.CollectVariable{}
	var existVars []model.CollectVariable
	db.Find(&existVars)
	devKeyByID := map[uint]string{} // 设备ID -> "通道名/设备名"
	for _, d := range existDevices {
		if n, ok := chIDByName[d.ChannelID]; ok {
			devKeyByID[d.ID] = n + "/" + d.Name
		}
	}
	for _, v := range existVars {
		if dk, ok := devKeyByID[v.DeviceID]; ok {
			varNameByKey[keyOf(dk, v.Name)] = v
		}
	}
	newVarKeys := map[string]bool{}
	vrows, err := f.GetRows(sheetVar)
	if err != nil {
		return res, fmt.Errorf("缺少「%s」sheet", sheetVar)
	}
	validTypes := map[string]bool{"INT16": true, "UINT16": true, "INT32": true, "UINT32": true, "FLOAT32": true, "FLOAT64": true, "BOOL": true, "STRING": true}
	for i, row := range vrows {
		if i == 0 || isEmptyRow(row) {
			continue
		}
		line := i + 1
		get := func(idx int) string {
			if idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}
		devKey := get(0) // 通道/设备
		name := get(1)
		addr := get(2)
		dataType := get(3)
		if devKey == "" || name == "" || addr == "" {
			res.Errors = append(res.Errors, ImportRowError{sheetVar, line, "所属设备/测点名称/地址为空"})
			continue
		}
		if !validTypes[dataType] {
			res.Errors = append(res.Errors, ImportRowError{sheetVar, line, "数据类型非法: " + dataType})
			continue
		}
		scale, _ := strconv.ParseFloat(get(4), 64)
		offset, _ := strconv.ParseFloat(get(5), 64)
		enable := get(10) != "false"
		parts := strings.SplitN(devKey, "/", 2)
		rowData := ImportVariableRow{
			CollectVariable: model.CollectVariable{
				Name: name, Addr: addr, DataType: dataType,
				Scale: scale, Offset: offset, Endian: get(6), CollectGroup: get(7),
				RW: get(8), Unit: get(9), Enable: &enable, Remark: get(11),
			},
			ChannelName: parts[0],
		}
		if len(parts) == 2 {
			rowData.DeviceName = parts[1]
		}
		key := keyOf(devKey, name)
		if old, ok := varNameByKey[key]; ok && !newVarKeys[key] {
			rowData.ID = old.ID
			rowData.DeviceID = old.DeviceID
		} else if newVarKeys[key] {
			res.Errors = append(res.Errors, ImportRowError{sheetVar, line, "测点在文件内重复: " + key})
			continue
		} else {
			newVarKeys[key] = true
		}
		res.Data.Variables = append(res.Data.Variables, rowData)
	}

	res.Stats["channels"] = len(res.Data.Channels)
	res.Stats["devices"] = len(res.Data.Devices)
	res.Stats["variables"] = len(res.Data.Variables)
	res.Stats["errors"] = len(res.Errors)
	return res, nil
}

func keyOf(devKey, varName string) string { return devKey + ":" + varName }

func isEmptyRow(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

// ---------- 提交（服务端重新校验 + 事务 upsert + 部署联动） ----------

// CommitResult 提交结果。
type CommitResult struct {
	Channels       int             `json:"channels"`
	Devices        int             `json:"devices"`
	Variables      int             `json:"variables"`
	Deployed       int             `json:"deployed"`
	Skipped        int             `json:"skipped"`
	DeployFailures map[uint]string `json:"deployFailures"`
}

// Commit 导入提交：重新校验 → 事务 upsert（通道→设备→测点，名称外键回填 ID）→ 自动部署。
func (ImportService) Commit(p ImportPayload, autoDeploy bool, operatorID uint) (CommitResult, error) {
	var res CommitResult
	res.DeployFailures = map[uint]string{}
	db := global.GVA_DB

	touchedChannels := map[uint]bool{}
	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. 通道
		for _, rc := range p.Channels {
			ch := rc.CollectChannel
			var existing model.CollectChannel
			if rc.ID != 0 && tx.First(&existing, rc.ID).Error == nil {
				ch.ID = existing.ID
				if err := (ChannelService{}).UpdateChannel(&ch); err != nil {
					return fmt.Errorf("更新通道 %s: %w", ch.Name, err)
				}
			} else {
				if err := (ChannelService{}).CreateChannel(&ch); err != nil {
					return fmt.Errorf("创建通道 %s: %w", ch.Name, err)
				}
			}
			res.Channels++
			touchedChannels[ch.ID] = true
		}
		// 通道名 → ID（供设备行回填）
		chIDByName := map[string]uint{}
		var allChannels []model.CollectChannel
		tx.Unscoped().Model(&model.CollectChannel{}).Find(&allChannels)
		for _, c := range allChannels {
			chIDByName[c.Name] = c.ID
		}
		// 2. 设备
		for _, rd := range p.Devices {
			d := rd.CollectDevice
			if d.ChannelID == 0 {
				id, ok := chIDByName[rd.ChannelName]
				if !ok {
					return fmt.Errorf("设备 %s 的通道不存在: %s", d.Name, rd.ChannelName)
				}
				d.ChannelID = id
			}
			var existing model.CollectDevice
			if rd.ID != 0 && tx.First(&existing, rd.ID).Error == nil {
				d.ID = existing.ID
				if err := (DeviceService{}).UpdateDevice(&d); err != nil {
					return fmt.Errorf("更新设备 %s: %w", d.Name, err)
				}
			} else {
				if err := (DeviceService{}).CreateDevice(&d); err != nil {
					return fmt.Errorf("创建设备 %s: %w", d.Name, err)
				}
			}
			res.Devices++
			touchedChannels[d.ChannelID] = true
		}
		// 设备唯一键（通道名/设备名）→ ID（供测点行回填）
		devIDByKey := map[string]uint{}
		chNameByID := map[uint]string{}
		var allChannels2 []model.CollectChannel
		tx.Unscoped().Model(&model.CollectChannel{}).Find(&allChannels2)
		for _, c := range allChannels2 {
			chNameByID[c.ID] = c.Name
		}
		var allDevices []model.CollectDevice
		tx.Unscoped().Model(&model.CollectDevice{}).Find(&allDevices)
		for _, d := range allDevices {
			if n, ok := chNameByID[d.ChannelID]; ok {
				devIDByKey[n+"/"+d.Name] = d.ID
			}
		}
		// 3. 测点
		for _, rv := range p.Variables {
			v := rv.CollectVariable
			if v.DeviceID == 0 {
				id, ok := devIDByKey[rv.ChannelName+"/"+rv.DeviceName]
				if !ok {
					return fmt.Errorf("测点 %s 的设备不存在: %s/%s", v.Name, rv.ChannelName, rv.DeviceName)
				}
				v.DeviceID = id
			}
			var existing model.CollectVariable
			if rv.ID != 0 && tx.First(&existing, rv.ID).Error == nil {
				v.ID = existing.ID
				if err := (VariableService{}).UpdateVariable(&v); err != nil {
					return fmt.Errorf("更新测点 %s: %w", v.Name, err)
				}
			} else {
				if err := (VariableService{}).CreateVariable(&v); err != nil {
					return fmt.Errorf("创建测点 %s: %w", v.Name, err)
				}
			}
			res.Variables++
		}
		return nil
	})
	if err != nil {
		return res, err
	}

	if autoDeploy {
		for chID := range touchedChannels {
			skipped, err := DeployChannel(chID, operatorID)
			if err != nil {
				res.DeployFailures[chID] = err.Error()
				continue
			}
			if skipped {
				res.Skipped++
			} else {
				res.Deployed++
			}
		}
	}
	return res, nil
}
