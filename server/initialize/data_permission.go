package initialize

import (
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemService "github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"go.uber.org/zap"
)

// SeedDataPermission 数据权限体系幂等种子
// 1. 菜单：superAdmin 目录下追加「公司管理」「部门管理」
// 2. 角色菜单：888（admin）关联新菜单
// 3. API：登记公司/部门/数据范围接口
// 4. casbin：为 888 插入 p 规则并热加载
// 5. 演示组织数据：仅当公司表为空时插入（不修改任何存量用户）
func SeedDataPermission() {
	if global.GVA_DB == nil {
		return
	}
	if !global.GVA_CONFIG.System.DataPermission.Seed {
		global.GVA_LOG.Info("data-permission seed disabled, skipping")
		return
	}
	seedMenus()
	seedApis()
	if global.GVA_CONFIG.System.DataPermission.SeedDemo {
		seedDemoOrg()
	}
}

// seedMenus 追加公司/部门管理菜单并授权给 888
func seedMenus() {
	db := global.GVA_DB
	// 超级管理员目录菜单
	var superAdmin system.SysBaseMenu
	if err := db.Where("name = ? AND parent_id = 0", "superAdmin").First(&superAdmin).Error; err != nil {
		global.GVA_LOG.Warn("数据权限种子：未找到 superAdmin 目录菜单，跳过菜单注册", zap.Error(err))
		return
	}
	type menuSeed struct {
		Name, Title, Icon, Component string
		Sort                         int
	}
	seeds := []menuSeed{
		{Name: "company", Title: "公司管理", Icon: "office-building", Component: "view/superAdmin/company/company.vue", Sort: 13},
		{Name: "department", Title: "部门管理", Icon: "connection", Component: "view/superAdmin/department/department.vue", Sort: 14},
	}
	insertedMenuIds := make([]uint, 0)
	for _, s := range seeds {
		var count int64
		db.Model(&system.SysBaseMenu{}).Where("parent_id = ? AND name = ?", superAdmin.ID, s.Name).Count(&count)
		if count > 0 {
			continue
		}
		menu := system.SysBaseMenu{
			ParentId:  superAdmin.ID,
			Path:      s.Name,
			Name:      s.Name,
			Hidden:    false,
			Component: s.Component,
			Sort:      s.Sort,
			Meta: system.Meta{
				Title:     s.Title,
				Icon:      s.Icon,
				KeepAlive: false,
			},
		}
		if err := db.Create(&menu).Error; err != nil {
			global.GVA_LOG.Error("数据权限种子：菜单创建失败", zap.String("name", s.Name), zap.Error(err))
			continue
		}
		insertedMenuIds = append(insertedMenuIds, menu.ID)
	}
	// 全部菜单（含此前已存在但未授权的）授权给 888
	var menus []system.SysBaseMenu
	db.Where("parent_id = ? AND name IN ?", superAdmin.ID, []string{"company", "department"}).Find(&menus)
	for _, m := range menus {
		var count int64
		db.Table("sys_authority_menus").
			Where("sys_base_menu_id = ? AND sys_authority_authority_id = ?", m.ID, 888).
			Count(&count)
		if count > 0 {
			continue
		}
		if err := db.Table("sys_authority_menus").
			Create(map[string]interface{}{
				"sys_base_menu_id":            m.ID,
				"sys_authority_authority_id": 888,
			}).Error; err != nil {
			global.GVA_LOG.Error("数据权限种子：角色菜单授权失败", zap.Uint("menuId", m.ID), zap.Error(err))
		}
	}
	_ = insertedMenuIds // 保留供日志扩展使用
}

// seedApis 登记接口并配置 casbin 规则
func seedApis() {
	db := global.GVA_DB
	type apiSeed struct {
		Path, Method, Group, Description string
	}
	seeds := []apiSeed{
		{"/company/createCompany", "POST", "公司管理", "创建公司"},
		{"/company/updateCompany", "PUT", "公司管理", "更新公司"},
		{"/company/deleteCompany", "DELETE", "公司管理", "删除公司"},
		{"/company/findCompany", "GET", "公司管理", "查询公司"},
		{"/company/getCompanyList", "GET", "公司管理", "获取公司树"},
		{"/department/createDepartment", "POST", "部门管理", "创建部门"},
		{"/department/updateDepartment", "PUT", "部门管理", "更新部门"},
		{"/department/deleteDepartment", "DELETE", "部门管理", "删除部门"},
		{"/department/findDepartment", "GET", "部门管理", "查询部门"},
		{"/department/getDepartmentList", "GET", "部门管理", "获取部门树"},
		{"/authority/setDataScope", "POST", "角色", "设置角色数据范围"},
		{"/authority/getDataScope", "GET", "角色", "获取角色数据范围"},
	}
	addedRules := make([][]string, 0)
	for _, s := range seeds {
		var count int64
		db.Model(&system.SysApi{}).Where("path = ? AND method = ?", s.Path, s.Method).Count(&count)
		if count == 0 {
			if err := db.Create(&system.SysApi{Path: s.Path, Method: s.Method, ApiGroup: s.Group, Description: s.Description}).Error; err != nil {
				global.GVA_LOG.Error("数据权限种子：API登记失败", zap.String("path", s.Path), zap.Error(err))
				continue
			}
		}
		// casbin 规则（唯一索引兜底，插入冲突忽略）
		var ruleCount int64
		db.Table("casbin_rule").
			Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", "888", s.Path, s.Method).
			Count(&ruleCount)
		if ruleCount == 0 {
			if err := db.Table("casbin_rule").Create(map[string]interface{}{
				"ptype": "p", "v0": "888", "v1": s.Path, "v2": s.Method,
			}).Error; err != nil {
				global.GVA_LOG.Error("数据权限种子：casbin规则插入失败", zap.String("path", s.Path), zap.Error(err))
				continue
			}
			addedRules = append(addedRules, []string{"888", s.Path, s.Method})
		}
	}
	if len(addedRules) > 0 {
		if err := systemService.CasbinServiceApp.FreshCasbin(); err != nil {
			global.GVA_LOG.Error("数据权限种子：casbin热加载失败（重启后生效）", zap.Error(err))
		} else {
			global.GVA_LOG.Info(fmt.Sprintf("数据权限种子：已为角色888新增 %d 条api规则", len(addedRules)))
		}
	}
}

// seedDemoOrg 公司表为空时插入演示组织数据
func seedDemoOrg() {
	var count int64
	global.GVA_DB.Model(&system.SysCompany{}).Count(&count)
	if count > 0 {
		return
	}
	companyService := systemService.CompanyServiceApp
	departmentService := systemService.DepartmentServiceApp

	type cSeed struct{ Name, Code string }
	companies := []cSeed{
		{Name: "GVA集团", Code: "GVA-GROUP"},
		{Name: "广州分公司", Code: "GVA-GZ"},
		{Name: "深圳分公司", Code: "GVA-SZ"},
	}
	companyIds := make([]uint, len(companies))
	for i, cs := range companies {
		var parentId uint
		if i > 0 {
			parentId = companyIds[0]
		}
		company := system.SysCompany{Name: cs.Name, Code: cs.Code, ParentId: parentId, Status: 1, Sort: i + 1}
		if err := companyService.CreateCompany(&company); err != nil {
			global.GVA_LOG.Error("数据权限种子：演示公司创建失败", zap.String("code", cs.Code), zap.Error(err))
			return
		}
		companyIds[i] = company.ID
	}
	type dSeed struct{ CompanyIdx int; Name, Code string; ParentCode string }
	departments := []dSeed{
		{0, "集团总部", "GVA-GROUP-HQ", ""},
		{0, "财务部", "GVA-GROUP-FIN", ""},
		{1, "总经办", "GVA-GZ-CEO", ""},
		{1, "技术部", "GVA-GZ-TECH", ""},
		{1, "人事部", "GVA-GZ-HR", ""},
		{1, "研发组", "GVA-GZ-TECH-RD", "GVA-GZ-TECH"},
		{1, "测试组", "GVA-GZ-TECH-QA", "GVA-GZ-TECH"},
		{2, "总经办", "GVA-SZ-CEO", ""},
		{2, "市场部", "GVA-SZ-MKT", ""},
	}
	codeToId := make(map[string]uint)
	for _, ds := range departments {
		var parentId uint
		if ds.ParentCode != "" {
			parentId = codeToId[ds.ParentCode]
		}
		department := system.SysDepartment{
			CompanyId: companyIds[ds.CompanyIdx],
			Name:      ds.Name,
			Code:      ds.Code,
			ParentId:  parentId,
			Status:    1,
			Sort:      len(codeToId) + 1,
		}
		if err := departmentService.CreateDepartment(&department); err != nil {
			global.GVA_LOG.Error("数据权限种子：演示部门创建失败", zap.String("code", ds.Code), zap.Error(err))
			return
		}
		codeToId[ds.Code] = department.ID
	}
	global.GVA_LOG.Info("数据权限种子：已插入演示组织数据（1集团+2公司+9部门），可在公司/部门管理页查看")
}
