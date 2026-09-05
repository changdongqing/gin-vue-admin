// GvaGrid 独立冒烟测试入口（不依赖后端/登录/路由）
// 访问 http://localhost:8080/grid-smoke.html
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import VxeUIAll from 'vxe-pc-ui'
import 'vxe-pc-ui/lib/style.css'
import VxeUITable from 'vxe-table'
import 'vxe-table/lib/style.css'
import 'uno.css'
import '@/components/gvaGrid'
import SmokeApp from './SmokeApp.vue'

const app = createApp(SmokeApp)
app.use(createPinia())
app.use(ElementPlus)
app.use(VxeUIAll)
app.use(VxeUITable)
app.mount('#app')
