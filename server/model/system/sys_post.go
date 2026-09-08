package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// SysPost 岗位表（参考 JeeSite js_sys_post；独立模块，不涉数据权限，岗位全局共享不挂部门）
type SysPost struct {
	global.GVA_MODEL
	PostCode string `json:"postCode" form:"postCode" gorm:"column:post_code;size:64;not null;uniqueIndex;comment:岗位编码;" binding:"required"`   //岗位编码
	PostName string `json:"postName" form:"postName" gorm:"column:post_name;size:64;not null;comment:岗位名称;" binding:"required"`               //岗位名称
	PostType string `json:"postType" form:"postType" gorm:"column:post_type;size:32;default:'';comment:岗位类型 字典post_type;" binding:"required"` //岗位类型
	Sort     int    `json:"sort" form:"sort" gorm:"column:sort;default:0;comment:排序;"`                                                        //排序
	Status   int    `json:"status" form:"status" gorm:"column:status;default:1;comment:状态 1启用 2停用;"`                                          //状态
	Remarks  string `json:"remarks" form:"remarks" gorm:"column:remarks;size:255;comment:备注;"`                                                //备注
	// 非持久化：岗位关联员工（详情场景，业务读写一律显式操作 SysUserPost，不走关联自动行为）
	Users []SysUser `json:"users" gorm:"many2many:sys_user_post;foreignKey:ID;references:ID;joinForeignKey:UserId;joinReferences:PostId"`
}

func (SysPost) TableName() string {
	return "sys_posts"
}

// SysUserPost 员工-岗位关联表（复合主键，硬删除，无审计诉求——操作留痕由写路由 OperationRecord 覆盖）
type SysUserPost struct {
	UserId uint `json:"userId" gorm:"column:user_id;primaryKey;comment:用户ID;"` //用户ID
	PostId uint `json:"postId" gorm:"column:post_id;primaryKey;comment:岗位ID;"` //岗位ID
}

func (SysUserPost) TableName() string {
	return "sys_user_post"
}

// SysPostListItem 岗位分页列表行（携带关联员工数子查询列）
type SysPostListItem struct {
	SysPost
	UserCount int64 `json:"userCount" gorm:"column:user_count;"` //关联员工数
}

// PostUserBrief 岗位已关联员工的精简信息（分配员工弹窗/回显用，不含手机号邮箱等敏感字段）
type PostUserBrief struct {
	ID             uint   `json:"ID" gorm:"column:id;"`
	UserName       string `json:"userName" gorm:"column:user_name;"`
	NickName       string `json:"nickName" gorm:"column:nick_name;"`
	DepartmentId   uint   `json:"departmentId" gorm:"column:department_id;"`
	DepartmentName string `json:"departmentName" gorm:"column:department_name;"`
}
