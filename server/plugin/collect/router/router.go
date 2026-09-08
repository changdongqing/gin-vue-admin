package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/collect/api"
	"github.com/gin-gonic/gin"
)

// CollectRouter 采集平台路由（03 文档 §六）。
// path 无 /api 前缀（GVA 惯例：PrivateGroup = Router.Group(RouterPrefix)，
// 前端 axios baseURL=/api 代理转换）；casbin 按 path+method 判权。
type CollectRouter struct{}

func (CollectRouter) Init(private *gin.RouterGroup) {
	a := api.CollectApi{}
	g := private.Group("collect")
	{
		// 通道
		g.POST("channels", a.CreateChannel)
		g.PUT("channels", a.UpdateChannel)
		g.DELETE("channels/:id", a.DeleteChannel)
		g.GET("channels/:id", a.FindChannel)
		g.GET("channels", a.GetChannelList)
		g.GET("tree", a.GetTree)
		// 设备
		g.POST("devices", a.CreateDevice)
		g.PUT("devices", a.UpdateDevice)
		g.DELETE("devices/:id", a.DeleteDevice)
		g.GET("devices", a.GetDeviceList)
		// 测点
		g.POST("variables", a.CreateVariable)
		g.PUT("variables", a.UpdateVariable)
		g.DELETE("variables/:id", a.DeleteVariable)
		g.GET("variables", a.GetVariableList)
		// 设备类型
		g.GET("deviceTypes", a.GetDeviceTypeList)
		g.POST("deviceTypes", a.CreateDeviceType)
		g.PUT("deviceTypes", a.UpdateDeviceType)
		g.DELETE("deviceTypes/:id", a.DeleteDeviceType)
		// 部署
		g.POST("deploy/:id", a.DeployChannel)
		g.DELETE("deploy/:id", a.UndeployChannel)
		g.POST("deploy/rebuild-all", a.RebuildAll)
		g.GET("deployments", a.GetDeployments)
		// 子流程库
		g.POST("parseChains", a.CreateParseChain)
		g.PUT("parseChains", a.UpdateParseChain)
		g.DELETE("parseChains/:id", a.DeleteParseChain)
		g.GET("parseChains", a.GetParseChainList)
		g.POST("parseChains/:id/publish", a.PublishParseChain)
		g.POST("parseChains/:id/test", a.TestParseChain)
		// Excel 导入导出
		g.GET("import/template", a.GetImportTemplate)
		g.POST("import/preview", a.ImportPreview)
		g.POST("import/commit", a.ImportCommit)
		g.GET("import/export", a.ExportConfig)
		// 实时数据与状态
		g.GET("realtime", a.GetRealtime)
		g.GET("channels/:id/status", a.GetChannelStatus)
	}
}
