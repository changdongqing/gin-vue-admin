<template>
  <div class="excel-preview">
    <!-- 页头 -->
    <div class="flex items-center justify-between py-2 px-2 border-b">
      <div class="flex items-center gap-2">
        <el-button icon="back" link @click="router.back()">返回</el-button>
        <span class="text-lg font-bold">报表预览：{{ reportName || reportCode }}</span>
        <el-tag v-if="truncatedNote" type="warning" size="small">{{ truncatedNote }}</el-tag>
      </div>
    </div>

    <div v-if="loadError" class="flex-1 flex items-center justify-center">
      <el-empty :description="loadError" />
    </div>

    <template v-else>
      <!-- 参数表单 -->
      <div v-if="params.length" class="px-2 pt-1">
        <param-form ref="paramFormRef" :params="params" @update="emitUpdate" />
      </div>
      <!-- 操作行 -->
      <div class="flex items-center gap-2 px-2 pb-1">
        <el-button type="primary" icon="search" :loading="loading" @click="handleQuery">查询</el-button>
        <el-button icon="refresh-left" @click="handleReset">重置</el-button>
        <el-button icon="download" :loading="exporting" @click="handleExport">导出</el-button>
        <div class="flex-1"></div>
        <el-pagination
          v-model:current-page="pagination.pageNo"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @current-change="(p) => loadData(p, pagination.pageSize)"
          @size-change="(s) => loadData(1, s)"
        />
      </div>

      <el-alert v-if="errorMsg" type="error" :title="errorMsg" class="mx-2 mb-1" :closable="true" />

      <!-- 只读表格区（Univer readonly，:key 重建加载新快照） -->
      <div class="flex-1 min-h-0 mx-2 mb-2">
        <univer-sheet :key="univerKey" :snapshot="renderedSnapshot" :readonly="true" />
      </div>
    </template>
  </div>
</template>

<script setup>
  // Excel 报表预览页（hidden 菜单 reportExcelViewer / path=reportpreview）：
  // reportCode 取值顺序：route.query.reportCode（设计器跳转）→ route.meta.queryInPath（菜单 path 内嵌 query）
  import { ref, reactive, onMounted } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import { previewExcelReport, getExcelReportParamDefs } from '@/api/report/excelPreview'
  import { getExcelReportList } from '@/api/report/excelReport'
  import ParamForm from '@/view/report/components/paramForm.vue'
  import UniverSheet from '../designer/components/UniverSheet.vue'
  import { exportSnapshotToXlsx } from './xlsxExport'

  defineOptions({ name: 'ReportExcelViewer' })

  const route = useRoute()
  const router = useRouter()

  const reportCode = ref('')
  const reportName = ref('')
  const params = ref([])
  const paramFormRef = ref(null)
  const renderedSnapshot = ref(null)
  const univerKey = ref(0)
  const loading = ref(false)
  const exporting = ref(false)
  const errorMsg = ref('')
  const loadError = ref('')
  const truncatedNote = ref('')
  const pagination = reactive({ pageNo: 1, pageSize: 100, total: 0 })

  const emitUpdate = () => {}

  const loadData = async (pageNo, pageSize) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const paramValues = paramFormRef.value ? paramFormRef.value.buildParamValues() : {}
      const res = await previewExcelReport({ reportCode: reportCode.value, paramValues, pageNo, pageSize })
      if (res.code === 0) {
        renderedSnapshot.value = res.data.snapshot
        pagination.total = res.data.total || 0
        pagination.pageNo = pageNo
        pagination.pageSize = pageSize
        univerKey.value++
      } else {
        errorMsg.value = res.msg
      }
    } finally {
      loading.value = false
    }
  }

  const handleQuery = () => loadData(1, pagination.pageSize)

  const handleReset = () => {
    paramFormRef.value?.resetValues()
    loadData(1, pagination.pageSize)
  }

  // 全量导出：pageNo=1、pageSize=50000（受 02 行数上限保护），前端 SheetJS 转 xlsx
  const handleExport = async () => {
    exporting.value = true
    try {
      const paramValues = paramFormRef.value ? paramFormRef.value.buildParamValues() : {}
      const res = await previewExcelReport({
        reportCode: reportCode.value,
        paramValues,
        pageNo: 1,
        pageSize: 50000
      })
      if (res.code === 0) {
        exportSnapshotToXlsx(res.data.snapshot, `${reportName.value || reportCode.value}.xlsx`)
        ElMessage.success('导出成功（值/合并/列宽；样式导出列后续迭代）')
      } else {
        errorMsg.value = res.msg
      }
    } finally {
      exporting.value = false
    }
  }

  onMounted(async () => {
    reportCode.value = route.query.reportCode || route.meta?.queryInPath?.reportCode || ''
    if (!reportCode.value) {
      loadError.value = '缺少报表编码（reportCode）'
      return
    }
    // page 接口按 reportCode 精确过滤（pageSize=1）校验存在 + 取名称（对齐实现：无 get-by-code）
    const pageRes = await getExcelReportList({ page: 1, pageSize: 1, reportCode: reportCode.value })
    if (pageRes.code !== 0 || !(pageRes.data?.list || []).length) {
      loadError.value = '报表不存在或已被删除'
      return
    }
    reportName.value = pageRes.data.list[0].reportName
    // 参数定义聚合
    const res = await getExcelReportParamDefs({ reportCode: reportCode.value })
    if (res.code === 0) {
      params.value = res.data || []
    }
    loadData(1, pagination.pageSize)
  })
</script>

<style lang="scss" scoped>
  .excel-preview {
    // GVA 布局仅提供 min-height：直接用视口高度（header+页签预留）
    height: calc(100vh - 150px);
    min-height: 520px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--el-bg-color);
  }
</style>
