package main

import (
	"flag"

	"github.com/flipped-aurora/gin-vue-admin/server/core"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/initialize"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

//go:generate go env -w GO111MODULE=on
//go:generate go env -w GOPROXY=https://goproxy.cn,direct
//go:generate go mod tidy
//go:generate go mod download

// 这部分 @Tag 设置用于排序, 需要排序的接口请按照下面的格式添加
// swag init 对 @Tag 只会从入口文件解析, 默认 main.go
// 也可通过 --generalInfo flag 指定其他文件
// @Tag.Name        Base
// @Tag.Name        SysUser
// @Tag.Description 用户

// @title                       Gin-Vue-Admin Swagger API接口文档
// @version                     v2.9.2
// @description                 使用gin+vue进行极速开发的全栈开发基础平台
// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        x-token
// @BasePath                    /
// 迁移运维命令（包级注册，与 core.Viper 内部注册的 -c 同集合，由 core.Viper 统一 flag.Parse）
var migrateCmd = flag.String("migrate-cmd", "", "版本化迁移命令: up | down | down:N | force:N | version")

func main() {
	// 初始化系统
	initializeSystem()
	// 迁移命令执行后直接退出，不启动 Web 服务
	if *migrateCmd != "" {
		initialize.RunMigrateCommand(*migrateCmd)
		return
	}
	// 运行服务器
	core.RunServer()
}

// initializeSystem 初始化系统所有组件
// 提取为单独函数以便于系统重载时调用
func initializeSystem() {
	global.GVA_VP = core.Viper() // 初始化Viper
	initialize.OtherInit()
	global.GVA_LOG = core.Zap() // 初始化zap日志库
	zap.ReplaceGlobals(global.GVA_LOG)
	global.GVA_DB = initialize.Gorm() // gorm连接数据库
	initialize.Timer()
	initialize.DBList()
	initialize.SetupHandlers() // 注册全局函数
	if global.GVA_DB != nil {
		initialize.MigrateDatabase()    // 版本化数据库迁移（先于 AutoMigrate，保证变更按版本落地）
		initialize.RegisterTables()     // AutoMigrate 兜底（纯新增表/列，幂等）
		initialize.SeedDataPermission() // 数据权限体系幂等种子（菜单/API/casbin/演示组织）
		initialize.SeedOntology()       // 本体治理幂等种子（菜单/API/casbin）
		initialize.SeedReport()         // 报表平台幂等种子（菜单/API/casbin/演示数据源）
	}
}
