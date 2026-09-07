package report

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// DefaultValueResolver：参数默认值动态表达式 → 实值（后端实时计算，前端仅 sampleItem 预填）
var dateLikePattern = regexp.MustCompile(`^\d{4}[-/]\d{1,2}[-/]\d{1,2}.*$`)

// RangeValue 日期范围双值（起/止）
type RangeValue [2]string

// ResolveDefaultValue 解析默认值表达式（非表达式原样返回）：
// 单值：today / now；范围：thisMonth / thisWeek / thisYear / lastMonth / last7Days / last30Days
func ResolveDefaultValue(exprOrValue string) interface{} {
	switch strings.TrimSpace(exprOrValue) {
	case "today":
		return time.Now().Format("2006-01-02")
	case "now":
		return time.Now().Format("2006-01-02 15:04:05")
	case "thisMonth":
		now := time.Now()
		first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return RangeValue{first.Format("2006-01-02"), first.AddDate(0, 1, -1).Format("2006-01-02")}
	case "thisWeek":
		now := time.Now()
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(weekday - 1))
		return RangeValue{monday.Format("2006-01-02"), monday.AddDate(0, 0, 6).Format("2006-01-02")}
	case "thisYear":
		now := time.Now()
		return RangeValue{
			time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02"),
			time.Date(now.Year(), 12, 31, 0, 0, 0, 0, now.Location()).Format("2006-01-02"),
		}
	case "lastMonth":
		now := time.Now()
		firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		lastMonthFirst := firstOfThisMonth.AddDate(0, -1, 0)
		return RangeValue{lastMonthFirst.Format("2006-01-02"), firstOfThisMonth.AddDate(0, 0, -1).Format("2006-01-02")}
	case "last7Days":
		today := time.Now()
		return RangeValue{today.AddDate(0, 0, -7).Format("2006-01-02"), today.Format("2006-01-02")}
	case "last30Days":
		today := time.Now()
		return RangeValue{today.AddDate(0, 0, -30).Format("2006-01-02"), today.Format("2006-01-02")}
	default:
		return exprOrValue
	}
}

// ResolveToMap 按参数类型展开到目标 map：
// dateRange → paramName_start / paramName_end 两个键（表达式或 "起,止" 字符串统一拆分）；
// 其余类型单值。defaultValue 为空时不写入
func ResolveToMap(paramName, paramType, defaultValue string, target map[string]interface{}) {
	if strings.TrimSpace(defaultValue) == "" {
		return
	}
	if paramType == "dateRange" {
		if rv, ok := ResolveDefaultValue(defaultValue).(RangeValue); ok {
			target[paramName+"_start"] = rv[0]
			target[paramName+"_end"] = rv[1]
			return
		}
		// "起,止" 字符串拆分
		parts := strings.SplitN(defaultValue, ",", 2)
		if len(parts) == 2 {
			target[paramName+"_start"] = strings.TrimSpace(parts[0])
			target[paramName+"_end"] = strings.TrimSpace(parts[1])
		} else {
			target[paramName+"_start"] = strings.TrimSpace(defaultValue)
			target[paramName+"_end"] = strings.TrimSpace(defaultValue)
		}
		return
	}
	if v, ok := ResolveDefaultValue(defaultValue).(RangeValue); ok {
		// 范围表达式配到了非 dateRange 类型：取起值兜底
		target[paramName] = v[0]
		return
	}
	target[paramName] = ResolveDefaultValue(defaultValue)
}

// SplitDateRangeValue 任意来源（用户提交/sampleItem）的 dateRange "起,止" 字符串拆分为 _start/_end 两键
func SplitDateRangeValue(paramName, value string, target map[string]interface{}) {
	if strings.TrimSpace(value) == "" {
		return
	}
	parts := strings.SplitN(value, ",", 2)
	if len(parts) == 2 {
		target[paramName+"_start"] = strings.TrimSpace(parts[0])
		target[paramName+"_end"] = strings.TrimSpace(parts[1])
	} else {
		target[paramName+"_start"] = strings.TrimSpace(value)
		target[paramName+"_end"] = strings.TrimSpace(value)
	}
}

// IsDateLike 值是否形如日期（05 类型推断兜底用）
func IsDateLike(s string) bool {
	return dateLikePattern.MatchString(strings.TrimSpace(s))
}

// formatValidateNote 预留：参数 JS 校验规则本期不执行（差异裁定见 02 §1.3）
var _ = fmt.Sprintf
