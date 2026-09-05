// GvaGrid 核心 hook：页面配置 → vxe-grid 的 gridOptions/gridEvents
// 用法：
//   const { gridRef, gridOptions, gridEvents, selectedRows, reload, refresh } = useGvaGrid({ api, columns, searchItems, ... })
//   <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents"> ... </GvaGrid>
import { ref, computed, provide } from 'vue'
import {
  GVA_GRID_DEFAULTS,
  GVA_TOOLBAR_DEFAULTS
} from '../config/defaults'
import { GVA_GRID_KEY } from '../config/constants'
import { normalizeColumns, buildFormConfig } from '../config/schema'
import { buildProxyConfig, buildPagerConfig } from './useGvaGridProxy'

/**
 * @param {Object} options
 *  - id               列设置持久化 key（建议 `模块-页面`）
 *  - api              列表接口 (params) => Promise<{code, data:{list,total}}>
 *  - columns          列 schema（GVA 列 schema）
 *  - searchItems      查询项 schema
 *  - searchDefaults   查询初始/重置默认值
 *  - searchCollapse   超过 N 项自动折叠
 *  - defaultSort      默认排序 { field: 'ID', order: 'desc' }；传 null 关闭默认排序
 *  - checkbox         是否启用多选列（跨页保留勾选）
 *  - pageSize/pageSizes/autoLoad/emptyPageFallback
 *  - operateWidth     操作列宽度（默认 appStore.operateMinWith）
 *  - gridConfig       原样 merge 进 vxe-grid 配置（逃生舱）
 */
export function useGvaGrid(options = {}) {
  const gridRef = ref(null)
  const selectedRows = ref([])

  const normalizedColumns = normalizeColumns(options.columns || [], {
    checkbox: options.checkbox,
    operateWidth: options.operateWidth
  })

  const formConfig = buildFormConfig(options.searchItems, {
    searchDefaults: options.searchDefaults,
    searchCollapse: options.searchCollapse
  })

  const gridOptions = {
    ...GVA_GRID_DEFAULTS,
    id: options.id,
    columns: normalizedColumns,
    proxyConfig: buildProxyConfig({ ...options, searchItems: options.searchItems }),
    // pager: false → 无分页（全量数据，如菜单树）
    pagerConfig: options.pager === false ? null : buildPagerConfig(options),
    formConfig,
    toolbarConfig: options.toolbarConfig === null
      ? null
      : { ...GVA_TOOLBAR_DEFAULTS, ...(options.toolbarConfig || {}) },
    ...(options.gridConfig || {})
  }

  const gridEvents = {
    checkboxChange() {
      syncSelection()
    },
    checkboxAll() {
      syncSelection()
    },
    // 每次代理查询后重同步（编程式 clearCheckboxRow 不触发 checkbox 事件）
    proxyQuery() {
      syncSelection()
    },
    ...(options.gridEvents || {})
  }

  function syncSelection() {
    const $grid = gridRef.value
    if ($grid && $grid.getCheckboxRecords) {
      selectedRows.value = $grid.getCheckboxRecords()
    }
  }

  /** 回到第 1 页重新查询（新增/删除后使用） */
  function reload() {
    return gridRef.value ? gridRef.value.commitProxy('reload') : Promise.resolve()
  }

  /** 保持当前页/条件重新查询（编辑保存后使用） */
  function refresh() {
    return gridRef.value ? gridRef.value.commitProxy('query') : Promise.resolve()
  }

  function getFormData() {
    return gridRef.value ? gridRef.value.getFormData() : {}
  }

  /** 读写查询条件（编程式查询） */
  function setFormData(patch = {}) {
    const $grid = gridRef.value
    if (!$grid) return
    const data = $grid.getFormData()
    Object.keys(patch).forEach((key) => {
      data[key] = patch[key]
    })
  }

  function clearSelection() {
    const $grid = gridRef.value
    if ($grid && $grid.clearCheckboxRow) {
      $grid.clearCheckboxRow()
    }
    selectedRows.value = []
    syncSelection()
  }

  /** 查询（回到第 1 页）；由 GvaGrid 内部查询按钮触发 */
  function submitSearch() {
    return reload()
  }

  /** 重置：清空查询条件、恢复默认值、清除排序并回到第 1 页（对齐旧页面 onReset 语义） */
  function resetSearch() {
    const $grid = gridRef.value
    if (!$grid) return
    const data = $grid.getFormData()
    const defaults = options.searchDefaults || {}
    Object.keys(data).forEach((key) => {
      data[key] = defaults[key] !== undefined ? defaults[key] : ''
    })
    if ($grid.clearSort) {
      $grid.clearSort()
    }
    return reload()
  }

  // —— 增删改工具配合 ——

  const formValues = computed(() => ({ ...getFormData() }))

  // 注入给 GvaGrid：内部查询/重置按钮自动调用（页面无需再接 @search/@reset）
  provide(GVA_GRID_KEY, {
    submitSearch,
    resetSearch
  })

  return {
    gridRef,
    gridOptions,
    gridEvents,
    selectedRows,
    reload,
    refresh,
    getFormData,
    setFormData,
    clearSelection,
    submitSearch,
    resetSearch,
    formValues
  }
}

export default useGvaGrid
