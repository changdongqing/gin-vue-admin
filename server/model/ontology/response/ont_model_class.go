package response

import (
	ontModel "github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
)

// ClassSummary 类摘要（详情父子类清单）
type ClassSummary struct {
	ID        uint   `json:"ID"`
	ClassIri  string `json:"classIri"`
	LocalName string `json:"localName"`
	Label     string `json:"label"`
	LabelCn   string `json:"labelCn"`
}

// PropertySummary 属性摘要（详情属性清单行）
type PropertySummary struct {
	ID           uint   `json:"ID"`
	PropertyIri  string `json:"propertyIri"`
	LocalName    string `json:"localName"`
	Label        string `json:"label"`
	TemplateCode string `json:"templateCode"`
	TypeOrRange  string `json:"typeOrRange"` // datatype=xsdType / object=range 类名（空=未定）
}

// ClassDetail 类详情聚合（类 + 双属性清单 + 直接父子类）
type ClassDetail struct {
	ontModel.OntModelClass
	DatatypeProperties []PropertySummary `json:"datatypeProperties"`
	ObjectProperties   []PropertySummary `json:"objectProperties"`
	ParentClasses      []ClassSummary    `json:"parentClasses"`
	ChildClasses       []ClassSummary    `json:"childClasses"`
}

// PropertyPreview 实例化预览属性行
type PropertyPreview struct {
	PropertyTemplateCode string `json:"propertyTemplateCode"`
	RefType              string `json:"refType"` // property/relationship
	Source               string `json:"source"`  // node/inherited/overridden
	Kind                 string `json:"kind"`    // datatype/object
	Label                string `json:"label"`
	Type                 string `json:"type"` // datatype=xsdType
	UnitRef              string `json:"unitRef"`
}

// InstantiatePreview 模板实例化预览（继承视图 + 模板批量富化，只读不落库）
type InstantiatePreview struct {
	TemplateCode       string            `json:"templateCode"`
	ClassificationCode string            `json:"classificationCode"`
	Icon               string            `json:"icon"`
	Color              string            `json:"color"`
	Properties         []PropertyPreview `json:"properties"`
}

// SuggestInverseResult 反向关系建议（前端确认后创建反向属性并双向 inverseOf 互指）
type SuggestInverseResult struct {
	SuggestedLocalName     string `json:"suggestedLocalName"`
	SuggestedLabel         string `json:"suggestedLabel"`
	SuggestedDomainClassId uint   `json:"suggestedDomainClassId"` //nolint:stylecheck // 对齐前端 domainClassId
	SuggestedRangeClassId  uint   `json:"suggestedRangeClassId"`  //nolint:stylecheck // 对齐前端 rangeClassId
}
