package ontology

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type ExtModuleRouter struct{}

// InitExtModuleRouter 初始化本体建模·外部模块注册路由
func (r *ExtModuleRouter) InitExtModuleRouter(Router *gin.RouterGroup) {
	priv := Router.Group("ontology")
	write := priv.Group("extModule").Use(middleware.OperationRecord())
	read := priv.Group("extModule")
	{
		write.POST("createExtModule", emApi.CreateExtModule)   // 新增模块
		write.PUT("updateExtModule", emApi.UpdateExtModule)    // 更新模块
		write.DELETE("deleteExtModule", emApi.DeleteExtModule) // 删除（有表拒绝）
	}
	{
		read.GET("getExtModuleList", emApi.GetExtModuleList) // 模块列表（含 tableCount）
	}
}

type ExtTableRouter struct{}

// InitExtTableRouter 初始化本体建模·外部表注册路由
func (r *ExtTableRouter) InitExtTableRouter(Router *gin.RouterGroup) {
	priv := Router.Group("ontology")
	write := priv.Group("extTable").Use(middleware.OperationRecord())
	read := priv.Group("extTable")
	{
		write.POST("registerExtTables", etApi.RegisterExtTables) // 探测勾选批量注册
		write.DELETE("deleteExtTable", etApi.DeleteExtTable)     // 删除（被绑定引用拒绝）
	}
	{
		read.GET("getExtTableList", etApi.GetExtTableList)       // 模块下注册表清单
		read.GET("probeExtTables", etApi.ProbeExtTables)         // 探测物理表
		read.GET("getExtTableColumns", etApi.GetExtTableColumns) // 探测表列
	}
}

type ExtBindingRouter struct{}

// InitExtBindingRouter 初始化本体建模·类绑定路由
func (r *ExtBindingRouter) InitExtBindingRouter(Router *gin.RouterGroup) {
	priv := Router.Group("ontology")
	write := priv.Group("extBinding").Use(middleware.OperationRecord())
	read := priv.Group("extBinding")
	{
		write.POST("createExtBinding", ebApi.CreateExtBinding)            // 保存草稿
		write.POST("updateExtBinding", ebApi.UpdateExtBinding)            // 更新草稿
		write.DELETE("deleteExtBinding", ebApi.DeleteExtBinding)          // 删除
		write.PUT("changeExtBindingStatus", ebApi.ChangeExtBindingStatus) // 生效（校验链）/停用
	}
	{
		read.GET("getExtBindingList", ebApi.GetExtBindingList) // 绑定分页（富化）
		read.GET("findExtBinding", ebApi.FindExtBinding)       // 绑定详情（双子表）
	}
}

type ExtSyncRouter struct{}

// InitExtSyncRouter 初始化本体建模·同步中心路由
func (r *ExtSyncRouter) InitExtSyncRouter(Router *gin.RouterGroup) {
	priv := Router.Group("ontology")
	write := priv.Group("extSync").Use(middleware.OperationRecord())
	read := priv.Group("extSync")
	{
		write.POST("triggerExtSync", esApi.TriggerExtSync) // 触发同步（1全量/2增量）
	}
	{
		read.POST("dryRunExtSync", esApi.DryRunExtSync)        // 试运行（只读预览，不挂 OperationRecord）
		read.GET("getExtSyncLogList", esApi.GetExtSyncLogList) // 同步日志分页
	}
}
