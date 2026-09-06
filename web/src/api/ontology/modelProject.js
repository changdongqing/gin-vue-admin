import service from '@/utils/request'

export const getModelProjectList = (params) => service({ url: '/ontology/modelProject/getModelProjectList', method: 'get', params })
export const getModelProjectAll = () => service({ url: '/ontology/modelProject/getModelProjectAll', method: 'get' })
export const findModelProject = (params) => service({ url: '/ontology/modelProject/findModelProject', method: 'get', params })
export const createModelProject = (data) => service({ url: '/ontology/modelProject/createModelProject', method: 'post', data })
export const updateModelProject = (data) => service({ url: '/ontology/modelProject/updateModelProject', method: 'put', data })
export const deleteModelProject = (params) => service({ url: '/ontology/modelProject/deleteModelProject', method: 'delete', params })
export const checkModelProjectCode = (params) => service({ url: '/ontology/modelProject/checkModelProjectCode', method: 'get', params })
export const getModelPrefixList = (params) => service({ url: '/ontology/modelPrefix/getModelPrefixList', method: 'get', params })
export const createModelPrefix = (data) => service({ url: '/ontology/modelPrefix/createModelPrefix', method: 'post', data })
export const updateModelPrefix = (data) => service({ url: '/ontology/modelPrefix/updateModelPrefix', method: 'put', data })
export const deleteModelPrefix = (params) => service({ url: '/ontology/modelPrefix/deleteModelPrefix', method: 'delete', params })
