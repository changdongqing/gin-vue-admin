import service from '@/utils/request'

export const getExtModuleList = (params) => service({ url: '/ontology/extModule/getExtModuleList', method: 'get', params })
export const createExtModule = (data) => service({ url: '/ontology/extModule/createExtModule', method: 'post', data })
export const updateExtModule = (data) => service({ url: '/ontology/extModule/updateExtModule', method: 'put', data })
export const deleteExtModule = (params) => service({ url: '/ontology/extModule/deleteExtModule', method: 'delete', params })
export const getExtTableList = (params) => service({ url: '/ontology/extTable/getExtTableList', method: 'get', params })
export const registerExtTables = (data) => service({ url: '/ontology/extTable/registerExtTables', method: 'post', data })
export const deleteExtTable = (params) => service({ url: '/ontology/extTable/deleteExtTable', method: 'delete', params })
export const probeExtTables = (params) => service({ url: '/ontology/extTable/probeExtTables', method: 'get', params })
export const getExtTableColumns = (params) => service({ url: '/ontology/extTable/getExtTableColumns', method: 'get', params })
