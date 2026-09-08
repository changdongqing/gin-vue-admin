# RuleGo 服务端源码集成分析（aiDoc/rulego）

> 分析视角：Go 高级后端工程师 / Go 高级架构师 / Git 高级仓库管理专家
> 分析日期：2026-09-05　｜　评审修订：2026-09-08（对照 GVA 当前代码与官方 v0.37.2 tag 逐项实测核验）
> 集成对象：RuleGo（核心库 + RuleGo Server 服务端）
> 集成源：**开源提供方（官方）仓库的最新发布版本**
> 官方源码：https://gitee.com/rulego/rulego （GitHub 同名仓库镜像），当前最新 release **v0.37.2**
> 本地参考克隆：`/home/changdq/src/gitee/ttshang/rulego`（fork，其 main 分支已偏离 v0.37.2，仅作对照参考，**不作为集成源**；对照 API 时须以 `git show v0.37.2:...` 为准）

## 结论速览（TL;DR）

1. **可行且成本低**：RuleGo Server 官方就为"被宿主应用嵌入"而设计（Bridge 桥接层适配标准 `http.Handler`，明确支持 Gin 宿主；认证、存储、模块、生命周期全部可替换）。与 gin-vue-admin（Gin 技术栈）的对接是"官方预设场景"，不需要 fork 式深度改码。
2. **集成源决策**：以**官方仓库最新发布版 v0.37.2** 为集成基线（官方仓库本身就是"核心库 + server/"单仓双模块结构，无需依赖任何第三方 fork）；后续跟随官方 release 节奏升级。
3. **推荐方案**：`git subtree --squash` 将官方整仓内嵌到本仓库 `server/rulego/`，主 `go.mod` 用两条 `replace` 指向树内路径；在 gin-vue-admin 侧新建 `server/plugin/rulego` 插件作为宿主胶水层（挂载路由 + 桥接认证）。
4. **主要代价**（评审修订）：Go 版本**已满足**——GVA 主 `go.mod` 现已是 `go 1.25.0`，无升级成本；真实代价在依赖侧——**`mcp-go` 将被 rulego-server 从 v0.41.1 抬升至 v0.44.0，GVA 自身 `server/mcp/` 等大量代码必须回归验证**；传递依赖树显著增大（components 六大生态模块 + eino 等）；`go.sum` 与二进制体积增长。
5. **两个必须用胶水层弥合的对接缝隙**（评审新增，详见 01 文档 §4.5/§4.6）：① rulego 只认标准 `Authorization` 头（或 `?token=`），不认 GVA 的 `x-token`，需头映射中间件；② rulego 鉴权 SPI 收到的是 `resource/action` 语义标签（如 `rule`/`write`）而非 HTTP 路径，映射到 GVA Casbin 需固定映射表。
6. **前端**：**webui-vben 明确舍弃、不纳入集成范围**；规则链的可视化管理由 GVA 前端（Vue3 + Element Plus）自建页面承载。
7. **许可证无障碍**：双方均为 Apache 2.0，保留 LICENSE 与来源声明即可。

## 文档导航

| 文档 | 内容 |
| ---- | ---- |
| [01-集成方案分析.md](./01-集成方案分析.md) | 集成对象盘点、两侧架构对接分析（含认证/鉴权 SPI 实测签名）、六种集成方式对比与选型、目录与 go.mod 规划、依赖与版本影响（含共享依赖冲突表）、运行时对接与认证/鉴权桥接设计（§4.5/§4.6）、治理纪律、许可证合规、风险清单 |
| [02-实施计划.md](./02-实施计划.md) | 分四个里程碑的落地步骤、具体命令与代码骨架、官方版本升级 SOP、验证清单 |

## 实施状态（2026-09-08）

| 批次 | 内容 | 提交 |
| ---- | ---- | ---- |
| 文档评审 | 本目录设计文档评审修订 | `fd66815e` |
| M1 源码入库 | subtree 官方 v0.37.2 内嵌 `server/rulego/`、go.mod 双 replace、依赖抬升（mcp-go v0.44.0）、NOTICE、顺手修复 announcement/gen.go 指令误置 | `c62750a2` / `550193a3` / `61b25afa` |
| M2 桥接运行 | 胶水插件（挂载/降级/mapXToken/watchShutdown/信封）、resource 配置基线、冒烟测试 | `522432c2` |
| M3 认证统一 | Authenticator/Authorizer + 映射表 + WithoutLocalAuth，三态测试 | `76a60f6c` |
| M4 配套 | Dockerfile 固定 golang:1.25-alpine、CI go 1.25、compose rulego 数据卷 | 见 git log |

**验收状态**：`go build ./...` 全量通过；`go test ./plugin/rulego/...` 通过（401 信封 / admin-token 200 / 本地登录关闭 / 映射表全组合 / 头映射）。**待运行环境验证**（需 DB 的完整 GVA 进程与 Docker）：M2/M3 的 curl 三态、compose 一键起、subtree pull 演练。

**实施口径与文档的偏差记录**：

1. go.mod 中 rulego require 为 tidy 生成的占位版本号（`v0.0.0-00010101…`）——filesystem replace 下由 go 工具链管理，语义等同"树内源码"，与文档示例的 `v0.37.2` 不一致属预期；
2. 888 超管在 Authorizer 中显式放行（原因见 01 文档 §4.6 实施口径说明）；
3. `app.WithTransportDisabled()` 未启用——v0.37.2 中该选项无消费者，与官方 bridge 示例保持一致；
4. M4 前端「规则链管理」页面（02 文档 M4.4）**未实施**，按计划"可拆分独立排期"留作后续；GVA「API 管理」录入 `/rulego/**` 虚拟路径与普通角色授权属运行时数据操作，部署时执行。
