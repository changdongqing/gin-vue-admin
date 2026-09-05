// GvaGrid 删除/批量删除：确认框 → 接口 → 成功提示 → 刷新 → 清空勾选
// 与旧页面 deleteUserFunc 语义对齐（文案、type: warning、成功后 getTableData）
import { ElMessage, ElMessageBox } from 'element-plus'

/**
 * @param {Object} gridRef useGvaGrid 返回的 gridRef（用于刷新与清空勾选）
 * @param {Object} opts
 *  - delete       必填 (rows) => Promise<{code}>，rows 为行数组（单行也包装为数组）
 *  - confirmText  确认文案，默认与现有页面一致
 *  - successText  成功提示：字符串，或 (res, rows) => string（如使用服务端返回的 msg）
 *  - reload       刷新方式：'refresh' 保持当前页（默认，对齐旧页面）| 'reload' 回第 1 页
 */
export function useGvaGridDelete(gridRef, opts = {}) {
  const {
    delete: deleteApi,
    confirmText = '此操作将永久删除该数据, 是否继续',
    successText = '删除成功',
    // 与旧页面删除后 getTableData() 语义一致：保持当前页（末页删空由空页回退兜底）
    reload: reloadMode = 'refresh'
  } = opts

  if (!deleteApi) {
    throw new Error('[GvaGrid] useGvaGridDelete 需要配置 delete 接口')
  }

  async function onDelete(rows) {
    const rowList = Array.isArray(rows) ? rows : [rows]
    if (!rowList.length) return false
    try {
      await ElMessageBox.confirm(confirmText, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch (e) {
      return false
    }
    const res = await deleteApi(rowList)
    if (res && res.code === 0) {
      const text = typeof successText === 'function' ? successText(res, rowList) : successText
      ElMessage({
        message: text,
        type: 'success'
      })
      const $grid = gridRef.value
      if ($grid) {
        if (reloadMode === 'refresh') {
          $grid.commitProxy('query')
        } else {
          $grid.commitProxy('reload')
        }
        if ($grid.clearCheckboxRow) {
          $grid.clearCheckboxRow()
        }
      }
      return true
    }
    // 失败提示已由 utils/request.js 响应拦截统一处理
    return false
  }

  /** 单行删除（模板 @click="onDelete(row)"） */
  function deleteRow(row) {
    return onDelete(row)
  }

  /** 批量删除（模板 @click="onBatchDelete(selectedRows)"） */
  function deleteRows(rows) {
    return onDelete(rows)
  }

  return { onDelete, deleteRow, deleteRows }
}

export default useGvaGridDelete
