package response

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
)

type SysUserResponse struct {
	User system.SysUser `json:"user"`
}

type LoginResponse struct {
	User      system.SysUser `json:"user"`
	Token     string         `json:"token"`
	ExpiresAt int64          `json:"expiresAt"`
}

// UserSimple 通用选择器用户精简行（字段白名单即安全边界：无手机号/邮箱等敏感字段）
type UserSimple struct {
	ID             uint   `json:"ID"`
	UserName       string `json:"userName"`
	NickName       string `json:"nickName"`
	DepartmentId   uint   `json:"departmentId"`
	DepartmentName string `json:"departmentName"`
	PostIds        []uint `json:"postIds"`
}
