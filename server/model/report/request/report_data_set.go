package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/report"
)

// SearchDataSet 数据集分页查询
type SearchDataSet struct {
	request.PageInfo
	SetType    string `form:"setType"` // sql/http，空=全部
	SourceCode string `form:"sourceCode"`
	EnableFlag *bool  `form:"enableFlag"`
}

// DataSetOps 数据集单条操作（详情/删除）
type DataSetOps struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// SaveDataSetReq 主子表合并保存请求（Create/Update 共用；Update 时 ID 必填）
type SaveDataSetReq struct {
	ID          uint                            `json:"ID"`
	SetCode     string                          `json:"setCode" binding:"required"`
	SetName     string                          `json:"setName" binding:"required"`
	SetDesc     string                          `json:"setDesc"`
	SourceCode  string                          `json:"sourceCode"`
	SetType     string                          `json:"setType" binding:"required,oneof=sql http"`
	DynSentence string                          `json:"dynSentence" binding:"required"`
	EnableFlag  *bool                           `json:"enableFlag"`
	Params      []report.ReportDataSetParam     `json:"params"`
	Transforms  []report.ReportDataSetTransform `json:"transforms"`
}

// TestPreviewReq 测试预览（两场景）：
// setCode 非空 = 已保存数据集测试（库中配置 + params + transforms 优先）；
// 否则用请求携带的配置即时测试（setType/dynSentence 必填）
type TestPreviewReq struct {
	SetCode     string                 `json:"setCode"`
	SetType     string                 `json:"setType"`
	SourceCode  string                 `json:"sourceCode"`
	DynSentence string                 `json:"dynSentence"`
	ParamValues map[string]interface{} `json:"paramValues"`
	PageNo      int                    `json:"pageNo"`   // 默认 1
	PageSize    int                    `json:"pageSize"` // 默认 20
}
