package service

import (
	"fmt"
	"net"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/model"
	"go.uber.org/zap"
)

// 启动对账（自愈）：三级模型是唯一事实源，规则链是编译派生物。引擎侧状态可能丢失
//（宿主升级期服务账号未注册被当孤儿目录跳过加载、数据目录损坏等），启动时以 DB 为准
// 把已发布子流程重新注册、启用通道重新编译部署。全部幂等：链在且 dsl_hash 一致则跳过。

// WhenBridgeReady 等待回环 bridge（同进程 gin）开始监听后执行 fn；超时（默认 2 分钟）
// 仅告警放弃。种子落库、启动对账等依赖 bridge 的动作统一经此调度。
func WhenBridgeReady(fn func()) {
	go func() {
		addr := fmt.Sprintf("127.0.0.1:%d", global.GVA_CONFIG.System.Addr)
		deadline := time.Now().Add(2 * time.Minute)
		for time.Now().Before(deadline) {
			c, err := net.DialTimeout("tcp", addr, time.Second)
			if err == nil {
				_ = c.Close()
				fn()
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
		global.GVA_LOG.Warn("采集平台：bridge 未就绪，跳过依赖引擎的启动动作", zap.String("addr", addr))
	}()
}

// ReconcileChainsOnStartup bridge 就绪后对账：重注册全部已发布子流程 → 强制重部署全部
// 启用通道。必须强制：store 里有链不代表引擎已加载（ChainExists 只查 store），引擎丢链
// 时只有真发 SaveChain 才能重建链实例。失败仅告警不阻塞。
func ReconcileChainsOnStartup() {
	WhenBridgeReady(func() {
		db := global.GVA_DB
		var chains []model.CollectParseChain
		if err := db.Where("status = ?", "published").Find(&chains).Error; err == nil {
			for _, pc := range chains {
				if pc.Dsl == "" {
					continue
				}
				if err := BridgeClient().SaveChain(ParseChainRuleID(pc.ID), []byte(pc.Dsl)); err != nil {
					global.GVA_LOG.Warn("采集启动对账：子流程注册失败", zap.Uint("id", pc.ID), zap.String("name", pc.Name), zap.Error(err))
				}
			}
		}
		var channels []model.CollectChannel
		if err := db.Where("enable = ?", true).Find(&channels).Error; err != nil {
			return
		}
		for _, ch := range channels {
			if err := ForceDeployChannel(ch.ID, 0); err != nil {
				global.GVA_LOG.Warn("采集启动对账：通道部署失败", zap.Uint("channelId", ch.ID), zap.String("name", ch.Name), zap.Error(err))
			}
		}
		global.GVA_LOG.Info("采集启动对账完成")
	})
}
