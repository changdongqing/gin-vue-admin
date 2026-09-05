import service from '@/utils/request'

// @Router /department/createDepartment [post]
export const createDepartment = (data) => {
  return service({
    url: '/department/createDepartment',
    method: 'post',
    data
  })
}

// @Router /department/updateDepartment [put]
export const updateDepartment = (data) => {
  return service({
    url: '/department/updateDepartment',
    method: 'put',
    data
  })
}

// @Router /department/deleteDepartment [delete]
export const deleteDepartment = (ID) => {
  return service({
    url: '/department/deleteDepartment',
    method: 'delete',
    params: { ID }
  })
}

// @Router /department/findDepartment [get]
export const findDepartment = (ID) => {
  return service({
    url: '/department/findDepartment',
    method: 'get',
    params: { ID }
  })
}

// @Router /department/getDepartmentList [get]
export const getDepartmentList = (params) => {
  return service({
    url: '/department/getDepartmentList',
    method: 'get',
    params
  })
}
