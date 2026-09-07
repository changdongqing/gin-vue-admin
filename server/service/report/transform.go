package report

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/flipped-aurora/gin-vue-admin/server/model/report"
)

// TransformExecutor 数据转换执行器（按 orderNum 串行执行）
type TransformExecutor interface {
	Support(transformType string) bool
	Execute(result *QueryResult, script string) (*QueryResult, error)
}

// RunTransforms 转换链执行（依次匹配执行器，无匹配执行器报错）
func RunTransforms(result *QueryResult, transforms []report.ReportDataSetTransform) (*QueryResult, error) {
	executors := []TransformExecutor{&JsTransformExecutor{}, &DictTransformExecutor{}}
	for _, tf := range transforms {
		var matched TransformExecutor
		for _, e := range executors {
			if e.Support(tf.TransformType) {
				matched = e
				break
			}
		}
		if matched == nil {
			return nil, fmt.Errorf("%w: 不支持的转换类型 %q", ErrSetTransformFailed, tf.TransformType)
		}
		out, err := matched.Execute(result, tf.TransformScript)
		if err != nil {
			return nil, err
		}
		result = out
	}
	return result, nil
}

// JsTransformExecutor goja 沙箱脚本转换。
// 沙箱要点：① 每次执行新建 Runtime；② 仅注入 data（行数组），不注册 require/console/fs 等
// 任何宿主对象——语言层无宿主可达，无文件/网络/协程能力；③ 3s Interrupt 防死循环；
// ④ 返回值须为数组，元素须为对象。
type JsTransformExecutor struct{}

func (e *JsTransformExecutor) Support(transformType string) bool { return transformType == "js" }

func (e *JsTransformExecutor) Execute(result *QueryResult, script string) (*QueryResult, error) {
	if strings.TrimSpace(script) == "" {
		return result, nil
	}
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	if err := vm.Set("data", result.Rows); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSetTransformFailed, err)
	}
	timer := time.AfterFunc(3*time.Second, func() { vm.Interrupt("execution timeout") })
	defer timer.Stop()
	value, err := vm.RunString(fmt.Sprintf("(function(data){ %s\n})(data)", script))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSetTransformFailed, err)
	}
	raw := value.Export()
	arr, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: 脚本返回值须为数组", ErrSetTransformFailed)
	}
	rows := make([]map[string]interface{}, 0, len(arr))
	for _, item := range arr {
		row, ok := item.(map[string]interface{})
		if !ok {
			// 允许数组元素为 goja 对象包装的结构，经 JSON 往返归一
			b, jerr := json.Marshal(item)
			if jerr != nil {
				return nil, fmt.Errorf("%w: 脚本返回数组元素须为对象", ErrSetTransformFailed)
			}
			row = map[string]interface{}{}
			if jerr = json.Unmarshal(b, &row); jerr != nil {
				return nil, fmt.Errorf("%w: 脚本返回数组元素须为对象", ErrSetTransformFailed)
			}
		}
		rows = append(rows, row)
	}
	return rebuildQueryResult(result.Columns, rows), nil
}

// DictTransformExecutor 字段值映射转换：script = {"field":"status","mapping":{"0":"停用"}}
// 未命中的值保持原样
type DictTransformExecutor struct{}

func (e *DictTransformExecutor) Support(transformType string) bool { return transformType == "dict" }

func (e *DictTransformExecutor) Execute(result *QueryResult, script string) (*QueryResult, error) {
	var cfg struct {
		Field   string            `json:"field"`
		Mapping map[string]string `json:"mapping"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(script)), &cfg); err != nil {
		return nil, fmt.Errorf("%w: dict 配置JSON格式错误: %v", ErrSetTransformFailed, err)
	}
	if cfg.Field == "" {
		return nil, fmt.Errorf("%w: dict 配置缺少 field", ErrSetTransformFailed)
	}
	rows := make([]map[string]interface{}, 0, len(result.Rows))
	for _, row := range result.Rows {
		out := make(map[string]interface{}, len(row))
		for k, v := range row {
			out[k] = v
		}
		if v, ok := row[cfg.Field]; ok {
			if mapped, hit := cfgMappingHit(cfg.Mapping, v); hit {
				out[cfg.Field] = mapped
			}
		}
		rows = append(rows, out)
	}
	return &QueryResult{Columns: append([]string{}, result.Columns...), Rows: rows}, nil
}

// cfgMappingHit 映射命中（键与值统一转字符串比对；数值 1 命中 "1"）
func cfgMappingHit(mapping map[string]string, v interface{}) (string, bool) {
	for k, target := range mapping {
		if valueToString(v) == k || fmt.Sprintf("%v", v) == k {
			return target, true
		}
	}
	return "", false
}

// rebuildQueryResult 重建结果：Columns = 原列序（仍存在的）在前 + 新列并集（按首行出现序）
func rebuildQueryResult(oldColumns []string, rows []map[string]interface{}) *QueryResult {
	seen := map[string]bool{}
	columns := make([]string, 0, len(oldColumns))
	for _, col := range oldColumns {
		for _, row := range rows {
			if _, ok := row[col]; ok {
				columns = append(columns, col)
				seen[col] = true
				break
			}
		}
	}
	for _, row := range rows {
		for k := range row {
			if !seen[k] {
				seen[k] = true
				columns = append(columns, k)
			}
		}
	}
	return &QueryResult{Columns: columns, Rows: rows}
}
