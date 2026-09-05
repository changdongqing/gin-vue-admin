package ontology

// ServiceGroup 本体治理领域服务聚合
type ServiceGroup struct {
	PropertyTemplateService
	ClassTemplateService
	ClassificationRuleService
	QuantityKindService
	UnitService
	AnnotationPropertyService
}
