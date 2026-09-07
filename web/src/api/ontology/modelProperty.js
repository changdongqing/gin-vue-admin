import service from '@/utils/request'

export const getModelDatatypePropertyList = (params) => service({ url: '/ontology/modelDatatypeProperty/getModelDatatypePropertyList', method: 'get', params })
export const findModelDatatypeProperty = (params) => service({ url: '/ontology/modelDatatypeProperty/findModelDatatypeProperty', method: 'get', params })
export const createModelDatatypeProperty = (data) => service({ url: '/ontology/modelDatatypeProperty/createModelDatatypeProperty', method: 'post', data })
export const updateModelDatatypeProperty = (data) => service({ url: '/ontology/modelDatatypeProperty/updateModelDatatypeProperty', method: 'put', data })
export const deleteModelDatatypeProperty = (params) => service({ url: '/ontology/modelDatatypeProperty/deleteModelDatatypeProperty', method: 'delete', params })
export const instantiateModelDatatypeProperty = (data) => service({ url: '/ontology/modelDatatypeProperty/instantiateModelDatatypeProperty', method: 'post', data })

export const getModelObjectPropertyList = (params) => service({ url: '/ontology/modelObjectProperty/getModelObjectPropertyList', method: 'get', params })
export const findModelObjectProperty = (params) => service({ url: '/ontology/modelObjectProperty/findModelObjectProperty', method: 'get', params })
export const createModelObjectProperty = (data) => service({ url: '/ontology/modelObjectProperty/createModelObjectProperty', method: 'post', data })
export const updateModelObjectProperty = (data) => service({ url: '/ontology/modelObjectProperty/updateModelObjectProperty', method: 'put', data })
export const deleteModelObjectProperty = (params) => service({ url: '/ontology/modelObjectProperty/deleteModelObjectProperty', method: 'delete', params })
export const instantiateModelObjectProperty = (data) => service({ url: '/ontology/modelObjectProperty/instantiateModelObjectProperty', method: 'post', data })
export const suggestInverseModelObjectProperty = (params) => service({ url: '/ontology/modelObjectProperty/suggestInverseModelObjectProperty', method: 'get', params })
