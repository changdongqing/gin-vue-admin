import service from '@/utils/request'

export const getAnalysisReportList = (params) => service({ url: '/report/analysisReport/getAnalysisReportList', method: 'get', params })
export const findAnalysisReport = (params) => service({ url: '/report/analysisReport/findAnalysisReport', method: 'get', params })
export const getAnalysisReportByCode = (params) => service({ url: '/report/analysisReport/getAnalysisReportByCode', method: 'get', params })
export const createAnalysisReport = (data) => service({ url: '/report/analysisReport/createAnalysisReport', method: 'post', data })
export const updateAnalysisReport = (data) => service({ url: '/report/analysisReport/updateAnalysisReport', method: 'put', data })
export const deleteAnalysisReport = (params) => service({ url: '/report/analysisReport/deleteAnalysisReport', method: 'delete', params })
export const copyAnalysisReport = (data) => service({ url: '/report/analysisReport/copyAnalysisReport', method: 'post', data })
export const saveAnalysisConfig = (data) => service({ url: '/report/analysisReport/saveAnalysisConfig', method: 'post', data })
export const previewAnalysisReport = (data) => service({ url: '/report/analysisReport/previewAnalysisReport', method: 'post', data })

/**
 * @typedef {Object} ValueConfig
 * @property {string} field
 * @property {'SUM'|'COUNT'|'AVG'|'MIN'|'MAX'|'NONE'} aggregation
 * @property {string} [alias]
 * @property {string} [format]   // numeral 风格模板，如 '#,##0.00'
 */
/**
 * @typedef {Object} ReportConfig
 * @property {'pivot'|'table'} sheetType
 * @property {{rows:string[], columns:string[], values:ValueConfig[], valueInRow:boolean}} fields
 * @property {{pagination:{open:boolean,current:number,pageSize:number}, frozenRowHeader:boolean,
 *   showSeriesNumber:boolean, hierarchyType:'grid'|'tree', layoutWidthType:'adaptive'|'colAdaptive'|'compact',
 *   adaptive:boolean, totals:{row:{showGrandTotals:boolean,showSubTotals:boolean,subTotalsDimensions:string[]},
 *   col:{showGrandTotals:boolean,showSubTotals:boolean,subTotalsDimensions:string[]}}, style?:{rowHeight?:number,colWidth?:number}}} options
 * @property {{name:'default'|'dark'|'colorful'|'gray'}} theme
 */

export const AGGREGATION_OPTIONS = [
  { label: '求和 (SUM)', value: 'SUM' },
  { label: '计数 (COUNT)', value: 'COUNT' },
  { label: '平均 (AVG)', value: 'AVG' },
  { label: '最小 (MIN)', value: 'MIN' },
  { label: '最大 (MAX)', value: 'MAX' },
  { label: '不聚合·直显 (NONE)', value: 'NONE' }
]

export const THEME_OPTIONS = [
  { label: '默认', value: 'default' },
  { label: '暗色', value: 'dark' },
  { label: '多彩', value: 'colorful' },
  { label: '灰色', value: 'gray' }
]

export const HIERARCHY_OPTIONS = [
  { label: '平铺', value: 'grid' },
  { label: '树状', value: 'tree' }
]

export const LAYOUT_WIDTH_OPTIONS = [
  { label: '自适应', value: 'adaptive' },
  { label: '列等宽', value: 'colAdaptive' },
  { label: '紧凑', value: 'compact' }
]

export const PAGE_SIZE_OPTIONS = [20, 50, 100]
