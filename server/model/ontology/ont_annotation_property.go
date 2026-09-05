package ontology

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// OntAnnotationProperty 本体注释属性注册表（ont:xxx 集中声明，建模侧序列化驱动）
// 注意：本表无 builtin 保护——声明式元数据，治理员可编辑/删除内置项
type OntAnnotationProperty struct {
	global.GVA_MODEL
	LocalName   string `json:"localName" form:"localName" gorm:"column:local_name;size:64;not null;comment:注释属性名（唯一）" binding:"required"`
	Label       string `json:"label" form:"label" gorm:"column:label;size:128;not null;comment:显示名" binding:"required"`
	RangeXsd    string `json:"rangeXsd" form:"rangeXsd" gorm:"column:range_xsd;size:64;comment:值域XSD类型"`
	AppliesTo   string `json:"appliesTo" form:"appliesTo" gorm:"column:applies_to;size:32;not null;default:all;comment:作用对象" binding:"required,oneof=class datatypeProperty objectProperty individual all"`
	Description string `json:"description" form:"description" gorm:"column:description;size:512;comment:描述"`
	Sort        int    `json:"sort" form:"sort" gorm:"column:sort;default:0;comment:排序"`
	CreatedBy   string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy   string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (OntAnnotationProperty) TableName() string {
	return "ont_annotation_properties"
}
