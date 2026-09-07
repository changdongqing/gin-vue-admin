<template>
  <div class="analysis-preview">
    <!-- 页头 -->
    <div class="flex items-center justify-between py-2 px-2 border-b">
      <div class="flex items-center gap-2">
        <el-button icon="back" link @click="router.back()">返回</el-button>
        <span class="text-lg font-bold">分析报表：{{ reportName || reportCode }}</span>
        <el-tag v-if="status === 1" type="warning" size="small">报表已禁用</el-tag>
      </div>
    </div>

    <div v-if="loadError" class="flex-1 flex items-center justify-center">
      <el-empty :description="loadError" />
    </div>

    <template v-else>
      <el-alert v-if="truncated" type="warning" title="数据超过 5 万行已截断展示，完整分析建议收窄参数或 SQL 预聚合" class="mx-2 mt-1" :closable="true" />

      <!-- 参数栏（复用设计器 ParamBar；改参刷新） -->
      <div class="mx-2">
        <param-bar ref="paramBarRef" :params="params" @reload="loadData" />
      </div>

      <!-- S2 渲染区 -->
      <div class="flex-1 min-h-0 mx-2 mb-2">
        <s2-sheet :sheet-type="config?.sheetType || 'pivot'" :data-cfg="dataCfg" :options="options" :theme-cfg="themeCfg" :loading="loading" />
      </div>
    </template>
  </div>
</template>

<script setup>
  // 分析报表预览页（hidden 菜单 reportAnalysisPreview / path=preview）：
  // reportCode 取值顺序：route.query.reportCode → route.meta.queryInPath（菜单 path 内嵌 query）
  import { ref, computed, onMounted } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { getAnalysisReportByCode, previewAnalysisReport } from '@/api/report/analysisReport'
  import { buildDataCfg, buildOptions, buildThemeCfg, loadConfig } from '../designer/utils/s2Config'
  import S2Sheet from '../designer/components/S2Sheet.vue'
  import ParamBar from '../designer/components/ParamBar.vue'

  defineOptions({ name: 'ReportAnalysisPreview' })

  const route = useRoute()
  const router = useRouter()

  const reportCode = ref('')
  const reportName = ref('')
  const status = ref(0)
  const params = ref([])
  const paramBarRef = ref(null)
  const rawRows = ref([])
  const loading = ref(false)
  const truncated = ref(false)
  const loadError = ref('')
  const loaded = ref(false)

  const config = computed(() => loaded.value ? loadConfig(configJson.value) : null)
  const configJson = ref('')

  const dataCfg = computed(() => (loaded.value ? buildDataCfg(config.value, rawRows.value) : null))
  const options = computed(() => (loaded.value ? buildOptions(config.value) : null))
  const themeCfg = computed(() => (loaded.value ? buildThemeCfg(config.value) : null))

  const loadData = async () => {
    loading.value = true
    try {
      const paramValues = paramBarRef.value ? paramBarRef.value.buildParamValues() : {}
      const res = await previewAnalysisReport({ reportCode: reportCode.value, paramValues })
      if (res.code === 0) {
        rawRows.value = res.data.rows || []
        truncated.value = !!res.data.truncated
      }
    } finally {
      loading.value = false
    }
  }

  onMounted(async () => {
    reportCode.value = route.query.reportCode || route.meta?.queryInPath?.reportCode || ''
    if (!reportCode.value) {
      loadError.value = '缺少报表编码（reportCode）'
      return
    }
    const res = await getAnalysisReportByCode({ reportCode: reportCode.value })
    if (res.code !== 0) {
      loadError.value = '报表不存在或已被删除'
      return
    }
    if (res.data.status === 1) {
      status.value = 1
    }
    reportName.value = res.data.reportName
    params.value = res.data.params || []
    configJson.value = res.data.configJson || ''
    loaded.value = true
    await loadData()
  })
</script>

<style lang="scss" scoped>
  .analysis-preview {
    height: calc(100vh - 150px);
    min-height: 520px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--el-bg-color);
  }
</style>
