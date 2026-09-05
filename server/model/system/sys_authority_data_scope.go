package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// SysAuthorityDataScope 角色自定义数据范围明细
// 角色数据范围为"自定义"(DataScopeCustom)时生效：勾选公司=整家公司全部部门，勾选部门=该部门子树
type SysAuthorityDataScope struct {
	global.GVA_MODEL
	AuthorityId uint   `json:"authorityId" gorm:"column:authority_id;index;not null;comment:角色ID;"`           //角色ID
	ScopeType   string `json:"scopeType" gorm:"column:scope_type;size:16;not null;comment:范围类型 company|department;"` //范围类型
	TargetId    uint   `json:"targetId" gorm:"column:target_id;not null;comment:公司或部门ID;"`                 //目标ID
}

func (SysAuthorityDataScope) TableName() string {
	return "sys_authority_data_scopes"
}

const (
	ScopeTypeCompany    = "company"
	ScopeTypeDepartment = "department"
)
