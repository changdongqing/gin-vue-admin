// Package report 报表平台领域路由聚合
package report

import api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"

// RouterGroup 报表平台领域路由聚合
type RouterGroup struct {
	DataSourceRouter
	DataSetRouter
}

var (
	dsApi      = api.ApiGroupApp.ReportApiGroup.DataSourceApi
	dataSetApi = api.ApiGroupApp.ReportApiGroup.DataSetApi
)
