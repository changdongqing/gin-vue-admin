package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/model/report"
)

// jsonUnmarshal 局部别名（避免各文件重复 import encoding/json 的困惑点统一）
func jsonUnmarshal(b []byte, v interface{}) error { return json.Unmarshal(b, v) }

// ResolveSetParam 参数缺省填充与校验（fillDefaultParamValues）：
// 用户值 > sampleItem > 默认值表达式（ResolveToMap）；
// dateRange 字符串（用户传入/sampleItem/默认值三种来源）统一拆 _start/_end；
// 必填参数（用户值与 sampleItem 均空且无默认值）报 ErrSetParamRequired。
// 返回展开后的新 map（不修改调用方 map）
func ResolveSetParam(params []report.ReportDataSetParam, userValues map[string]interface{}) (map[string]interface{}, error) {
	resolved := map[string]interface{}{}
	for _, param := range params {
		name := param.ParamName
		userVal, hasUser := userValues[name]

		// ① 用户显式传值优先
		if hasUser && hasValue(userVal) {
			if param.ParamType == "dateRange" {
				SplitDateRangeValue(name, valueToString(userVal), resolved)
			} else {
				resolved[name] = userVal
			}
			continue
		}

		// ② sampleItem（示例值）预填
		if strings.TrimSpace(param.SampleItem) != "" {
			if param.ParamType == "dateRange" {
				SplitDateRangeValue(name, param.SampleItem, resolved)
			} else {
				resolved[name] = ResolveDefaultValue(param.SampleItem)
			}
			continue
		}

		// ③ 默认值表达式（后端解析）
		if strings.TrimSpace(param.DefaultValue) != "" {
			ResolveToMap(name, param.ParamType, param.DefaultValue, resolved)
			continue
		}

		// ④ 必填校验
		if param.RequiredFlag {
			return nil, fmt.Errorf("%w: %s", ErrSetParamRequired, name)
		}
	}
	// 合并用户显式传入的、未在参数表中定义的值（宽松透传，供 <if>/自定义使用）
	for k, v := range userValues {
		if _, defined := resolved[k]; !defined && !rangeHandled(params, k) {
			resolved[k] = v
		}
	}
	return resolved, nil
}

// rangeHandled 判断该键是否为某 dateRange 参数展开出的 _start/_end（避免原键误透传）
func rangeHandled(params []report.ReportDataSetParam, key string) bool {
	for _, param := range params {
		if param.ParamType == "dateRange" {
			if key == param.ParamName+"_start" || key == param.ParamName+"_end" {
				return true
			}
		}
	}
	return false
}
