package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func boolPtr(b bool) *bool {
	return &b
}

// SeedPost 岗位管理幂等种子（菜单/API/casbin/字典）
// 1. 菜单：「超级管理员」目录下新增「岗位管理」（锚定部门管理菜单的父级，失败降级顶级）并授权 888
// 2. API：登记 /post/** 路由 + casbin p 规则（888）
// 3. 字典：post_type 岗位类型（迁移 000011 已种，此处幂等兜底）
func SeedPost() {
	if global.GVA_DB == nil {
		return
	}
	seedPostMenus()
	seedPostApis()
	seedPostDict()
}

// postParentMenuId 锚定「超级管理员」目录：取部门管理菜单的 parent_id，查不到则降级顶级
func postParentMenuId(db *gorm.DB) uint {
	var department system.SysBaseMenu
	if err := db.Where("name = ?", "department").First(&department).Error; err != nil {
		global.GVA_LOG.Warn("岗位种子：未找到部门管理菜单，岗位管理降级为顶级菜单")
		return 0
	}
	return department.ParentId
}

// seedPostMenus 追加岗位管理菜单并授权给 888
func seedPostMenus() {
	db := global.GVA_DB
	parentId := postParentMenuId(db)
	var count int64
	db.Model(&system.SysBaseMenu{}).Where("name = ? AND parent_id = ?", "post", parentId).Count(&count)
	if count == 0 {
		menu := system.SysBaseMenu{
			ParentId:  parentId,
			Path:      "post",
			Name:      "post",
			Hidden:    false,
			Component: "view/superAdmin/post/post.vue",
			Sort:      3,
			Meta: system.Meta{
				Title:     "岗位管理",
				Icon:      "suitcase",
				KeepAlive: false,
			},
		}
		if err := db.Create(&menu).Error; err != nil {
			global.GVA_LOG.Error("岗位种子：菜单创建失败", zap.Error(err))
			return
		}
	}
	var menu system.SysBaseMenu
	if err := db.Where("name = ? AND parent_id = ?", "post", parentId).First(&menu).Error; err != nil {
		return
	}
	var linked int64
	db.Table("sys_authority_menus").
		Where("sys_base_menu_id = ? AND sys_authority_authority_id = ?", menu.ID, 888).
		Count(&linked)
	if linked == 0 {
		if err := db.Table("sys_authority_menus").Create(map[string]interface{}{
			"sys_base_menu_id":           menu.ID,
			"sys_authority_authority_id": 888,
		}).Error; err != nil {
			global.GVA_LOG.Error("岗位种子：角色菜单授权失败", zap.Uint("menuId", menu.ID), zap.Error(err))
		}
	}
}

// seedPostApis 登记岗位管理接口并配置 casbin 规则
func seedPostApis() {
	db := global.GVA_DB
	type apiSeed struct {
		Path, Method, Group, Description string
	}
	seeds := []apiSeed{
		{"/post/createPost", "POST", "岗位管理", "创建岗位"},
		{"/post/updatePost", "PUT", "岗位管理", "更新岗位"},
		{"/post/deletePost", "DELETE", "岗位管理", "删除岗位"},
		{"/post/findPost", "GET", "岗位管理", "查询岗位详情"},
		{"/post/getPostList", "POST", "岗位管理", "分页查询岗位列表"},
		{"/post/getPostListAll", "GET", "岗位管理", "全部启用岗位下拉"},
		{"/post/getPostUsers", "GET", "岗位管理", "岗位已关联员工"},
		{"/post/setPostUsers", "POST", "岗位管理", "岗位绑定员工（全量覆盖）"},
		{"/post/getUserPosts", "GET", "岗位管理", "员工已关联岗位"},
		{"/post/setUserPosts", "POST", "岗位管理", "员工绑定岗位（全量覆盖）"},
	}
	for _, s := range seeds {
		var count int64
		db.Model(&system.SysApi{}).Where("path = ? AND method = ?", s.Path, s.Method).Count(&count)
		if count == 0 {
			if err := db.Create(&system.SysApi{Path: s.Path, Method: s.Method, ApiGroup: s.Group, Description: s.Description}).Error; err != nil {
				global.GVA_LOG.Error("岗位种子：API登记失败", zap.String("path", s.Path), zap.Error(err))
				continue
			}
		}
		var ruleCount int64
		db.Table("casbin_rule").
			Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", "888", s.Path, s.Method).
			Count(&ruleCount)
		if ruleCount > 0 {
			continue
		}
		if err := db.Table("casbin_rule").Create(map[string]interface{}{
			"ptype": "p", "v0": "888", "v1": s.Path, "v2": s.Method,
		}).Error; err != nil {
			global.GVA_LOG.Error("岗位种子：casbin规则插入失败", zap.String("path", s.Path), zap.Error(err))
		}
	}
}

// seedPostDict 岗位类型字典幂等兜底（迁移 000011 已种，双保险）
func seedPostDict() {
	db := global.GVA_DB
	var dict system.SysDictionary
	if err := db.Where("type = ?", "post_type").First(&dict).Error; err != nil {
		dict = system.SysDictionary{
			Name:   "岗位类型",
			Type:   "post_type",
			Status: boolPtr(true),
			Desc:   "岗位管理-岗位类型（参考 JeeSite sys_post_type）",
		}
		if err := db.Create(&dict).Error; err != nil {
			global.GVA_LOG.Error("岗位种子：字典类型创建失败", zap.Error(err))
			return
		}
	}
	details := []struct {
		Label string
		Value string
		Sort  int
	}{
		{"高层岗位", "high", 1},
		{"中层岗位", "middle", 2},
		{"基层岗位", "primary", 3},
		{"普通岗位", "common", 4},
	}
	for _, d := range details {
		var count int64
		db.Model(&system.SysDictionaryDetail{}).
			Where("sys_dictionary_id = ? AND value = ?", dict.ID, d.Value).
			Count(&count)
		if count > 0 {
			continue
		}
		if err := db.Create(&system.SysDictionaryDetail{
			Label:           d.Label,
			Value:           d.Value,
			Sort:            d.Sort,
			Status:          boolPtr(true),
			SysDictionaryID: int(dict.ID),
		}).Error; err != nil {
			global.GVA_LOG.Error("岗位种子：字典明细创建失败", zap.String("value", d.Value), zap.Error(err))
		}
	}
}
