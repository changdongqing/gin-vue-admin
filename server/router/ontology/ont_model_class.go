package ontology

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type ModelClassRouter struct{}

// InitModelClassRouter 初始化本体建模·类建模路由
func (r *ModelClassRouter) InitModelClassRouter(Router *gin.RouterGroup) {
	priv := Router.Group("ontology")
	write := priv.Group("modelClass").Use(middleware.OperationRecord())
	read := priv.Group("modelClass")
	{
		write.POST("createModelClass", mcApi.CreateModelClass)           // 创建（IRI 生成/查重）
		write.PUT("updateModelClass", mcApi.UpdateModelClass)            // 更新（编辑白名单）
		write.DELETE("deleteModelClass", mcApi.DeleteModelClass)         // 删除（三重守卫）
		write.POST("instantiateModelClass", mcApi.InstantiateModelClass) // 模板实例化（单事务）
	}
	{
		read.GET("findModelClass", mcApi.FindModelClass)                             // 详情（单实体）
		read.GET("getModelClassDetail", mcApi.GetModelClassDetail)                   // 详情聚合（属性+父子类）
		read.GET("getModelClassList", mcApi.GetModelClassList)                       // 分页
		read.GET("getModelClassByProject", mcApi.GetModelClassByProject)             // 项目内全量（03 选类）
		read.GET("checkModelClassLocalName", mcApi.CheckModelClassLocalName)         // 本地名查重
		read.GET("previewInstantiateModelClass", mcApi.PreviewInstantiateModelClass) // 实例化预览
	}
}

type ModelDatatypePropertyRouter struct{}

// InitModelDatatypePropertyRouter 初始化本体建模·数据属性路由
func (r *ModelDatatypePropertyRouter) InitModelDatatypePropertyRouter(Router *gin.RouterGroup) {
	priv := Router.Group("ontology")
	write := priv.Group("modelDatatypeProperty").Use(middleware.OperationRecord())
	read := priv.Group("modelDatatypeProperty")
	{
		write.POST("createModelDatatypeProperty", mdpApi.CreateModelDatatypeProperty)           // 创建
		write.PUT("updateModelDatatypeProperty", mdpApi.UpdateModelDatatypeProperty)            // 更新
		write.DELETE("deleteModelDatatypeProperty", mdpApi.DeleteModelDatatypeProperty)         // 删除
		write.POST("instantiateModelDatatypeProperty", mdpApi.InstantiateModelDatatypeProperty) // 单模板挂载
	}
	{
		read.GET("findModelDatatypeProperty", mdpApi.FindModelDatatypeProperty)       // 详情
		read.GET("getModelDatatypePropertyList", mdpApi.GetModelDatatypePropertyList) // 分页
	}
}

type ModelObjectPropertyRouter struct{}

// InitModelObjectPropertyRouter 初始化本体建模·对象属性路由
func (r *ModelObjectPropertyRouter) InitModelObjectPropertyRouter(Router *gin.RouterGroup) {
	priv := Router.Group("ontology")
	write := priv.Group("modelObjectProperty").Use(middleware.OperationRecord())
	read := priv.Group("modelObjectProperty")
	{
		write.POST("createModelObjectProperty", mopApi.CreateModelObjectProperty)           // 创建（range 可空/inverseOf 互指）
		write.PUT("updateModelObjectProperty", mopApi.UpdateModelObjectProperty)            // 更新
		write.DELETE("deleteModelObjectProperty", mopApi.DeleteModelObjectProperty)         // 删除（清互指）
		write.POST("instantiateModelObjectProperty", mopApi.InstantiateModelObjectProperty) // 单模板挂载
	}
	{
		read.GET("findModelObjectProperty", mopApi.FindModelObjectProperty)                     // 详情
		read.GET("getModelObjectPropertyList", mopApi.GetModelObjectPropertyList)               // 分页
		read.GET("suggestInverseModelObjectProperty", mopApi.SuggestModelObjectPropertyInverse) // 反向关系建议
	}
}
