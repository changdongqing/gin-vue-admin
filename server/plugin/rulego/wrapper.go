package rulego

import (
	"encoding/json"
	"net/http"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
)

// gvaEnvelope 实现 bridge.ResponseWrapper（见官方 bridge.WithResponseWrapper）：
// 把 rulego-server 的裸 JSON 响应重组为 GVA 统一信封 {code, data, msg}。
// SSE/流式/二进制/204/HEAD 响应由 bridge 直接透传、不经此函数，只需处理缓冲的 JSON。
// HTTP status 沿用上游（无 token 401 / 无权限 403 语义保留），前端同时按信封 code 判定业务结果。
func gvaEnvelope(status int, body []byte) ([]byte, int) {
	var parsed interface{}
	_ = json.Unmarshal(body, &parsed)

	code := response.SUCCESS
	msg := "操作成功"
	data := parsed
	if status >= http.StatusBadRequest {
		code = response.ERROR
		data = nil
		msg = errMessage(parsed, status)
	}

	out, err := json.Marshal(response.Response{Code: code, Data: data, Msg: msg})
	if err != nil {
		return body, status // 包装失败兜底返回原始响应
	}
	return out, status
}

// errMessage 从 rulego 错误响应体提取人读消息；rulego 错误格式为 {"error":"..."}。
func errMessage(parsed interface{}, status int) string {
	if m, ok := parsed.(map[string]interface{}); ok {
		for _, key := range []string{"error", "message", "msg"} {
			if v, ok := m[key].(string); ok && v != "" {
				return v
			}
		}
	}
	return http.StatusText(status)
}
