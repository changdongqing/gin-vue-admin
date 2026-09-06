package ontology

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type ModelProjectRouter struct{}

// InitModelProjectRouter 初始化本体建模·项目管理路由
func (r *ModelProjectRouter) InitModelProjectRouter(Router *gin.RouterGroup) {
	priv := Router.Group("ontology")
	write := priv.Group("modelProject").Use(middleware.OperationRecord())
	read := priv.Group("modelProject")
	{
		write.POST("createModelProject", mpApi.CreateModelProject)   // 创建
		write.PUT("updateModelProject", mpApi.UpdateModelProject)    // 更新（archived 拒绝/状态机）
		write.DELETE("deleteModelProject", mpApi.DeleteModelProject) // 删除（archived 拒绝/有类拒绝）
	}
	{
		read.GET("findModelProject", mpApi.FindModelProject)           // 详情
		read.GET("getModelProjectList", mpApi.GetModelProjectList)     // 分页
		read.GET("getModelProjectAll", mpApi.GetModelProjectAll)       // 全量下拉
		read.GET("checkModelProjectCode", mpApi.CheckModelProjectCode) // 编码查重
	}
}

type ModelPrefixRouter struct{}

// InitModelPrefixRouter 初始化本体建模·IRI前缀注册路由
func (r *ModelPrefixRouter) InitModelPrefixRouter(Router *gin.RouterGroup) {
	priv := Router.Group("ontology")
	write := priv.Group("modelPrefix").Use(middleware.OperationRecord())
	read := priv.Group("modelPrefix")
	{
		write.POST("createModelPrefix", mpPrefixApi.CreateModelPrefix)   // 新增前缀
		write.PUT("updateModelPrefix", mpPrefixApi.UpdateModelPrefix)    // 更新前缀
		write.DELETE("deleteModelPrefix", mpPrefixApi.DeleteModelPrefix) // 删除前缀
	}
	{
		read.GET("getModelPrefixList", mpPrefixApi.GetModelPrefixList) // 项目前缀列表
	}
}
