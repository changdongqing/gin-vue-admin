// Package rulego 是 gin-vue-admin 嵌入 RuleGo 服务端（官方 v0.37.2，源码内嵌于
// server/rulego/，上游只读）的宿主胶水层。设计依据见 aiDoc/rulego/。
//
// 职责边界（树的只读纪律）：对 rulego 的全部定制只允许落在本包，方式为官方扩展点——
// bridge.WithoutLocalAuth / WithResponseWrapper、app.WithAuthenticator / WithAuthorizer 等。
package rulego

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	interfaces "github.com/flipped-aurora/gin-vue-admin/server/utils/plugin/v2"
	"github.com/gin-gonic/gin"
	"github.com/rulego/rulego/server/app"
	"github.com/rulego/rulego/server/bootstrap"
	"github.com/rulego/rulego/server/bridge"
	"go.uber.org/zap"
)

const prefix = "/rulego"

// App 暴露 rulego App 实例，供后续扩展使用（如配置热更新 App.Reload）。
var App *app.App

type plugin struct{}

var Plugin = new(plugin)

func init() {
	interfaces.Register(Plugin)
}

func (p *plugin) Register(group *gin.Engine) {
	// 本方法经 initialize.InstallPlugin 调用；GVA_DB 未初始化时会被延迟到数据库就绪
	// 回调后执行（rulego 不依赖 GVA DB，接受该时序，见 aiDoc/rulego 风险 #7）。
	cfgFile := configPath()
	if _, err := os.Stat(cfgFile); err != nil {
		global.GVA_LOG.Error("rulego 配置文件缺失，/rulego/** 不挂载", zap.String("config", cfgFile), zap.Error(err))
		return
	}

	// 失败降级：规则引擎是增强能力，构造失败只告警、不挂路由，不得 panic
	//（Register 运行在 GVA 启动链路上，panic 会杀死整个进程）。
	//
	// 认证统一：WithoutLocalAuth 关闭 rulego 自带 /login、/users* 端点（其硬约束
	// require_auth=true 已在 resource/rulego/config.conf 落实），身份体系桥接为
	// GVA JWT（Authenticator）+ Casbin（Authorizer），rulego 侧不再产生独立登录态。
	b, err := bridge.New(
		bridge.WithAppOptions(
			app.WithConfigFile(cfgFile), // base_path=/rulego、require_auth=true 等基线见该配置文件
			app.WithModules(bootstrap.DefaultModules()...),
			app.WithAuthenticator(&gvaAuthenticator{}),
			app.WithAuthorizer(&gvaAuthorizer{}),
		),
		bridge.WithoutLocalAuth(),
		bridge.WithResponseWrapper(gvaEnvelope),
	)
	if err != nil {
		global.GVA_LOG.Error("rulego bridge 初始化失败，/rulego/** 不可用", zap.Error(err))
		return
	}
	App = b.App()
	ensureServiceUser(App)
	watchShutdown(b)

	// rulego 路由注册时已含 BasePath 前缀，挂载时勿 StripPrefix（官方 bridge 约定）。
	// mapXToken 把 GVA 前端的 x-token 映射为标准 Authorization 头（rulego 只认
	// Authorization / ?token=，不识别 x-token）。
	group.Any(prefix+"/*path", mapXToken(), gin.WrapH(b.Handler()))
	global.GVA_LOG.Info("rulego bridge 已挂载", zap.String("prefix", prefix))
}

// mapXToken GVA 前端统一携带 x-token（web/src/utils/request.js），而 rulego 的
// extractAuthorization 只读标准 Authorization 头或 ?token= query。此处补映射，前端零改动。
func mapXToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			if t := c.GetHeader("x-token"); t != "" {
				c.Request.Header.Set("Authorization", "Bearer "+t)
			}
		}
		c.Next()
	}
}

// watchShutdown GVA v2 插件接口只有 Register、无 Stop 钩子，core/server_run.go 的优雅
// 退出只调用 srv.Shutdown，因此胶水层自订阅退出信号关闭 rulego App（signal.Notify 支持
// 多订阅者，与 initServer 的信号监听并存安全）。
func watchShutdown(b *bridge.Bridge) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		if err := b.Stop(); err != nil {
			global.GVA_LOG.Error("rulego App 停止异常", zap.Error(err))
		}
	}()
}
