import service from '@/utils/request'

export const getModelClassList = (params) => service({ url: '/ontology/modelClass/getModelClassList', method: 'get', params })
export const getModelClassByProject = (params) => service({ url: '/ontology/modelClass/getModelClassByProject', method: 'get', params })
export const findModelClass = (params) => service({ url: '/ontology/modelClass/findModelClass', method: 'get', params })
export const getModelClassDetail = (params) => service({ url: '/ontology/modelClass/getModelClassDetail', method: 'get', params })
export const createModelClass = (data) => service({ url: '/ontology/modelClass/createModelClass', method: 'post', data })
export const updateModelClass = (data) => service({ url: '/ontology/modelClass/updateModelClass', method: 'put', data })
export const deleteModelClass = (params) => service({ url: '/ontology/modelClass/deleteModelClass', method: 'delete', params })
export const checkModelClassLocalName = (params) => service({ url: '/ontology/modelClass/checkModelClassLocalName', method: 'get', params })
export const instantiateModelClass = (data) => service({ url: '/ontology/modelClass/instantiateModelClass', method: 'post', data })
export const previewInstantiateModelClass = (params) => service({ url: '/ontology/modelClass/previewInstantiateModelClass', method: 'get', params })
