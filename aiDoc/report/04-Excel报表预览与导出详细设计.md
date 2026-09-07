# 报表平台 · 04 Excel 报表预览与导出 详细设计

> 版本：v1.1（**按 dongqing 实际实现校准**：渲染引擎/平铺参数/前端导出均以 `ExcelReportRenderServiceImpl` + `views/report/excel/reportpreview/` 实现为准）｜ 日期：2026-09-07 ｜ 分支：dq
> 对应参考实现：dongqing `service/excel/render/ExcelReportRenderServiceImpl.java` + `views/report/excel/reportpreview/index.vue` + `utils/xlsx-export.ts` + `ExcelReportController#preview/#dataset-params`
> 前置依赖：[03](./03-Excel报表设计与模板管理详细设计.md)（模板 JSON / bind-datasets 产出）、[02](./02-数据集管理详细设计.md)（`QueryPage`/参数解析）
> 本文档为详细设计，不含代码交付；文中代码为实现指引的范式示例。
> 页面路由：预览页隐藏菜单 `/report/reportpreview?reportCode=…`

---

## 目录

- [一、功能概述](#一功能概述)
- [二、术语与业务规则](#二术语与业务规则)
- [三、核心算法：渲染引擎（对齐实现）](#三核心算法渲染引擎对齐实现)
- [四、后端设计](#四后端设计)
- [五、前端设计](#五前端设计)
- [六、菜单 / API / casbin 种子](#六菜单--api--casbin-种子)
- [七、实施步骤](#七实施步骤)
- [八、验证清单（AC 对照）](#八验证清单ac-对照)
- [九、风险与注意事项](#九风险与注意事项)

---

## 一、功能概述

### 1.1 功能定位

预览/导出是 Excel 报表的**消费端**：后端渲染引擎将「模板快照 + 数据集数据 + 平铺查询参数」渲染为最终报表快照；预览页以**主数据集服务端分页**逐页取渲染结果并用只读 Univer 展示；**导出在前端完成**——按当前条件拉全量渲染快照（pageNo=1、pageSize=50000），用 SheetJS 转为 .xlsx 下载（与设计器 xlsx 导入互为逆过程）。

### 1.2 功能范围

**含**：

1. **渲染引擎 `ExcelReportRenderService`**：定位主数据集（cellData 中含完整占位符的第一个）→ 主数据集分页查询 → 明细模板行多数据集对齐展开 → 单值/文本内占位符替换 → 行数裁剪；
2. **预览接口** `preview`：入参 `{reportCode, paramValues(平铺), pageNo, pageSize}`，响应 `{snapshot(渲染后快照对象), total(主数据集总行数)}`；
3. **参数定义聚合接口** `dataset-params`：报表关联的全部数据集参数平铺返回（**跨数据集按参数名去重**），供预览页查询条件表单；
4. **预览页**：参数表单（sampleItem 预填；默认值表达式由后端解析）+ 只读 Univer + el-pagination（默认 100/页，可选 10/20/50/100）；
5. **前端导出**：`exportSnapshotToXlsx`（值 + 合并单元格 + 列宽，无样式迁移），文件名 `{报表名}.xlsx`。

**不含**：后端二进制导出接口（实现未做，列后续迭代增强：excelize 全量渲染导出可复用本引擎）、PDF 导出、明细区合并单元格位移（实现明确不做，见 §九）、warnings 容错机制（实现中任一数据集查询失败预览直接报错）、禁用报表拦截（实现未校验 status，列后续增强）。

### 1.3 与 v1.0 设计的差异校准（以实现为准）

| # | v1.0 设计 | 实际实现（**采用**） | 校准说明 |
|---|---|---|---|
| 1 | `setParam` 嵌套 `{setCode:{param:value}}` | **`paramValues` 平铺** `{paramName: value}`；同名参数跨数据集共用（dataset-params 按参数名去重，保留先出现者） | 实现的参数模型；同名参数在不同数据集中应语义一致（如同为 `date`） |
| 2 | 主数据集 = 绑定单元格数量最多 | **主数据集 = cellData 中含「完整占位符」的第一个**（按 sheet→行→列扫描序；找不到取 setCodes 首个兜底） | 实现语义更简单直观 |
| 3 | 其余数据集全量取数 + 首行替换 | **明细模板行内引用的其它数据集按同一分页各自查询**（各显示各的）；仅单值场景数据集 `queryPage(1,1)` 取首行 | 实现支持"同行多数据集并排展开"（对账单明细+单价表场景） |
| 4 | 展开行数 = 主数据集行数 | **N = 模板行内各数据集本页行数的最大值**；第 j 行各数据集取自身第 j 行，**超出自身行数的单元格留空**（互不截断、互不重复） | 实现的"各显示各的"算法 |
| 5 | 合并单元格随行拷贝/跨行扩展/下方位移 | **不做 mergeData 处理**（仅参与行数裁剪计算）——明细区下方/跨明细区的合并会错位，接受为已知限制 | 实现取舍；模板设计时合并单元格放标题区 |
| 6 | 渲染输出 jsonStr 字符串 | **响应 `snapshot` 为 JSON 对象**（非字符串），前端直接使用 | 实现契约 |
| 7 | warnings 容错 + 多 sheet 独立分页 | **无容错**（数据集查询失败整体报错）；**全局单主数据集**（跨 sheet 扫描第一个，分页作用于此） | 实现语义 |
| 8 | 后端 excelize 导出 + downloadBlob | **前端 SheetJS 导出**：`preview(1, 50000)` → `exportSnapshotToXlsx(snapshot, filename)` | 实现方案（零后端导出接口；`xlsx` 包已在 03 引入） |
| 9 | 参数表单前端解析默认值表达式 | **前端仅用 sampleItem 预填**（日期类转 dayjs）；默认值表达式（today/thisMonth…）**由后端 fillDefaultParamValues 解析** | 实现分工，避免双端表达式漂移 |
| 10 | dateRange 前端拆 `_start/_end` 提交 | **前端拼 `"start,end"` 字符串提交，后端拆分**（fillDefaultParamValues 内统一处理 sampleItem/默认值/用户值三种来源） | 实现契约 |
| 11 | 自建 Go 渲染数据模型 | 渲染算法对齐，**Go 实现用 `map[string]interface{}` 树 + 行键数值排序**遍历（与 03 §3.1 的 RawMessage 方案合并为：cell 仅需读写 `v` 字段，整 cell 对象深拷贝） | Go 落地方式（§三附实现要点） |

---

## 二、术语与业务规则

| 术语 | 含义 |
|---|---|
| 完整占位符 | 单元格 `v` 整体匹配 `^#\{([^{}]+)\.([^{}]+)}$`（用于判定明细行）；字段名宽匹配（支持中文/连字符等非 `{}` 字符） |
| 片段占位符 | `#{set.field}` 混在文本中（如 `合计：#{a.qty} 件`），替换时仅替换片段 |
| 主数据集 | cellData 中含完整占位符的**第一个**数据集；它决定服务端分页（`pageNo/pageSize`/`total`） |
| 明细模板行 R | 每个 sheet 中主数据集完整占位符所在的**最小行号**行；整行（含同行静态单元格与样式）作为明细行模板 |
| 明细行数据集 | 模板行 R 内引用的**全部**数据集（含主数据集与同行其它数据集）——均按 `queryPage(pageNo,pageSize)` 取数、逐行对齐展开 |
| 单值数据集 | 不在模板行内、仅出现在标题等处的数据集——`queryPage(1,1)` 取首行做单值替换 |
| 行数裁剪 | 渲染后把 sheet `rowCount` 调整为实际内容行数（cellData 最大行号与 merge endRow 的最大值 +1），去掉模板默认空白行 |

**核心业务规则**：

1. **参数平铺去重**：`dataset-params` 跨数据集按 `paramName` 去重（同名保留先出现的数据集定义），按 `orderNum` 排序；渲染时**每个数据集各自**用平铺 `paramValues` 做缺省填充（用户值 > sampleItem > 默认值表达式）；
2. **分页语义**：`pageNo`（默认 1）/`pageSize`（后端默认 20，前端预览页默认 100）；`total` = 主数据集 COUNT；翻页只改 pageNo 重新渲染；
3. **展开规则**：N = 模板行内各明细数据集本页行数的最大值；第 j 份明细行（j=0..N-1）中每个数据集的占位符取**自身第 j 行**数据，超出留空；
4. **退化路径**：主数据集无完整占位符（全是文本内片段）或 N=0（本页无数据）→ 整表退化为**单值替换**（各数据集取首行，无数据替换为空串），保留模板行不展开；
5. **无模板/无关联数据集**：模板 `json_str` 为空 → 返回空快照对象 + total=0；`set_codes` 为空 → 原样返回模板 + total=0；
6. **导出**：前端按当前查询条件拉 `pageNo=1、pageSize=50000` 的全量渲染快照导出（受 02 行数上限保护）；文件名 `{报表名}.xlsx`。

---

## 三、核心算法：渲染引擎（对齐实现）

### 3.1 渲染总流程

```
render(reportCode, paramValues, pageNo, pageSize):
① 加载模板（json_str → 快照对象；为空 → 返回 {空快照, 0}）；set_codes 拆分 → setCodes
   （setCodes 为空 → 原样返回 {snapshot, 0}）
② 主数据集 = findMainSetCode：按 sheet→行→列扫描，首个「完整占位符」且 setCode ∈ setCodes 者；
   无 → setCodes[0] 兜底
③ 主数据集查询：fillDefaultParamValues(mainSetCode, paramValues) → queryPage(pageNo, pageSize, transforms)
   → mainData（本页行）+ total（分页依据）
④ 明细行数据集集合 detailSetCodes = 各 sheet 模板行 R 内引用的其它数据集（除主数据集）
⑤ 其余数据集查询：detail 数据集 queryPage(pageNo, pageSize)（与主数据集同页对齐）；
   单值数据集 queryPage(1, 1)（仅取首行）
⑥ 逐 sheet renderSheet（§3.2）→ adjustSheetRowCount（§3.3）
⑦ 返回 {渲染后快照对象, total}
```

### 3.2 单 sheet 渲染（明细行展开 + 占位符替换）

```
renderSheet(sheet, mainSetCode, mainData, otherData):
① R = 主数据集完整占位符所在最小行号；无 → 整表单值替换（firstRowSource）后返回
② templateRow = cellData[R]；rowSetCodes = 该行引用的全部数据集（完整或片段占位符）
③ N = max(各 rowSetCode 数据集本页行数)；N=0 → 整表单值替换后返回
④ 位移：cellData 中行键 > R 的行整体 +（N-1）（从大到小处理防覆盖）；
   ★ mergeData 不做位移（已知限制，见 §九）
⑤ 生成明细行：j = 0..N-1，深拷贝 templateRow 整行（含样式 cell 对象），占位符替换源 =
   rowSourceAt(j)（各数据集取自身第 j 行，超出 → null → 替换为空串）
⑥ 其余行（标题区等）单值替换：占位符替换源 = firstRowSource（各数据集首行）
```

**占位符替换**（一行内逐 cell）：

- 仅处理 `v` 为字符串且含 `#{` 的 cell；用片段正则 `#\{([^{}]+)\.([^{}]+)}` 逐个替换为 `String.valueOf(数据行[field])`（null → 空串）；替换后写回 `v`（`m`/`s` 等字段不动）。

### 3.3 行数裁剪（adjustSheetRowCount）

```
maxRow = max(cellData 所有行键, mergeData[].endRow)
sheet.rowCount = maxRow + 1（maxRow ≥ 0 时）
```

预览页每页只显示「标题行 + 本页数据行」，不再出现模板默认空白行。

### 3.4 Go 落地要点（实现指引）

- 快照用 `map[string]any` 树解析（`sheets→sheetId→cellData→row→col→cell`），cell 为 `map[string]any`（深拷贝 = 序列化往返或递归拷贝）；行键遍历需**数值排序**（Go map 无序）；
- 两组正则常量与实现一致：完整 `^#\{([^{}]+)\.([^{}]+)}$`、片段 `#\{([^{}]+)\.([^{}]+)}`；
- `fillDefaultParamValues` 复用 02 §4.7（每个数据集各自调用：用户值 > sampleItem > 默认值表达式；dateRange 字符串 `"起,止"` 拆 `_start/_end`——**用户传入的 dateRange 同样拆分**）；
- 渲染结果以 `json.RawMessage`（或 marshal 后的 `map`）直接放入响应 `data.snapshot`，前端免二次 parse。

---

## 四、后端设计

### 4.1 领域文件清单

```
server/
├── service/report/
│   ├── excel_render.go        # ★ ExcelReportRenderService（§3 算法 + RenderResult）
│   ├── param_resolve.go       # fillDefaultParamValues/ResolveSetParam（02 已建，此处复用）
│   └── errors.go              # 追加 04 段（精简：实现无 disabled/warnings 语义）
├── api/v1/report/excel_report.go   # 追加 2 handler（preview / dataset-params）
└── router/report/excel_report.go
```

### 4.2 错误定义（追加）

```go
// 04 预览
var ErrExcelRenderFailed = errors.New("报表渲染失败") // 拼接数据集查询等底层原因
// 复用：ErrExcelReportNotExists（模板不存在即报表未保存模板）、02 的数据集错误段
```

> 实现无「禁用报表拦截」与「warnings 容错」：预览只校验模板存在；任一数据集查询失败整体报错（前端 message 展示后端 msg）。

### 4.3 请求/响应结构

```go
// 预览请求（paramValues 平铺；dateRange 为 "起,止" 字符串，后端拆分）
type PreviewExcelReportReq struct {
	ReportCode  string                 `json:"reportCode" binding:"required"`
	ParamValues map[string]interface{} `json:"paramValues"`
	PageNo      int                    `json:"pageNo"`   // 默认 1
	PageSize    int                    `json:"pageSize"` // 默认 20
}
// 预览响应 data（snapshot 为渲染后快照【对象】，非字符串）
type PreviewExcelReportResp struct {
	Snapshot interface{} `json:"snapshot"`
	Total    int64       `json:"total"`
}
// 参数定义聚合响应（dataset-params，平铺 + 跨数据集按参数名去重，按 orderNum 排序）
type ReportParamDef struct {
	SetCode      string `json:"setCode"`      // 所属数据集（标签展示）
	ParamName    string `json:"paramName"`
	ParamDesc    string `json:"paramDesc"`
	ParamType    string `json:"paramType"`
	SampleItem   string `json:"sampleItem"`
	DefaultValue string `json:"defaultValue"`
	DictType     string `json:"dictType"`
	CustomOptions string `json:"customOptions"`
	DateFormat   string `json:"dateFormat"`
	RequiredFlag bool   `json:"requiredFlag"`
	OrderNum     int    `json:"orderNum"`
}
```

### 4.4 渲染服务（excel_render.go）

```go
type ExcelReportRenderService struct{ /* 依赖 templateMapper/dataSetMapper/paramMapper/transformMapper/queryService */ }

type RenderResult struct {
	Snapshot map[string]interface{} // 渲染后快照
	Total    int64                   // 主数据集总行数
}

// Render 预览渲染（分页，§3.1 流程；参数缺省填充按数据集各自执行）
func (r *ExcelReportRenderService) Render(reportCode string, paramValues map[string]interface{},
	pageNo, pageSize int) (*RenderResult, error)
// pageNo<1→1；pageSize<1→20；模板缺失 → ErrExcelReportNotExists
```

### 4.5 API

| 方法 | 路由 | 说明 |
|---|---|---|
| POST | `/report/excelReport/previewExcelReport` | 分页渲染（不记操作日志）；响应 `{snapshot, total}` |
| GET | `/report/excelReport/getExcelReportParamDefs` | 关联数据集参数定义平铺（query：reportCode） |

> 路由名沿用本设计集 camelCase 范式（实现为 `/preview`、`/dataset-params`，语义一致）；导出**无后端接口**（§5.4 前端导出）。

---

## 五、前端设计

### 5.1 目录结构

```
web/src/
├── api/report/excelPreview.js                  # preview / paramDefs API
└── view/report/excel/preview/
    ├── preview.vue                             # 预览页（hidden 菜单）
    └── xlsxExport.js                           # exportSnapshotToXlsx（SheetJS，与 03 xlsxImport 互逆）
```

### 5.2 API 层（api/report/excelPreview.js）

```javascript
import service from '@/utils/request'
export const previewExcelReport = (data) => service({ url: '/report/excelReport/previewExcelReport', method: 'post', data })
export const getExcelReportParamDefs = (params) => service({ url: '/report/excelReport/getExcelReportParamDefs', method: 'get', params })
```

### 5.3 预览页（preview.vue，对齐 reportpreview/index.vue）

```
布局（整页 flex 列、overflow-hidden，滚动只发生在 Univer 内部）：
┌ 页头：报表预览：{报表名} ｜ [返回]
├ 参数表单（params.length > 0 时）：grid 布局（label = 参数描述(参数名)+必填*）
│    ★ 建议抽公共组件 paramForm.vue（05 分析报表 ParamBar 共用：类型→组件映射/sampleItem 预填/buildParamValues 拼接）
│    number→el-input-number｜date→el-date-picker(date)｜datetime→el-date-picker(datetime)
│    dateRange→el-date-picker(daterange)｜select/multipleSelect→el-select（选项见下）｜其余→el-input
├ 操作行：[查询]（回第1页）[重置]（恢复示例值并查询）[导出]
├ 表格区：ReadonlySheet（03 UniverSheet readonly 模式，:key 重建加载新快照）
└ 分页：el-pagination（total=主数据集行数；默认 pageSize=100；选项 10/20/50/100；翻页/改页大小直接重新预览）

初始化 onMounted：
① page 接口按 reportCode 精确过滤（pageSize=1）校验存在 + 取 reportName（Excel 侧无 get-by-code，对齐实现）
② getExcelReportParamDefs → params
③ initFormValues（sampleItem 预填：dateRange 拆两段 dayjs、date/datetime 转 dayjs、其余原样；
   ★ 默认值表达式不在前端解析——空值不提交，交给后端 fillDefaultParamValues）
④ loadData(1, 100)

提交构造 buildParamValues：
空值跳过（交给后端默认值）；dateRange 拼提交 `"YYYY-MM-DD,YYYY-MM-DD"`；date/datetime 按 dateFormat
（缺省 YYYY-MM-DD / YYYY-MM-DD HH:mm:ss）格式化；multipleSelect 数组 join(',')（配合 02 resolve 的 IN 展开）

下拉参数选项 getParamOptions(p)：dictType 优先（GVA 字典接口）→ customOptions JSON（label/value，
兼容 text/name 键）→ 空选项
```

### 5.4 前端导出（xlsxExport.js，对齐 xlsx-export.ts）

```javascript
// exportSnapshotToXlsx(snapshot, filename)：SheetJS 社区版
// ① 逐 sheet：cellData → ws[addr] = { v: cell.v ?? cell.m ?? '' }（统计真实行列范围）
// ② mergeData → ws['!merges']（s/e 端点同步纳入范围）
// ③ columnData → ws['!cols']（w × 1.2 近似字符宽）
// ④ !ref = encode_range(真实范围；空表用模板 rowCount/columnCount 兜底)
// ⑤ XLSX.writeFile(workbook, filename)
// 调用：previewExcelReport({ reportCode, paramValues, pageNo:1, pageSize:50000 })
//       → exportSnapshotToXlsx(res.snapshot, `${reportName || reportCode}.xlsx`)
```

> **导出范围 = 当前查询条件下的全量数据**（一次 50000 上限内拉齐）；值/合并/列宽完整，**不迁移样式**（SheetJS 社区版能力边界，与 03 导入对称）；后端流式导出（excelize）列后续迭代增强。

---

## 六、菜单 / API / casbin 种子

`report_seed.go` 追加：

1. **hidden 菜单** `reportExcelViewer`（父=report 目录，path=**reportpreview**（对齐实现的路径段），`Hidden:true`，component=`view/report/excel/preview/preview.vue`，meta.title=报表预览）；
2. API/casbin：§4.5 两条（ApiGroup=报表平台，888）；03 列表页与设计器的「预览」按钮改跳 `router.push({ name:'reportExcelViewer', query:{ reportCode } })`。

**「添加到菜单」的预览路由适配**（03 §5.8 建立机制，本功能为落地对象）：预览组件同时支持 `route.query.reportCode` 与 path 内嵌 query（`reportpreview?reportCode=x` 经 asyncRouter 拆分注入）两种取值，兼容侧边栏菜单进入。

---

## 七、实施步骤

```
1. errors 精简 + param_resolve 复用确认（02 已建）
2. excel_render.go（§3 算法 + 单测：主数据集定位/多数据集对齐展开/退化单值/位移/行数裁剪/
   片段占位符文本替换/dateRange 用户值拆分）
3. api 追加 2 handler（preview 响应 snapshot 为对象）+ router + 种子（hidden 菜单 + 2 条 API/casbin）
4. 前端：api/excelPreview.js → xlsxExport.js → preview.vue（参数表单/分页/导出/ReadonlySheet 复用）
   → 03 列表页与设计器「预览」跳转接通
5. 联调 AC1~AC10（导出文件 Office/WPS 打开验证）
```

---

## 八、验证清单（AC 对照）

| AC | 验收项 | 标准 |
|----|--------|------|
| AC1 | 预览渲染 | 完整占位符替换正确；明细行按 N 展开；标题区片段占位符（`合计：#{a.qty}`）文本内替换正确 |
| AC2 | 多数据集对齐 | 模板行内多个数据集各显示各的（行数不同时短的留空，互不截断） |
| AC3 | 参数化查询 | 平铺参数表单正确生成（同名参数跨数据集仅一项）；改参查询刷新；默认值表达式由后端解析生效 |
| AC4 | dateRange | 前端拼 `起,止` 提交、后端拆 `_start/_end` 注入；必填缺失被拦截 |
| AC5 | 服务端分页 | total=主数据集 COUNT；默认 100/页；翻页/改页大小正常；大表不触发全量查询 |
| AC6 | 退化路径 | 无完整占位符 / 本页无数据 → 单值替换且保留模板行；空模板/无关联数据集返回空态 |
| AC7 | 行数裁剪 | 预览每页仅显示「标题+本页数据」行，无模板空白尾行 |
| AC8 | 前端导出 | 全量导出（值/合并/列宽），`{报表名}.xlsx`，Office/WPS 正常打开 |
| AC9 | 只读与布局 | 预览页 Univer 只读、无浏览器级滚动条；反复进出无泄漏 |
| AC10 | 权限/入口 | hidden 菜单不显示但路由可达；设计器/列表「预览」跳转正常；无权限接口 403 |

---

## 九、风险与注意事项

| 风险 | 等级 | 应对 |
|---|---|---|
| **明细区合并单元格错位**（实现已知限制） | 高（模板设计约束） | mergeData 不参与位移/扩展：明细模板行**下方或跨明细区的合并单元格**会错位——模板设计规范：合并单元格仅用于标题区（明细行之上）；完整合并位移列后续迭代 |
| 同名参数跨数据集语义漂移 | 中 | 平铺参数模型要求不同数据集的同名参数语义一致（如同为日期）；dataset-params 去重保留先出现者，参数描述中带 setCode 标签提示来源 |
| 导出一次性拉 50000 行渲染 | 中 | 受 02 行数上限保护；超限报错提示收窄条件；后端 excelize 流式导出列后续迭代 |
| 单值数据集每页重复取首行 | 低 | queryPage(1,1) 走 COUNT+LIMIT 1，开销可控；后续可加数据集级短 TTL 缓存 |
| Go map 无序遍历 | 低 | 行/列键必须数值排序后再遍历（明细行生成与位移顺序敏感）；单测覆盖乱序键用例 |
| 前端导出无样式 | 低 | 与 03 SheetJS 导入对称的能力边界；样式保真需求走后续后端导出 |
| 大快照传输 | 低 | 预览页仅当前页数据量；snapshot 为对象传输与实现一致 |

---

## 附：与 03/02/05 的衔接点

| 衔接点 | 状态 | 说明 |
|---|---|---|
| 模板 `json_str` / `set_codes` | 03 产出（bind-datasets 独立维护） | 渲染引擎输入 |
| `QueryPage`（分页方言/内存分页） | 02 已实现 | 主/明细/单值数据集取数 |
| `fillDefaultParamValues`（param_resolve.go） | 02 已建 | 渲染时按数据集各自解析（含用户 dateRange 拆分） |
| `getExcelReportParamDefs` | 本期实现 | 预览页参数表单；05 分析报表参数栏复用同一参数定义结构 |
| hidden 菜单 + path 内嵌 query 适配 | 本期落地（03 建机制） | 「添加到菜单」的预览入口 |
| 03 列表页/设计器「预览」占位 | 本期替换为真实跳转 | — |
