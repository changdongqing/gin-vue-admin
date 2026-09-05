// GvaGrid 数据代理：gin-vue-admin 后端协议适配
// 入参 { page, pageSize, orderKey, desc, ...搜索字段 } → POST
// 出参 { code: 0, data: { list, total, page, pageSize } }
import { toSQLLine } from '@/utils/stringFun'
import { GVA_PAGER_DEFAULTS } from '../config/defaults'

// vxe sorts → 后端排序参数
// sortProtocol:
//   'gva'（默认）→ { orderKey: 蛇形, desc: bool }，核心模块约定
//   'autoCode'  → { sort: 蛇形列名, order: 'ascending'|'descending' }，代码生成器约定
export function toOrderParams(sorts, defaultSort, sortProtocol = 'gva') {
  const sort = sorts && sorts.length ? sorts[0] : null
  const isAutoCode = sortProtocol === 'autoCode'
  if (!sort || !sort.field || !sort.order) {
    if (defaultSort && defaultSort.field) {
      const field = defaultSort.field
      if (isAutoCode) {
        return {
          sort: field === 'ID' ? 'id' : toSQLLine(field),
          order: (defaultSort.order || 'desc') === 'asc' ? 'ascending' : 'descending'
        }
      }
      return {
        orderKey: field === 'ID' ? 'id' : toSQLLine(field),
        desc: (defaultSort.order || 'desc') !== 'asc'
      }
    }
    return {}
  }
  if (isAutoCode) {
    return {
      sort: sort.field === 'ID' ? 'id' : toSQLLine(sort.field),
      order: sort.order !== 'asc' ? 'descending' : 'ascending'
    }
  }
  return {
    orderKey: sort.field === 'ID' ? 'id' : toSQLLine(sort.field),
    desc: sort.order !== 'asc'
  }
}

/**
 * 构建 proxyConfig
 * @param {Object} options useGvaGrid 的 options（api/beforeQuery/afterQuery/defaultSort/autoLoad/searchItems…）
 */
export function buildProxyConfig(options = {}) {
  const { api, beforeQuery, afterQuery, defaultSort } = options

  // 日期范围拆分：gvaDateRange 存 [start, end] 数组，
  // 若查询项声明了 itemRender.props.startField/endField，提交前拆为独立查询字段
  const applyDateRangeSplit = (params) => {
    const items = options.searchItems || []
    items.forEach((item) => {
      const render = item.itemRender || {}
      const props = render.props || {}
      if (render.name === 'gvaDateRange' && props.startField && props.endField) {
        const value = params[item.field]
        if (Array.isArray(value) && value.length === 2) {
          params[props.startField] = value[0]
          params[props.endField] = value[1]
        }
        delete params[item.field]
      }
    })
  }

  const query = async ({ page, sorts, form }) => {
    if (!api) {
      console.warn('[GvaGrid] useGvaGrid 缺少 api 配置')
      return { result: [], total: 0 }
    }
    const params = {
      page: (page && page.currentPage) || 1,
      pageSize: (page && page.pageSize) || options.pageSize || 10,
      ...(form || {}),
      ...toOrderParams(
        sorts,
        defaultSort === undefined ? { field: 'ID', order: 'desc' } : defaultSort,
        options.sortProtocol || 'gva'
      )
    }
    applyDateRangeSplit(params)
    if (beforeQuery) beforeQuery(params)

    // loadingOption: false → 关闭全局 ElLoading，加载态交给 vxe 表格遮罩
    const res = await api({ ...params, loadingOption: false })
    if (!res || res.code !== 0) {
      // 错误提示已由 utils/request.js 响应拦截统一处理
      return { result: [], total: 0 }
    }
    if (afterQuery) afterQuery(res)
    const data = res.data || {}
    return { result: data.list || [], total: Number(data.total || 0) }
  }

  return {
    form: true,
    // 注意：必须用 response 而非 props（vxe 全局默认 response 永远覆盖 props 旧式配置）
    // list 供 pagerConfig 为 null（pager: false）时使用：vxe 无分页分支只认 response.list 或数组返回
    response: { result: 'result', list: 'result', total: 'total', message: 'message' },
    autoLoad: options.autoLoad !== false,
    ajax: {
      query,
      querySuccess(params) {
        // 空页回退（旧页面 customer.vue 的“末页删空回退页码”补丁，统一到封装）
        if (options.emptyPageFallback === false) return
        const { $grid, page, response } = params
        const currentPage = Number((page && page.currentPage) || $grid.getCurrentPage())
        if (currentPage > 1 && (!response || !response.result || !response.result.length)) {
          $grid.setCurrentPage(currentPage - 1)
          $grid.commitProxy('query')
        }
      }
    }
  }
}

export function buildPagerConfig(options = {}) {
  return {
    ...GVA_PAGER_DEFAULTS,
    pageSize: options.pageSize || GVA_PAGER_DEFAULTS.pageSize,
    pageSizes: options.pageSizes || GVA_PAGER_DEFAULTS.pageSizes
  }
}
