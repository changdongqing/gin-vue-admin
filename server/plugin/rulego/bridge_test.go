package rulego

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/rulego/rulego/server/app"
	"github.com/rulego/rulego/server/bootstrap"
	"github.com/rulego/rulego/server/bridge"
	"github.com/rulego/rulego/server/config"
)

// newTestBridge 构造与生产 Register 相同形态的 bridge：GVA JWT/Casbin 桥接 +
// WithoutLocalAuth + require_auth=true 安全基线；filestore 数据目录指向测试临时目录，
// MCP 关闭以减少测试面。
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
			app.WithAuthenticator(&gvaAuthenticator{}),
			app.WithAuthorizer(&gvaAuthorizer{}),
		),
		bridge.WithoutLocalAuth(),
		bridge.WithResponseWrapper(gvaEnvelope),
	)
	if err != nil {
		t.Fatalf("bridge.New: %v", err)
	}
	t.Cleanup(func() { _ = b.Stop() })
	return b
}

// gvaAdminToken 用测试签名密钥签发 GVA 超管（888）JWT，模拟前端 x-token 的载荷。
func gvaAdminToken(t *testing.T) string {
	t.Helper()
	global.GVA_CONFIG.JWT.SigningKey = "unit-test-signing-key"
	claims := request.CustomClaims{
		BaseClaims: request.BaseClaims{Username: "admin", NickName: "测试管理员", AuthorityId: 888},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Issuer:    "unit-test",
		},
	}
	token, err := utils.NewJWT().CreateToken(claims)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	return token
}

// M2/M3 冒烟：require_auth=true 且无凭据访问规则链列表 → 上游 401，响应体已被
// ResponseWrapper 重组为 GVA 信封 {code:7, data:null, msg}。
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

// M3 验收：GVA 超管 token（经 Authorization 头，对应 mapXToken 映射后的形态）
// 访问规则链列表 → 200 + 信封 code=0。
func TestBridgeGVAAdminToken(t *testing.T) {
	b := newTestBridge(t)
	srv := httptest.NewServer(b.Handler())
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+prefix+"/api/v1/rules", nil)
	req.Header.Set("Authorization", "Bearer "+gvaAdminToken(t))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET with gva token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 200, body=%s", resp.StatusCode, body)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var got response.Response
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("body 不是合法信封 JSON: %v, body=%s", err, body)
	}
	if got.Code != response.SUCCESS {
		t.Fatalf("code = %d, want %d, msg=%s", got.Code, response.SUCCESS, got.Msg)
	}
	if got.Data == nil {
		t.Fatal("data 为空，期望携带规则链列表数据")
	}
}

// M3 验收：WithoutLocalAuth 后 rulego 自带登录端点已关闭。
func TestBridgeLocalAuthDisabled(t *testing.T) {
	b := newTestBridge(t)
	srv := httptest.NewServer(b.Handler())
	defer srv.Close()

	resp, err := http.Post(srv.URL+prefix+"/login", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /rulego/login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("本地登录端点应已关闭，但返回 200: body=%s", body)
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
