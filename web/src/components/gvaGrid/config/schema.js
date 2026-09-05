// GvaGrid 列 schema 与查询项 schema 归一化
// GVA 列 schema ≈ vxe 列 schema 的子集 + 少量约定：
//  - 操作列不写 field（归一化时补 'operate'，vxe customConfig.storage 要求列必须有 field）
//  - 操作列宽度默认取 appStore.operateMinWith
import { useAppStore } from '@/pinia'

// 查询/重置按钮组：由 GvaGrid 内部插槽渲染（Element 按钮，视觉与现有一致），
// 这里只声明一个使用该插槽的 form item
export const FORM_BUTTONS_SLOT = 'gva-form-buttons'
export const FORM_COLLAPSE_SLOT = 'gva-form-collapse'

/**
 * 归一化列 schema
 * @param {Array} columns 页面声明的列
 * @param {Object} options { checkbox, operateWidth }
 */
export function normalizeColumns(columns = [], options = {}) {
  const appStore = useAppStore()
  const result = []

  if (options.checkbox) {
    result.push({ type: 'checkbox', width: 50, align: 'center' })
  }

  columns.forEach((col) => {
    const item = { ...col }
    // 操作列缺省命名（vxe customConfig.storage 要求列必须有 field）
    if (
      !item.field &&
      !item.type &&
      item.slots &&
      item.slots.default === 'operate'
    ) {
      item.field = 'operate'
    }
    if (item.slots && item.slots.default === 'operate' && !item.width) {
      item.width = Number(options.operateWidth || appStore.operateMinWith || 240)
    }
    result.push(item)
  })

  return result
}

/**
 * 构建查询区 formConfig
 * @param {Array} searchItems 查询项 schema（空数组/null 表示无查询区）
 * @param {Object} options { searchDefaults, searchCollapse }
 */
export function buildFormConfig(searchItems, options = {}) {
  if (!searchItems || !searchItems.length) {
    return null
  }
  const collapseCount = Number(options.searchCollapse || 0)
  const items = searchItems.map((raw) => {
    const item = { span: 6, titleWidth: 'auto', ...raw }
    if (collapseCount > 0) {
      item.folding = true
    }
    return item
  })

  if (collapseCount > 0 && searchItems.length > collapseCount) {
    // 前 collapseCount 项始终可见，其余折叠
    items.forEach((item, index) => {
      if (index < collapseCount) item.folding = false
    })
    // 折叠触发节点 + 查询/重置按钮（内容由 GvaGrid 内部插槽渲染）
    items.push({
      span: 24,
      align: 'left',
      collapseNode: true,
      slots: { default: FORM_COLLAPSE_SLOT }
    })
    items.push({
      span: 24,
      align: 'left',
      slots: { default: FORM_BUTTONS_SLOT }
    })
  } else {
    items.push({
      span: 24,
      align: 'left',
      slots: { default: FORM_BUTTONS_SLOT }
    })
  }

  return {
    data: { ...(options.searchDefaults || {}) },
    titleWidth: 80,
    titleAsterisk: false,
    vertical: false,
    // 有折叠时默认收起
    collapseStatus: collapseCount > 0 && searchItems.length > collapseCount ? false : undefined,
    items
  }
}
