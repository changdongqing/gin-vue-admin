package initialize

import (
	"errors"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"
)

// MigrateDatabase 版本化数据库迁移（golang-migrate，迁移文件 go:embed 内嵌）
//
// 三种场景（在 RegisterTables/AutoMigrate 之前执行，保证版本化变更先行）：
//  1. 全新空库：无任何业务表 → 从零执行全部迁移（含 000001_baseline 建全量表）；
//  2. 存量库升级（无 schema_migrations 表但有业务表）：自动"基线固化"——
//     不执行 baseline 脚本（表已存在），仅将版本记录置为 baseline-version，
//     之后的结构变更从 000002 增量迁移开始受版本管理；
//  3. 已版本化库：执行未应用的增量迁移（Up），出错时保持 dirty 状态人工介入。
//
// 业务库决策：仅适配 PostgreSQL（pgsql），不扩展其它数据库的迁移支持；
// 非 pgsql 配置直接告警并跳过迁移（AutoMigrate 兜底），遇到时应改用 pgsql。
func MigrateDatabase() {
	cfg := global.GVA_CONFIG
	if !cfg.Migrate.Enable {
		global.GVA_LOG.Info("versioned migration disabled by config, skipping")
		return
	}
	if cfg.System.DbType != "pgsql" {
		global.GVA_LOG.Warn(fmt.Sprintf("versioned migration supports pgsql only, db-type %q skipped (fallback to AutoMigrate)", cfg.System.DbType))
		return
	}
	// 存量库自动基线固化：无版本表但有业务表。
	// 注意：必须先探测再创建 migrate 实例——postgres.WithInstance 会自动创建
	// schema_migrations 表，先建实例会让"存量库判定"永远失效。
	if hasBizTable() && !hasVersionTable() {
		version := cfg.Migrate.BaselineVersion
		if version <= 0 {
			version = 1
		}
		m, err := newMigrateInstance()
		if err != nil {
			global.GVA_LOG.Error("migrate: init failed", zap.Error(err))
			return
		}
		if err := m.Force(version); err != nil {
			global.GVA_LOG.Error("migrate: baseline force failed", zap.Int("version", version), zap.Error(err))
			return
		}
		global.GVA_LOG.Info(fmt.Sprintf("migrate: 存量库检测到业务表且无版本记录，已自动基线固化到版本 %d（后续结构变更请编写 00000%d 增量迁移）", version, version+1))
		return
	}

	m, err := newMigrateInstance()
	if err != nil {
		global.GVA_LOG.Error("migrate: init failed", zap.Error(err))
		return
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		global.GVA_LOG.Error("migrate: up failed, database may be dirty, fix then run again", zap.Error(err))
		return
	}
	version, dirty, vErr := m.Version()
	if vErr != nil && !errors.Is(vErr, migrate.ErrNilVersion) {
		global.GVA_LOG.Error("migrate: read version failed", zap.Error(vErr))
		return
	}
	if !errors.Is(err, migrate.ErrNoChange) {
		global.GVA_LOG.Info("migrate: versioned migrations applied successfully")
	}
	global.GVA_LOG.Info(fmt.Sprintf("migrate: current schema version=%d dirty=%v", version, dirty))
}

// RunMigrateCommand 执行运维迁移命令（./server -c config.yaml -migrate-cmd <cmd>），
// 返回 true 表示命令已执行完毕、主进程应退出（不启动 Web 服务）。
// 支持命令：
//
//	up          应用全部未执行迁移
//	down        回退最近一个迁移
//	down:N      回退 N 个迁移
//	force:N     将版本强制置为 N（dirty 修复用，需人工确保 schema 与 N 一致）
//	version     打印当前版本与 dirty 状态
func RunMigrateCommand(cmd string) bool {
	if global.GVA_DB == nil {
		global.GVA_LOG.Error("migrate: db not initialized")
		return true
	}
	if !global.GVA_CONFIG.Migrate.Enable {
		global.GVA_LOG.Error("migrate: disabled by config, cannot run command")
		return true
	}
	m, err := newMigrateInstance()
	if err != nil {
		global.GVA_LOG.Error("migrate: init failed", zap.Error(err))
		return true
	}

	action, n, hasN := parseMigrateCommand(cmd)
	switch action {
	case "up":
		err = m.Up()
		if errors.Is(err, migrate.ErrNoChange) {
			global.GVA_LOG.Info("migrate: already up to date")
			err = nil
		}
	case "down":
		if !hasN {
			n = 1
		}
		err = m.Steps(-n)
	case "force":
		if !hasN {
			global.GVA_LOG.Error("migrate: force requires version, usage: -migrate-cmd force:N")
			return true
		}
		err = m.Force(n)
	case "version":
		v, dirty, vErr := m.Version()
		if vErr != nil && !errors.Is(vErr, migrate.ErrNilVersion) {
			err = vErr
		} else {
			global.GVA_LOG.Info(fmt.Sprintf("migrate: version=%d dirty=%v", v, dirty))
		}
	default:
		global.GVA_LOG.Error(fmt.Sprintf("migrate: unknown command %q, usage: up | down | down:N | force:N | version", cmd))
		return true
	}
	if err != nil {
		global.GVA_LOG.Error("migrate: command failed", zap.String("cmd", cmd), zap.Error(err))
		return true
	}
	global.GVA_LOG.Info(fmt.Sprintf("migrate: command %q executed successfully", cmd))
	return true
}

// parseMigrateCommand 解析 "down" / "down:2" / "force:3" 形式命令
func parseMigrateCommand(cmd string) (action string, n int, hasN bool) {
	idx := -1
	for i, c := range cmd {
		if c == ':' {
			idx = i
			break
		}
	}
	if idx == -1 {
		return cmd, 0, false
	}
	action = cmd[:idx]
	fmt.Sscanf(cmd[idx+1:], "%d", &n)
	return action, n, n != 0
}

// newMigrateInstance 创建 golang-migrate 实例（内嵌迁移文件 + 当前主库连接）
func newMigrateInstance() (*migrate.Migrate, error) {
	sqlDB, err := global.GVA_DB.DB()
	if err != nil {
		return nil, err
	}
	instance, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return nil, err
	}
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, err
	}
	return migrate.NewWithInstance("iofs", src, "postgres", instance)
}

// hasVersionTable schema_migrations 是否存在
func hasVersionTable() bool {
	var count int64
	tx := global.GVA_DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = 'schema_migrations'").Scan(&count)
	if tx.Error != nil {
		global.GVA_LOG.Error("migrate: check version table failed", zap.Error(tx.Error))
		return false
	}
	return count > 0
}

// hasBizTable 是否存在任意业务表（用于区分"全新空库"与"存量库"）
func hasBizTable() bool {
	var count int64
	tx := global.GVA_DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = current_schema() AND table_name <> 'schema_migrations'").Scan(&count)
	if tx.Error != nil {
		global.GVA_LOG.Error("migrate: check biz table failed", zap.Error(tx.Error))
		return false
	}
	return count > 0
}
