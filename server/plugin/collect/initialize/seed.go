package initialize

import (
	"fmt"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemService "github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"go.uber.org/zap"
)

// 采集平台幂等种子（照 report_seed 范式）：菜单 + 888 角色授权 + API 登记 + casbin p 规则。

// SeedCollect 幂等注册采集平台菜单/API/casbin（global.GVA_DB 就绪后调用）。
func SeedCollect() {
	if global.GVA_DB == nil {
		return
	}
	seedCollectMenus()
	seedCollectApis()
	seedCollectDemo()
}

func seedCollectMenus() {
	db := global.GVA_DB
	// 顶级目录
	var dir system.SysBaseMenu
	if err := db.Where("name = ? AND parent_id = 0", "collect").First(&dir).Error; err != nil {
		dir = system.SysBaseMenu{
			ParentId:  0,
			Path:      "collect",
			Name:      "collect",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      22,
			Meta: system.Meta{
				Title:     "采集平台",
				Icon:      "odometer",
				KeepAlive: false,
			},
		}
		if err := db.Create(&dir).Error; err != nil {
			global.GVA_LOG.Error("采集种子：目录菜单创建失败", zap.Error(err))
			return
		}
	}
	seeds := []struct {
		Name, Title, Icon, Component string
		Sort                         int
	}{
		{Name: "collectConfig", Title: "采集配置", Icon: "setting", Component: "view/collect/config/index.vue", Sort: 1},
		{Name: "collectMonitor", Title: "采集监控", Icon: "data-line", Component: "view/collect/monitor/index.vue", Sort: 2},
		{Name: "collectParseChains", Title: "子流程库", Icon: "document", Component: "view/collect/parseChains/index.vue", Sort: 3},
	}
	menuIDs := make([]uint, 0)
	for _, s := range seeds {
		var count int64
		db.Model(&system.SysBaseMenu{}).Where("parent_id = ? AND name = ?", dir.ID, s.Name).Count(&count)
		if count > 0 {
			var m system.SysBaseMenu
			db.Where("parent_id = ? AND name = ?", dir.ID, s.Name).First(&m)
			menuIDs = append(menuIDs, m.ID)
			continue
		}
		menu := system.SysBaseMenu{
			ParentId: dir.ID, Path: s.Name, Name: s.Name, Hidden: false,
			Component: s.Component, Sort: s.Sort,
			Meta: system.Meta{Title: s.Title, Icon: s.Icon, KeepAlive: false},
		}
		if err := db.Create(&menu).Error; err != nil {
			global.GVA_LOG.Error("采集种子：菜单创建失败", zap.String("name", s.Name), zap.Error(err))
			continue
		}
		menuIDs = append(menuIDs, menu.ID)
	}
	// 授权 888
	for _, id := range menuIDs {
		var count int64
		db.Model(&system.SysAuthorityMenu{}).Where("sys_base_menu_id = ? AND sys_authority_authority_id = ?", id, "888").Count(&count)
		if count == 0 {
			if err := db.Create(&system.SysAuthorityMenu{
				MenuId:      strconv.FormatUint(uint64(id), 10),
				AuthorityId: "888",
			}).Error; err != nil {
				global.GVA_LOG.Error("采集种子：菜单授权失败", zap.Uint("menuId", id), zap.Error(err))
			}
		}
	}
	// RunLog 调试 WS 走 bridge（?token=），不经 casbin；此处无需登记
	_ = fmt.Sprint()
}

func seedCollectApis() {
	db := global.GVA_DB
	type apiSeed struct{ Path, Method, Group, Description string }
	seeds := []apiSeed{
		{"/collect/channels", "GET", "采集平台", "通道列表"},
		{"/collect/channels/:id", "GET", "采集平台", "通道详情"},
		{"/collect/channels", "POST", "采集平台", "创建通道"},
		{"/collect/channels", "PUT", "采集平台", "更新通道"},
		{"/collect/channels/:id", "DELETE", "采集平台", "删除通道"},
		{"/collect/tree", "GET", "采集平台", "通道设备测点树"},
		{"/collect/devices", "GET", "采集平台", "设备列表"},
		{"/collect/devices", "POST", "采集平台", "创建设备"},
		{"/collect/devices", "PUT", "采集平台", "更新设备"},
		{"/collect/devices/:id", "DELETE", "采集平台", "删除设备"},
		{"/collect/variables", "GET", "采集平台", "测点列表"},
		{"/collect/variables", "POST", "采集平台", "创建测点"},
		{"/collect/variables", "PUT", "采集平台", "更新测点"},
		{"/collect/variables/:id", "DELETE", "采集平台", "删除测点"},
		{"/collect/deploy/:id", "POST", "采集平台", "编译部署通道"},
		{"/collect/deploy/:id", "DELETE", "采集平台", "下线通道"},
		{"/collect/deploy/rebuild-all", "POST", "采集平台", "全量重编译"},
		{"/collect/deployments", "GET", "采集平台", "部署历史"},
		{"/collect/deviceTypes", "GET", "采集平台", "设备类型列表"},
		{"/collect/deviceTypes", "POST", "采集平台", "创建设备类型"},
		{"/collect/deviceTypes", "PUT", "采集平台", "更新设备类型"},
		{"/collect/deviceTypes/:id", "DELETE", "采集平台", "删除设备类型"},
		{"/collect/parseChains", "GET", "采集平台", "子流程列表"},
		{"/collect/parseChains", "POST", "采集平台", "创建子流程"},
		{"/collect/parseChains", "PUT", "采集平台", "更新子流程"},
		{"/collect/parseChains/:id", "DELETE", "采集平台", "删除子流程"},
		{"/collect/parseChains/:id/publish", "POST", "采集平台", "发布子流程"},
		{"/collect/parseChains/:id/test", "POST", "采集平台", "子流程回放测试"},
		{"/collect/import/template", "GET", "采集平台", "下载导入模板"},
		{"/collect/import/preview", "POST", "采集平台", "导入预览"},
		{"/collect/import/commit", "POST", "采集平台", "导入提交"},
		{"/collect/import/export", "GET", "采集平台", "配置导出"},
		{"/collect/realtime", "GET", "采集平台", "实时数据查询"},
		{"/collect/channels/:id/status", "GET", "采集平台", "通道运行状态"},
	}
	addedRules := make([][]string, 0)
	for _, s := range seeds {
		var count int64
		db.Model(&system.SysApi{}).Where("path = ? AND method = ?", s.Path, s.Method).Count(&count)
		if count == 0 {
			if err := db.Create(&system.SysApi{Path: s.Path, Method: s.Method, ApiGroup: s.Group, Description: s.Description}).Error; err != nil {
				global.GVA_LOG.Error("采集种子：API登记失败", zap.String("path", s.Path), zap.Error(err))
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
			global.GVA_LOG.Error("采集种子：casbin规则插入失败", zap.String("path", s.Path), zap.Error(err))
			continue
		}
		addedRules = append(addedRules, []string{"888", s.Path, s.Method})
	}
	if len(addedRules) > 0 {
		if err := systemService.CasbinServiceApp.FreshCasbin(); err != nil {
			global.GVA_LOG.Error("采集种子：casbin热加载失败（重启后生效）", zap.Error(err))
		} else {
			global.GVA_LOG.Info(fmt.Sprintf("采集种子：已新增 %d 条api规则", len(addedRules)))
		}
	}
}
