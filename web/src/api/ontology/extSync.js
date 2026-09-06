import service from '@/utils/request'

export const dryRunExtSync = (params) => service({ url: '/ontology/extSync/dryRunExtSync', method: 'post', params })
export const triggerExtSync = (data) => service({ url: '/ontology/extSync/triggerExtSync', method: 'post', data })
export const getExtSyncLogList = (params) => service({ url: '/ontology/extSync/getExtSyncLogList', method: 'get', params })
