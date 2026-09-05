package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

// SearchClassTemplate 分类模板分页查询（treeRoot 精确 + 编码右模糊 + 名称模糊）
type SearchClassTemplate struct {
	request.PageInfo
	TreeRoot           string `form:"treeRoot"`
	ClassificationCode string `form:"classificationCode"` // 右模糊（前缀匹配）
	Name               string `form:"name"`               // 模糊匹配 label/label_cn
}

// ClassTemplateOps 单条操作（删除/详情/弃用）
type ClassTemplateOps struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// RefListOps 骨架子表回填
type RefListOps struct {
	ClassTemplateId uint `form:"classTemplateId" binding:"required"` //nolint:stylecheck // 对齐前端 classTemplateId
}

// PreviewCodeOps 编码预览
type PreviewCodeOps struct {
	ParentId uint   `form:"parentId"`
	TreeRoot string `form:"treeRoot" binding:"required"`
}

// RuleOps 编码规则查询
type RuleOps struct {
	TreeRoot string `form:"treeRoot" binding:"required"`
}

// InheritedOps 继承视图查询
type InheritedOps struct {
	TemplateCode string `form:"templateCode" binding:"required"`
}

// SuggestOps 类层级建议查询（FR-9 供给）
type SuggestOps struct {
	TemplateCode string `form:"templateCode" binding:"required"`
}
