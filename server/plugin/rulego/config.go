package rulego

import "path/filepath"

// configPath 返回 rulego 配置文件路径（相对 GVA 工作目录 server/，与 resource 下其他
// 资产的定位方式一致）。文件缺失时 Register 降级：不挂载 /rulego/**，不影响 GVA 主功能。
func configPath() string {
	return filepath.Join("resource", "rulego", "config.conf")
}
