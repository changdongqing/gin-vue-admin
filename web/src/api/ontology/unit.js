import service from '@/utils/request'

// @Router /ontology/quantityKind/getQuantityKindList [get]
export const getQuantityKindList = () => {
  return service({
    url: '/ontology/quantityKind/getQuantityKindList',
    method: 'get'
  })
}

// @Router /ontology/unit/getUnitPage [get]
export const getUnitPage = (params) => {
  return service({
    url: '/ontology/unit/getUnitPage',
    method: 'get',
    params
  })
}

// @Router /ontology/unit/getUnitAll [get]
export const getUnitAll = (params) => {
  return service({
    url: '/ontology/unit/getUnitAll',
    method: 'get',
    params
  })
}

// @Router /ontology/unit/findUnit [get]
export const findUnit = (params) => {
  return service({
    url: '/ontology/unit/findUnit',
    method: 'get',
    params
  })
}

// @Router /ontology/unit/createUnit [post]
export const createUnit = (data) => {
  return service({
    url: '/ontology/unit/createUnit',
    method: 'post',
    data
  })
}

// @Router /ontology/unit/updateUnit [put]
export const updateUnit = (data) => {
  return service({
    url: '/ontology/unit/updateUnit',
    method: 'put',
    data
  })
}

// @Router /ontology/unit/deleteUnit [delete]
export const deleteUnit = (params) => {
  return service({
    url: '/ontology/unit/deleteUnit',
    method: 'delete',
    params
  })
}

// @Router /ontology/unit/disableUnit [put]
export const disableUnit = (params) => {
  return service({
    url: '/ontology/unit/disableUnit',
    method: 'put',
    params
  })
}

// @Router /ontology/unit/convertUnit [get]
export const convertUnit = (params) => {
  return service({
    url: '/ontology/unit/convertUnit',
    method: 'get',
    params
  })
}

// 供给（建模侧用）
// @Router /ontology/supply/v1/units [get]
export const getUnitsForSupply = (params) => {
  return service({
    url: '/ontology/supply/v1/units',
    method: 'get',
    params
  })
}

// @Router /ontology/supply/v1/units/convert [get]
export const convertUnitForSupply = (params) => {
  return service({
    url: '/ontology/supply/v1/units/convert',
    method: 'get',
    params
  })
}
