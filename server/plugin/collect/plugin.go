// Package collect 采集平台插件：三级模型管理、编译部署联动、Excel 点表导入、
// 触发面（cron / MQTT 订阅）与实时数据查询。设计依据 aiDoc/rulego/03。
package collect

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/router"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/service"
	interfaces "github.com/flipped-aurora/gin-vue-admin/server/utils/plugin/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type plugin struct{}

var Plugin = new(plugin)

func init() { interfaces.Register(Plugin) }

func (p *plugin) Register(group *gin.Engine) {
	// 本方法经 InstallPlugin 调用；GVA_DB 未就绪时被延迟到数据库就绪回调后执行，
	// 此处依赖 GVA_DB 完成迁移与种子（仍判空防御）。
	if global.GVA_DB == nil {
		global.GVA_LOG.Error("采集平台：数据库未就绪，插件功能不可用")
		return
	}
	if err := global.GVA_DB.AutoMigrate(
		&model.CollectChannel{}, &model.CollectDevice{}, &model.CollectVariable{},
		&model.CollectDeviceType{}, &model.CollectParseChain{},
		&model.CollectDeployment{}, &model.CollectRealtime{},
	); err != nil {
		global.GVA_LOG.Error("采集平台：表结构迁移失败", zap.Error(err))
		return
	}
	initialize.SeedCollect()
	// 与 GVA 私有路由同惯例（JWT + Casbin），path 无 /api 前缀
	private := group.Group(global.GVA_CONFIG.System.RouterPrefix)
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	collectRouter := router.CollectRouter{}
	collectRouter.Init(private)
	global.GVA_LOG.Info("采集平台路由已挂载（/collect/**）")
	// 触发面（cron / MQTT 订阅）启动；引擎侧链状态随后以 DB 为准对账自愈
	service.StartTriggers()
	service.ReconcileChainsOnStartup()
}
