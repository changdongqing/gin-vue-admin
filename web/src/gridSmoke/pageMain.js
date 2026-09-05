// 迁移页运行时接线验证入口（不依赖后端；API 请求失败 → 空数据，仅验证页面渲染与事件接线）
// 访问 http://localhost:8080/page-smoke.html#user 或 #api
import { createApp, defineComponent, h, computed, ref } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHashHistory } from 'vue-router'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import VxeUIAll from 'vxe-pc-ui'
import 'vxe-pc-ui/lib/style.css'
import VxeUITable from 'vxe-table'
import 'vxe-table/lib/style.css'
import 'uno.css'
import '@/components/gvaGrid'

import UserPage from '@/view/superAdmin/user/user.vue'
import ApiPage from '@/view/superAdmin/api/api.vue'

// 模拟 layout 容器结构（自适应高度依赖 .gva-container2）
const Layout = defineComponent({
  setup() {
    const current = ref(location.hash.includes('api') ? 'api' : 'user')
    window.addEventListener('hashchange', () => {
      current.value = location.hash.includes('api') ? 'api' : 'user'
    })
    const comp = computed(() => (current.value === 'api' ? ApiPage : UserPage))
    return () =>
      h('div', { style: { height: '100%', boxSizing: 'border-box', paddingTop: '3rem' } }, [
        h('div', { class: 'gva-container2', style: { height: 'calc(100% - 1.5rem)', overflow: 'auto', padding: '0 8px' } }, [
          h('div', { class: 'gva-body-h' }, [h(comp.value, { key: current.value })]),
          h('div', { style: { height: '28px', textAlign: 'center', color: '#94a3b8', fontSize: '12px' } }, 'BottomInfo 模拟')
        ])
      ])
  }
})

const router = createRouter({
  history: createWebHashHistory(),
  routes: [{ path: '/:pathMatch(.*)*', component: Layout }]
})

const app = createApp(Layout)
app.use(createPinia())
app.use(router)
app.use(ElementPlus)
app.use(VxeUIAll)
app.use(VxeUITable)
app.mount('#app')
