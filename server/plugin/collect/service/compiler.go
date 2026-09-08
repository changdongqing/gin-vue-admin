package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/model"
)

// 编译器：三级模型 → RuleGo 规则链 DSL（纯函数，无 IO）。
// 模板 v1 形态依据 P0 PoC 实测（aiDoc/rulego/03 §十 P0，server/plugin/rulego/poc_test.go）：
//  1. DSL 边字段为 connections/fromId/toId/type；
//  2. join(mergeToMap=false) 合并各分支输出为 JSON 数组，元素为 WrapperMsg
//     {nodeId, err, msg:{data(分支输出JSON字符串), metadata, ts}}；
//  3. join 超时走 Failure 关系；x/iotRead 连接级失败走 Failure 关系；
//  4. 链节点注册表无周期触发组件（endpoint/schedule 仅存在于端点注册表）——
//     触发面收敛在 GVA 侧：collect 插件按通道周期调 notify 端点驱动链（§十 P0 结论）。
//  5. execute/notify 的响应经 gvaEnvelope 包装。

// TemplateVersion 编译模板版本（模板结构演进时递增，全量重编译用）。
// v2 修正两处实测 bug（2026-09 端到端联调）：
//  1. ts 列类型：collect_realtime.ts 为 bigint(unix ms)，v1 误生成 to_timestamp()/now()
//     （timestamptz），所有落库 INSERT 被 PgSQL 拒绝；且 x/iotRead 点位 timestamp 为
//     纳秒，需归一化为 ms（sqlScript 统一处理）；
//  2. flow 节点输出为 WrapperMsg 数组（{nodeId,msg:{data},err}，与 join 同形态），
//     v1 reportBuildScript 直接当设备数组用导致上报数据全部丢弃，v2 先解包再聚合。
const TemplateVersion = "2"

// ChainID 通道主链在 rulego 侧的 ID。
func ChainID(channelID uint) string { return fmt.Sprintf("collect_ch_%d", channelID) }

// ParseChainID 设备类型子流程链在 rulego 侧的 ID。
func ParseChainRuleID(parseChainID uint) string { return fmt.Sprintf("collect_pc_%d", parseChainID) }

// compileInput 编译输入（调用方保证 Devices/Variables 已按 enable 过滤）。
type compileInput struct {
	Channel     model.CollectChannel
	Devices     []model.CollectDevice
	Variables   map[uint][]model.CollectVariable // deviceId -> 测点（使能）
	DeviceTypes map[uint]model.CollectDeviceType // 报文型设备引用
	ParseChains map[uint]model.CollectParseChain // 设备类型绑定的子流程
	RealtimeDSN string                           // collect_realtime 所在库 DSN（dbClient 用）
}

// dslNode/dslConnection/dslChain 规则链 DSL 结构（对齐 types.RuleChain 实测 json tag）。
type dslNode struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Name          string                 `json:"name,omitempty"`
	DebugMode     bool                   `json:"debugMode,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	Routers       []dslRouter            `json:"routers,omitempty"`
}

type dslRouter struct {
	From struct {
		Path string `json:"path"`
	} `json:"from"`
	To struct {
		Path string `json:"path"`
	} `json:"to"`
}

type dslConnection struct {
	FromID string `json:"fromId"`
	ToID   string `json:"toId"`
	Type   string `json:"type"`
}

type dslChain struct {
	RuleChain struct {
		ID             string                 `json:"id"`
		Name           string                 `json:"name"`
		Root           bool                   `json:"root"`
		AdditionalInfo map[string]interface{} `json:"additionalInfo,omitempty"`
	} `json:"ruleChain"`
	Metadata struct {
		FirstNodeIndex int             `json:"firstNodeIndex"`
		Nodes          []dslNode       `json:"nodes"`
		Connections    []dslConnection `json:"connections"`
	} `json:"metadata"`
}

func (c *dslChain) addNode(n dslNode) string {
	c.Metadata.Nodes = append(c.Metadata.Nodes, n)
	return n.ID
}

func (c *dslChain) link(from, to, rel string) {
	c.Metadata.Connections = append(c.Metadata.Connections, dslConnection{FromID: from, ToID: to, Type: rel})
}

// Compile 编译通道主链。规则：
//   - poll 通道：root 直通 → 每设备一个 x/iotRead 分支（Failure→兜底）→ join → build → 落库/发布；
//   - report 通道：归一 → flow(子流程) → build → 落库/发布（一期一通道绑一设备类型，conn_config.deviceTypeId）；
//   - fail-closed：未知 driver / 缺子流程 / 空设备及测点 → 返回错误不产出 DSL。
func Compile(in compileInput) (dsl []byte, hash string, err error) {
	if in.Channel.Enable != nil && !*in.Channel.Enable {
		return nil, "", fmt.Errorf("通道 %s 未启用", in.Channel.Name)
	}
	if len(in.Devices) == 0 {
		return nil, "", fmt.Errorf("通道 %s 无启用设备", in.Channel.Name)
	}

	c := &dslChain{}
	c.RuleChain.ID = ChainID(in.Channel.ID)
	c.RuleChain.Name = "采集主链-" + in.Channel.Name
	c.RuleChain.Root = true
	c.RuleChain.AdditionalInfo = map[string]interface{}{
		"kind":            "collect-main",
		"templateVersion": TemplateVersion,
	}

	var outNodes []string // 大 JSON 之后的数据面节点（落库/发布），degrade 与 build 共同汇入

	if in.Channel.AccessMode == "poll" {
		outNodes, err = compilePollBranches(c, in)
	} else {
		outNodes, err = compileReportBranches(c, in)
	}
	if err != nil {
		return nil, "", err
	}

	// 输出段：build 与 degrade 汇入
	output := in.Channel.OutputConfig
	mqttCfg, _ := jsonStr(output, "mqttPublish")
	dbCfg, _ := jsonStr(output, "dbWrite")
	mqttOn := mqttCfg != ""

	dbOn := true
	if dbCfg != "" {
		var m map[string]interface{}
		if json.Unmarshal([]byte(dbCfg), &m) == nil {
			if v, ok := m["enable"].(bool); ok {
				dbOn = v
			}
		}
	}
	if dbOn {
		nSQL := c.addNode(dslNode{ID: "n_sql", Type: "jsTransform", Name: "展开UPSERT",
			Configuration: map[string]interface{}{"jsScript": sqlScript()}})
		dsn := in.RealtimeDSN
		if dsn == "" {
			return nil, "", fmt.Errorf("实时库 DSN 为空，无法编译落库节点")
		}
		nDB := c.addNode(dslNode{ID: "n_db", Type: "dbClient", Name: "实时值落库",
			Configuration: map[string]interface{}{
				"driverName": "postgres", "dsn": dsn,
				"sql": "${msg.upsertSql}", "opType": "INSERT",
			}})
		for _, src := range outNodes {
			c.link(src, nSQL, "Success")
		}
		c.link(nSQL, nDB, "Success")
		// 落库支路终结于 n_db；发布支路仍从原汇点出发（与落库并行，互不阻塞）
		if !mqttOn {
			outNodes = []string{nDB}
		}
	}
	if mqttCfg != "" {
		var m map[string]interface{}
		if json.Unmarshal([]byte(mqttCfg), &m) != nil {
			return nil, "", fmt.Errorf("通道 %s 的 mqttPublish 配置非法", in.Channel.Name)
		}
		topic, _ := m["topic"].(string)
		if topic == "" {
			return nil, "", fmt.Errorf("通道 %s 的 mqttPublish.topic 为空", in.Channel.Name)
		}
		// topic 附加大 JSON 输出（调用方可用 ${msg.xxx} 组合，这里按原样透传）
		cfg := map[string]interface{}{"server": m["server"], "topic": topic, "qos": m["qos"]}
		if v, ok := m["username"].(string); ok && v != "" {
			cfg["username"] = v
		}
		if v, ok := m["password"].(string); ok && v != "" {
			cfg["password"] = v
		}
		nMQTT := c.addNode(dslNode{ID: "n_mqtt", Type: "mqttClient", Name: "大JSON发布",
			Configuration: cfg})
		for _, src := range outNodes {
			c.link(src, nMQTT, "Success")
		}
	}
	if len(outNodes) == 0 {
		return nil, "", fmt.Errorf("通道 %s 未启用任何输出（dbWrite/mqttPublish）", in.Channel.Name)
	}

	dsl, err = json.Marshal(c)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(dsl)
	return dsl, hex.EncodeToString(sum[:]), nil
}

// compilePollBranches 寄存器型分支：x/iotRead × N → join → build。
func compilePollBranches(c *dslChain, in compileInput) (outNodes []string, err error) {
	// root 直通分发节点（fork 起点）
	c.addNode(dslNode{ID: "n_split", Type: "jsTransform", Name: "触发分发",
		Configuration: map[string]interface{}{"jsScript": passthroughScript}})

	// 设备名映射表（build 脚本内嵌：nodeId → {id, name}）
	names := map[string]map[string]interface{}{}
	joinID := ""
	// 聚合窗口：3 倍通道周期（秒），下限 10s 上限 60s
	aggTimeout := 3 * in.Channel.PollInterval / 1000
	if aggTimeout < 10 {
		aggTimeout = 10
	}
	if aggTimeout > 60 {
		aggTimeout = 60
	}

	for _, dev := range in.Devices {
		if dev.DeviceKind != "register" {
			continue
		}
		vars := in.Variables[dev.ID]
		if len(vars) == 0 {
			continue
		}
		cfg, err := iotReadConfig(in.Channel, dev, vars)
		if err != nil {
			return nil, err
		}
		nodeID := "n_dev_" + strconv.FormatUint(uint64(dev.ID), 10)
		c.addNode(dslNode{ID: nodeID, Type: "x/iotRead", Name: "设备-" + dev.Name, Configuration: cfg})
		names[nodeID] = map[string]interface{}{"id": dev.ID, "name": dev.Name}

		// 连接级失败兜底：Failure → 兜底节点 → join（保证 join 等齐全部分支）
		failID := "n_fail_" + strconv.FormatUint(uint64(dev.ID), 10)
		c.addNode(dslNode{ID: failID, Type: "jsTransform", Name: "失败兜底-" + dev.Name,
			Configuration: map[string]interface{}{"jsScript": failScript(dev.ID, dev.Name)}})
		c.link("n_split", nodeID, "Success")
		c.link(nodeID, joinTarget(c, aggTimeout, &joinID), "Success")
		c.link(nodeID, failID, "Failure")
		c.link(failID, joinID, "Success")
	}
	if joinID == "" {
		return nil, fmt.Errorf("通道 %s 无有效寄存器型设备（缺少使能测点）", in.Channel.Name)
	}

	// build：WrapperMsg 数组 → 大 JSON
	buildID := c.addNode(dslNode{ID: "n_build", Type: "jsTransform", Name: "组装大JSON",
		Configuration: map[string]interface{}{"jsScript": pollBuildScript(names, in.Channel.ID, in.Channel.Name)}})
	c.link(joinID, buildID, "Success")

	// join 超时降级：Failure → degrade → 输出
	degradeID := c.addNode(dslNode{ID: "n_degrade", Type: "jsTransform", Name: "聚合超时降级",
		Configuration: map[string]interface{}{"jsScript": degradeScript(in.Channel.ID, in.Channel.Name)}})
	c.link(joinID, degradeID, "Failure")

	return []string{buildID, degradeID}, nil
}

// joinTarget 取或创建 join 节点，返回 join 节点 ID。
func joinTarget(c *dslChain, timeout int, joinID *string) string {
	if *joinID == "" {
		*joinID = c.addNode(dslNode{ID: "n_join", Type: "join", Name: "设备分支聚合",
			Configuration: map[string]interface{}{"timeout": timeout, "mergeToMap": false}})
	}
	return *joinID
}

// compileReportBranches 报文型分支：归一 → flow(子流程) → build（一期一通道一设备类型）。
func compileReportBranches(c *dslChain, in compileInput) (outNodes []string, err error) {
	// conn_config.deviceTypeId 决定绑定的设备类型与子流程
	typeID := 0
	var conn map[string]interface{}
	if len(in.Channel.ConnConfig) > 0 {
		if json.Unmarshal(in.Channel.ConnConfig, &conn) != nil {
			return nil, fmt.Errorf("通道 %s conn_config 非法 JSON", in.Channel.Name)
		}
	}
	if v, ok := conn["deviceTypeId"].(float64); ok {
		typeID = int(v)
	}
	if typeID == 0 {
		return nil, fmt.Errorf("报文型通道 %s 未在 conn_config 指定 deviceTypeId", in.Channel.Name)
	}
	dt, ok := in.DeviceTypes[uint(typeID)]
	if !ok {
		return nil, fmt.Errorf("通道 %s 引用的设备类型 %d 不存在", in.Channel.Name, typeID)
	}
	if dt.PayloadType != "string" && dt.PayloadType != "json" && dt.PayloadType != "bytes" {
		return nil, fmt.Errorf("设备类型 %s 的 payloadType 非法: %s", dt.Name, dt.PayloadType)
	}
	pc, ok := in.ParseChains[uint(derefUint(dt.ParseChainID))]
	if !ok {
		return nil, fmt.Errorf("设备类型 %s 未绑定已存在的子流程", dt.Name)
	}
	if pc.Status != "published" {
		return nil, fmt.Errorf("子流程 %s 未发布，通道 %s 无法编译", pc.Name, in.Channel.Name)
	}

	c.addNode(dslNode{ID: "n_split", Type: "jsTransform", Name: "报文入口",
		Configuration: map[string]interface{}{"jsScript": passthroughScript}})
	nNorm := c.addNode(dslNode{ID: "n_norm", Type: "jsTransform", Name: "类型归一(" + dt.PayloadType + ")",
		Configuration: map[string]interface{}{"jsScript": reportNormScript(dt.PayloadType)}})
	nFlow := c.addNode(dslNode{ID: "n_flow", Type: "flow", Name: "子流程-" + pc.Name,
		Configuration: map[string]interface{}{"targetId": ParseChainRuleID(pc.ID), "extend": false}})
	nBuild := c.addNode(dslNode{ID: "n_build", Type: "jsTransform", Name: "组装大JSON",
		Configuration: map[string]interface{}{"jsScript": reportBuildScript(in.Channel.ID, in.Channel.Name)}})

	c.link("n_split", nNorm, "Success")
	c.link(nNorm, nFlow, "Success")
	c.link(nFlow, nBuild, "Success")
	return []string{nBuild}, nil
}

// iotReadConfig 组装 x/iotRead 节点配置：conn_config 展开 + unitId + points。
func iotReadConfig(ch model.CollectChannel, dev model.CollectDevice, vars []model.CollectVariable) (map[string]interface{}, error) {
	cfg := map[string]interface{}{}
	if len(ch.ConnConfig) > 0 {
		if err := json.Unmarshal(ch.ConnConfig, &cfg); err != nil {
			return nil, fmt.Errorf("通道 %s conn_config 非法 JSON: %w", ch.Name, err)
		}
	}
	if ch.Driver == "" {
		return nil, fmt.Errorf("通道 %s 未指定 driver", ch.Name)
	}
	cfg["driver"] = ch.Driver
	if dev.UnitID != nil {
		cfg["unitId"] = *dev.UnitID
	}
	points := make([]map[string]interface{}, 0, len(vars))
	for _, v := range vars {
		if v.RW != "" && v.RW != "R" && v.RW != "RW" {
			return nil, fmt.Errorf("测点 %s 读写标识非法: %s", v.Name, v.RW)
		}
		p := map[string]interface{}{"name": v.Name, "addr": v.Addr, "type": v.DataType}
		if v.Scale != 0 {
			p["scale"] = v.Scale
		}
		if v.Offset != 0 {
			p["offset"] = v.Offset
		}
		if v.Endian != "" {
			p["endian"] = v.Endian
		}
		points = append(points, p)
	}
	cfg["points"] = points
	return cfg, nil
}

func derefUint(p *uint) uint {
	if p == nil {
		return 0
	}
	return *p
}

// jsonStr 从 JSON 对象文本中取顶层字符串字段原文。
func jsonStr(raw []byte, key string) (string, bool) {
	if len(raw) == 0 {
		return "", false
	}
	var m map[string]json.RawMessage
	if json.Unmarshal(raw, &m) != nil {
		return "", false
	}
	v, ok := m[key]
	if !ok {
		return "", false
	}
	return string(v), true
}

// ---------- 模板内置 JS 脚本（v1，随模板版本演进） ----------

// passthroughScript 直通脚本。
const passthroughScript = "return {'msg':msg,'metadata':metadata,'msgType':msgType,'dataType':dataType};"

// failScript 设备分支连接级失败兜底：输出该设备的 timeout 片段（join 等齐分支用）。
func failScript(devID uint, devName string) string {
	return "return {'msg':{'id':" + strconv.FormatUint(uint64(devID), 10) +
		",'device':'" + jsEscape(devName) + "','points':[],'quality':'timeout'}," +
		"'metadata':metadata,'msgType':msgType};"
}

// pollBuildScript 寄存器型组装脚本：解析 join 的 WrapperMsg 数组（P0 实测形态），
// 按 nodeId 映射设备，点位 error → quality=bad，分支 err → quality=timeout。
func pollBuildScript(names map[string]map[string]interface{}, chID uint, chName string) string {
	namesJSON, _ := json.Marshal(names)
	var b strings.Builder
	b.WriteString("var NAMES = " + string(namesJSON) + ";\n")
	b.WriteString("var devices = [];\n")
	b.WriteString("for (var i = 0; i < msg.length; i++) {\n")
	b.WriteString("  var w = msg[i];\n")
	b.WriteString("  var meta = NAMES[w.nodeId] || {id:0, name:w.nodeId};\n")
	b.WriteString("  var dev = {id: meta.id, device: meta.name, points: [], err: w.err || ''};\n")
	b.WriteString("  if (!w.err && w.msg && w.msg.data) {\n")
	b.WriteString("    var arr = JSON.parse(w.msg.data);\n")
	b.WriteString("    for (var j = 0; j < arr.length; j++) {\n")
	b.WriteString("      var p = arr[j];\n")
	b.WriteString("      var pt = {name: p.name, value: (p.error ? null : p.value), quality: (p.error ? 'bad' : 'good'), ts: p.timestamp || null};\n")
	b.WriteString("      if (p.error) { pt.error = p.error; }\n")
	b.WriteString("      dev.points.push(pt);\n")
	b.WriteString("    }\n")
	b.WriteString("  } else if (w.err) { dev.quality = 'timeout'; }\n")
	b.WriteString("  devices.push(dev);\n")
	b.WriteString("}\n")
	b.WriteString("return {'msg':{'channel':{'id':" + strconv.FormatUint(uint64(chID), 10) +
		",'name':'" + jsEscape(chName) + "','ts':Date.now()},'devices':devices},'metadata':metadata,'msgType':'JSON'};")
	return b.String()
}

// degradeScript 聚合超时降级：输出空 devices 的大 JSON（契约不中断）。
func degradeScript(chID uint, chName string) string {
	return "return {'msg':{'channel':{'id':" + strconv.FormatUint(uint64(chID), 10) +
		",'name':'" + jsEscape(chName) + "','ts':Date.now()},'devices':[],'error':'aggregation timeout'},'metadata':metadata,'msgType':'JSON'};"
}

// reportNormScript 报文归一（一期直通占位：类型校验由子流程负责，节点保留便于后续插桩）。
func reportNormScript(payloadType string) string {
	_ = payloadType
	return passthroughScript
}

// reportBuildScript 报文型组装：解包 flow 的 WrapperMsg（{nodeId,msg:{data},err}），
// 取子流程输出的设备数组（[{id,device,points}]，亦兼容单对象/直出形态）→ 大 JSON。
func reportBuildScript(chID uint, chName string) string {
	var b strings.Builder
	b.WriteString("var devices = [];\n")
	b.WriteString("var arr = Array.isArray(msg) ? msg : [msg];\n")
	b.WriteString("for (var i = 0; i < arr.length; i++) {\n")
	b.WriteString("  var w = arr[i];\n")
	b.WriteString("  if (w.nodeId !== undefined) {\n")
	b.WriteString("    if (w.err) { devices.push({id:0, device:String(w.nodeId), points:[], err:String(w.err), quality:'timeout'}); continue; }\n")
	b.WriteString("    if (!w.msg || !w.msg.data) { continue; }\n")
	b.WriteString("    var inner = JSON.parse(w.msg.data);\n")
	b.WriteString("    if (!Array.isArray(inner)) { inner = [inner]; }\n")
	b.WriteString("    for (var j = 0; j < inner.length; j++) { devices.push(inner[j]); }\n")
	b.WriteString("  } else { devices.push(w); }\n")
	b.WriteString("}\n")
	b.WriteString("return {'msg':{'channel':{'id':" + strconv.FormatUint(uint64(chID), 10) +
		",'name':'" + jsEscape(chName) + "','ts':Date.now()},'devices':devices},'metadata':metadata,'msgType':'JSON'};")
	return b.String()
}

// sqlScript 大 JSON → 多行 UPSERT SQL（collect_realtime）。值统一字符串化（text 列）。
// ts 列为 bigint(unix ms)：点位时间戳在此统一归一化（x/iotRead 为纳秒、网关报文为毫秒，
// 兜底 Date.now()），不得用 to_timestamp()/now()（timestamptz，类型不符）。
func sqlScript() string {
	var b strings.Builder
	b.WriteString("function esc(s){return String(s).replace(/'/g,\"''\");}\n")
	b.WriteString("var ch = msg.channel; var rows = [];\n")
	b.WriteString("for (var i = 0; i < msg.devices.length; i++) {\n")
	b.WriteString("  var d = msg.devices[i];\n")
	b.WriteString("  if (d.id === undefined || d.id === 0) { continue; }\n")
	b.WriteString("  var pts = d.points || [];\n")
	b.WriteString("  for (var j = 0; j < pts.length; j++) {\n")
	b.WriteString("    var p = pts[j];\n")
	b.WriteString("    var pk = ch.id + ':' + d.id + ':' + p.name;\n")
	b.WriteString("    var t = p.ts ? Math.floor(p.ts) : 0;\n")
	b.WriteString("    if (t > 1e17) { t = Math.floor(t / 1e6); } else if (t > 1e14) { t = Math.floor(t / 1e3); }\n")
	b.WriteString("    if (!t) { t = Date.now(); }\n")
	b.WriteString("    rows.push(\"('\" + esc(pk) + \"',\" + ch.id + \",\" + d.id + \",'\" + esc(p.name) + \"','\" + esc(p.value === null ? '' : p.value) + \"','\" + (p.quality || 'good') + \"','\" + esc(p.error || '') + \"',\" + t + \")\");\n")
	b.WriteString("  }\n")
	b.WriteString("}\n")
	b.WriteString("if (rows.length === 0) { return {'msg':msg,'metadata':metadata,'msgType':msgType}; }\n")
	b.WriteString("var sql = \"INSERT INTO collect_realtime (point_key, channel_id, device_id, name, value, quality, error_msg, ts) VALUES \" + rows.join(',') + \" ON CONFLICT (point_key) DO UPDATE SET value = EXCLUDED.value, quality = EXCLUDED.quality, error_msg = EXCLUDED.error_msg, ts = EXCLUDED.ts\";\n")
	b.WriteString("msg.upsertSql = sql;\n")
	b.WriteString("return {'msg':msg,'metadata':metadata,'msgType':msgType};\n")
	return b.String()
}

// jsEscape JS 字符串字面量内的单引号转义（脚本用单引号字符串）。
func jsEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	return strings.ReplaceAll(s, "'", "\\'")
}
