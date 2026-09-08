package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/model"
	"gorm.io/gorm"
)

// 通道/设备/测点 CRUD、树查询、实时数据查询与运行状态推导（03 文档 §六/§八）。

// ChannelService 通道服务。
type ChannelService struct{}

// CreateChannel 创建通道（名称唯一）。
func (ChannelService) CreateChannel(ch *model.CollectChannel) error {
	db := global.GVA_DB
	var count int64
	db.Model(&model.CollectChannel{}).Where("name = ?", ch.Name).Count(&count)
	if count > 0 {
		return errors.New("通道名称已存在: " + ch.Name)
	}
	return db.Create(ch).Error
}

// UpdateChannel 更新通道。
func (ChannelService) UpdateChannel(ch *model.CollectChannel) error {
	db := global.GVA_DB
	var old model.CollectChannel
	if err := db.First(&old, ch.ID).Error; err != nil {
		return errors.New("通道不存在")
	}
	var dup int64
	db.Model(&model.CollectChannel{}).Where("name = ? AND id <> ?", ch.Name, ch.ID).Count(&dup)
	if dup > 0 {
		return errors.New("通道名称已存在: " + ch.Name)
	}
	return db.Model(&old).Omit("created_at").Updates(map[string]interface{}{
		"name": ch.Name, "access_mode": ch.AccessMode, "driver": ch.Driver,
		"conn_config": ch.ConnConfig, "poll_interval": ch.PollInterval,
		"output_config": ch.OutputConfig, "enable": ch.Enable, "remark": ch.Remark,
	}).Error
}

// DeleteChannel 删除通道（级联软删设备/测点，并下线 rulego 链）。
func (ChannelService) DeleteChannel(id uint) error {
	db := global.GVA_DB
	var ch model.CollectChannel
	if err := db.First(&ch, id).Error; err != nil {
		return errors.New("通道不存在")
	}
	var deviceIds []uint
	db.Model(&model.CollectDevice{}).Where("channel_id = ?", id).Pluck("id", &deviceIds)
	return db.Transaction(func(tx *gorm.DB) error {
		if len(deviceIds) > 0 {
			if err := tx.Where("device_id IN ?", deviceIds).Delete(&model.CollectVariable{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("channel_id = ?", id).Delete(&model.CollectDevice{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&ch).Error; err != nil {
			return err
		}
		// 实时值同步清理
		return tx.Where("channel_id = ?", id).Delete(&model.CollectRealtime{}).Error
	})
}

// GetChannelList 通道全量列表（量级小，不分页）。
func (ChannelService) GetChannelList() ([]model.CollectChannel, error) {
	var list []model.CollectChannel
	err := global.GVA_DB.Order("id").Find(&list).Error
	return list, err
}

// FindChannel 通道详情。
func (ChannelService) FindChannel(id uint) (model.CollectChannel, error) {
	var ch model.CollectChannel
	err := global.GVA_DB.First(&ch, id).Error
	return ch, err
}

// DeviceService 设备服务。
type DeviceService struct{}

// CreateDevice 创建设备（通道内名称唯一）。
func (DeviceService) CreateDevice(d *model.CollectDevice) error {
	db := global.GVA_DB
	var count int64
	db.Model(&model.CollectDevice{}).Where("channel_id = ? AND name = ?", d.ChannelID, d.Name).Count(&count)
	if count > 0 {
		return fmt.Errorf("设备名称已存在: %s", d.Name)
	}
	if d.DeviceKind == "report" && d.DeviceTypeID == nil {
		return errors.New("报文型设备必须绑定设备类型")
	}
	return db.Create(d).Error
}

// UpdateDevice 更新设备。
func (DeviceService) UpdateDevice(d *model.CollectDevice) error {
	db := global.GVA_DB
	var old model.CollectDevice
	if err := db.First(&old, d.ID).Error; err != nil {
		return errors.New("设备不存在")
	}
	var dup int64
	db.Model(&model.CollectDevice{}).Where("channel_id = ? AND name = ? AND id <> ?", d.ChannelID, d.Name, d.ID).Count(&dup)
	if dup > 0 {
		return fmt.Errorf("设备名称已存在: %s", d.Name)
	}
	return db.Model(&old).Omit("created_at").Updates(map[string]interface{}{
		"channel_id": d.ChannelID, "name": d.Name, "device_kind": d.DeviceKind,
		"device_type_id": d.DeviceTypeID, "unit_id": d.UnitID, "poll_interval": d.PollInterval,
		"props": d.Props, "enable": d.Enable, "remark": d.Remark,
	}).Error
}

// DeleteDevice 删除设备（级联软删测点 + 清实时值）。
func (DeviceService) DeleteDevice(id uint) error {
	db := global.GVA_DB
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("device_id = ?", id).Delete(&model.CollectVariable{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).Delete(&model.CollectDevice{}).Error; err != nil {
			return err
		}
		return tx.Where("device_id = ?", id).Delete(&model.CollectRealtime{}).Error
	})
}

// GetDeviceListByChannel 通道下设备列表。
func (DeviceService) GetDeviceListByChannel(channelID uint) ([]model.CollectDevice, error) {
	var list []model.CollectDevice
	err := global.GVA_DB.Where("channel_id = ?", channelID).Order("id").Find(&list).Error
	return list, err
}

// VariableService 测点服务。
type VariableService struct{}

// CreateVariable 创建测点（设备内名称唯一）。
func (VariableService) CreateVariable(v *model.CollectVariable) error {
	db := global.GVA_DB
	var count int64
	db.Model(&model.CollectVariable{}).Where("device_id = ? AND name = ?", v.DeviceID, v.Name).Count(&count)
	if count > 0 {
		return fmt.Errorf("测点名称已存在: %s", v.Name)
	}
	return db.Create(v).Error
}

// UpdateVariable 更新测点。
func (VariableService) UpdateVariable(v *model.CollectVariable) error {
	db := global.GVA_DB
	var old model.CollectVariable
	if err := db.First(&old, v.ID).Error; err != nil {
		return errors.New("测点不存在")
	}
	var dup int64
	db.Model(&model.CollectVariable{}).Where("device_id = ? AND name = ? AND id <> ?", v.DeviceID, v.Name, v.ID).Count(&dup)
	if dup > 0 {
		return fmt.Errorf("测点名称已存在: %s", v.Name)
	}
	return db.Model(&old).Omit("created_at").Updates(map[string]interface{}{
		"device_id": v.DeviceID, "name": v.Name, "addr": v.Addr, "data_type": v.DataType,
		"scale": v.Scale, "offset": v.Offset, "endian": v.Endian, "collect_group": v.CollectGroup,
		"read_expr": v.ReadExpr, "rw": v.RW, "unit": v.Unit, "enable": v.Enable, "remark": v.Remark,
	}).Error
}

// DeleteVariable 删除测点（清实时值）。
func (VariableService) DeleteVariable(id uint) error {
	db := global.GVA_DB
	var v model.CollectVariable
	if err := db.First(&v, id).Error; err != nil {
		return errors.New("测点不存在")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&v).Error; err != nil {
			return err
		}
		return tx.Where("device_id = ? AND name = ?", v.DeviceID, v.Name).Delete(&model.CollectRealtime{}).Error
	})
}

// GetVariableListByDevice 设备下测点列表。
func (VariableService) GetVariableListByDevice(deviceID uint) ([]model.CollectVariable, error) {
	var list []model.CollectVariable
	err := global.GVA_DB.Where("device_id = ?", deviceID).Order("id").Find(&list).Error
	return list, err
}

// CollectTree 通道-设备-测点树节点。
type CollectTree struct {
	model.CollectChannel
	Devices []CollectTreeDevice `json:"devices"`
}

// CollectTreeDevice 树中设备节点。
type CollectTreeDevice struct {
	model.CollectDevice
	Variables []model.CollectVariable `json:"variables"`
}

// GetTree 全量三级树（左侧导航用）。
func (ChannelService) GetTree() ([]CollectTree, error) {
	chs, err := (ChannelService{}).GetChannelList()
	if err != nil {
		return nil, err
	}
	db := global.GVA_DB
	var devices []model.CollectDevice
	if err := db.Order("id").Find(&devices).Error; err != nil {
		return nil, err
	}
	var variables []model.CollectVariable
	if err := db.Order("id").Find(&variables).Error; err != nil {
		return nil, err
	}
	varsByDev := map[uint][]model.CollectVariable{}
	for _, v := range variables {
		varsByDev[v.DeviceID] = append(varsByDev[v.DeviceID], v)
	}
	devsByCh := map[uint][]CollectTreeDevice{}
	for _, d := range devices {
		devsByCh[d.ChannelID] = append(devsByCh[d.ChannelID], CollectTreeDevice{CollectDevice: d, Variables: varsByDev[d.ID]})
	}
	tree := make([]CollectTree, 0, len(chs))
	for _, c := range chs {
		tree = append(tree, CollectTree{CollectChannel: c, Devices: devsByCh[c.ID]})
	}
	return tree, nil
}

// ParseChainService 子流程库服务。
type ParseChainService struct{}

// CreateParseChain 创建子流程（草稿）。
func (ParseChainService) CreateParseChain(pc *model.CollectParseChain) error {
	db := global.GVA_DB
	var count int64
	db.Model(&model.CollectParseChain{}).Where("name = ?", pc.Name).Count(&count)
	if count > 0 {
		return errors.New("子流程名已存在: " + pc.Name)
	}
	pc.Status = "draft"
	pc.Version = 1
	return db.Create(pc).Error
}

// UpdateParseChain 更新子流程（仅 draft 可改 DSL）。
func (ParseChainService) UpdateParseChain(pc *model.CollectParseChain) error {
	db := global.GVA_DB
	var old model.CollectParseChain
	if err := db.First(&old, pc.ID).Error; err != nil {
		return errors.New("子流程不存在")
	}
	if old.Status == "published" {
		return errors.New("已发布子流程不可编辑，请新建版本")
	}
	return db.Model(&old).Omit("created_at").Updates(map[string]interface{}{
		"name": pc.Name, "dsl": pc.Dsl, "input_contract": pc.InputContract,
		"test_payload": pc.TestPayload, "remark": pc.Remark,
	}).Error
}

// PublishParseChain 发布子流程（部署到 rulego 供通道引用）。
func (ParseChainService) PublishParseChain(id uint, operatorID uint) error {
	db := global.GVA_DB
	var pc model.CollectParseChain
	if err := db.First(&pc, id).Error; err != nil {
		return errors.New("子流程不存在")
	}
	if pc.Dsl == "" {
		return errors.New("子流程 DSL 为空")
	}
	chainID := ParseChainRuleID(pc.ID)
	if err := BridgeClient().SaveChain(chainID, []byte(pc.Dsl)); err != nil {
		return err
	}
	return db.Model(&pc).Updates(map[string]interface{}{"status": "published", "version": pc.Version + 1}).Error
}

// DeleteParseChain 删除子流程（published 且被引用时拒绝）。
func (ParseChainService) DeleteParseChain(id uint) error {
	db := global.GVA_DB
	var count int64
	db.Model(&model.CollectDeviceType{}).Where("parse_chain_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("子流程被设备类型引用，先解绑再删除")
	}
	var pc model.CollectParseChain
	if err := db.First(&pc, id).Error; err != nil {
		return errors.New("子流程不存在")
	}
	_ = BridgeClient().DeleteChain(ParseChainRuleID(id))
	return db.Delete(&pc).Error
}

// GetParseChainList 子流程列表。
func (ParseChainService) GetParseChainList() ([]model.CollectParseChain, error) {
	var list []model.CollectParseChain
	err := global.GVA_DB.Order("id").Find(&list).Error
	return list, err
}

// RealtimeService 实时数据服务。
type RealtimeService struct{}

// RealtimeQuery 实时值查询条件。
type RealtimeQuery struct {
	ChannelID uint   `json:"channelId" form:"channelId"`
	DeviceID  uint   `json:"deviceId" form:"deviceId"`
	Quality   string `json:"quality" form:"quality"`
	Limit     int    `json:"limit" form:"limit"`
}

// GetRealtime 实时值查询（默认上限 1000）。
func (RealtimeService) GetRealtime(q RealtimeQuery) ([]model.CollectRealtime, error) {
	db := global.GVA_DB.Model(&model.CollectRealtime{})
	if q.ChannelID != 0 {
		db = db.Where("channel_id = ?", q.ChannelID)
	}
	if q.DeviceID != 0 {
		db = db.Where("device_id = ?", q.DeviceID)
	}
	if q.Quality != "" {
		db = db.Where("quality = ?", q.Quality)
	}
	if q.Limit <= 0 || q.Limit > 5000 {
		q.Limit = 1000
	}
	var list []model.CollectRealtime
	err := db.Order("point_key").Limit(q.Limit).Find(&list).Error
	return list, err
}

// ChannelRuntimeStatus 通道运行状态（03 文档 §8.3 无侵入推导）。
type ChannelRuntimeStatus struct {
	ChannelID    uint    `json:"channelId"`
	DeployStatus string  `json:"deployStatus"` // deployed/none/rolled_back
	Running      bool    `json:"running"`
	GoodRate     float64 `json:"goodRate"` // 0~1，实时表 good 占比
	TotalPoints  int64   `json:"totalPoints"`
	LastTs       int64   `json:"lastTs"` // 全通道最新采集时间（unix ms，0=无数据）
}

// GetChannelStatus 通道运行状态（部署状态 + 数据新鲜度推导）。
func (RealtimeService) GetChannelStatus(channelID uint) (ChannelRuntimeStatus, error) {
	db := global.GVA_DB
	st := ChannelRuntimeStatus{ChannelID: channelID, DeployStatus: "none"}

	var dep model.CollectDeployment
	if err := db.Where("channel_id = ?", channelID).Order("id desc").First(&dep).Error; err == nil {
		st.DeployStatus = dep.Status
		st.Running = dep.Status == "deployed"
	}

	var ch model.CollectChannel
	if err := db.First(&ch, channelID).Error; err == nil && (ch.Enable == nil || !*ch.Enable) {
		st.Running = false
	}

	var total, good int64
	db.Model(&model.CollectRealtime{}).Where("channel_id = ?", channelID).Count(&total)
	db.Model(&model.CollectRealtime{}).Where("channel_id = ? AND quality = ?", channelID, "good").Count(&good)
	st.TotalPoints = total
	if total > 0 {
		st.GoodRate = float64(good) / float64(total)
	}
	var last model.CollectRealtime
	if err := db.Where("channel_id = ?", channelID).Order("ts desc").First(&last).Error; err == nil {
		st.LastTs = last.Ts
	}
	return st, nil
}

// IsDeviceOnline 设备在线判定：最新数据距今 < 3×采集周期。
func IsDeviceOnline(device model.CollectDevice, channelInterval int, lastTs int64) bool {
	if lastTs == 0 {
		return false
	}
	period := device.PollInterval
	if period <= 0 {
		period = channelInterval
	}
	if period <= 0 {
		period = 5000
	}
	return time.Now().UnixMilli()-lastTs < int64(3*period)
}

// DB 库句柄（api 层便捷访问）。
func DB() *gorm.DB { return global.GVA_DB }

// TestParseChain 子流程回放：将 DSL 临时保存为回放链，同步执行后清理，返回引擎输出。
func (ParseChainService) TestParseChain(id uint) (string, error) {
	db := global.GVA_DB
	var pc model.CollectParseChain
	if err := db.First(&pc, id).Error; err != nil {
		return "", errors.New("子流程不存在")
	}
	if pc.Dsl == "" {
		return "", errors.New("子流程 DSL 为空")
	}
	testID := fmt.Sprintf("collect_pc_test_%d", pc.ID)
	bc := BridgeClient()
	if err := bc.SaveChain(testID, []byte(pc.Dsl)); err != nil {
		return "", fmt.Errorf("回放链部署失败: %w", err)
	}
	defer func() { _ = bc.DeleteChain(testID) }()
	msgType := "JSON"
	if pc.TestPayload != "" && pc.TestPayload[0] != '{' && pc.TestPayload[0] != '[' {
		msgType = "STRING"
	}
	return bc.ExecuteChain(testID, msgType, pc.TestPayload)
}

// DeviceTypeService 设备类型服务（报文型绑定子流程）。
type DeviceTypeService struct{}

// GetDeviceTypeList 设备类型列表。
func (DeviceTypeService) GetDeviceTypeList() ([]model.CollectDeviceType, error) {
	var list []model.CollectDeviceType
	err := global.GVA_DB.Order("id").Find(&list).Error
	return list, err
}

// CreateDeviceType 创建设备类型。
func (DeviceTypeService) CreateDeviceType(dt *model.CollectDeviceType) error {
	var count int64
	global.GVA_DB.Model(&model.CollectDeviceType{}).Where("name = ?", dt.Name).Count(&count)
	if count > 0 {
		return errors.New("设备类型名已存在: " + dt.Name)
	}
	return global.GVA_DB.Create(dt).Error
}

// UpdateDeviceType 更新设备类型。
func (DeviceTypeService) UpdateDeviceType(dt *model.CollectDeviceType) error {
	var old model.CollectDeviceType
	if err := global.GVA_DB.First(&old, dt.ID).Error; err != nil {
		return errors.New("设备类型不存在")
	}
	return global.GVA_DB.Model(&old).Omit("created_at").Updates(map[string]interface{}{
		"name": dt.Name, "payload_type": dt.PayloadType, "parse_chain_id": dt.ParseChainID,
		"enable": dt.Enable, "remark": dt.Remark,
	}).Error
}

// DeleteDeviceType 删除设备类型（被设备引用时拒绝）。
func (DeviceTypeService) DeleteDeviceType(id uint) error {
	var count int64
	global.GVA_DB.Model(&model.CollectDevice{}).Where("device_type_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("设备类型被设备引用，先删除或改绑设备")
	}
	return global.GVA_DB.Delete(&model.CollectDeviceType{}, id).Error
}
