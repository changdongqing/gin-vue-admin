package service

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

// bridgeClient 服务端内部调用本进程 rulego bridge（HTTP 回环）。
// 认证：服务账号 JWT（888 超管，复用 GVA 签名密钥），与 03 文档 §4.6 一致。

type bridgeClient struct {
	base string
	hc   *http.Client
}

var (
	bridgeOnce    sync.Once
	bridgeIns     *bridgeClient
	serviceTokMu  sync.Mutex
	serviceTokVal string
	serviceTokExp time.Time
)

// BridgeClient 取回环 bridge 客户端（单例）。
func BridgeClient() *bridgeClient {
	bridgeOnce.Do(func() {
		base := "http://127.0.0.1:" + fmt.Sprintf("%d", global.GVA_CONFIG.System.Addr)
		bridgeIns = &bridgeClient{base: base, hc: &http.Client{Timeout: 30 * time.Second}}
	})
	return bridgeIns
}

// serviceToken 服务账号 JWT（888），到期前 5 分钟刷新，进程内缓存。
func serviceToken() (string, error) {
	serviceTokMu.Lock()
	defer serviceTokMu.Unlock()
	if serviceTokVal != "" && time.Now().Before(serviceTokExp.Add(-5*time.Minute)) {
		return serviceTokVal, nil
	}
	claims := request.CustomClaims{
		BaseClaims: request.BaseClaims{Username: "collect-service", NickName: "采集服务账号", AuthorityId: 888},
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
			Issuer:    "gva-collect",
		},
	}
	tok, err := utils.NewJWT().CreateToken(claims)
	if err != nil {
		return "", err
	}
	serviceTokVal = tok
	serviceTokExp = time.Now().Add(time.Hour)
	return tok, nil
}

func (b *bridgeClient) do(method, path, body string) (int, []byte, error) {
	tok, err := serviceToken()
	if err != nil {
		return 0, nil, fmt.Errorf("签发服务账号token失败: %w", err)
	}
	var req *http.Request
	if body != "" {
		req, err = http.NewRequest(method, b.base+path, strings.NewReader(body))
	} else {
		req, err = http.NewRequest(method, b.base+path, nil)
	}
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := b.hc.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	return resp.StatusCode, data, err
}

// SaveChain 保存（含校验+自动部署）规则链：POST /rulego/api/v1/rules/:id。
func (b *bridgeClient) SaveChain(chainID string, dsl []byte) error {
	code, body, err := b.do(http.MethodPost, "/rulego/api/v1/rules/"+chainID, string(dsl))
	if err != nil {
		return err
	}
	if code != http.StatusOK {
		return fmt.Errorf("bridge 保存链 %s 失败: status=%d body=%s", chainID, code, body)
	}
	return nil
}

// StopChain 下线规则链：POST /rulego/api/v1/rules/:id/operate/stop。
func (b *bridgeClient) StopChain(chainID string) error {
	code, body, err := b.do(http.MethodPost, "/rulego/api/v1/rules/"+chainID+"/operate/stop", "")
	if err != nil {
		return err
	}
	if code != http.StatusOK {
		return fmt.Errorf("bridge 下线链 %s 失败: status=%d body=%s", chainID, code, body)
	}
	return nil
}

// DeleteChain 删除规则链。
func (b *bridgeClient) DeleteChain(chainID string) error {
	code, body, err := b.do(http.MethodDelete, "/rulego/api/v1/rules/"+chainID, "")
	if err != nil {
		return err
	}
	if code != http.StatusOK {
		return fmt.Errorf("bridge 删除链 %s 失败: status=%d body=%s", chainID, code, body)
	}
	return nil
}

// NotifyChain 异步触发链（触发面入口）：POST /rulego/api/v1/rules/:id/notify/:msgType。
func (b *bridgeClient) NotifyChain(chainID, msgType, payload string) error {
	code, body, err := b.do(http.MethodPost, "/rulego/api/v1/rules/"+chainID+"/notify/"+msgType, payload)
	if err != nil {
		return err
	}
	if code != http.StatusOK {
		return fmt.Errorf("bridge 触发链 %s 失败: status=%d body=%s", chainID, code, body)
	}
	return nil
}

// ChainExists 探测链是否已部署（GET 单链）。
func (b *bridgeClient) ChainExists(chainID string) bool {
	code, _, err := b.do(http.MethodGet, "/rulego/api/v1/rules/"+chainID, "")
	return err == nil && code == http.StatusOK
}

// ExecuteChain 同步执行链并返回输出体（子流程回放测试用）。
func (b *bridgeClient) ExecuteChain(chainID, msgType, payload string) (string, error) {
	code, body, err := b.do(http.MethodPost, "/rulego/api/v1/rules/"+chainID+"/execute/"+msgType, payload)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK {
		return "", fmt.Errorf("bridge 执行链 %s 失败: status=%d body=%s", chainID, code, body)
	}
	return string(body), nil
}
