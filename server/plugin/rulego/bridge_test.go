package rulego

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
	"github.com/rulego/rulego/server/app"
	"github.com/rulego/rulego/server/bootstrap"
	"github.com/rulego/rulego/server/bridge"
	"github.com/rulego/rulego/server/config"
)

// newTestBridge 构造与生产 Register 相同形态的 bridge（require_auth=true 安全基线，
// filestore 数据目录指向测试临时目录，MCP 关闭以减少测试面）。
func newTestBridge(t *testing.T) *bridge.Bridge {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.BasePath = prefix
	cfg.RequireAuth = true
	cfg.DataDir = t.TempDir()
	cfg.MCP.Enable = false
	b, err := bridge.New(
		bridge.WithAppOptions(
			app.WithConfig(&cfg),
			app.WithModules(bootstrap.DefaultModules()...),
		),
		bridge.WithResponseWrapper(gvaEnvelope),
	)
	if err != nil {
		t.Fatalf("bridge.New: %v", err)
	}
	t.Cleanup(func() { _ = b.Stop() })
	return b
}

// M2 冒烟：require_auth=true 下未携带凭据访问规则链列表 → 上游 401，且响应体
// 已被 ResponseWrapper 重组为 GVA 信封 {code:7, data:null, msg}。
func TestBridgeUnauthorizedEnvelope(t *testing.T) {
	b := newTestBridge(t)
	srv := httptest.NewServer(b.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + prefix + "/api/v1/rules")
	if err != nil {
		t.Fatalf("GET /rulego/api/v1/rules: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var got response.Response
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("body 不是合法信封 JSON: %v, body=%s", err, body)
	}
	if got.Code != response.ERROR {
		t.Fatalf("code = %d, want %d", got.Code, response.ERROR)
	}
	if got.Msg == "" {
		t.Fatal("msg 为空，期望携带上游错误信息（如 unauthorized）")
	}
	if got.Data != nil {
		t.Fatalf("data = %v, want nil", got.Data)
	}
}

// mapXToken：Authorization 缺失时从 x-token 补映射；已有 Authorization 时不覆盖。
func TestMapXToken(t *testing.T) {
	run := func(req *http.Request) *http.Request {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = req
		mapXToken()(c)
		return req
	}

	t.Run("x-token补映射", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, prefix+"/api/v1/rules", nil)
		req.Header.Set("x-token", "abc123")
		if got := run(req).Header.Get("Authorization"); got != "Bearer abc123" {
			t.Fatalf("Authorization = %q, want %q", got, "Bearer abc123")
		}
	})

	t.Run("已有Authorization不覆盖", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, prefix+"/api/v1/rules", nil)
		req.Header.Set("Authorization", "Bearer origin")
		req.Header.Set("x-token", "other")
		if got := run(req).Header.Get("Authorization"); got != "Bearer origin" {
			t.Fatalf("Authorization = %q, want %q", got, "Bearer origin")
		}
	})

	t.Run("无token不动作", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, prefix+"/api/v1/rules", nil)
		if got := run(req).Header.Get("Authorization"); got != "" {
			t.Fatalf("Authorization = %q, want empty", got)
		}
	})
}

// gvaEnvelope：成功响应包装为 code=0 并保留上游数据；错误响应包装为 code=7、
// msg 取上游 error 字段、data 置空。
func TestGvaEnvelope(t *testing.T) {
	t.Run("成功响应", func(t *testing.T) {
		out, status := gvaEnvelope(http.StatusOK, []byte(`{"rules":[1,2],"total":2}`))
		if status != http.StatusOK {
			t.Fatalf("status = %d, want 200", status)
		}
		var got response.Response
		if err := json.Unmarshal(out, &got); err != nil {
			t.Fatalf("unmarshal: %v, out=%s", err, out)
		}
		if got.Code != response.SUCCESS || got.Msg != "操作成功" || got.Data == nil {
			t.Fatalf("got %+v, want code=0 data非空", got)
		}
	})

	t.Run("错误响应", func(t *testing.T) {
		out, status := gvaEnvelope(http.StatusUnauthorized, []byte(`{"error":"unauthorized"}`))
		if status != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", status)
		}
		var got response.Response
		if err := json.Unmarshal(out, &got); err != nil {
			t.Fatalf("unmarshal: %v, out=%s", err, out)
		}
		if got.Code != response.ERROR || got.Data != nil || got.Msg != "unauthorized" {
			t.Fatalf("got %+v, want code=7 msg=unauthorized data=nil", got)
		}
	})
}
