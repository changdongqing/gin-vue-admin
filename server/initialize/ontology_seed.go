package initialize

import (
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemService "github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"go.uber.org/zap"
)

// SeedOntology 本体治理幂等种子（菜单/API/casbin；字典随迁移 SQL 种入，不做 Go 双写）
// 1. 菜单：顶级目录「本体治理」+ 子菜单「属性模板库」（02/03/04 陆续追加子菜单）
// 2. 角色菜单：888（admin）关联新菜单
// 3. API：登记治理 CRUD 与供给接口（供给独立「本体供给」组）
// 4. casbin：为 888 插入 p 规则；供给规则同时授予「建模师」角色（若存在）并热加载
func SeedOntology() {
	if global.GVA_DB == nil {
		return
	}
	seedOntologyMenus()
	seedOntologyApis()
}

// seedOntologyMenus 追加本体治理目录与子菜单并授权给 888
func seedOntologyMenus() {
	db := global.GVA_DB
	// 顶级目录（幂等）
	var dir system.SysBaseMenu
	if err := db.Where("name = ? AND parent_id = 0", "ontology").First(&dir).Error; err != nil {
		dir = system.SysBaseMenu{
			ParentId:  0,
			Path:      "ontology",
			Name:      "ontology",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      20,
			Meta: system.Meta{
				Title:     "本体治理",
				Icon:      "share",
				KeepAlive: false,
			},
		}
		if err := db.Create(&dir).Error; err != nil {
			global.GVA_LOG.Error("本体种子：目录菜单创建失败", zap.Error(err))
			return
		}
	}
	type menuSeed struct {
		Name, Title, Icon, Component string
		Sort                         int
	}
	seeds := []menuSeed{
		{Name: "propertyTemplate", Title: "属性模板库", Icon: "document", Component: "view/ontology/propertyTemplate/propertyTemplate.vue", Sort: 1},
		{Name: "classTemplate", Title: "分类模板", Icon: "grid", Component: "view/ontology/classTemplate/classTemplate.vue", Sort: 2},
		{Name: "unit", Title: "单位注册表", Icon: "scale-to-original", Component: "view/ontology/unit/unit.vue", Sort: 3},
		{Name: "annotationProperty", Title: "注释属性注册表", Icon: "collection", Component: "view/ontology/annotationProperty/annotationProperty.vue", Sort: 4},
	}
	for _, s := range seeds {
		var count int64
		db.Model(&system.SysBaseMenu{}).Where("parent_id = ? AND name = ?", dir.ID, s.Name).Count(&count)
		if count > 0 {
			continue
		}
		menu := system.SysBaseMenu{
			ParentId:  dir.ID,
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
			global.GVA_LOG.Error("本体种子：菜单创建失败", zap.String("name", s.Name), zap.Error(err))
			continue
		}
	}
	// 目录 + 全部子菜单授权给 888
	var menus []system.SysBaseMenu
	db.Where("name = ? OR parent_id = ?", "ontology", dir.ID).Find(&menus)
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
				"sys_base_menu_id":           m.ID,
				"sys_authority_authority_id": 888,
			}).Error; err != nil {
			global.GVA_LOG.Error("本体种子：角色菜单授权失败", zap.Uint("menuId", m.ID), zap.Error(err))
		}
	}
}

// seedOntologyApis 登记本体接口并配置 casbin 规则
func seedOntologyApis() {
	db := global.GVA_DB
	type apiSeed struct {
		Path, Method, Group, Description string
	}
	seeds := []apiSeed{
		{"/ontology/propertyTemplate/createPropertyTemplate", "POST", "本体治理", "创建属性模板"},
		{"/ontology/propertyTemplate/updatePropertyTemplate", "PUT", "本体治理", "更新属性模板"},
		{"/ontology/propertyTemplate/deletePropertyTemplate", "DELETE", "本体治理", "删除属性模板"},
		{"/ontology/propertyTemplate/disablePropertyTemplate", "PUT", "本体治理", "弃用/取消弃用属性模板"},
		{"/ontology/propertyTemplate/findPropertyTemplate", "GET", "本体治理", "查询属性模板详情"},
		{"/ontology/propertyTemplate/getPropertyTemplateList", "GET", "本体治理", "分页查询属性模板"},
		{"/ontology/propertyTemplate/getPropertyTemplateAll", "GET", "本体治理", "全量属性模板下拉"},
		{"/ontology/propertyTemplate/checkPropertyTemplateCode", "GET", "本体治理", "属性模板编码查重"},
		{"/ontology/propertyTemplate/promotePropertyTemplate", "POST", "本体治理", "属性快照提升为模板"},
		{"/ontology/classTemplate/createClassTemplate", "POST", "本体治理", "创建分类模板"},
		{"/ontology/classTemplate/updateClassTemplate", "PUT", "本体治理", "更新分类模板"},
		{"/ontology/classTemplate/deleteClassTemplate", "DELETE", "本体治理", "删除分类模板"},
		{"/ontology/classTemplate/disableClassTemplate", "PUT", "本体治理", "弃用/取消弃用分类模板"},
		{"/ontology/classTemplate/findClassTemplate", "GET", "本体治理", "查询分类模板详情"},
		{"/ontology/classTemplate/getClassTemplateList", "GET", "本体治理", "扁平全量分类模板"},
		{"/ontology/classTemplate/getClassTemplatePage", "GET", "本体治理", "分页查询分类模板"},
		{"/ontology/classTemplate/getClassTemplateTreeRoots", "GET", "本体治理", "分类树下拉"},
		{"/ontology/classTemplate/getClassTemplateRefList", "GET", "本体治理", "骨架引用列表"},
		{"/ontology/classTemplate/getClassTemplateInherited", "GET", "本体治理", "分类模板继承视图"},
		{"/ontology/classTemplate/previewClassificationCode", "GET", "本体治理", "分类编码预览"},
		{"/ontology/classificationRule/findClassificationRule", "GET", "本体治理", "查询分类编码规则"},
		{"/ontology/classificationRule/saveClassificationRule", "PUT", "本体治理", "保存分类编码规则"},
		{"/ontology/unit/createUnit", "POST", "本体治理", "创建单位"},
		{"/ontology/unit/updateUnit", "PUT", "本体治理", "更新单位"},
		{"/ontology/unit/deleteUnit", "DELETE", "本体治理", "删除单位"},
		{"/ontology/unit/disableUnit", "PUT", "本体治理", "停用/启用单位"},
		{"/ontology/unit/findUnit", "GET", "本体治理", "查询单位详情"},
		{"/ontology/unit/getUnitPage", "GET", "本体治理", "分页查询单位"},
		{"/ontology/unit/getUnitAll", "GET", "本体治理", "全量单位下拉"},
		{"/ontology/unit/convertUnit", "GET", "本体治理", "单位换算试算"},
		{"/ontology/quantityKind/getQuantityKindList", "GET", "本体治理", "量纲列表"},
		{"/ontology/annotationProperty/createAnnotationProperty", "POST", "本体治理", "创建注释属性"},
		{"/ontology/annotationProperty/updateAnnotationProperty", "PUT", "本体治理", "更新注释属性"},
		{"/ontology/annotationProperty/deleteAnnotationProperty", "DELETE", "本体治理", "删除注释属性"},
		{"/ontology/annotationProperty/findAnnotationProperty", "GET", "本体治理", "查询注释属性详情"},
		{"/ontology/annotationProperty/getAnnotationPropertyList", "GET", "本体治理", "分页查询注释属性"},
		{"/ontology/annotationProperty/getAnnotationPropertyAll", "GET", "本体治理", "全量注释属性"},
		{"/ontology/annotationProperty/exportAnnotationPropertyExcel", "GET", "本体治理", "导出注释属性清单"},
		{"/ontology/supply/v1/propertyTemplates", "GET", "本体供给", "供给属性模板列表"},
		{"/ontology/supply/v1/classTemplate/tree", "GET", "本体供给", "供给分类模板树"},
		{"/ontology/supply/v1/classHierarchy/suggest", "GET", "本体供给", "类层级建议"},
		{"/ontology/supply/v1/units", "GET", "本体供给", "供给单位清单"},
		{"/ontology/supply/v1/units/convert", "GET", "本体供给", "供给单位换算"},
		{"/ontology/supply/v1/annotationProperties", "GET", "本体供给", "供给注释属性注册表"},
	}
	// 治理 CRUD 授 888；供给接口额外授予「建模师」角色（若存在）
	governanceRoles := []string{"888"}
	supplyRoles := append([]string{"888"}, findModelerAuthorityIds()...)

	addedRules := make([][]string, 0)
	for _, s := range seeds {
		var count int64
		db.Model(&system.SysApi{}).Where("path = ? AND method = ?", s.Path, s.Method).Count(&count)
		if count == 0 {
			if err := db.Create(&system.SysApi{Path: s.Path, Method: s.Method, ApiGroup: s.Group, Description: s.Description}).Error; err != nil {
				global.GVA_LOG.Error("本体种子：API登记失败", zap.String("path", s.Path), zap.Error(err))
				continue
			}
		}
		roles := governanceRoles
		if s.Group == "本体供给" {
			roles = supplyRoles
		}
		for _, role := range roles {
			var ruleCount int64
			db.Table("casbin_rule").
				Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", role, s.Path, s.Method).
				Count(&ruleCount)
			if ruleCount > 0 {
				continue
			}
			if err := db.Table("casbin_rule").Create(map[string]interface{}{
				"ptype": "p", "v0": role, "v1": s.Path, "v2": s.Method,
			}).Error; err != nil {
				global.GVA_LOG.Error("本体种子：casbin规则插入失败", zap.String("path", s.Path), zap.String("role", role), zap.Error(err))
				continue
			}
			addedRules = append(addedRules, []string{role, s.Path, s.Method})
		}
	}
	if len(addedRules) > 0 {
		if err := systemService.CasbinServiceApp.FreshCasbin(); err != nil {
			global.GVA_LOG.Error("本体种子：casbin热加载失败（重启后生效）", zap.Error(err))
		} else {
			global.GVA_LOG.Info(fmt.Sprintf("本体种子：已新增 %d 条api规则", len(addedRules)))
		}
	}
}

// findModelerAuthorityIds 查找「建模师」角色ID（未配置则返回空，仅授 888）
func findModelerAuthorityIds() []string {
	var ids []string
	global.GVA_DB.Table("sys_authorities").
		Where("authority_name = ?", "建模师").
		Pluck("authority_id", &ids)
	return ids
}
