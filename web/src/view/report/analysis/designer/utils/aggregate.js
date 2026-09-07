/**
 * 前端聚合层：透视模式预聚合（S2 2.x 对相同维度组合多行是"后行覆盖"，不聚合——
 * 必须先在本层产出"每维度组合恰好一行"，S2 的索引/小计/总计/分页语义才正确）。
 * COUNT 计非空样本数；NONE 直显（重复组合取首条，等值语义）；
 * SUM/AVG/MIN/MAX 空值按非空样本计算（AVG 分母为非空样本数）。
 */

/** 数值列派生键：同字段可配多种聚合不冲突 */
export const valueKeyOf = (v) => `${v.field}__${v.aggregation}`

const groupKeyOf = (row, dims) => dims.map((d) => String(row[d] ?? '')).join('')

const toNumber = (v) => {
  if (v === null || v === undefined || v === '') return null
  const n = Number(v)
  return Number.isNaN(n) ? null : n
}

/**
 * 透视模式预聚合
 * @param {import('@/api/report/analysisReport').ReportConfig} config
 * @param {Array<Object>} rawRows 后端明细行
 * @returns {{ rows: Array<Object>, valueKeys: string[] }} 聚合行（含派生键列）与派生键数组
 */
export function aggregateData(config, rawRows) {
  const dims = [...(config.fields.rows || []), ...(config.fields.columns || [])]
  const values = config.fields.values || []
  const valueKeys = values.map(valueKeyOf)

  // 无维度：聚合全量为单行
  if (!dims.length) {
    const row = {}
    for (const v of values) {
      row[valueKeyOf(v)] = aggregateValues(v, rawRows.map((r) => r[v.field]))
    }
    return { rows: [row], valueKeys }
  }

  const groups = new Map()
  for (const row of rawRows || []) {
    const key = groupKeyOf(row, dims)
    let bucket = groups.get(key)
    if (!bucket) {
      bucket = { first: row, rows: [] }
      groups.set(key, bucket)
    }
    bucket.rows.push(row)
  }

  const out = []
  for (const bucket of groups.values()) {
    const row = {}
    for (const d of dims) {
      row[d] = bucket.first[d]
    }
    for (const v of values) {
      if (v.aggregation === 'NONE') {
        // 直显：取首条（等值语义）
        row[valueKeyOf(v)] = bucket.rows[0]?.[v.field] ?? null
      } else {
        row[valueKeyOf(v)] = aggregateValues(v, bucket.rows.map((r) => r[v.field]))
      }
    }
    out.push(row)
  }
  return { rows: out, valueKeys }
}

/**
 * 单值聚合计算
 */
function aggregateValues(valueCfg, vals) {
  const { aggregation } = valueCfg
  switch (aggregation) {
    case 'COUNT':
      return vals.filter((v) => v !== null && v !== undefined && v !== '').length
    case 'NONE':
      return vals.length ? vals[0] ?? null : null
    default: {
      const nums = (vals || []).map(toNumber).filter((n) => n !== null)
      if (!nums.length) return null
      switch (aggregation) {
        case 'SUM':
          return nums.reduce((a, b) => a + b, 0)
        case 'AVG':
          return nums.reduce((a, b) => a + b, 0) / nums.length
        case 'MIN':
          return Math.min(...nums)
        case 'MAX':
          return Math.max(...nums)
        default:
          return null
      }
    }
  }
}
