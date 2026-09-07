package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

// SearchExcelReport Excel 报表分页查询
type SearchExcelReport struct {
	request.PageInfo
	ReportGroup string `form:"reportGroup"`
	ReportCode  string `form:"reportCode"` // 精确过滤（设计器/预览页定位，对齐实现无 get-by-code）
}

// ExcelReportOps 单条操作（详情/删除）
type ExcelReportOps struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// CopyExcelReportReq 复制报表
type CopyExcelReportReq struct {
	SourceReportCode string `json:"sourceReportCode" binding:"required"`
	ReportCode       string `json:"reportCode" binding:"required"`
	ReportName       string `json:"reportName" binding:"required"`
}

// SaveExcelTemplateReq 保存模板（仅 jsonStr/setParam；SetCodes 兼容保留但服务端忽略——
// 由 bindExcelReportDataSets 独立维护，打破循环依赖）
type SaveExcelTemplateReq struct {
	ReportCode string `json:"reportCode" binding:"required"`
	SetCodes   string `json:"setCodes"`
	SetParam   string `json:"setParam"`
	JsonStr    string `json:"jsonStr" binding:"required"`
}

// BindDataSetsReq 设计器「关联数据集」弹窗提交（仅 setCodes）
type BindDataSetsReq struct {
	ReportCode string   `json:"reportCode" binding:"required"`
	SetCodes   []string `json:"setCodes"`
}

// PreviewExcelReportReq 预览请求（paramValues 平铺；dateRange 为 "起,止" 字符串，后端拆分）
type PreviewExcelReportReq struct {
	ReportCode  string                 `json:"reportCode" binding:"required"`
	ParamValues map[string]interface{} `json:"paramValues"`
	PageNo      int                    `json:"pageNo"`   // 默认 1
	PageSize    int                    `json:"pageSize"` // 默认 20
}
