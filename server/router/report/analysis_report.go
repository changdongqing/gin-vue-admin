package report

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type AnalysisReportRouter struct{}

// InitAnalysisReportRouter 初始化报表平台·分析报表路由
func (r *AnalysisReportRouter) InitAnalysisReportRouter(Router *gin.RouterGroup) {
	priv := Router.Group("report")
	write := priv.Group("analysisReport").Use(middleware.OperationRecord())
	read := priv.Group("analysisReport")
	{
		write.POST("createAnalysisReport", analysisApi.CreateAnalysisReport)   // 新增
		write.PUT("updateAnalysisReport", analysisApi.UpdateAnalysisReport)    // 更新
		write.DELETE("deleteAnalysisReport", analysisApi.DeleteAnalysisReport) // 删除（级联配置）
		write.POST("copyAnalysisReport", analysisApi.CopyAnalysisReport)       // 复制
		write.POST("saveAnalysisConfig", analysisApi.SaveAnalysisConfig)       // 保存设计器配置
	}
	{
		read.GET("getAnalysisReportList", analysisApi.GetAnalysisReportList)     // 分页
		read.GET("findAnalysisReport", analysisApi.FindAnalysisReport)           // 详情
		read.GET("getAnalysisReportByCode", analysisApi.GetAnalysisReportByCode) // 按编码详情
		read.POST("previewAnalysisReport", analysisApi.PreviewAnalysisReport)    // 预览取数（查询性质，不记操作日志）
	}
}
