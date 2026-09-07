<template>
  <div class="analysis-designer">
    <div v-if="loadError" class="flex items-center justify-center h-full">
      <el-empty :description="loadError" />
    </div>
    <template v-else>
      <!-- 工具栏 -->
      <div class="toolbar flex items-center gap-2 py-2 px-1">
        <el-tag type="primary" size="large">{{ reportCode }}</el-tag>
        <span class="text-sm text-gray-500">{{ reportName }}{{ setCode ? ` · ${setName}` : '' }}</span>
        <el-tag v-if="truncated" type="warning" size="small">数据超 5 万行已截断，建议 SQL 预聚合</el-tag>
        <el-button icon="promotion" type="primary" plain @click="handleAddToMenu">添加到菜单</el-button>
        <div class="flex-1"></div>
        <el-button icon="refresh" @click="reloadData">刷新数据</el-button>
        <el-button icon="checked" type="primary" :loading="saving" @click="handleSave">保存</el-button>
        <el-button icon="close" @click="emit('close')">关闭</el-button>
      </div>

      <!-- 参数栏 -->
      <param-bar ref="paramBarRef" :params="params" @reload="reloadData" />

      <!-- 三栏：字段面板 18% ｜ S2 预览 60% ｜ 配置面板 22% -->
      <splitpanes class="default-theme designer-split" @resized="onResize">
        <pane :size="18" min-size="12" max-size="30">
          <field-panel
            :field-meta="fieldMeta"
            :config="config"
            :set-code="setCode"
            :data-set-name="setName"
            @add-row="addToRows"
            @add-column="addToColumns"
            @add-value="addToValues"
          />
        </pane>
        <pane :size="60">
          <div class="s2-wrap" @dragover.prevent @drop="handleDrop">
            <s2-sheet :sheet-type="config.sheetType" :data-cfg="dataCfg" :options="options" :theme-cfg="themeCfg" :loading="loading" />
          </div>
        </pane>
        <pane :size="22" min-size="16" max-size="32">
          <config-panel :config="config" @set-sheet-type="setSheetType" @set-value-in-row="setValueInRow" @reset-page="resetPage" />
        </pane>
      </splitpanes>
    </template>

    <add-to-menu-modal
      ref="menuModalRef"
      :report-code="reportCode"
      :report-name="reportName"
      preview-component="view/report/analysis/preview/preview.vue"
      preview-path="preview"
    />
  </div>
</template>

<script setup>
  // 分析报表设计器（被列表页 el-tabs 承载）：配置即渲染 —— config 改动经 computed 派生 S2 props，
  // 不手动重建实例；rawRows 用 shallowRef（5 万行深代理会拖垮遍历并可能破坏 S2 内部 instanceof）
  import { ref, computed, shallowRef, onMounted } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useRouter } from 'vue-router'
  import { Splitpanes, Pane } from 'splitpanes'
  import 'splitpanes/dist/splitpanes.css'
  import { getAnalysisReportByCode, saveAnalysisConfig, previewAnalysisReport } from '@/api/report/analysisReport'
  import { buildDataCfg, buildOptions, buildThemeCfg, loadConfig } from './utils/s2Config'
  import S2Sheet from './components/S2Sheet.vue'
  import FieldPanel from './components/FieldPanel.vue'
  import ConfigPanel from './components/ConfigPanel.vue'
  import ParamBar from './components/ParamBar.vue'
  import AddToMenuModal from '@/view/report/excel/designer/components/AddToMenuModal.vue'

  const props = defineProps({
    reportCode: { type: String, required: true }
  })
  const emit = defineEmits(['close'])

  const router = useRouter()

  const config = ref(loadConfig(''))
  const fieldMeta = ref([]) // [{name, type}]
  const rawRows = shallowRef([]) // 后端明细（shallowRef：深代理会拖垮大数组遍历）
  const total = ref(0)
  const truncated = ref(false)
  const loading = ref(false)
  const saving = ref(false)
  const loadError = ref('')
  const reportName = ref('')
  const setCode = ref('')
  const setName = ref('')
  const params = ref([])
  const paramBarRef = ref(null)
  const menuModalRef = ref(null)

  // S2 props 为 computed 派生 —— 配置改动即重渲染
  const dataCfg = computed(() => (fieldMeta.value.length ? buildDataCfg(config.value, rawRows.value) : null))
  const options = computed(() => buildOptions(config.value))
  const themeCfg = computed(() => buildThemeCfg(config.value))

  // 分页 current 重置联动：改 pageSize/行头/列头/数值 → current=1（ConfigPanel emit）
  const resetPage = () => {
    config.value.options.pagination.current = 1
  }

  const setSheetType = (type) => {
    config.value.sheetType = type
    resetPage()
  }
  const setValueInRow = (v) => {
    config.value.fields.valueInRow = v
  }

  const addToRows = (field) => {
    if (!config.value.fields.rows.includes(field)) config.value.fields.rows.push(field)
    resetPage()
  }
  const addToColumns = (field) => {
    if (!config.value.fields.columns.includes(field)) config.value.fields.columns.push(field)
    resetPage()
  }
  const addToValues = (field) => {
    if (!config.value.fields.values.some((v) => v.field === field)) {
      config.value.fields.values.push({ field, aggregation: 'SUM', alias: '', format: '' })
    }
    resetPage()
  }

  // 左栏拖拽落点 = 右栏对应列表头部（Canvas 无 DOM 落点）：拖入字段默认进行头
  const handleDrop = (event) => {
    const raw = event.dataTransfer.getData('application/json')
    if (!raw) return
    try {
      const { field } = JSON.parse(raw)
      const meta = fieldMeta.value.find((f) => f.name === field)
      if (meta?.type === 'number') addToValues(field)
      else addToRows(field)
    } catch (e) {
      /* 忽略非法拖拽数据 */
    }
  }

  const reloadData = async () => {
    loading.value = true
    try {
      const paramValues = paramBarRef.value ? paramBarRef.value.buildParamValues() : {}
      const res = await previewAnalysisReport({ reportCode: props.reportCode, paramValues })
      if (res.code === 0) {
        fieldMeta.value = res.data.columns || []
        rawRows.value = res.data.rows || []
        total.value = res.data.total || 0
        truncated.value = !!res.data.truncated
      } else {
        ElMessage.error(res.msg)
      }
    } finally {
      loading.value = false
    }
  }

  const handleSave = async () => {
    if (!config.value.fields.values.length) {
      ElMessage.warning('请先配置数值列')
      return
    }
    saving.value = true
    try {
      const res = await saveAnalysisConfig({ reportCode: props.reportCode, configJson: JSON.stringify(config.value) })
      if (res.code === 0) {
        ElMessage.success('配置已保存')
      }
    } finally {
      saving.value = false
    }
  }

  const handleAddToMenu = () => menuModalRef.value?.open()

  const onResize = () => {
    /* splitpanes 拖动后 S2 自适应（layoutWidthType=adaptive） */
  }

  onMounted(async () => {
    const res = await getAnalysisReportByCode({ reportCode: props.reportCode })
    if (res.code !== 0) {
      loadError.value = '报表不存在或已被删除'
      return
    }
    reportName.value = res.data.reportName
    setCode.value = res.data.setCode
    setName.value = res.data.setName || res.data.setCode
    params.value = res.data.params || []
    config.value = loadConfig(res.data.configJson)
    await reloadData()
  })
</script>

<style lang="scss" scoped>
  .analysis-designer {
    height: calc(100vh - 268px);
    min-height: 520px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    padding: 0 8px 8px;
    background: var(--el-bg-color);

    .toolbar {
      border-bottom: 1px solid var(--el-border-color-lighter);
      flex-wrap: wrap;
    }

    .designer-split {
      flex: 1;
      min-height: 0;
    }

    .s2-wrap {
      height: 100%;
      overflow: hidden;
    }
  }
</style>
