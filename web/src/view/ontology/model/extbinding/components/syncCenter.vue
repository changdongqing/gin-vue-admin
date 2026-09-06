<template>
  <div>
    <!-- 生效绑定卡片流 -->
    <div class="flex flex-wrap gap-3 mb-4">
      <el-card v-for="b in activeBindings" :key="b.ID" shadow="never" class="w-[340px]">
        <div class="font-medium mb-1">{{ b.classLocalName }} ↔ {{ b.tableName }}</div>
        <div class="text-xs text-gray-400 mb-3">
          水位 {{ b.watermark || '无' }} · 模式 {{ b.syncMode === 2 ? '定时+手动' : '手动' }}
          <template v-if="b.lastSyncTime"> · 最近 {{ b.lastSyncTime }}</template>
        </div>
        <div class="text-xs text-gray-500 mb-3 min-h-4">{{ b.lastSyncSummary || '尚未同步' }}</div>
        <div>
          <el-button size="small" type="primary" plain :loading="triggering === b.ID + '-2'" @click="trigger(b, 2)">增量同步</el-button>
          <el-button size="small" type="primary" :loading="triggering === b.ID + '-1'" @click="triggerFull(b)">全量同步</el-button>
        </div>
      </el-card>
      <el-empty v-if="!activeBindings.length" description="暂无生效绑定" :image-size="48" class="w-full" />
    </div>

    <!-- 同步结果弹窗 -->
    <el-dialog v-model="resultVisible" title="同步结果" width="560px" append-to-body>
      <template v-if="lastResult">
        <div class="mb-2">扫描 <b>{{ lastResult.totalRows }}</b> 行 · 耗时 {{ lastResult.durationMs }} ms</div>
        <div class="flex gap-2 mb-3 flex-wrap">
          <el-tag type="success">新建 {{ lastResult.created }}</el-tag>
          <el-tag type="primary">更新 {{ lastResult.updated }}</el-tag>
          <el-tag type="info">跳过 {{ lastResult.skipped }}</el-tag>
          <el-tag type="danger">失败 {{ lastResult.failed }}</el-tag>
          <el-tag type="warning">问题 {{ lastResult.issuesCount }}</el-tag>
        </div>
        <div v-if="lastResult.skipReasons?.length" class="text-xs text-gray-500 max-h-40 overflow-auto">
          <div v-for="(r, i) in lastResult.skipReasons" :key="i">· {{ r }}</div>
        </div>
      </template>
      <template #footer>
        <el-button type="primary" @click="resultVisible = false">关 闭</el-button>
      </template>
    </el-dialog>

    <!-- 同步日志 -->
    <el-divider content-position="left">同步日志</el-divider>
    <GvaGrid ref="logGridRef" v-bind="logGridOptions" v-on="logGridEvents">
      <template #statusTag="{ row }">
        <el-tag v-if="row.status === 1" type="success">成功</el-tag>
        <el-tag v-else-if="row.status === 2" type="warning">部分失败</el-tag>
        <el-tag v-else type="danger">失败</el-tag>
      </template>
    </GvaGrid>
  </div>
</template>

<script setup>
  import { getExtBindingList } from '@/api/ontology/extBinding'
  import { triggerExtSync, getExtSyncLogList } from '@/api/ontology/extSync'
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import GvaGrid, { useGvaGrid } from '@/components/gvaGrid'

  defineOptions({ name: 'SyncCenter' })

  const activeBindings = ref([])
  const triggering = ref('')
  const resultVisible = ref(false)
  const lastResult = ref(null)

  const { gridRef: logGridRef, gridOptions: logGridOptions, gridEvents: logGridEvents, refresh: refreshLogs } = useGvaGrid({
    id: 'ontology-extSyncLog',
    defaultSort: null,
    api: getExtSyncLogList,
    searchItems: [],
    columns: [
      { field: 'startTime', title: '开始时间', width: 170, cellRender: { name: 'gvaDate' } },
      { field: 'triggerType', title: '触发', width: 90, formatter: ({ cellValue }) => ({ 1: '反向', 2: '手动', 3: '定时', 4: '试运行' }[cellValue] || '-') },
      { field: 'syncScope', title: '范围', width: 80, formatter: ({ cellValue }) => (cellValue === 2 ? '增量' : '全量') },
      { field: 'totalRows', title: '扫描', width: 70 },
      { field: 'created', title: '新建', width: 70 },
      { field: 'updated', title: '更新', width: 70 },
      { field: 'skipped', title: '跳过', width: 70 },
      { field: 'failed', title: '失败', width: 70 },
      { field: 'issuesCount', title: '问题', width: 70 },
      { field: 'status', title: '状态', width: 90, slots: { default: 'statusTag' } },
      { field: 'durationMs', title: '耗时', width: 90, formatter: ({ cellValue }) => `${cellValue} ms` },
      { field: 'summary', title: '摘要', minWidth: 200, showOverflow: true }
    ]
  })

  const loadBindings = async () => {
    const res = await getExtBindingList({ page: 1, pageSize: 50, bindingStatus: 1 })
    if (res.code === 0) {
      activeBindings.value = res.data?.list || []
    }
  }
  const refresh = () => {
    loadBindings()
    refreshLogs()
  }
  defineExpose({ refresh })
  refresh()

  const trigger = async (b, scope) => {
    triggering.value = `${b.ID}-${scope}`
    try {
      const res = await triggerExtSync({ bindingId: b.ID, scope })
      if (res.code === 0) {
        lastResult.value = res.data
        resultVisible.value = true
        refresh()
      }
    } finally {
      triggering.value = ''
    }
  }
  const triggerFull = (b) => {
    ElMessageBox.confirm('全量同步将做孤儿检测（失联对象标记且不物理删除），是否继续?', '提示', { type: 'warning' })
      .then(() => trigger(b, 1))
      .catch(() => {})
  }
  void ElMessage
</script>
