package config

// Migrate 版本化数据库迁移配置（golang-migrate）
type Migrate struct {
	Enable          bool `mapstructure:"enable" json:"enable" yaml:"enable"`                         // 启动时执行版本化迁移
	BaselineVersion int  `mapstructure:"baseline-version" json:"baseline-version" yaml:"baseline-version"` // 存量库自动固化到的基线版本号
}
