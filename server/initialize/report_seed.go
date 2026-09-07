package initialize

import (
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/report"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	reportService "github.com/flipped-aurora/gin-vue-admin/server/service/report"
	systemService "github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"go.uber.org/zap"
)

// SeedReport 报表平台幂等种子（菜单/API/casbin/演示数据源；字典不种，枚举走前端本地 options）
// 1. 菜单：顶级目录「自定义报表」+ 子菜单（01 数据源管理先行，02~05 陆续追加）
// 2. 角色菜单：888（admin）关联新菜单
// 3. API：登记各功能路由 + casbin p 规则（888）
// 4. 演示数据源：demo_pg（按本库 pgsql 连接配置生成，供演示数据集取数）
func SeedReport() {
	if global.GVA_DB == nil {
		return
	}
	if err := reportService.ValidateAesKey(); err != nil {
		global.GVA_LOG.Error("报表平台：report.aes-key 配置非法，数据源密码加密不可用", zap.Error(err))
	}
	seedReportMenus()
	seedReportApis()
	seedDemoDataSource()
}

// seedReportMenus 追加报表平台目录与子菜单并授权给 888
func seedReportMenus() {
	db := global.GVA_DB
	var dir system.SysBaseMenu
	if err := db.Where("name = ? AND parent_id = 0", "report").First(&dir).Error; err != nil {
		dir = system.SysBaseMenu{
			ParentId:  0,
			Path:      "report",
			Name:      "report",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      21,
			Meta: system.Meta{
				Title:     "自定义报表",
				Icon:      "data-analysis",
				KeepAlive: false,
			},
		}
		if err := db.Create(&dir).Error; err != nil {
			global.GVA_LOG.Error("报表种子：目录菜单创建失败", zap.Error(err))
			return
		}
	}
	type menuSeed struct {
		Name, Title, Icon, Component string
		Sort                         int
	}
	seeds := []menuSeed{
		{Name: "dataSource", Title: "数据源管理", Icon: "coin", Component: "view/report/dataSource/dataSource.vue", Sort: 1},
		{Name: "dataSet", Title: "数据集管理", Icon: "tickets", Component: "view/report/dataSet/dataSet.vue", Sort: 2},
		{Name: "excelReport", Title: "Excel报表", Icon: "document", Component: "view/report/excel/excelReport.vue", Sort: 3},
		// 后续：analysisReport(sort=4,icon=data-line)
	}
	// hidden 菜单：Excel 报表预览页（「添加到菜单」的目标路由；Hidden 仍注册路由）
	hiddenSeeds := []menuSeed{
		{Name: "reportExcelViewer", Title: "报表预览", Icon: "view", Component: "view/report/excel/preview/preview.vue", Sort: 90},
	}
	for _, h := range hiddenSeeds {
		var count int64
		db.Model(&system.SysBaseMenu{}).Where("parent_id = ? AND name = ?", dir.ID, h.Name).Count(&count)
		if count > 0 {
			continue
		}
		menu := system.SysBaseMenu{
			ParentId:  dir.ID,
			Path:      "reportpreview",
			Name:      h.Name,
			Hidden:    true,
			Component: h.Component,
			Sort:      h.Sort,
			Meta: system.Meta{
				Title:     h.Title,
				Icon:      h.Icon,
				KeepAlive: false,
			},
		}
		if err := db.Create(&menu).Error; err != nil {
			global.GVA_LOG.Error("报表种子：hidden菜单创建失败", zap.String("name", h.Name), zap.Error(err))
		}
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
			global.GVA_LOG.Error("报表种子：菜单创建失败", zap.String("name", s.Name), zap.Error(err))
			continue
		}
	}
	// 目录 + 全部子菜单授权给 888
	var menus []system.SysBaseMenu
	db.Where("name = ? OR parent_id = ?", "report", dir.ID).Find(&menus)
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
			global.GVA_LOG.Error("报表种子：角色菜单授权失败", zap.Uint("menuId", m.ID), zap.Error(err))
		}
	}
}

// seedReportApis 登记报表平台接口并配置 casbin 规则
func seedReportApis() {
	db := global.GVA_DB
	type apiSeed struct {
		Path, Method, Group, Description string
	}
	seeds := []apiSeed{
		{"/report/dataSource/getDataSourceList", "GET", "报表平台", "分页查询数据源"},
		{"/report/dataSource/findDataSource", "GET", "报表平台", "查询数据源详情"},
		{"/report/dataSource/getDataSourceAll", "GET", "报表平台", "已启用数据源全量下拉"},
		{"/report/dataSource/createDataSource", "POST", "报表平台", "创建数据源"},
		{"/report/dataSource/updateDataSource", "PUT", "报表平台", "更新数据源"},
		{"/report/dataSource/deleteDataSource", "DELETE", "报表平台", "删除数据源"},
		{"/report/dataSource/testDataSourceConnection", "POST", "报表平台", "测试数据源连接"},
		{"/report/dataSet/getDataSetList", "GET", "报表平台", "分页查询数据集"},
		{"/report/dataSet/findDataSet", "GET", "报表平台", "查询数据集详情"},
		{"/report/dataSet/getDataSetAll", "GET", "报表平台", "已启用数据集全量下拉"},
		{"/report/dataSet/createDataSet", "POST", "报表平台", "创建数据集"},
		{"/report/dataSet/updateDataSet", "PUT", "报表平台", "更新数据集"},
		{"/report/dataSet/deleteDataSet", "DELETE", "报表平台", "删除数据集"},
		{"/report/dataSet/testDataSetPreview", "POST", "报表平台", "数据集测试预览"},
		{"/report/excelReport/getExcelReportList", "GET", "报表平台", "分页查询Excel报表"},
		{"/report/excelReport/findExcelReport", "GET", "报表平台", "查询Excel报表详情"},
		{"/report/excelReport/getExcelReportAll", "GET", "报表平台", "已启用Excel报表全量"},
		{"/report/excelReport/getExcelReportDataSetFields", "GET", "报表平台", "报表关联数据集字段列表"},
		{"/report/excelReport/createExcelReport", "POST", "报表平台", "创建Excel报表"},
		{"/report/excelReport/updateExcelReport", "PUT", "报表平台", "更新Excel报表"},
		{"/report/excelReport/deleteExcelReport", "DELETE", "报表平台", "删除Excel报表"},
		{"/report/excelReport/copyExcelReport", "POST", "报表平台", "复制Excel报表"},
		{"/report/excelReport/saveExcelTemplate", "POST", "报表平台", "保存Excel报表模板"},
		{"/report/excelReport/bindExcelReportDataSets", "POST", "报表平台", "Excel报表关联数据集"},
		{"/report/excelReport/previewExcelReport", "POST", "报表平台", "Excel报表预览渲染"},
		{"/report/excelReport/getExcelReportParamDefs", "GET", "报表平台", "Excel报表参数定义聚合"},
	}
	addedRules := make([][]string, 0)
	for _, s := range seeds {
		var count int64
		db.Model(&system.SysApi{}).Where("path = ? AND method = ?", s.Path, s.Method).Count(&count)
		if count == 0 {
			if err := db.Create(&system.SysApi{Path: s.Path, Method: s.Method, ApiGroup: s.Group, Description: s.Description}).Error; err != nil {
				global.GVA_LOG.Error("报表种子：API登记失败", zap.String("path", s.Path), zap.Error(err))
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
			global.GVA_LOG.Error("报表种子：casbin规则插入失败", zap.String("path", s.Path), zap.Error(err))
			continue
		}
		addedRules = append(addedRules, []string{"888", s.Path, s.Method})
	}
	if len(addedRules) > 0 {
		if err := systemService.CasbinServiceApp.FreshCasbin(); err != nil {
			global.GVA_LOG.Error("报表种子：casbin热加载失败（重启后生效）", zap.Error(err))
		} else {
			global.GVA_LOG.Info(fmt.Sprintf("报表种子：已新增 %d 条api规则", len(addedRules)))
		}
	}
}

// seedDemoDataSource 演示数据源 demo_pg：按本库 pgsql 连接配置生成（幂等；未配置 pgsql 跳过）。
// 密码经 AES 加密落库；aes-key 未配置时跳过（演示数据源非必需资产）
func seedDemoDataSource() {
	p := global.GVA_CONFIG.Pgsql
	if p.Dbname == "" {
		return
	}
	var count int64
	global.GVA_DB.Model(&report.ReportDataSource{}).
		Where("source_code = ? AND deleted_at IS NULL", "demo_pg").Count(&count)
	if count > 0 {
		return
	}
	// dsn 不内嵌密码，凭据走独立字段（password 经 AES 加密落库）
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s sslmode=disable", p.Path, p.Port, p.Dbname)
	sourceConfig := fmt.Sprintf(`{"dsn":%q,"username":%q,"password":%q}`, dsn, p.Username, p.Password)
	encrypted, err := reportService.EncryptSourceConfig(sourceConfig)
	if err != nil {
		global.GVA_LOG.Warn("报表种子：演示数据源创建跳过（加密不可用，请配置 report.aes-key）", zap.Error(err))
		return
	}
	ds := report.ReportDataSource{
		SourceCode:   "demo_pg",
		SourceName:   "演示数据源（本库）",
		SourceType:   "postgresql",
		SourceDesc:   "种子自动创建：指向本程序业务库，供演示数据集（demo_*）取数",
		SourceConfig: encrypted,
		EnableFlag:   true,
		CreatedBy:    "seed",
		UpdatedBy:    "seed",
	}
	if err := global.GVA_DB.Create(&ds).Error; err != nil {
		global.GVA_LOG.Error("报表种子：演示数据源创建失败", zap.Error(err))
	}
}
