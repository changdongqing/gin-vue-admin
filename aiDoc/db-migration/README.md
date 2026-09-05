# 数据库版本化迁移（golang-migrate）设计与使用

> 落地日期：2026-09-05｜涉及代码：`server/migrations/`、`server/initialize/migrate.go`、`config.yaml migrate` 节
> 参考：JeeSite（flyway）与 yudao 的脚本化迁移实践，为 gin-vue-admin 补齐"版本化、可审计、可回滚"的库表演进能力

## 1. 机制总览（双轨）

| 轨道 | 工具 | 职责 | 触发 |
| --- | --- | --- | --- |
| 版本化迁移 | golang-migrate v4（`go:embed` 内嵌 SQL） | 结构/数据变更的**版本化落地**：基线建库、增量变更、回滚、dirty 修复 | 启动时 `MigrateDatabase()`（先于 AutoMigrate）；运维命令 `-migrate-cmd` |
| AutoMigrate 兜底 | gorm AutoMigrate | 开发期"改模型即生效"的兜底（幂等加表/加列） | 启动时 `RegisterTables()`（`disable-auto-migrate: false` 时） |

执行顺序（main.go）：

```
Gorm() 连接 → MigrateDatabase() → RegisterTables() → SeedDataPermission() → RunServer
```

版本化变更永远先于 AutoMigrate 落地；生产建议 `disable-auto-migrate: true`，
此时一切结构变更必须写迁移文件，AutoMigrate 只作为开发模式兜底。

## 2. 三种库场景

| 场景 | 判定 | 行为 |
| --- | --- | --- |
| 全新空库 | 无任何业务表 | 从零执行全部迁移（000001_baseline 建全量 36 张表） |
| 存量库升级 | 有业务表、无 `schema_migrations` | **自动基线固化**：不执行 baseline 脚本，仅把版本记录置为 `baseline-version`（默认 1），此后变更从 000002 起受版本管理 |
| 已版本化库 | 有 `schema_migrations` | 应用未执行的增量迁移（Up）；失败进入 dirty 状态，修复后 `force` 恢复 |

> 判定顺序很重要：**必须先探测、后创建 migrate 实例**——`postgres.WithInstance`
> 会自动创建 `schema_migrations` 表，先建实例会让"存量库判定"永远失效（线上踩坑记录）。

## 3. 迁移文件规范（server/migrations/）

```
000001_baseline.up.sql   # 基线：当前全量 schema（pg_dump --schema-only 快照，36 表）
000001_baseline.down.sql # 基线回滚：DROP 全部业务表
000002_xxxx.up.sql       # 第一个真实增量迁移示例（尚未创建，等待首个变更）
000002_xxxx.down.sql
embed.go                 # go:embed *.sql，迁移文件编译进二进制
```

- 命名：`<6位版本号>_<描述>.{up,down}.sql`，版本号必须连续且唯一（重复版本直接报错拒绝启动）。
- 每个迁移必须成对提供 up/down；up 与 down 都要**可重复执行安全**（多用
  `IF NOT EXISTS` / `IF EXISTS`），便于故障后重放。
- 迁移文件内嵌于二进制：不依赖运行目录、不依赖部署时拷贝；改迁移 = 走发布流程，与单体应用一致。
- 禁止修改已发布的迁移文件——错误只能通过"新迁移反向修正"或 down+force 回退处理。

## 4. 运维命令

```bash
./server -c config.yaml -migrate-cmd version   # 查看当前版本与 dirty 状态
./server -c config.yaml -migrate-cmd up        # 应用全部未执行迁移（启动时自动执行，一般无需手动）
./server -c config.yaml -migrate-cmd down      # 回退最近一个迁移
./server -c config.yaml -migrate-cmd down:2    # 回退最近 2 个迁移
./server -c config.yaml -migrate-cmd force:1   # dirty 修复：人工确认 schema 与版本 N 一致后强制置位
```

命令执行完自动退出，不启动 Web 服务。

## 5. 日常开发/升级 SOP

### 新增表/加列（推荐双轨写法）

1. 先写迁移：`migrations/000002_add_xxx.{up,down}.sql`
2. 再改 model 结构体（AutoMigrate 兜底对纯新增幂等，两边不冲突）
3. 启动验证：日志出现 `current schema version=2 dirty=false`

### 破坏性变更（删列/改类型/改名字/数据订正）

只写迁移，**不要依赖 AutoMigrate**（它只做加法）：
生产 `disable-auto-migrate: true`，up/down 完整实现，先测试库演练。

### 升级发布

1. 代码合并（含新迁移文件）→ 编译/出包
2. 测试库跑 `-migrate-cmd up` + 数据校验
3. 生产：备份 → 停旧实例 → 启动新实例（自动 Up）→ 观察日志
   `versioned migrations applied successfully` 与 `current schema version=N dirty=false`
4. 回滚：跑旧版二进制前先 `-migrate-cmd down:N` 回退结构（数据变更类迁移需自行评估）

### 迁移失败（dirty）处理

1. 日志出现 `up failed, database may be dirty`，`schema_migrations` 为 `(N, dirty=t)`
2. 定位失败原因（查看日志 SQL 错误）
3. 修复：若迁移文件本身写错 → 修正后 `force:N-1` 再启动自动 Up；
   若是数据问题 → 人工处理后 `force:N-1` 或按需回退
4. 原则：**先确认 schema 与版本一致，再 force**，绝不盲目 force 跳过失败

## 6. 验证记录（2026-09-05 实测）

| 场景 | 结果 |
| --- | --- |
| 存量库（gvadq，36 表无版本表）启动 | ✅ 自动固化 version=1 dirty=false |
| 全新空库启动 | ✅ 执行 baseline 建 36 表 + 种子正常，version=1 |
| `-migrate-cmd down`（基线） | ✅ 业务表清空（0 表） |
| 增量 000002 up | ✅ 加列成功 version=2 |
| 增量 down | ✅ 删列回滚 version=1，其余表无损 |
| 坏迁移制造 dirty | ✅ (2, dirty=t)，后续 up 被拒 |
| `force:1` 修复 + 重启 | ✅ version=1 dirty=false，正常启动 |
| 迁移文件版本冲突（两个 000002） | ✅ 启动即报 duplicate 拒绝，防呆生效 |

## 7. 范围与决策

- **业务库仅适配 PostgreSQL（`db-type: pgsql`）**：迁移执行器与全部迁移 SQL 只按
  pgsql 语义编写与验证；其他库类型配置会告警并跳过迁移（AutoMigrate 兜底），
  遇到时应改用 pgsql，不为此扩展 mysql/mssql/oracle/sqlite 的迁移支持。
- `schema_migrations` 由 golang-migrate 自动管理，勿手工删改；彻底重建环境时
  直接 drop database 或 `down` 至空。
- 与 `sys_versions`（发版公告）职责不同：前者是结构迁移史，后者是业务发版记录；
  如需打通可后续将发版记录关联迁移版本号。
- 长期规范（后续开发必须遵守）见 `aiDoc/memory/long-term/db-migration-rules.md`。
