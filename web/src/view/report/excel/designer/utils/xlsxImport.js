/**
 * xlsx → Univer 快照映射（SheetJS 社区版，cellStyles:true）
 * 覆盖：单元格值（优先格式化文本 w）、合并单元格、列宽（wpx/7 近似）、行列数（!ref）
 * 样式尽力迁移为后续迭代；与 04 前端导出互为逆过程
 */
import * as XLSX from 'xlsx'

const colName = (index) => {
  let s = ''
  let n = index + 1
  while (n > 0) {
    const rem = (n - 1) % 26
    s = String.fromCharCode(65 + rem) + s
    n = Math.floor((n - 1) / 26)
  }
  return s
}

export const createEmptySnapshot = () => ({
  id: 'workbook-01',
  locale: 'zh-CN',
  name: '工作簿',
  sheetOrder: ['sheet-01'],
  sheets: {
    'sheet-01': {
      id: 'sheet-01',
      name: 'Sheet1',
      tabColor: '',
      hidden: false,
      rowCount: 40,
      columnCount: 20,
      zoomRatio: 1,
      cellData: {},
      rowData: {},
      columnData: {},
      mergeData: []
    }
  },
  views: [{}]
})

export const importXlsxToSnapshot = async (file) => {
  const buffer = await file.arrayBuffer()
  const wb = XLSX.read(buffer, { type: 'array', cellStyles: true, cellDates: false })
  const snapshot = createEmptySnapshot()
  snapshot.sheets = {}
  const sheetOrder = []

  wb.SheetNames.forEach((name, sheetIdx) => {
    const ws = wb.Sheets[name]
    const sheetId = `sheet-${String(sheetIdx + 1).padStart(2, '0')}`
    sheetOrder.push(sheetId)

    const range = ws['!ref'] ? XLSX.utils.decode_range(ws['!ref']) : { s: { r: 0, c: 0 }, e: { r: 0, c: 0 } }
    const rowCount = Math.max(range.e.r + 1, 20)
    const columnCount = Math.max(range.e.c + 1, 10)

    const cellData = {}
    for (let r = range.s.r; r <= range.e.r; r++) {
      for (let c = range.s.c; c <= range.e.c; c++) {
        const addr = XLSX.utils.encode_cell({ r, c })
        const cell = ws[addr]
        if (!cell) continue
        const out = {}
        // 优先格式化文本，其次原始值
        if (cell.w !== undefined) {
          out.v = cell.w
          out.t = 1 // CellValueType.String
        } else if (cell.v !== undefined) {
          if (typeof cell.v === 'number') {
            out.v = cell.v
            out.t = 3 // CellValueType.Number
          } else {
            out.v = String(cell.v)
            out.t = 1
          }
        }
        if (Object.keys(out).length) {
          cellData[r] = cellData[r] || {}
          cellData[r][c] = out
        }
      }
    }

    // 合并单元格
    const mergeData = (ws['!merges'] || []).map((m) => ({
      startRow: m.s.r,
      endRow: m.e.r,
      startColumn: m.s.c,
      endColumn: m.e.c
    }))

    // 列宽（wpx 像素 → Univer 列宽近似 1:1；w 字符宽 ×7）
    const columnData = {}
    ;(ws['!cols'] || []).forEach((col, c) => {
      const width = col?.wpx ?? (col?.w ? Math.round(col.w * 7) : undefined)
      if (width) {
        columnData[c] = { w: width, hd: 0 }
      }
    })

    snapshot.sheets[sheetId] = {
      id: sheetId,
      name,
      tabColor: '',
      hidden: false,
      rowCount,
      columnCount,
      zoomRatio: 1,
      cellData,
      rowData: {},
      columnData,
      mergeData
    }
  })
  snapshot.sheetOrder = sheetOrder
  return snapshot
}

export { colName }
