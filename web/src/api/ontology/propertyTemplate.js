import service from '@/utils/request'

// @Router /ontology/propertyTemplate/getPropertyTemplateList [get]
export const getPropertyTemplateList = (params) => {
  return service({
    url: '/ontology/propertyTemplate/getPropertyTemplateList',
    method: 'get',
    params
  })
}

// @Router /ontology/propertyTemplate/getPropertyTemplateAll [get]
export const getPropertyTemplateAll = (params) => {
  return service({
    url: '/ontology/propertyTemplate/getPropertyTemplateAll',
    method: 'get',
    params
  })
}

// @Router /ontology/propertyTemplate/findPropertyTemplate [get]
export const findPropertyTemplate = (params) => {
  return service({
    url: '/ontology/propertyTemplate/findPropertyTemplate',
    method: 'get',
    params
  })
}

// @Router /ontology/propertyTemplate/createPropertyTemplate [post]
export const createPropertyTemplate = (data) => {
  return service({
    url: '/ontology/propertyTemplate/createPropertyTemplate',
    method: 'post',
    data
  })
}

// @Router /ontology/propertyTemplate/updatePropertyTemplate [put]
export const updatePropertyTemplate = (data) => {
  return service({
    url: '/ontology/propertyTemplate/updatePropertyTemplate',
    method: 'put',
    data
  })
}

// @Router /ontology/propertyTemplate/deletePropertyTemplate [delete]
export const deletePropertyTemplate = (params) => {
  return service({
    url: '/ontology/propertyTemplate/deletePropertyTemplate',
    method: 'delete',
    params
  })
}

// @Router /ontology/propertyTemplate/disablePropertyTemplate [put]
export const disablePropertyTemplate = (params) => {
  return service({
    url: '/ontology/propertyTemplate/disablePropertyTemplate',
    method: 'put',
    params
  })
}

// @Router /ontology/propertyTemplate/checkPropertyTemplateCode [get]
export const checkPropertyTemplateCode = (params) => {
  return service({
    url: '/ontology/propertyTemplate/checkPropertyTemplateCode',
    method: 'get',
    params
  })
}

// 供给（建模侧用）
// @Router /ontology/supply/v1/propertyTemplates [get]
export const getPropertyTemplatesForSupply = (params) => {
  return service({
    url: '/ontology/supply/v1/propertyTemplates',
    method: 'get',
    params
  })
}
