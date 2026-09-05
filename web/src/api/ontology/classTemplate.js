import service from '@/utils/request'

// @Router /ontology/classTemplate/getClassTemplateList [get]
export const getClassTemplateList = (params) => {
  return service({
    url: '/ontology/classTemplate/getClassTemplateList',
    method: 'get',
    params
  })
}

// @Router /ontology/classTemplate/getClassTemplatePage [get]
export const getClassTemplatePage = (params) => {
  return service({
    url: '/ontology/classTemplate/getClassTemplatePage',
    method: 'get',
    params
  })
}

// @Router /ontology/classTemplate/getClassTemplateTreeRoots [get]
export const getClassTemplateTreeRoots = () => {
  return service({
    url: '/ontology/classTemplate/getClassTemplateTreeRoots',
    method: 'get'
  })
}

// @Router /ontology/classTemplate/findClassTemplate [get]
export const findClassTemplate = (params) => {
  return service({
    url: '/ontology/classTemplate/findClassTemplate',
    method: 'get',
    params
  })
}

// @Router /ontology/classTemplate/createClassTemplate [post]
export const createClassTemplate = (data) => {
  return service({
    url: '/ontology/classTemplate/createClassTemplate',
    method: 'post',
    data
  })
}

// @Router /ontology/classTemplate/updateClassTemplate [put]
export const updateClassTemplate = (data) => {
  return service({
    url: '/ontology/classTemplate/updateClassTemplate',
    method: 'put',
    data
  })
}

// @Router /ontology/classTemplate/deleteClassTemplate [delete]
export const deleteClassTemplate = (params) => {
  return service({
    url: '/ontology/classTemplate/deleteClassTemplate',
    method: 'delete',
    params
  })
}

// @Router /ontology/classTemplate/getClassTemplateRefList [get]
export const getClassTemplateRefList = (params) => {
  return service({
    url: '/ontology/classTemplate/getClassTemplateRefList',
    method: 'get',
    params
  })
}

// @Router /ontology/classTemplate/previewClassificationCode [get]
export const previewClassificationCode = (params) => {
  return service({
    url: '/ontology/classTemplate/previewClassificationCode',
    method: 'get',
    params
  })
}

// @Router /ontology/classificationRule/findClassificationRule [get]
export const findClassificationRule = (params) => {
  return service({
    url: '/ontology/classificationRule/findClassificationRule',
    method: 'get',
    params
  })
}

// @Router /ontology/classificationRule/saveClassificationRule [put]
export const saveClassificationRule = (data) => {
  return service({
    url: '/ontology/classificationRule/saveClassificationRule',
    method: 'put',
    data
  })
}

// 供给（建模侧用）
// @Router /ontology/supply/v1/classTemplate/tree [get]
export const getClassTemplateTreeForSupply = (params) => {
  return service({
    url: '/ontology/supply/v1/classTemplate/tree',
    method: 'get',
    params
  })
}
