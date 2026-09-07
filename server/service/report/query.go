package report

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/model/report"
)

// DataSetQueryService 数据集查询执行服务（SQL 多库执行 + HTTP 取数 + 转换链 + 行数上限 + 服务端分页）
type DataSetQueryService struct{}

// QueryServiceApp 领域内单例（04 渲染引擎复用）
var QueryServiceApp = &DataSetQueryService{}

// QueryResult 查询结果（Go 无序 map：列序显式携带，渲染引擎/预览表头依赖）
type QueryResult struct {
	Columns []string                 `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
}

// maxQueryRows 单次全量查询行数上限（防 OOM）
const maxQueryRows = 50000

// queryTimeout 单条查询执行超时
const queryTimeout = 30 * time.Second

// trailingSemicolon 去分页包装前的尾分号
var trailingSemicolon = regexp.MustCompile(`(?is);+\s*$`)

// Execute 按类型分发执行（编辑中即时测试用：直接传配置，配置明文未加密）
func (q *DataSetQueryService) Execute(setType, sourceCode, dynSentence string,
	paramValues map[string]interface{}, transforms []report.ReportDataSetTransform) (*QueryResult, error) {
	if paramValues == nil {
		paramValues = map[string]interface{}{}
	}
	switch setType {
	case "http":
		return q.ExecuteHTTP(dynSentence, paramValues)
	case "sql":
		if strings.TrimSpace(sourceCode) == "" {
			return nil, ErrSetSourceMissing
		}
		return q.ExecuteSQL(sourceCode, dynSentence, paramValues)
	default:
		return nil, fmt.Errorf("%w: %s", ErrSetParamTypeInvalid, setType)
	}
}

// Query 按已保存数据集执行（查库组装 → 缺省填充 → 必填校验 → Execute → 转换链）
func (q *DataSetQueryService) Query(setCode string, paramValues map[string]interface{}) (*QueryResult, error) {
	set, transforms, params, err := DataSetServiceApp.LoadForQuery(setCode)
	if err != nil {
		return nil, err
	}
	resolved, err := ResolveSetParam(params, paramValues)
	if err != nil {
		return nil, err
	}
	result, err := q.Execute(set.SetType, set.SourceCode, set.DynSentence, resolved, transforms)
	if err != nil {
		return nil, err
	}
	return RunTransforms(result, transforms)
}

// QueryPage 服务端分页（04 预览主数据集分页 / 05 超限回退用）。
// page 从 1 起；total 由 COUNT 包装求得；含转换器的数据集走「全量查询 + 内存分页」
func (q *DataSetQueryService) QueryPage(setType, sourceCode, dynSentence string,
	paramValues map[string]interface{}, page, pageSize int, transforms []report.ReportDataSetTransform) (*QueryResult, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	// 含转换器：全量查询（受 50000 上限）→ 内存分页
	if len(transforms) > 0 {
		result, err := q.Execute(setType, sourceCode, dynSentence, paramValues, transforms)
		if err != nil {
			return nil, 0, err
		}
		result, err = RunTransforms(result, transforms)
		if err != nil {
			return nil, 0, err
		}
		total := int64(len(result.Rows))
		start := (page - 1) * pageSize
		if start >= int(total) {
			return &QueryResult{Columns: result.Columns, Rows: []map[string]interface{}{}}, total, nil
		}
		end := start + pageSize
		if end > int(total) {
			end = int(total)
		}
		return &QueryResult{Columns: result.Columns, Rows: result.Rows[start:end]}, total, nil
	}

	switch setType {
	case "http":
		// HTTP 无分页语义：全量取回后内存分页
		result, err := q.ExecuteHTTP(dynSentence, paramValues)
		if err != nil {
			return nil, 0, err
		}
		total := int64(len(result.Rows))
		start := (page - 1) * pageSize
		if start >= int(total) {
			return &QueryResult{Columns: result.Columns, Rows: []map[string]interface{}{}}, total, nil
		}
		end := start + pageSize
		if end > int(total) {
			end = int(total)
		}
		return &QueryResult{Columns: result.Columns, Rows: result.Rows[start:end]}, total, nil
	case "sql":
		return q.queryPageSQL(sourceCode, dynSentence, paramValues, page, pageSize)
	default:
		return nil, 0, fmt.Errorf("%w: %s", ErrSetParamTypeInvalid, setType)
	}
}

// queryPageSQL SQL 分页方言包装
func (q *DataSetQueryService) queryPageSQL(sourceCode, dynSentence string,
	paramValues map[string]interface{}, page, pageSize int) (*QueryResult, int64, error) {
	if err := ValidateSQL(dynSentence); err != nil {
		return nil, 0, err
	}
	ds, err := DataSourceServiceApp.GetDataSourceByCode(sourceCode)
	if err != nil {
		return nil, 0, err
	}
	if !ds.EnableFlag {
		return nil, 0, ErrSetSourceDisabled
	}
	spec, ok := driverRegistry[ds.SourceType]
	if !ok || !spec.enabled {
		return nil, 0, ErrSourceDriverDisabled
	}
	base := trailingSemicolon.ReplaceAllString(strings.TrimSpace(dynSentence), "")
	resolved, err := Resolve(StripEmptyIfBlocks(base, paramValues), paramValues)
	if err != nil {
		return nil, 0, err
	}

	query := resolved.Query
	var countQuery string
	var args []interface{}
	switch ds.SourceType {
	case "sqlserver":
		// sqlserver：OFFSET/FETCH（须 ORDER BY，用 ORDER BY (SELECT NULL) 绕开强制排序）
		args = append(append([]interface{}{}, resolved.Args...), (page-1)*pageSize, pageSize)
		query = fmt.Sprintf("SELECT * FROM (%s) t ORDER BY (SELECT NULL) OFFSET ? ROWS FETCH NEXT ? ROWS ONLY", resolved.Query)
		countQuery = fmt.Sprintf("SELECT COUNT(*) FROM (%s) t", resolved.Query)
	default:
		// pg 系 / mysql：LIMIT ? OFFSET ?
		args = append(append([]interface{}{}, resolved.Args...), pageSize, (page-1)*pageSize)
		query = fmt.Sprintf("SELECT * FROM (%s) t LIMIT ? OFFSET ?", resolved.Query)
		countQuery = fmt.Sprintf("SELECT COUNT(*) FROM (%s) t", resolved.Query)
	}
	if spec.rebindDollar {
		query = RebindDollar(query)
		countQuery = RebindDollar(countQuery)
	}

	decrypted, err := DecryptSourceConfig(ds.SourceConfig)
	if err != nil {
		return nil, 0, err
	}
	db, err := GlobalPoolManager.GetOrCreate(ds.SourceCode, ds.SourceType, decrypted)
	if err != nil {
		return nil, 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	var total int64
	if err := db.QueryRowContext(ctx, countQuery, resolved.Args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("%w: %v", ErrSetQueryFailed, err)
	}
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", ErrSetQueryFailed, err)
	}
	defer rows.Close()
	result, err := scanRows(rows, maxQueryRows)
	if err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

// ExecuteSQL SQL 取数全流程：SELECT-only → 数据源启用 → 解密取池 → <if> 剥离 → 参数化 → rebind → 扫描
func (q *DataSetQueryService) ExecuteSQL(sourceCode, sqlText string, paramValues map[string]interface{}) (*QueryResult, error) {
	if err := ValidateSQL(sqlText); err != nil {
		return nil, err
	}
	ds, err := DataSourceServiceApp.GetDataSourceByCode(sourceCode)
	if err != nil {
		return nil, err
	}
	if !ds.EnableFlag {
		return nil, ErrSetSourceDisabled
	}
	decrypted, err := DecryptSourceConfig(ds.SourceConfig)
	if err != nil {
		return nil, err
	}
	db, err := GlobalPoolManager.GetOrCreate(ds.SourceCode, ds.SourceType, decrypted)
	if err != nil {
		return nil, err
	}
	resolved, err := Resolve(StripEmptyIfBlocks(sqlText, paramValues), paramValues)
	if err != nil {
		return nil, err
	}
	query := resolved.Query
	if RebindDollarNeeded(ds.SourceType) {
		query = RebindDollar(query)
	}
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	rows, err := db.QueryContext(ctx, query, resolved.Args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSetQueryFailed, err)
	}
	defer rows.Close()
	return scanRows(rows, maxQueryRows)
}

// httpClient 领域内单例（连接 10s / 总 30s）
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
	},
}

// httpConfigJSON HTTP 数据集请求配置 {apiUrl, method, headers, body}
type httpConfigJSON struct {
	ApiUrl  string                 `json:"apiUrl"`
	Method  string                 `json:"method"`
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body"`
}

// parseHTTPConfig 解析并校验 HTTP 请求配置
func parseHTTPConfig(dynSentence string) (*httpConfigJSON, error) {
	var cfg httpConfigJSON
	if err := json.Unmarshal([]byte(dynSentence), &cfg); err != nil {
		return nil, ErrSetHttpConfigInvalid
	}
	if strings.TrimSpace(cfg.ApiUrl) == "" {
		return nil, ErrSetHttpConfigInvalid
	}
	if cfg.Method == "" {
		cfg.Method = "GET"
	}
	return &cfg, nil
}

// ExecuteHTTP HTTP 取数：参数替换（URL/headers 字符串、body 深替换）→ 请求 → 三档响应解析
func (q *DataSetQueryService) ExecuteHTTP(dynSentence string, paramValues map[string]interface{}) (*QueryResult, error) {
	cfg, err := parseHTTPConfig(dynSentence)
	if err != nil {
		return nil, err
	}
	url := ReplaceParams(cfg.ApiUrl, paramValues)
	method := strings.ToUpper(cfg.Method)
	var bodyReader io.Reader
	if cfg.Body != nil && len(cfg.Body) > 0 {
		bodyData := ReplaceParamsDeep(cfg.Body, paramValues)
		b, mErr := json.Marshal(bodyData)
		if mErr != nil {
			return nil, fmt.Errorf("%w: %v", ErrSetHttpConfigInvalid, mErr)
		}
		bodyReader = bytes.NewReader(b)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSetHttpFailed, err)
	}
	hasContentType := false
	for k, v := range cfg.Headers {
		req.Header.Set(k, ReplaceParams(v, paramValues))
		if strings.EqualFold(k, "Content-Type") {
			hasContentType = true
		}
	}
	if bodyReader != nil && !hasContentType {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSetHttpFailed, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSetHttpFailed, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		preview := string(body)
		if len(preview) > 200 {
			preview = preview[:200]
		}
		return nil, fmt.Errorf("%w: HTTP %d %s", ErrSetHttpFailed, resp.StatusCode, preview)
	}
	return parseHTTPResponse(body)
}

// parseHTTPResponse 三档解析：JSON 数组 → 对象的 data 数组 → 对象本身单行
func parseHTTPResponse(body []byte) (*QueryResult, error) {
	var payload interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("%w: 响应不是合法 JSON: %v", ErrSetHttpFailed, err)
	}
	var rows []map[string]interface{}
	switch v := payload.(type) {
	case []interface{}:
		for _, item := range v {
			if row, ok := item.(map[string]interface{}); ok {
				rows = append(rows, row)
			}
		}
	case map[string]interface{}:
		if dataArr, ok := v["data"].([]interface{}); ok {
			for _, item := range dataArr {
				if row, ok := item.(map[string]interface{}); ok {
					rows = append(rows, row)
				}
			}
		} else {
			rows = append(rows, v)
		}
	default:
		return nil, fmt.Errorf("%w: 响应结构不可解析（数组/含data数组的对象/单对象）", ErrSetHttpFailed)
	}
	if rows == nil {
		rows = []map[string]interface{}{}
	}
	return &QueryResult{Columns: columnsFromRows(rows), Rows: rows}, nil
}

// columnsFromRows 首行 keys 为列序
func columnsFromRows(rows []map[string]interface{}) []string {
	if len(rows) == 0 {
		return []string{}
	}
	columns := make([]string, 0, len(rows[0]))
	seen := map[string]bool{}
	for k := range rows[0] {
		columns = append(columns, k)
		seen[k] = true
	}
	sort.Strings(columns)
	return columns
}

// scanRows 逐行扫描（数值/时间驱动态，json.Marshal 兼容）；超限报 ErrSetResultTooLarge
func scanRows(rows *sql.Rows, limit int) (*QueryResult, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSetQueryFailed, err)
	}
	result := &QueryResult{Columns: columns, Rows: []map[string]interface{}{}}
	for rows.Next() {
		raw := make([]interface{}, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSetQueryFailed, err)
		}
		row := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			row[col] = normalizeValue(raw[i])
		}
		result.Rows = append(result.Rows, row)
		if len(result.Rows) > limit {
			return nil, ErrSetResultTooLarge
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSetQueryFailed, err)
	}
	return result, nil
}

// normalizeValue 驱动返回值归一（[]byte → string / time → 格式化，保证 json.Marshal 兼容）
func normalizeValue(v interface{}) interface{} {
	switch t := v.(type) {
	case []byte:
		return string(t)
	case time.Time:
		if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() == 0 {
			return t.Format("2006-01-02")
		}
		return t.Format("2006-01-02 15:04:05")
	default:
		return v
	}
}
