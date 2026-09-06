package ontology

import (
	"errors"
	"regexp"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
	ontReq "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/request"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ModelProjectService struct{}

// ncNameRe 前缀名 NCName 语义正则（如 ex/qudt/fire）
var ncNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.-]*$`)

// BuildIri 命名空间基址拼接 IRI：结尾 `/` 或 `#` 直拼，否则补 `/`
// 建模域统一入口（01 项目/02 类与属性 IRI 生成共用，双侧不得各自实现）
func BuildIri(namespaceBase, localName string) string {
	if namespaceBase == "" {
		return localName
	}
	last := namespaceBase[len(namespaceBase)-1]
	if last != '/' && last != '#' {
		return namespaceBase + "/" + localName
	}
	return namespaceBase + localName
}

// CreateModelProject 创建项目（编码查重 → 默认值兜底 → 方案A拒绝）
func (s *ModelProjectService) CreateModelProject(p *ontology.OntModelProject, operator string) error {
	if p.SerializationStrategy == "A" {
		return errors.New("方案A暂未开放")
	}
	if err := s.checkCodeUnique(global.GVA_DB, p.ProjectCode, 0); err != nil {
		return err
	}
	applyProjectDefaults(p)
	p.CreatedBy, p.UpdatedBy = operator, operator
	return global.GVA_DB.Create(p).Error
}

// UpdateModelProject 更新项目（archived 拒绝 → 状态机单向 → 编码查重排除自身 → 方案A拒绝）
func (s *ModelProjectService) UpdateModelProject(p *ontology.OntModelProject, operator string) error {
	var old ontology.OntModelProject
	if err := global.GVA_DB.First(&old, p.ID).Error; err != nil {
		return errors.New("项目不存在")
	}
	if old.Status == "archived" {
		return errors.New("项目已归档，只读")
	}
	status := p.Status
	if status == "" {
		status = old.Status
	}
	if !validProjectStatusTransition(old.Status, status) {
		return errors.New("项目状态仅允许 draft→active→archived 单向流转")
	}
	if p.SerializationStrategy == "A" {
		return errors.New("方案A暂未开放")
	}
	if err := s.checkCodeUnique(global.GVA_DB, p.ProjectCode, p.ID); err != nil {
		return err
	}
	return global.GVA_DB.Model(&ontology.OntModelProject{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
		"project_code":           p.ProjectCode,
		"name":                   p.Name,
		"description":            p.Description,
		"namespace_base":         p.NamespaceBase,
		"default_format":         withDefault(p.DefaultFormat, old.DefaultFormat, "TTL"),
		"serialization_strategy": withDefault(p.SerializationStrategy, old.SerializationStrategy, "B"),
		"status":                 status,
		"updated_by":             operator,
	}).Error
}

// DeleteModelProject 删除项目（archived 拒绝；项目下存在本体类拒绝——02 落地后生效）
// 前缀随项目语义废弃，不做级联删除（保留恢复可能）
func (s *ModelProjectService) DeleteModelProject(id uint) error {
	var old ontology.OntModelProject
	if err := global.GVA_DB.First(&old, id).Error; err != nil {
		return errors.New("项目不存在")
	}
	if old.Status == "archived" {
		return errors.New("项目已归档，只读")
	}
	var classCount int64
	if global.GVA_DB.Migrator().HasTable("ont_model_classes") {
		if err := global.GVA_DB.Table("ont_model_classes").
			Where("project_id = ? AND deleted_at IS NULL", id).Count(&classCount).Error; err != nil {
			return err
		}
	}
	if classCount > 0 {
		return errors.New("项目下存在本体类，无法删除")
	}
	return global.GVA_DB.Delete(&ontology.OntModelProject{}, id).Error
}

// GetModelProject 详情
func (s *ModelProjectService) GetModelProject(id uint) (p ontology.OntModelProject, err error) {
	err = global.GVA_DB.First(&p, id).Error
	return
}

// GetModelProjectList 分页（keyword 匹配 project_code/name + status 等值）
func (s *ModelProjectService) GetModelProjectList(info ontReq.SearchModelProject) (list []ontology.OntModelProject, total int64, err error) {
	db := s.buildListQuery(info)
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Order("id DESC").Limit(info.PageSize).Offset((info.Page - 1) * info.PageSize).Find(&list).Error
	return
}

// GetModelProjectAll 全量下拉数据源（02 类建模搜索区 / 03 绑定表单选项目）
func (s *ModelProjectService) GetModelProjectAll() (list []ontology.OntModelProject, err error) {
	err = global.GVA_DB.Order("id ASC").Find(&list).Error
	return
}

// CheckModelProjectCode 编码查重：true=已存在
func (s *ModelProjectService) CheckModelProjectCode(projectCode string, excludeId uint) (bool, error) { //nolint:stylecheck // 对齐前端 excludeId
	var count int64
	q := global.GVA_DB.Model(&ontology.OntModelProject{}).Where("project_code = ?", projectCode)
	if excludeId > 0 {
		q = q.Where("id <> ?", excludeId)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// AssertProjectWritable 建模域公共校验（02 类/属性/层级、03 绑定/同步的写入口复用）：
// 项目存在且未归档
func (s *ModelProjectService) AssertProjectWritable(projectId uint) error {
	var p ontology.OntModelProject
	if err := global.GVA_DB.First(&p, projectId).Error; err != nil {
		return errors.New("项目不存在")
	}
	if p.Status == "archived" {
		return errors.New("项目已归档，只读")
	}
	return nil
}

func (s *ModelProjectService) buildListQuery(info ontReq.SearchModelProject) *gorm.DB {
	db := global.GVA_DB.Model(&ontology.OntModelProject{})
	if info.Status != "" {
		db = db.Where("status = ?", info.Status)
	}
	if info.Keyword != "" {
		kw := "%" + info.Keyword + "%"
		db = db.Where("project_code LIKE ? OR name LIKE ?", kw, kw)
	}
	return db
}

func (s *ModelProjectService) checkCodeUnique(db *gorm.DB, projectCode string, excludeID uint) error {
	var count int64
	q := db.Model(&ontology.OntModelProject{}).Where("project_code = ?", projectCode)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("项目编码已存在")
	}
	return nil
}

// applyProjectDefaults 创建默认值兜底（前端表单同默认值）
func applyProjectDefaults(p *ontology.OntModelProject) {
	if p.DefaultFormat == "" {
		p.DefaultFormat = "TTL"
	}
	if p.SerializationStrategy == "" {
		p.SerializationStrategy = "B"
	}
	if p.Status == "" {
		p.Status = "draft"
	}
}

// validProjectStatusTransition 状态机单向：draft→active→archived（相同状态视为无流转）
func validProjectStatusTransition(oldStatus, newStatus string) bool {
	if oldStatus == newStatus {
		return true
	}
	return (oldStatus == "draft" && newStatus == "active") ||
		(oldStatus == "active" && newStatus == "archived")
}

// withDefault 提交值为空时沿用库中旧值，仍为空则兜底默认
func withDefault(submitted, old, dft string) string {
	if submitted != "" {
		return submitted
	}
	if old != "" {
		return old
	}
	return dft
}

type ModelPrefixService struct{}

// GetModelPrefixList 项目前缀列表（默认前缀置顶）
func (s *ModelPrefixService) GetModelPrefixList(projectId uint) (list []ontology.OntModelPrefix, err error) { //nolint:stylecheck // 对齐前端 projectId
	err = global.GVA_DB.Where("project_id = ?", projectId).
		Order("is_default DESC, id ASC").Find(&list).Error
	return
}

// CreateModelPrefix 新增前缀：项目未归档 → NCName → 同项目唯一 → 默认前缀唯一
// 设默认走项目行 FOR UPDATE 串行，防并发双默认
func (s *ModelPrefixService) CreateModelPrefix(p *ontology.OntModelPrefix, operator string) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := lockProjectWritable(tx, p.ProjectId); err != nil {
			return err
		}
		if !ncNameRe.MatchString(p.Prefix) {
			return errors.New("前缀名不合法（须以字母或下划线开头，仅含字母/数字/下划线/点/连字符）")
		}
		if err := checkPrefixUnique(tx, p.ProjectId, p.Prefix, 0); err != nil {
			return err
		}
		if p.IsDefault != 1 {
			p.IsDefault = 0
		} else if err := checkNoOtherDefault(tx, p.ProjectId, 0); err != nil {
			return err
		}
		p.CreatedBy, p.UpdatedBy = operator, operator
		return tx.Create(p).Error
	})
}

// UpdateModelPrefix 更新前缀：存在+归属 → 项目未归档 → 同上校验（排除自身）
func (s *ModelPrefixService) UpdateModelPrefix(p *ontology.OntModelPrefix, operator string) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var old ontology.OntModelPrefix
		if err := tx.First(&old, p.ID).Error; err != nil {
			return errors.New("前缀不存在")
		}
		if old.ProjectId != p.ProjectId {
			return errors.New("前缀不属于该项目")
		}
		if err := lockProjectWritable(tx, p.ProjectId); err != nil {
			return err
		}
		if !ncNameRe.MatchString(p.Prefix) {
			return errors.New("前缀名不合法（须以字母或下划线开头，仅含字母/数字/下划线/点/连字符）")
		}
		if err := checkPrefixUnique(tx, p.ProjectId, p.Prefix, p.ID); err != nil {
			return err
		}
		if p.IsDefault != 1 {
			p.IsDefault = 0
		} else if err := checkNoOtherDefault(tx, p.ProjectId, p.ID); err != nil {
			return err
		}
		return tx.Model(&ontology.OntModelPrefix{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
			"prefix":     p.Prefix,
			"namespace":  p.Namespace,
			"is_default": p.IsDefault,
			"updated_by": operator,
		}).Error
	})
}

// DeleteModelPrefix 删除前缀：存在+归属 → 项目未归档 → 软删（取消默认无需校验）
func (s *ModelPrefixService) DeleteModelPrefix(id, projectId uint) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var old ontology.OntModelPrefix
		if err := tx.First(&old, id).Error; err != nil {
			return errors.New("前缀不存在")
		}
		if old.ProjectId != projectId {
			return errors.New("前缀不属于该项目")
		}
		if err := lockProjectWritable(tx, projectId); err != nil {
			return err
		}
		return tx.Delete(&ontology.OntModelPrefix{}, id).Error
	})
}

// lockProjectWritable 项目行 FOR UPDATE + 存在/归档校验（设默认读-改-写的项目级串行点）
func lockProjectWritable(tx *gorm.DB, projectId uint) error {
	var p ontology.OntModelProject
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&p, projectId).Error; err != nil {
		return errors.New("项目不存在")
	}
	if p.Status == "archived" {
		return errors.New("项目已归档，只读")
	}
	return nil
}

func checkPrefixUnique(db *gorm.DB, projectId uint, prefix string, excludeID uint) error {
	var count int64
	q := db.Model(&ontology.OntModelPrefix{}).Where("project_id = ? AND prefix = ?", projectId, prefix)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("前缀已存在")
	}
	return nil
}

func checkNoOtherDefault(db *gorm.DB, projectId uint, excludeID uint) error {
	var count int64
	q := db.Model(&ontology.OntModelPrefix{}).Where("project_id = ? AND is_default = 1", projectId)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("已存在默认前缀")
	}
	return nil
}
