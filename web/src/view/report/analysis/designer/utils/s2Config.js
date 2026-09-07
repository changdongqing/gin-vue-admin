/**
 * ReportConfig ↔ S2 dataCfg/options/themeCfg 映射层（纯函数，便于控制台用例验证）。
 * 明细模式：columns = rows 维度顺序 + values 字段，直显不聚合；
 * 透视模式：先经 aggregate.js 预聚合，S2 values 为派生键 field__aggregation。
 */
import { aggregateData, valueKeyOf } from './aggregate'

/** 轻量 numeral 风格格式化（#,##0.00 / 0.00% / 0.00 / #,##0） */
export function formatNumber(value, format) {
  if (value === null || value === undefined || value === '') return ''
  const n = Number(value)
  if (Number.isNaN(n)) return String(value)
  const fmt = format || ''
  if (fmt.includes('%')) {
    const decimals = (fmt.split('.')[1] || '').replace('%', '').length
    return `${(n * 100).toFixed(decimals)}%`
  }
  const decimals = fmt.includes('.') ? fmt.split('.')[1].length : 0
  const thousands = fmt.includes(',')
  let out = n.toFixed(decimals)
  if (thousands) {
    const [int, dec] = out.split('.')
    out = int.replace(/\B(?=(\d{3})+(?!\d))/g, ',') + (dec ? `.${dec}` : '')
  }
  return out
}

/** 派生键元信息（alias/formatter 挂 meta） */
const buildMeta = (values, isPivot) =>
  (values || []).map((v) => ({
    field: isPivot ? valueKeyOf(v) : v.field,
    name: v.alias || `${v.field}${v.aggregation && v.aggregation !== 'NONE' ? `(${v.aggregation})` : ''}`,
    formatter: (cell) => (v.format ? formatNumber(cell, v.format) : cell ?? '')
  }))

/** 构造 S2 dataCfg */
export function buildDataCfg(config, rawRows) {
  if (config.sheetType === 'table') {
    // 明细模式：columns = rows 维度顺序 + values 字段，直显不聚合
    const columns = [...(config.fields.rows || []), ...(config.fields.values || []).map((v) => v.field)]
    const seen = new Set()
    return {
      fields: {
        columns: columns.filter((c) => (seen.has(c) ? false : (seen.add(c), true)))
      },
      data: rawRows || [],
      meta: buildMeta(config.fields.values, false)
    }
  }
  const { rows, valueKeys } = aggregateData(config, rawRows || [])
  return {
    fields: {
      rows: [...(config.fields.rows || [])],
      columns: [...(config.fields.columns || [])],
      values: valueKeys,
      valueInCols: !config.fields.valueInRow // 取反映射（用户语义 valueInRow）
    },
    data: rows,
    meta: buildMeta(config.fields.values, true)
  }
}

/** 构造 S2 options */
export function buildOptions(config) {
  const isTable = config.sheetType === 'table'
  const options = {
    width: undefined,
    height: undefined,
    hierarchyType: isTable ? 'grid' : (config.options.hierarchyType || 'grid'),
    layoutWidthType: config.options.layoutWidthType || 'adaptive',
    frozenRowHeader: isTable ? false : !!config.options.frozenRowHeader,
    showSeriesNumber: !!config.options.showSeriesNumber,
    interaction: { autoResetSheetStyle: false },
    pagination: config.options.pagination?.open
      ? {
          open: true,
          current: config.options.pagination.current || 1,
          pageSize: config.options.pagination.pageSize || 50
        }
      : { open: false }
  }
  if (!isTable && config.options.totals) {
    // 维度级全局 SUM（AVG 列为"平均的平均"语义限制，界面有提示）
    const buildTotals = (t) => ({
      showGrandTotals: !!t?.showGrandTotals,
      showSubTotals: !!t?.showSubTotals,
      subTotalsDimensions: t?.subTotalsDimensions || [],
      calcGrandTotals: { aggregation: 'SUM' },
      calcSubTotals: { aggregation: 'SUM' }
    })
    options.totals = { row: buildTotals(config.options.totals.row), col: buildTotals(config.options.totals.col) }
  }
  if (config.options.style?.rowHeight) options.style = { ...options.style, rowHeight: config.options.style.rowHeight }
  if (config.options.style?.colWidth) options.style = { ...(options.style || {}), colWidth: config.options.style.colWidth }
  return options
}

/** 构造 themeCfg */
export function buildThemeCfg(config) {
  return { name: config.theme?.name || 'default' }
}

/** 行小计维度 = rows 去最外层 */
export function calcSubTotalsDimensions(rows) {
  return (rows || []).slice(1)
}

/** 默认配置（§6.1） */
export function defaultConfig() {
  return {
    sheetType: 'pivot',
    fields: { rows: [], columns: [], values: [], valueInRow: false },
    options: {
      pagination: { open: false, current: 1, pageSize: 50 },
      frozenRowHeader: true,
      showSeriesNumber: false,
      hierarchyType: 'grid',
      layoutWidthType: 'adaptive',
      adaptive: true,
      totals: {
        row: { showGrandTotals: false, showSubTotals: false, subTotalsDimensions: [] },
        col: { showGrandTotals: false, showSubTotals: false, subTotalsDimensions: [] }
      },
      style: {}
    },
    theme: { name: 'default' }
  }
}

/** 加载配置：JSON.parse(try-catch) + 与 defaultConfig 逐层合并兜底（新字段带默认值不破坏旧配置） */
export function loadConfig(jsonStr) {
  const base = defaultConfig()
  if (!jsonStr) return base
  let parsed
  try {
    parsed = JSON.parse(jsonStr)
  } catch (e) {
    return base
  }
  return {
    ...base,
    ...parsed,
    fields: { ...base.fields, ...(parsed.fields || {}) },
    options: {
      ...base.options,
      ...(parsed.options || {}),
      pagination: { ...base.options.pagination, ...(parsed.options?.pagination || {}) },
      totals: {
        row: { ...base.options.totals.row, ...(parsed.options?.totals?.row || {}) },
        col: { ...base.options.totals.col, ...(parsed.options?.totals?.col || {}) }
      },
      style: { ...base.options.style, ...(parsed.options?.style || {}) }
    },
    theme: { ...base.theme, ...(parsed.theme || {}) }
  }
}
