<template>
  <div class="analysis-report">
    <warning-bar title="注：分析报表为配置式多维分析（AntV S2），绑定单个数据集；聚合发生在前端，后端仅返回明细（超 5 万行截断）" />
    <el-tabs v-model="activeTab" type="card" closable @tab-remove="removeTab">
      <el-tab-pane label="报表列表" name="list">
        <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
          <template #toolbar-buttons>
            <el-button type="primary" icon="plus" @click="openForm('add')">新增分析报表</el-button>
          </template>

          <template #status="{ row }">
            <el-tag :type="row.status === 0 ? 'success' : 'info'" size="small">{{ row.status === 0 ? '启用' : '禁用' }}</el-tag>
          </template>

          <template #operate="{ row }">
            <el-button icon="edit-outline" type="primary" link @click="openDesigner(row)">设计</el-button>
            <el-button icon="view" type="primary" link @click="handlePreview(row)">预览</el-button>
            <el-button icon="copy-document" type="primary" link @click="openCopy(row)">复制</el-button>
            <el-button icon="edit" type="primary" link @click="openForm('edit', row)">编辑</el-button>
            <el-button icon="delete" type="primary" link @click="deleteRow(row)">删除</el-button>
          </template>
        </GvaGrid>
      </el-tab-pane>
      <el-tab-pane v-for="t in designerTabs" :key="t.key" :name="t.key" :label="t.title" :closable="true">
        <analysis-designer v-if="t.key === activeTab" :report-code="t.key" @close="removeTab(t.key)" />
      </el-tab-pane>
    </el-tabs>

    <!-- 元数据表单抽屉 -->
    <el-drawer v-model="formVisible" :size="appStore.drawerSize" :show-close="false">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ titleForm }}</span>
          <div>
            <el-button @click="formVisible = false">取 消</el-button>
            <el-button type="primary" @click="submitForm">确 定</el-button>
          </div>
        </div>
      </template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="报表编码" prop="reportCode">
          <el-input v-model="form.reportCode" :disabled="dialogType === 'edit'" placeholder="唯一标识，创建后不可修改" />
        </el-form-item>
        <el-form-item label="报表名称" prop="reportName">
          <el-input v-model="form.reportName" />
        </el-form-item>
        <el-form-item label="分组" prop="reportGroup">
          <el-input v-model="form.reportGroup" />
        </el-form-item>
        <el-form-item label="描述" prop="reportDesc">
          <el-input v-model="form.reportDesc" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="关联数据集" prop="setCode">
          <el-select v-model="form.setCode" filterable placeholder="单选已启用数据集">
            <el-option v-for="item in dataSetOptions" :key="item.setCode" :label="`${item.setName}（${item.setCode}）`" :value="item.setCode" />
          </el-select>
          <el-alert
            v-if="dialogType === 'edit' && form.setCode !== originSetCode"
            type="warning"
            :closable="false"
            title="更换数据集后已保存配置将失效，需重新设计"
            class="mt-1"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch v-model="form.status" :active-value="0" :inactive-value="1" active-text="启用" inactive-text="禁用" />
        </el-form-item>
      </el-form>
    </el-drawer>

    <!-- 复制弹窗 -->
    <el-dialog v-model="copyVisible" title="复制分析报表" width="460px">
      <el-form ref="copyFormRef" :model="copyForm" :rules="copyRules" label-width="100px">
        <el-form-item label="源报表">
          <el-input :model-value="copyForm.sourceReportCode" disabled />
        </el-form-item>
        <el-form-item label="新编码" prop="reportCode">
          <el-input v-model="copyForm.reportCode" />
        </el-form-item>
        <el-form-item label="新名称" prop="reportName">
          <el-input v-model="copyForm.reportName" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="copyVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitCopy">复 制</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import {
    getAnalysisReportList,
    findAnalysisReport,
    createAnalysisReport,
    updateAnalysisReport,
    deleteAnalysisReport,
    copyAnalysisReport
  } from '@/api/report/analysisReport'
  import { getDataSetAll } from '@/api/report/dataSet'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useRouter } from 'vue-router'
  import { useAppStore } from '@/pinia'
  import GvaGrid, { useGvaGrid } from '@/components/gvaGrid'
  import AnalysisDesigner from './designer/designer.vue'

  defineOptions({
    name: 'ReportAnalysis'
  })

  const appStore = useAppStore()
  const router = useRouter()

  const { gridRef, gridOptions, gridEvents, refresh } = useGvaGrid({
    id: 'report-analysisReport',
    defaultSort: null,
    api: getAnalysisReportList,
    searchItems: [
      {
        field: 'keyword',
        title: '关键字',
        span: 6,
        itemRender: { name: 'VxeInput', props: { placeholder: '报表编码或名称', clearable: true } }
      },
      {
        field: 'status',
        title: '状态',
        span: 6,
        itemRender: {
          name: 'VxeSelect',
          props: {
            placeholder: '不选则查全部',
            clearable: true,
            options: [
              { label: '启用', value: '0' },
              { label: '禁用', value: '1' }
            ]
          }
        }
      }
    ],
    columns: [
      { field: 'reportCode', title: '报表编码', width: 160 },
      { field: 'reportName', title: '报表名称', width: 160 },
      { field: 'reportGroup', title: '分组', width: 100 },
      { field: 'setCode', title: '数据集', width: 140 },
      { field: 'status', title: '状态', width: 80, slots: { default: 'status' } },
      { field: 'CreatedAt', title: '创建时间', width: 170, cellRender: { name: 'gvaDate' } },
      { title: '操作', width: 270, fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  // ── 页内 Tab ────────────────────────────────
  const activeTab = ref('list')
  const designerTabs = ref([])

  const openDesigner = (row) => {
    if (!designerTabs.value.some((t) => t.key === row.reportCode)) {
      designerTabs.value.push({ key: row.reportCode, title: `${row.reportName}-设计` })
    }
    activeTab.value = row.reportCode
  }

  const removeTab = (key) => {
    const idx = designerTabs.value.findIndex((t) => t.key === key)
    if (idx === -1) return
    designerTabs.value.splice(idx, 1)
    if (activeTab.value === key) activeTab.value = 'list'
  }

  const handlePreview = (row) => {
    router.push({ name: 'reportAnalysisPreview', query: { reportCode: row.reportCode } })
  }

  // ── 元数据表单 ────────────────────────────────
  const formVisible = ref(false)
  const titleForm = ref('新增分析报表')
  const dialogType = ref('add')
  const formRef = ref(null)
  const form = ref({})
  const originSetCode = ref('')
  const dataSetOptions = ref([])
  const rules = ref({
    reportCode: [
      { required: true, message: '请输入报表编码', trigger: 'blur' },
      { pattern: /^[a-zA-Z0-9_]{1,100}$/, message: '仅允许字母/数字/下划线（≤100字符）', trigger: 'blur' }
    ],
    reportName: [{ required: true, message: '请输入报表名称', trigger: 'blur' }],
    setCode: [{ required: true, message: '请选择关联数据集', trigger: 'change' }]
  })

  const initForm = () => {
    formRef.value && formRef.value.resetFields()
    form.value = { ID: 0, reportCode: '', reportName: '', reportGroup: '', reportDesc: '', setCode: '', status: 0 }
    originSetCode.value = ''
  }

  const openForm = async (type, row) => {
    initForm()
    dialogType.value = type
    titleForm.value = type === 'add' ? '新增分析报表' : '编辑分析报表'
    if (!dataSetOptions.value.length) {
      const res = await getDataSetAll()
      if (res.code === 0) dataSetOptions.value = res.data || []
    }
    if (type === 'edit') {
      const res = await findAnalysisReport({ ID: row.ID })
      if (res.code === 0) {
        form.value = { ...form.value, ...res.data }
        originSetCode.value = res.data.setCode
      }
      formRef.value && formRef.value.clearValidate()
    }
    formVisible.value = true
  }

  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      const isAdd = dialogType.value === 'add'
      const res = isAdd ? await createAnalysisReport(form.value) : await updateAnalysisReport(form.value)
      if (res.code === 0) {
        ElMessage({ type: 'success', message: isAdd ? '创建成功' : '更新成功' })
        await refresh()
        formVisible.value = false
      }
    })
  }

  // ── 复制 ────────────────────────────────
  const copyVisible = ref(false)
  const copyFormRef = ref(null)
  const copyForm = ref({})
  const copyRules = ref({
    reportCode: [{ required: true, message: '请输入新报表编码', trigger: 'blur' }],
    reportName: [{ required: true, message: '请输入新报表名称', trigger: 'blur' }]
  })

  const openCopy = (row) => {
    copyForm.value = { sourceReportCode: row.reportCode, reportCode: `${row.reportCode}_copy`, reportName: `${row.reportName}-副本` }
    copyVisible.value = true
  }

  const submitCopy = () => {
    copyFormRef.value.validate(async (valid) => {
      if (!valid) return
      const res = await copyAnalysisReport(copyForm.value)
      if (res.code === 0) {
        ElMessage.success('复制成功')
        copyVisible.value = false
        refresh()
      }
    })
  }

  // ── 删除（联动移除设计器 Tab）────────────────
  const deleteRow = (row) => {
    ElMessageBox.confirm(`此操作将删除分析报表「${row.reportName}」及其配置，是否继续?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        const res = await deleteAnalysisReport({ ID: row.ID })
        if (res.code === 0) {
          ElMessage({ type: 'success', message: '删除成功!' })
          removeTab(row.reportCode)
          refresh()
        }
      })
      .catch(() => {})
  }
</script>

<style lang="scss" scoped>
  .analysis-report {
    height: 100%;
    display: flex;
    flex-direction: column;

    :deep(.el-tabs) {
      flex: 1;
      display: flex;
      flex-direction: column;
      overflow: hidden;

      .el-tabs__content {
        flex: 1;
        min-height: 0;
        overflow: hidden;

        .el-tab-pane {
          height: 100%;
          overflow: auto;
        }
      }
    }
  }
</style>
