// GvaGrid 全局默认配置
// 与 gin-vue-admin 现有列表页视觉/协议对齐，详见 aiDoc/vxegrid/01-详细设计方案.md

export const GVA_GRID_DEFAULTS = {
  // 对齐现有 el-table 视觉（36/32px 行高，element_visiable.scss 覆写值）
  size: 'medium',
  // 全边框：所有单元格显示横竖分割线（vxe border='full'）
  border: 'full',
  stripe: false,
  align: 'left',
  // 全局单行省略 + tooltip（替代逐列 show-overflow-tooltip）
  showOverflow: true,
  rowConfig: {
    keyField: 'ID',
    isHover: true,
    isCurrent: true
  },
  // 全部走后端排序（orderKey + desc 协议）
  sortConfig: {
    remote: true,
    trigger: 'cell'
  },
  checkboxConfig: {
    reserve: true,
    highlight: true
  },
  // 列设置持久化（需要 grid id，由 useGvaGrid 注入 options.id）
  customConfig: {
    storage: true
  },
  // 虚拟滚动阈值：大数据量页面（操作记录/登录日志）受益
  scrollY: {
    enabled: true,
    gt: 60
  },
  scrollX: {
    enabled: true,
    gt: 12
  },
  tooltipConfig: {
    enterable: true
  }
}

export const GVA_PAGER_DEFAULTS = {
  // 对齐现有 el-pagination layout="total, sizes, prev, pager, next, jumper"
  layouts: ['Total', 'Sizes', 'PrevJump', 'PrevPage', 'Number', 'NextPage', 'NextJump', 'FullJump'],
  pageSizes: [10, 30, 50, 100],
  pageSize: 10
}

export const GVA_TOOLBAR_DEFAULTS = {
  // code: 'query' → 刷新保持当前页/条件（reload 会回第 1 页）
  refresh: { code: 'query' },
  custom: true,
  zoom: true,
  slots: {
    buttons: 'toolbar-buttons',
    tools: 'gva-grid-tools'
  }
}

// 工具栏密度选项（size 联动 vxe-grid）
export const GVA_SIZE_OPTIONS = [
  { label: '默认', value: 'medium' },
  { label: '紧凑', value: 'small' },
  { label: '迷你', value: 'mini' }
]
