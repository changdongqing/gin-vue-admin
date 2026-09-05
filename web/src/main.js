import './style/element_visiable.scss'
import 'element-plus/theme-chalk/dark/css-vars.css'
import 'uno.css'
import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import { setupVueRootValidator } from 'vite-check-multiple-dom/client';

import 'element-plus/dist/index.css'
// vxe-table v4：表格核心与基础组件库（GvaGrid 统一表格封装依赖，见 aiDoc/vxegrid）
import VxeUIAll from 'vxe-pc-ui'
import 'vxe-pc-ui/lib/style.css'
import VxeUITable from 'vxe-table'
import 'vxe-table/lib/style.css'
// 引入gin-vue-admin前端初始化相关内容
import './core/gin-vue-admin'
// 引入封装的router
import router from '@/router/index'
import '@/permission'
import run from '@/core/gin-vue-admin.js'
import auth from '@/directive/auth'
import clickOutSide from '@/directive/clickOutSide'
import { store } from '@/pinia'
import App from './App.vue'
import '@/core/error-handel'
// GvaGrid 统一表格封装（副作用导入：注册渲染器与主题样式，见 aiDoc/vxegrid）
import '@/components/gvaGrid'

const app = createApp(App)

app.config.productionTip = false

setupVueRootValidator(app, {
    lang: 'zh'
  })

app
  .use(run)
  .use(ElementPlus)
  .use(VxeUIAll)
  .use(VxeUITable)
  .use(store)
  .use(auth)
  .use(clickOutSide)
  .use(router)
  .mount('#app')
export default app
