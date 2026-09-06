package ontology

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// OntModelProject 本体项目（建模域顶层容器：命名空间基址/前缀注册/序列化策略）
// 唯一业务索引用迁移 SQL 的部分唯一索引（WHERE deleted_at IS NULL）+ Service 查重共同保证，
// 此处不声明 uniqueIndex tag（AutoMigrate 对 PG 部分索引支持有限，结构以迁移为准）
type OntModelProject struct {
	global.GVA_MODEL
	ProjectCode           string `json:"projectCode" form:"projectCode" gorm:"column:project_code;size:64;not null;comment:项目编码（唯一）" binding:"required"`
	Name                  string `json:"name" form:"name" gorm:"column:name;size:128;not null;comment:项目名称" binding:"required"`
	Description           string `json:"description" form:"description" gorm:"column:description;size:512;comment:描述"`
	NamespaceBase         string `json:"namespaceBase" form:"namespaceBase" gorm:"column:namespace_base;size:255;not null;comment:IRI命名空间基址" binding:"required"`
	DefaultFormat         string `json:"defaultFormat" form:"defaultFormat" gorm:"column:default_format;size:16;not null;default:TTL;comment:默认序列化格式 TTL/OWL_XML" binding:"omitempty,oneof=TTL OWL_XML"`
	SerializationStrategy string `json:"serializationStrategy" form:"serializationStrategy" gorm:"column:serialization_strategy;size:8;not null;default:B;comment:序列化策略 B=独立副本/A=共享属性(预留禁用)" binding:"omitempty,oneof=B A"`
	Status                string `json:"status" form:"status" gorm:"column:status;size:16;not null;default:draft;comment:项目状态 draft/active/archived" binding:"omitempty,oneof=draft active archived"`
	CreatedBy             string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy             string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (OntModelProject) TableName() string {
	return "ont_model_projects"
}

// OntModelPrefix 本体项目 IRI 前缀注册（序列化 @prefix 声明来源）
type OntModelPrefix struct {
	global.GVA_MODEL
	ProjectId uint   `json:"projectId" form:"projectId" gorm:"column:project_id;not null;comment:所属项目ID" binding:"required"` //nolint:stylecheck // 对齐前端 projectId
	Prefix    string `json:"prefix" form:"prefix" gorm:"column:prefix;size:64;not null;comment:前缀名(NCName)" binding:"required"`
	Namespace string `json:"namespace" form:"namespace" gorm:"column:namespace;size:255;not null;comment:命名空间URI" binding:"required"`
	IsDefault int    `json:"isDefault" form:"isDefault" gorm:"column:is_default;default:0;comment:默认前缀0/1"`
	CreatedBy string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (OntModelPrefix) TableName() string {
	return "ont_model_prefixes"
}
