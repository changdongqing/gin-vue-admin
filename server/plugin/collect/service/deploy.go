package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 部署联动（03 文档 §4.6）：编译 → dsl_hash 幂等比较 → bridge 保存（自动部署）→ 记录 deployment。

var deployLock sync.Mutex // 通道部署互斥（单实例部署语义）

// realtimeDSN collect_realtime 所在库 DSN（与 GVA 主库同库）。
func realtimeDSN() string {
	return global.GVA_CONFIG.Pgsql.Dsn()
}

// loadCompileInput 从库装配编译输入（使能过滤：设备/测点 enable，子流程 published）。
func loadCompileInput(db *gorm.DB, ch model.CollectChannel) (compileInput, error) {
	in := compileInput{Channel: ch, RealtimeDSN: realtimeDSN()}

	var devices []model.CollectDevice
	if err := db.Where("channel_id = ? AND enable = ?", ch.ID, true).Find(&devices).Error; err != nil {
		return in, err
	}
	in.Devices = devices

	var variables []model.CollectVariable
	if err := db.Where("enable = ?", true).Find(&variables).Error; err != nil {
		return in, err
	}
	in.Variables = map[uint][]model.CollectVariable{}
	for _, v := range variables {
		in.Variables[v.DeviceID] = append(in.Variables[v.DeviceID], v)
	}

	var types []model.CollectDeviceType
	if err := db.Where("enable = ?", true).Find(&types).Error; err != nil {
		return in, err
	}
	in.DeviceTypes = map[uint]model.CollectDeviceType{}
	for _, t := range types {
		in.DeviceTypes[t.ID] = t
	}

	var chains []model.CollectParseChain
	if err := db.Find(&chains).Error; err != nil {
		return in, err
	}
	in.ParseChains = map[uint]model.CollectParseChain{}
	for _, pc := range chains {
		in.ParseChains[pc.ID] = pc
	}
	return in, nil
}

// DeployChannel 编译并部署单通道（幂等：dsl_hash 与最新 deployed 记录一致则跳过）。
// 返回 skipped=true 表示无变更未触发引擎动作。
func DeployChannel(channelID uint, operatorID uint) (skipped bool, err error) {
	deployLock.Lock()
	defer deployLock.Unlock()

	db := global.GVA_DB
	var ch model.CollectChannel
	if err := db.First(&ch, channelID).Error; err != nil {
		return false, fmt.Errorf("通道不存在: %w", err)
	}

	in, err := loadCompileInput(db, ch)
	if err != nil {
		return false, err
	}
	dsl, hash, err := Compile(in)
	if err != nil {
		return false, fmt.Errorf("编译失败: %w", err)
	}

	// 幂等：与当前 deployed 哈希一致 → 跳过
	var last model.CollectDeployment
	hasLast := db.Where("channel_id = ? AND status = ?", ch.ID, "deployed").
		Order("id desc").First(&last).Error == nil
	if hasLast && last.DslHash == hash && BridgeClient().ChainExists(ChainID(ch.ID)) {
		return true, nil
	}

	chainID := ChainID(ch.ID)
	if err := BridgeClient().SaveChain(chainID, dsl); err != nil {
		return false, err
	}
	rec := model.CollectDeployment{
		ChannelID:       ch.ID,
		ChainID:         chainID,
		DslHash:         hash,
		DslSnapshot:     string(dsl),
		TemplateVersion: TemplateVersion,
		Status:          "deployed",
		OperatorID:      operatorID,
	}
	if err := db.Create(&rec).Error; err != nil {
		return false, err
	}
	global.GVA_LOG.Info("采集通道已部署", zap.Uint("channelId", ch.ID), zap.String("chainId", chainID), zap.String("hash", hash[:8]))
	ReloadChannelTrigger(ch) // 部署成功即恢复/刷新触发面
	return false, nil
}

// UndeployChannel 下线通道（链 stop + 记录 rolled_back）。
func UndeployChannel(channelID uint, operatorID uint) error {
	deployLock.Lock()
	defer deployLock.Unlock()

	db := global.GVA_DB
	var ch model.CollectChannel
	if err := db.First(&ch, channelID).Error; err != nil {
		return fmt.Errorf("通道不存在: %w", err)
	}
	chainID := ChainID(ch.ID)
	if err := BridgeClient().StopChain(chainID); err != nil {
		// 链不存在视为已下线（首次部署前下线场景）
		global.GVA_LOG.Warn("下线链返回错误（可能未部署）", zap.String("chainId", chainID), zap.Error(err))
	}
	rec := model.CollectDeployment{
		ChannelID: ch.ID, ChainID: chainID, Status: "rolled_back", OperatorID: operatorID,
	}
	if err := db.Create(&rec).Error; err != nil {
		return err
	}
	StopChannelTrigger(ch.ID) // 下线即移除触发面
	return nil
}

// RebuildAll 全量重编译（模板升级后），返回 (总数, 成功, 跳过, 失败明细)。
func RebuildAll(operatorID uint) (total, ok, skipped int, failures map[uint]string) {
	deployLock.Lock()
	defer deployLock.Unlock()
	failures = map[uint]string{}

	db := global.GVA_DB
	var channels []model.CollectChannel
	if err := db.Where("enable = ?", true).Find(&channels).Error; err != nil {
		return 0, 0, 0, map[uint]string{0: err.Error()}
	}
	total = len(channels)
	for _, ch := range channels {
		in, err := loadCompileInput(db, ch)
		if err != nil {
			failures[ch.ID] = err.Error()
			continue
		}
		dsl, hash, err := Compile(in)
		if err != nil {
			failures[ch.ID] = err.Error()
			continue
		}
		var last model.CollectDeployment
		hasLast := db.Where("channel_id = ? AND status = ?", ch.ID, "deployed").
			Order("id desc").First(&last).Error == nil
		if hasLast && last.DslHash == hash && BridgeClient().ChainExists(ChainID(ch.ID)) {
			skipped++
			continue
		}
		if err := BridgeClient().SaveChain(ChainID(ch.ID), dsl); err != nil {
			failures[ch.ID] = err.Error()
			continue
		}
		_ = db.Create(&model.CollectDeployment{
			ChannelID: ch.ID, ChainID: ChainID(ch.ID), DslHash: hash,
			DslSnapshot: string(dsl), TemplateVersion: TemplateVersion,
			Status: "deployed", OperatorID: operatorID,
		}).Error
		ok++
	}
	return total, ok, skipped, failures
}

// RemoveChannelChain 通道删除时移除 rulego 侧链。
func RemoveChannelChain(channelID uint) {
	StopChannelTrigger(channelID)
	_ = BridgeClient().DeleteChain(ChainID(channelID))
}

var _ = sha256.Sum256
var _ = hex.EncodeToString
