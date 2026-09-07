package report

import (
	"github.com/flipped-aurora/gin-vue-admin/server/service"
)

type ApiGroup struct {
	DataSourceApi
}

var (
	DataSourceService = service.ServiceGroupApp.ReportServiceGroup.DataSourceService
)
