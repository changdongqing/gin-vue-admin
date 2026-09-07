package report

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type ExcelReportRouter struct{}

// InitExcelReportRouter 初始化报表平台·Excel报表设计与模板管理路由
func (r *ExcelReportRouter) InitExcelReportRouter(Router *gin.RouterGroup) {
	priv := Router.Group("report")
	write := priv.Group("excelReport").Use(middleware.OperationRecord())
	read := priv.Group("excelReport")
	{
		write.POST("createExcelReport", excelApi.CreateExcelReport)             // 新增元数据
		write.PUT("updateExcelReport", excelApi.UpdateExcelReport)              // 更新元数据
		write.DELETE("deleteExcelReport", excelApi.DeleteExcelReport)           // 删除（级联模板）
		write.POST("copyExcelReport", excelApi.CopyExcelReport)                 // 复制（元数据+模板深拷贝）
		write.POST("saveExcelTemplate", excelApi.SaveExcelTemplate)             // 保存模板（仅 jsonStr/setParam）
		write.POST("bindExcelReportDataSets", excelApi.BindExcelReportDataSets) // 关联数据集（仅 setCodes）
	}
	{
		read.GET("getExcelReportList", excelApi.GetExcelReportList)                   // 分页（含 reportCode 精确过滤）
		read.GET("findExcelReport", excelApi.FindExcelReport)                         // 详情（含模板）
		read.GET("getExcelReportAll", excelApi.GetExcelReportAll)                     // 全量下拉（预留）
		read.GET("getExcelReportDataSetFields", excelApi.GetExcelReportDataSetFields) // 设计器左栏字段列表
	}
}
