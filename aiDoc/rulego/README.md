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
