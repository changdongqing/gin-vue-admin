package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

// SearchModelProject 本体项目分页查询
type SearchModelProject struct {
	request.PageInfo
	Status string `form:"status"` // draft/active/archived，空=全部
}

// ModelProjectOps 单条操作（删除/详情）
type ModelProjectOps struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// CheckModelProjectCodeOps 项目编码查重
type CheckModelProjectCodeOps struct {
	ProjectCode string `form:"projectCode" binding:"required"`
	ExcludeId   uint   `form:"excludeId"` //nolint:stylecheck // 对齐前端 excludeId
}

// PrefixOps 前缀操作（列表定位/删除）
type PrefixOps struct {
	ID        uint `json:"ID" form:"ID"`
	ProjectId uint `json:"projectId" form:"projectId" binding:"required"` //nolint:stylecheck // 对齐前端 projectId
}
