package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

// SearchUnit 单位分页查询
type SearchUnit struct {
	request.PageInfo
	QuantityKindCode string `form:"quantityKindCode"`
	Source           string `form:"source"`
	Status           *int   `form:"status"`
}

// UnitOps 单条操作（删除/详情/停用）
type UnitOps struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// ConvertUnitReq 换算请求（GET query；value 用字符串承载，避免 float 精度损失）
type ConvertUnitReq struct {
	Value   string `form:"value" binding:"required"`
	FromIri string `form:"fromIri" binding:"required"`
	ToIri   string `form:"toIri" binding:"required"`
}
