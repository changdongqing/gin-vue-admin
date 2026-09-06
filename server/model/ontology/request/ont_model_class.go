package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

// SearchModelClass 本体类分页查询
type SearchModelClass struct {
	request.PageInfo
	ProjectId   uint   `form:"projectId"`   //nolint:stylecheck // 对齐前端 projectId
	Keyword     string `form:"keyword"`     // label/labelCn/localName 模糊
	HasTemplate string `form:"hasTemplate"` // "true"/"false"/"" 三态
}

// ModelClassOps 单条操作（删除/详情）
type ModelClassOps struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// ModelClassByProjectOps 项目内全量类（03 选类用）
type ModelClassByProjectOps struct {
	ProjectId uint `form:"projectId" binding:"required"` //nolint:stylecheck // 对齐前端 projectId
}

// CheckModelClassLocalNameOps 类本地名查重
type CheckModelClassLocalNameOps struct {
	ProjectId uint   `form:"projectId" binding:"required"` //nolint:stylecheck // 对齐前端 projectId
	LocalName string `form:"localName" binding:"required"`
	ExcludeId uint   `form:"excludeId"` //nolint:stylecheck // 对齐前端 excludeId
}

// InstantiateClassOps 类级模板实例化
type InstantiateClassOps struct {
	ProjectId    uint   `json:"projectId" form:"projectId" binding:"required"` //nolint:stylecheck // 对齐前端 projectId
	ClassId      uint   `json:"classId" form:"classId" binding:"required"`
	TemplateCode string `json:"templateCode" form:"templateCode" binding:"required"`
}

// PreviewInstantiateOps 模板实例化预览（只读）
type PreviewInstantiateOps struct {
	ProjectId    uint   `form:"projectId" binding:"required"` //nolint:stylecheck // 对齐前端 projectId
	TemplateCode string `form:"templateCode" binding:"required"`
}

// SearchDatatypeProperty 数据属性分页查询
type SearchDatatypeProperty struct {
	request.PageInfo
	ProjectId    uint   `form:"projectId"` //nolint:stylecheck // 对齐前端 projectId
	ClassId      uint   `form:"classId"`
	Keyword      string `form:"keyword"` // localName/label 模糊
	XsdType      string `form:"xsdType"`
	TemplateCode string `form:"templateCode"`
}

// SearchObjectProperty 对象属性分页查询
type SearchObjectProperty struct {
	request.PageInfo
	ProjectId     uint   `form:"projectId"`     //nolint:stylecheck // 对齐前端 projectId
	DomainClassId uint   `form:"domainClassId"` //nolint:stylecheck // 对齐前端 domainClassId
	Keyword       string `form:"keyword"`       // localName/label 模糊
	TemplateCode  string `form:"templateCode"`
	RangeClassId  uint   `form:"rangeClassId"` //nolint:stylecheck // 对齐前端 rangeClassId
}

// ModelPropertyOps 属性单条操作（删除/详情/反向建议）
type ModelPropertyOps struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// InstantiatePropertyOps 属性级模板实例化（从属性库挂载）
type InstantiatePropertyOps struct {
	ProjectId    uint   `json:"projectId" form:"projectId" binding:"required"` //nolint:stylecheck // 对齐前端 projectId
	ClassId      uint   `json:"classId" form:"classId" binding:"required"`
	TemplateCode string `json:"templateCode" form:"templateCode" binding:"required"`
}
