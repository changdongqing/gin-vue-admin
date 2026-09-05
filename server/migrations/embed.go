// Package migrations 通过 go:embed 将版本化 SQL 迁移文件编译进二进制，
// 使迁移不依赖运行目录/部署拷贝（与 iofs source driver 配合使用）。
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
