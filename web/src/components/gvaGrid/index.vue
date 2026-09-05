<template>
  <div
    ref="wrapRef"
    class="gva-grid"
    :class="{ 'gva-grid--fit': isFit, 'gva-grid--fallback': isFit && !fitHeight }"
    :style="wrapStyle"
  >
    <vxe-grid ref="gridRef" v-bind="gridBind">
      <!-- 查询/重置按钮（Element 视觉，语义对齐旧页面 onSubmit/onReset） -->
      <template #gva-form-buttons>
        <el-button type="primary" icon="search" @click="submitSearch">查询</el-button>
        <el-button icon="refresh" @click="submitReset">重置</el-button>
        <slot name="form-buttons-extra" />
      </template>

      <!-- 折叠触发节点（展开/收起） -->
      <template #gva-form-collapse>
        <el-button link type="primary" @click="toggleCollapse">
          {{ collapsed ? '展开' : '收起' }}
          <el-icon class="gva-grid__collapse-icon" :class="{ 'is-collapsed': collapsed }">
            <arrow-down />
          </el-icon>
        </el-button>
      </template>

      <!-- 工具栏右侧：密度切换 + 页面自定义工具按钮（toolbar-tools） -->
      <template #gva-grid-tools>
        <el-dropdown trigger="click" @command="onSizeCommand">
          <span class="gva-grid__tool-btn" title="密度">
            <el-icon><Menu /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item
                v-for="opt in sizeOptions"
                :key="opt.value"
                :command="opt.value"
                :class="{ 'is-active': currentSize === opt.value }"
              >
                {{ opt.label }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <slot name="toolbar-tools" />
      </template>

      <!-- 全量透传页面插槽（单元格/工具栏按钮/空态等） -->
      <template v-for="(_, name) in passSlots" :key="name" #[name]="scope">
        <slot :name="name" v-bind="scope ?? {}" />
      </template>
    </vxe-grid>
  </div>
</template>

<script setup>
import { ref, computed, useSlots, useAttrs, inject } from 'vue'
import { ArrowDown, Menu } from '@element-plus/icons-vue'
import { useGridHeight } from './hooks/useGridHeight'
import { GVA_SIZE_OPTIONS } from './config/defaults'
import { GVA_GRID_KEY } from './config/constants'

defineOptions({ name: 'GvaGrid' })

const props = defineProps({
  /** 'fit' 测量填充（列表页默认）| number px 固定（弹窗/抽屉内）| 'auto' 填充有确定高度的父级 */
  height: { type: [String, Number], default: 'fit' },
  /** fit 模式底部预留（BottomInfo 高度 + 容器 padding） */
  bottomReserve: { type: Number, default: 56 },
  /** 高度测量参照的滚动容器选择器 */
  fitContainerSelector: { type: String, default: '.gva-container, .gva-container2' }
})

const emit = defineEmits(['search', 'reset'])
const attrs = useAttrs()
const slots = useSlots()

// useGvaGrid 注入的查询/重置控制器（未经过 hook 使用时退化为仅发事件）
const gridController = inject(GVA_GRID_KEY, null)

// —— 高度自适应 ——
const isFit = computed(() => props.height === 'fit')
const { wrapRef, height: fitHeight } = useGridHeight({
  selector: props.fitContainerSelector,
  bottomReserve: props.bottomReserve
})

const wrapStyle = computed(() => {
  if (isFit.value && fitHeight.value) return { height: `${fitHeight.value}px` }
  // 'auto' 模式：填充已有确定高度的父级
  if (props.height === 'auto') return { height: '100%' }
  return {}
})
const gridHeight = computed(() => {
  if (isFit.value) {
    // 测量成功：.gva-grid 定高 + vxe height auto（填充）；测量失败：交由降级样式 max-height
    return fitHeight.value ? 'auto' : undefined
  }
  return props.height
})

// —— 查询/重置/折叠 ——
// vxe formConfig 语义：collapseStatus=true 表示折叠（folding 项隐藏）
const collapsed = ref(true)
const formConfig = computed(() => attrs.formConfig || null)
const hasCollapseNode = computed(
  () =>
    formConfig.value &&
    formConfig.value.items &&
    formConfig.value.items.some((item) => item.collapseNode)
)

const gridBind = computed(() => ({
  ...attrs,
  height: gridHeight.value,
  size: currentSize.value,
  formConfig: formConfig.value
    ? {
        ...formConfig.value,
        collapseStatus: hasCollapseNode.value ? collapsed.value : formConfig.value.collapseStatus
      }
    : undefined
}))

function toggleCollapse() {
  collapsed.value = !collapsed.value
}

function submitSearch() {
  if (gridController && gridController.submitSearch) {
    gridController.submitSearch()
  }
  emit('search')
}

function submitReset() {
  collapsed.value = true
  if (gridController && gridController.resetSearch) {
    gridController.resetSearch()
  }
  emit('reset')
}

// —— 密度切换 ——
const currentSize = ref('medium')
const sizeOptions = GVA_SIZE_OPTIONS
function onSizeCommand(size) {
  currentSize.value = size
}

// —— 插槽透传（排除内部使用的插槽名） ——
const RESERVED_SLOTS = new Set(['gva-form-buttons', 'gva-form-collapse', 'gva-grid-tools', 'form-buttons-extra'])
const passSlots = computed(() => {
  const names = {}
  Object.keys(slots).forEach((name) => {
    if (!RESERVED_SLOTS.has(name)) names[name] = true
  })
  return names
})

// —— vxe-grid 实例方法委托（页面通过 useGvaGrid 的 gridRef 直调） ——
const gridRef = ref(null)
const delegatedMethods = [
  'commitProxy',
  'getFormData',
  'getCheckboxRecords',
  'clearCheckboxRow',
  'clearSort',
  'getCurrentPage',
  'setCurrentPage',
  'zoom',
  'getCheckedFiltersData'
]
const exposed = {
  /** vxe-grid 原始实例（白名单外的全量方法） */
  get $grid() {
    return gridRef.value
  }
}
delegatedMethods.forEach((method) => {
  exposed[method] = function (...args) {
    return gridRef.value && gridRef.value[method](...args)
  }
})
defineExpose(exposed)
</script>
