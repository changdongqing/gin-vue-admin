package ontology

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
	ontReq "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/request"
	ontRes "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/response"
	"gorm.io/gorm"
)

// ClassInstantiationService 模板实例化（快照语义：复制字段值，不回写治理域模板）
type ClassInstantiationService struct{}

var classTemplateSvc = &ClassTemplateService{}

// InstantiateClass 类级模板实例化（单事务）：
// 归档校验 → 继承视图 → 二次查分类编码 → 更新类溯源/外观 → 批量复制属性（同类已有跳过）
func (s *ClassInstantiationService) InstantiateClass(req ontReq.InstantiateClassOps, operator string) error {
	if err := modelProjectSvc.AssertProjectWritable(req.ProjectId); err != nil {
		return err
	}
	var cls ontology.OntModelClass
	if err := global.GVA_DB.First(&cls, req.ClassId).Error; err != nil {
		return errors.New("本体类不存在")
	}
	if cls.ProjectId != req.ProjectId {
		return errors.New("类不属于该项目")
	}
	// 治理域继承视图（只读调用，R-23 边界：不写治理域任何表）
	view, err := classTemplateSvc.GetClassTemplateInherited(req.TemplateCode)
	if err != nil {
		return err
	}
	if len(view.Properties) == 0 {
		return errors.New("该模板（含继承链）未声明任何属性")
	}
	var tmpl ontology.OntClassTemplate
	if err := global.GVA_DB.Where("template_code = ?", req.TemplateCode).First(&tmpl).Error; err != nil {
		return errors.New("分类模板不存在")
	}
	var project ontology.OntModelProject
	if err := global.GVA_DB.First(&project, cls.ProjectId).Error; err != nil {
		return errors.New("项目不存在")
	}
	// 属性模板批量 IN（缺失任一整体回滚）
	codes := make([]string, 0, len(view.Properties))
	for _, r := range view.Properties {
		codes = append(codes, r.PropertyTemplateCode)
	}
	var pts []ontology.OntPropertyTemplate
	if err := global.GVA_DB.Where("template_code IN ?", codes).Find(&pts).Error; err != nil {
		return err
	}
	ptByCode := make(map[string]ontology.OntPropertyTemplate, len(pts))
	for _, pt := range pts {
		ptByCode[pt.TemplateCode] = pt
	}
	for _, code := range codes {
		if _, ok := ptByCode[code]; !ok {
			return errors.New("属性模板不存在: " + code)
		}
	}

	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		// 类溯源 + 外观继承（已有值保留）
		updates := map[string]interface{}{
			"template_code":       req.TemplateCode,
			"classification_code": tmpl.ClassificationCode,
			"updated_by":          operator,
		}
		if cls.Icon == "" && view.Icon != "" {
			updates["icon"] = view.Icon
		}
		if cls.Color == "" && view.Color != "" {
			updates["color"] = view.Color
		}
		if err := tx.Model(&ontology.OntModelClass{}).Where("id = ?", cls.ID).Updates(updates).Error; err != nil {
			return err
		}
		// 同类已存在 localName 的属性跳过（幂等：重复实例化视为补挂/刷新溯源）
		var exists []string
		if err := tx.Model(&ontology.OntModelDatatypeProperty{}).
			Where("class_id = ? AND local_name IN ?", cls.ID, codes).
			Pluck("local_name", &exists).Error; err != nil {
			return err
		}
		existSet := make(map[string]bool, len(exists))
		for _, e := range exists {
			existSet[e] = true
		}
		var objExists []string
		if err := tx.Model(&ontology.OntModelObjectProperty{}).
			Where("domain_class_id = ? AND local_name IN ?", cls.ID, codes).
			Pluck("local_name", &objExists).Error; err != nil {
			return err
		}
		for _, e := range objExists {
			existSet[e] = true
		}

		dts := make([]ontology.OntModelDatatypeProperty, 0, len(codes))
		objs := make([]ontology.OntModelObjectProperty, 0)
		for _, r := range view.Properties {
			if existSet[r.PropertyTemplateCode] {
				continue
			}
			pt := ptByCode[r.PropertyTemplateCode]
			propIri := BuildIri(project.NamespaceBase, cls.LocalName+"_"+pt.TemplateCode)
			if pt.Kind == "object" {
				objs = append(objs, ontology.OntModelObjectProperty{
					ProjectId:      cls.ProjectId,
					DomainClassId:  cls.ID,
					PropertyIri:    propIri,
					LocalName:      pt.TemplateCode,
					Label:          pt.Label,
					TemplateCode:   pt.TemplateCode,
					MinCardinality: 0,
					MaxCardinality: -1,
					SortOrder:      r.SortOrder,
					CreatedBy:      operator,
					UpdatedBy:      operator,
				})
				continue
			}
			dts = append(dts, ontology.OntModelDatatypeProperty{
				ProjectId:      cls.ProjectId,
				ClassId:        cls.ID,
				PropertyIri:    propIri,
				LocalName:      pt.TemplateCode,
				Label:          pt.Label,
				TemplateCode:   pt.TemplateCode,
				XsdType:        toXsdType(pt.Type),
				UnitRef:        pt.UnitRef,
				EnumValues:     valuesToJsonArray(pt.Values),
				MinCardinality: 0,
				MaxCardinality: -1,
				IsIdentifier:   pt.IsIdentifier,
				SortOrder:      r.SortOrder,
				CreatedBy:      operator,
				UpdatedBy:      operator,
			})
		}
		if len(dts) > 0 {
			if err := tx.Create(&dts).Error; err != nil {
				return err
			}
		}
		if len(objs) > 0 {
			if err := tx.Create(&objs).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// PreviewInstantiate 实例化预览（只读组装：继承视图 + 模板批量查富化，不落库）
func (s *ClassInstantiationService) PreviewInstantiate(req ontReq.PreviewInstantiateOps) (preview ontRes.InstantiatePreview, err error) {
	view, err := classTemplateSvc.GetClassTemplateInherited(req.TemplateCode)
	if err != nil {
		return preview, err
	}
	if len(view.Properties) == 0 {
		return preview, errors.New("该模板（含继承链）未声明任何属性")
	}
	var tmpl ontology.OntClassTemplate
	if err = global.GVA_DB.Where("template_code = ?", req.TemplateCode).First(&tmpl).Error; err != nil {
		return preview, errors.New("分类模板不存在")
	}
	codes := make([]string, 0, len(view.Properties))
	for _, r := range view.Properties {
		codes = append(codes, r.PropertyTemplateCode)
	}
	var pts []ontology.OntPropertyTemplate
	if err = global.GVA_DB.Where("template_code IN ?", codes).Find(&pts).Error; err != nil {
		return
	}
	ptByCode := make(map[string]ontology.OntPropertyTemplate, len(pts))
	for _, pt := range pts {
		ptByCode[pt.TemplateCode] = pt
	}
	preview = ontRes.InstantiatePreview{
		TemplateCode:       req.TemplateCode,
		ClassificationCode: tmpl.ClassificationCode,
		Icon:               view.Icon,
		Color:              view.Color,
		Properties:         make([]ontRes.PropertyPreview, 0, len(view.Properties)),
	}
	for _, r := range view.Properties {
		item := ontRes.PropertyPreview{
			PropertyTemplateCode: r.PropertyTemplateCode,
			RefType:              r.RefType,
			Source:               r.Source,
		}
		if pt, ok := ptByCode[r.PropertyTemplateCode]; ok {
			item.Kind = pt.Kind
			item.Label = pt.Label
			if pt.Kind == "datatype" {
				item.Type = toXsdType(pt.Type)
				item.UnitRef = pt.UnitRef
			}
		}
		preview.Properties = append(preview.Properties, item)
	}
	return
}

// toXsdType 治理域模板短类型 → xsd 全称（空/未知回退 xsd:string）
func toXsdType(t string) string {
	t = strings.TrimSpace(t)
	if t == "" {
		return "xsd:string"
	}
	if strings.HasPrefix(t, "xsd:") {
		return t
	}
	switch t {
	case "string", "integer", "decimal", "boolean", "datetime":
		return "xsd:" + t
	default:
		return "xsd:string"
	}
}

// valuesToJsonArray 逗号分隔枚举 → JSON 数组字符串（治理域同口径）
func valuesToJsonArray(values string) string {
	values = strings.TrimSpace(values)
	if values == "" {
		return ""
	}
	parts := strings.Split(values, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	b, err := json.Marshal(parts)
	if err != nil {
		return ""
	}
	return string(b)
}
