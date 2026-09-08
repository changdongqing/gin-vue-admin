package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/model"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// 触发面（03 文档 §4.2 修正版，P0 结论）：RuleGo 链节点注册表无周期触发组件，
// 触发收敛在 GVA 侧——poll 通道由 cron 按周期调 notify 端点驱动链；
// report 通道由本插件 MQTT 订阅者收报文调 notify。链本身是纯数据处理器。
// GVA 启动/插件注册时全量恢复；部署/下线/删除通道时同步增删对应触发器。

var (
	trigMu      sync.Mutex
	trigCron    *cron.Cron
	cronEntries = map[uint]cron.EntryID{} // channelId -> cron entry
	mqttSubs    = map[uint]mqtt.Client{}  // channelId -> mqtt client
)

// StartTriggers 启动触发面（插件 Register 调用；重启后全量恢复）。
func StartTriggers() {
	trigMu.Lock()
	defer trigMu.Unlock()
	trigCron = cron.New(cron.WithSeconds())
	trigCron.Start()

	db := global.GVA_DB
	var channels []model.CollectChannel
	if err := db.Where("enable = ?", true).Find(&channels).Error; err != nil {
		global.GVA_LOG.Error("触发面恢复失败", zap.Error(err))
		return
	}
	for _, ch := range channels {
		startChannelTriggerLocked(ch)
	}
	global.GVA_LOG.Info("采集触发面已启动", zap.Int("pollChannels", len(cronEntries)), zap.Int("reportChannels", len(mqttSubs)))
}

// ReloadChannelTrigger 通道部署/启用后刷新其触发器。
func ReloadChannelTrigger(ch model.CollectChannel) {
	trigMu.Lock()
	defer trigMu.Unlock()
	stopChannelTriggerLocked(ch.ID)
	if ch.Enable == nil || !*ch.Enable {
		return
	}
	startChannelTriggerLocked(ch)
}

// StopChannelTrigger 通道下线/删除时移除触发器。
func StopChannelTrigger(channelID uint) {
	trigMu.Lock()
	defer trigMu.Unlock()
	stopChannelTriggerLocked(channelID)
}

func stopChannelTriggerLocked(channelID uint) {
	if id, ok := cronEntries[channelID]; ok {
		trigCron.Remove(id)
		delete(cronEntries, channelID)
	}
	if cl, ok := mqttSubs[channelID]; ok {
		cl.Disconnect(500)
		delete(mqttSubs, channelID)
	}
}

func startChannelTriggerLocked(ch model.CollectChannel) {
	chainID := ChainID(ch.ID)
	switch ch.AccessMode {
	case "poll":
		period := ch.PollInterval
		if period <= 0 {
			period = 5000
		}
		sec := period / 1000
		if sec < 1 {
			sec = 1 // cron 秒级精度，周期下限 1s
		}
		expr := "*/" + strconv.Itoa(sec) + " * * * * *"
		id, err := trigCron.AddFunc(expr, func() {
			if err := BridgeClient().NotifyChain(chainID, "JSON", "{}"); err != nil {
				global.GVA_LOG.Error("采集触发失败", zap.Uint("channelId", ch.ID), zap.Error(err))
			}
		})
		if err != nil {
			global.GVA_LOG.Error("采集 cron 注册失败", zap.Uint("channelId", ch.ID), zap.String("expr", expr), zap.Error(err))
			return
		}
		cronEntries[ch.ID] = id
	case "report":
		var conn struct {
			Server   string `json:"server"`
			Username string `json:"username"`
			Password string `json:"password"`
			Topic    string `json:"topic"`
			QoS      int    `json:"qos"`
		}
		if len(ch.ConnConfig) > 0 {
			if err := json.Unmarshal(ch.ConnConfig, &conn); err != nil || conn.Server == "" || conn.Topic == "" {
				global.GVA_LOG.Error("上报通道 conn_config 缺少 server/topic", zap.Uint("channelId", ch.ID))
				return
			}
		} else {
			return
		}
		opts := mqtt.NewClientOptions().
			AddBroker(conn.Server).
			SetClientID(fmt.Sprintf("gva-collect-%d-%d", ch.ID, time.Now().UnixMilli())).
			SetAutoReconnect(true).
			SetCleanSession(true)
		if conn.Username != "" {
			opts = opts.SetUsername(conn.Username).SetPassword(conn.Password)
		}
		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			global.GVA_LOG.Error("上报通道 MQTT 连接失败", zap.Uint("channelId", ch.ID), zap.Error(token.Error()))
			return
		}
		token := client.Subscribe(conn.Topic, byte(conn.QoS), func(_ mqtt.Client, m mqtt.Message) {
			payload := string(m.Payload())
			msgType := "STRING"
			if len(payload) > 0 && (payload[0] == '{' || payload[0] == '[') {
				msgType = "JSON"
			}
			if err := BridgeClient().NotifyChain(chainID, msgType, payload); err != nil {
				global.GVA_LOG.Error("上报报文触发失败", zap.Uint("channelId", ch.ID), zap.Error(err))
			}
		})
		if token.Wait() && token.Error() != nil {
			global.GVA_LOG.Error("上报通道订阅失败", zap.Uint("channelId", ch.ID), zap.Error(token.Error()))
			return
		}
		mqttSubs[ch.ID] = client
	}
}
