package ontology

import (
	"github.com/flipped-aurora/gin-vue-admin/server/service"
)

type ApiGroup struct {
	PropertyTemplateApi
	ClassTemplateApi
	ClassificationRuleApi
	UnitApi
	QuantityKindApi
	AnnotationPropertyApi
	SupplyApi
	ModelProjectApi
	ModelPrefixApi
	ModelClassApi
	ModelDatatypePropertyApi
	ModelObjectPropertyApi
	ExtModuleApi
	ExtTableApi
	ExtBindingApi
	ExtSyncApi
}

var (
	PropertyTemplateService   = service.ServiceGroupApp.OntologyServiceGroup.PropertyTemplateService
	ClassTemplateService      = service.ServiceGroupApp.OntologyServiceGroup.ClassTemplateService
	ClassificationRuleService = service.ServiceGroupApp.OntologyServiceGroup.ClassificationRuleService
	UnitService               = service.ServiceGroupApp.OntologyServiceGroup.UnitService
	QuantityKindService       = service.ServiceGroupApp.OntologyServiceGroup.QuantityKindService
	AnnotationPropertyService = service.ServiceGroupApp.OntologyServiceGroup.AnnotationPropertyService
	ModelProjectService       = service.ServiceGroupApp.OntologyServiceGroup.ModelProjectService
	ModelPrefixService        = service.ServiceGroupApp.OntologyServiceGroup.ModelPrefixService
	ModelClassService         = service.ServiceGroupApp.OntologyServiceGroup.ModelClassService
	ClassInstantiationService = service.ServiceGroupApp.OntologyServiceGroup.ClassInstantiationService
	DatatypePropertyService   = service.ServiceGroupApp.OntologyServiceGroup.DatatypePropertyService
	ObjectPropertyService     = service.ServiceGroupApp.OntologyServiceGroup.ObjectPropertyService
	ExtModuleService          = service.ServiceGroupApp.OntologyServiceGroup.ExtModuleService
	ExtTableService           = service.ServiceGroupApp.OntologyServiceGroup.ExtTableService
	ExtBindingService         = service.ServiceGroupApp.OntologyServiceGroup.ExtBindingService
	ExtSyncService            = service.ServiceGroupApp.OntologyServiceGroup.ExtSyncService
)
