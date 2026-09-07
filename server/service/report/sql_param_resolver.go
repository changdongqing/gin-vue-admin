package report

import (
	"fmt"
	"regexp"
	"strings"
)

// SqlParamResolver：查询语句参数化解析与防注入（Go 移植，对齐参考实现三层防线）
// ① ValidateSQL：SELECT-only + 危险词拦截（<if> 剥离前的原文上执行）
// ② StripEmptyIfBlocks：空值条件片段整段剥离
// ③ Resolve：${p} → ? 占位符 + args 传值（根本防线）；逗号分隔值展开 IN
// ④ RebindDollar：PG 系驱动 ? → $1..$N
var (
	paramPattern  = regexp.MustCompile(`\$\{(\w+)\}`)
	ifPattern     = regexp.MustCompile(`(?s)<if\s+param="(\w+)"\s*>(.*?)</if>`)
	dangerPattern = regexp.MustCompile(`(?i)\b(DROP|DELETE|UPDATE|INSERT|TRUNCATE|ALTER|CREATE|GRANT|REVOKE|MERGE|CALL|EXEC|EXECUTE)\b`)
	selectStart   = regexp.MustCompile(`(?is)^\s*(--[^\n]*\n|/\*.*?\*/|\s)*(SELECT|WITH)\b`)
)

// ValidateSQL SELECT-only + 危险词拦截（在 <if> 剥离前的原文上执行）
func ValidateSQL(sqlText string) error {
	trimmed := strings.TrimSpace(sqlText)
	if trimmed == "" {
		return ErrSetSqlInvalid
	}
	if !selectStart.MatchString(trimmed) {
		return ErrSetSqlInvalid
	}
	if m := dangerPattern.FindString(trimmed); m != "" {
		return fmt.Errorf("%w: %s", ErrSetSqlDangerous, strings.ToUpper(m))
	}
	return nil
}

// StripEmptyIfBlocks 条件片段剥离：
// 参数值为空（不存在/null/空白字符串）→ 整段删除；非空 → 仅保留片段内容（参与后续参数化）。
// 仅支持单层不嵌套（设计约束）
func StripEmptyIfBlocks(sqlText string, values map[string]interface{}) string {
	return ifPattern.ReplaceAllStringFunc(sqlText, func(match string) string {
		sub := ifPattern.FindStringSubmatch(match)
		if len(sub) < 3 {
			return ""
		}
		if !hasValue(values[sub[1]]) {
			return ""
		}
		return sub[2]
	})
}

// hasValue 参数值非空判定（null/空白串视为空；数值 0、布尔 false 视为有值）
func hasValue(v interface{}) bool {
	if v == nil {
		return false
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s) != ""
	}
	return true
}

// ResolvedSQL 解析产物（Query 含 ? 占位符，未 rebind）
type ResolvedSQL struct {
	Query string
	Args  []interface{}
}

// Resolve ${p} → ?；未提供取值的参数缺省为 ""；
// 逗号分隔字符串展开为多个占位符（IN 场景：workshop IN (${ws})，ws="A,B" → IN (?,?)）
func Resolve(sqlText string, values map[string]interface{}) (*ResolvedSQL, error) {
	var args []interface{}
	query := paramPattern.ReplaceAllStringFunc(sqlText, func(match string) string {
		name := paramPattern.FindStringSubmatch(match)[1]
		v, ok := values[name]
		if !ok || v == nil {
			v = ""
		}
		if s, isStr := v.(string); isStr && strings.Contains(s, ",") && strings.TrimSpace(s) != "" {
			for _, part := range strings.Split(s, ",") {
				args = append(args, strings.TrimSpace(part))
			}
			return strings.TrimSuffix(strings.Repeat("?,", len(strings.Split(s, ","))), ",")
		}
		args = append(args, v)
		return "?"
	})
	return &ResolvedSQL{Query: query, Args: args}, nil
}

// RebindDollar ? → $1..$N（PG 系驱动）
func RebindDollar(query string) string {
	var b strings.Builder
	n := 0
	for _, r := range query {
		if r == '?' {
			n++
			b.WriteString(fmt.Sprintf("$%d", n))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// ReplaceParams HTTP 场景纯字符串替换（URL/header）：${p} → 字符串值（缺失替换为空串）
func ReplaceParams(text string, values map[string]interface{}) string {
	return paramPattern.ReplaceAllStringFunc(text, func(match string) string {
		name := paramPattern.FindStringSubmatch(match)[1]
		return valueToString(values[name])
	})
}

// ReplaceParamsDeep HTTP body 深替换：map/slice 递归，标量字段做字符串替换
func ReplaceParamsDeep(v interface{}, values map[string]interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(t))
		for k, val := range t {
			out[k] = ReplaceParamsDeep(val, values)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(t))
		for i, val := range t {
			out[i] = ReplaceParamsDeep(val, values)
		}
		return out
	case string:
		return ReplaceParams(t, values)
	default:
		return v
	}
}

func valueToString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
