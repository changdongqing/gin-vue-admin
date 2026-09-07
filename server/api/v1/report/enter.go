package report

import (
	"github.com/flipped-aurora/gin-vue-admin/server/service"
	"github.com/flipped-aurora/gin-vue-admin/server/service/report"
)

type ApiGroup struct {
	DataSourceApi
	DataSetApi
	ExcelReportApi
}

var (
	DataSourceService  = service.ServiceGroupApp.ReportServiceGroup.DataSourceService
	DataSetService     = service.ServiceGroupApp.ReportServiceGroup.DataSetService
	ExcelReportService = service.ServiceGroupApp.ReportServiceGroup.ExcelReportService
	// QueryServiceApp 数据集查询执行服务（04 渲染引擎复用同一实例）
	QueryServiceApp = &report.DataSetQueryService{}
)
