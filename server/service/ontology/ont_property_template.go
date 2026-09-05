package ontology

import (
	"errors"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
	ontReq "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/request"
	"gorm.io/gorm"
)

var errCodeDuplicated = errors.New("属性模板编码已存在")

type PropertyTemplateService struct{}

// CreatePropertyTemplate 创建（强制 source=custom，builtin 仅迁移种子分发）
func (s *PropertyTemplateService) CreatePropertyTemplate(t *ontology.OntPropertyTemplate, operator string) error {
	if err := s.validateTemplateCodeUnique(global.GVA_DB, t.TemplateCode, 0); err != nil {
		return err
	}
	t.Source = "custom"
	t.Status = 0
	t.CreatedBy, t.UpdatedBy = operator, operator
	return global.GVA_DB.Create(t).Error
}

// UpdatePropertyTemplate 更新（builtin 拒绝；source/编码归属不可变）
func (s *PropertyTemplateService) UpdatePropertyTemplate(t *ontology.OntPropertyTemplate, operator string) error {
	var existing ontology.OntPropertyTemplate
	if err := global.GVA_DB.First(&existing, t.ID).Error; err != nil {
		return errors.New("属性模板不存在")
	}
	if existing.Source == "builtin" {
		return errors.New("内置属性模板不可编辑")
	}
	if err := s.validateTemplateCodeUnique(global.GVA_DB, t.TemplateCode, t.ID); err != nil {
		return err
	}
	// 用 map 更新避免结构体 Updates 忽略零值（deprecated=0/isIdentifier=0 等场景）
	return global.GVA_DB.Model(&ontology.OntPropertyTemplate{}).Where("id = ?", t.ID).Updates(map[string]interface{}{
		"template_code":       t.TemplateCode,
		"kind":                t.Kind,
		"label":               t.Label,
		"alias":               t.Alias,
		"description":         t.Description,
		"category":            t.Category,
		"type":                t.Type,
		"is_identifier":       t.IsIdentifier,
		"unit_ref":            t.UnitRef,
		"values":              t.Values,
		"default_cardinality": t.DefaultCardinality,
		"deprecated":          t.Deprecated,
		"updated_by":          operator,
	}).Error
}

// DeletePropertyTemplate 删除（builtin 拒绝；被分类模板骨架引用时拒绝；软删除，软删后编码可复用）
func (s *PropertyTemplateService) DeletePropertyTemplate(id uint) error {
	var existing ontology.OntPropertyTemplate
	if err := global.GVA_DB.First(&existing, id).Error; err != nil {
		return errors.New("属性模板不存在")
	}
	if existing.Source == "builtin" {
		return errors.New("内置属性模板不可删除")
	}
	// 引用计数校验（02 分类模板骨架落地）：被未删除的骨架引用则拒绝
	var refCount int64
	global.GVA_DB.Model(&ontology.OntClassTemplateRef{}).
		Where("property_template_code = ?", existing.TemplateCode).
		Count(&refCount)
	if refCount > 0 {
		return errors.New("属性模板已被分类模板骨架引用，不可删除")
	}
	return global.GVA_DB.Delete(&ontology.OntPropertyTemplate{}, id).Error
}

// DisablePropertyTemplate 弃用/取消弃用（幂等双向切换；builtin 也可弃用）
func (s *PropertyTemplateService) DisablePropertyTemplate(id uint, operator string) error {
	var existing ontology.OntPropertyTemplate
	if err := global.GVA_DB.First(&existing, id).Error; err != nil {
		return errors.New("属性模板不存在")
	}
	target := 1 - existing.Deprecated // 0→1 弃用，1→0 取消弃用
	return global.GVA_DB.Model(&ontology.OntPropertyTemplate{}).Where("id = ?", id).
		Updates(map[string]interface{}{"deprecated": target, "updated_by": operator}).Error
}

// GetPropertyTemplate 分页查询的详情查询
func (s *PropertyTemplateService) GetPropertyTemplate(id uint) (t ontology.OntPropertyTemplate, err error) {
	err = global.GVA_DB.First(&t, id).Error
	return
}

// GetPropertyTemplateList 分页查询（keyword 匹配编码/显示名/别名三字段）
func (s *PropertyTemplateService) GetPropertyTemplateList(info ontReq.SearchPropertyTemplate) (list []ontology.OntPropertyTemplate, total int64, err error) {
	db := s.buildListQuery(info)
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Order("id DESC").Limit(info.PageSize).Offset((info.Page - 1) * info.PageSize).Find(&list).Error
	return
}

func (s *PropertyTemplateService) buildListQuery(info ontReq.SearchPropertyTemplate) *gorm.DB {
	db := global.GVA_DB.Model(&ontology.OntPropertyTemplate{})
	if info.Kind != "" {
		db = db.Where("kind = ?", info.Kind)
	}
	if info.Category != "" {
		db = db.Where("category = ?", info.Category)
	}
	if info.Source != "" {
		db = db.Where("source = ?", info.Source)
	}
	if info.Deprecated != nil {
		db = db.Where("deprecated = ?", *info.Deprecated)
	}
	if info.Keyword != "" {
		kw := "%" + info.Keyword + "%"
		db = db.Where("template_code LIKE ? OR label LIKE ? OR alias LIKE ?", kw, kw, kw)
	}
	return db
}

// GetPropertyTemplateAll 全量列表（下拉/选择器用，可按 kind 过滤，默认排除弃用）
func (s *PropertyTemplateService) GetPropertyTemplateAll(kind string) (list []ontology.OntPropertyTemplate, err error) {
	db := global.GVA_DB.Where("deprecated = 0 AND status = 0")
	if kind != "" {
		db = db.Where("kind = ?", kind)
	}
	err = db.Order("kind ASC, category ASC, id ASC").Find(&list).Error
	return
}

// GetPropertyTemplatesForSupply 供给查询（FR-5）：默认排除弃用、仅启用
func (s *PropertyTemplateService) GetPropertyTemplatesForSupply(kind, category string, includeDeprecated bool) (list []ontology.OntPropertyTemplate, err error) {
	db := global.GVA_DB.Where("status = 0")
	if !includeDeprecated {
		db = db.Where("deprecated = 0")
	}
	if kind != "" {
		db = db.Where("kind = ?", kind)
	}
	if category != "" {
		db = db.Where("category = ?", category)
	}
	err = db.Order("kind ASC, category ASC, id ASC").Find(&list).Error
	return
}

// IsTemplateCodeUnique 编码远程查重（true=可用）
func (s *PropertyTemplateService) IsTemplateCodeUnique(code string, excludeID uint) (bool, error) {
	err := s.validateTemplateCodeUnique(global.GVA_DB, code, excludeID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, errCodeDuplicated) {
		return false, nil
	}
	return false, err
}

// PromotePropertyTemplate 提升（FR-6）：建模侧属性快照生成 custom 模板，重复编码返回业务错误不覆盖
func (s *PropertyTemplateService) PromotePropertyTemplate(req ontReq.PromotePropertyTemplate, operator string) error {
	code := req.TemplateCode
	if code == "" {
		code = deriveCamelCase(req.Name)
		if code == "" {
			return errors.New("无法从 name 派生模板编码，请显式传入 templateCode")
		}
	}
	kind := req.Kind
	if kind == "" {
		kind = "datatype"
	}
	label := req.Label
	if label == "" {
		label = req.Name
	}
	t := ontology.OntPropertyTemplate{
		TemplateCode:       code,
		Kind:               kind,
		Label:              label,
		Alias:              req.Alias,
		Description:        req.Description,
		Category:           req.Category,
		Type:               req.Type,
		IsIdentifier:       req.IsIdentifier,
		UnitRef:            req.UnitRef,
		Values:             req.Values,
		DefaultCardinality: req.DefaultCardinality,
	}
	if err := s.validateTemplateCodeUnique(global.GVA_DB, t.TemplateCode, 0); err != nil {
		return err
	}
	t.Source = "custom"
	t.Status = 0
	t.CreatedBy, t.UpdatedBy = operator, operator
	return global.GVA_DB.Create(&t).Error
}

// deriveCamelCase 从名称派生 camelCase 编码（按非字母数字切词，首词小写；全中文等无法派生时返回空串）
func deriveCamelCase(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9')
	})
	var b strings.Builder
	for i, w := range words {
		if w == "" {
			continue
		}
		if i == 0 {
			b.WriteString(strings.ToLower(w))
		} else {
			b.WriteString(strings.ToUpper(w[:1]) + strings.ToLower(w[1:]))
		}
	}
	return b.String()
}

func (s *PropertyTemplateService) validateTemplateCodeUnique(db *gorm.DB, code string, excludeID uint) error {
	var count int64
	q := db.Model(&ontology.OntPropertyTemplate{}).Where("template_code = ?", code)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errCodeDuplicated
	}
	return nil
}
