package ontology

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
	ontReq "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/request"
	ontRes "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/response"
)

// ExtSyncService 同步引擎：dry-run 复用同一「配置加载→取数→行组装」路径，差异仅在写侧短路
type ExtSyncService struct{}

const syncBatchSize = 500

// syncConfig 同步配置加载（绑定 + 主表 + 子表 + 启用属性绑定 + 关键映射）
type syncConfig struct {
	binding      ontology.OntExtBinding
	table        ontology.OntExtTable
	class        ontology.OntModelClass
	details      []ontology.OntExtBindingDetail
	properties   []ontology.OntExtBindingProperty
	detailTables map[uint]ontology.OntExtTable              // detail.ID → 注册表
	dtProps      map[uint]ontology.OntModelDatatypeProperty // propertyId → 数据属性
	objProps     map[uint]ontology.OntModelObjectProperty   // propertyId → 对象属性
	// range 类 → 其生效绑定的 (biz_table)；关系解析用
	rangeActiveTable map[uint]string
	// 列缓存（白名单校验时落下的 dataType 映射）
	mainCols   map[string]string
	detailCols map[uint]map[string]string
}

// loadSyncConfig 加载并校验同步配置（主表+子表白名单校验）
func (s *ExtSyncService) loadSyncConfig(bindingId uint) (*syncConfig, error) {
	var b ontology.OntExtBinding
	if err := global.GVA_DB.First(&b, bindingId).Error; err != nil {
		return nil, errors.New("绑定不存在")
	}
	var table ontology.OntExtTable
	if err := global.GVA_DB.First(&table, b.TableId).Error; err != nil {
		return nil, errors.New("主表注册不存在")
	}
	var cls ontology.OntModelClass
	if err := global.GVA_DB.First(&cls, b.ClassId).Error; err != nil {
		return nil, errors.New("本体类不存在")
	}
	var details []ontology.OntExtBindingDetail
	if err := global.GVA_DB.Where("binding_id = ?", bindingId).Order("sort_order ASC, id ASC").Find(&details).Error; err != nil {
		return nil, err
	}
	var properties []ontology.OntExtBindingProperty
	if err := global.GVA_DB.Where("binding_id = ? AND enabled = 1", bindingId).Order("sort_order ASC, id ASC").Find(&properties).Error; err != nil {
		return nil, err
	}
	cfg := &syncConfig{
		binding: b, table: table, class: cls, details: details, properties: properties,
		detailTables:     map[uint]ontology.OntExtTable{},
		dtProps:          map[uint]ontology.OntModelDatatypeProperty{},
		objProps:         map[uint]ontology.OntModelObjectProperty{},
		rangeActiveTable: map[uint]string{},
	}
	detailTableIds := make([]uint, 0, len(details))
	for _, d := range details {
		detailTableIds = append(detailTableIds, d.TableId)
	}
	if len(detailTableIds) > 0 {
		var tables []ontology.OntExtTable
		if err := global.GVA_DB.Where("id IN ?", detailTableIds).Find(&tables).Error; err != nil {
			return nil, err
		}
		for _, t := range tables {
			cfg.detailTables[t.ID] = t
		}
	}
	dtIds, objIds, rangeClassIds := []uint{}, []uint{}, []uint{}
	for _, p := range properties {
		if p.BindingType == 2 {
			objIds = append(objIds, p.PropertyId)
		} else {
			dtIds = append(dtIds, p.PropertyId)
		}
	}
	if len(dtIds) > 0 {
		var rows []ontology.OntModelDatatypeProperty
		if err := global.GVA_DB.Where("id IN ?", dtIds).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, r := range rows {
			cfg.dtProps[r.ID] = r
		}
	}
	if len(objIds) > 0 {
		var rows []ontology.OntModelObjectProperty
		if err := global.GVA_DB.Where("id IN ?", objIds).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, r := range rows {
			cfg.objProps[r.ID] = r
			if r.RangeClassId != nil {
				rangeClassIds = append(rangeClassIds, *r.RangeClassId)
			}
		}
	}
	// range 类生效绑定 → 业务表名（关系解析定位目标对象）
	if len(rangeClassIds) > 0 {
		var activeBindings []ontology.OntExtBinding
		if err := global.GVA_DB.Where("class_id IN ? AND binding_status = 1 AND deleted_at IS NULL", rangeClassIds).
			Find(&activeBindings).Error; err != nil {
			return nil, err
		}
		activeTableIds := make([]uint, 0, len(activeBindings))
		for _, ab := range activeBindings {
			activeTableIds = append(activeTableIds, ab.TableId)
			cfg.rangeActiveTable[ab.ClassId] = "" // 先占位
		}
		var activeTables []ontology.OntExtTable
		if len(activeTableIds) > 0 {
			if err := global.GVA_DB.Where("id IN ?", activeTableIds).Find(&activeTables).Error; err != nil {
				return nil, err
			}
		}
		tableById := map[uint]string{}
		for _, t := range activeTables {
			tableById[t.ID] = t.Table
		}
		for _, ab := range activeBindings {
			cfg.rangeActiveTable[ab.ClassId] = tableById[ab.TableId]
		}
	}
	// 主表 + 全部子表白名单校验（列缓存复用于值转换）
	cfg.mainCols = map[string]string{}
	cfg.detailCols = map[uint]map[string]string{}
	mainCols, err := extDynamicQuerySvc.AssertTableValid(cfg.table.Table)
	if err != nil {
		return nil, fmt.Errorf("主表白名单校验失败: %w", err)
	}
	cfg.mainCols = mainCols
	for _, d := range cfg.details {
		t := cfg.detailTables[d.TableId]
		cols, cErr := extDynamicQuerySvc.AssertTableValid(t.Table)
		if cErr != nil {
			return nil, fmt.Errorf("子表白名单校验失败: %w", cErr)
		}
		cfg.detailCols[d.ID] = cols
	}
	return cfg, nil
}

// DryRunExtBinding 试运行（任意状态可试运行；完全只读，零写入零日志）
func (s *ExtSyncService) DryRunExtBinding(bindingId uint, limit int) (ontRes.DryRunResult, error) {
	result := ontRes.DryRunResult{Rows: make([]ontRes.DryRunRow, 0), Warnings: make([]string, 0)}
	if limit <= 0 {
		limit = 20
	}
	cfg, err := s.loadSyncConfig(bindingId)
	if err != nil {
		return result, err
	}
	// 静态配置 warnings（取数前生成）
	if cfg.binding.CodeColumn == "" {
		result.Warnings = append(result.Warnings, "未配置编码列，将生成「类名-主键」形式编码")
	}
	mainCols, _ := extDynamicQuerySvc.AssertTableValid(cfg.table.Table)
	for _, p := range cfg.properties {
		if p.BindingType != 1 {
			// 对象属性：目标类无生效绑定提示
			op := cfg.objProps[p.PropertyId]
			if op.RangeClassId == nil || cfg.rangeActiveTable[*op.RangeClassId] == "" {
				result.Warnings = append(result.Warnings, "对象属性「"+opLabel(cfg, p.PropertyId)+"」的目标类无生效绑定，关系将无法解析")
			}
			continue
		}
		dp := cfg.dtProps[p.PropertyId]
		if p.DetailBindingId == nil {
			dataType := mainCols[p.BizColumn]
			if dataType != "" && !compatGroupMatch(dp.XsdType, dataType) {
				result.Warnings = append(result.Warnings, "属性「"+dpLabel(cfg, p.PropertyId)+"」的列 "+p.BizColumn+" 类型("+dataType+")与 "+dp.XsdType+" 不兼容，将转换失败记 issue")
			}
		} else if d := s.detailById(cfg, *p.DetailBindingId); d != nil && d.DetailKind == 1 {
			dt := cfg.detailTables[d.TableId]
			cols, _ := extDynamicQuerySvc.AssertTableValid(dt.Table)
			dataType := cols[p.BizColumn]
			if dataType != "" && !compatGroupMatch(dp.XsdType, dataType) {
				result.Warnings = append(result.Warnings, "属性「"+dpLabel(cfg, p.PropertyId)+"」的子表列 "+p.BizColumn+" 类型与 "+dp.XsdType+" 不兼容，将转换失败记 issue")
			}
		}
	}

	// 取前 limit 行（全量口径；dry-run 不推进水位）
	plan, err := s.buildSelectPlan(cfg, false, nil)
	if err != nil {
		return result, err
	}
	res, err := extDynamicQuerySvc.SelectRows(cfg.table.Table, plan.cols, plan.conds, plan.args, cfg.table.PkColumn, limit, 0)
	if err != nil {
		return result, fmt.Errorf("业务表读取失败: %w", err)
	}
	// 子表数据（dry-run 批量取）
	rowMaps, err := s.loadDetailMaps(cfg, res)
	if err != nil {
		return result, err
	}
	for _, row := range res.Rows {
		bizKey := trimValue(res.ValueAt(row, cfg.binding.KeyColumn))
		preview := ontRes.DryRunRow{BizKey: bizKey, Values: make([]ontRes.DryRunValue, 0), Relations: make([]ontRes.DryRunRelation, 0)}
		preview.PreviewCode = s.buildCode(cfg, row, res, bizKey)
		preview.PreviewName = s.buildName(cfg, row, res, preview.PreviewCode)
		for _, p := range cfg.properties {
			if p.BindingType != 1 {
				// 关系解析（只读反查目标对象）
				op := cfg.objProps[p.PropertyId]
				rel := ontRes.DryRunRelation{Label: opLabel(cfg, p.PropertyId), BizColumnValue: trimValue(res.ValueAt(row, p.BizColumn)), Resolved: false}
				if op.RangeClassId != nil {
					if targetTable := cfg.rangeActiveTable[*op.RangeClassId]; targetTable != "" {
						var target ontology.OntObject
						if e := global.GVA_DB.Where("class_id = ? AND biz_table = ? AND biz_key = ? AND deleted_at IS NULL",
							*op.RangeClassId, targetTable, rel.BizColumnValue).First(&target).Error; e == nil {
							rel.Resolved = true
							rel.TargetObjectName = target.ObjectName
						}
					}
				}
				if rel.TargetObjectName == "" {
					rel.TargetObjectName = "未解析"
				}
				preview.Relations = append(preview.Relations, rel)
				continue
			}
			dp := cfg.dtProps[p.PropertyId]
			raw, dataType, source := s.rawValue(cfg, p, row, res, rowMaps)
			_, ok := convertValue(raw, dataType, dp.XsdType)
			preview.Values = append(preview.Values, ontRes.DryRunValue{
				Label: dpLabel(cfg, p.PropertyId), Source: source,
				RawValue: fmt.Sprint(raw), Converted: ok,
			})
		}
		result.Rows = append(result.Rows, preview)
	}
	return result, nil
}

// TriggerExtSync 触发同步（1全量含孤儿检测 / 2增量水位）；返回 SyncResult 并落日志
func (s *ExtSyncService) TriggerExtSync(req ontReq.TriggerSyncOps, operator string) (ontRes.SyncResult, error) {
	start := time.Now()
	result := ontRes.SyncResult{BindingId: req.BindingId, TriggerType: 2, SyncScope: req.Scope, SkipReasons: make([]string, 0)}
	cfg, err := s.loadSyncConfig(req.BindingId)
	if err != nil {
		return result, err
	}
	result.ClassId = cfg.binding.ClassId
	if cfg.binding.BindingStatus != 1 {
		return result, errors.New("绑定未生效，无法同步")
	}
	if err := modelProjectSvc.AssertProjectWritable(cfg.binding.ProjectId); err != nil {
		return result, err
	}
	// 互斥：绑定行 FOR UPDATE（覆盖取数+行处理窗口；收尾写之前释放，避免与自身收尾 UPDATE 争锁自死锁）
	lockTx := global.GVA_DB.Begin()
	if lockTx.Error != nil {
		return result, lockTx.Error
	}
	lockCommitted := false
	defer func() {
		if !lockCommitted {
			_ = lockTx.Rollback()
		}
	}()
	var lockedId uint
	if err := lockTx.Raw("SELECT id FROM ont_ext_bindings WHERE id = ? FOR UPDATE", req.BindingId).Row().Scan(&lockedId); err != nil {
		lockTx.Rollback()
		return result, errors.New("绑定加锁失败")
	}

	// 取数计划：增量 → WHERE 水位列 > watermark；不支持增量强制全量
	incremental := req.Scope == 2
	warnings := make([]string, 0)
	if incremental && cfg.table.SupportsIncremental != 1 {
		incremental = false
		warnings = append(warnings, "表不支持增量（无水位列），已按全量执行")
	}
	if incremental && cfg.binding.Watermark == nil {
		warnings = append(warnings, "无历史水位，本次按全量扫描")
		incremental = false
	}
	plan, err := s.buildSelectPlan(cfg, incremental, cfg.binding.Watermark)
	if err != nil {
		return result, err
	}

	// 宽/窄子表数据按批加载进内存 map（join值 → 行；窄表 (join值,identifier) → 值）
	var maxWatermark *time.Time
	skipReasons := make([]string, 0)
	offset := 0
	for {
		res, err := extDynamicQuerySvc.SelectRows(cfg.table.Table, plan.cols, plan.conds, plan.args, cfg.table.PkColumn, syncBatchSize, offset)
		if err != nil {
			return result, fmt.Errorf("业务表读取失败: %w", err)
		}
		if len(res.Rows) == 0 {
			break
		}
		rowMaps, err := s.loadDetailMaps(cfg, res)
		if err != nil {
			return result, err
		}
		if t := batchMaxWatermark(cfg, res); t != nil {
			if maxWatermark == nil || t.After(*maxWatermark) {
				maxWatermark = t
			}
		}
		for _, row := range res.Rows {
			s.processRow(cfg, row, res, rowMaps, &result, &skipReasons, operator)
		}
		result.TotalRows += len(res.Rows)
		if len(res.Rows) < syncBatchSize {
			break
		}
		offset += syncBatchSize
	}

	// 全量收尾：孤儿检测（不带状态过滤的 key 全集 vs 已物化对象）
	if req.Scope == 1 {
		orphans, err := s.detectOrphans(cfg, plan, operator)
		if err != nil {
			return result, err
		}
		result.IssuesCount += len(orphans)
		for _, o := range orphans {
			skipReasons = append(skipReasons, "孤儿: biz_key="+o)
		}
	}

	// 释放互斥锁（行处理完成，收尾写不再占用绑定行锁）
	if err := lockTx.Commit().Error; err != nil {
		return result, err
	}
	lockCommitted = true

	// 水位推进（增量；空批不回退水位）
	if req.Scope == 2 && maxWatermark != nil {
		if err := global.GVA_DB.Model(&ontology.OntExtBinding{}).Where("id = ?", req.BindingId).
			Update("watermark", maxWatermark).Error; err != nil {
			return result, err
		}
	}
	// 回写增量模式下的水位（全量亦更新为本次扫描最大值，后续可切增量）
	if req.Scope == 1 && maxWatermark != nil && cfg.table.SupportsIncremental == 1 {
		_ = global.GVA_DB.Model(&ontology.OntExtBinding{}).Where("id = ?", req.BindingId).
			Update("watermark", maxWatermark).Error
	}

	// status 判定 + 日志 + 绑定摘要
	result.DurationMs = int(time.Since(start).Milliseconds())
	switch {
	case result.Failed > 0 && result.Failed == result.TotalRows:
		result.Status = 3
	case result.Failed > 0 || result.IssuesCount > 0:
		result.Status = 2
	default:
		result.Status = 1
	}
	summary := fmt.Sprintf("创建%d/更新%d/跳过%d/失败%d/问题%d", result.Created, result.Updated, result.Skipped, result.Failed, result.IssuesCount)
	for _, w := range warnings {
		summary += "；" + w
	}
	if len(summary) > 500 {
		summary = summary[:500]
	}
	result.SkipReasons = skipReasons
	reasonsJson, _ := json.Marshal(skipReasons)
	now := time.Now()
	log := ontology.OntExtSyncLog{
		BindingId: req.BindingId, ClassId: result.ClassId, TriggerType: 2, SyncScope: req.Scope,
		TotalRows: result.TotalRows, Created: result.Created, Updated: result.Updated,
		Skipped: result.Skipped, Failed: result.Failed, IssuesCount: result.IssuesCount,
		Status: result.Status, DurationMs: result.DurationMs, Summary: summary,
		SkipReasons: string(reasonsJson), StartTime: start, EndTime: &now,
	}
	if err := global.GVA_DB.Create(&log).Error; err != nil {
		return result, err
	}
	if err := global.GVA_DB.Model(&ontology.OntExtBinding{}).Where("id = ?", req.BindingId).
		Updates(map[string]interface{}{"last_sync_time": now, "last_sync_summary": summary}).Error; err != nil {
		return result, err
	}
	return result, nil
}

// GetExtSyncLogList 同步日志分页
func (s *ExtSyncService) GetExtSyncLogList(info ontReq.SearchExtSyncLog) (list []ontology.OntExtSyncLog, total int64, err error) {
	db := global.GVA_DB.Model(&ontology.OntExtSyncLog{})
	if info.BindingId > 0 {
		db = db.Where("binding_id = ?", info.BindingId)
	}
	if info.ClassId > 0 {
		db = db.Where("class_id = ?", info.ClassId)
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Order("start_time DESC").Limit(info.PageSize).Offset((info.Page - 1) * info.PageSize).Find(&list).Error
	return
}
