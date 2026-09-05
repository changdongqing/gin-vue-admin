package ontology

import (
	"github.com/flipped-aurora/gin-vue-admin/server/service"
)

type ApiGroup struct {
	PropertyTemplateApi
	SupplyApi
}

var PropertyTemplateService = service.ServiceGroupApp.OntologyServiceGroup.PropertyTemplateService
