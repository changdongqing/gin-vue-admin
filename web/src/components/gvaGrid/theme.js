// vxe 内置主题与 gin-vue-admin 暗色模式（html.dark）同步
// vxe v4 通过 data-vxe-ui-theme 属性切换完整明/暗变量集（[data-vxe-ui-theme=dark]）
import { VxeUI } from 'vxe-pc-ui'

export function syncVxeTheme() {
  const isDark = document.documentElement.classList.contains('dark')
  VxeUI.setTheme(isDark ? 'dark' : 'light')
}

let observer = null

export function setupGvaGridTheme() {
  if (typeof document === 'undefined') return
  syncVxeTheme()
  if (observer) return
  // gin-vue-admin 通过 useDark 切换 html.dark，这里监听其变化同步 vxe 主题
  observer = new MutationObserver(syncVxeTheme)
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
}
