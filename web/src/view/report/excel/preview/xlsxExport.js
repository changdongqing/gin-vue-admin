/**
 * exportSnapshotToXlsx：渲染快照 → .xlsx（SheetJS 社区版；与 03 xlsxImport 互为逆过程）
 * 覆盖：值（v ?? m）、合并单元格、列宽（w × 1.2 近似字符宽）；不迁移样式（社区版能力边界）
 */
import * as XLSX from 'xlsx'

const colNumToName = (index) => {
  let s = ''
  let n = Math.floor(index) + 1
  while (n > 0) {
    const rem = (n - 1) % 26
    s = String.fromCharCode(65 + rem) + s
    n = Math.floor((n - 1) / 26)
  }
  return s
}

export const exportSnapshotToXlsx = (snapshot, filename) => {
  const workbook = XLSX.utils.book_new()
  const sheets = snapshot?.sheets || {}
  const order = Array.isArray(snapshot?.sheetOrder) ? snapshot.sheetOrder : Object.keys(sheets)

  order.forEach((sheetId) => {
    const sheet = sheets[sheetId]
    if (!sheet) return
    const ws = {}
    const cellData = sheet.cellData || {}
    const merges = sheet.mergeData || []

    // 真实行列范围（含合并端点）
    let maxRow = -1
    let maxCol = -1
    Object.keys(cellData).forEach((r) => {
      const rowNum = Number(r)
      if (rowNum > maxRow) maxRow = rowNum
      const row = cellData[r] || {}
      Object.keys(row).forEach((c) => {
        const colNum = Number(c)
        if (colNum > maxCol) maxCol = colNum
        const cell = row[c]
        if (!cell) return
        const value = cell.v ?? cell.m ?? ''
        if (value === '') return
        ws[`${colNumToName(colNum)}${rowNum + 1}`] = { v: value }
      })
    })
    merges.forEach((m) => {
      if ((m.endRow ?? -1) > maxRow) maxRow = Number(m.endRow)
      if ((m.endColumn ?? -1) > maxCol) maxCol = Number(m.endColumn)
    })

    // 合并区域（s/e 0-based → XLSX 0-based）
    if (merges.length) {
      ws['!merges'] = merges.map((m) => ({
        s: { r: m.startRow, c: m.startColumn },
        e: { r: m.endRow, c: m.endColumn }
      }))
    }

    // 列宽（Univer 列宽像素 → SheetJS 字符宽近似 ÷1.2 → 设计口径 w×1.2 的逆过程）
    const columnData = sheet.columnData || {}
    const colKeys = Object.keys(columnData)
    if (colKeys.length) {
      ws['!cols'] = colKeys.map(() => undefined)
      ws['!cols'] = []
      let prev = 0
      const colList = []
      colKeys
        .map(Number)
        .sort((a, b) => a - b)
        .forEach((c) => {
          // 填充缺失列的默认宽
          for (let i = prev; i < c; i++) colList.push({ wch: 8.43 })
          const w = columnData[c]?.w
          colList.push({ wch: w ? Math.max(4, Math.round(w / 1.2 * 10) / 10) : 8.43 })
          prev = c + 1
        })
      ws['!cols'] = colList
    }

    // !ref：真实范围；空表用模板 rowCount/columnCount 兜底
    if (maxRow < 0 || maxCol < 0) {
      maxRow = Math.max((sheet.rowCount || 1) - 1, 0)
      maxCol = Math.max((sheet.columnCount || 1) - 1, 0)
    }
    ws['!ref'] = XLSX.utils.encode_range({ s: { r: 0, c: 0 }, e: { r: maxRow, c: Math.max(maxCol, 0) } })

    XLSX.utils.book_append_sheet(workbook, ws, sheet.name || `Sheet${order.indexOf(sheetId) + 1}`)
  })

  XLSX.writeFile(workbook, filename)
}
