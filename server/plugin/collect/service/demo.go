package service

import (
	"encoding/json"
	"fmt"
)

// 采集演示种子（落库见 initialize/demo_seed.go）的场景一契约与子流程 DSL 构造。
// 场景一（report）：网关→8路模拟量采集模块→温湿度/压力/水浸三传感器，网关每 5s 向
// MQTT 发一次包含所有设备所有测点的大 JSON；平台经 conn_config.server/topic 订阅，
// 子流程按 port 映射平台设备 ID 并展开测点。
// 输出契约对齐主链模板：n_build(reportBuildScript) 接收数组，元素 {id,device,points}——
// id 必须是 collect_devices.ID（sqlScript 仅落库 id 非零设备）；points[].name 须与
// collect_variables.name 严格一致（实时表按 通道:设备:测点 定位）。

const (
	DemoChainName = "演示-模拟量网关大JSON解析"
	DemoTypeName  = "演示-模拟量网关大JSON"
	DemoReportCh  = "演示-网关模拟量上报"
	DemoModbusCh  = "演示-ModbusTCP电表"

	// DemoGatewaySample 演示网关大 JSON 样例：port 1=温湿度 2=压力 3=水浸，ts 毫秒。
	// 同时用作子流程 TestPayload（子流程库「回放测试」直接可用）。
	DemoGatewaySample = `{"gateway":"gw-analog-01","ts":1730000000000,"sensors":[` +
		`{"port":1,"temp":26.5,"hum":58.2},{"port":2,"pressure":0.432},{"port":3,"leak":false}]}`
)

// BuildDemoGatewayParseDSL 构造演示子流程 DSL：单 jsTransform 节点，
// 脚本内嵌 port→平台设备ID 映射（%d 依次为温湿度/压力/水浸设备 ID）。
func BuildDemoGatewayParseDSL(chainID, thID, ptID, lkID uint) ([]byte, error) {
	dsl := map[string]interface{}{
		"ruleChain": map[string]interface{}{
			"id": ParseChainRuleID(chainID), "name": DemoChainName, "root": true,
		},
		"metadata": map[string]interface{}{
			"firstNodeIndex": 0,
			"nodes": []interface{}{map[string]interface{}{
				"id": "parse", "type": "jsTransform", "name": "大JSON解析",
				"configuration": map[string]interface{}{"jsScript": fmt.Sprintf(demoGatewayParseScript, thID, ptID, lkID)},
			}},
			"connections": []interface{}{},
		},
	}
	return json.Marshal(dsl)
}

// demoGatewayParseScript 大 JSON → devices 数组；未知 port 跳过；缺测点字段不产出该点。
// 实测约束（P0 回归锚点 TestDemoGatewayParseChainSeed）：jsTransform 返回数组会被引擎按
// 字节数组校验（元素须 0-255 整数）而失败，故以 JSON.stringify 字符串承载、msgType 保持
// JSON——下游 n_build 的 GetDataByType 会重新解码，Array.isArray(msg) 命中数组分支。
const demoGatewayParseScript = `var MAP = {'1':{id:%d,device:'温湿度传感器'},'2':{id:%d,device:'压力传感器'},'3':{id:%d,device:'水浸传感器'}};
var ts = msg.ts || Date.now();
var out = [];
var arr = msg.sensors || [];
for (var i = 0; i < arr.length; i++) {
  var s = arr[i];
  var m = MAP[String(s.port)];
  if (!m) { continue; }
  var pts = [];
  if (s.temp !== undefined && s.temp !== null) { pts.push({name:'温度', value:s.temp, quality:'good', ts:ts}); }
  if (s.hum !== undefined && s.hum !== null) { pts.push({name:'湿度', value:s.hum, quality:'good', ts:ts}); }
  if (s.pressure !== undefined && s.pressure !== null) { pts.push({name:'压力', value:s.pressure, quality:'good', ts:ts}); }
  if (s.leak !== undefined && s.leak !== null) { pts.push({name:'水浸状态', value:s.leak ? '报警' : '正常', quality:'good', ts:ts}); }
  if (pts.length > 0) { out.push({id:m.id, device:m.device, points:pts}); }
}
return {'msg': JSON.stringify(out), 'metadata': metadata, 'msgType': 'JSON'};`
