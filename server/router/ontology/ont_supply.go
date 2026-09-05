package ontology

import (
	"github.com/gin-gonic/gin"
)

type SupplyRouter struct{}

// InitSupplyRouter 初始化本体供给路由（只读、GET、版本化路径，与治理 CRUD 权限分离）
// 后续 02/03/04 的供给端点（classTemplate/tree、units、annotationProperties 等）在此追加
func (r *SupplyRouter) InitSupplyRouter(Router *gin.RouterGroup) {
	supply := Router.Group("ontology/supply/v1")
	{
		supply.GET("propertyTemplates", supplyApi.GetPropertyTemplatesForSupply)
	}
}
