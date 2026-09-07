const viewModules = import.meta.glob('../view/**/*.vue')
const pluginModules = import.meta.glob('../plugin/**/*.vue')

export const asyncRouterHandle = (asyncRouter) => {
  asyncRouter.forEach((item) => {
    // path 内嵌 query 适配（「添加到菜单」的预览路由）：注册前拆出 path 与注入 query，
    // 路由 path 保持干净（vue-router 不支持 '?'）；组件从 route.meta.queryInPath 兜底取参。
    // 不含 '?' 的菜单完全不受影响（向后兼容）
    if (item.path && item.path.includes('?')) {
      const [pathPart, queryPart] = item.path.split('?')
      item.path = pathPart
      item.meta = item.meta || {}
      item.meta.queryInPath = Object.fromEntries(new URLSearchParams(queryPart))
    }
    if (item.component && typeof item.component === 'string') {
      item.meta.path = '/src/' + item.component
      if (item.component.split('/')[0] === 'view') {
        item.component = dynamicImport(viewModules, item.component)
      } else if (item.component.split('/')[0] === 'plugin') {
        item.component = dynamicImport(pluginModules, item.component)
      }
    }
    if (item.children) {
      asyncRouterHandle(item.children)
    }
  })
}

function dynamicImport(dynamicViewsModules, component) {
  const keys = Object.keys(dynamicViewsModules)
  const matchKeys = keys.filter((key) => {
    const k = key.replace('../', '')
    return k === component
  })
  const matchKey = matchKeys[0]

  return dynamicViewsModules[matchKey]
}
