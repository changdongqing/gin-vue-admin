package initialize

import (
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemService "github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"go.uber.org/zap"
)

// SeedOntology 本体治理幂等种子（菜单/API/casbin；字典随迁移 SQL 种入，不做 Go 双写）
// 1. 菜单：顶级目录「本体」+ 子目录「本体治理」+ 四个治理叶子（旧结构叶子直挂顶级目录时自动迁移）
// 2. 角色菜单：888（admin）关联新菜单
// 3. API：登记治理 CRUD 与供给接口（供给独立「本体供给」组）
// 4. casbin：为 888 插入 p 规则；供给规则同时授予「建模师」角色（若存在）并热加载
func SeedOntology() {
	if global.GVA_DB == nil {
		return
	}
	seedOntologyMenus()
	seedOntologyModelMenus()
	seedOntologyApis()
}

// seedOntologyMenus 追加本体治理子目录与叶子菜单并授权给 888
// 结构：本体(ontology) → 本体治理(governance) → 四个治理叶子；
// 旧结构（治理叶子直挂顶级目录）自动迁移到子目录，保留菜单 id 与既有角色授权
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
				Title:     "本体",
				Icon:      "share",
				KeepAlive: false,
			},
		}
		if err := db.Create(&dir).Error; err != nil {
			global.GVA_LOG.Error("本体种子：目录菜单创建失败", zap.Error(err))
			return
		}
	}
	// 子目录 governance（幂等；运维手建的「本体治理」目录按标题识别并归一化 name/path/sort/icon，避免重复建目录）
	var govDir system.SysBaseMenu
	if err := db.Where("name = ? AND parent_id = ?", "governance", dir.ID).First(&govDir).Error; err != nil {
		if err := db.Where("title = ? AND parent_id = ?", "本体治理", dir.ID).First(&govDir).Error; err != nil {
			govDir = system.SysBaseMenu{
				ParentId:  dir.ID,
				Path:      "governance",
				Name:      "governance",
				Hidden:    false,
				Component: "view/routerHolder.vue",
				Sort:      80,
				Meta: system.Meta{
					Title:     "本体治理",
					Icon:      "operation",
					KeepAlive: false,
				},
			}
			if err := db.Create(&govDir).Error; err != nil {
				global.GVA_LOG.Error("本体种子：治理子目录创建失败", zap.Error(err))
				return
			}
		} else if err := db.Model(&govDir).Updates(map[string]interface{}{
			"name": "governance", "path": "governance", "sort": 80, "icon": "operation",
		}).Error; err != nil {
			global.GVA_LOG.Error("本体种子：治理子目录归一化失败", zap.Error(err))
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
		var menu system.SysBaseMenu
		// 已挂在本体治理子目录下（幂等跳过）
		if err := db.Where("parent_id = ? AND name = ?", govDir.ID, s.Name).First(&menu).Error; err == nil {
			continue
		}
		// 旧结构：直挂顶级目录 → 迁移进子目录（保留 id 与既有授权，仅改归属）
		if err := db.Where("parent_id = ? AND name = ?", dir.ID, s.Name).First(&menu).Error; err == nil {
			if err := db.Model(&menu).Update("parent_id", govDir.ID).Error; err != nil {
				global.GVA_LOG.Error("本体种子：治理菜单迁移失败", zap.String("name", s.Name), zap.Error(err))
			}
			continue
		}
		menu = system.SysBaseMenu{
			ParentId:  govDir.ID,
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
	// 本体目录 + 治理子目录 + 全部治理叶子授权给 888（建模域由 seedOntologyModelMenus 授权）
	var menus []system.SysBaseMenu
	db.Where("id = ? OR id = ? OR parent_id = ?", dir.ID, govDir.ID, govDir.ID).Find(&menus)
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

// seedOntologyModelMenus 本体建模域菜单（01/02/03 陆续追加）：
// 1. 顶级目录 ontology 标题「本体治理」→「本体」（该目录同时容纳治理与建模两组子菜单，幂等 UPDATE 仅改标题）
// 2. 子目录 model（本体建模，routerHolder）
// 3. 叶子菜单 modelProject（path=project，Path≠Name，最终 URL /ontology/model/project）
func seedOntologyModelMenus() {
	db := global.GVA_DB
	var dir system.SysBaseMenu
	if err := db.Where("name = ? AND parent_id = 0", "ontology").First(&dir).Error; err != nil {
		global.GVA_LOG.Warn("本体建模种子：顶级目录不存在（先启动治理域种子）")
		return
	}
	// 目录更名：仅当仍为旧标题时更新（运维自定义过标题则不动）
	if dir.Title == "本体治理" {
		if err := db.Model(&system.SysBaseMenu{}).Where("id = ?", dir.ID).Update("title", "本体").Error; err != nil {
			global.GVA_LOG.Error("本体建模种子：目录标题更新失败", zap.Error(err))
		}
	}
	// 子目录 model（幂等）
	var modelDir system.SysBaseMenu
	if err := db.Where("name = ? AND parent_id = ?", "model", dir.ID).First(&modelDir).Error; err != nil {
		modelDir = system.SysBaseMenu{
			ParentId:  dir.ID,
			Path:      "model",
			Name:      "model",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      90,
			Meta: system.Meta{
				Title:     "本体建模",
				Icon:      "edit-outline",
				KeepAlive: false,
			},
		}
		if err := db.Create(&modelDir).Error; err != nil {
			global.GVA_LOG.Error("本体建模种子：子目录创建失败", zap.Error(err))
			return
		}
	}
	// 叶子菜单（显式 Path；现有治理范式 Path=Name，此处 Path≠Name 必须单独处理）
	type leafSeed struct {
		Name, Path, Title, Icon, Component string
		Sort                               int
	}
	leaves := []leafSeed{
		{Name: "modelProject", Path: "project", Title: "本体项目管理", Icon: "folder-opened", Component: "view/ontology/model/project/project.vue", Sort: 1},
		{Name: "modelClass", Path: "class", Title: "本体类建模", Icon: "connection", Component: "view/ontology/model/class/class.vue", Sort: 2},
		{Name: "extBinding", Path: "extbinding", Title: "外部模块关联", Icon: "link", Component: "view/ontology/model/extbinding/extbinding.vue", Sort: 3},
	}
	for _, s := range leaves {
		var count int64
		db.Model(&system.SysBaseMenu{}).Where("parent_id = ? AND name = ?", modelDir.ID, s.Name).Count(&count)
		if count > 0 {
			continue
		}
		menu := system.SysBaseMenu{
			ParentId:  modelDir.ID,
			Path:      s.Path,
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
			global.GVA_LOG.Error("本体建模种子：菜单创建失败", zap.String("name", s.Name), zap.Error(err))
			continue
		}
	}
	// ontology 目录 + model 子目录 + 全部叶子授权给 888
	var menus []system.SysBaseMenu
	db.Where("id = ? OR parent_id = ? OR parent_id = ?", dir.ID, dir.ID, modelDir.ID).Find(&menus)
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
			global.GVA_LOG.Error("本体建模种子：角色菜单授权失败", zap.Uint("menuId", m.ID), zap.Error(err))
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
		{"/ontology/modelProject/createModelProject", "POST", "本体建模", "创建本体项目"},
		{"/ontology/modelProject/updateModelProject", "PUT", "本体建模", "更新本体项目"},
		{"/ontology/modelProject/deleteModelProject", "DELETE", "本体建模", "删除本体项目"},
		{"/ontology/modelProject/findModelProject", "GET", "本体建模", "查询本体项目详情"},
		{"/ontology/modelProject/getModelProjectList", "GET", "本体建模", "分页查询本体项目"},
		{"/ontology/modelProject/getModelProjectAll", "GET", "本体建模", "全量本体项目下拉"},
		{"/ontology/modelProject/checkModelProjectCode", "GET", "本体建模", "本体项目编码查重"},
		{"/ontology/modelPrefix/getModelPrefixList", "GET", "本体建模", "项目IRI前缀列表"},
		{"/ontology/modelPrefix/createModelPrefix", "POST", "本体建模", "新增项目IRI前缀"},
		{"/ontology/modelPrefix/updateModelPrefix", "PUT", "本体建模", "更新项目IRI前缀"},
		{"/ontology/modelPrefix/deleteModelPrefix", "DELETE", "本体建模", "删除项目IRI前缀"},
		{"/ontology/modelClass/createModelClass", "POST", "本体建模", "创建本体类"},
		{"/ontology/modelClass/updateModelClass", "PUT", "本体建模", "更新本体类"},
		{"/ontology/modelClass/deleteModelClass", "DELETE", "本体建模", "删除本体类"},
		{"/ontology/modelClass/findModelClass", "GET", "本体建模", "查询本体类详情"},
		{"/ontology/modelClass/getModelClassDetail", "GET", "本体建模", "类详情聚合"},
		{"/ontology/modelClass/getModelClassList", "GET", "本体建模", "分页查询本体类"},
		{"/ontology/modelClass/getModelClassByProject", "GET", "本体建模", "项目内全量类"},
		{"/ontology/modelClass/checkModelClassLocalName", "GET", "本体建模", "类本地名查重"},
		{"/ontology/modelClass/instantiateModelClass", "POST", "本体建模", "分类模板实例化"},
		{"/ontology/modelClass/previewInstantiateModelClass", "GET", "本体建模", "实例化预览"},
		{"/ontology/modelDatatypeProperty/createModelDatatypeProperty", "POST", "本体建模", "创建数据属性"},
		{"/ontology/modelDatatypeProperty/updateModelDatatypeProperty", "PUT", "本体建模", "更新数据属性"},
		{"/ontology/modelDatatypeProperty/deleteModelDatatypeProperty", "DELETE", "本体建模", "删除数据属性"},
		{"/ontology/modelDatatypeProperty/findModelDatatypeProperty", "GET", "本体建模", "查询数据属性详情"},
		{"/ontology/modelDatatypeProperty/getModelDatatypePropertyList", "GET", "本体建模", "分页查询数据属性"},
		{"/ontology/modelDatatypeProperty/instantiateModelDatatypeProperty", "POST", "本体建模", "数据属性模板挂载"},
		{"/ontology/modelObjectProperty/createModelObjectProperty", "POST", "本体建模", "创建对象属性"},
		{"/ontology/modelObjectProperty/updateModelObjectProperty", "PUT", "本体建模", "更新对象属性"},
		{"/ontology/modelObjectProperty/deleteModelObjectProperty", "DELETE", "本体建模", "删除对象属性"},
		{"/ontology/modelObjectProperty/findModelObjectProperty", "GET", "本体建模", "查询对象属性详情"},
		{"/ontology/modelObjectProperty/getModelObjectPropertyList", "GET", "本体建模", "分页查询对象属性"},
		{"/ontology/modelObjectProperty/suggestInverseModelObjectProperty", "GET", "本体建模", "反向关系建议"},
		{"/ontology/modelObjectProperty/instantiateModelObjectProperty", "POST", "本体建模", "对象属性模板挂载"},
		{"/ontology/extModule/getExtModuleList", "GET", "本体建模", "外部模块列表"},
		{"/ontology/extModule/createExtModule", "POST", "本体建模", "新增外部模块"},
		{"/ontology/extModule/updateExtModule", "PUT", "本体建模", "更新外部模块"},
		{"/ontology/extModule/deleteExtModule", "DELETE", "本体建模", "删除外部模块"},
		{"/ontology/extTable/getExtTableList", "GET", "本体建模", "模块下注册表清单"},
		{"/ontology/extTable/registerExtTables", "POST", "本体建模", "批量注册物理表"},
		{"/ontology/extTable/deleteExtTable", "DELETE", "本体建模", "删除注册表"},
		{"/ontology/extTable/probeExtTables", "GET", "本体建模", "探测物理表"},
		{"/ontology/extTable/getExtTableColumns", "GET", "本体建模", "探测表列"},
		{"/ontology/extBinding/getExtBindingList", "GET", "本体建模", "绑定分页"},
		{"/ontology/extBinding/findExtBinding", "GET", "本体建模", "绑定详情"},
		{"/ontology/extBinding/createExtBinding", "POST", "本体建模", "新建绑定草稿"},
		{"/ontology/extBinding/updateExtBinding", "POST", "本体建模", "更新绑定草稿"},
		{"/ontology/extBinding/deleteExtBinding", "DELETE", "本体建模", "删除绑定"},
		{"/ontology/extBinding/changeExtBindingStatus", "PUT", "本体建模", "绑定生效/停用"},
		{"/ontology/extSync/dryRunExtSync", "POST", "本体建模", "绑定试运行"},
		{"/ontology/extSync/triggerExtSync", "POST", "本体建模", "触发同步"},
		{"/ontology/extSync/getExtSyncLogList", "GET", "本体建模", "同步日志分页"},
	}
	// 治理 CRUD 授 888；供给/建模接口额外授予「建模师」角色（若存在）
	governanceRoles := []string{"888"}
	modelerRoles := append([]string{"888"}, findModelerAuthorityIds()...)

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
		if s.Group == "本体供给" || s.Group == "本体建模" {
			roles = modelerRoles
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
