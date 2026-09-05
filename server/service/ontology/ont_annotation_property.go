package ontology

import (
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
	ontReq "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/request"
	"gorm.io/gorm"
)

type AnnotationPropertyService struct{}

// CreateAnnotationProperty 创建（localName 唯一校验）
func (s *AnnotationPropertyService) CreateAnnotationProperty(p *ontology.OntAnnotationProperty, operator string) error {
	if err := s.validateLocalNameUnique(global.GVA_DB, p.LocalName, 0); err != nil {
		return err
	}
	p.CreatedBy, p.UpdatedBy = operator, operator
	return global.GVA_DB.Create(p).Error
}

// UpdateAnnotationProperty 更新（localName 唯一校验排除自身；无 builtin 保护）
func (s *AnnotationPropertyService) UpdateAnnotationProperty(p *ontology.OntAnnotationProperty, operator string) error {
	var existing ontology.OntAnnotationProperty
	if err := global.GVA_DB.First(&existing, p.ID).Error; err != nil {
		return errors.New("注释属性不存在")
	}
	if err := s.validateLocalNameUnique(global.GVA_DB, p.LocalName, p.ID); err != nil {
		return err
	}
	return global.GVA_DB.Model(&ontology.OntAnnotationProperty{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
		"local_name":  p.LocalName,
		"label":       p.Label,
		"range_xsd":   p.RangeXsd,
		"applies_to":  p.AppliesTo,
		"description": p.Description,
		"sort":        p.Sort,
		"updated_by":  operator,
	}).Error
}

// DeleteAnnotationProperty 删除（无 builtin 保护；删除后建模侧不再序列化对应 ont:xxx）
func (s *AnnotationPropertyService) DeleteAnnotationProperty(id uint) error {
	var existing ontology.OntAnnotationProperty
	if err := global.GVA_DB.First(&existing, id).Error; err != nil {
		return errors.New("注释属性不存在")
	}
	return global.GVA_DB.Delete(&ontology.OntAnnotationProperty{}, id).Error
}

// GetAnnotationProperty 详情
func (s *AnnotationPropertyService) GetAnnotationProperty(id uint) (p ontology.OntAnnotationProperty, err error) {
	err = global.GVA_DB.First(&p, id).Error
	return
}

// GetAnnotationPropertyList 分页（keyword 匹配 localName/label；可选 appliesTo/时间范围）
func (s *AnnotationPropertyService) GetAnnotationPropertyList(info ontReq.SearchAnnotationProperty) (list []ontology.OntAnnotationProperty, total int64, err error) {
	db := s.buildListQuery(info)
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Order("sort ASC, id ASC").Limit(info.PageSize).Offset((info.Page - 1) * info.PageSize).Find(&list).Error
	return
}

// GetAnnotationPropertyAll 导出/下拉数据源：同分页条件、去分页（避免两套条件逻辑漂移）
func (s *AnnotationPropertyService) GetAnnotationPropertyAll(info ontReq.SearchAnnotationProperty) (list []ontology.OntAnnotationProperty, err error) {
	err = s.buildListQuery(info).Order("sort ASC, id ASC").Find(&list).Error
	return
}

// GetAnnotationPropertiesForSupply 供给：全量注册表（Order sort,id；数量预期 ≤ 数十条不分页）
func (s *AnnotationPropertyService) GetAnnotationPropertiesForSupply() (list []ontology.OntAnnotationProperty, err error) {
	err = global.GVA_DB.Order("sort ASC, id ASC").Find(&list).Error
	return
}

func (s *AnnotationPropertyService) buildListQuery(info ontReq.SearchAnnotationProperty) *gorm.DB {
	db := global.GVA_DB.Model(&ontology.OntAnnotationProperty{})
	if info.AppliesTo != "" {
		db = db.Where("applies_to = ?", info.AppliesTo)
	}
	if info.Keyword != "" {
		kw := "%" + info.Keyword + "%"
		db = db.Where("local_name LIKE ? OR label LIKE ?", kw, kw)
	}
	if info.StartTime != "" {
		db = db.Where("created_at >= ?", info.StartTime)
	}
	if info.EndTime != "" {
		db = db.Where("created_at <= ?", info.EndTime)
	}
	return db
}

func (s *AnnotationPropertyService) validateLocalNameUnique(db *gorm.DB, localName string, excludeID uint) error {
	var count int64
	q := db.Model(&ontology.OntAnnotationProperty{}).Where("local_name = ?", localName)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("注释属性名已存在")
	}
	return nil
}
