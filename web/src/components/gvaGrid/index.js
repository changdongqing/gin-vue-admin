// GvaGrid 统一导出入口
// 页面用法：
//   import GvaGrid, { useGvaGrid, useGvaGridDelete } from '@/components/gvaGrid'
//   const { gridRef, gridOptions, gridEvents, selectedRows } = useGvaGrid({ api, columns, searchItems })
//   <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">...</GvaGrid>
import './renderers'
import { setupGvaGridTheme } from './theme'
import './style/index.scss'
import GvaGrid from './index.vue'
import { useGvaGrid } from './hooks/useGvaGrid'
import { useGvaGridDelete } from './hooks/useGvaGridDelete'

// 同步 vxe 内置明/暗主题（跟随 html.dark）
setupGvaGridTheme()

export { GvaGrid, useGvaGrid, useGvaGridDelete }
export default GvaGrid
