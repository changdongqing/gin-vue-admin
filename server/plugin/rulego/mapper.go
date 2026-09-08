package rulego

// rulego 的鉴权 SPI 按语义标签 (resource, action) 判权，而 GVA Casbin 策略粒度是
// (path, method)（middleware/casbin_rbac.go）。本文件维护二者之间的固定映射表：把
// 上游标签翻译为可录入 GVA「API 管理」的虚拟路径，再交由 Casbin 判权。设计依据与
// 枚举实测见 aiDoc/rulego/01-集成方案分析.md §4.6。
//
// 上游升级时若 resource/action 枚举变化（升级 SOP 已列入检查项），须同步本表与
// mapper_test.go；映射表外的一律拒绝（fail-closed）。

// gvaResources rulego v0.37.2 全部 resource 枚举（internal/constants + 各 endpoint 实测）。
// user 端点在 WithoutLocalAuth 下已整体关闭，仍保留映射作防御：上游若放开 user 资源，
// 也能落到 Casbin 策略上而不是漏过。
var gvaResources = map[string]bool{
	"rule":        true,
	"component":   true,
	"config":      true,
	"locale":      true,
	"log":         true,
	"marketplace": true,
	"skill":       true,
	"user":        true,
}

// actionMethod 每个 action 对应的虚拟 method。
var actionMethod = map[string]string{
	"read":    "GET",
	"write":   "POST",
	"delete":  "DELETE",
	"execute": "POST",
	"operate": "POST",
}

// actionPathSuffix execute/operate 类动作映射到资源的子路径（与上游路由
// /rules/:id/operate/:type 等执行类端点对齐）。
var actionPathSuffix = map[string]string{
	"execute": "/execute",
	"operate": "/execute",
}

// mapAction 把 rulego (resource, action) 翻译为 GVA Casbin 判权的虚拟 (path, method)。
// 未知组合返回 ok=false，调用方一律拒绝。
func mapAction(resource, action string) (path, method string, ok bool) {
	if !gvaResources[resource] {
		return "", "", false
	}
	method, ok = actionMethod[action]
	if !ok {
		return "", "", false
	}
	return prefix + "/" + resource + actionPathSuffix[action], method, true
}
