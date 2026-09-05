# 数据库版本开发规范（长期约束）

> 建立日期：2026-09-05｜触发条件：**所有涉及 `server/model/` 变更或直接修改数据库结构的开发任务**

## 为什么会有这条记忆

项目已落地 golang-migrate 版本化迁移（见 `aiDoc/db-migration/README.md`）。
此后数据库结构的演进必须"版本化、可审计、可回滚"，不能再依赖纯 AutoMigrate 隐式改库。

## 强制规则

1. **结构变更必须写版本化迁移**：新增
   `server/migrations/00000N_描述.{up,down}.sql`（`N` = 当前最大版本号 + 1，
   查看版本用 `./server -c config.yaml -migrate-cmd version`）。
2. **禁止修改已发布的迁移文件**：错误只能通过"新迁移反向修正"或 down+force 回退。
3. **迁移文件要求**：up/down 成对提供；尽量可重复执行安全（`IF NOT EXISTS` / `IF EXISTS`）；
   版本号唯一且连续（重复版本会导致启动直接报错）。
4. **与 AutoMigrate 的关系（双轨）**：
   - 破坏性变更（删列、改类型、改列名、数据订正）**只能走迁移**，AutoMigrate 不做也不可靠；
   - 纯新增表/新增列可同时改 model（AutoMigrate 幂等兜底），但生产环境
     `disable-auto-migrate: true` 时**一律以迁移文件为准**，模型改动不算完成；
   - 启动顺序固定：`MigrateDatabase()` 先于 `RegisterTables()`，不要在 main.go 中调整。
5. **迁移文件内嵌于二进制**（`go:embed`）：新增迁移后必须重新编译，迁移才会生效。
6. **验收**：变更落地后运行并确认日志 `current schema version=N dirty=false`；
   失败（dirty）时按 `aiDoc/db-migration/README.md` 第 5 节流程修复，禁止盲目 force。
7. **仅适配 PostgreSQL**：业务库只支持 `db-type: pgsql`。迁移执行器、迁移 SQL、
   AutoMigrate 兜底均只按 PostgreSQL 语义编写与验证；
   **不再为 mysql / mssql / oracle / sqlite 扩展迁移支持**，遇到其它库类型场景应提示改用 pgsql。
8. **数据/种子订正**：存量数据的升级订正写入迁移 up 段；启动幂等种子
   （如 `SeedDataPermission`）只负责"当前态补齐"，不承担历史数据升级职责。

## 参考文档

- `aiDoc/db-migration/README.md`：机制、迁移写法、升级/回滚 SOP、运维命令
- `server/migrations/000001_baseline.up.sql`：基线写法样例（pg_dump 快照，勿改动）
