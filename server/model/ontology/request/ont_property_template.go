package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

// SearchPropertyTemplate 属性模板分页查询
type SearchPropertyTemplate struct {
	request.PageInfo        // page/pageSize/keyword
	Kind             string `form:"kind"`       // datatype/object（可选）
	Category         string `form:"category"`   // 分组（可选）
	Source           string `form:"source"`     // builtin/custom（可选）
	Deprecated       *int   `form:"deprecated"` // 0/1（可选，指针区分未传与 0）
}

// PropertyTemplateOps 单条操作（删除/详情/弃用）
type PropertyTemplateOps struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// CheckPropertyTemplateCode 编码远程查重
type CheckPropertyTemplateCode struct {
	TemplateCode string `form:"templateCode" binding:"required"`
	ExcludeID    uint   `form:"excludeId"`
}

// PromotePropertyTemplate 属性快照提升（FR-6，建模侧调用；templateCode 空则从 name 派生 camelCase）
type PromotePropertyTemplate struct {
	TemplateCode       string `json:"templateCode"`
	Name               string `json:"name" binding:"required"`
	Kind               string `json:"kind"`
	Label              string `json:"label"`
	Alias              string `json:"alias"`
	Description        string `json:"description"`
	Category           string `json:"category"`
	Type               string `json:"type"`
	IsIdentifier       int    `json:"isIdentifier"`
	UnitRef            string `json:"unitRef"`
	Values             string `json:"values"`
	DefaultCardinality string `json:"defaultCardinality"`
}
