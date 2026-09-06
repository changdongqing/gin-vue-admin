import service from '@/utils/request'

export const getExtBindingList = (params) => service({ url: '/ontology/extBinding/getExtBindingList', method: 'get', params })
export const findExtBinding = (params) => service({ url: '/ontology/extBinding/findExtBinding', method: 'get', params })
export const createExtBinding = (data) => service({ url: '/ontology/extBinding/createExtBinding', method: 'post', data })
export const updateExtBinding = (data) => service({ url: '/ontology/extBinding/updateExtBinding', method: 'post', data })
export const deleteExtBinding = (params) => service({ url: '/ontology/extBinding/deleteExtBinding', method: 'delete', params })
export const changeExtBindingStatus = (params) => service({ url: '/ontology/extBinding/changeExtBindingStatus', method: 'put', params })
