package report

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	// 按 sourceType 注册 database/sql 驱动（开箱三类，信创/Oracle 按需启用，见 driverRegistry）
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/microsoft/go-mssqldb"
)

// driverSpec 类型-驱动-验证 SQL 注册表
type driverSpec struct {
	driverName   string // database/sql 注册名
	validator    string // 测试连接验证 SQL
	rebindDollar bool   // true=占位符 ? 需 rebind 为 $N（PG 系）
	enabled      bool   // 是否随本期启用（未启用类型报 ErrSourceDriverDisabled）
}

// driverRegistry 数据源类型注册表（前端 DATA_SOURCE_TYPE_OPTIONS 全量枚举对齐；
// 启用信创三库：go get 驱动 → import 触发 sql.Register → 本表置 enabled: true）
var driverRegistry = map[string]driverSpec{
	"postgresql": {driverName: "pgx", validator: "SELECT 1", rebindDollar: true, enabled: true},
	"mysql":      {driverName: "mysql", validator: "SELECT 1", enabled: true},
	"sqlserver":  {driverName: "sqlserver", validator: "SELECT 1", enabled: true},
	"mariadb":    {driverName: "mysql", validator: "SELECT 1"},                         // MySQL 协议复用 mysql 驱动，按需启用
	"dameng":     {driverName: "dm", validator: "SELECT 1", rebindDollar: true},        // gitee.com/chunanyong/dm，按需启用
	"kingbase":   {driverName: "kingbase", validator: "SELECT 1", rebindDollar: true},  // 按需启用
	"opengauss":  {driverName: "opengauss", validator: "SELECT 1", rebindDollar: true}, // 按需启用
	"oracle":     {driverName: "godror", validator: "SELECT 1 FROM DUAL"},              // 需 CGO，默认不启用
	"http":       {},                                                                   // HTTP 数据集不建连接池
	// highgo/shentong/gbase：驱动生态不成熟，仅枚举预留
}

// IsSourceTypeValid sourceType 是否在枚举注册表内
func IsSourceTypeValid(sourceType string) bool {
	_, ok := driverRegistry[sourceType]
	return ok
}

// IsSourceTypeEnabled sourceType 是否随本期启用（http 类型无连接配置，视为可用）
func IsSourceTypeEnabled(sourceType string) bool {
	spec, ok := driverRegistry[sourceType]
	return ok && (spec.enabled || sourceType == "http")
}

// RebindDollarNeeded 驱动是否需要 ? → $N 占位符改写（02 SqlParamResolver 消费）
func RebindDollarNeeded(sourceType string) bool {
	return driverRegistry[sourceType].rebindDollar
}

// DataSourcePoolManager 连接池管理器：按 sourceCode 缓存各数据源 *sql.DB（02 取数复用）
type DataSourcePoolManager struct {
	pool sync.Map // sourceCode → *sql.DB
}

// GetOrCreate 获取或创建连接池（sourceConfig 中 password 须已解密）。
// 建池即 Ping（10s 超时）拦截坏配置；并发下 LoadOrStore 防重复建池。
func (m *DataSourcePoolManager) GetOrCreate(sourceCode, sourceType, sourceConfig string) (*sql.DB, error) {
	if db, ok := m.pool.Load(sourceCode); ok {
		return db.(*sql.DB), nil
	}
	spec, ok := driverRegistry[sourceType]
	if !ok {
		return nil, ErrSourceTypeInvalid
	}
	if !spec.enabled {
		return nil, ErrSourceDriverDisabled
	}
	dsn, err := buildDsn(sourceType, sourceConfig)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(spec.driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSourceConnFailed, err)
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(30 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("%w: %v", ErrSourceConnFailed, err)
	}
	actual, _ := m.pool.LoadOrStore(sourceCode, db)
	if actual != db { // 并发下别人先建了池，关掉自己这份多余连接
		_ = db.Close()
	}
	return actual.(*sql.DB), nil
}

// TestConnection 测试连接（临时连接执行验证 SQL，不写入池缓存；调用方无需 Close）
func (m *DataSourcePoolManager) TestConnection(sourceType, sourceConfig string) error {
	spec, ok := driverRegistry[sourceType]
	if !ok {
		return ErrSourceTypeInvalid
	}
	if sourceType == "http" {
		return nil // HTTP 类型无连接配置
	}
	if !spec.enabled {
		return ErrSourceDriverDisabled
	}
	dsn, err := buildDsn(sourceType, sourceConfig)
	if err != nil {
		return err
	}
	db, err := sql.Open(spec.driverName, dsn)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSourceConnFailed, err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrSourceConnFailed, err)
	}
	return nil
}

// Remove 数据源配置变更/删除时关闭旧池（下次取数自动重建）
func (m *DataSourcePoolManager) Remove(sourceCode string) {
	if db, ok := m.pool.LoadAndDelete(sourceCode); ok {
		_ = db.(*sql.DB).Close()
	}
}

// buildDsn 组装驱动 DSN：sourceConfig = {dsn, username, password}；
// 独立 username/password 优先拼入（未提供则信任 dsn 内嵌凭据）
func buildDsn(sourceType, sourceConfig string) (string, error) {
	cfg, err := parseSourceConfig(sourceConfig)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(cfg.Dsn) == "" {
		return "", fmt.Errorf("%w: dsn 不能为空", ErrSourceConfigInvalid)
	}
	dsn := strings.TrimSpace(cfg.Dsn)
	if cfg.Username == "" {
		return dsn, nil
	}
	switch sourceType {
	case "postgresql", "kingbase", "opengauss", "highgo", "shentong":
		return withPgCredentials(dsn, cfg.Username, cfg.Password)
	case "mysql", "mariadb":
		if strings.Contains(dsn, "@") { // 已内嵌凭据
			return dsn, nil
		}
		return fmt.Sprintf("%s:%s@%s", cfg.Username, cfg.Password, dsn), nil
	case "sqlserver":
		if strings.Contains(strings.ToLower(dsn), "user id") {
			return dsn, nil
		}
		sep := ";"
		if !strings.HasSuffix(dsn, ";") && strings.Contains(dsn, "=") {
			sep = ";"
		}
		return fmt.Sprintf("%s%suser id=%s;password=%s;", dsn, sep, cfg.Username, cfg.Password), nil
	case "dameng":
		return withPgCredentials(dsn, cfg.Username, cfg.Password) // dm:// URL 形态同族
	default:
		return dsn, nil
	}
}

// withPgCredentials 向 PG 族 DSN 拼入凭据：URL 形态改写 userinfo，keyword 形态追加 user/password
func withPgCredentials(dsn, username, password string) (string, error) {
	lower := strings.ToLower(dsn)
	if strings.HasPrefix(lower, "postgres://") || strings.HasPrefix(lower, "postgresql://") ||
		strings.HasPrefix(lower, "dm://") {
		u, err := url.Parse(dsn)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrSourceConfigInvalid, err)
		}
		if password != "" {
			u.User = url.UserPassword(username, password)
		} else {
			u.User = url.User(username)
		}
		return u.String(), nil
	}
	// keyword/value 形态：去掉已有 user/password 关键字后追加
	parts := splitKeywordDsn(dsn)
	kept := make([]string, 0, len(parts))
	for i := 0; i < len(parts); i++ {
		key := strings.ToLower(strings.TrimSpace(parts[i]))
		if key == "user" || key == "password" {
			i++ // 跳过其值
			continue
		}
		kept = append(kept, parts[i])
	}
	kept = append(kept, "user="+username, "password="+password)
	return strings.Join(kept, " "), nil
}

// splitKeywordDsn 简易拆分 keyword/value DSN（容忍引号值内空格）
func splitKeywordDsn(dsn string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	for _, r := range dsn {
		switch {
		case r == '\'':
			inQuote = !inQuote
			current.WriteRune(r)
		case r == ' ' && !inQuote:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

// GetOrCreateFromStored 按已保存数据源取池（解密配置 → GetOrCreate；01/02 衔接入口）
func (m *DataSourcePoolManager) GetOrCreateFromStored(sourceCode string) (*sql.DB, error) {
	ds, err := DataSourceServiceApp.GetDataSourceByCode(sourceCode)
	if err != nil {
		return nil, err
	}
	if !ds.EnableFlag {
		return nil, ErrSourceDisabled
	}
	decrypted, err := DecryptSourceConfig(ds.SourceConfig)
	if err != nil {
		return nil, err
	}
	return GlobalPoolManager.GetOrCreate(ds.SourceCode, ds.SourceType, decrypted)
}

// DataSourceServiceApp / GlobalPoolManager 领域内单例（pool 与 data_source 互相依赖，
// 经单例变量解耦，避免 enter.go 聚合结构上的循环引用）
var (
	GlobalPoolManager    = &DataSourcePoolManager{}
	DataSourceServiceApp = &DataSourceService{}
)
