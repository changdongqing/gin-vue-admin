# 报表平台 · 03 Excel 报表设计与模板管理 详细设计

> 版本：v1.1（**按 dongqing 实际实现校准**：设计器为列表页内 Tab、bind-datasets 独立维护关联、保存模板仅提交 jsonStr、page 过滤加载、新增「添加到菜单」）｜ 日期：2026-09-07 ｜ 分支：dq
> 对应参考实现：dongqing `web/apps/web-antd/src/views/report/excel/{report,designer}/` + `ExcelReportController#save-template/#bind-datasets/#dataset-fields`
> 前置依赖：[02-数据集管理](./02-数据集管理详细设计.md)（`getDataSetAll` / `case_result` 字段来源已交付）
> 本文档为详细设计，不含代码交付；文中代码为实现指引的范式示例。
> 页面路由：`/report/excelReport`（设计器为列表页**页内 Tab**，无独立路由）

---

## 目录

- [一、功能概述](#一功能概述)
- [二、术语与业务规则](#二术语与业务规则)
- [三、数据库设计](#三数据库设计)
- [四、后端设计](#四后端设计)
- [五、前端设计](#五前端设计)
- [六、菜单 / API / casbin 种子](#六菜单--api--casbin-种子)
- [七、实施步骤](#七实施步骤)
- [八、验证清单（AC 对照）](#八验证清单ac-对照)
- [九、风险与注意事项](#九风险与注意事项)

---

## 一、功能概述

### 1.1 功能定位

报表是 Excel 报表的顶层实体：**元数据**（编码/名称/分组/状态）+ **模板内容**（Univer 工作簿快照 JSON、关联数据集、参数默认值），两表按 `report_code` 1:1 关联。本功能交付报表元数据 CRUD/复制 + **Univer 在线设计器**（列表页内 Tab 承载，三栏：数据集字段树 / 表格编辑区 / 单元格属性面板）+ **拖拽数据绑定**（字段→单元格写 `#{setCode.fieldName}` 占位符）+ **关联数据集独立维护**（bind-datasets）+ **导入 xlsx** + **添加到菜单**，为 04（预览渲染/导出）提供"含占位符的模板 JSON"产出能力。

### 1.2 功能范围

**含**：

1. 报表元数据 CRUD（分页 keyword/分组/状态、详情、create/update、删除）、复制（元数据+模板深拷贝，新编码）；
2. 模板保存 `saveExcelTemplate`（**仅提交 jsonStr（+setParam 预留），setCodes 由 bind 接口独立管理**）；
3. **关联数据集独立接口 `bindExcelReportDataSets`**：设计器左栏「关联数据集」弹窗多选提交，upsert 模板行的 `set_codes`（**不动 jsonStr**）——打破「保存模板需带 setCodes / 绑定需先有模板」的循环依赖；
4. 设计器：Univer 集成（preset 动态加载 + 模块级单例缓存）、三栏 splitpanes、工具栏（导入xlsx / 预览（跳 04 预览页）/ 保存 / 添加到菜单）；
5. 拖拽绑定（拖入**当前选中单元格**，无选区默认 A1）+「点 + 插入」等价交互（无选区时提示先选格）；
6. 属性面板：选中单元格坐标/值/占位符识别/值编辑回写；
7. 数据集字段接口（模板 `set_codes` → 各数据集字段列表，来自 `case_result` 首行 keys；**数据集不存在时抛错**而非跳过）；
8. **添加到菜单**：设计器工具栏 → 弹窗选择上级菜单/名称/路由/排序 → 前端调菜单新增 API，创建指向预览页并携带 reportCode 的菜单；
9. 数据集删除引用校验落地（`report_excel_templates.set_codes`：LIKE 粗筛 + Go 内按 `|` 拆分精确比对，命中拒删）。

**不含**：预览渲染/导出（04）、聚合绑定语法与 custom metadata（后续迭代）、协作编辑/版本管理、模板缩略图。

### 1.3 与 v1.0 设计的差异校准（以实现为准）

| # | v1.0 设计（源自 M3 文档） | 实际实现（**采用**） | 校准说明 |
|---|---|---|---|
| 1 | 设计器 = hideInMenu 静态路由（fullContent 全屏） | **设计器 = 列表页内 Tab**（`designerTabs`，key=reportCode，可多开/关闭；删除报表联动移除 Tab）；仅**预览页**为隐藏路由（04） | 实现把设计器做成列表页页签（与分析报表一致），交互更连贯 |
| 2 | 直接提供 `get-by-code`（采纳分析报表 v1.1 修订） | **Excel 侧无 get-by-code**：设计器/预览页经 `page` 接口按 reportCode 精确过滤（pageSize=1）取 id/名称，再 `find`（get）取详情；分析报表侧保留 get-by-code | 实现差异；Excel 报表编码可作为 page 精确过滤条件 |
| 3 | 关联数据集随 save-template 提交（左栏本地维护） | **bind-datasets 独立接口**：`{reportCode, setCodes[]}` → 校验报表/各数据集存在 → upsert 模板行**仅写 set_codes**；save-template 更新分支**不动 setCodes** | 打破 setCodes 与 jsonStr 的循环依赖；新建报表未保存模板也可先绑定数据集 |
| 4 | （未设计） | **添加到菜单**（AddToMenuModal）：前端调菜单 createMenu，`component = 预览页组件 + ?reportCode=xxx`（query 内嵌 component 字符串，菜单点击自动拼到 URL） | 实现新增能力；GVA 适配见 §5.8 |
| 5 | UniverSheet 基础封装 | 实装细节补齐：**Univer 0.25.x pin**；preset 模块**动态 import + 模块级单例缓存**；zh-CN locale **补丁**（forceString 两个缺失 key）；preset 配置 `sheets:{disableForceStringAlert, disableForceStringMark}`；readonly = `workbook.setEditable(false)`；选区监听 `onCommandExecuted`（命令 id 为**小写** selection 系列）；snapshot 变化统一走**父组件 `:key` 重建**（组件内不 watch） | 实装的兼容性经验，全部纳入 |
| 6 | `getReportDataSetFields` 对不存在数据集跳过 | **抛 `ErrExcelReportDataSetNotExists`**（编码被误删/拼写错误应显式暴露） | 实现语义 |
| 7 | 保存模板提交 setCodes | SaveTemplateReq 保留 setCodes 字段但**服务端忽略**（insert 分支置空，update 分支不动） | 对齐实现，字段保留兼容 |

其余（1:1 模板表、`set_codes` 用 `|` 分隔、字段来自 case_result、先选中再拖拽、页面列表不返回 jsonStr 大字段、复制深拷贝）与实现一致。

---

## 二、术语与业务规则

| 术语 | 含义 |
|---|---|
| reportCode | 报表编码，全局唯一，`^[a-zA-Z0-9_]+$` ≤100；创建后不可改（编辑禁用） |
| 模板（template） | `report_excel_templates` 行：`set_codes`（bind 接口维护）+ `set_param`（参数默认值 JSON，预留）+ `json_str`（save-template 维护）；与报表按 report_code **1:1**，两路写入互不干扰 |
| 快照（snapshot） | Univer `workbook.save()` 产出的 `IWorkbookData` JSON；**只读副本**，改单元格必须走命令 API（`getRange().setValue()`） |
| `#{setCode.fieldName}` | 数据绑定占位符；字段名宽匹配（非 `{}` 字符，支持中文/连字符）；04 渲染按「完整占位符判明细行 / 片段占位符文本替换」两态处理 |
| 设计器 Tab | 列表页行「设计」打开的页内页签（title=`报表名-设计`，key=reportCode），支持多报表并行、关闭、删除联动 |
| 添加到菜单 | 将当前报表的预览页注册为系统菜单项（携 reportCode），业务人员可从侧边栏直达 |

**核心业务规则**：

1. **reportCode 唯一**（软删可复用）；删除报表 → 级联软删模板；
2. **模板两路写入**：`save-template` 只写 `json_str/set_param`；`bind-datasets` 只写 `set_codes`（两者均 upsert 模板行，先到者 insert 空骨架行）；
3. **关联数据集约束**：bind 时逐个校验数据集存在；字段列表来自 `case_result` 首行 keys，为空则提示「请先在数据集管理准备结果案例」（02 演示种子/后续维护入口）；
4. **复制**：新 reportCode/新名称必填，元数据+模板（set_codes/set_param/json_str 深拷贝）复制，状态默认启用；
5. **数据集删除引用**：任一模板 `set_codes` 拆分后含该编码 → `ErrSetReferencedByReport`（LIKE 粗筛防子串误判由 Go 精确比对兜底）。

**Univer 集成铁律**（实装验证结论）：

1. `onMounted` 初始化、`onBeforeUnmount` dispose（workbook + univer 双 dispose，选区监听 disposable 一并释放）；
2. Univer 实例/门面用**模块级裸变量或 shallowRef**，绝不入 Vue 深响应式；
3. 快照仅在初始化时输入；运行期改值走 `setCellValue`；**快照更换（导入 xlsx）由父组件 `:key` 递增整体重建**，组件内不 watch snapshot。

---

## 三、数据库设计

迁移 `000010_create_report_platform` 中本功能两张表（v1.0 版本不变，此处为 v1.1 复述）：

```sql
-- 报表元数据
CREATE TABLE IF NOT EXISTS report_excel_reports (
    id           bigserial     PRIMARY KEY,
    report_code  varchar(100)  NOT NULL,
    report_name  varchar(100)  NOT NULL DEFAULT '',
    report_group varchar(100)  NOT NULL DEFAULT '',
    report_desc  varchar(255)  NOT NULL DEFAULT '',
    created_by   varchar(64)   NOT NULL DEFAULT '',
    updated_by   varchar(64)   NOT NULL DEFAULT '',
    created_at   timestamptz   NOT NULL DEFAULT now(),
    updated_at   timestamptz   NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);
COMMENT ON TABLE report_excel_reports IS '报表平台-Excel报表元数据表';
CREATE UNIQUE INDEX IF NOT EXISTS uk_report_excel_report_code
    ON report_excel_reports (report_code) WHERE deleted_at IS NULL;

-- 报表模板内容（与元数据按 report_code 1:1；set_codes 与 json_str 两路独立维护）
CREATE TABLE IF NOT EXISTS report_excel_templates (
    id          bigserial    PRIMARY KEY,
    report_code varchar(100) NOT NULL,
    set_codes   varchar(500) NOT NULL DEFAULT '',      -- bind-datasets 维护，| 分隔
    set_param   text,                                  -- 参数默认值 JSON（预留）
    json_str    text,                                  -- save-template 维护（Univer 快照）
    created_by  varchar(64) NOT NULL DEFAULT '',
    updated_by  varchar(64) NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);
COMMENT ON TABLE report_excel_templates IS '报表平台-Excel报表模板内容表';
CREATE UNIQUE INDEX IF NOT EXISTS uk_report_excel_template_code
    ON report_excel_templates (report_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_report_excel_template_set_codes
    ON report_excel_templates (set_codes) WHERE deleted_at IS NULL; -- 数据集引用校验粗筛
```

> 与 v1.0 的差异：去掉 `report_author`/`status` 两列——实现的状态开关未在列表页使用（预览也不校验 status），审计由 created_by 承担；若需禁用开关，列后续迭代再加（保持最小对齐实现）。

GORM 模型 `model/report/report_excel.go`（`ReportExcelReport` / `ReportExcelTemplate`，显式 TableName；`ensure_tables.go` 追加）。

---

## 四、后端设计

### 4.1 领域文件清单

```
server/
├── model/report/report_excel.go            # 两 DO
├── model/report/request/report_excel.go    # SearchExcelReport / ExcelReportOps / CopyReq / SaveTemplateReq / BindDataSetsReq
├── service/report/excel_report.go          # 元数据 CRUD/复制/模板 upsert(仅json)/bindDataSets/字段列表/引用校验
├── service/report/errors.go                # 追加 03 段
├── api/v1/report/excel_report.go           # 10 handler
└── router/report/excel_report.go
```

### 4.2 错误定义（追加）

```go
// 03 Excel 报表
var (
	ErrExcelReportNotExists       = errors.New("Excel报表不存在")
	ErrExcelReportCodeDuplicate   = errors.New("报表编码已存在")
	ErrExcelTemplateJSONInvalid   = errors.New("报表模板JSON格式错误")
	ErrExcelCopySourceMissing     = errors.New("复制源报表不存在")
	ErrExcelReportDataSetNotExists = errors.New("报表关联的数据集不存在") // 拼接 setCode
)
```

### 4.3 请求结构

```go
type SearchExcelReport struct {
	request.PageInfo
	ReportGroup string `form:"reportGroup"`
}
type ExcelReportOps struct{ ID uint `json:"ID" form:"ID" binding:"required"` }
type CopyExcelReportReq struct {
	SourceReportCode string `json:"sourceReportCode" binding:"required"`
	ReportCode       string `json:"reportCode" binding:"required"`
	ReportName       string `json:"reportName" binding:"required"`
}
// SaveTemplateReq：ReportCode required；JsonStr required（json.Valid 校验）；
// SetCodes 字段保留兼容但服务端忽略；SetParam 预留
type SaveExcelTemplateReq struct {
	ReportCode string `json:"reportCode" binding:"required"`
	SetCodes   string `json:"setCodes"`  // 兼容保留，服务端忽略
	SetParam   string `json:"setParam"`
	JsonStr    string `json:"jsonStr" binding:"required"`
}
// BindDataSetsReq：设计器「关联数据集」弹窗提交
type BindDataSetsReq struct {
	ReportCode string   `json:"reportCode" binding:"required"`
	SetCodes   []string `json:"setCodes"`
}
```

### 4.4 Service 层（excel_report.go）

```go
type ExcelReportService struct{}

// CreateReport / UpdateReport 元数据（编码查重排除自身；与 01 的 create/update 分离风格一致）
func (s *ExcelReportService) CreateReport(p *report.ReportExcelReport, operator string) error
func (s *ExcelReportService) UpdateReport(p *report.ReportExcelReport, operator string) error

// DeleteReport 级联软删模板
func (s *ExcelReportService) DeleteReport(id uint) error

// GetReport 详情 = 元数据 + 模板（setCodes/setParam/jsonStr 附加字段，聚合结构体返回）
func (s *ExcelReportService) GetReport(id uint) (*ExcelReportDetail, error)

// GetReportList 列表（keyword 匹配 reportCode/reportName；支持 reportCode 精确过滤——
// 设计器/预览页以 page(pageSize=1, reportCode 精确) 取 id/名称，对齐实现）
func (s *ExcelReportService) GetReportList(info req.SearchExcelReport) (list, total int64, err error) // 不取 json_str
func (s *ExcelReportService) GetEnabledReportAll() ([]report.ReportExcelReport, error)

// SaveTemplate upsert 模板行【仅 json_str/set_param】：报表存在 → json.Valid →
//   行不存在则 insert（set_codes 置空）→ 存在则 Updates(json_str, set_param)，不动 set_codes
func (s *ExcelReportService) SaveTemplate(r *req.SaveExcelTemplateReq, operator string) error

// BindDataSets upsert 模板行【仅 set_codes】：报表存在 → 逐个校验数据集存在（不存在抛
//   ErrExcelReportDataSetNotExists）→ setCodes join("|") → 行不存在 insert 骨架 / 存在仅 Updates(set_codes)
func (s *ExcelReportService) BindDataSets(r *req.BindDataSetsReq, operator string) error

// CopyReport 元数据+模板深拷贝（源存在/新编码唯一）
func (s *ExcelReportService) CopyReport(r *req.CopyExcelReportReq, operator string) (uint, error)

// GetReportDataSetFields 设计器左栏：set_codes 拆分 → 每数据集 {setCode,setName,fields}
// fields = extractFieldsFromCaseResult(case_result)（JSON 数组首行 keys）；
// ★ 数据集不存在 → 抛 ErrExcelReportDataSetNotExists（显式暴露脏引用，不静默跳过）
func (s *ExcelReportService) GetReportDataSetFields(reportCode string) ([]DataSetFields, error)

// IsSetCodeReferenced 数据集删除引用校验（挂到 02 DeleteDataSet）：
// LIKE '%'||code||'%' 粗筛 → Go 内按 | 拆分精确比对（防 abc 误匹配 xabc）
func (s *ExcelReportService) IsSetCodeReferenced(setCode string) (bool, error)
```

### 4.5 API 与路由

| 方法 | 路由 | 说明 |
|---|---|---|
| GET | `/report/excelReport/getExcelReportList` | 分页（keyword / reportGroup；**reportCode 精确过滤**供设计器/预览页定位） |
| GET | `/report/excelReport/findExcelReport` | 详情（含模板） |
| POST | `/report/excelReport/createExcelReport` | 新增元数据（记操作日志） |
| PUT | `/report/excelReport/updateExcelReport` | 更新元数据（记操作日志） |
| DELETE | `/report/excelReport/deleteExcelReport` | 删除（级联模板） |
| GET | `/report/excelReport/getExcelReportAll` | 已启用全量（预留） |
| POST | `/report/excelReport/copyExcelReport` | 复制 |
| POST | `/report/excelReport/saveExcelTemplate` | 保存模板（仅 jsonStr/setParam；记操作日志） |
| POST | `/report/excelReport/bindExcelReportDataSets` | 关联数据集（仅 setCodes；记操作日志） |
| GET | `/report/excelReport/getExcelReportDataSetFields` | 数据集字段列表（query：reportCode） |

（`preview` / `getExcelReportParamDefs` 两条归 04。）**响应约定**：`jsonStr`/`setParam` 普通字符串返回；列表接口不含 json_str。

---

## 五、前端设计

### 5.1 目录结构与新增依赖

```
web/src/
├── api/report/excelReport.js             # API + 常量
└── view/report/excel/
    ├── excelReport.vue                   # 报表列表页（GvaGrid + el-tabs 承载设计器 + 抽屉/复制弹窗）
    ├── designer/
    │   ├── designer.vue                  # 设计器主页（被列表页 Tab 承载；三栏 + 工具栏）
    │   ├── components/
    │   │   ├── UniverSheet.vue           # Univer 封装（§5.5 实装细节）
    │   │   ├── DatasetPanel.vue          # 左栏：关联数据集折叠面板（emit bind/insert）
    │   │   ├── DataSetSelectModal.vue    # 「关联数据集」多选弹窗（el-dialog）
    │   │   ├── AddToMenuModal.vue        # 「添加到菜单」弹窗（§5.8）
    │   │   └── PropertyPanel.vue         # 右栏：单元格属性/占位符识别/值回写
    │   └── utils/xlsxImport.js           # SheetJS → Univer 快照映射
```

```bash
cd web && npm i @univerjs/presets @univerjs/preset-sheets-core xlsx splitpanes
# Univer pin 到 0.25.x 具体 patch（对齐实装版本线），lockfile 入库
```

### 5.2 API 层（api/report/excelReport.js）

```javascript
import service from '@/utils/request'

export const getExcelReportList = (params) => service({ url: '/report/excelReport/getExcelReportList', method: 'get', params })
export const findExcelReport = (params) => service({ url: '/report/excelReport/findExcelReport', method: 'get', params })
export const createExcelReport = (data) => service({ url: '/report/excelReport/createExcelReport', method: 'post', data })
export const updateExcelReport = (data) => service({ url: '/report/excelReport/updateExcelReport', method: 'put', data })
export const deleteExcelReport = (params) => service({ url: '/report/excelReport/deleteExcelReport', method: 'delete', params })
export const copyExcelReport = (data) => service({ url: '/report/excelReport/copyExcelReport', method: 'post', data })
export const saveExcelTemplate = (data) => service({ url: '/report/excelReport/saveExcelTemplate', method: 'post', data })
export const bindExcelReportDataSets = (data) => service({ url: '/report/excelReport/bindExcelReportDataSets', method: 'post', data })
export const getExcelReportDataSetFields = (params) => service({ url: '/report/excelReport/getExcelReportDataSetFields', method: 'get', params })
```

### 5.3 列表页（excelReport.vue，页内 Tab 承载设计器）

1. 整页结构：`el-tabs`（v-model activeKey）——固定首个 Tab 为**列表**（GvaGrid），「设计」打开的每个报表追加一个设计器 Tab（`title=报表名-设计`，`key=reportCode`），支持多开、关闭（关当前 Tab 回列表）、**删除报表时联动移除对应 Tab**（对齐实现）；
2. GvaGrid：searchItems（keyword / reportGroup）；columns：reportCode / reportName / reportGroup / reportDesc / CreatedAt / 操作；
3. 操作列：**设计**（打开/激活设计器 Tab）/ 预览（04 交付后跳 `router.push({name:'reportExcelViewer', query:{reportCode}})`，03 阶段占位提示）/ 编辑 / 复制（el-dialog：新编码+新名称）/ 删除（confirm）；
4. 新增/编辑抽屉：reportCode（编辑禁用）/ reportName / reportGroup / reportDesc。

### 5.4 设计器主页（designer.vue）

```vue
<script setup>
  // props: reportCode（列表页 Tab 传入）；职责：加载报表 → 初始快照 → 字段列表 → 保存/绑定/导入/预览/加菜单
  // 关键状态：
  // univerRef（expose: getSnapshot/setCellValue/getSelection/rebuild）
  // initialSnapshot / univerKey（导入 xlsx 后 ++ 重建）
  // dataSetFields  [{ setCode, setName, fields: string[] }]
  // selectedCell { sheetName, row, col, value }
  // loadError（缺少编码/报表不存在 → 页签内空态文案，对齐实现）
  // 关键动作：
  // onMounted: getExcelReportList({ page:1, pageSize:1, reportCode 精确 }) → 校验存在 + 取 reportName
  //            → findExcelReport(id) → JSON.parse(jsonStr)(try-catch，失败提示并创建空白工作簿)
  //            → refreshDataSetFields()
  // handleSave: getSnapshot() → saveExcelTemplate({ reportCode, jsonStr })   // ★ 只提交 jsonStr
  // handleFieldDrop({setCode,fieldName}): getSelection() 无选区 → 插入 A1；有 → 插入选中格（§5.5）
  // handleInsertField: 无选区 → ElMessage 提示先选格；有 → 插入 + 成功提示（含行列号）
  // handleUpdateValue(value): 属性面板回写 setCellValue（闭环）
  // handleImportXlsx(file): importXlsxToSnapshot → initialSnapshot 更新 + univerKey++
  // handlePreview: router.push({ name:'reportExcelViewer', query:{ reportCode } })  // 04 交付
  // handleAddToMenu: 打开 AddToMenuModal（§5.8）
</script>
<template>
  <!-- 工具栏：{reportCode} | 导入xlsx(el-upload .xlsx) | 预览 | 保存 | 添加到菜单 -->
  <!-- splitpanes 三栏：DatasetPanel 18%（@bind 打开 DataSetSelectModal / @insert 点+插入）
       ｜ 中栏容器（@dragover.prevent @click 同步选区）内 UniverSheet 62%（:key=univerKey @cell-select @field-drop）
       ｜ PropertyPanel 20% -->
</template>
```

### 5.5 UniverSheet.vue（核心封装，实装细节）

```javascript
// ① 动态加载（模块级单例缓存，避免几 MB 进主 bundle）：
//    const [presets, corePreset, zhCNModule] = await Promise.all([
//      import('@univerjs/presets'), import('@univerjs/preset-sheets-core'),
//      import('@univerjs/preset-sheets-core/locales/zh-CN')])
//    await import('@univerjs/preset-sheets-core/lib/index.css')   // 样式随 chunk 按需加载
// ② locale 补丁：mergeLocales(zhCN, { 'sheets-ui': { info: { error:'错误',
//    forceStringInfo:'以文本形式存储的数字' } } })   // 0.25.x 语言包缺失 key，缺失时悬停告警显示 key 原文
// ③ createUniver({ locale: ZH_CN, locales, presets: [UniverSheetsCorePreset({
//      container, sheets: { disableForceStringAlert: true, disableForceStringMark: true } })]})
//    ★ sheets 相关选项必须嵌套在 preset 配置的 sheets 子对象内；关闭"强制文本数字"悬停告警与绿三角
// ④ workbook = api.createWorkbook(snapshot ?? createEmptyWorkbook())
//    readonly → toRaw(workbook).setEditable(false)   // 预览复用（04）
// ⑤ 选区监听：api.onCommandExecuted(cmd)（命令 id 为小写 'selection' 系列——SetSelections/
//    MoveSelection/SelectAll 等），取活动选区 emit('cell-select', {sheetName,row,col,value})；
//    返回 disposable，卸载时 dispose
// ⑥ 卸载：selectionDisposable.dispose() + workbook.dispose() + univer.dispose()（铁律）
// ⑦ expose：getSnapshot()=workbook.save()；setCellValue(r,c,v)=ActiveSheet().getRange(r,c,1,1).setValue(v)
//    getSelection()（拖拽/插入定位用）；快照更换由父组件 :key 重建（组件内不 watch snapshot）
// ⑧ 拖放：容器层捕获 drop 事件（@dragover.prevent），从 dataTransfer 解析 {setCode,fieldName}
//    后 emit('field-drop')，由父组件按当前选区写入
```

### 5.6 DatasetPanel.vue（左栏）+ DataSetSelectModal.vue

- **DatasetPanel**：`emit('bind')`（顶部「关联数据集」按钮 → 父组件打开弹窗）；每个已关联数据集折叠面板（头 `setName（setCode）`），字段行 `draggable`（dragstart 写 dataTransfer JSON）+ 行尾「+」按钮 `emit('insert', setCode, field)`；fields 为空提示准备结果案例（02）；
- **DataSetSelectModal**：`el-dialog`，从 `getDataSetAll()`（已启用）加载**多选**表格（回显当前已绑定 setCodes），确认 → `bindExcelReportDataSets({ reportCode, setCodes })` → emit success → 父组件 `refreshDataSetFields()`（左栏即时刷新）。

### 5.7 PropertyPanel.vue（右栏）

`el-descriptions`（Sheet 名 / A1 样式坐标 + 行列号）+ 值编辑 `el-input type=textarea`（change → emit 父组件回写）+ 占位符识别卡（正则 `^#\{([^{}]+)\.([^{}]+)}$` 宽匹配，显示数据集/字段）。

### 5.8 添加到菜单（AddToMenuModal.vue）

```
表单：上级菜单（树选择，来自菜单列表接口）/ 菜单名（默认=报表名）/ 路由 path / 排序；
提交前同父级菜单名查重（前端预检）→ 调 GVA 菜单新增 API（web/src/api/menu.js addMenu）创建：
  { parentId, name, path（顶级补 / 前缀、子级去 / 前缀）, component, sort, hidden:false }
★ GVA 适配（实现的原机制是 yudao component 字符串内嵌 query）：
  GVA 菜单 component 不支持 query → 采用「path 内嵌 query」方案：
  path = `reportpreview?reportCode={reportCode}`，并给 GVA 前端路由注册器（asyncRouter.js）
  增加 3~5 行小改造：注册路由时若 path 含 '?'，拆分为 route.path（? 之前）与注入 route 的
  query（? 之后），侧边栏跳转 URL 自然携带 query；不含 '?' 的菜单完全不受影响（向后兼容）。
  预览组件取值顺序：route.query.reportCode（兼容设计器 router.push 与菜单进入两种来源）。
成功提示：重新登录/刷新后可在侧边栏查看新菜单（菜单路由需重载生效，与 GVA 行为一致）。
```

### 5.9 xlsx 导入（utils/xlsxImport.js）

SheetJS（`cellStyles:true`）→ Univer 快照映射，覆盖：单元格值（优先格式化文本 `w`）、合并单元格、列宽（wpx/7 近似）、行列数（`!ref` range）；样式尽力迁移。导入产物 → `univerKey++` 重建 Univer（与 04 前端导出互为逆过程）。

---

## 六、菜单 / API / casbin 种子

`report_seed.go` 追加：

1. **叶子菜单** `excelReport`（父=report 目录，标题「Excel报表」，icon=document，sort=3，component=`view/report/excel/excelReport.vue`）；
2. **设计器无独立菜单**（页内 Tab 承载）；预览页 hidden 菜单（`reportExcelViewer`，path=reportpreview）归 04 种子追加；
3. API/casbin：§4.5 路由表 10 条（ApiGroup=报表平台，888）。

---

## 七、实施步骤

```
1. model 两 DO（去 status/reportAuthor，对齐实现）+ ensure_tables + request 结构 + errors 追加
2. service/excel_report.go（CRUD/复制/upsert 两路写入/bindDataSets/字段列表抛错语义/引用校验）
   + 挂接 02 DeleteDataSet 引用校验
3. api + router + 种子（叶子菜单 + 10 条 API/casbin）
4. 前端依赖安装（Univer 0.25.x pin）→ Univer 技术验证（★关键里程碑：空工作簿渲染/快照往返/
   locale 补丁/selection 命令监听/dispose 无泄漏；失败立即评估替代方案）
5. excelReport.js + 列表页（GvaGrid + el-tabs 设计器承载 + 复制弹窗）
6. designer/：designer.vue 三栏 + UniverSheet.vue（§5.5 全部实装细节）→ DatasetPanel +
   DataSetSelectModal（bind 流程）→ 拖放/点+插入 → PropertyPanel → xlsxImport.js
7. asyncRouter path 内嵌 query 小改造（§5.8）+ AddToMenuModal
8. 联调 AC1~AC12
```

---

## 八、验证清单（AC 对照）

| AC | 验收项 | 标准 |
|----|--------|------|
| AC1 | 报表 CRUD | 编码唯一；分页正常；删除级联软删模板并联动移除设计器 Tab |
| AC2 | 模板保存/还原 | 保存后重开设计器快照完整还原（含样式/合并/多 sheet）；保存仅写 json_str |
| AC3 | 复制 | 新编码+元数据+模板深拷贝一致；原报表不受影响 |
| AC4 | Univer 集成 | 编辑/合并/基础样式/多 sheet 可用；无"强制文本数字"悬停告警；反复进出无泄漏 |
| AC5 | 拖拽绑定 | 拖字段到选中格写入 `#{set.field}`；无选区拖拽默认 A1；「点+插入」无选区有提示 |
| AC6 | 字段加载 | 左栏展示关联数据集与字段；数据集被误删时字段接口显式报错 |
| AC7 | bind-datasets | 弹窗多选绑定即时生效；新建未保存模板的报表也可先绑定；绑定不动 json_str、保存不动 set_codes |
| AC8 | 导入 xlsx | .xlsx 载入（值/合并/列宽）；导入后 `:key` 重建正常 |
| AC9 | 页内 Tab | 设计器多开/关闭/删除联动；Tab 内表格无浏览器级滚动条 |
| AC10 | 值编辑闭环 | 属性面板改值回写单元格并触发快照变化 |
| AC11 | 数据集引用校验 | 删除被模板关联的数据集被拦截；前缀包含不误判（abc 不命中 xabc） |
| AC12 | 添加到菜单 | 生成的菜单进入预览页且 reportCode 正确（path 内嵌 query 适配生效） |

---

## 九、风险与注意事项

| 风险 | 等级 | 应对 |
|---|---|---|
| Univer 与本项目 Vite/Vue3.5 兼容 | 高 | 实施步骤 4 技术验证里程碑（参考实测 0.25.x 可用）；pin + lockfile；失败评估替代 |
| asyncRouter path 内嵌 query 改造 | 中 | 仅拆分含 `?` 的 path，改动 3~5 行向后兼容；AC12 专项回归存量菜单路由 |
| bind/save 两路 upsert 并发 | 低 | 均以 report_code 唯一行 upsert（先 select 后 insert/update），GVA 单用户设计场景无高并发；如需加固可改 `ON CONFLICT` |
| case_result 为空致左栏无字段 | 中 | 02 演示种子提供样例；「保存案例」维护入口列后续迭代（对齐实现现状） |
| 包体积（gzip 约 0.7~0.9MB） | 中 | preset 动态 import + 模块级缓存；设计器组件随列表页 Tab 懒渲染 |
| 拖拽落点体验 | 低 | 「先选中后拖入（无选区默认 A1）+ 点+插入（无选区提示）」双通道，对齐实装 |
| SheetJS 样式丢失 | 中 | 预期管理；与 04 前端导出对称 |

---

## 附：与 04 的衔接点

| 衔接点 | 03 状态 | 04 消费 |
|---|---|---|
| 模板 `json_str` / `set_codes`（两路独立维护） | 本期产出 | 渲染引擎输入（主数据集定位/明细行数据集收集） |
| `getExcelReportList` reportCode 精确过滤 | 本期实现 | 预览页存在性校验与名称获取（Excel 侧无 get-by-code，对齐实现） |
| UniverSheet readonly 模式 | 本期实现 | 预览页只读展示复用 |
| hidden 菜单 + path 内嵌 query 适配 | 本期建机制（asyncRouter 小改造） | 04 落地预览 hidden 菜单；「添加到菜单」直达预览 |
| 设计器/列表「预览」按钮 | 本期占位提示 | 04 替换为真实跳转 |
