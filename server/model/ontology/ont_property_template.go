package ontology

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// OntPropertyTemplate 本体属性模板（数据属性 + 对象属性同构，kind 区分）
// 唯一业务编码由迁移 SQL 的部分唯一索引（WHERE deleted_at IS NULL）+ Service 查重共同保证，
// 此处不声明 uniqueIndex tag（AutoMigrate 对 PG 部分索引支持有限，结构以迁移为准）
type OntPropertyTemplate struct {
	global.GVA_MODEL
	TemplateCode       string `json:"templateCode" form:"templateCode" gorm:"column:template_code;size:64;not null;comment:模板编码（唯一业务标识）" binding:"required"` //nolint:stylecheck // 驼峰 JSON 契约
	Kind               string `json:"kind" form:"kind" gorm:"column:kind;size:16;not null;default:datatype;comment:属性类型 datatype/object" binding:"required,oneof=datatype object"`
	Label              string `json:"label" form:"label" gorm:"column:label;size:128;not null;comment:显示名" binding:"required"`
	Alias              string `json:"alias" form:"alias" gorm:"column:alias;size:512;comment:同义别名（逗号分隔）"`
	Description        string `json:"description" form:"description" gorm:"column:description;size:512;comment:业务说明"`
	Category           string `json:"category" form:"category" gorm:"column:category;size:64;comment:分组"`
	Type               string `json:"type" form:"type" gorm:"column:type;size:32;comment:数据类型(datatype 专属)"`
	IsIdentifier       int    `json:"isIdentifier" form:"isIdentifier" gorm:"column:is_identifier;default:0;comment:是否标识 0/1"`
	UnitRef            string `json:"unitRef" form:"unitRef" gorm:"column:unit_ref;size:255;comment:预设 QUDT 单位 IRI"`
	Values             string `json:"values" form:"values" gorm:"column:values;size:1000;comment:枚举值(逗号分隔)"` //nolint:stylecheck // 对齐前端字段 values
	DefaultCardinality string `json:"defaultCardinality" form:"defaultCardinality" gorm:"column:default_cardinality;size:32;comment:默认基数(object 专属)"`
	Source             string `json:"source" form:"source" gorm:"column:source;size:16;not null;default:custom;comment:来源 builtin/custom"`
	Deprecated         int    `json:"deprecated" form:"deprecated" gorm:"column:deprecated;default:0;comment:是否弃用 0/1"`
	Status             int    `json:"status" form:"status" gorm:"column:status;default:0;comment:业务启停 0/1"`
	CreatedBy          string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy          string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (OntPropertyTemplate) TableName() string {
	return "ont_property_templates"
}
