<template>
  <div class="collect-monitor">
    <warning-bar title="采集监控：通道卡片状态由部署状态与数据新鲜度推导（3×采集周期无数据视为离线）；实时表格 5 秒轮询" />
    <!-- 通道卡片流 -->
    <div class="flex flex-wrap gap-3 mb-4">
      <el-card v-for="c in cards" :key="c.channelId" shadow="hover"
        class="w-64 cursor-pointer" :class="{ 'ring-2 ring-blue-400': filterChannelId === c.channelId }"
        @click="toggleFilter(c.channelId)">
        <div class="flex justify-between items-center">
          <span class="font-bold">{{ c.name }}</span>
          <el-tag size="small" :type="statusTag(c).type">{{ statusTag(c).text }}</el-tag>
        </div>
        <div class="text-xs text-gray-500 mt-2">
          <div>驱动：{{ c.driver }} ｜ 测点：{{ c.totalPoints }}</div>
          <div>good 率：{{ (c.goodRate * 100).toFixed(1) }}%</div>
          <div>最后上数：{{ fmtTime(c.lastTs) }}</div>
        </div>
      </el-card>
      <el-empty v-if="!cards.length" description="暂无通道，请先在采集配置页创建/导入" />
    </div>

    <!-- 实时数据表 -->
    <el-card shadow="never">
      <template #header>
        <div class="flex justify-between items-center">
          <span>实时数据{{ filterChannelId ? `（通道 ${filterChannelId}）` : '' }}</span>
          <div>
            <el-select v-model="qualityFilter" placeholder="质量" clearable size="small" class="w-28 mr-2" @change="load">
              <el-option label="good" value="good" />
              <el-option label="bad" value="bad" />
              <el-option label="timeout" value="timeout" />
            </el-select>
            <el-switch v-model="autoRefresh" active-text="自动刷新" />
          </div>
        </div>
      </template>
      <el-table :data="rows" size="small" border max-height="480">
        <el-table-column prop="channelId" label="通道" width="70" />
        <el-table-column prop="deviceId" label="设备" width="70" />
        <el-table-column prop="name" label="测点" min-width="120" />
        <el-table-column prop="value" label="值" min-width="120" />
        <el-table-column label="质量" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="row.quality === 'good' ? 'success' : row.quality === 'bad' ? 'danger' : 'warning'">
              {{ row.quality }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="errorMsg" label="错误" min-width="120" show-overflow-tooltip />
        <el-table-column label="时间" width="170">
          <template #default="{ row }">{{ fmtTime(row.ts) }}</template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import WarningBar from '@/components/warningBar/warningBar.vue'
import { getRealtime, getChannelList, getChannelStatus } from '@/api/collect'

defineOptions({ name: 'CollectMonitor' })

const cards = ref([])
const rows = ref([])
const filterChannelId = ref(0)
const qualityFilter = ref('')
const autoRefresh = ref(true)
let timer = null

const statusTag = (c) => {
  if (c.deployStatus === 'deployed' && c.running) {
    if (c.lastTs && Date.now() - c.lastTs > 300000) return { text: '疑似离线', type: 'warning' }
    return { text: '运行中', type: 'success' }
  }
  if (c.deployStatus === 'deployed') return { text: '已停止', type: 'info' }
  return { text: '未部署', type: 'info' }
}
const fmtTime = (ts) => (ts ? new Date(ts).toLocaleString() : '—')
const toggleFilter = (id) => {
  filterChannelId.value = filterChannelId.value === id ? 0 : id
  load()
}

const loadCards = async () => {
  const r = await getChannelList()
  const list = r.data || []
  const out = []
  for (const c of list) {
    try {
      const s = await getChannelStatus(c.ID)
      out.push({ ...s.data, name: c.name, driver: c.driver })
    } catch (e) { /* 单通道状态失败不阻塞 */ }
  }
  cards.value = out
}
const load = async () => {
  const params = {}
  if (filterChannelId.value) params.channelId = filterChannelId.value
  if (qualityFilter.value) params.quality = qualityFilter.value
  const r = await getRealtime(params)
  rows.value = r.data || []
}

const tick = async () => {
  if (!autoRefresh.value) return
  await Promise.all([loadCards(), load()])
}

onMounted(() => {
  tick()
  timer = setInterval(tick, 5000)
})
onBeforeUnmount(() => clearInterval(timer))
</script>
