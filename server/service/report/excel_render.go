package report

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/report"
)

// ExcelReportRenderService 渲染引擎（对齐参考实现）：
// 主数据集 = cellData 中含「完整占位符」的第一个（按 sheet→行→列扫描序，找不到取 setCodes 首个兜底）；
// 明细模板行内引用的其它数据集按同一分页各自查询（各显示各的，超出留空）；
// 单值数据集 queryPage(1,1) 取首行；合并单元格不做位移（已知限制，模板设计约束）。
type ExcelReportRenderService struct{}

// 完整占位符（整格匹配，判定明细行）/ 片段占位符（文本内替换）——字段名宽匹配
var (
	fullPlaceholderRe = regexp.MustCompile(`^#\{([^{}]+)\.([^{}]+)}$`)
	fragPlaceholderRe = regexp.MustCompile(`#\{([^{}]+)\.([^{}]+)}`)
)

// loadSetQueryParts 数据集查询前置装载（参数定义 + 转换链）
func loadSetQueryParts(setCode string) ([]report.ReportDataSetParam, []report.ReportDataSetTransform, error) {
	params, err := DataSetServiceApp.GetParams(setCode)
	if err != nil {
		return nil, nil, err
	}
	transforms, err := DataSetServiceApp.GetTransforms(setCode)
	if err != nil {
		return nil, nil, err
	}
	return params, transforms, nil
}

// RenderResult 渲染产物（snapshot 为渲染后快照【对象】，前端免二次 parse）
type RenderResult struct {
	Snapshot map[string]interface{} `json:"snapshot"`
	Total    int64                  `json:"total"`
}

// Render 分页渲染（§3.1 流程）
func (r *ExcelReportRenderService) Render(reportCode string, paramValues map[string]interface{}, pageNo, pageSize int) (*RenderResult, error) {
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if paramValues == nil {
		paramValues = map[string]interface{}{}
	}
	var tpl report.ReportExcelTemplate
	if err := global.GVA_DB.Where("report_code = ?", reportCode).First(&tpl).Error; err != nil {
		return nil, ErrExcelReportNotExists
	}
	// 无模板：空快照 + total=0
	if strings.TrimSpace(tpl.JsonStr) == "" {
		return &RenderResult{Snapshot: map[string]interface{}{}}, nil
	}
	setCodes := splitSetCodes(tpl.SetCodes)
	if len(setCodes) == 0 {
		// 无关联数据集：原样返回模板 + total=0
		snapshot, _, err := parseSnapshot(tpl.JsonStr)
		if err != nil {
			return nil, ErrExcelTemplateJSONInvalid
		}
		return &RenderResult{Snapshot: snapshot}, nil
	}

	// ① 主数据集定位 + 全部被引用数据集收集
	snapshot, sheetOrder, err := parseSnapshot(tpl.JsonStr)
	if err != nil {
		return nil, ErrExcelTemplateJSONInvalid
	}
	mainSetCode := findMainSetCode(snapshot, setCodes)
	if mainSetCode == "" {
		mainSetCode = setCodes[0]
	}

	// ② 主数据集分页查询（fillDefaultParamValues 按数据集各自解析，含用户 dateRange 拆分）
	mainParams, mainTransforms, err := loadSetQueryParts(mainSetCode)
	if err != nil {
		return nil, err
	}
	mainResolved, err := ResolveSetParam(mainParams, paramValues)
	if err != nil {
		return nil, err
	}
	set, err := DataSetServiceApp.GetDataSetByCode(mainSetCode)
	if err != nil {
		return nil, err
	}
	mainData, total, err := QueryServiceApp.QueryPage(set.SetType, set.SourceCode, set.DynSentence, mainResolved, pageNo, pageSize, mainTransforms)
	if err != nil {
		return nil, fmt.Errorf("%w: 主数据集[%s]查询失败: %v", ErrExcelRenderFailed, mainSetCode, err)
	}

	// ③ 各 sheet 渲染（缓存按 setCode+场景：detail=当前分页 / single=首行）
	pageCache := map[string]*QueryResult{"main": mainData}
	loadPage := func(setCode string, detail bool) (*QueryResult, error) {
		key := setCode
		if !detail {
			key += "|single"
		}
		if cached, ok := pageCache[key]; ok {
			return cached, nil
		}
		params, transforms, err := loadSetQueryParts(setCode)
		if err != nil {
			return nil, err
		}
		resolved, err := ResolveSetParam(params, paramValues)
		if err != nil {
			return nil, err
		}
		ds, err := DataSetServiceApp.GetDataSetByCode(setCode)
		if err != nil {
			return nil, err
		}
		page, size := 1, 1
		if detail {
			page, size = pageNo, pageSize
		}
		data, _, err := QueryServiceApp.QueryPage(ds.SetType, ds.SourceCode, ds.DynSentence, resolved, page, size, transforms)
		if err != nil {
			return nil, err
		}
		pageCache[key] = data
		return data, nil
	}

	for _, sheetID := range sheetOrder {
		sheetMap, ok := snapshot["sheets"].(map[string]interface{})[sheetID].(map[string]interface{})
		if !ok {
			continue
		}
		if err := r.renderSheet(sheetMap, mainSetCode, mainData, setCodes, loadPage); err != nil {
			return nil, err
		}
		adjustSheetRowCount(sheetMap)
	}
	return &RenderResult{Snapshot: snapshot, Total: total}, nil
}

// renderSheet 单 sheet 渲染（明细行展开 + 占位符替换）
func (r *ExcelReportRenderService) renderSheet(sheet map[string]interface{}, mainSetCode string, mainData *QueryResult,
	setCodes []string, loadPage func(setCode string, detail bool) (*QueryResult, error)) error {
	cellData, _ := sheet["cellData"].(map[string]interface{})
	if cellData == nil {
		return nil
	}
	rows := sortedRowKeys(cellData)

	// R = 主数据集完整占位符所在最小行号
	mainRow := -1
	for _, rowKey := range rows {
		cols := sortedColKeys(cellData[rowKeyString(rowKey)])
		for _, colKey := range cols {
			v := cellValue(cellData[rowKeyString(rowKey)], colKey)
			if m := fullPlaceholderRe.FindStringSubmatch(v); m != nil && m[1] == mainSetCode {
				mainRow = rowKey
				break
			}
		}
		if mainRow >= 0 {
			break
		}
	}
	if mainRow < 0 {
		// 退化路径：整表单值替换（保留模板行不展开）
		return replaceAllRows(cellData, rows, setCodes, loadPage)
	}

	templateRow, _ := cellData[rowKeyString(mainRow)].(map[string]interface{})
	if templateRow == nil {
		return nil
	}
	// 模板行内引用的全部数据集（完整或片段占位符）
	rowSetCodes := collectRowSetCodes(templateRow, setCodes)

	// 各数据集本页行数据（主数据集用 mainData；其余 detail 按同一分页对齐）
	pageData := map[string][]map[string]interface{}{}
	maxN := 0
	for _, code := range rowSetCodes {
		var data *QueryResult
		var err error
		if code == mainSetCode {
			data = mainData
		} else {
			data, err = loadPage(code, true)
			if err != nil {
				return fmt.Errorf("%w: 数据集[%s]查询失败: %v", ErrExcelRenderFailed, code, err)
			}
		}
		pageData[code] = data.Rows
		if len(data.Rows) > maxN {
			maxN = len(data.Rows)
		}
	}
	// N=0（本页无数据）→ 整表单值替换
	if maxN == 0 {
		return replaceAllRows(cellData, rows, setCodes, loadPage)
	}

	// 位移：行键 > R 的行整体 +(N-1)（从大到小处理防覆盖）
	newCellData := map[string]interface{}{}
	for _, rowKey := range rows {
		row := cellData[rowKeyString(rowKey)]
		if rowKey > mainRow {
			newCellData[rowKeyString(rowKey+maxN-1)] = row
		} else {
			newCellData[rowKeyString(rowKey)] = row
		}
	}

	// 生成明细行：j=0..N-1 深拷贝模板行，占位符源 = 各数据集自身第 j 行（超出留空）
	for j := 0; j < maxN; j++ {
		rowCopy, err := deepCopyRow(templateRow)
		if err != nil {
			return err
		}
		source := map[string]map[string]interface{}{}
		for _, code := range rowSetCodes {
			rowsOfSet := pageData[code]
			if j < len(rowsOfSet) {
				source[code] = rowsOfSet[j]
			}
		}
		replaceRowPlaceholders(rowCopy, source)
		newCellData[rowKeyString(mainRow+j)] = rowCopy
	}
	sheet["cellData"] = newCellData

	// 其余行（含标题区）单值替换：此时 cellData 已含明细行，仅处理非明细行
	for _, rowKey := range sortedRowKeys(newCellData) {
		if rowKey >= mainRow && rowKey < mainRow+maxN {
			continue
		}
		row, _ := newCellData[rowKeyString(rowKey)].(map[string]interface{})
		if row == nil {
			continue
		}
		replaceRowPlaceholdersSingle(row, setCodes, loadPage)
	}
	return nil
}

// replaceAllRows 整表单值替换（无完整占位符/本页无数据退化路径）
func replaceAllRows(cellData map[string]interface{}, rows []int, setCodes []string, loadPage func(string, bool) (*QueryResult, error)) error {
	for _, rowKey := range rows {
		row, _ := cellData[rowKeyString(rowKey)].(map[string]interface{})
		if row == nil {
			continue
		}
		if err := replaceRowPlaceholdersSingle(row, setCodes, loadPage); err != nil {
			return err
		}
	}
	return nil
}

// replaceRowPlaceholdersSingle 单值替换：占位符源 = 各数据集首行（queryPage(1,1)）
func replaceRowPlaceholdersSingle(row map[string]interface{}, setCodes []string, loadPage func(string, bool) (*QueryResult, error)) error {
	codes := collectRowSetCodes(row, setCodes)
	source := map[string]map[string]interface{}{}
	for _, code := range codes {
		data, err := loadPage(code, false)
		if err != nil {
			return err
		}
		if len(data.Rows) > 0 {
			source[code] = data.Rows[0]
		}
	}
	replaceRowPlaceholders(row, source)
	return nil
}

// replaceRowPlaceholders 行内占位符替换：仅处理 v 为字符串且含 #{ 的 cell（片段正则逐个替换；null → 空串）
func replaceRowPlaceholders(row map[string]interface{}, source map[string]map[string]interface{}) {
	for _, cellRaw := range row {
		cell, ok := cellRaw.(map[string]interface{})
		if !ok {
			continue
		}
		v, ok := cell["v"].(string)
		if !ok || !strings.Contains(v, "#{") {
			continue
		}
		newV := fragPlaceholderRe.ReplaceAllStringFunc(v, func(match string) string {
			sub := fragPlaceholderRe.FindStringSubmatch(match)
			if len(sub) < 3 {
				return ""
			}
			rowData := source[sub[1]]
			if rowData == nil {
				return ""
			}
			return valueToString(rowData[sub[2]])
		})
		cell["v"] = newV
	}
}

// collectRowSetCodes 行内引用的全部数据集（完整或片段占位符，且 ∈ setCodes）
func collectRowSetCodes(row map[string]interface{}, setCodes []string) []string {
	allowed := map[string]bool{}
	for _, c := range setCodes {
		allowed[c] = true
	}
	seen := map[string]bool{}
	codes := []string{}
	for _, cellRaw := range row {
		cell, ok := cellRaw.(map[string]interface{})
		if !ok {
			continue
		}
		v, ok := cell["v"].(string)
		if !ok {
			continue
		}
		for _, m := range fragPlaceholderRe.FindAllStringSubmatch(v, -1) {
			if allowed[m[1]] && !seen[m[1]] {
				seen[m[1]] = true
				codes = append(codes, m[1])
			}
		}
	}
	return codes
}

// findMainSetCode 按 sheet→行→列扫描，首个「完整占位符」且 setCode ∈ setCodes 者
func findMainSetCode(snapshot map[string]interface{}, setCodes []string) string {
	allowed := map[string]bool{}
	for _, c := range setCodes {
		allowed[c] = true
	}
	sheets, _ := snapshot["sheets"].(map[string]interface{})
	order := mapKeyOrder(sheets)
	if len(order) > 0 {
		if ord, ok := snapshot["sheetOrder"].([]interface{}); ok {
			order = nil
			for _, o := range ord {
				if _, exists := sheets[fmt.Sprintf("%v", o)]; exists {
					order = append(order, fmt.Sprintf("%v", o))
				}
			}
		}
	}
	for _, sheetID := range order {
		sheet, _ := sheets[sheetID].(map[string]interface{})
		cellData, _ := sheet["cellData"].(map[string]interface{})
		for _, rowKey := range sortedRowKeys(cellData) {
			for _, colKey := range sortedColKeys(cellData[rowKeyString(rowKey)]) {
				v := cellValue(cellData[rowKeyString(rowKey)], colKey)
				if m := fullPlaceholderRe.FindStringSubmatch(v); m != nil && allowed[m[1]] {
					return m[1]
				}
			}
		}
	}
	return ""
}

// adjustSheetRowCount 行数裁剪：rowCount = max(cellData 行键, mergeData endRow) + 1
func adjustSheetRowCount(sheet map[string]interface{}) {
	cellData, _ := sheet["cellData"].(map[string]interface{})
	maxRow := -1
	for _, rowKey := range sortedRowKeys(cellData) {
		if rowKey > maxRow {
			maxRow = rowKey
		}
	}
	if mergeData, ok := sheet["mergeData"].([]interface{}); ok {
		for _, m := range mergeData {
			mm, _ := m.(map[string]interface{})
			if mm == nil {
				continue
			}
			if end, err := strconv.Atoi(fmt.Sprintf("%v", mm["endRow"])); err == nil && end > maxRow {
				maxRow = end
			}
		}
	}
	if maxRow >= 0 {
		sheet["rowCount"] = float64(maxRow + 1)
	}
}

// ── 快照解析辅助 ─────────────────────────────────────

func parseSnapshot(jsonStr string) (map[string]interface{}, []string, error) {
	var snapshot map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &snapshot); err != nil {
		return nil, nil, err
	}
	sheetOrder := []string{}
	if raw, ok := snapshot["sheetOrder"].([]interface{}); ok {
		for _, s := range raw {
			sheetOrder = append(sheetOrder, fmt.Sprintf("%v", s))
		}
	} else if sheets, ok := snapshot["sheets"].(map[string]interface{}); ok {
		sheetOrder = mapKeyOrder(sheets)
	}
	return snapshot, sheetOrder, nil
}

func splitSetCodes(setCodes string) []string {
	result := []string{}
	for _, code := range strings.Split(setCodes, "|") {
		if code = strings.TrimSpace(code); code != "" {
			result = append(result, code)
		}
	}
	return result
}

// sortedRowKeys / sortedColKeys：Go map 无序，行/列键必须数值排序后遍历（明细行生成与位移顺序敏感）
func sortedRowKeys(cellData map[string]interface{}) []int {
	keys := make([]int, 0, len(cellData))
	for k := range cellData {
		if n, err := strconv.Atoi(k); err == nil {
			keys = append(keys, n)
		}
	}
	sort.Ints(keys)
	return keys
}

func sortedColKeys(rowRaw interface{}) []int {
	row, ok := rowRaw.(map[string]interface{})
	if !ok {
		return nil
	}
	keys := make([]int, 0, len(row))
	for k := range row {
		if n, err := strconv.Atoi(k); err == nil {
			keys = append(keys, n)
		}
	}
	sort.Ints(keys)
	return keys
}

func cellValue(rowRaw interface{}, colKey int) string {
	row, ok := rowRaw.(map[string]interface{})
	if !ok {
		return ""
	}
	cell, ok := row[strconv.Itoa(colKey)].(map[string]interface{})
	if !ok {
		return ""
	}
	v, _ := cell["v"].(string)
	return v
}

func rowKeyString(n int) string { return strconv.Itoa(n) }

func deepCopyRow(row map[string]interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(row)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExcelRenderFailed, err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExcelRenderFailed, err)
	}
	return out, nil
}

func mapKeyOrder(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ReportParamDef 参数定义聚合（dataset-params 平铺返回；跨数据集按参数名去重）
type ReportParamDef struct {
	SetCode       string `json:"setCode"`
	ParamName     string `json:"paramName"`
	ParamDesc     string `json:"paramDesc"`
	ParamType     string `json:"paramType"`
	SampleItem    string `json:"sampleItem"`
	DefaultValue  string `json:"defaultValue"`
	DictType      string `json:"dictType"`
	CustomOptions string `json:"customOptions"`
	DateFormat    string `json:"dateFormat"`
	RequiredFlag  bool   `json:"requiredFlag"`
	OrderNum      int    `json:"orderNum"`
}

// GetParamDefs 报表关联的全部数据集参数定义（按 setCodes 顺序扫描，同名参数保留先出现者）
func (r *ExcelReportRenderService) GetParamDefs(reportCode string) ([]ReportParamDef, error) {
	var tpl report.ReportExcelTemplate
	if err := global.GVA_DB.Where("report_code = ?", reportCode).First(&tpl).Error; err != nil {
		return nil, ErrExcelReportNotExists
	}
	defs := []ReportParamDef{}
	seen := map[string]bool{}
	for _, code := range splitSetCodes(tpl.SetCodes) {
		params, err := DataSetServiceApp.GetParams(code)
		if err != nil {
			return nil, err
		}
		for _, p := range params {
			if seen[p.ParamName] {
				continue
			}
			seen[p.ParamName] = true
			defs = append(defs, ReportParamDef{
				SetCode:       code,
				ParamName:     p.ParamName,
				ParamDesc:     p.ParamDesc,
				ParamType:     p.ParamType,
				SampleItem:    p.SampleItem,
				DefaultValue:  p.DefaultValue,
				DictType:      p.DictType,
				CustomOptions: p.CustomOptions,
				DateFormat:    p.DateFormat,
				RequiredFlag:  p.RequiredFlag,
				OrderNum:      p.OrderNum,
			})
		}
	}
	return defs, nil
}
