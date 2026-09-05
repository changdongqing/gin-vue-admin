// GvaGrid 自适应高度（fit 模式）
// 页面滚动发生在 layout 中间层 .gva-container/.gva-container2（有确定高度），
// 而页面内容层 .gva-body-h 只有 min-height，百分比高度链断裂，
// 因此组件高度 = 滚动容器高度 − 组件顶边到容器顶边的占位 − 底部预留。
import { ref, onMounted, onActivated, onBeforeUnmount, nextTick } from 'vue'

const DEFAULT_CONTAINER_SELECTOR = '.gva-container, .gva-container2'
const TRANSITION_RECHECK_DELAY = 260
const MIN_HEIGHT = 240

export function useGridHeight({ selector, bottomReserve = 56, min = MIN_HEIGHT } = {}) {
  const wrapRef = ref(null)
  // 0 表示测量不可用（触发降级模式）
  const height = ref(0)
  let resizeObserver = null
  let rafId = null
  let recheckTimer = null

  const compute = () => {
    if (rafId) cancelAnimationFrame(rafId)
    rafId = requestAnimationFrame(() => {
      rafId = null
      const el = wrapRef.value
      if (!el) return
      const container = el.closest(selector || DEFAULT_CONTAINER_SELECTOR)
      if (!container) {
        height.value = 0
        return
      }
      const occupied =
        el.getBoundingClientRect().top - container.getBoundingClientRect().top
      height.value = Math.max(container.clientHeight - occupied - bottomReserve, min)
    })
  }

  const observeContainer = () => {
    const el = wrapRef.value
    const container = el && el.closest(selector || DEFAULT_CONTAINER_SELECTOR)
    if (!container || typeof ResizeObserver === 'undefined') return
    if (resizeObserver) resizeObserver.disconnect()
    resizeObserver = new ResizeObserver(() => compute())
    resizeObserver.observe(container)
  }

  onMounted(async () => {
    await nextTick()
    compute()
    observeContainer()
    // 路由 transition / 查询区首帧渲染完成后复算兜底
    recheckTimer = setTimeout(compute, TRANSITION_RECHECK_DELAY)
    window.addEventListener('resize', compute)
  })

  // keep-alive 切回时容器尺寸可能已变化
  onActivated(() => {
    nextTick(compute)
  })

  onBeforeUnmount(() => {
    if (resizeObserver) {
      resizeObserver.disconnect()
      resizeObserver = null
    }
    window.removeEventListener('resize', compute)
    if (rafId) cancelAnimationFrame(rafId)
    if (recheckTimer) clearTimeout(recheckTimer)
  })

  return { wrapRef, height, compute }
}
