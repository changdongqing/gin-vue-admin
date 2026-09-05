package ontology

import (
	"github.com/flipped-aurora/gin-vue-admin/server/service"
)

type ApiGroup struct {
	PropertyTemplateApi
	ClassTemplateApi
	ClassificationRuleApi
	SupplyApi
}

var (
	PropertyTemplateService   = service.ServiceGroupApp.OntologyServiceGroup.PropertyTemplateService
	ClassTemplateService      = service.ServiceGroupApp.OntologyServiceGroup.ClassTemplateService
	ClassificationRuleService = service.ServiceGroupApp.OntologyServiceGroup.ClassificationRuleService
)
