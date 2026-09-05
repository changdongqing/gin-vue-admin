package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DataScopeService 数据权限引擎
//
// 使用方式（业务查询接入，一行生效）：
//
//	db = db.Scopes(DataScopeServiceApp.Scope(userId, "dept_id", "create_user_id"))
//
// 语义：解析当前用户（多角色取并集）的可见机构范围，展开为部门ID集合后追加
// (deptCol IN (...) OR userCol = userId)；任一角色为"全部数据"时不追加条件。
// 解析失败按失败关闭处理（1 <> 1），绝不放大权限。
type DataScopeService struct{}

var DataScopeServiceApp = new(DataScopeService)

// UserScope 用户数据权限解析结果
type UserScope struct {
	All     bool   // 全部数据（无需过滤）
	UserId  uint   // 当前用户ID
	DeptIds []uint // 机构范围展开后的部门ID集合（含公司展开、部门子树展开）
}

// HasFilter 是否需要追加过滤条件
func (u *UserScope) HasFilter() bool { return !u.All }

// ResolveUserScope 解析用户数据权限（多角色并集，就宽原则）
func (s *DataScopeService) ResolveUserScope(userId uint) (*UserScope, error) {
	// 总开关关闭直接放行
	if !global.GVA_CONFIG.System.DataPermission.Enable {
		return &UserScope{All: true, UserId: userId}, nil
	}

	// 1. 用户主属部门
	var user system.SysUser
	if err := global.GVA_DB.Select("id, department_id").Where("id = ?", userId).First(&user).Error; err != nil {
		return nil, err
	}

	// 2. 用户全部角色的数据范围档位
	var roleRecords []system.SysUserAuthority
	if err := global.GVA_DB.Where("sys_user_id = ?", userId).Find(&roleRecords).Error; err != nil {
		return nil, err
	}
	authorityIds := make([]uint, 0, len(roleRecords))
	for _, r := range roleRecords {
		authorityIds = append(authorityIds, r.SysAuthorityAuthorityId)
	}
	if len(authorityIds) == 0 {
		// 无任何角色：按仅本人处理
		return &UserScope{All: false, UserId: userId, DeptIds: []uint{}}, nil
	}
	var authorities []system.SysAuthority
	if err := global.GVA_DB.Select("authority_id, data_scope").Where("authority_id IN ?", authorityIds).Find(&authorities).Error; err != nil {
		return nil, err
	}

	// 3. 多角色并集合并
	// 说明：机构类范围（自定义/本公司/本部门及以下/本部门）在用户无主属部门时
	// 展开为空集合 → 查询退化为仅本人（宁紧勿松）
	scope := &UserScope{All: false, UserId: userId, DeptIds: []uint{}}
	deptIdSet := make(map[uint]struct{})
	customAuthorityIds := make([]uint, 0)

	for _, a := range authorities {
		switch a.DataScope {
		case DataScopeAll:
			scope.All = true
			return scope, nil
		case DataScopeCustom:
			customAuthorityIds = append(customAuthorityIds, a.AuthorityId)
		case DataScopeCompany:
			if user.DepartmentId > 0 {
				ids, err := DepartmentServiceApp.GetDepartmentIdsByCompany(s.userCompanyId(user.DepartmentId))
				if err != nil {
					return nil, err
				}
				for _, id := range ids {
					deptIdSet[id] = struct{}{}
				}
			}
		case DataScopeDeptChildren:
			if user.DepartmentId > 0 {
				ids, err := DepartmentServiceApp.GetDepartmentSubTreeIds(user.DepartmentId)
				if err != nil {
					return nil, err
				}
				for _, id := range ids {
					deptIdSet[id] = struct{}{}
				}
			}
		case DataScopeDept:
			if user.DepartmentId > 0 {
				deptIdSet[user.DepartmentId] = struct{}{}
			}
		case DataScopeSelf:
			// 不贡献部门集合，本人条件由 userCol 兜底
		default:
			// 未知档位按全部处理（与存量默认值语义一致）
			return &UserScope{All: true, UserId: userId}, nil
		}
	}

	// 4. 自定义明细展开（公司→全部部门；部门→子树）
	if len(customAuthorityIds) > 0 {
		var details []system.SysAuthorityDataScope
		if err := global.GVA_DB.Where("authority_id IN ?", customAuthorityIds).Find(&details).Error; err != nil {
			return nil, err
		}
		for _, d := range details {
			var ids []uint
			var err error
			switch d.ScopeType {
			case system.ScopeTypeCompany:
				ids, err = DepartmentServiceApp.GetDepartmentIdsByCompany(d.TargetId)
			case system.ScopeTypeDepartment:
				ids, err = DepartmentServiceApp.GetDepartmentSubTreeIds(d.TargetId)
			}
			if err != nil {
				return nil, err
			}
			for _, id := range ids {
				deptIdSet[id] = struct{}{}
			}
		}
	}

	for id := range deptIdSet {
		scope.DeptIds = append(scope.DeptIds, id)
	}
	return scope, nil
}

// userCompanyId 取部门所属公司ID（部门不存在返回0）
func (s *DataScopeService) userCompanyId(departmentId uint) uint {
	var department system.SysDepartment
	if err := global.GVA_DB.Select("id, company_id").Where("id = ?", departmentId).First(&department).Error; err != nil {
		return 0
	}
	return department.CompanyId
}

// GetUserDeptId 取用户主属部门ID（无部门返回0），业务创建数据时填充 dept_id 用
func (s *DataScopeService) GetUserDeptId(userId uint) (uint, error) {
	var user system.SysUser
	if err := global.GVA_DB.Select("id, department_id").Where("id = ?", userId).First(&user).Error; err != nil {
		return 0, err
	}
	return user.DepartmentId, nil
}

// Scope 生成可直接用于 db.Scopes(...) 的数据权限查询作用域
//
// deptCol：业务表部门归属列，如 "dept_id"
// userCol：本人判定列，如 "create_user_id"，传空串表示不启用本人兜底（严格机构隔离）
func (s *DataScopeService) Scope(userId uint, deptCol, userCol string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if !global.GVA_CONFIG.System.DataPermission.Enable {
			return db
		}
		scope, err := s.ResolveUserScope(userId)
		if err != nil {
			global.GVA_LOG.Error("数据权限解析失败，按空结果处理", zap.Uint("userId", userId), zap.Error(err))
			return db.Where("1 <> 1")
		}
		if scope.All {
			return db
		}
		hasDept := len(scope.DeptIds) > 0
		switch {
		case hasDept && userCol != "":
			return db.Where("("+deptCol+" IN ? OR "+userCol+" = ?)", scope.DeptIds, userId)
		case hasDept:
			return db.Where(deptCol+" IN ?", scope.DeptIds)
		case userCol != "":
			// 集合为空：仅本人档位、无主属部门退化、或自定义明细为空
			return db.Where(userCol+" = ?", userId)
		default:
			// 无部门维度且无本人列：失败关闭
			return db.Where("1 <> 1")
		}
	}
}

// CanAccess 判断用户是否可访问指定业务表的某一行（详情/修改/删除接口的归属校验）
func (s *DataScopeService) CanAccess(userId uint, tableName, deptCol, userCol string, rowId uint) (bool, error) {
	if !global.GVA_CONFIG.System.DataPermission.Enable {
		return true, nil
	}
	scope, err := s.ResolveUserScope(userId)
	if err != nil {
		return false, err
	}
	if scope.All {
		return true, nil
	}
	query := global.GVA_DB.Table(tableName).Where("id = ?", rowId)
	switch {
	case len(scope.DeptIds) > 0 && userCol != "":
		query = query.Where("("+deptCol+" IN ? OR "+userCol+" = ?)", scope.DeptIds, userId)
	case len(scope.DeptIds) > 0:
		query = query.Where(deptCol+" IN ?", scope.DeptIds)
	case userCol != "":
		query = query.Where(userCol+" = ?", userId)
	default:
		return false, nil
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
