import service from '@/utils/request'

// @Router /ontology/annotationProperty/getAnnotationPropertyList [get]
export const getAnnotationPropertyList = (params) => {
  return service({
    url: '/ontology/annotationProperty/getAnnotationPropertyList',
    method: 'get',
    params
  })
}

// @Router /ontology/annotationProperty/getAnnotationPropertyAll [get]
export const getAnnotationPropertyAll = (params) => {
  return service({
    url: '/ontology/annotationProperty/getAnnotationPropertyAll',
    method: 'get',
    params
  })
}

// @Router /ontology/annotationProperty/findAnnotationProperty [get]
export const findAnnotationProperty = (params) => {
  return service({
    url: '/ontology/annotationProperty/findAnnotationProperty',
    method: 'get',
    params
  })
}

// @Router /ontology/annotationProperty/createAnnotationProperty [post]
export const createAnnotationProperty = (data) => {
  return service({
    url: '/ontology/annotationProperty/createAnnotationProperty',
    method: 'post',
    data
  })
}

// @Router /ontology/annotationProperty/updateAnnotationProperty [put]
export const updateAnnotationProperty = (data) => {
  return service({
    url: '/ontology/annotationProperty/updateAnnotationProperty',
    method: 'put',
    data
  })
}

// @Router /ontology/annotationProperty/deleteAnnotationProperty [delete]
export const deleteAnnotationProperty = (params) => {
  return service({
    url: '/ontology/annotationProperty/deleteAnnotationProperty',
    method: 'delete',
    params
  })
}

// 导出（blob 流下载）
// @Router /ontology/annotationProperty/exportAnnotationPropertyExcel [get]
export const exportAnnotationPropertyExcel = (params) => {
  return service({
    url: '/ontology/annotationProperty/exportAnnotationPropertyExcel',
    method: 'get',
    params,
    responseType: 'blob'
  })
}

// 供给（建模侧用）
// @Router /ontology/supply/v1/annotationProperties [get]
export const getAnnotationPropertiesForSupply = (params) => {
  return service({
    url: '/ontology/supply/v1/annotationProperties',
    method: 'get',
    params
  })
}
