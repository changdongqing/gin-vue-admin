package ontology

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
	ontRes "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/response"
	"gorm.io/gorm"
)

// selectPlan 动态取数计划（列清单已白名单校验；条件值全参数化）
type selectPlan struct {
	cols  []string
	conds []string
	args  []interface{}
}

// buildSelectPlan 构造主表取数计划：软删条件 [+ 水位] [+ 状态过滤]，显式列清单
func (s *ExtSyncService) buildSelectPlan(cfg *syncConfig, incremental bool, watermark *time.Time) (*selectPlan, error) {
	b := cfg.binding
	colSet := map[string]bool{
		b.KeyColumn: true, cfg.table.PkColumn: true,
	}
	for _, c := range []string{b.CodeColumn, b.NameColumn, b.ParentColumn, b.StatusColumn, cfg.table.UpdateTimeColumn} {
		if c != "" {
			colSet[c] = true
		}
	}
	for _, p := range cfg.properties {
		if p.DetailBindingId == nil && p.BizColumn != "" {
			colSet[p.BizColumn] = true
		}
	}
	cols := make([]string, 0, len(colSet))
	for c := range colSet {
		if _, ok := cfg.mainCols[c]; !ok {
			return nil, errors.New("列不在主表白名单内: " + c)
		}
		cols = append(cols, c)
	}
	conds := make([]string, 0)
	args := make([]interface{}, 0)
	if cond := extDynamicQuerySvc.SoftDeleteCond(cfg.table.DeletedColumn, cfg.mainCols[cfg.table.DeletedColumn]); cond != "" {
		conds = append(conds, cond)
	}
	if incremental && cfg.table.UpdateTimeColumn != "" {
		conds = append(conds, extDynamicQuerySvc.QuoteIdent(cfg.table.UpdateTimeColumn)+" > ?")
		args = append(args, watermark)
	}
	if b.IncludeDisabled == 0 && b.StatusColumn != "" {
		conds = append(conds, extDynamicQuerySvc.QuoteIdent(b.StatusColumn)+" = ?")
		args = append(args, b.StatusActiveValue)
	}
	return &selectPlan{cols: cols, conds: conds, args: args}, nil
}

// detailMap 子表批数据的内存索引
type detailMap struct {
	kind    int
	joinCol string
	wide    map[string]map[string]interface{} // join值 → 列 → 值
	narrow  map[string]string                 // join值\x00identifier → 值列值
}

// loadDetailMaps 按批 IN 取子表数据（禁止逐行 N+1）
func (s *ExtSyncService) loadDetailMaps(cfg *syncConfig, res *RowsResult) (map[uint]*detailMap, error) {
	maps := make(map[uint]*detailMap, len(cfg.details))
	for _, d := range cfg.details {
		keys := make([]string, 0, len(res.Rows))
		for _, row := range res.Rows {
			keys = append(keys, trimValue(res.ValueAt(row, cfg.binding.KeyColumn)))
		}
		dm := &detailMap{kind: d.DetailKind, joinCol: d.JoinColumn}
		if d.DetailKind == 2 {
			dm.narrow = map[string]string{}
		} else {
			dm.wide = map[string]map[string]interface{}{}
		}
		if len(keys) > 0 {
			planCols := []string{d.JoinColumn}
			if d.DetailKind == 2 {
				planCols = append(planCols, d.IdentifierColumn, d.ValueColumn)
			} else {
				// 宽表取属性行引用到的全部列
				seen := map[string]bool{d.JoinColumn: true}
				for _, p := range cfg.properties {
					if p.BindingType == 1 && p.DetailBindingId != nil && *p.DetailBindingId == d.ID && p.BizColumn != "" && !seen[p.BizColumn] {
						seen[p.BizColumn] = true
						planCols = append(planCols, p.BizColumn)
					}
				}
			}
			cols := cfg.detailCols[d.ID]
			for _, c := range planCols {
				if _, ok := cols[c]; !ok {
					return nil, errors.New("列不在子表白名单内: " + c)
				}
			}
			conds := make([]string, 0)
			args := make([]interface{}, 0)
			if cond := extDynamicQuerySvc.SoftDeleteCond(softColOf(cfg, d.ID), cols[softColOf(cfg, d.ID)]); cond != "" {
				conds = append(conds, cond)
			}
			conds = append(conds, extDynamicQuerySvc.QuoteIdent(d.JoinColumn)+" IN ?")
			args = append(args, keys)
			res2, err := extDynamicQuerySvc.SelectRows(cfg.detailTables[d.TableId].Table, planCols, conds, args, "", 0, 0)
			if err != nil {
				return nil, fmt.Errorf("子表读取失败: %w", err)
			}
			for _, r2 := range res2.Rows {
				joinVal := trimValue(res2.ValueAt(r2, d.JoinColumn))
				if d.DetailKind == 2 {
					ident := trimValue(res2.ValueAt(r2, d.IdentifierColumn))
					dm.narrow[joinVal+"\x00"+ident] = trimValue(res2.ValueAt(r2, d.ValueColumn))
				} else {
					rowMap := map[string]interface{}{}
					for i, c := range res2.Columns {
						rowMap[c] = r2[i]
					}
					dm.wide[joinVal] = rowMap
				}
			}
		}
		maps[d.ID] = dm
	}
	return maps, nil
}

// softColOf 取子表注册的软删列
func softColOf(cfg *syncConfig, detailID uint) string {
	t := cfg.detailTables[detailID]
	if t.DeletedColumn == "" {
		return ""
	}
	return t.DeletedColumn
}

// batchMaxWatermark 本批主表水位最大值
func batchMaxWatermark(cfg *syncConfig, res *RowsResult) *time.Time {
	if cfg.table.UpdateTimeColumn == "" {
		return nil
	}
	var max *time.Time
	for _, row := range res.Rows {
		if t := asTime(res.ValueAt(row, cfg.table.UpdateTimeColumn)); t != nil {
			if max == nil || t.After(*max) {
				tt := *t
				max = &tt
			}
		}
	}
	return max
}

// processRow 单行处理（行级小事务；部分成功是正常业务结果）
func (s *ExtSyncService) processRow(cfg *syncConfig, row []interface{}, res *RowsResult, rowMaps map[uint]*detailMap, result *ontRes.SyncResult, skipReasons *[]string, operator string) {
	b := cfg.binding
	bizKey := trimValue(res.ValueAt(row, b.KeyColumn))
	if bizKey == "" {
		result.Skipped++
		*skipReasons = append(*skipReasons, "主键列值为空: 行主键="+trimValue(res.ValueAt(row, cfg.table.PkColumn)))
		return
	}
	code := s.buildCode(cfg, row, res, bizKey)
	name := s.buildName(cfg, row, res, code)

	// 编码冲突：同名对象编码已存在且非本 biz_key（防唯一索引冲突，跳过不失败）
	var conflict ontology.OntObject
	if err := global.GVA_DB.Where("project_id = ? AND object_code = ? AND deleted_at IS NULL", b.ProjectId, code).
		First(&conflict).Error; err == nil {
		if !(conflict.ClassId == b.ClassId && conflict.BizTable == cfg.table.Table && conflict.BizKey == bizKey) {
			result.Skipped++
			*skipReasons = append(*skipReasons, "对象编码冲突: "+code)
			return
		}
	}
	// 身份一致性：查既有对象 (class_id, biz_table, biz_key)
	var existing ontology.OntObject
	found := global.GVA_DB.Where("class_id = ? AND biz_table = ? AND biz_key = ? AND deleted_at IS NULL",
		b.ClassId, cfg.table.Table, bizKey).First(&existing).Error == nil

	// 先做只读解析（父/属性值/关系），policy=2 时整行不落库
	type valuePlan struct {
		propertyId uint
		valueJson  string
	}
	type relationPlan struct {
		objectPropertyId uint
		targetObjectId   uint
	}
	valuePlans := make([]valuePlan, 0)
	relationPlans := make([]relationPlan, 0)
	issues := 0
	policy2 := b.MissingTargetPolicy == 2
	rowTxFailed := false

	// 树形父解析（同绑定 biz_table+parent_key 反查）
	var parentId *uint
	if b.ParentColumn != "" {
		parentKey := trimValue(res.ValueAt(row, b.ParentColumn))
		if parentKey != "" {
			var parent ontology.OntObject
			if err := global.GVA_DB.Where("class_id = ? AND biz_table = ? AND biz_key = ? AND deleted_at IS NULL",
				b.ClassId, cfg.table.Table, parentKey).First(&parent).Error; err == nil {
				parentId = &parent.ID
			} else if policy2 {
				result.Failed++
				*skipReasons = append(*skipReasons, "目标缺失(父): "+parentKey)
				rowTxFailed = true
			} else {
				issues++
				*skipReasons = append(*skipReasons, "父对象未解析(跳过): "+parentKey)
			}
		}
	}
	if rowTxFailed {
		return
	}

	// 数据属性值（转换失败该值跳过 + issues，行继续）
	for _, p := range cfg.properties {
		if p.BindingType != 1 {
			continue
		}
		dp := cfg.dtProps[p.PropertyId]
		raw, dataType, _ := s.rawValue(cfg, p, row, res, rowMaps)
		valueJson, ok := convertValue(raw, dataType, dp.XsdType)
		if !ok {
			issues++
			*skipReasons = append(*skipReasons, "属性「"+dpLabel(cfg, p.PropertyId)+"」列 "+p.BizColumn+" 类型不兼容，值未落库")
			continue
		}
		valuePlans = append(valuePlans, valuePlan{propertyId: p.PropertyId, valueJson: valueJson})
	}
	// 对象属性关系解析（目标类生效绑定 biz_table+biz_key 反查对象）
	for _, p := range cfg.properties {
		if p.BindingType != 2 {
			continue
		}
		op := cfg.objProps[p.PropertyId]
		raw := trimValue(res.ValueAt(row, p.BizColumn))
		if raw == "" {
			continue
		}
		resolved := false
		if op.RangeClassId != nil {
			if targetTable := cfg.rangeActiveTable[*op.RangeClassId]; targetTable != "" {
				var target ontology.OntObject
				if err := global.GVA_DB.Where("class_id = ? AND biz_table = ? AND biz_key = ? AND deleted_at IS NULL",
					*op.RangeClassId, targetTable, raw).First(&target).Error; err == nil {
					relationPlans = append(relationPlans, relationPlan{objectPropertyId: p.PropertyId, targetObjectId: target.ID})
					resolved = true
				}
			}
		}
		if !resolved {
			if policy2 {
				result.Failed++
				*skipReasons = append(*skipReasons, "目标缺失(关系 "+opLabel(cfg, p.PropertyId)+"): "+raw)
				rowTxFailed = true
				break
			}
			issues++
			*skipReasons = append(*skipReasons, "关系未解析(跳过 "+opLabel(cfg, p.PropertyId)+"): "+raw)
		}
	}
	if rowTxFailed {
		return
	}

	// 行级小事务写库
	now := time.Now()
	state := 1
	if b.StatusColumn != "" {
		state = boolToInt(trimValue(res.ValueAt(row, b.StatusColumn)) == b.StatusActiveValue)
		if state == 0 {
			state = 2
		}
	}
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if !found {
			obj := ontology.OntObject{
				ProjectId: b.ProjectId, ClassId: b.ClassId,
				ObjectCode: code, ObjectName: name, ParentId: parentId,
				State: state, SourceType: 2, BizTable: cfg.table.Table, BizKey: bizKey,
				LastSyncTime: &now, CreatedBy: operator, UpdatedBy: operator,
			}
			if err := tx.Create(&obj).Error; err != nil {
				return err
			}
			existing = obj
			result.Created++
		} else {
			// 值变才写、才计 updated（AC-19.7：无变化重跑 updated=0）
			drift := map[string]interface{}{}
			if existing.ObjectName != name {
				drift["object_name"] = name
			}
			if existing.State != state {
				drift["state"] = state
			}
			if !ptrEqual(existing.ParentId, parentId) {
				drift["parent_id"] = parentId
			}
			if len(drift) > 0 {
				if b.ConflictStrategy == 1 {
					issues++ // 仅报告：diff 记 issue，不覆盖
				} else {
					drift["last_sync_time"] = now
					drift["updated_by"] = operator
					if err := tx.Model(&ontology.OntObject{}).Where("id = ?", existing.ID).Updates(drift).Error; err != nil {
						return err
					}
					result.Updated++
				}
			} else {
				if err := tx.Model(&ontology.OntObject{}).Where("id = ?", existing.ID).
					Updates(map[string]interface{}{"last_sync_time": now}).Error; err != nil {
					return err
				}
			}
		}
		// 数据属性值 upsert
		for _, vp := range valuePlans {
			var av ontology.OntObjectAttrValue
			if err := tx.Where("object_id = ? AND property_id = ? AND deleted_at IS NULL", existing.ID, vp.propertyId).
				First(&av).Error; err == nil {
				if av.ValueJson != vp.valueJson {
					if err := tx.Model(&ontology.OntObjectAttrValue{}).Where("id = ?", av.ID).
						Update("value_json", vp.valueJson).Error; err != nil {
						return err
					}
				}
			} else {
				if err := tx.Create(&ontology.OntObjectAttrValue{ObjectId: existing.ID, PropertyId: vp.propertyId, ValueJson: vp.valueJson}).Error; err != nil {
					return err
				}
			}
		}
		// 对象关系 upsert
		for _, rp := range relationPlans {
			var cnt int64
			if err := tx.Model(&ontology.OntObjectRelation{}).
				Where("object_id = ? AND object_property_id = ? AND target_object_id = ? AND deleted_at IS NULL",
					existing.ID, rp.objectPropertyId, rp.targetObjectId).Count(&cnt).Error; err != nil {
				return err
			}
			if cnt == 0 {
				if err := tx.Create(&ontology.OntObjectRelation{ObjectId: existing.ID, ObjectPropertyId: rp.objectPropertyId, TargetObjectId: rp.targetObjectId}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		result.Failed++
		*skipReasons = append(*skipReasons, "行写入失败: biz_key="+bizKey)
		return
	}
	result.IssuesCount += issues
}

// detectOrphans 全量孤儿检测：不带状态过滤的 key 全集 vs 已物化对象 → state=3
func (s *ExtSyncService) detectOrphans(cfg *syncConfig, plan *selectPlan, operator string) ([]string, error) {
	// key 全集（仅主键列，忽略状态过滤）
	fullConds := make([]string, 0)
	if cond := extDynamicQuerySvc.SoftDeleteCond(cfg.table.DeletedColumn, cfg.mainCols[cfg.table.DeletedColumn]); cond != "" {
		fullConds = append(fullConds, cond)
	}
	keys := map[string]bool{}
	offset := 0
	for {
		res, err := extDynamicQuerySvc.SelectRows(cfg.table.Table, []string{cfg.binding.KeyColumn}, fullConds, nil, cfg.table.PkColumn, syncBatchSize, offset)
		if err != nil {
			return nil, err
		}
		if len(res.Rows) == 0 {
			break
		}
		for _, row := range res.Rows {
			keys[trimValue(res.ValueAt(row, cfg.binding.KeyColumn))] = true
		}
		if len(res.Rows) < syncBatchSize {
			break
		}
		offset += syncBatchSize
	}
	var objects []ontology.OntObject
	if err := global.GVA_DB.Select("id, biz_key").
		Where("class_id = ? AND biz_table = ? AND deleted_at IS NULL", cfg.binding.ClassId, cfg.table.Table).
		Find(&objects).Error; err != nil {
		return nil, err
	}
	orphanKeys := make([]string, 0)
	for _, o := range objects {
		if !keys[o.BizKey] {
			orphanKeys = append(orphanKeys, o.BizKey)
			if err := global.GVA_DB.Model(&ontology.OntObject{}).Where("id = ?", o.ID).
				Updates(map[string]interface{}{"state": 3, "updated_by": operator}).Error; err != nil {
				return nil, err
			}
		}
	}
	return orphanKeys, nil
}

// ── 行组装辅助（dry-run 与同步复用） ─────────────────────

// buildCode 编码生成：codeColumn 列值优先，空则 "{类名小写}-{biz_key}"（禁 name 列由生效校验链拦截）
func (s *ExtSyncService) buildCode(cfg *syncConfig, row []interface{}, res *RowsResult, bizKey string) string {
	if cfg.binding.CodeColumn != "" {
		if v := trimValue(res.ValueAt(row, cfg.binding.CodeColumn)); v != "" {
			return v
		}
	}
	return strings.ToLower(cfg.class.LocalName) + "-" + bizKey
}

func (s *ExtSyncService) buildName(cfg *syncConfig, row []interface{}, res *RowsResult, fallback string) string {
	if cfg.binding.NameColumn != "" {
		if v := trimValue(res.ValueAt(row, cfg.binding.NameColumn)); v != "" {
			return v
		}
	}
	return fallback
}

// rawValue 取属性原始值：主表列 / 宽表子表列 / 窄表 identifier 匹配
func (s *ExtSyncService) rawValue(cfg *syncConfig, p ontology.OntExtBindingProperty, row []interface{}, res *RowsResult, rowMaps map[uint]*detailMap) (raw interface{}, dataType, source string) {
	if p.DetailBindingId == nil {
		return res.ValueAt(row, p.BizColumn), cfg.mainCols[p.BizColumn], "主表"
	}
	dm := rowMaps[*p.DetailBindingId]
	if dm == nil {
		return nil, "", "子表"
	}
	d := s.detailById(cfg, *p.DetailBindingId)
	detailTable := cfg.detailTables[d.TableId]
	bizKey := trimValue(res.ValueAt(row, cfg.binding.KeyColumn))
	if dm.kind == 2 {
		if v, ok := dm.narrow[bizKey+"\x00"+p.BizColumn]; ok {
			return v, cfg.detailCols[d.ID][d.ValueColumn], "窄表:" + detailTable.Table + "(" + p.BizColumn + ")"
		}
		return nil, cfg.detailCols[d.ID][d.ValueColumn], "窄表:" + detailTable.Table + "(" + p.BizColumn + ")"
	}
	if rowMap, ok := dm.wide[bizKey]; ok {
		return rowMap[p.BizColumn], cfg.detailCols[d.ID][p.BizColumn], "宽表:" + detailTable.Table
	}
	return nil, "", "宽表:" + detailTable.Table
}

func (s *ExtSyncService) detailById(cfg *syncConfig, id uint) *ontology.OntExtBindingDetail {
	for i := range cfg.details {
		if cfg.details[i].ID == id {
			return &cfg.details[i]
		}
	}
	return nil
}

// ── xsd ↔ PG 兼容组与值转换（附录A） ─────────────────────

func xsdGroupOf(xsdType string) string {
	switch xsdType {
	case "xsd:string":
		return "string"
	case "xsd:integer":
		return "integer"
	case "xsd:decimal":
		return "decimal"
	case "xsd:boolean":
		return "boolean"
	case "xsd:datetime":
		return "datetime"
	default:
		return "any"
	}
}

// compatGroupMatch xsdType 与 PG dataType 是否兼容（xsdType 空/未知 → any 宽松放行）
func compatGroupMatch(xsdType, dataType string) bool {
	group := xsdGroupOf(xsdType)
	if group == "any" || dataType == "" {
		return true
	}
	switch group {
	case "string":
		return map[string]bool{"character varying": true, "character": true, "text": true, "jsonb": true, "uuid": true}[dataType]
	case "integer":
		return map[string]bool{"smallint": true, "integer": true, "bigint": true}[dataType]
	case "decimal":
		return map[string]bool{"numeric": true, "real": true, "double precision": true, "money": true}[dataType]
	case "boolean":
		return map[string]bool{"boolean": true, "smallint": true}[dataType]
	case "datetime":
		return map[string]bool{"timestamp without time zone": true, "timestamp with time zone": true, "date": true}[dataType]
	}
	return false
}

// convertValue 原始值 → JSON 标量字符串（不兼容/解析失败返回 ok=false）
func convertValue(raw interface{}, dataType, xsdType string) (string, bool) {
	if raw == nil {
		return "", true
	}
	group := xsdGroupOf(xsdType)
	if group == "any" {
		return jsonString(fmt.Sprint(raw)), true
	}
	switch group {
	case "string":
		if !compatGroupMatch(xsdType, dataType) && dataType != "" {
			return "", false
		}
		if dataType == "jsonb" {
			b, err := json.Marshal(raw)
			if err != nil {
				return jsonString(fmt.Sprint(raw)), true
			}
			return string(b), true
		}
		return jsonString(trimValue(raw)), true
	case "integer":
		if !compatGroupMatch(xsdType, dataType) {
			return "", false
		}
		switch v := raw.(type) {
		case int64:
			return strconv.FormatInt(v, 10), true
		case int:
			return strconv.Itoa(v), true
		case float64:
			return strconv.FormatInt(int64(v), 10), true
		case []byte:
			n, err := strconv.ParseInt(strings.TrimSpace(string(v)), 10, 64)
			if err != nil {
				return "", false
			}
			return strconv.FormatInt(n, 10), true
		case string:
			n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			if err != nil {
				return "", false
			}
			return strconv.FormatInt(n, 10), true
		}
		return "", false
	case "decimal":
		if !compatGroupMatch(xsdType, dataType) {
			return "", false
		}
		switch v := raw.(type) {
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64), true
		case int64:
			return strconv.FormatInt(v, 10), true
		case []byte:
			f, err := strconv.ParseFloat(strings.TrimSpace(string(v)), 64)
			if err != nil {
				return "", false
			}
			return strconv.FormatFloat(f, 'f', -1, 64), true
		case string:
			f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err != nil {
				return "", false
			}
			return strconv.FormatFloat(f, 'f', -1, 64), true
		}
		return "", false
	case "boolean":
		if !compatGroupMatch(xsdType, dataType) {
			return "", false
		}
		switch v := raw.(type) {
		case bool:
			return strconv.FormatBool(v), true
		case int64:
			return strconv.FormatBool(v != 0), true
		case []byte:
			return parseBoolText(string(v))
		case string:
			return parseBoolText(v)
		}
		return "", false
	case "datetime":
		if !compatGroupMatch(xsdType, dataType) {
			return "", false
		}
		if t, ok := raw.(time.Time); ok {
			return jsonString(t.Format(time.RFC3339)), true
		}
		return jsonString(trimValue(raw)), true
	}
	return "", false
}

func parseBoolText(v string) (string, bool) {
	v = strings.TrimSpace(strings.ToLower(v))
	switch v {
	case "true", "t", "1":
		return "true", true
	case "false", "f", "0":
		return "false", true
	}
	return "", false
}

// ── 值/标签小工具 ────────────────────────────────────────

func trimValue(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case []byte:
		return strings.TrimSpace(string(t))
	case string:
		return strings.TrimSpace(t)
	case time.Time:
		return t.Format(time.RFC3339)
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func asTime(v interface{}) *time.Time {
	if t, ok := v.(time.Time); ok {
		return &t
	}
	if s := trimValue(v); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return &t
		}
	}
	return nil
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func dpLabel(cfg *syncConfig, propertyId uint) string {
	if p, ok := cfg.dtProps[propertyId]; ok {
		if p.Label != "" {
			return p.Label
		}
		return p.LocalName
	}
	return strconv.FormatUint(uint64(propertyId), 10)
}

func opLabel(cfg *syncConfig, propertyId uint) string {
	if p, ok := cfg.objProps[propertyId]; ok {
		if p.Label != "" {
			return p.Label
		}
		return p.LocalName
	}
	return strconv.FormatUint(uint64(propertyId), 10)
}

func ptrEqual(a, b *uint) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
