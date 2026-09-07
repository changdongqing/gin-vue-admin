package report

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type DataSourceRouter struct{}

// InitDataSourceRouter 初始化报表平台·数据源管理路由
func (r *DataSourceRouter) InitDataSourceRouter(Router *gin.RouterGroup) {
	priv := Router.Group("report")
	write := priv.Group("dataSource").Use(middleware.OperationRecord())
	read := priv.Group("dataSource")
	{
		write.POST("createDataSource", dsApi.CreateDataSource)   // 创建
		write.PUT("updateDataSource", dsApi.UpdateDataSource)    // 更新（类型不可变/密码三态）
		write.DELETE("deleteDataSource", dsApi.DeleteDataSource) // 删除（引用校验/销毁池）
	}
	{
		read.GET("getDataSourceList", dsApi.GetDataSourceList)      // 分页
		read.GET("findDataSource", dsApi.FindDataSource)            // 详情
		read.GET("getDataSourceAll", dsApi.GetDataSourceAll)        // 已启用全量下拉
		read.POST("testDataSourceConnection", dsApi.TestConnection) // 测试连接（查询性质，不记操作日志）
	}
}
