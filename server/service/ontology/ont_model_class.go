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

type ModelClassService struct{}

// modelProjectSvc 无状态服务引用（跨域复用 01 的归档校验）
var modelProjectSvc = &ModelProjectService{}

// classLocalNameRe 类本地名正则（创建后不可改）
var classLocalNameRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

// CreateModelClass 空白创建：归档校验 → localName 正则 → IRI 生成 → 同项目查重
// 溯源字段强制空串（只允许 instantiate 写入）
func (s *ModelClassService) CreateModelClass(p *ontology.OntModelClass, operator string) error {
	if err := modelProjectSvc.AssertProjectWritable(p.ProjectId); err != nil {
		return err
	}
	if !classLocalNameRe.MatchString(p.LocalName) {
		return errors.New("本地名不合法（须以字母开头，仅含字母/数字/下划线/连字符）")
	}
	var project ontology.OntModelProject
	if err := global.GVA_DB.First(&project, p.ProjectId).Error; err != nil {
		return errors.New("项目不存在")
	}
	p.ClassIri = BuildIri(project.NamespaceBase, p.LocalName)
	if err := s.checkIriUnique(global.GVA_DB, p.ProjectId, p.ClassIri, 0); err != nil {
		return err
	}
	p.TemplateCode, p.ClassificationCode = "", ""
	p.Table = "" // 预留列不写（业务表映射已迁出至外部模块关联）
	p.CreatedBy, p.UpdatedBy = operator, operator
	return global.GVA_DB.Create(p).Error
}

// UpdateModelClass 编辑白名单：label/labelCn/description/icon/color/sortOrder/isInstantiable
// classIri/localName/projectId/溯源字段不可改
func (s *ModelClassService) UpdateModelClass(p *ontology.OntModelClass, operator string) error {
	var old ontology.OntModelClass
	if err := global.GVA_DB.First(&old, p.ID).Error; err != nil {
		return errors.New("本体类不存在")
	}
	if err := modelProjectSvc.AssertProjectWritable(old.ProjectId); err != nil {
		return err
	}
	return global.GVA_DB.Model(&ontology.OntModelClass{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
		"label":           p.Label,
		"label_cn":        p.LabelCn,
		"description":     p.Description,
		"icon":            p.Icon,
		"color":           p.Color,
		"sort_order":      p.SortOrder,
		"is_instantiable": p.IsInstantiable,
		"updated_by":      operator,
	}).Error
}

// DeleteModelClass 三重守卫式删除：归档 → 被子类引用 → 类下存在属性
func (s *ModelClassService) DeleteModelClass(id uint) error {
	var old ontology.OntModelClass
	if err := global.GVA_DB.First(&old, id).Error; err != nil {
		return errors.New("本体类不存在")
	}
	if err := modelProjectSvc.AssertProjectWritable(old.ProjectId); err != nil {
		return err
	}
	var subCount int64
	if err := global.GVA_DB.Model(&ontology.OntModelSubclassOf{}).
		Where("parent_class_id = ?", id).Count(&subCount).Error; err != nil {
		return err
	}
	if subCount > 0 {
		return errors.New("该类被子类引用，无法删除")
	}
	var dtCount int64
	if err := global.GVA_DB.Model(&ontology.OntModelDatatypeProperty{}).
		Where("class_id = ?", id).Count(&dtCount).Error; err != nil {
		return err
	}
	var objCount int64
	if err := global.GVA_DB.Model(&ontology.OntModelObjectProperty{}).
		Where("domain_class_id = ? OR range_class_id = ?", id, id).Count(&objCount).Error; err != nil {
		return err
	}
	if dtCount+objCount > 0 {
		return errors.New("该类下存在属性，无法删除")
	}
	return global.GVA_DB.Delete(&ontology.OntModelClass{}, id).Error
}

// GetModelClass 单实体详情（编辑回填）
func (s *ModelClassService) GetModelClass(id uint) (p ontology.OntModelClass, err error) {
	err = global.GVA_DB.First(&p, id).Error
	return
}

// GetModelClassList 分页（projectId 等值 + keyword + hasTemplate 三态）
func (s *ModelClassService) GetModelClassList(info ontReq.SearchModelClass) (list []ontology.OntModelClass, total int64, err error) {
	db := s.buildListQuery(info)
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Order("sort_order ASC, id ASC").Limit(info.PageSize).Offset((info.Page - 1) * info.PageSize).Find(&list).Error
	return
}

// GetModelClassByProject 项目内全量类（03 外部模块关联选类用；isInstantiable 过滤由前端做）
func (s *ModelClassService) GetModelClassByProject(projectId uint) (list []ontology.OntModelClass, err error) { //nolint:stylecheck // 对齐前端 projectId
	err = global.GVA_DB.Where("project_id = ?", projectId).Order("sort_order ASC, id ASC").Find(&list).Error
	return
}

// GetModelClassDetail 详情聚合：类 + 属性双清单（Order sort_order,id）+ 父子类双向 join
func (s *ModelClassService) GetModelClassDetail(id uint) (detail ontRes.ClassDetail, err error) {
	var cls ontology.OntModelClass
	if err = global.GVA_DB.First(&cls, id).Error; err != nil {
		return detail, errors.New("本体类不存在")
	}
	detail.OntModelClass = cls

	var dts []ontology.OntModelDatatypeProperty
	if err = global.GVA_DB.Where("class_id = ?", id).Order("sort_order ASC, id ASC").Find(&dts).Error; err != nil {
		return
	}
	detail.DatatypeProperties = make([]ontRes.PropertySummary, 0, len(dts))
	for _, d := range dts {
		detail.DatatypeProperties = append(detail.DatatypeProperties, ontRes.PropertySummary{
			ID: d.ID, PropertyIri: d.PropertyIri, LocalName: d.LocalName,
			Label: d.Label, TemplateCode: d.TemplateCode, TypeOrRange: d.XsdType,
		})
	}

	var objs []ontology.OntModelObjectProperty
	if err = global.GVA_DB.Where("domain_class_id = ?", id).Order("sort_order ASC, id ASC").Find(&objs).Error; err != nil {
		return
	}
	rangeIds := make([]uint, 0, len(objs))
	for _, o := range objs {
		if o.RangeClassId != nil {
			rangeIds = append(rangeIds, *o.RangeClassId)
		}
	}
	rangeNames := map[uint]string{}
	if len(rangeIds) > 0 {
		var ranges []ontology.OntModelClass
		if err = global.GVA_DB.Where("id IN ?", rangeIds).Find(&ranges).Error; err != nil {
			return
		}
		for _, r := range ranges {
			rangeNames[r.ID] = r.LabelCn
			if rangeNames[r.ID] == "" {
				rangeNames[r.ID] = r.Label
			}
			if rangeNames[r.ID] == "" {
				rangeNames[r.ID] = r.LocalName
			}
		}
	}
	detail.ObjectProperties = make([]ontRes.PropertySummary, 0, len(objs))
	for _, o := range objs {
		typeOrRange := ""
		if o.RangeClassId != nil {
			typeOrRange = rangeNames[*o.RangeClassId]
		}
		detail.ObjectProperties = append(detail.ObjectProperties, ontRes.PropertySummary{
			ID: o.ID, PropertyIri: o.PropertyIri, LocalName: o.LocalName,
			Label: o.Label, TemplateCode: o.TemplateCode, TypeOrRange: typeOrRange,
		})
	}

	// 父类：child_class_id = id 的 parent 集合；子类：parent_class_id = id 的 child 集合
	if detail.ParentClasses, err = s.joinClassSummaries("child_class_id", id); err != nil {
		return
	}
	detail.ChildClasses, err = s.joinClassSummaries("parent_class_id", id)
	return
}

// joinClassSummaries subclassofs 单向 join 类表（column=本类所在列，取对侧类摘要）
func (s *ModelClassService) joinClassSummaries(column string, classId uint) (list []ontRes.ClassSummary, err error) { //nolint:stylecheck // 局部小驼峰
	other := "parent_class_id"
	if column == "parent_class_id" {
		other = "child_class_id"
	}
	err = global.GVA_DB.Model(&ontology.OntModelSubclassOf{}).
		Select("ont_model_classes.id, ont_model_classes.class_iri, ont_model_classes.local_name, ont_model_classes.label, ont_model_classes.label_cn").
		Joins("JOIN ont_model_classes ON ont_model_classes.id = ont_model_subclassofs."+other).
		Where("ont_model_subclassofs."+column+" = ? AND ont_model_subclassofs.deleted_at IS NULL", classId).
		Scan(&list).Error
	return
}

// CheckModelClassLocalName 本地名查重（同项目）：true=已存在
func (s *ModelClassService) CheckModelClassLocalName(projectId uint, localName string, excludeId uint) (bool, error) { //nolint:stylecheck // 对齐前端
	var count int64
	q := global.GVA_DB.Model(&ontology.OntModelClass{}).Where("project_id = ? AND local_name = ?", projectId, localName)
	if excludeId > 0 {
		q = q.Where("id <> ?", excludeId)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *ModelClassService) buildListQuery(info ontReq.SearchModelClass) *gorm.DB {
	db := global.GVA_DB.Model(&ontology.OntModelClass{})
	if info.ProjectId > 0 {
		db = db.Where("project_id = ?", info.ProjectId)
	}
	if info.Keyword != "" {
		kw := "%" + info.Keyword + "%"
		db = db.Where("label LIKE ? OR label_cn LIKE ? OR local_name LIKE ?", kw, kw, kw)
	}
	switch info.HasTemplate {
	case "true":
		db = db.Where("template_code <> ''")
	case "false":
		db = db.Where("template_code = ''")
	}
	return db
}

func (s *ModelClassService) checkIriUnique(db *gorm.DB, projectId uint, classIri string, excludeID uint) error {
	var count int64
	q := db.Model(&ontology.OntModelClass{}).Where("project_id = ? AND class_iri = ?", projectId, classIri)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("同类本地名已存在")
	}
	return nil
}
