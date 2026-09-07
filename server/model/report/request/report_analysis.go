package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

// SearchAnalysis 分析报表分页查询
type SearchAnalysis struct {
	request.PageInfo
	ReportGroup string `form:"reportGroup"`
	SetCode     string `form:"setCode"`
	Status      *int   `form:"status"`
}

// AnalysisReportOps 单条操作（详情/删除）
type AnalysisReportOps struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// SaveAnalysisReq 元数据请求（Create/Update 共用；setCode 单选）
type SaveAnalysisReq struct {
	ID          uint   `json:"ID"`
	ReportCode  string `json:"reportCode" binding:"required"`
	ReportName  string `json:"reportName" binding:"required"`
	ReportGroup string `json:"reportGroup"`
	ReportDesc  string `json:"reportDesc"`
	SetCode     string `json:"setCode" binding:"required"`
	Status      *int   `json:"status"`
}

// CopyAnalysisReq 复制分析报表
type CopyAnalysisReq struct {
	SourceReportCode string `json:"sourceReportCode" binding:"required"`
	ReportCode       string `json:"reportCode" binding:"required"`
	ReportName       string `json:"reportName" binding:"required"`
}

// SaveAnalysisConfigReq 保存设计器配置（configJson 必填；setParam 预留）
type SaveAnalysisConfigReq struct {
	ReportCode string `json:"reportCode" binding:"required"`
	ConfigJson string `json:"configJson" binding:"required"`
	SetParam   string `json:"setParam"`
}

// PreviewAnalysisReq 预览取数（明细不聚合；超限回退分页截断置 truncated）
type PreviewAnalysisReq struct {
	ReportCode  string                 `json:"reportCode" binding:"required"`
	ParamValues map[string]interface{} `json:"paramValues"`
}

// ColumnMeta 字段类型推断（设计器维度/度量分区依据）
type ColumnMeta struct {
	Name string `json:"name"`
	Type string `json:"type"` // string / number / date
}
