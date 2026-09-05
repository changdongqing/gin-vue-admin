package ontology

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type UnitRouter struct{}

// InitUnitRouter 初始化单位/量纲路由
func (r *UnitRouter) InitUnitRouter(Router *gin.RouterGroup) {
	priv := Router.Group("ontology")
	write := priv.Group("unit").Use(middleware.OperationRecord())
	read := priv.Group("unit")
	{
		write.POST("createUnit", unitApi.CreateUnit)   // 创建单位
		write.PUT("updateUnit", unitApi.UpdateUnit)    // 更新单位
		write.PUT("disableUnit", unitApi.DisableUnit)  // 停用/启用
		write.DELETE("deleteUnit", unitApi.DeleteUnit) // 删除单位
	}
	{
		read.GET("findUnit", unitApi.FindUnit)       // 详情
		read.GET("getUnitPage", unitApi.GetUnitPage) // 分页
		read.GET("getUnitAll", unitApi.GetUnitAll)   // 全量下拉
		read.GET("convertUnit", unitApi.ConvertUnit) // 换算试算
	}
	qk := priv.Group("quantityKind")
	{
		qk.GET("getQuantityKindList", quantityKindApi.GetQuantityKindList) // 量纲列表（只读）
	}
}
