import service from '@/utils/request'

export const getDataSourceList = (params) => service({ url: '/report/dataSource/getDataSourceList', method: 'get', params })
export const findDataSource = (params) => service({ url: '/report/dataSource/findDataSource', method: 'get', params })
export const createDataSource = (data) => service({ url: '/report/dataSource/createDataSource', method: 'post', data })
export const updateDataSource = (data) => service({ url: '/report/dataSource/updateDataSource', method: 'put', data })
export const deleteDataSource = (params) => service({ url: '/report/dataSource/deleteDataSource', method: 'delete', params })
export const getDataSourceAll = () => service({ url: '/report/dataSource/getDataSourceAll', method: 'get' })
export const testDataSourceConnection = (data) => service({ url: '/report/dataSource/testDataSourceConnection', method: 'post', data })

/**
 * 数据源类型选项（分组展示；dsnTemplate 供「填充模板」生成配置骨架）
 * enabled=false 的类型前端置灰并提示「驱动未启用」
 */
export const DATA_SOURCE_TYPE_OPTIONS = [
  {
    label: 'PostgreSQL',
    value: 'postgresql',
    group: '主流数据库',
    enabled: true,
    dsnTemplate: 'postgres://user:pass@127.0.0.1:5432/your_db'
  },
  {
    label: 'MySQL',
    value: 'mysql',
    group: '主流数据库',
    enabled: true,
    dsnTemplate: 'user:pass@tcp(127.0.0.1:3306)/your_db?charset=utf8mb4&parseTime=true'
  },
  {
    label: 'SQL Server',
    value: 'sqlserver',
    group: '主流数据库',
    enabled: true,
    dsnTemplate: 'sqlserver://user:pass@127.0.0.1:1433?database=your_db'
  },
  {
    label: '达梦数据库',
    value: 'dameng',
    group: '信创数据库',
    enabled: false,
    dsnTemplate: 'dm://user:pass@127.0.0.1:5236'
  },
  {
    label: '人大金仓',
    value: 'kingbase',
    group: '信创数据库',
    enabled: false,
    dsnTemplate: 'postgresql://user:pass@127.0.0.1:54321/your_db'
  },
  {
    label: 'openGauss',
    value: 'opengauss',
    group: '信创数据库',
    enabled: false,
    dsnTemplate: 'opengauss://user:pass@127.0.0.1:5432/your_db'
  },
  {
    label: 'Oracle',
    value: 'oracle',
    group: '主流数据库',
    enabled: false,
    dsnTemplate: 'oracle://user:pass@127.0.0.1:1521/ORCLCDB'
  },
  {
    label: 'HTTP接口',
    value: 'http',
    group: '其他',
    enabled: true,
    dsnTemplate: ''
  }
]

/** 按类型分组（el-option-group 渲染用） */
export const DATA_SOURCE_TYPE_GROUPS = [...new Set(DATA_SOURCE_TYPE_OPTIONS.map((o) => o.group))]

/** sourceType → 中文标签 */
export const sourceTypeLabel = (value) => DATA_SOURCE_TYPE_OPTIONS.find((o) => o.value === value)?.label || value
