package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

// SearchAnnotationProperty 分页查询（含创建时间范围）
type SearchAnnotationProperty struct {
	request.PageInfo
	AppliesTo string `form:"appliesTo"`
	StartTime string `form:"startTime"` // 创建时间起（yyyy-MM-dd HH:mm:ss，前端 RangePicker 映射）
	EndTime   string `form:"endTime"`   // 创建时间止
}

// AnnotationPropertyOps 单条操作（删除/详情）
type AnnotationPropertyOps struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}
