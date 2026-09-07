import service from '@/utils/request'

export const previewExcelReport = (data) => service({ url: '/report/excelReport/previewExcelReport', method: 'post', data })
export const getExcelReportParamDefs = (params) => service({ url: '/report/excelReport/getExcelReportParamDefs', method: 'get', params })
