import service from '@/utils/request'

export const getExcelReportList = (params) => service({ url: '/report/excelReport/getExcelReportList', method: 'get', params })
export const findExcelReport = (params) => service({ url: '/report/excelReport/findExcelReport', method: 'get', params })
export const createExcelReport = (data) => service({ url: '/report/excelReport/createExcelReport', method: 'post', data })
export const updateExcelReport = (data) => service({ url: '/report/excelReport/updateExcelReport', method: 'put', data })
export const deleteExcelReport = (params) => service({ url: '/report/excelReport/deleteExcelReport', method: 'delete', params })
export const copyExcelReport = (data) => service({ url: '/report/excelReport/copyExcelReport', method: 'post', data })
export const saveExcelTemplate = (data) => service({ url: '/report/excelReport/saveExcelTemplate', method: 'post', data })
export const bindExcelReportDataSets = (data) => service({ url: '/report/excelReport/bindExcelReportDataSets', method: 'post', data })
export const getExcelReportDataSetFields = (params) => service({ url: '/report/excelReport/getExcelReportDataSetFields', method: 'get', params })
