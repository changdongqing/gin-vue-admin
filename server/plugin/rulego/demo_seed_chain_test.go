package rulego

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	collectservice "github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/service"
)

// 演示种子回归锚点：采集演示子流程 DSL（service.BuildDemoGatewayParseDSL）必须在真实
// bridge 引擎上可执行，且输出满足主链模板契约——devices 数组元素 {id(平台设备ID),device,points}，
// id 直接来自种子脚本内嵌映射（initialize/demo_seed.go 落库与 DeployChannel 依赖此契约）。
func TestDemoGatewayParseChainSeed(t *testing.T) {
	b := newTestBridge(t)
	srv := httptest.NewServer(b.Handler())
	defer srv.Close()

	dsl, err := collectservice.BuildDemoGatewayParseDSL(301, 311, 312, 313)
	if err != nil {
		t.Fatalf("构造演示子流程 DSL: %v", err)
	}
	id := "collect_pc_demo_seed"
	pocSaveChain(t, srv, id, string(dsl))
	code, out := demoExecuteJSON(t, srv, id, collectservice.DemoGatewaySample)
	if code != http.StatusOK {
		t.Fatalf("execute: status=%d body=%s", code, out)
	}

	// execute 响应经 gvaEnvelope 包装为 {code,data,msg}，data 即末节点 msg（devices 数组）
	var envelope response.Response
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("信封解析失败: %v, out=%s", err, out)
	}
	if envelope.Code != 0 {
		t.Fatalf("信封 code=%d msg=%s", envelope.Code, envelope.Msg)
	}
	var devices []struct {
		ID     int    `json:"id"`
		Device string `json:"device"`
		Points []struct {
			Name    string      `json:"name"`
			Value   interface{} `json:"value"`
			Quality string      `json:"quality"`
			Ts      int64       `json:"ts"`
		} `json:"points"`
	}
	dataBytes, _ := json.Marshal(envelope.Data)
	if err := json.Unmarshal(dataBytes, &devices); err != nil {
		t.Fatalf("devices 解析失败: %v, out=%s", err, out)
	}

	if len(devices) != 3 {
		t.Fatalf("应产出3台设备, got %d: %s", len(devices), out)
	}
	byID := map[int]map[string]interface{}{}
	for _, d := range devices {
		m := map[string]interface{}{}
		for _, p := range d.Points {
			m[p.Name] = p.Value
		}
		byID[d.ID] = m
	}
	th, ok := byID[311]
	if !ok || th["温度"].(float64) != 26.5 || th["湿度"].(float64) != 58.2 {
		t.Fatalf("温湿度设备(311)解析不符: %v", th)
	}
	pt, ok := byID[312]
	if !ok || pt["压力"].(float64) != 0.432 {
		t.Fatalf("压力设备(312)解析不符: %v", pt)
	}
	lk, ok := byID[313]
	if !ok || lk["水浸状态"] != "正常" {
		t.Fatalf("水浸设备(313)解析不符: %v", lk)
	}
}

// demoExecuteJSON 同步执行链；rest 端点按 Content-Type 判定 DataType（与生产 bridge.go
// do() 恒带 application/json 一致）——pocReq 未带头会退化为 TEXT，msg 不解码。
func demoExecuteJSON(t *testing.T, srv *httptest.Server, chainID, payload string) (int, string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, srv.URL+prefix+"/api/v1/rules/"+chainID+"/execute/JSON",
		bytes.NewBufferString(payload))
	req.Header.Set("Authorization", "Bearer "+gvaAdminToken(t))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}
