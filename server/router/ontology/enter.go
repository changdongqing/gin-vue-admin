// Package ontology 本体治理领域路由聚合
package ontology

import api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"

// RouterGroup 本体治理领域路由聚合
type RouterGroup struct {
	PropertyTemplateRouter
	ClassTemplateRouter
	UnitRouter
	AnnotationPropertyRouter
	SupplyRouter
	ModelProjectRouter
	ModelPrefixRouter
}

var (
	ontApi          = api.ApiGroupApp.OntologyApiGroup.PropertyTemplateApi
	classApi        = api.ApiGroupApp.OntologyApiGroup.ClassTemplateApi
	ruleApi         = api.ApiGroupApp.OntologyApiGroup.ClassificationRuleApi
	unitApi         = api.ApiGroupApp.OntologyApiGroup.UnitApi
	quantityKindApi = api.ApiGroupApp.OntologyApiGroup.QuantityKindApi
	apApi           = api.ApiGroupApp.OntologyApiGroup.AnnotationPropertyApi
	supplyApi       = api.ApiGroupApp.OntologyApiGroup.SupplyApi
	mpApi           = api.ApiGroupApp.OntologyApiGroup.ModelProjectApi
	mpPrefixApi     = api.ApiGroupApp.OntologyApiGroup.ModelPrefixApi
)
