package ontology

import (
	"github.com/gin-gonic/gin"
)

type SupplyRouter struct{}

// InitSupplyRouter 初始化本体供给路由（只读、GET、版本化路径，与治理 CRUD 权限分离）
// 后续 03/04 的供给端点（units、annotationProperties 等）在此追加
func (r *SupplyRouter) InitSupplyRouter(Router *gin.RouterGroup) {
	supply := Router.Group("ontology/supply/v1")
	{
		supply.GET("propertyTemplates", supplyApi.GetPropertyTemplatesForSupply)
		supply.GET("classTemplate/tree", supplyApi.GetClassTemplateTreeForSupply)
		supply.GET("classHierarchy/suggest", supplyApi.SuggestClassHierarchy)
		supply.GET("units", supplyApi.GetUnitsForSupply)
		supply.GET("units/convert", supplyApi.ConvertUnitForSupply)
		supply.GET("annotationProperties", supplyApi.GetAnnotationPropertiesForSupply)
	}
}
