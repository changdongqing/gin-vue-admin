package config

// Report 报表平台配置
type Report struct {
	// AesKey 数据源密码加密密钥（32字节随机数的base64，生成：openssl rand -base64 32）
	// 未配置时数据源密码无法加密存储（创建/修改数据源会被拒绝），不影响系统其它功能启动
	AesKey string `mapstructure:"aes-key" json:"aes-key" yaml:"aes-key"`
}
