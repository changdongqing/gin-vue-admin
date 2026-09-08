package rulego

import "testing"

// 全组合覆盖：v0.37.2 的 8 个 resource × 5 个 action。上游升级引入新枚举时，本测试
// 的 unknown 组合断言会失败并提示同步映射表（升级 SOP 检查项）。
func TestMapAction(t *testing.T) {
	cases := []struct {
		resource string
		action   string
		wantPath string
		wantMeth string
		wantOK   bool
	}{
		{"rule", "read", "/rulego/rule", "GET", true},
		{"rule", "write", "/rulego/rule", "POST", true},
		{"rule", "delete", "/rulego/rule", "DELETE", true},
		{"rule", "execute", "/rulego/rule/execute", "POST", true},
		{"rule", "operate", "/rulego/rule/execute", "POST", true},

		{"component", "read", "/rulego/component", "GET", true},
		{"component", "write", "/rulego/component", "POST", true},
		{"component", "delete", "/rulego/component", "DELETE", true},
		{"config", "read", "/rulego/config", "GET", true},
		{"config", "write", "/rulego/config", "POST", true},
		{"locale", "read", "/rulego/locale", "GET", true},
		{"locale", "write", "/rulego/locale", "POST", true},
		{"log", "read", "/rulego/log", "GET", true},
		{"log", "delete", "/rulego/log", "DELETE", true},
		{"marketplace", "read", "/rulego/marketplace", "GET", true},
		{"skill", "read", "/rulego/skill", "GET", true},
		{"skill", "write", "/rulego/skill", "POST", true},
		{"skill", "delete", "/rulego/skill", "DELETE", true},
		{"user", "read", "/rulego/user", "GET", true},
		{"user", "write", "/rulego/user", "POST", true},

		// fail-closed：未知 resource / 未知 action 一律拒绝
		{"node", "read", "", "", false},
		{"mcp", "write", "", "", false},
		{"", "read", "", "", false},
		{"rule", "", "", "", false},
		{"rule", "admin", "", "", false},
		{"rule", "list", "", "", false},
	}

	for _, tc := range cases {
		path, method, ok := mapAction(tc.resource, tc.action)
		if ok != tc.wantOK || path != tc.wantPath || method != tc.wantMeth {
			t.Errorf("mapAction(%q,%q) = (%q,%q,%v), want (%q,%q,%v)",
				tc.resource, tc.action, path, method, ok, tc.wantPath, tc.wantMeth, tc.wantOK)
		}
	}
}
