package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// SysDepartment 机构部门表（参考 JeeSite js_sys_office，通过 CompanyId 挂靠公司）
type SysDepartment struct {
	global.GVA_MODEL
	CompanyId uint            `json:"companyId" form:"companyId" gorm:"column:company_id;index;not null;comment:挂靠公司ID;"`   //挂靠公司ID
	Name      string          `json:"name" form:"name" gorm:"column:name;size:100;not null;comment:部门名称;" binding:"required"` //部门名称
	Code      string          `json:"code" form:"code" gorm:"column:code;size:64;not null;uniqueIndex;comment:部门编码;"`        //部门编码
	ParentId  uint            `json:"parentId" form:"parentId" gorm:"column:parent_id;index;default:0;comment:父部门ID 0为根;"`  //父部门ID
	ParentIds string          `json:"parentIds" form:"parentIds" gorm:"column:parent_ids;size:500;index;comment:物化路径 0,1,3,"` //物化路径
	TreeLevel int             `json:"treeLevel" form:"treeLevel" gorm:"column:tree_level;default:0;comment:层级 根为0;"`         //层级
	Sort      int             `json:"sort" form:"sort" gorm:"column:sort;default:0;comment:排序;"`                             //排序
	Leader    string          `json:"leader" form:"leader" gorm:"column:leader;size:64;comment:负责人;"`                        //负责人
	Phone     string          `json:"phone" form:"phone" gorm:"column:phone;size:32;comment:联系电话;"`                          //联系电话
	Email     string          `json:"email" form:"email" gorm:"column:email;size:128;comment:邮箱;"`                            //邮箱
	Address   string          `json:"address" form:"address" gorm:"column:address;size:255;comment:地址;"`                     //地址
	Remarks   string          `json:"remarks" form:"remarks" gorm:"column:remarks;size:255;comment:备注;"`                      //备注
	Status    int             `json:"status" form:"status" gorm:"column:status;default:1;comment:状态 1启用 2停用;"`             //状态
	Children  []SysDepartment `json:"children" gorm:"-"`                                                                    //子部门
}

func (SysDepartment) TableName() string {
	return "sys_departments"
}
