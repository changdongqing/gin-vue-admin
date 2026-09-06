package ontology

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
	ontReq "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/request"
	ontRes "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/response"
	"gorm.io/gorm"
)

type ExtBindingService struct{}

// CreateExtBinding 保存草稿：归档 → 类可实例化 → 表/子表已注册 → 属性属类 → 同事务插入
func (s *ExtBindingService) CreateExtBinding(req *ontReq.SaveExtBinding, operator string) (uint, error) {
	if err := s.validateDraft(req); err != nil {
		return 0, err
	}
	b := &req.OntExtBinding
	b.BindingStatus = 0 // 保存即草稿态，生效需显式操作
	b.ID = 0
	b.CreatedBy, b.UpdatedBy = operator, operator
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(b).Error; err != nil {
			return err
		}
		for i := range req.Details {
			req.Details[i].BindingId = b.ID
			req.Details[i].ID = 0
		}
		if len(req.Details) > 0 {
			if err := tx.Create(&req.Details).Error; err != nil {
				return err
			}
		}
		for i := range req.Properties {
			req.Properties[i].BindingId = b.ID
			req.Properties[i].ID = 0
		}
		if len(req.Properties) > 0 {
			if err := tx.Create(&req.Properties).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return b.ID, nil
}

// UpdateExtBinding 更新草稿（生效态不可改，先停用）→ 删旧插新整体替换
func (s *ExtBindingService) UpdateExtBinding(req *ontReq.SaveExtBinding, operator string) error {
	var old ontology.OntExtBinding
	if err := global.GVA_DB.First(&old, req.ID).Error; err != nil {
		return errors.New("绑定不存在")
	}
	if old.BindingStatus == 1 {
		return errors.New("生效绑定不可编辑，请先停用")
	}
	if err := s.validateDraft(req); err != nil {
		return err
	}
	b := &req.OntExtBinding
	b.BindingStatus = old.BindingStatus
	b.UpdatedBy = operator
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&ontology.OntExtBinding{}).Where("id = ?", old.ID).Updates(map[string]interface{}{
			"project_id": b.ProjectId, "class_id": b.ClassId, "table_id": b.TableId,
			"key_column": b.KeyColumn, "code_column": b.CodeColumn, "name_column": b.NameColumn,
			"parent_column": b.ParentColumn, "status_column": b.StatusColumn, "status_active_value": b.StatusActiveValue,
			"sync_mode": b.SyncMode, "sync_order": b.SyncOrder, "include_disabled": b.IncludeDisabled,
			"conflict_strategy": b.ConflictStrategy, "missing_target_policy": b.MissingTargetPolicy,
			"telemetry_enabled": b.TelemetryEnabled, "remark": b.Remark, "updated_by": operator,
		}).Error; err != nil {
			return err
		}
		// 整体替换：子表/属性行物理删旧插新（软删会残留孤儿）
		if err := tx.Unscoped().Where("binding_id = ?", old.ID).
			Delete(&ontology.OntExtBindingDetail{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("binding_id = ?", old.ID).
			Delete(&ontology.OntExtBindingProperty{}).Error; err != nil {
			return err
		}
		for i := range req.Details {
			req.Details[i].BindingId = old.ID
			req.Details[i].ID = 0
		}
		if len(req.Details) > 0 {
			if err := tx.Create(&req.Details).Error; err != nil {
				return err
			}
		}
		for i := range req.Properties {
			req.Properties[i].BindingId = old.ID
			req.Properties[i].ID = 0
		}
		if len(req.Properties) > 0 {
			if err := tx.Create(&req.Properties).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// validateDraft 草稿校验（创建/更新共用）
func (s *ExtBindingService) validateDraft(req *ontReq.SaveExtBinding) error {
	b := &req.OntExtBinding
	if err := modelProjectSvc.AssertProjectWritable(b.ProjectId); err != nil {
		return err
	}
	var cls ontology.OntModelClass
	if err := global.GVA_DB.First(&cls, b.ClassId).Error; err != nil {
		return errors.New("本体类不存在")
	}
	if cls.ProjectId != b.ProjectId {
		return errors.New("类不属于该项目")
	}
	if cls.IsInstantiable != 1 {
		return errors.New("类未开启实例化")
	}
	var table ontology.OntExtTable
	if err := global.GVA_DB.First(&table, b.TableId).Error; err != nil {
		return errors.New("主表未注册")
	}
	detailTableIds := make([]uint, 0, len(req.Details))
	for _, d := range req.Details {
		detailTableIds = append(detailTableIds, d.TableId)
	}
	if len(detailTableIds) > 0 {
		var cnt int64
		if err := global.GVA_DB.Model(&ontology.OntExtTable{}).
			Where("id IN ?", detailTableIds).Count(&cnt).Error; err != nil {
			return err
		}
		if int(cnt) != len(detailTableIds) {
			return errors.New("子表未注册")
		}
	}
	// 属性绑定：property_id 必须属于该类（数据=class_id / 对象=domain_class_id）
	dtIds, objIds := []uint{}, []uint{}
	for _, p := range req.Properties {
		if p.BindingType == 2 {
			objIds = append(objIds, p.PropertyId)
		} else {
			dtIds = append(dtIds, p.PropertyId)
		}
	}
	if len(dtIds) > 0 {
		var cnt int64
		if err := global.GVA_DB.Model(&ontology.OntModelDatatypeProperty{}).
			Where("id IN ? AND class_id = ?", dtIds, b.ClassId).Count(&cnt).Error; err != nil {
			return err
		}
		if int(cnt) != len(dtIds) {
			return errors.New("属性绑定包含不属于该类的数据属性")
		}
	}
	if len(objIds) > 0 {
		var cnt int64
		if err := global.GVA_DB.Model(&ontology.OntModelObjectProperty{}).
			Where("id IN ? AND domain_class_id = ?", objIds, b.ClassId).Count(&cnt).Error; err != nil {
			return err
		}
		if int(cnt) != len(objIds) {
			return errors.New("属性绑定包含不属于该类的对象属性")
		}
	}
	return nil
}

// GetExtBinding 绑定详情（主表 + 双子表清单，编辑回填）
func (s *ExtBindingService) GetExtBinding(id uint) (binding ontology.OntExtBinding, details []ontology.OntExtBindingDetail, properties []ontology.OntExtBindingProperty, err error) {
	if err = global.GVA_DB.First(&binding, id).Error; err != nil {
		return binding, nil, nil, errors.New("绑定不存在")
	}
	if err = global.GVA_DB.Where("binding_id = ?", id).Order("sort_order ASC, id ASC").Find(&details).Error; err != nil {
		return
	}
	err = global.GVA_DB.Where("binding_id = ?", id).Order("sort_order ASC, id ASC").Find(&properties).Error
	return
}

// GetExtBindingList 分页 + 富化（项目编码/类信息/表名/行数计数）
func (s *ExtBindingService) GetExtBindingList(info ontReq.SearchExtBinding) (list []ontRes.ExtBindingPageItem, total int64, err error) {
	db := global.GVA_DB.Model(&ontology.OntExtBinding{})
	if info.ProjectId > 0 {
		db = db.Where("project_id = ?", info.ProjectId)
	}
	if info.ClassId > 0 {
		db = db.Where("class_id = ?", info.ClassId)
	}
	if info.BindingStatus > 0 {
		db = db.Where("binding_status = ?", info.BindingStatus)
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	var bindings []ontology.OntExtBinding
	if err = db.Order("sync_order ASC, id DESC").Limit(info.PageSize).Offset((info.Page - 1) * info.PageSize).Find(&bindings).Error; err != nil {
		return
	}
	list = make([]ontRes.ExtBindingPageItem, 0, len(bindings))
	if len(bindings) == 0 {
		return
	}
	// 批量 IN 富化（无 N+1）
	projectIds, classIds, tableIds, bindingIds := []uint{}, []uint{}, []uint{}, []uint{}
	for _, b := range bindings {
		projectIds = append(projectIds, b.ProjectId)
		classIds = append(classIds, b.ClassId)
		tableIds = append(tableIds, b.TableId)
		bindingIds = append(bindingIds, b.ID)
	}
	projectNames := map[uint]string{}
	var projects []ontology.OntModelProject
	if err = global.GVA_DB.Where("id IN ?", projectIds).Find(&projects).Error; err != nil {
		return
	}
	for _, p := range projects {
		projectNames[p.ID] = p.ProjectCode
	}
	classes := map[uint]ontology.OntModelClass{}
	var classRows []ontology.OntModelClass
	if err = global.GVA_DB.Where("id IN ?", classIds).Find(&classRows).Error; err != nil {
		return
	}
	for _, c := range classRows {
		classes[c.ID] = c
	}
	tableNames := map[uint]string{}
	var tables []ontology.OntExtTable
	if err = global.GVA_DB.Where("id IN ?", tableIds).Find(&tables).Error; err != nil {
		return
	}
	for _, t := range tables {
		tableNames[t.ID] = t.Table
	}
	type cntRow struct {
		BindingId uint
		Cnt       int64
	}
	propCounts := map[uint]int64{}
	var propCnt []cntRow
	if err = global.GVA_DB.Model(&ontology.OntExtBindingProperty{}).
		Select("binding_id, COUNT(*) AS cnt").Where("binding_id IN ? AND deleted_at IS NULL", bindingIds).
		Group("binding_id").Scan(&propCnt).Error; err != nil {
		return
	}
	for _, r := range propCnt {
		propCounts[r.BindingId] = r.Cnt
	}
	detailCounts := map[uint]int64{}
	var detailCnt []cntRow
	if err = global.GVA_DB.Model(&ontology.OntExtBindingDetail{}).
		Select("binding_id, COUNT(*) AS cnt").Where("binding_id IN ? AND deleted_at IS NULL", bindingIds).
		Group("binding_id").Scan(&detailCnt).Error; err != nil {
		return
	}
	for _, r := range detailCnt {
		detailCounts[r.BindingId] = r.Cnt
	}
	for _, b := range bindings {
		item := ontRes.ExtBindingPageItem{
			OntExtBinding: b,
			ProjectCode:   projectNames[b.ProjectId],
			TableName:     tableNames[b.TableId],
			PropertyCount: propCounts[b.ID],
			DetailCount:   detailCounts[b.ID],
		}
		if c, ok := classes[b.ClassId]; ok {
			item.ClassLocalName = c.LocalName
			item.ClassLabelCn = c.LabelCn
		}
		list = append(list, item)
	}
	return
}

// ActivateExtBinding 生效校验链（§二.3）：逐条拼接失败原因，任一失败整体拒绝
func (s *ExtBindingService) ActivateExtBinding(id uint) error {
	var b ontology.OntExtBinding
	if err := global.GVA_DB.First(&b, id).Error; err != nil {
		return errors.New("绑定不存在")
	}
	if b.BindingStatus == 1 {
		return nil
	}
	if err := modelProjectSvc.AssertProjectWritable(b.ProjectId); err != nil {
		return err
	}
	reasons := make([]string, 0)
	var table ontology.OntExtTable
	if err := global.GVA_DB.First(&table, b.TableId).Error; err != nil {
		return errors.New("主表注册不存在")
	}
	mainCols, err := extDynamicQuerySvc.AssertTableValid(table.Table)
	if err != nil {
		return fmt.Errorf("主表校验失败: %w", err)
	}
	need := func(col, label string) {
		if col != "" {
			if _, ok := mainCols[col]; !ok {
				reasons = append(reasons, label+"列不存在: "+col)
			}
		}
	}
	if _, ok := mainCols[b.KeyColumn]; !ok {
		reasons = append(reasons, "主键列不存在: "+b.KeyColumn)
	}
	need(b.CodeColumn, "编码")
	need(b.NameColumn, "名称")
	need(b.ParentColumn, "父列")
	need(b.StatusColumn, "状态")
	if b.CodeColumn != "" && b.CodeColumn == b.NameColumn {
		reasons = append(reasons, "编码列与名称列不可为同一列（防唯一索引冲突）")
	}
	if (b.StatusColumn == "") != (b.StatusActiveValue == "") {
		reasons = append(reasons, "状态列与状态在用值须成对配置")
	}
	var details []ontology.OntExtBindingDetail
	if err := global.GVA_DB.Where("binding_id = ?", id).Find(&details).Error; err != nil {
		return err
	}
	detailCols := map[uint]map[string]string{}
	for _, d := range details {
		var dt ontology.OntExtTable
		if err := global.GVA_DB.First(&dt, d.TableId).Error; err != nil {
			reasons = append(reasons, "子表注册不存在: tableId="+fmt.Sprint(d.TableId))
			continue
		}
		cols, cErr := extDynamicQuerySvc.AssertTableValid(dt.Table)
		if cErr != nil {
			reasons = append(reasons, "子表校验失败: "+dt.Table)
			continue
		}
		detailCols[d.ID] = cols
		if _, ok := cols[d.JoinColumn]; !ok {
			reasons = append(reasons, "子表 "+dt.Table+" 回连列不存在: "+d.JoinColumn)
		}
		if d.IdentifierColumn != "" {
			if _, ok := cols[d.IdentifierColumn]; !ok {
				reasons = append(reasons, "子表 "+dt.Table+" 标识列不存在: "+d.IdentifierColumn)
			}
		}
		if d.ValueColumn != "" {
			if _, ok := cols[d.ValueColumn]; !ok {
				reasons = append(reasons, "子表 "+dt.Table+" 值列不存在: "+d.ValueColumn)
			}
		}
	}
	var properties []ontology.OntExtBindingProperty
	if err := global.GVA_DB.Where("binding_id = ? AND enabled = 1", id).Find(&properties).Error; err != nil {
		return err
	}
	for _, p := range properties {
		if p.DetailBindingId == nil {
			if p.BizColumn != "" {
				if _, ok := mainCols[p.BizColumn]; !ok {
					reasons = append(reasons, "属性绑定列不存在(主表): "+p.BizColumn)
				}
			}
			continue
		}
		if cols, ok := detailCols[*p.DetailBindingId]; ok {
			// 窄表来源 bizColumn 为 identifier 匹配值，不校验列；宽表来源校验列
			isNarrow := false
			for _, d := range details {
				if d.ID == *p.DetailBindingId && d.DetailKind == 2 {
					isNarrow = true
					break
				}
			}
			if !isNarrow && p.BizColumn != "" {
				if _, ok := cols[p.BizColumn]; !ok {
					reasons = append(reasons, "属性绑定列不存在(子表): "+p.BizColumn)
				}
			}
		}
	}
	if len(reasons) > 0 {
		return errors.New("生效校验失败: " + strings.Join(reasons, "；"))
	}
	err = global.GVA_DB.Model(&ontology.OntExtBinding{}).Where("id = ?", id).
		Update("binding_status", 1).Error
	if err != nil && strings.Contains(err.Error(), "uk_ont_ext_binding_class_active") {
		return errors.New("该类已存在生效绑定（一个类至多一条）")
	}
	return err
}

// DeactivateExtBinding 停用（直接置停用）
func (s *ExtBindingService) DeactivateExtBinding(id uint) error {
	var b ontology.OntExtBinding
	if err := global.GVA_DB.First(&b, id).Error; err != nil {
		return errors.New("绑定不存在")
	}
	if b.BindingStatus != 1 {
		return errors.New("仅生效绑定可停用")
	}
	return global.GVA_DB.Model(&ontology.OntExtBinding{}).Where("id = ?", id).
		Update("binding_status", 2).Error
}

// DeleteExtBinding 删除绑定（同步日志保留不级联；双子表随删）
func (s *ExtBindingService) DeleteExtBinding(id uint) error {
	var b ontology.OntExtBinding
	if err := global.GVA_DB.First(&b, id).Error; err != nil {
		return errors.New("绑定不存在")
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&ontology.OntExtBinding{}, id).Error; err != nil {
			return err
		}
		if err := tx.Where("binding_id = ?", id).Delete(&ontology.OntExtBindingDetail{}).Error; err != nil {
			return err
		}
		return tx.Where("binding_id = ?", id).Delete(&ontology.OntExtBindingProperty{}).Error
	})
}
