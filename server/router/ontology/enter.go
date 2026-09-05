package ontology

import api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"

// RouterGroup 本体治理领域路由聚合
type RouterGroup struct {
	PropertyTemplateRouter
	SupplyRouter
}

var (
	ontApi    = api.ApiGroupApp.OntologyApiGroup.PropertyTemplateApi
	supplyApi = api.ApiGroupApp.OntologyApiGroup.SupplyApi
)
