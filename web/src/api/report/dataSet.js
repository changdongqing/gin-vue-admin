import service from '@/utils/request'

export const getDataSetList = (params) => service({ url: '/report/dataSet/getDataSetList', method: 'get', params })
export const findDataSet = (params) => service({ url: '/report/dataSet/findDataSet', method: 'get', params })
export const createDataSet = (data) => service({ url: '/report/dataSet/createDataSet', method: 'post', data })
export const updateDataSet = (data) => service({ url: '/report/dataSet/updateDataSet', method: 'put', data })
export const deleteDataSet = (params) => service({ url: '/report/dataSet/deleteDataSet', method: 'delete', params })
export const getDataSetAll = () => service({ url: '/report/dataSet/getDataSetAll', method: 'get' })
export const testDataSetPreview = (data) => service({ url: '/report/dataSet/testDataSetPreview', method: 'post', data })

/**
 * 参数类型选项（03/04/05 参数表单共用）
 */
export const PARAM_TYPE_OPTIONS = [
  { label: '文本', value: 'string' },
  { label: '数字', value: 'number' },
  { label: '日期', value: 'date' },
  { label: '日期时间', value: 'datetime' },
  { label: '日期范围', value: 'dateRange' },
  { label: '下拉单选', value: 'select' },
  { label: '下拉多选', value: 'multipleSelect' }
]

/** 参数类型分组（参数配置表格下拉用） */
export const PARAM_TYPE_GROUPS = [
  {
    label: '基础',
    options: [
      { label: '文本', value: 'string' },
      { label: '数字', value: 'number' }
    ]
  },
  {
    label: '日期',
    options: [
      { label: '日期', value: 'date' },
      { label: '日期时间', value: 'datetime' },
      { label: '日期范围', value: 'dateRange' }
    ]
  },
  {
    label: '下拉',
    options: [
      { label: '下拉单选', value: 'select' },
      { label: '下拉多选', value: 'multipleSelect' }
    ]
  }
]

export const TRANSFORM_TYPE_OPTIONS = [
  { label: 'JS脚本', value: 'js' },
  { label: '字典映射', value: 'dict' }
]

export const DEFAULT_VALUE_EXPR_OPTIONS = [
  { label: '当天 (today)', value: 'today' },
  { label: '当前时间 (now)', value: 'now' },
  { label: '本月 (thisMonth)', value: 'thisMonth' },
  { label: '本周 (thisWeek)', value: 'thisWeek' },
  { label: '本年 (thisYear)', value: 'thisYear' },
  { label: '上月 (lastMonth)', value: 'lastMonth' },
  { label: '最近7天 (last7Days)', value: 'last7Days' },
  { label: '最近30天 (last30Days)', value: 'last30Days' }
]

/** setType → 中文标签 */
export const setTypeLabel = (value) => (value === 'http' ? 'HTTP' : 'SQL')
