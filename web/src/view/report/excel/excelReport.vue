<template>
  <div class="excel-report">
    <warning-bar title="注：Excel 报表为画布式模板（Univer 在线设计器），拖拽字段写入 #{数据集.字段} 占位符；预览与导出按占位符渲染数据" />
    <!-- 固定列表 Tab + 设计器页内 Tab（key=reportCode，可多开/关闭/删除联动） -->
    <el-tabs v-model="activeTab" type="card" closable @tab-remove="removeTab">
      <el-tab-pane label="报表列表" name="list">
        <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
          <template #toolbar-buttons>
            <el-button type="primary" icon="plus" @click="openForm('add')">新增报表</el-button>
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
        <excel-designer v-if="t.key === activeTab" :report-code="t.key" @close="removeTab(t.key)" />
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
          <el-input v-model="form.reportGroup" placeholder="如 生产报表" />
        </el-form-item>
        <el-form-item label="描述" prop="reportDesc">
          <el-input v-model="form.reportDesc" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
    </el-drawer>

    <!-- 复制弹窗 -->
    <el-dialog v-model="copyVisible" title="复制报表" width="460px">
      <el-form ref="copyFormRef" :model="copyForm" :rules="copyRules" label-width="100px">
        <el-form-item label="源报表">
          <el-input :model-value="copyForm.sourceReportCode" disabled />
        </el-form-item>
        <el-form-item label="新编码" prop="reportCode">
          <el-input v-model="copyForm.reportCode" placeholder="新报表编码" />
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
    getExcelReportList,
    findExcelReport,
    createExcelReport,
    updateExcelReport,
    deleteExcelReport,
    copyExcelReport
  } from '@/api/report/excelReport'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useRouter } from 'vue-router'
  import { useAppStore } from '@/pinia'
  import GvaGrid, { useGvaGrid } from '@/components/gvaGrid'
  import ExcelDesigner from './designer/designer.vue'

  defineOptions({
    name: 'ReportExcel'
  })

  const appStore = useAppStore()
  const router = useRouter()

  const { gridRef, gridOptions, gridEvents, refresh } = useGvaGrid({
    id: 'report-excelReport',
    defaultSort: null,
    api: getExcelReportList,
    searchItems: [
      {
        field: 'keyword',
        title: '关键字',
        span: 6,
        itemRender: { name: 'VxeInput', props: { placeholder: '报表编码或名称', clearable: true } }
      },
      {
        field: 'reportGroup',
        title: '分组',
        span: 6,
        itemRender: { name: 'VxeInput', props: { placeholder: '报表分组', clearable: true } }
      }
    ],
    columns: [
      { field: 'reportCode', title: '报表编码', width: 160 },
      { field: 'reportName', title: '报表名称', width: 160 },
      { field: 'reportGroup', title: '分组', width: 110 },
      { field: 'reportDesc', title: '描述', minWidth: 160, showOverflow: true },
      { field: 'CreatedAt', title: '创建时间', width: 170, cellRender: { name: 'gvaDate' } },
      { title: '操作', width: 270, fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  // ── 页内 Tab（设计器承载）──────────────────
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
    if (activeTab.value === key) {
      activeTab.value = 'list'
    }
  }

  const handlePreview = (row) => {
    router.push({ name: 'reportExcelViewer', query: { reportCode: row.reportCode } })
  }

  // ── 元数据表单 ────────────────────────────────
  const formVisible = ref(false)
  const titleForm = ref('新增报表')
  const dialogType = ref('add')
  const formRef = ref(null)
  const form = ref({})
  const rules = ref({
    reportCode: [
      { required: true, message: '请输入报表编码', trigger: 'blur' },
      { pattern: /^[a-zA-Z0-9_]{1,100}$/, message: '仅允许字母/数字/下划线（≤100字符）', trigger: 'blur' }
    ],
    reportName: [{ required: true, message: '请输入报表名称', trigger: 'blur' }]
  })

  const initForm = () => {
    formRef.value && formRef.value.resetFields()
    form.value = { ID: 0, reportCode: '', reportName: '', reportGroup: '', reportDesc: '' }
  }

  const openForm = async (type, row) => {
    initForm()
    dialogType.value = type
    titleForm.value = type === 'add' ? '新增报表' : '编辑报表'
    if (type === 'edit') {
      const res = await findExcelReport({ ID: row.ID })
      if (res.code === 0) {
        form.value = { ...form.value, ...res.data }
      }
      formRef.value && formRef.value.clearValidate()
    }
    formVisible.value = true
  }

  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      const isAdd = dialogType.value === 'add'
      const res = isAdd ? await createExcelReport(form.value) : await updateExcelReport(form.value)
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
      const res = await copyExcelReport(copyForm.value)
      if (res.code === 0) {
        ElMessage.success('复制成功')
        copyVisible.value = false
        refresh()
      }
    })
  }

  // ── 删除（联动移除设计器 Tab）────────────────
  const deleteRow = (row) => {
    ElMessageBox.confirm(`此操作将删除报表「${row.reportName}」及其模板内容，是否继续?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        const res = await deleteExcelReport({ ID: row.ID })
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
  .excel-report {
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
