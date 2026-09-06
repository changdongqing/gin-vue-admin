package ontology

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// ExtDynamicQueryService 动态读业务表（安全关键，只读）：
// 三重防线——information_schema 白名单 → 标识符双引号转义 → 值/分页占位符参数化。
// 禁止任何用户输入直接拼接进 SQL 文本。
type ExtDynamicQueryService struct{}

// AssertTableValid 白名单校验：表必须存在于 public schema，返回列名→dataType 映射
func (s *ExtDynamicQueryService) AssertTableValid(tableName string) (map[string]string, error) {
	type colRow struct {
		ColumnName string
		DataType   string
	}
	var rows []colRow
	err := global.GVA_DB.Raw(
		`SELECT column_name, data_type FROM information_schema.columns
		 WHERE table_schema = 'public' AND table_name = ? ORDER BY ordinal_position`, tableName).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("表未注册或不存在: " + tableName)
	}
	cols := make(map[string]string, len(rows))
	for _, r := range rows {
		cols[r.ColumnName] = r.DataType
	}
	return cols, nil
}

// QuoteIdent 标识符统一转义（表名/列名）：双引号包裹 + 内部双引号翻倍
func (s *ExtDynamicQueryService) QuoteIdent(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}

// EscapeLike ILIKE 通配符转义（% _ \）
func (s *ExtDynamicQueryService) EscapeLike(v string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(v)
}

// SoftDeleteCond 软删条件片段：timestamptz → IS NULL；数值/布尔 → = 0
// 未配置软删列返回空串（业务行物理删除的场景直接全表扫）
func (s *ExtDynamicQueryService) SoftDeleteCond(deletedColumn, dataType string) string {
	if deletedColumn == "" {
		return ""
	}
	q := s.QuoteIdent(deletedColumn)
	switch dataType {
	case "timestamp without time zone", "timestamp with time zone", "date":
		return q + " IS NULL"
	default:
		return q + " = 0"
	}
}

// SelectRows 动态只读查行：显式列清单 + 固定白名单标识符 + 占位符参数
// whereConds 为已转义标识符拼好的条件片段；args 与 ? 一一对应
func (s *ExtDynamicQueryService) SelectRows(tableName string, selectCols []string, whereConds []string, args []interface{}, orderByCol string, limit, offset int) (*RowsResult, error) {
	cols := make([]string, 0, len(selectCols))
	for _, c := range selectCols {
		cols = append(cols, s.QuoteIdent(c))
	}
	sqlText := "SELECT " + strings.Join(cols, ", ") + " FROM " + s.QuoteIdent(tableName)
	if len(whereConds) > 0 {
		sqlText += " WHERE " + strings.Join(whereConds, " AND ")
	}
	if orderByCol != "" {
		sqlText += " ORDER BY " + s.QuoteIdent(orderByCol)
	}
	if limit > 0 {
		sqlText += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}
	rows, err := global.GVA_DB.Raw(sqlText, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result := &RowsResult{Columns: columns, Rows: make([][]interface{}, 0, 64)}
	for rows.Next() {
		vals := make([]interface{}, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		result.Rows = append(result.Rows, vals)
	}
	return result, rows.Err()
}

// RowsResult 动态查询结果（列名 + 原始值矩阵）
type RowsResult struct {
	Columns []string
	Rows    [][]interface{}
}

// ValueAt 按列名取行值
func (r *RowsResult) ValueAt(row []interface{}, column string) interface{} {
	for i, c := range r.Columns {
		if c == column {
			return row[i]
		}
	}
	return nil
}
