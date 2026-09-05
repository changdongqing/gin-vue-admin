package ontology

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
	"gorm.io/gorm"
)

type ClassificationRuleService struct{}

// defaultRule 无规则时的兜底默认值（改规则只影响新节点，存量编码冻结）
func (s *ClassificationRuleService) defaultRule(treeRoot string) ontology.OntClassificationRule {
	return ontology.OntClassificationRule{
		TreeRoot:    treeRoot,
		Separator:   "-",
		LevelDigits: 2,
		BaseNumber:  1,
		ZeroPad:     1,
	}
}

// FindClassificationRule 查询某棵树的编码规则（无则返回默认值，不落库）
func (s *ClassificationRuleService) FindClassificationRule(treeRoot string) (ontology.OntClassificationRule, error) {
	var rule ontology.OntClassificationRule
	err := global.GVA_DB.Where("tree_root = ?", treeRoot).First(&rule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.defaultRule(treeRoot), nil
	}
	return rule, err
}

// SaveClassificationRule 保存编码规则（按 treeRoot upsert；改规则只影响新节点，存量编码冻结）
func (s *ClassificationRuleService) SaveClassificationRule(rule *ontology.OntClassificationRule) error {
	var existing ontology.OntClassificationRule
	err := global.GVA_DB.Where("tree_root = ?", rule.TreeRoot).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return global.GVA_DB.Create(rule).Error
	}
	if err != nil {
		return err
	}
	return global.GVA_DB.Model(&ontology.OntClassificationRule{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{
		"separator":    rule.Separator,
		"level_digits": rule.LevelDigits,
		"base_number":  rule.BaseNumber,
		"zero_pad":     rule.ZeroPad,
		"description":  rule.Description,
	}).Error
}

// GenerateClassificationCode 编码生成：根节点=baseNumber；子节点=父编码+分隔符+零填充同级序号
// 序号 = 同级现有子节点数 + 1（并发重复由 (treeRoot, classification_code) 唯一索引兜底）
func (s *ClassificationRuleService) GenerateClassificationCode(db *gorm.DB, parentId uint, treeRoot string) (string, error) {
	rule, err := s.FindClassificationRule(treeRoot)
	if err != nil {
		return "", err
	}
	if parentId == 0 {
		return strconv.Itoa(rule.BaseNumber), nil
	}
	var parent ontology.OntClassTemplate
	if err := db.First(&parent, parentId).Error; err != nil {
		return "", errors.New("父分类模板不存在")
	}
	var count int64
	db.Model(&ontology.OntClassTemplate{}).Where("parent_id = ?", parentId).Count(&count)
	seq := int(count) + 1
	suffix := strconv.Itoa(seq)
	if rule.ZeroPad == 1 {
		suffix = fmt.Sprintf("%0"+strconv.Itoa(rule.LevelDigits)+"d", seq)
	}
	return parent.ClassificationCode + rule.Separator + suffix, nil
}
