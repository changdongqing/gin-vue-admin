import service from '@/utils/request'

// ============ 采集平台 API（后端路由 /collect/**，前端经 /api 代理） ============

// 通道
export const getChannelList = () => service({ url: '/collect/channels', method: 'get' })
export const findChannel = (id) => service({ url: `/collect/channels/${id}`, method: 'get' })
export const createChannel = (data) => service({ url: '/collect/channels', method: 'post', data })
export const updateChannel = (data) => service({ url: '/collect/channels', method: 'put', data })
export const deleteChannel = (params) => service({ url: `/collect/channels/${params.id}`, method: 'delete', params })

// 设备
export const getDeviceList = (params) => service({ url: '/collect/devices', method: 'get', params })
export const createDevice = (data) => service({ url: '/collect/devices', method: 'post', data })
export const updateDevice = (data) => service({ url: '/collect/devices', method: 'put', data })
export const deleteDevice = (params) => service({ url: `/collect/devices/${params.id}`, method: 'delete', params })

// 测点
export const getVariableList = (params) => service({ url: '/collect/variables', method: 'get', params })
export const createVariable = (data) => service({ url: '/collect/variables', method: 'post', data })
export const updateVariable = (data) => service({ url: '/collect/variables', method: 'put', data })
export const deleteVariable = (params) => service({ url: `/collect/variables/${params.id}`, method: 'delete', params })

// 设备类型（报文型绑定）
export const getDeviceTypeList = () => service({ url: '/collect/deviceTypes', method: 'get' })
export const createDeviceType = (data) => service({ url: '/collect/deviceTypes', method: 'post', data })
export const updateDeviceType = (data) => service({ url: '/collect/deviceTypes', method: 'put', data })
export const deleteDeviceType = (params) => service({ url: `/collect/deviceTypes/${params.id}`, method: 'delete', params })

// 树
export const getCollectTree = () => service({ url: '/collect/tree', method: 'get' })

// 部署
export const deployChannel = (id) => service({ url: `/collect/deploy/${id}`, method: 'post' })
export const undeployChannel = (id) => service({ url: `/collect/deploy/${id}`, method: 'delete' })
export const rebuildAll = () => service({ url: '/collect/deploy/rebuild-all', method: 'post' })
export const getDeployments = (params) => service({ url: '/collect/deployments', method: 'get', params })

// 子流程库
export const getParseChainList = () => service({ url: '/collect/parseChains', method: 'get' })
export const createParseChain = (data) => service({ url: '/collect/parseChains', method: 'post', data })
export const updateParseChain = (data) => service({ url: '/collect/parseChains', method: 'put', data })
export const deleteParseChain = (params) => service({ url: `/collect/parseChains/${params.id}`, method: 'delete', params })
export const publishParseChain = (id) => service({ url: `/collect/parseChains/${id}/publish`, method: 'post' })
export const testParseChain = (id) => service({ url: `/collect/parseChains/${id}/test`, method: 'post' })

// 实时数据与状态
export const getRealtime = (params) => service({ url: '/collect/realtime', method: 'get', params })
export const getChannelStatus = (id) => service({ url: `/collect/channels/${id}/status`, method: 'get' })

// Excel 导入导出（下载走 axios blob，携带 x-token 拦截器头）
export const downloadImportTemplate = () =>
  service({ url: '/collect/import/template', method: 'get', responseType: 'blob' })
export const exportConfig = () =>
  service({ url: '/collect/import/export', method: 'get', responseType: 'blob' })
export const importPreview = (formData) =>
  service({ url: '/collect/import/preview', method: 'post', data: formData, headers: { 'Content-Type': 'multipart/form-data' } })
export const importCommit = (data) => service({ url: '/collect/import/commit', method: 'post', data })
