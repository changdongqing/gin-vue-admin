package initialize

import (
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/service"
	"go.uber.org/zap"
)

// 采集演示业务种子：两套模拟采集场景（幂等——任一主体曾存在（含软删）即整套跳过）。
//
// 场景一 report 通道「演示-网关模拟量上报」：网关→8路模拟量采集模块→温湿度/压力/水浸
// 三传感器，网关每 5s 向 MQTT topic gateway/analog/data 发包含所有设备所有测点的大 JSON
// （样例 service.DemoGatewaySample，也是子流程回放样例）；平台 conn_config.server/topic
// 订阅，子流程按 port 映射平台设备。喂数：mosquitto_pub -t gateway/analog/data -m '<样例>'。
//
// 场景二 poll 通道「演示-ModbusTCP电表」：平台直连 Modbus TCP（默认 tcp://127.0.0.1:5020，
// 可用 diagslave -m tcp -p 5020 起模拟器），站号 1/2/3 对应三块电表，每表 7 测点
// （正向有功电能 + 三相电压/电流），按 E/U/I 分组批量读取，通道周期 5s。
//
// 落库后异步等 bridge（同进程 gin）开始监听：注册子流程到引擎并部署两个通道（失败仅告警）。

// seedCollectDemo 演示业务种子入口（SeedCollect 末尾调用）。
func seedCollectDemo() {
	db := global.GVA_DB
	var n int64
	db.Unscoped().Model(&model.CollectParseChain{}).Where("name = ?", service.DemoChainName).Count(&n)
	if n > 0 {
		return
	}

	// ---- 场景二：Modbus TCP 电表（独立场景，先建）----
	mb := model.CollectChannel{
		Name: service.DemoModbusCh, AccessMode: "poll", Driver: "modbus",
		ConnConfig:   []byte(`{"server":"tcp://127.0.0.1:5020"}`),
		PollInterval: 5000, Enable: boolPtr(true),
		Remark: "演示：平台直连 Modbus TCP，站号1/2/3=三块电表（电能+三相电压电流）；默认指向本机 5020 模拟器，按需改 server",
	}
	if err := db.Create(&mb).Error; err != nil {
		global.GVA_LOG.Error("采集演示种子：modbus通道创建失败", zap.Error(err))
		return
	}
	for unit := 1; unit <= 3; unit++ {
		dev := model.CollectDevice{
			ChannelID: mb.ID, Name: fmt.Sprintf("电表%d", unit),
			DeviceKind: "register", UnitID: &unit, Enable: boolPtr(true),
			Remark: "演示电表（站号 " + fmt.Sprintf("%d", unit) + "）",
		}
		if err := db.Create(&dev).Error; err != nil {
			global.GVA_LOG.Error("采集演示种子：电表设备创建失败", zap.Error(err))
			return
		}
		vars := meterVariables(dev.ID)
		if err := db.Create(&vars).Error; err != nil {
			global.GVA_LOG.Error("采集演示种子：电表测点创建失败", zap.Uint("deviceId", dev.ID), zap.Error(err))
			return
		}
	}

	// ---- 场景一：网关大 JSON（子流程 → 设备类型 → 通道 → 设备/测点）----
	pc := model.CollectParseChain{
		Name: service.DemoChainName, Version: 1, Status: "draft",
		InputContract: "msgType=JSON：{gateway,ts,sensors:[{port,temp?,hum?,pressure?,leak?}]}，" +
			"port 1=温湿度 2=压力 3=水浸，ts 毫秒；输出 devices[] 元素 {id(平台设备ID),device,points:[{name,value,quality,ts}]}",
		TestPayload: service.DemoGatewaySample,
		Remark:      "演示：8路模拟量采集模块经网关上报的全量大 JSON 解析（port→平台设备ID 映射内嵌脚本）",
	}
	if err := db.Create(&pc).Error; err != nil {
		global.GVA_LOG.Error("采集演示种子：子流程创建失败", zap.Error(err))
		return
	}
	dt := model.CollectDeviceType{
		Name: service.DemoTypeName, PayloadType: "json", ParseChainID: &pc.ID, Enable: boolPtr(true),
		Remark: "演示：模拟量网关大 JSON（绑定演示解析子流程）",
	}
	if err := db.Create(&dt).Error; err != nil {
		global.GVA_LOG.Error("采集演示种子：设备类型创建失败", zap.Error(err))
		return
	}
	rep := model.CollectChannel{
		Name: service.DemoReportCh, AccessMode: "report", Driver: "",
		ConnConfig:   []byte(fmt.Sprintf(`{"server":"tcp://127.0.0.1:1883","topic":"gateway/analog/data","qos":0,"deviceTypeId":%d}`, dt.ID)),
		Enable:       boolPtr(true),
		Remark:       "演示：网关每5s上报全量大JSON（温湿度/压力/水浸），平台 MQTT 订阅；broker 默认本机 1883，需有 broker 才能挂载订阅",
	}
	if err := db.Create(&rep).Error; err != nil {
		global.GVA_LOG.Error("采集演示种子：上报通道创建失败", zap.Error(err))
		return
	}
	type sensorSeed struct {
		name, remark string
		vars         []model.CollectVariable
	}
	sensors := []sensorSeed{
		{"温湿度传感器", "演示：8路模块 port1", []model.CollectVariable{
			{Name: "温度", Addr: "port1.temp", DataType: "FLOAT32", RW: "R", Unit: "℃", Enable: boolPtr(true)},
			{Name: "湿度", Addr: "port1.hum", DataType: "FLOAT32", RW: "R", Unit: "%RH", Enable: boolPtr(true)},
		}},
		{"压力传感器", "演示：8路模块 port2", []model.CollectVariable{
			{Name: "压力", Addr: "port2.pressure", DataType: "FLOAT32", RW: "R", Unit: "MPa", Enable: boolPtr(true)},
		}},
		{"水浸传感器", "演示：8路模块 port3", []model.CollectVariable{
			{Name: "水浸状态", Addr: "port3.leak", DataType: "STRING", RW: "R", Unit: "-", Enable: boolPtr(true)},
		}},
	}
	devIDs := make(map[string]uint, len(sensors))
	for _, s := range sensors {
		dev := model.CollectDevice{
			ChannelID: rep.ID, Name: s.name, DeviceKind: "report",
			DeviceTypeID: &dt.ID, Enable: boolPtr(true), Remark: s.remark,
		}
		if err := db.Create(&dev).Error; err != nil {
			global.GVA_LOG.Error("采集演示种子：上报设备创建失败", zap.String("name", s.name), zap.Error(err))
			return
		}
		devIDs[s.name] = dev.ID
		for i := range s.vars {
			s.vars[i].DeviceID = dev.ID
		}
		if err := db.Create(&s.vars).Error; err != nil {
			global.GVA_LOG.Error("采集演示种子：上报测点创建失败", zap.String("name", s.name), zap.Error(err))
			return
		}
	}

	// 子流程 DSL 依赖设备 ID，最后回填并直接置为已发布（引擎注册在 bridge 就绪后异步补做）
	dsl, err := service.BuildDemoGatewayParseDSL(pc.ID, devIDs["温湿度传感器"], devIDs["压力传感器"], devIDs["水浸传感器"])
	if err != nil {
		global.GVA_LOG.Error("采集演示种子：子流程DSL构造失败", zap.Error(err))
		return
	}
	if err := db.Model(&pc).Updates(map[string]interface{}{"dsl": string(dsl), "status": "published"}).Error; err != nil {
		global.GVA_LOG.Error("采集演示种子：子流程DSL回填失败", zap.Error(err))
		return
	}

	// 子流程引擎注册 + 通道部署统一走启动对账（bridge 就绪后执行，幂等）
	global.GVA_LOG.Info("采集演示种子：两场景配置已就绪，等待 bridge 部署")
	service.ReconcileChainsOnStartup()
}

// meterVariables 单块电表的 7 个测点（Modicon 地址，E/U/I 分组批量读取）。
func meterVariables(deviceID uint) []model.CollectVariable {
	return []model.CollectVariable{
		{DeviceID: deviceID, Name: "正向有功电能", Addr: "40001", DataType: "FLOAT32", Scale: 1, CollectGroup: "E", RW: "R", Unit: "kWh", Enable: boolPtr(true)},
		{DeviceID: deviceID, Name: "A相电压", Addr: "40011", DataType: "UINT16", Scale: 0.1, CollectGroup: "U", RW: "R", Unit: "V", Enable: boolPtr(true)},
		{DeviceID: deviceID, Name: "B相电压", Addr: "40012", DataType: "UINT16", Scale: 0.1, CollectGroup: "U", RW: "R", Unit: "V", Enable: boolPtr(true)},
		{DeviceID: deviceID, Name: "C相电压", Addr: "40013", DataType: "UINT16", Scale: 0.1, CollectGroup: "U", RW: "R", Unit: "V", Enable: boolPtr(true)},
		{DeviceID: deviceID, Name: "A相电流", Addr: "40021", DataType: "UINT16", Scale: 0.001, CollectGroup: "I", RW: "R", Unit: "A", Enable: boolPtr(true)},
		{DeviceID: deviceID, Name: "B相电流", Addr: "40022", DataType: "UINT16", Scale: 0.001, CollectGroup: "I", RW: "R", Unit: "A", Enable: boolPtr(true)},
		{DeviceID: deviceID, Name: "C相电流", Addr: "40023", DataType: "UINT16", Scale: 0.001, CollectGroup: "I", RW: "R", Unit: "A", Enable: boolPtr(true)},
	}
}

func boolPtr(v bool) *bool { return &v }
