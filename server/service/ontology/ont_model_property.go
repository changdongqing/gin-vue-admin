package ontology

import (
	"errors"
	"regexp"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
	ontReq "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/request"
	ontRes "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/response"
	"gorm.io/gorm"
)

type DatatypePropertyService struct{}

// propertyLocalNameRe 属性本地名正则（与类同口径；创建后不可改）
var propertyLocalNameRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

// CreateModelDatatypeProperty 创建数据属性：归档 → 类存在 → localName → 单位校验 → 方案B IRI
func (s *DatatypePropertyService) CreateModelDatatypeProperty(p *ontology.OntModelDatatypeProperty, operator string) error {
	if err := modelProjectSvc.AssertProjectWritable(p.ProjectId); err != nil {
		return err
	}
	var cls ontology.OntModelClass
	if err := global.GVA_DB.First(&cls, p.ClassId).Error; err != nil {
		return errors.New("所属类不存在")
	}
	if cls.ProjectId != p.ProjectId {
		return errors.New("类不属于该项目")
	}
	if !propertyLocalNameRe.MatchString(p.LocalName) {
		return errors.New("本地名不合法（须以字母开头，仅含字母/数字/下划线/连字符）")
	}
	if err := s.checkLocalNameUnique(global.GVA_DB, cls.ID, p.LocalName, 0); err != nil {
		return err
	}
	if p.XsdType == "" {
		p.XsdType = "xsd:string"
	}
	if err := validateUnitBinding(p.XsdType, p.UnitRef); err != nil {
		return err
	}
	var project ontology.OntModelProject
	if err := global.GVA_DB.First(&project, p.ProjectId).Error; err != nil {
		return errors.New("项目不存在")
	}
	p.PropertyIri = BuildIri(project.NamespaceBase, cls.LocalName+"_"+p.LocalName)
	p.TemplateCode = "" // 手建无溯源；模板挂载走 InstantiateFromTemplate
	p.CreatedBy, p.UpdatedBy = operator, operator
	return global.GVA_DB.Create(p).Error
}

// UpdateModelDatatypeProperty 编辑白名单：label/unitRef/enumValues/基数/isIdentifier/sortOrder
// localName/xsdType 创建后不可改（IRI/值域稳定）
func (s *DatatypePropertyService) UpdateModelDatatypeProperty(p *ontology.OntModelDatatypeProperty, operator string) error {
	var old ontology.OntModelDatatypeProperty
	if err := global.GVA_DB.First(&old, p.ID).Error; err != nil {
		return errors.New("数据属性不存在")
	}
	if err := modelProjectSvc.AssertProjectWritable(old.ProjectId); err != nil {
		return err
	}
	if err := validateUnitBinding(old.XsdType, p.UnitRef); err != nil {
		return err
	}
	return global.GVA_DB.Model(&ontology.OntModelDatatypeProperty{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
		"label":           p.Label,
		"unit_ref":        p.UnitRef,
		"enum_values":     p.EnumValues,
		"min_cardinality": p.MinCardinality,
		"max_cardinality": p.MaxCardinality,
		"is_identifier":   p.IsIdentifier,
		"sort_order":      p.SortOrder,
		"updated_by":      operator,
	}).Error
}

// DeleteModelDatatypeProperty 删除数据属性（不做值级联——对象域未建）
func (s *DatatypePropertyService) DeleteModelDatatypeProperty(id uint) error {
	var old ontology.OntModelDatatypeProperty
	if err := global.GVA_DB.First(&old, id).Error; err != nil {
		return errors.New("数据属性不存在")
	}
	if err := modelProjectSvc.AssertProjectWritable(old.ProjectId); err != nil {
		return err
	}
	return global.GVA_DB.Delete(&ontology.OntModelDatatypeProperty{}, id).Error
}

// GetModelDatatypeProperty 详情
func (s *DatatypePropertyService) GetModelDatatypeProperty(id uint) (p ontology.OntModelDatatypeProperty, err error) {
	err = global.GVA_DB.First(&p, id).Error
	return
}

// GetModelDatatypePropertyList 分页
func (s *DatatypePropertyService) GetModelDatatypePropertyList(info ontReq.SearchDatatypeProperty) (list []ontology.OntModelDatatypeProperty, total int64, err error) {
	db := global.GVA_DB.Model(&ontology.OntModelDatatypeProperty{})
	if info.ProjectId > 0 {
		db = db.Where("project_id = ?", info.ProjectId)
	}
	if info.ClassId > 0 {
		db = db.Where("class_id = ?", info.ClassId)
	}
	if info.XsdType != "" {
		db = db.Where("xsd_type = ?", info.XsdType)
	}
	if info.TemplateCode != "" {
		db = db.Where("template_code = ?", info.TemplateCode)
	}
	if info.Keyword != "" {
		kw := "%" + info.Keyword + "%"
		db = db.Where("local_name LIKE ? OR label LIKE ?", kw, kw)
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Order("sort_order ASC, id ASC").Limit(info.PageSize).Offset((info.Page - 1) * info.PageSize).Find(&list).Error
	return
}

// InstantiateModelDatatypeProperty 属性级模板挂载：复制单条属性模板（同类已存在按模板编码跳过）
// 返回 skipped=true 表示同类已有该属性（前端计入成功并提示）
func (s *DatatypePropertyService) InstantiateModelDatatypeProperty(req ontReq.InstantiatePropertyOps, operator string) (skipped bool, err error) {
	if err = modelProjectSvc.AssertProjectWritable(req.ProjectId); err != nil {
		return
	}
	var cls ontology.OntModelClass
	if err = global.GVA_DB.First(&cls, req.ClassId).Error; err != nil {
		return false, errors.New("所属类不存在")
	}
	var pt ontology.OntPropertyTemplate
	if err = global.GVA_DB.Where("template_code = ?", req.TemplateCode).First(&pt).Error; err != nil {
		return false, errors.New("属性模板不存在: " + req.TemplateCode)
	}
	var count int64
	if err = global.GVA_DB.Model(&ontology.OntModelDatatypeProperty{}).
		Where("class_id = ? AND local_name = ?", cls.ID, pt.TemplateCode).Count(&count).Error; err != nil {
		return
	}
	if count > 0 {
		return true, nil
	}
	var project ontology.OntModelProject
	if err = global.GVA_DB.First(&project, cls.ProjectId).Error; err != nil {
		return false, errors.New("项目不存在")
	}
	p := ontology.OntModelDatatypeProperty{
		ProjectId:      cls.ProjectId,
		ClassId:        cls.ID,
		PropertyIri:    BuildIri(project.NamespaceBase, cls.LocalName+"_"+pt.TemplateCode),
		LocalName:      pt.TemplateCode,
		Label:          pt.Label,
		TemplateCode:   pt.TemplateCode,
		XsdType:        toXsdType(pt.Type),
		UnitRef:        pt.UnitRef,
		EnumValues:     valuesToJsonArray(pt.Values),
		MinCardinality: 0,
		MaxCardinality: -1,
		IsIdentifier:   pt.IsIdentifier,
		CreatedBy:      operator,
		UpdatedBy:      operator,
	}
	return false, global.GVA_DB.Create(&p).Error
}

func (s *DatatypePropertyService) checkLocalNameUnique(db *gorm.DB, classId uint, localName string, excludeID uint) error {
	var count int64
	q := db.Model(&ontology.OntModelDatatypeProperty{}).Where("class_id = ? AND local_name = ?", classId, localName)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("同类下本地名已存在")
	}
	return nil
}

// validateUnitBinding 单位绑定校验：仅数值型（xsd:integer/xsd:decimal）可绑 unitRef
func validateUnitBinding(xsdType, unitRef string) error {
	if unitRef == "" {
		return nil
	}
	if xsdType != "xsd:integer" && xsdType != "xsd:decimal" {
		return errors.New("仅数值型属性可绑定单位")
	}
	return nil
}

type ObjectPropertyService struct{}

// CreateModelObjectProperty 创建对象属性：domainClassId=所属类自动语义；range 可空
// body 携带 inverseOf 时同事务双向互指（反向关系确认创建入口）
func (s *ObjectPropertyService) CreateModelObjectProperty(p *ontology.OntModelObjectProperty, operator string) error {
	if err := modelProjectSvc.AssertProjectWritable(p.ProjectId); err != nil {
		return err
	}
	var cls ontology.OntModelClass
	if err := global.GVA_DB.First(&cls, p.DomainClassId).Error; err != nil {
		return errors.New("所属类不存在")
	}
	if cls.ProjectId != p.ProjectId {
		return errors.New("类不属于该项目")
	}
	if p.RangeClassId != nil && *p.RangeClassId > 0 {
		var rangeCls ontology.OntModelClass
		if err := global.GVA_DB.First(&rangeCls, *p.RangeClassId).Error; err != nil {
			return errors.New("值域类不存在")
		}
		if rangeCls.ProjectId != p.ProjectId {
			return errors.New("值域类不属于该项目")
		}
	} else {
		p.RangeClassId = nil
	}
	if !propertyLocalNameRe.MatchString(p.LocalName) {
		return errors.New("本地名不合法（须以字母开头，仅含字母/数字/下划线/连字符）")
	}
	if err := s.checkLocalNameUnique(global.GVA_DB, p.DomainClassId, p.LocalName, 0); err != nil {
		return err
	}
	var project ontology.OntModelProject
	if err := global.GVA_DB.First(&project, p.ProjectId).Error; err != nil {
		return errors.New("项目不存在")
	}
	p.PropertyIri = BuildIri(project.NamespaceBase, cls.LocalName+"_"+p.LocalName)
	p.TemplateCode = ""
	p.CreatedBy, p.UpdatedBy = operator, operator
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			return err
		}
		if p.InverseOf != nil && *p.InverseOf > 0 {
			if err := tx.Model(&ontology.OntModelObjectProperty{}).
				Where("id = ?", *p.InverseOf).Update("inverse_of", p.ID).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateModelObjectProperty 编辑白名单：label/rangeClassId/基数/inverseOf/sortOrder
func (s *ObjectPropertyService) UpdateModelObjectProperty(p *ontology.OntModelObjectProperty, operator string) error {
	var old ontology.OntModelObjectProperty
	if err := global.GVA_DB.First(&old, p.ID).Error; err != nil {
		return errors.New("对象属性不存在")
	}
	if err := modelProjectSvc.AssertProjectWritable(old.ProjectId); err != nil {
		return err
	}
	if p.RangeClassId != nil && *p.RangeClassId > 0 {
		var rangeCls ontology.OntModelClass
		if err := global.GVA_DB.First(&rangeCls, *p.RangeClassId).Error; err != nil {
			return errors.New("值域类不存在")
		}
		if rangeCls.ProjectId != old.ProjectId {
			return errors.New("值域类不属于该项目")
		}
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&ontology.OntModelObjectProperty{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
			"label":           p.Label,
			"range_class_id":  p.RangeClassId,
			"min_cardinality": p.MinCardinality,
			"max_cardinality": p.MaxCardinality,
			"inverse_of":      p.InverseOf,
			"sort_order":      p.SortOrder,
			"updated_by":      operator,
		}).Error; err != nil {
			return err
		}
		// inverseOf 互指同步：改指/清空时同步对方侧
		oldInverse := uint(0)
		if old.InverseOf != nil {
			oldInverse = *old.InverseOf
		}
		newInverse := uint(0)
		if p.InverseOf != nil {
			newInverse = *p.InverseOf
		}
		if oldInverse != newInverse {
			if oldInverse > 0 {
				if err := tx.Model(&ontology.OntModelObjectProperty{}).
					Where("id = ? AND inverse_of = ?", oldInverse, p.ID).
					Update("inverse_of", gorm.Expr("NULL")).Error; err != nil {
					return err
				}
			}
			if newInverse > 0 {
				if err := tx.Model(&ontology.OntModelObjectProperty{}).
					Where("id = ?", newInverse).Update("inverse_of", p.ID).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// DeleteModelObjectProperty 删除：同事务清对方 inverseOf 互指 → 软删
func (s *ObjectPropertyService) DeleteModelObjectProperty(id uint) error {
	var old ontology.OntModelObjectProperty
	if err := global.GVA_DB.First(&old, id).Error; err != nil {
		return errors.New("对象属性不存在")
	}
	if err := modelProjectSvc.AssertProjectWritable(old.ProjectId); err != nil {
		return err
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if old.InverseOf != nil && *old.InverseOf > 0 {
			if err := tx.Model(&ontology.OntModelObjectProperty{}).
				Where("id = ? AND inverse_of = ?", *old.InverseOf, id).
				Update("inverse_of", gorm.Expr("NULL")).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&ontology.OntModelObjectProperty{}, id).Error
	})
}

// GetModelObjectProperty 详情
func (s *ObjectPropertyService) GetModelObjectProperty(id uint) (p ontology.OntModelObjectProperty, err error) {
	err = global.GVA_DB.First(&p, id).Error
	return
}

// GetModelObjectPropertyList 分页
func (s *ObjectPropertyService) GetModelObjectPropertyList(info ontReq.SearchObjectProperty) (list []ontology.OntModelObjectProperty, total int64, err error) {
	db := global.GVA_DB.Model(&ontology.OntModelObjectProperty{})
	if info.ProjectId > 0 {
		db = db.Where("project_id = ?", info.ProjectId)
	}
	if info.DomainClassId > 0 {
		db = db.Where("domain_class_id = ?", info.DomainClassId)
	}
	if info.RangeClassId > 0 {
		db = db.Where("range_class_id = ?", info.RangeClassId)
	}
	if info.TemplateCode != "" {
		db = db.Where("template_code = ?", info.TemplateCode)
	}
	if info.Keyword != "" {
		kw := "%" + info.Keyword + "%"
		db = db.Where("local_name LIKE ? OR label LIKE ?", kw, kw)
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Order("sort_order ASC, id ASC").Limit(info.PageSize).Offset((info.Page - 1) * info.PageSize).Find(&list).Error
	return
}

// SuggestModelObjectPropertyInverse 反向关系建议（不落库）：
// suggestedLocalName=belongsTo、domain/range 对调；range 为空时提示先补 range
func (s *ObjectPropertyService) SuggestModelObjectPropertyInverse(id uint) (ontRes.SuggestInverseResult, error) {
	var p ontology.OntModelObjectProperty
	if err := global.GVA_DB.First(&p, id).Error; err != nil {
		return ontRes.SuggestInverseResult{}, errors.New("对象属性不存在")
	}
	if p.RangeClassId == nil || *p.RangeClassId == 0 {
		return ontRes.SuggestInverseResult{}, errors.New("请先补充值域类，再建议反向关系")
	}
	var domainCls ontology.OntModelClass
	if err := global.GVA_DB.First(&domainCls, p.DomainClassId).Error; err != nil {
		return ontRes.SuggestInverseResult{}, errors.New("域类不存在")
	}
	peerName := domainCls.LabelCn
	if peerName == "" {
		peerName = domainCls.Label
	}
	if peerName == "" {
		peerName = domainCls.LocalName
	}
	return ontRes.SuggestInverseResult{
		SuggestedLocalName:     "belongsTo",
		SuggestedLabel:         peerName + "归属",
		SuggestedDomainClassId: *p.RangeClassId,
		SuggestedRangeClassId:  p.DomainClassId,
	}, nil
}

// InstantiateModelObjectProperty 属性级模板挂载（对象属性骨架：range 留空待补）
func (s *ObjectPropertyService) InstantiateModelObjectProperty(req ontReq.InstantiatePropertyOps, operator string) (skipped bool, err error) {
	if err = modelProjectSvc.AssertProjectWritable(req.ProjectId); err != nil {
		return
	}
	var cls ontology.OntModelClass
	if err = global.GVA_DB.First(&cls, req.ClassId).Error; err != nil {
		return false, errors.New("所属类不存在")
	}
	var pt ontology.OntPropertyTemplate
	if err = global.GVA_DB.Where("template_code = ?", req.TemplateCode).First(&pt).Error; err != nil {
		return false, errors.New("属性模板不存在: " + req.TemplateCode)
	}
	var count int64
	if err = global.GVA_DB.Model(&ontology.OntModelObjectProperty{}).
		Where("domain_class_id = ? AND local_name = ?", cls.ID, pt.TemplateCode).Count(&count).Error; err != nil {
		return
	}
	if count > 0 {
		return true, nil
	}
	var project ontology.OntModelProject
	if err = global.GVA_DB.First(&project, cls.ProjectId).Error; err != nil {
		return false, errors.New("项目不存在")
	}
	p := ontology.OntModelObjectProperty{
		ProjectId:      cls.ProjectId,
		DomainClassId:  cls.ID,
		PropertyIri:    BuildIri(project.NamespaceBase, cls.LocalName+"_"+pt.TemplateCode),
		LocalName:      pt.TemplateCode,
		Label:          pt.Label,
		TemplateCode:   pt.TemplateCode,
		MinCardinality: 0,
		MaxCardinality: -1,
		CreatedBy:      operator,
		UpdatedBy:      operator,
	}
	return false, global.GVA_DB.Create(&p).Error
}

func (s *ObjectPropertyService) checkLocalNameUnique(db *gorm.DB, domainClassId uint, localName string, excludeID uint) error { //nolint:stylecheck // 对齐前端 domainClassId
	var count int64
	q := db.Model(&ontology.OntModelObjectProperty{}).Where("domain_class_id = ? AND local_name = ?", domainClassId, localName)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("同类下本地名已存在")
	}
	return nil
}
