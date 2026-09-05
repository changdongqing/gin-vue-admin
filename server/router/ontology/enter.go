package ontology

import api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"

// RouterGroup 本体治理领域路由聚合
type RouterGroup struct {
	PropertyTemplateRouter
	ClassTemplateRouter
	SupplyRouter
}

var (
	ontApi    = api.ApiGroupApp.OntologyApiGroup.PropertyTemplateApi
	classApi  = api.ApiGroupApp.OntologyApiGroup.ClassTemplateApi
	ruleApi   = api.ApiGroupApp.OntologyApiGroup.ClassificationRuleApi
	supplyApi = api.ApiGroupApp.OntologyApiGroup.SupplyApi
)
