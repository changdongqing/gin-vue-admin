package rulego

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
)

// P0 PoC（aiDoc/rulego/03 §十 P0）：在真实 bridge 引擎上手写采集主链模板的核心结构，
// 实测回填设计文档的三个不确定点：
//  1. join 聚合：Fork→分支→join 合并为 JSON 数组、Success/Failure 语义（§风险 2）；
//  2. join 超时：慢分支触发超时后走 Failure 关系，模板需为 Failure 设计降级路径；
//  3. x/iotRead 节点注册与不可达 server 时的行为（§风险 1 数据面表现）。
// 结论以注释形式固化在对应测试内，并回填 03 文档。

// pocReq 通过 bridge HTTP 面以 GVA 超管身份调用 rulego REST。
func pocReq(t *testing.T, srv *httptest.Server, method, path, body string) (int, []byte) {
	t.Helper()
	var req *http.Request
	if body != "" {
		req, _ = http.NewRequest(method, srv.URL+path, bytes.NewBufferString(body))
	} else {
		req, _ = http.NewRequest(method, srv.URL+path, nil)
	}
	req.Header.Set("Authorization", "Bearer "+gvaAdminToken(t))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, b
}

// pocSaveChain 保存规则链（DSL 格式同官方 README：ruleChain + metadata）。
func pocSaveChain(t *testing.T, srv *httptest.Server, id, dsl string) {
	t.Helper()
	code, body := pocReq(t, srv, http.MethodPost, prefix+"/api/v1/rules/"+id, dsl)
	if code != http.StatusOK {
		t.Fatalf("save chain %s: status=%d body=%s", id, code, body)
	}
}

// pocExecute 同步执行规则链（POST /rules/:id/execute/:msgType），响应体即链最终输出。
func pocExecute(t *testing.T, srv *httptest.Server, id, payload string) (int, string) {
	t.Helper()
	code, body := pocReq(t, srv, http.MethodPost, prefix+"/api/v1/rules/"+id+"/execute/JSON", payload)
	return code, string(body)
}

const pocSplitScript = "return {'msg':msg,'metadata':metadata,'msgType':msgType,'dataType':dataType};"

// TestPocJoinAggregation 验证 Fork→两设备分支→join 合并为 JSON 数组→组装大 JSON 的全链路。
// 实测结论（回填 03 文档 §风险 2 / §4.4）：
//  1. join 默认（mergeToMap=false）把各分支输出合并为 JSON 数组并走 Success；
//  2. 数组元素为 WrapperMsg：{nodeId, err, msg:{data(分支输出JSON字符串), metadata, ts}}——
//     组装节点必须 JSON.parse(w.msg.data)，并可读取 w.err 作为分支级错误（quality 判定来源）。
func TestPocJoinAggregation(t *testing.T) {
	b := newTestBridge(t)
	srv := httptest.NewServer(b.Handler())
	defer srv.Close()

	// build 脚本：解析 join 的 WrapperMsg 数组，取各分支输出并附带分支级 err（模板 v1 蓝本）
	const buildScript = "var devices = msg.map(function(w){ var d = JSON.parse(w.msg.data); d.err = w.err; return d; });" +
		" return {'msg':{'channel':{'id':1,'name':'poc'},'devices':devices},'metadata':metadata,'msgType':msgType};"

	dsl := `{
	  "ruleChain": {"id": "poc_join", "name": "poc-join", "root": true},
	  "metadata": {
	    "nodes": [
	      {"id": "split", "type": "jsTransform", "name": "入口分发",
	       "configuration": {"jsScript": "` + pocSplitScript + `"}},
	      {"id": "dev1", "type": "jsTransform", "name": "设备1分支",
	       "configuration": {"jsScript": "return {'msg':{'device':'dev1','points':[{'name':'电压','value':220.1,'quality':'good'}]},'metadata':metadata,'msgType':msgType};"}},
	      {"id": "dev2", "type": "jsTransform", "name": "设备2分支",
	       "configuration": {"jsScript": "return {'msg':{'device':'dev2','points':[{'name':'电流','value':1.5,'quality':'good'}]},'metadata':metadata,'msgType':msgType};"}},
	      {"id": "join", "type": "join", "name": "设备分支聚合",
	       "configuration": {"timeout": 5, "mergeToMap": false}},
	      {"id": "build", "type": "jsTransform", "name": "组装大JSON",
	       "configuration": {"jsScript": "` + buildScript + `"}}
	    ],
	    "connections": [
	      {"fromId": "split", "toId": "dev1", "type": "Success"},
	      {"fromId": "split", "toId": "dev2", "type": "Success"},
	      {"fromId": "dev1", "toId": "join", "type": "Success"},
	      {"fromId": "dev2", "toId": "join", "type": "Success"},
	      {"fromId": "join", "toId": "build", "type": "Success"}
	    ]
	  }
	}`
	pocSaveChain(t, srv, "poc_join", dsl)

	code, out := pocExecute(t, srv, "poc_join", `{"trigger":"manual"}`)
	if code != http.StatusOK {
		t.Fatalf("execute: status=%d out=%s", code, out)
	}

	// execute 响应经 gvaEnvelope 包装为 {code,data,msg}，大 JSON 在 data 字段
	var envelope response.Response
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("信封解析失败: %v, out=%s", err, out)
	}
	if envelope.Code != response.SUCCESS {
		t.Fatalf("信封 code=%d msg=%s", envelope.Code, envelope.Msg)
	}

	var got struct {
		Channel struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"channel"`
		Devices []struct {
			Device string `json:"device"`
			Err    string `json:"err"`
			Points []struct {
				Name    string      `json:"name"`
				Value   interface{} `json:"value"`
				Quality string      `json:"quality"`
			} `json:"points"`
		} `json:"devices"`
	}
	dataBytes, _ := json.Marshal(envelope.Data)
	if err := json.Unmarshal(dataBytes, &got); err != nil {
		t.Fatalf("大 JSON 解析失败: %v, out=%s", err, out)
	}
	if got.Channel.Name != "poc" || len(got.Devices) != 2 {
		t.Fatalf("大 JSON 结构不符: %s", out)
	}
	found := map[string]string{}
	for _, d := range got.Devices {
		if len(d.Points) == 1 {
			found[d.Device] = d.Points[0].Quality
		}
	}
	if found["dev1"] != "good" || found["dev2"] != "good" {
		t.Fatalf("分支测点丢失: %v, out=%s", found, out)
	}
}

// TestPocJoinTimeout 验证慢分支触发 join 超时后走 Failure 关系（模板降级路径设计依据）。
// 实测结论（回填 03 文档 §风险 2）：join 超时通过 TellFailure 走 Failure 关系，
// 且 join(Failure) 的消息仍是 wrapperMsg（原 msg 副本）；模板据此为每个 join 配
// Failure 边 → 降级组装节点，保证大 JSON 契约不中断。
func TestPocJoinTimeout(t *testing.T) {
	b := newTestBridge(t)
	srv := httptest.NewServer(b.Handler())
	defer srv.Close()

	dsl := `{
	  "ruleChain": {"id": "poc_join_to", "name": "poc-join-timeout", "root": true},
	  "metadata": {
	    "nodes": [
	      {"id": "split", "type": "jsTransform", "name": "入口分发",
	       "configuration": {"jsScript": "` + pocSplitScript + `"}},
	      {"id": "fast", "type": "jsTransform", "name": "快分支",
	       "configuration": {"jsScript": "return {'msg':{'device':'fast','points':[]},'metadata':metadata,'msgType':msgType};"}},
	      {"id": "slow", "type": "delay", "name": "慢分支(5s)",
	       "configuration": {"delayMs": "5000"}},
	      {"id": "join", "type": "join", "name": "聚合(1s超时)",
	       "configuration": {"timeout": 1, "mergeToMap": false}},
	      {"id": "build", "type": "jsTransform", "name": "组装大JSON",
	       "configuration": {"jsScript": "return {'msg':{'channel':{'id':1},'devices':msg},'metadata':metadata,'msgType':msgType};"}},
	      {"id": "degrade", "type": "jsTransform", "name": "超时降级",
	       "configuration": {"jsScript": "return {'msg':{'channel':{'id':1},'devices':[],'error':'aggregation timeout'},'metadata':metadata,'msgType':msgType};"}}
	    ],
	      "connections": [
	      {"fromId": "split", "toId": "fast", "type": "Success"},
	      {"fromId": "split", "toId": "slow", "type": "Success"},
	      {"fromId": "fast", "toId": "join", "type": "Success"},
	      {"fromId": "slow", "toId": "join", "type": "Success"},
	      {"fromId": "join", "toId": "build", "type": "Success"},
	      {"fromId": "join", "toId": "degrade", "type": "Failure"}
	    ]
	  }
	}`
	pocSaveChain(t, srv, "poc_join_to", dsl)

	start := time.Now()
	code, out := pocExecute(t, srv, "poc_join_to", `{}`)
	elapsed := time.Since(start)
	if code != http.StatusOK {
		t.Fatalf("execute: status=%d out=%s", code, out)
	}
	// 超时应在 join timeout(1s) 附近返回，而不是等待慢分支 5s
	if elapsed > 4*time.Second {
		t.Fatalf("join 超时未生效，耗时 %v, out=%s", elapsed, out)
	}
	if !bytes.Contains([]byte(out), []byte("aggregation timeout")) {
		t.Fatalf("未走 Failure 降级路径, out=%s", out)
	}
}

// TestPocIotReadRegistered 验证 x/iotRead 节点经 components.go 引入后已在引擎注册，
// 并实测连接级失败的行为（回填 03 文档 §4.3 模板兜底设计依据）：
// 实测结论：x/iotRead 连接失败走 Failure 关系（execute 同步端点曾表现为 400 GetError），
// 模板必须为每个设备分支配置 Failure 边 → 兜底节点，输出 quality=timeout 的设备片段，
// 保证 join 等齐全部分支、大 JSON 契约不中断。
func TestPocIotReadRegistered(t *testing.T) {
	b := newTestBridge(t)
	srv := httptest.NewServer(b.Handler())
	defer srv.Close()

	dsl := `{
	  "ruleChain": {"id": "poc_iotread", "name": "poc-iotread", "root": true},
	  "metadata": {
	    "nodes": [
	      {"id": "dev", "type": "x/iotRead", "name": "不可达Modbus设备",
	       "configuration": {"driver": "modbus", "server": "tcp://127.0.0.1:1", "unitId": 1,
	         "points": [{"name": "电压", "addr": "40001", "type": "FLOAT32", "scale": 0.1}]}},
	      {"id": "fail", "type": "jsTransform", "name": "连接失败兜底",
	       "configuration": {"jsScript": "return {'msg':{'device':'poc','points':[{'name':'电压','value':null,'quality':'timeout','error':String(metadata.error || 'connect failed')}]},'metadata':metadata,'msgType':msgType};"}}
	    ],
	    "connections": [
	      {"fromId": "dev", "toId": "fail", "type": "Failure"}
	    ]
	  }
	}`
	pocSaveChain(t, srv, "poc_iotread", dsl)

	code, out := pocExecute(t, srv, "poc_iotread", `{}`)
	if code != http.StatusOK {
		t.Fatalf("execute: status=%d out=%s", code, out)
	}
	if !bytes.Contains([]byte(out), []byte("timeout")) {
		t.Fatalf("Failure 兜底路径未生效, out=%s", out)
	}
}
