package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

// SearchDataSource 数据源分页查询
type SearchDataSource struct {
	request.PageInfo
	SourceType string `form:"sourceType"` // 空=全部
	EnableFlag *bool  `form:"enableFlag"` // 指针区分「未传」与「false」
}

// DataSourceOps 数据源单条操作（详情/删除）
type DataSourceOps struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// TestConnectionReq 测试连接（前端当前表单值，不先保存；sourceConfig 为明文配置 JSON）
type TestConnectionReq struct {
	SourceType   string `json:"sourceType" binding:"required"`
	SourceConfig string `json:"sourceConfig"`
}
