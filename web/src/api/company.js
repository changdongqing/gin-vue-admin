import service from '@/utils/request'

// @Router /company/createCompany [post]
export const createCompany = (data) => {
  return service({
    url: '/company/createCompany',
    method: 'post',
    data
  })
}

// @Router /company/updateCompany [put]
export const updateCompany = (data) => {
  return service({
    url: '/company/updateCompany',
    method: 'put',
    data
  })
}

// @Router /company/deleteCompany [delete]
export const deleteCompany = (ID) => {
  return service({
    url: '/company/deleteCompany',
    method: 'delete',
    params: { ID }
  })
}

// @Router /company/findCompany [get]
export const findCompany = (ID) => {
  return service({
    url: '/company/findCompany',
    method: 'get',
    params: { ID }
  })
}

// @Router /company/getCompanyList [get]
export const getCompanyList = (params) => {
  return service({
    url: '/company/getCompanyList',
    method: 'get',
    params
  })
}
