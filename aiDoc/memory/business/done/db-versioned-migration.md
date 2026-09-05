# 数据库版本化迁移（方案A）落地与 pgsql 决策固化

## 基本信息

- 提出日期：2026-09-05
- 当前状态：`done`
- 需求类型：基础设施 / 工程规则
- 优先级：高
- 需求文件：`aiDoc/memory/business/done/db-versioned-migration.md`
- 长期规范：`aiDoc/memory/long-term/db-migration-rules.md`

## 用户原始意图摘要

评估本项目数据库版本管理现状后，按"方案 A"落地版本化迁移（golang-migrate），
并将后续开发必须遵守的数据库版本开发规范固化为长期记忆；
明确业务库仅需适配 PostgreSQL。

## 影响范围

- 后端：`server/migrations/`（基线+内嵌）、`server/initialize/migrate.go`（三场景+CLI）、
  `server/main.go`、`server/config/migrate.go`、`server/config.yaml`、`server/core/viper.go`
- 文档：`aiDoc/db-migration/README.md`、`AGENT.MD`（数据库与迁移规则）、
  `aiDoc/memory/long-term/db-migration-rules.md`、`aiDoc/memory/project-memory.md`
- 数据库：新增 `schema_migrations` 版本表（自动管理）

## 涉及对象

- 接口：无对外 API 变更
- 页面：无
- 配置：`migrate.enable`（默认 true）、`migrate.baseline-version`（默认 1）；
  运维 CLI `-migrate-cmd {version|up|down|down:N|force:N}`
- 提交：`211adaf4`（核心）、`e81ed867`（gofmt）

## 已确认约束

- 业务库仅适配 PostgreSQL（pgsql）：迁移执行器与 SQL 不扩展 mysql/mssql/oracle/sqlite
- 结构变更一律写版本化迁移；禁止修改已发布迁移文件
- AutoMigrate 仅开发期兜底；生产 `disable-auto-migrate: true` 时以迁移为准

## 结果与验证

- 存量库自动基线固化、空库从零建 36 表、增量 up/down、坏迁移 dirty+force 修复，
  全链路实测通过（详见 `aiDoc/db-migration/README.md` 第 6 节）
