package response

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
)

// ClassTemplatePageItem 分页列表项（平铺展示用富化字段）
type ClassTemplatePageItem struct {
	ontology.OntClassTemplate
	ParentLabel   string `json:"parentLabel"`   // 父节点 label（顶级为空）
	TreeRootLabel string `json:"treeRootLabel"` // 分类树中文名（取该树根节点 label）
	HasChildren   bool   `json:"hasChildren"`   // 是否存在子节点（删除按钮禁用判断）
}

// InheritedRefItem 继承视图属性项
type InheritedRefItem struct {
	PropertyTemplateCode string `json:"propertyTemplateCode"`
	RefType              string `json:"refType"`
	SortOrder            int    `json:"sortOrder"`
	Source               string `json:"source"` // node(本节点声明)/inherited(继承)/overridden(覆盖父级同名)
}

// ClassTemplateInheritedView 继承视图（FR-2 核心：结构骨架 + 外观合并视图）
type ClassTemplateInheritedView struct {
	TemplateCode     string             `json:"templateCode"`
	Label            string             `json:"label"`
	Properties       []InheritedRefItem `json:"properties"`
	Icon             string             `json:"icon"`
	Color            string             `json:"color"`
	AppearanceSource string             `json:"appearanceSource"` // node/inherited
}

// ClassTreeRootItem 分类树下拉项
type ClassTreeRootItem struct {
	TreeRoot string `json:"treeRoot"`
	Label    string `json:"label"`
}

// ClassHierarchySuggest 类层级建议（FR-9 供给；本期无镜像数据，返回父分类模板信息供参考）
type ClassHierarchySuggest struct {
	SuggestedParentTemplateCode string `json:"suggestedParentTemplateCode"`
	SuggestedParentLabel        string `json:"suggestedParentLabel"`
	Note                        string `json:"note"`
}
