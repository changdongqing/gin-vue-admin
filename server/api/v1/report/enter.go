package report

import (
	"github.com/flipped-aurora/gin-vue-admin/server/service"
	"github.com/flipped-aurora/gin-vue-admin/server/service/report"
)

type ApiGroup struct {
	DataSourceApi
	DataSetApi
	ExcelReportApi
	AnalysisReportApi
}

var (
	DataSourceService     = service.ServiceGroupApp.ReportServiceGroup.DataSourceService
	DataSetService        = service.ServiceGroupApp.ReportServiceGroup.DataSetService
	ExcelReportService    = service.ServiceGroupApp.ReportServiceGroup.ExcelReportService
	AnalysisReportService = service.ServiceGroupApp.ReportServiceGroup.AnalysisReportService
	// QueryServiceApp / RenderServiceApp 查询执行与渲染引擎（04 预览复用）
	QueryServiceApp  = &report.DataSetQueryService{}
	RenderServiceApp = &report.ExcelReportRenderService{}
)
