package request

import (
	common "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

// SysPostSearch 岗位分页查询
type SysPostSearch struct {
	common.PageInfo
	PostType string `json:"postType" form:"postType"` //岗位类型，空为全部
	Status   int    `json:"status" form:"status"`     //状态 1启用 2停用 0为全部
}

// SetPostUsersReq 岗位绑定员工（全量覆盖）
type SetPostUsersReq struct {
	PostId  uint   `json:"postId" binding:"required"`
	UserIds []uint `json:"userIds"`
}

// SetUserPostsReq 员工绑定岗位（全量覆盖）
type SetUserPostsReq struct {
	UserId  uint   `json:"userId" binding:"required"`
	PostIds []uint `json:"postIds"`
}
