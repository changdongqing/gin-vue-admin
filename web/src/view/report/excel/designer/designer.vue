<template>
  <div class="excel-designer">
    <div v-if="loadError" class="flex items-center justify-center h-full">
      <el-empty :description="loadError" />
    </div>
    <template v-else>
      <!-- 工具栏 -->
      <div class="toolbar flex items-center gap-2 py-2 px-1">
        <el-tag type="primary" size="large">{{ reportCode }}</el-tag>
        <el-upload :show-file-list="false" accept=".xlsx" :auto-upload="false" :on-change="handleImportXlsx">
          <el-button icon="upload">导入xlsx</el-button>
        </el-upload>
        <el-button icon="view" type="primary" plain @click="handlePreview">预览</el-button>
        <el-button icon="promotion" type="primary" plain @click="handleAddToMenu">添加到菜单</el-button>
        <div class="flex-1"></div>
        <el-button icon="link" @click="openBindModal">关联数据集</el-button>
        <el-button icon="select" @click="handleInsertField">点+插入</el-button>
        <el-button icon="checked" type="primary" :loading="saving" @click="handleSave">保存</el-button>
        <el-button icon="close" @click="emit('close')">关闭</el-button>
      </div>

      <!-- 三栏：字段面板 18% ｜ Univer 62% ｜ 属性面板 20% -->
      <splitpanes class="default-theme designer-split">
        <pane :size="18" min-size="12" max-size="30">
          <dataset-panel :data-set-fields="dataSetFields" @bind="openBindModal" @insert="handleInsert" />
        </pane>
        <pane :size="62">
          <div class="univer-wrap" @dragover.prevent @drop="handleDrop">
            <univer-sheet :key="univerKey" ref="univerRef" :snapshot="initialSnapshot" @cell-select="onCellSelect" />
          </div>
        </pane>
        <pane :size="20" min-size="14" max-size="30">
          <property-panel :selected-cell="selectedCell" @update-value="handleUpdateValue" />
        </pane>
      </splitpanes>
    </template>

    <data-set-select-modal ref="bindModalRef" :report-code="reportCode" :bound-set-codes="boundSetCodes" @success="refreshDataSetFields" />
    <add-to-menu-modal ref="menuModalRef" :report-code="reportCode" :report-name="reportName" />
  </div>
</template>

<script setup>
  // 设计器主页（被列表页 el-tabs 承载）：加载报表 → 初始快照 → 字段列表 → 保存/绑定/导入/预览/加菜单
  import { ref, computed, onMounted } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useRouter } from 'vue-router'
  import { Splitpanes, Pane } from 'splitpanes'
  import 'splitpanes/dist/splitpanes.css'
  import { getExcelReportList, findExcelReport, saveExcelTemplate, getExcelReportDataSetFields } from '@/api/report/excelReport'
  import UniverSheet from './components/UniverSheet.vue'
  import DatasetPanel from './components/DatasetPanel.vue'
  import DataSetSelectModal from './components/DataSetSelectModal.vue'
  import PropertyPanel from './components/PropertyPanel.vue'
  import AddToMenuModal from './components/AddToMenuModal.vue'
  import { importXlsxToSnapshot, createEmptySnapshot } from './utils/xlsxImport'

  const props = defineProps({
    reportCode: { type: String, required: true }
  })
  const emit = defineEmits(['close'])

  const router = useRouter()

  const univerRef = ref(null)
  const initialSnapshot = ref(null)
  const univerKey = ref(0)
  const dataSetFields = ref([])
  const selectedCell = ref(null)
  const loadError = ref('')
  const saving = ref(false)
  const reportName = ref('')
  const bindModalRef = ref(null)
  const menuModalRef = ref(null)

  const boundSetCodes = computed(() => (dataSetFields.value || []).map((d) => d.setCode))

  const onCellSelect = (cell) => {
    selectedCell.value = cell
  }

  const handleUpdateValue = (value) => {
    if (!selectedCell.value) return
    univerRef.value?.setCellValue(selectedCell.value.row, selectedCell.value.col, value)
  }

  // 字段插入：有选区插选中格，无选区默认 A1
  const insertField = (setCode, fieldName) => {
    const sel = univerRef.value?.getSelection()
    const target = sel || { row: 0, col: 0 }
    univerRef.value?.setCellValue(target.row, target.col, `#{${setCode}.${fieldName}}`)
    ElMessage.success(`已插入 ${target.row === 0 && target.col === 0 && !sel ? 'A1' : target.row + 1} 行`)
  }

  const handleInsert = (setCode, fieldName) => insertField(setCode, fieldName)

  const handleDrop = (event) => {
    const raw = event.dataTransfer.getData('application/json')
    if (!raw) return
    try {
      const { setCode, fieldName } = JSON.parse(raw)
      insertField(setCode, fieldName)
    } catch (e) {
      /* 忽略非法拖拽数据 */
    }
  }

  const handleSave = async () => {
    saving.value = true
    try {
      const snapshot = univerRef.value?.getSnapshot()
      if (!snapshot) {
        ElMessage.error('表格尚未就绪')
        return
      }
      const res = await saveExcelTemplate({ reportCode: props.reportCode, jsonStr: JSON.stringify(snapshot) })
      if (res.code === 0) {
        ElMessage.success('模板已保存')
      }
    } finally {
      saving.value = false
    }
  }

  const handlePreview = () => {
    router.push({ name: 'reportExcelViewer', query: { reportCode: props.reportCode } })
  }

  const openBindModal = () => {
    bindModalRef.value?.open()
  }

  const handleAddToMenu = () => {
    menuModalRef.value?.open()
  }

  const handleImportXlsx = async (uploadFile) => {
    const file = uploadFile?.raw
    if (!file) return
    try {
      const snapshot = await importXlsxToSnapshot(file)
      initialSnapshot.value = snapshot
      univerKey.value++ // 父组件 :key 重建 Univer（组件内不 watch 快照）
      ElMessage.success('导入成功（值/合并/列宽），样式迁移为后续迭代')
    } catch (e) {
      ElMessage.error('导入失败：' + (e?.message || '文件解析异常'))
    }
  }

  const refreshDataSetFields = async () => {
    const res = await getExcelReportDataSetFields({ reportCode: props.reportCode })
    if (res.code === 0) {
      dataSetFields.value = res.data || []
    }
  }

  onMounted(async () => {
    // page 接口 reportCode 精确过滤（pageSize=1）校验存在 + 取名称（对齐实现：无 get-by-code）
    const pageRes = await getExcelReportList({ page: 1, pageSize: 1, reportCode: props.reportCode })
    if (pageRes.code !== 0 || !(pageRes.data?.list || []).length) {
      loadError.value = '报表不存在或已被删除'
      return
    }
    reportName.value = pageRes.data.list[0].reportName
    const res = await findExcelReport({ ID: pageRes.data.list[0].ID })
    if (res.code !== 0) {
      loadError.value = '报表详情加载失败'
      return
    }
    try {
      initialSnapshot.value = res.data.jsonStr ? JSON.parse(res.data.jsonStr) : createEmptySnapshot()
    } catch (e) {
      ElMessage.error('模板 JSON 解析失败，已加载空白工作簿')
      initialSnapshot.value = createEmptySnapshot()
    }
    univerKey.value++
    await refreshDataSetFields()
  })
</script>

<style lang="scss" scoped>
  .excel-designer {
    // GVA 布局仅提供 min-height，百分比链不成立：直接用视口高度（header+页签+工具栏预留）
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

    .univer-wrap {
      height: 100%;
      overflow: hidden;
    }
  }
</style>
