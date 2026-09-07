package report

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type DataSetRouter struct{}

// InitDataSetRouter 初始化报表平台·数据集管理路由
func (r *DataSetRouter) InitDataSetRouter(Router *gin.RouterGroup) {
	priv := Router.Group("report")
	write := priv.Group("dataSet").Use(middleware.OperationRecord())
	read := priv.Group("dataSet")
	{
		write.POST("createDataSet", dataSetApi.CreateDataSet)   // 新增（主子表事务）
		write.PUT("updateDataSet", dataSetApi.UpdateDataSet)    // 更新（子表先删后插）
		write.DELETE("deleteDataSet", dataSetApi.DeleteDataSet) // 删除（级联子表/引用校验）
	}
	{
		read.GET("getDataSetList", dataSetApi.GetDataSetList)          // 分页
		read.GET("findDataSet", dataSetApi.FindDataSet)                // 详情（含参数/转换）
		read.GET("getDataSetAll", dataSetApi.GetDataSetAll)            // 已启用全量下拉
		read.POST("testDataSetPreview", dataSetApi.TestDataSetPreview) // 测试预览（查询性质，不记操作日志）
	}
}
