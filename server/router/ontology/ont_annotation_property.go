package ontology

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type AnnotationPropertyRouter struct{}

// InitAnnotationPropertyRouter 初始化注释属性注册表路由
func (r *AnnotationPropertyRouter) InitAnnotationPropertyRouter(Router *gin.RouterGroup) {
	priv := Router.Group("ontology")
	write := priv.Group("annotationProperty").Use(middleware.OperationRecord())
	read := priv.Group("annotationProperty")
	{
		write.POST("createAnnotationProperty", apApi.CreateAnnotationProperty)          // 创建
		write.PUT("updateAnnotationProperty", apApi.UpdateAnnotationProperty)           // 更新
		write.DELETE("deleteAnnotationProperty", apApi.DeleteAnnotationProperty)        // 删除（无 builtin 保护）
		write.GET("exportAnnotationPropertyExcel", apApi.ExportAnnotationPropertyExcel) // 导出 xlsx
	}
	{
		read.GET("findAnnotationProperty", apApi.FindAnnotationProperty)       // 详情
		read.GET("getAnnotationPropertyList", apApi.GetAnnotationPropertyList) // 分页
		read.GET("getAnnotationPropertyAll", apApi.GetAnnotationPropertyAll)   // 全量
	}
}
