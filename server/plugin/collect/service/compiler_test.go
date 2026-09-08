package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/model"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/rulego/rulego/server/app"
	"github.com/rulego/rulego/server/bootstrap"
	"github.com/rulego/rulego/server/bridge"
	"github.com/rulego/rulego/server/config"
	rgmodel "github.com/rulego/rulego/server/model"

	// x/iotRead 等采集节点注册（生产环境由 plugin/rulego/components.go 全进程注册，测试进程内需显式引入）
	_ "github.com/rulego/rulego-components-iot/external/deviceio"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"time"
)

// 编译器单测：结构断言 + 用真实 bridge 引擎验证编译产物可保存部署（fail-closed 前置校验）。

func testChannel(id uint, mode string) model.CollectChannel {
	enable := true
	return model.CollectChannel{
		GVA_MODEL:    global.GVA_MODEL{ID: id},
		Name:         "测试通道",
		AccessMode:   mode,
		Driver:       "modbus",
		ConnConfig:   []byte(`{"server":"tcp://127.0.0.1:502"}`),
		PollInterval: 5000,
		OutputConfig: []byte(`{"dbWrite":{"enable":true},"mqttPublish":{"server":"tcp://127.0.0.1:1883","topic":"gva/collect/1","qos":1}}`),
		Enable:       &enable,
	}
}

func testRegisterDevice(id, channelID uint, name string) model.CollectDevice {
	enable := true
	unit := 1
	return model.CollectDevice{
		GVA_MODEL:  global.GVA_MODEL{ID: id},
		ChannelID:  channelID,
		Name:       name,
		DeviceKind: "register",
		UnitID:     &unit,
		Enable:     &enable,
	}
}

func testVar(id, deviceID uint, name, addr string) model.CollectVariable {
	enable := true
	return model.CollectVariable{
		GVA_MODEL: global.GVA_MODEL{ID: id},
		DeviceID:  deviceID,
		Name:      name,
		Addr:      addr,
		DataType:  "FLOAT32",
		Scale:     0.1,
		RW:        "R",
		Enable:    &enable,
	}
}

func compilePoll() ([]byte, string, error) {
	return Compile(compileInput{
		Channel: testChannel(1, "poll"),
		Devices: []model.CollectDevice{
			testRegisterDevice(11, 1, "1号电表"),
			testRegisterDevice(12, 1, "2号电表"),
		},
		Variables: map[uint][]model.CollectVariable{
			11: {testVar(101, 11, "电压", "40001"), testVar(102, 11, "电流", "40003")},
			12: {testVar(103, 12, "电压", "40001")},
		},
		RealtimeDSN: "host=127.0.0.1 user=test password=test dbname=test port=5432 sslmode=disable",
	})
}

func TestCompilePollChannel(t *testing.T) {
	dsl, hash, err := compilePoll()
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if hash == "" || len(dsl) == 0 {
		t.Fatal("dsl/hash 为空")
	}

	var chain struct {
		RuleChain struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"ruleChain"`
		Metadata struct {
			Nodes []struct {
				ID            string                 `json:"id"`
				Type          string                 `json:"type"`
				Configuration map[string]interface{} `json:"configuration"`
			} `json:"nodes"`
			Connections []struct {
				FromID string `json:"fromId"`
				ToID   string `json:"toId"`
				Type   string `json:"type"`
			} `json:"connections"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(dsl, &chain); err != nil {
		t.Fatalf("dsl 解析: %v", err)
	}

	if chain.RuleChain.ID != "collect_ch_1" {
		t.Fatalf("chain id = %s", chain.RuleChain.ID)
	}

	// 节点集合：split + 2 iotRead + 2 fail + join + build + degrade + sql + db + mqtt = 10
	wantNodes := map[string]string{
		"n_split": "jsTransform", "n_dev_11": "x/iotRead", "n_dev_12": "x/iotRead",
		"n_fail_11": "jsTransform", "n_fail_12": "jsTransform",
		"n_join": "join", "n_build": "jsTransform", "n_degrade": "jsTransform",
		"n_sql": "jsTransform", "n_db": "dbClient", "n_mqtt": "mqttClient",
	}
	gotNodes := map[string]string{}
	for _, n := range chain.Metadata.Nodes {
		gotNodes[n.ID] = n.Type
	}
	if len(gotNodes) != len(wantNodes) {
		t.Fatalf("节点数 %d != %d: %v", len(gotNodes), len(wantNodes), gotNodes)
	}
	for id, typ := range wantNodes {
		if gotNodes[id] != typ {
			t.Fatalf("节点 %s type = %s, want %s", id, gotNodes[id], typ)
		}
	}

	// 关键边：iotRead 分支 Success/Failure、兜底回 join、join Failure→degrade
	edges := map[string]bool{}
	for _, e := range chain.Metadata.Connections {
		edges[e.FromID+"->"+e.ToID+":"+e.Type] = true
	}
	for _, want := range []string{
		"n_dev_11->n_join:Success", "n_dev_11->n_fail_11:Failure", "n_fail_11->n_join:Success",
		"n_dev_12->n_join:Success", "n_dev_12->n_fail_12:Failure", "n_fail_12->n_join:Success",
		"n_join->n_build:Success", "n_join->n_degrade:Failure",
		"n_build->n_sql:Success", "n_sql->n_db:Success", "n_build->n_mqtt:Success",
	} {
		if !edges[want] {
			t.Fatalf("缺少边 %s, edges=%v", want, edges)
		}
	}

	// x/iotRead 配置：driver/server/unitId/points
	for _, n := range chain.Metadata.Nodes {
		if n.ID != "n_dev_11" {
			continue
		}
		if n.Configuration["driver"] != "modbus" || n.Configuration["server"] != "tcp://127.0.0.1:502" {
			t.Fatalf("iotRead driver/server 不符: %v", n.Configuration)
		}
		if n.Configuration["unitId"] != float64(1) {
			t.Fatalf("iotRead unitId 不符: %v", n.Configuration["unitId"])
		}
		points, ok := n.Configuration["points"].([]interface{})
		if !ok || len(points) != 2 {
			t.Fatalf("points 数量不符: %v", n.Configuration["points"])
		}
		p0, _ := points[0].(map[string]interface{})
		if p0["name"] != "电压" || p0["addr"] != "40001" || p0["type"] != "FLOAT32" || p0["scale"] != 0.1 {
			t.Fatalf("points[0] 翻译不符: %v", p0)
		}
	}

	// 同输入幂等：hash 一致
	_, hash2, err := compilePoll()
	if err != nil || hash != hash2 {
		t.Fatalf("编译不幂等: %s vs %s (%v)", hash, hash2, err)
	}
}

// TestCompileEngineAccepted 用真实 bridge 引擎验证编译产物可保存（节点类型/配置被引擎接受）。
func TestCompileEngineAccepted(t *testing.T) {
	dsl, _, err := compilePoll()
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.BasePath = "/rulego"
	cfg.RequireAuth = true
	cfg.DataDir = t.TempDir()
	cfg.MCP.Enable = false
	b, err := bridge.New(
		bridge.WithAppOptions(
			app.WithConfig(&cfg),
			app.WithModules(bootstrap.DefaultModules()...),
			app.WithAuthenticator(&collectTestAuth{}),
			app.WithAuthorizer(&collectTestAuthz{}),
		),
		bridge.WithoutLocalAuth(),
		bridge.WithResponseWrapper(func(status int, body []byte) ([]byte, int) { return body, status }),
	)
	if err != nil {
		t.Fatalf("bridge.New: %v", err)
	}
	defer func() { _ = b.Stop() }()
	srv := httptest.NewServer(b.Handler())
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/rulego/api/v1/rules/collect_ch_1", bytes.NewReader(dsl))
	req.Header.Set("Authorization", "Bearer "+testServiceToken(t))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("编译产物被引擎拒绝: status=%d body=%s", resp.StatusCode, body)
	}
}

func TestCompileReportChannel(t *testing.T) {
	enable := true
	pc := model.CollectParseChain{GVA_MODEL: global.GVA_MODEL{ID: 21}, Name: "温度传感器JSON解析", Status: "published"}
	dt := model.CollectDeviceType{GVA_MODEL: global.GVA_MODEL{ID: 5}, Name: "XX温度传感器", PayloadType: "json", ParseChainID: &pc.ID, Enable: &enable}

	dsl, _, err := Compile(compileInput{
		Channel: model.CollectChannel{
			GVA_MODEL:  global.GVA_MODEL{ID: 2},
			Name:       "MQTT上报通道",
			AccessMode: "report",
			ConnConfig: []byte(`{"server":"127.0.0.1:1883","deviceTypeId":5}`),
			Enable:     &enable,
		},
		Devices:     []model.CollectDevice{{GVA_MODEL: global.GVA_MODEL{ID: 31}, ChannelID: 2, Name: "传感器A", DeviceKind: "report", DeviceTypeID: &dt.ID, Enable: &enable}},
		DeviceTypes: map[uint]model.CollectDeviceType{5: dt},
		ParseChains: map[uint]model.CollectParseChain{21: pc},
		RealtimeDSN: "host=127.0.0.1",
	})
	if err != nil {
		t.Fatalf("compile report: %v", err)
	}
	var chain struct {
		Metadata struct {
			Nodes []struct {
				ID            string `json:"id"`
				Type          string `json:"type"`
				Configuration struct {
					TargetID string `json:"targetId"`
				} `json:"configuration"`
			} `json:"nodes"`
		} `json:"metadata"`
	}
	if json.Unmarshal(dsl, &chain) != nil {
		t.Fatal("dsl 解析失败")
	}
	foundFlow := false
	for _, n := range chain.Metadata.Nodes {
		if n.Type == "flow" {
			foundFlow = true
			if n.Configuration.TargetID != "collect_pc_21" {
				t.Fatalf("flow targetId = %s", n.Configuration.TargetID)
			}
		}
	}
	if !foundFlow {
		t.Fatal("report 链缺少 flow 节点")
	}
}

func TestCompileFailClosed(t *testing.T) {
	enable := true
	// 空设备
	if _, _, err := Compile(compileInput{Channel: testChannel(1, "poll")}); err == nil {
		t.Fatal("空设备应报错")
	}
	// 报文型通道缺 deviceTypeId
	if _, _, err := Compile(compileInput{
		Channel: model.CollectChannel{GVA_MODEL: global.GVA_MODEL{ID: 2}, Name: "r", AccessMode: "report", Enable: &enable},
		Devices: []model.CollectDevice{{GVA_MODEL: global.GVA_MODEL{ID: 31}, ChannelID: 2, Name: "A", DeviceKind: "report", Enable: &enable}},
	}); err == nil {
		t.Fatal("缺 deviceTypeId 应报错")
	}
	// 子流程未发布
	pc := model.CollectParseChain{GVA_MODEL: global.GVA_MODEL{ID: 21}, Name: "未发布", Status: "draft"}
	dt := model.CollectDeviceType{GVA_MODEL: global.GVA_MODEL{ID: 5}, Name: "T", PayloadType: "json", ParseChainID: &pc.ID, Enable: &enable}
	if _, _, err := Compile(compileInput{
		Channel: model.CollectChannel{
			GVA_MODEL: global.GVA_MODEL{ID: 2}, Name: "r2", AccessMode: "report",
			ConnConfig: []byte(`{"deviceTypeId":5}`), Enable: &enable,
		},
		Devices:     []model.CollectDevice{{GVA_MODEL: global.GVA_MODEL{ID: 31}, ChannelID: 2, Name: "A", DeviceKind: "report", Enable: &enable}},
		DeviceTypes: map[uint]model.CollectDeviceType{5: dt},
		ParseChains: map[uint]model.CollectParseChain{21: pc},
	}); err == nil {
		t.Fatal("子流程未发布应报错")
	}
	// 通道未启用
	disabled := testChannel(1, "poll")
	disabled.Enable = &[]bool{false}[0]
	if _, _, err := Compile(compileInput{Channel: disabled, Devices: []model.CollectDevice{testRegisterDevice(11, 1, "d")}}); err == nil {
		t.Fatal("通道未启用应报错")
	}
}

// ---- 测试专用 GVA 身份桥接（与 plugin/rulego 生产实现同契约） ----

type collectTestAuth struct{}

func (collectTestAuth) Authenticate(authorization string) (*rgmodel.UserContext, error) {
	tokenStr := authorization
	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}
	claims, err := utils.NewJWT().ParseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	return &rgmodel.UserContext{Username: claims.Username, Roles: []string{"888"}}, nil
}

type collectTestAuthz struct{}

func (collectTestAuthz) Authorize(user *rgmodel.UserContext, resource, action string) error {
	if user != nil && len(user.Roles) > 0 && user.Roles[0] == "888" {
		return nil
	}
	return fmt.Errorf("forbidden")
}

// testServiceToken 供引擎保存测试使用（888 超管）。
func testServiceToken(t *testing.T) string {
	t.Helper()
	global.GVA_CONFIG.JWT.SigningKey = "unit-test-signing-key"
	claims := request.CustomClaims{
		BaseClaims: request.BaseClaims{Username: "collect-service", AuthorityId: 888},
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok, err := utils.NewJWT().CreateToken(claims)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}
