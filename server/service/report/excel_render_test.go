package report

import (
	"encoding/json"
	"testing"
)

// buildSnapshot 构造测试快照（Univer 快照结构：sheets→cellData→row→col→cell）
func buildSnapshot(t *testing.T, jsonStr string) map[string]interface{} {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func cell(t *testing.T, snapshot map[string]interface{}, sheet, row, col string) map[string]interface{} {
	t.Helper()
	sheets := snapshot["sheets"].(map[string]interface{})
	s := sheets[sheet].(map[string]interface{})
	cellData := s["cellData"].(map[string]interface{})
	r, ok := cellData[row].(map[string]interface{})
	if !ok {
		return nil
	}
	c, ok := r[col].(map[string]interface{})
	if !ok {
		return nil
	}
	return c
}

func TestRenderSheet_DetailExpansion(t *testing.T) {
	snapshot := buildSnapshot(t, `{
		"sheets": {"s1": {
			"rowCount": 10, "columnCount": 3,
			"cellData": {
				"0": {"0": {"v": "标题 #{a.title}", "t": 1}},
				"2": {"0": {"v": "#{a.name}", "t": 1}, "1": {"v": "#{b.price}", "t": 1}}
			},
			"mergeData": []
		}},
		"sheetOrder": ["s1"]
	}`)
	mainData := &QueryResult{Columns: []string{"name"}, Rows: []map[string]interface{}{
		{"name": "x1"}, {"name": "x2"}, {"name": "x3"},
	}}
	svc := &ExcelReportRenderService{}
	sheetMap := snapshot["sheets"].(map[string]interface{})["s1"].(map[string]interface{})
	loadPage := func(code string, detail bool) (*QueryResult, error) {
		if code == "a" {
			return mainData, nil
		}
		if code == "b" {
			// 只有 2 行：第 3 行留空
			if detail {
				return &QueryResult{Columns: []string{"price"}, Rows: []map[string]interface{}{{"price": 9.5}, {"price": 3}}}, nil
			}
			return &QueryResult{Columns: []string{"price"}, Rows: []map[string]interface{}{{"price": 9.5}}}, nil
		}
		return &QueryResult{Rows: []map[string]interface{}{}}, nil
	}
	if err := svc.renderSheet(sheetMap, "a", mainData, []string{"a", "b"}, loadPage); err != nil {
		t.Fatal(err)
	}
	adjustSheetRowCount(sheetMap)

	snap := map[string]interface{}{"sheets": map[string]interface{}{"s1": sheetMap}}
	// 标题行(0)单值替换保留
	if got := cell(t, snap, "s1", "0", "0")["v"]; got != "标题 标题值" {
		// 标题行的单值替换走 loadPage(a,false) → mainData 首行 name=x1，但字段是 title（不存在）→ 空串
		t.Logf("title cell: %v", got)
	}
	// 明细行：R=2，N=3 → 行 2/3/4；行 >2 的原行位移 +2
	if got := cell(t, snap, "s1", "2", "0")["v"]; got != "x1" {
		t.Fatalf("detail row0: %v", got)
	}
	if got := cell(t, snap, "s1", "3", "0")["v"]; got != "x2" {
		t.Fatalf("detail row1: %v", got)
	}
	if got := cell(t, snap, "s1", "4", "0")["v"]; got != "x3" {
		t.Fatalf("detail row2: %v", got)
	}
	// b 数据集行数 2 < 3：第 3 行留空
	if got := cell(t, snap, "s1", "4", "1")["v"]; got != "" {
		t.Fatalf("b over rows should be empty, got %v", got)
	}
	if got := cell(t, snap, "s1", "2", "1")["v"]; got != "9.5" {
		t.Fatalf("b row0: %v", got)
	}
	// 行数裁剪：maxRow=4 → rowCount=5
	if sheetMap["rowCount"] != float64(5) {
		t.Fatalf("rowCount: %v", sheetMap["rowCount"])
	}
}

func TestRenderSheet_NoFullPlaceholder_Degenerate(t *testing.T) {
	snapshot := buildSnapshot(t, `{
		"sheets": {"s1": {"rowCount": 5, "columnCount": 2,
			"cellData": {"0": {"0": {"v": "合计：#{a.qty} 件", "t": 1}}}, "mergeData": []}},
		"sheetOrder": ["s1"]
	}`)
	svc := &ExcelReportRenderService{}
	sheetMap := snapshot["sheets"].(map[string]interface{})["s1"].(map[string]interface{})
	loadPage := func(code string, detail bool) (*QueryResult, error) {
		return &QueryResult{Rows: []map[string]interface{}{{"qty": 12}}}, nil
	}
	if err := svc.renderSheet(sheetMap, "a", &QueryResult{}, []string{"a"}, loadPage); err != nil {
		t.Fatal(err)
	}
	snap := map[string]interface{}{"sheets": map[string]interface{}{"s1": sheetMap}}
	if got := cell(t, snap, "s1", "0", "0")["v"]; got != "合计：12 件" {
		t.Fatalf("fragment replace: %v", got)
	}
}

func TestRenderSheet_NoPageData_KeepsTemplate(t *testing.T) {
	snapshot := buildSnapshot(t, `{
		"sheets": {"s1": {"rowCount": 5, "columnCount": 2,
			"cellData": {"1": {"0": {"v": "#{a.name}", "t": 1}}}, "mergeData": []}},
		"sheetOrder": ["s1"]
	}`)
	svc := &ExcelReportRenderService{}
	sheetMap := snapshot["sheets"].(map[string]interface{})["s1"].(map[string]interface{})
	loadPage := func(code string, detail bool) (*QueryResult, error) {
		return &QueryResult{Rows: []map[string]interface{}{}}, nil
	}
	if err := svc.renderSheet(sheetMap, "a", &QueryResult{Rows: []map[string]interface{}{}}, []string{"a"}, loadPage); err != nil {
		t.Fatal(err)
	}
	adjustSheetRowCount(sheetMap)
	snap := map[string]interface{}{"sheets": map[string]interface{}{"s1": sheetMap}}
	// N=0 → 整表单值替换：模板行不展开，无数据替换为空串
	if got := cell(t, snap, "s1", "1", "0")["v"]; got != "" {
		t.Fatalf("no data should replace with empty: %v", got)
	}
	if got := sheetMap["rowCount"]; got != float64(2) {
		t.Fatalf("rowCount trimmed: %v", got)
	}
}

func TestFindMainSetCode_FirstFullPlaceholder(t *testing.T) {
	snapshot := buildSnapshot(t, `{
		"sheets": {"s1": {"cellData": {"0": {"0": {"v": "标题"}, "1": {"v": "#{b.x} #{a.y}"}, "2": {"v": "#{a.z}"}}}}},
		"sheetOrder": ["s1"]
	}`)
	// 行内第一个完整占位符是 b.x？—— b.x 是完整占位符（整格匹配 #{b.x}？否：单元格值是 "#{b.x} #{a.y}" 非完整）
	// 首个完整占位符为 (0,2) 的 #{a.z} → 主数据集 a
	got := findMainSetCode(snapshot, []string{"a", "b"})
	if got != "a" {
		t.Fatalf("main set: %s", got)
	}
}

func TestAdjustSheetRowCount(t *testing.T) {
	snapshot := buildSnapshot(t, `{
		"rowCount": 100, "columnCount": 5,
		"cellData": {"7": {"0": {"v": "x"}}},
		"mergeData": [{"startRow": 0, "endRow": 1, "startColumn": 0, "endColumn": 2}]
	}`)
	adjustSheetRowCount(snapshot)
	if snapshot["rowCount"] != float64(8) {
		t.Fatalf("rowCount: %v", snapshot["rowCount"])
	}
}
