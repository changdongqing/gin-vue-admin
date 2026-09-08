package rulego

import (
	"github.com/rulego/rulego/server/app"
	"github.com/rulego/rulego/server/model"
	"github.com/rulego/rulego/server/services"
	"go.uber.org/zap"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// ServiceUsername 采集插件回环 bridge 的服务账号名（collect/service/bridge.go 签发
// 的服务 JWT 即此身份，其名下保存采集主链/子流程链）。
const ServiceUsername = "collect-service"

// ensureServiceUser 把服务账号注册进 rulego 用户存储（users INI，落盘持久）。
// 引擎 InitUserEngines 启动时按该注册表逐一校验 workflows/<username> 目录，未注册
// 账号的目录被视作"已删用户孤儿数据"跳过加载（skip orphan data dir for removed
// user）——服务账号不是真实 GVA 用户、也不会走 rulego 本地注册（WithoutLocalAuth 已
// 关闭 /users*），若不补登，其名下全部采集链在重启后消失，采集触发器 notify 一律
// chain not found。仅影响存储层租户归属判定，不产生可登录账号（本地认证已关闭）。
func ensureServiceUser(a *app.App) {
	userAdmin, err := app.GetAs[services.UserAdmin](a.Container(), services.KeyUserAdmin)
	if err != nil {
		global.GVA_LOG.Warn("rulego 服务账号注册跳过：user 模块不可用", zap.Error(err))
		return
	}
	if _, ok := userAdmin.Get(ServiceUsername); ok {
		return
	}
	if err := userAdmin.Save(model.User{Username: ServiceUsername}); err != nil {
		global.GVA_LOG.Warn("rulego 服务账号注册失败：重启后采集链可能无法恢复", zap.Error(err))
		return
	}
	global.GVA_LOG.Info("rulego 服务账号已注册（采集链重启可恢复）", zap.String("username", ServiceUsername))
}
