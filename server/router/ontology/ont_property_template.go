package ontology

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type PropertyTemplateRouter struct{}

// InitPropertyTemplateRouter 初始化属性模板路由（JWT + casbin 已在 PrivateGroup 挂载）
func (r *PropertyTemplateRouter) InitPropertyTemplateRouter(Router *gin.RouterGroup) {
	priv := Router.Group("ontology")
	write := priv.Group("propertyTemplate").Use(middleware.OperationRecord())
	read := priv.Group("propertyTemplate")
	{
		write.POST("createPropertyTemplate", ontApi.CreatePropertyTemplate)   // 新建属性模板
		write.PUT("updatePropertyTemplate", ontApi.UpdatePropertyTemplate)    // 更新属性模板
		write.PUT("disablePropertyTemplate", ontApi.DisablePropertyTemplate)  // 弃用/取消弃用
		write.DELETE("deletePropertyTemplate", ontApi.DeletePropertyTemplate) // 删除属性模板
		write.POST("promotePropertyTemplate", ontApi.PromotePropertyTemplate) // 提升（FR-6 预留）
	}
	{
		read.GET("findPropertyTemplate", ontApi.FindPropertyTemplate)           // 详情
		read.GET("getPropertyTemplateList", ontApi.GetPropertyTemplateList)     // 分页
		read.GET("getPropertyTemplateAll", ontApi.GetPropertyTemplateAll)       // 全量下拉
		read.GET("checkPropertyTemplateCode", ontApi.CheckPropertyTemplateCode) // 编码查重
	}
}
