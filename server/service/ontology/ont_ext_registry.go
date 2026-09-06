package ontology

import (
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
	ontReq "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/request"
	ontRes "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/response"
)

type ExtModuleService struct{}

// GetExtModuleList 全量模块 + 注册表计数
func (s *ExtModuleService) GetExtModuleList() (list []ontRes.ExtModulePageItem, err error) {
	var modules []ontology.OntExtModule
	if err = global.GVA_DB.Order("sort_order ASC, id ASC").Find(&modules).Error; err != nil {
		return
	}
	type cntRow struct {
		ModuleId uint
		Cnt      int64
	}
	var counts []cntRow
	if err = global.GVA_DB.Model(&ontology.OntExtTable{}).
		Select("module_id, COUNT(*) AS cnt").
		Where("deleted_at IS NULL").
		Group("module_id").Scan(&counts).Error; err != nil {
		return
	}
	cntMap := make(map[uint]int64, len(counts))
	for _, c := range counts {
		cntMap[c.ModuleId] = c.Cnt
	}
	list = make([]ontRes.ExtModulePageItem, 0, len(modules))
	for _, m := range modules {
		list = append(list, ontRes.ExtModulePageItem{OntExtModule: m, TableCount: cntMap[m.ID]})
	}
	return
}

// CreateExtModule 新增模块（moduleCode 全局唯一）
func (s *ExtModuleService) CreateExtModule(p *ontology.OntExtModule, operator string) error {
	var count int64
	if err := global.GVA_DB.Model(&ontology.OntExtModule{}).
		Where("module_code = ?", p.ModuleCode).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("模块编码已存在")
	}
	p.CreatedBy, p.UpdatedBy = operator, operator
	return global.GVA_DB.Create(p).Error
}

// UpdateExtModule 更新模块（moduleCode 查重排除自身）
func (s *ExtModuleService) UpdateExtModule(p *ontology.OntExtModule, operator string) error {
	var old ontology.OntExtModule
	if err := global.GVA_DB.First(&old, p.ID).Error; err != nil {
		return errors.New("模块不存在")
	}
	var count int64
	if err := global.GVA_DB.Model(&ontology.OntExtModule{}).
		Where("module_code = ? AND id <> ?", p.ModuleCode, p.ID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("模块编码已存在")
	}
	return global.GVA_DB.Model(&ontology.OntExtModule{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
		"module_code":  p.ModuleCode,
		"name":         p.Name,
		"service_name": p.ServiceName,
		"description":  p.Description,
		"status":       p.Status,
		"sort_order":   p.SortOrder,
		"updated_by":   operator,
	}).Error
}

// DeleteExtModule 删除模块（有注册表拒绝）
func (s *ExtModuleService) DeleteExtModule(id uint) error {
	var count int64
	if err := global.GVA_DB.Model(&ontology.OntExtTable{}).
		Where("module_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("模块下存在已注册表，无法删除")
	}
	return global.GVA_DB.Delete(&ontology.OntExtModule{}, id).Error
}

type ExtTableService struct{}

var extDynamicQuerySvc = &ExtDynamicQueryService{}

// GetExtTableList 模块下注册表清单
func (s *ExtTableService) GetExtTableList(moduleId uint) (list []ontology.OntExtTable, err error) { //nolint:stylecheck // 对齐前端
	err = global.GVA_DB.Where("module_id = ?", moduleId).Order("id ASC").Find(&list).Error
	return
}

// ProbeExtTables 探测 public schema 物理表（关键字 ILIKE + 注释 + 已注册标注）
func (s *ExtTableService) ProbeExtTables(keyword string) (list []ontRes.ProbeExtTableItem, err error) {
	type probeRow struct {
		TableName    string
		TableComment string
	}
	kw := "%" + extDynamicQuerySvc.EscapeLike(keyword) + "%"
	rows := []probeRow{}
	q := `SELECT c.relname AS table_name, COALESCE(obj_description(c.oid), '') AS table_comment
	      FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
	      WHERE n.nspname = 'public' AND c.relkind = 'BASE TABLE'`
	args := []interface{}{}
	if keyword != "" {
		q += ` AND c.relname ILIKE ?`
		args = append(args, kw)
	}
	q += ` ORDER BY c.relname`
	if err = global.GVA_DB.Raw(q, args...).Scan(&rows).Error; err != nil {
		return
	}
	var registered []ontology.OntExtTable
	if err = global.GVA_DB.Find(&registered).Error; err != nil {
		return
	}
	regSet := make(map[string]bool, len(registered))
	for _, r := range registered {
		regSet[r.Table] = true
	}
	list = make([]ontRes.ProbeExtTableItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, ontRes.ProbeExtTableItem{
			TableName: r.TableName, TableComment: r.TableComment, Registered: regSet[r.TableName],
		})
	}
	return
}

// GetExtTableColumns 探测表列（列名/dataType/注释）
func (s *ExtTableService) GetExtTableColumns(tableName string) (list []ontRes.ExtTableColumnItem, err error) {
	type colRow struct {
		ColumnName    string
		DataType      string
		ColumnComment string
	}
	rows := []colRow{}
	err = global.GVA_DB.Raw(
		`SELECT c.column_name, c.data_type,
		        COALESCE(col_description(cl.oid, c.ordinal_position), '') AS column_comment
		 FROM information_schema.columns c
		 JOIN pg_class cl ON cl.relname = c.table_name AND cl.relnamespace = 'public'::regnamespace
		 WHERE c.table_schema = 'public' AND c.table_name = ?
		 ORDER BY c.ordinal_position`, tableName).Scan(&rows).Error
	if err != nil {
		return
	}
	if len(rows) == 0 {
		return nil, errors.New("表不存在: " + tableName)
	}
	list = make([]ontRes.ExtTableColumnItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, ontRes.ExtTableColumnItem{ColumnName: r.ColumnName, DataType: r.DataType, ColumnComment: r.ColumnComment})
	}
	return
}

// RegisterExtTables 探测勾选批量注册：已全局注册跳过 → 表存在校验 → 探测回填软删/水位列
func (s *ExtTableService) RegisterExtTables(req ontReq.RegisterExtTablesOps, operator string) (registered int, err error) {
	var module ontology.OntExtModule
	if err = global.GVA_DB.First(&module, req.ModuleId).Error; err != nil {
		return 0, errors.New("模块不存在")
	}
	for _, item := range req.Items {
		var count int64
		if err = global.GVA_DB.Model(&ontology.OntExtTable{}).
			Where("table_name = ?", item.TableName).Count(&count).Error; err != nil {
			return registered, err
		}
		if count > 0 {
			continue // 物理表全局至多注册一次
		}
		cols, cErr := extDynamicQuerySvc.AssertTableValid(item.TableName)
		if cErr != nil {
			return registered, cErr
		}
		pk := "id"
		if _, ok := cols[pk]; !ok {
			for c := range cols { // 无 id 列回退首列
				pk = c
				break
			}
		}
		deletedCol := pickExisting(cols, "deleted_at", "deleted")
		updateTimeCol := pickExisting(cols, "updated_at", "update_time")
		rec := ontology.OntExtTable{
			ModuleId: req.ModuleId, Table: item.TableName, DisplayName: item.DisplayName,
			PkColumn: pk, PkType: cols[pk], DeletedColumn: deletedCol, UpdateTimeColumn: updateTimeCol,
			SupportsIncremental: boolToInt(updateTimeCol != ""), Remark: item.Remark,
			CreatedBy: operator, UpdatedBy: operator,
		}
		if err = global.GVA_DB.Create(&rec).Error; err != nil {
			return registered, err
		}
		registered++
	}
	return registered, nil
}

// DeleteExtTable 删除注册表（被绑定引用拒绝）
func (s *ExtTableService) DeleteExtTable(id uint) error {
	var bCount int64
	if err := global.GVA_DB.Model(&ontology.OntExtBinding{}).
		Where("table_id = ?", id).Count(&bCount).Error; err != nil {
		return err
	}
	var dCount int64
	if err := global.GVA_DB.Model(&ontology.OntExtBindingDetail{}).
		Where("table_id = ?", id).Count(&dCount).Error; err != nil {
		return err
	}
	if bCount+dCount > 0 {
		return errors.New("表已被类绑定引用，无法删除")
	}
	return global.GVA_DB.Delete(&ontology.OntExtTable{}, id).Error
}

func pickExisting(cols map[string]string, candidates ...string) string {
	for _, c := range candidates {
		if _, ok := cols[c]; ok {
			return c
		}
	}
	return ""
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
