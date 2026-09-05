# GvaGrid —— 基于 vxe-grid 的统一表格封装

> 状态：**已交付**（2026-09-05）。设计 → 开发 → 冒烟验证 → 代码生成器模板改造已完成；实现差异见《01-详细设计方案》第 21 节"实现注记"。
> 交付范围：`web/src/components/gvaGrid/` 封装组件全套、18 个列表页迁移、`server/utils/autocode/template_funcs.go` 新增 4 个 GvaGrid 辅助函数、`table.vue.tpl` / `view.vue.tpl` 生成器模板改造。

本目录存放 gin-vue-admin 前端表格体系迁移到 vxe-grid 的完整设计资料。

| 文档 | 说明 |
| --- | --- |
| [01-详细设计方案.md](./01-详细设计方案.md) | 现状分析、总体架构、组件 API、各功能区（查询区 / 工具栏 / 表格 / 分页 / 数据代理）、自适应高度方案、主题暗色适配、迁移示例、风险与验收标准 |
| [02-开发计划.md](./02-开发计划.md) | 里程碑划分（M0 基建 → M1 试点 → M2 批量迁移 → M3 生成器与收尾 → M4 验收）、任务分解与工作量估算、24 个页面迁移清单与批次、回归 checklist、协作规范 |

## 一句话方案

以 vben-admin v5 的 `useVbenVxeGrid + BasicTable` 分层思想为蓝本，落地一套适合 gin-vue-admin 的 **`useGvaGrid` hook（配置与数据协议适配层）+ `GvaGrid` 薄壳组件（布局 / 插槽透传 / 高度自适应）**，底层完全使用 `vxe-grid` 原生一体化能力（`formConfig` 查询区 + `toolbarConfig` 工具栏 + 表格本体 + `pagerConfig` 分页 + `proxyConfig` 数据代理），全站 24 个列表页分五批迁移，并同步改造服务端代码生成器模板（`table.vue.tpl` / `view.vue.tpl`）。

## 关键技术决策速览

- **依赖**：`vxe-table@~4.21.5` + `vxe-pc-ui@~4.17.29`（官方推荐组合，用 `~` 锁版本）
- **数据协议适配**：吸收 gin-vue-admin 后端约定 —— `{ page, pageSize, orderKey, desc, ...搜索字段 }` 入参、`{ code: 0, data: { list, total } }` 出参、`orderKey` 驼峰转下划线（复用 `toSQLLine`）、行主键 `ID`
- **自适应高度**：封装组件用 `ResizeObserver` 相对布局滚动容器（`.gva-container/.gva-container2`）测量剩余高度；内部 `vxe-grid` 用 `height="auto"` 填充，表格区内部滚动；测量失败自动降级为 `max-height` 模式
- **主题**：vxe v4 全 CSS 变量驱动，`--vxe-ui-*` 映射到 Element Plus 运行时变量（主色跟随 `--el-color-primary`，暗色跟随 `html.dark`）
- **逃生舱**：任意查询项、工具栏、单元格都可退回插槽写法（兼容 `el-*` 组件、`v-auth` 指令），复杂页面（如 autoCode）不受封装约束
