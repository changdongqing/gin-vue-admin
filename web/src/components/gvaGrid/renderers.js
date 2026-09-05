// GvaGrid 单元格/查询项渲染器注册（VxeUI.renderer.add）
// 单元格钩子：renderTableDefault(renderOpts, { row, column })
// 查询项钩子：renderFormItemContent(renderOpts, { data, field })
// 重置钩子：formItemResetMethod({ data, field })
import { h, reactive } from 'vue'
import { VxeUI } from 'vxe-pc-ui'
import XEUtils from 'xe-utils'
import { ElTag, ElImage, ElButton, ElSelect, ElOption, ElDatePicker } from 'element-plus'
import { useDictionaryStore } from '@/pinia/modules/dictionary'
import { useUserStore } from '@/pinia/modules/user'
import { getUrl } from '@/utils/image'

const DATETIME_FORMAT = 'yyyy-MM-dd hh:mm:ss'

// 字典缓存：{ [type_depth]: [{ label, value }] }（reactive，异步加载完成后自动重渲染）
const dictCache = reactive({})

function loadDict(type, depth = 1) {
  const key = `${type}_${depth}`
  if (dictCache[key] !== undefined) return dictCache[key]
  dictCache[key] = []
  const dictionaryStore = useDictionaryStore()
  dictionaryStore.getDictionary(type, depth).then((list) => {
    dictCache[key] = list || []
  })
  return dictCache[key]
}

function getDictList(props) {
  return props.dict ? loadDict(props.dict, props.depth || 1) : props.options || []
}

// —— 单元格渲染器 ——

/** 日期时间格式化 */
VxeUI.renderer.add('gvaDate', {
  renderTableDefault(renderOpts, { row, column }) {
    const value = XEUtils.get(row, column.field)
    if (!value) return []
    const format = (renderOpts.props && renderOpts.props.format) || DATETIME_FORMAT
    return [h('span', XEUtils.toDateString(new Date(value), format))]
  }
})

/** 字典值 → el-tag 标签 */
VxeUI.renderer.add('gvaDict', {
  renderTableDefault(renderOpts, { row, column }) {
    const value = XEUtils.get(row, column.field)
    if (value === '' || value === null || value === undefined) return []
    const props = renderOpts.props || {}
    const list = getDictList(props)
    const hit = list.find((item) => String(item.value) === String(value))
    return [
      h(
        ElTag,
        { type: props.tagType || 'primary', size: 'small' },
        { default: () => (hit ? hit.label : value) }
      )
    ]
  }
})

/** 图片列（props: { width, height, round, preview, raw: true 时不经 getUrl 解析 }） */
VxeUI.renderer.add('gvaImage', {
  renderTableDefault(renderOpts, { row, column }) {
    const value = XEUtils.get(row, column.field)
    if (!value) return []
    const props = renderOpts.props || {}
    const src = props.raw ? value : getUrl(value)
    return [
      h(ElImage, {
        src,
        fit: 'cover',
        lazy: true,
        previewSrcList: props.preview === false ? undefined : [src],
        previewTeleported: true,
        style: {
          width: `${props.width || 40}px`,
          height: `${props.height || 40}px`,
          borderRadius: props.round ? '50%' : '4px',
          verticalAlign: 'middle',
          display: 'block'
        }
      })
    ]
  }
})

// v-auth 指令等价逻辑：button.auth 为允许的 authorityId 数组
function filterAuthButtons(buttons) {
  const userStore = useUserStore()
  const authorityId = userStore.userInfo && userStore.userInfo.authorityId
  return buttons.filter((btn) => {
    if (!btn.auth || !btn.auth.length) return true
    return btn.auth.some((item) => Number(item) === Number(authorityId))
  })
}

/** 操作列按钮组（含权限按钮的页面建议改用插槽 + v-auth） */
VxeUI.renderer.add('gvaOperate', {
  renderTableDefault(renderOpts, { row, column }) {
    const props = renderOpts.props || {}
    const buttons = filterAuthButtons(props.buttons || [])
    return [
      h(
        'div',
        { style: { display: 'flex', flexWrap: 'wrap', gap: '8px' } },
        buttons.map((btn) =>
          h(
            ElButton,
            {
              type: btn.type || 'primary',
              link: btn.link !== false,
              icon: btn.icon,
              onClick: () => btn.onClick && btn.onClick(row, column)
            },
            { default: () => btn.label }
          )
        )
      )
    ]
  }
})

// —— 查询项渲染器 ——

/** 字典下拉（props: { dict, depth, placeholder, clearable, options }） */
VxeUI.renderer.add('gvaDictSelect', {
  renderFormItemContent(renderOpts, { data, field }) {
    const props = renderOpts.props || {}
    const list = getDictList(props)
    return [
      h(
        ElSelect,
        {
          modelValue: data[field],
          'onUpdate:modelValue': (value) => {
            data[field] = value
          },
          placeholder: props.placeholder || '请选择',
          clearable: props.clearable !== false,
          style: { width: '100%' }
        },
        {
          default: () =>
            list.map((item) =>
              h(ElOption, { key: String(item.value), label: item.label, value: item.value })
            )
        }
      )
    ]
  },
  formItemResetMethod({ data, field }) {
    data[field] = undefined
  }
})

/**
 * 日期范围（值存 data[field] = [start, end]）
 * 若配置 props.startField/props.endField，在 beforeQuery 中拆分为独立查询字段
 */
VxeUI.renderer.add('gvaDateRange', {
  renderFormItemContent(renderOpts, { data, field }) {
    const props = renderOpts.props || {}
    return [
      h(ElDatePicker, {
        modelValue: data[field],
        'onUpdate:modelValue': (value) => {
          data[field] = value
        },
        type: props.type || 'daterange',
        valueFormat: props.valueFormat || 'YYYY-MM-DD',
        startPlaceholder: '开始日期',
        endPlaceholder: '结束日期',
        style: { width: '100%' }
      })
    ]
  },
  formItemResetMethod({ data, field }) {
    data[field] = undefined
  }
})

export function setupGvaGridRenderers() {
  // renderer.add 为模块加载时全局注册，此处保留安装函数形式便于统一初始化
}
