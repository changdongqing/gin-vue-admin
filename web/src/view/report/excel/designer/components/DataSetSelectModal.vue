<template>
  <el-dialog v-model="visible" title="关联数据集" width="640px" :close-on-click-modal="false" @open="loadDataSets">
    <el-alert type="info" :closable="false" class="mb-2" title="从已启用数据集中多选；绑定仅写入关联关系，不影响已保存的模板内容" />
    <el-table ref="tableRef" :data="dataSets" v-loading="loading" border max-height="380" row-key="setCode" @selection-change="onSelectionChange">
      <el-table-column type="selection" width="42" />
      <el-table-column prop="setCode" label="编码" width="160" show-overflow-tooltip />
      <el-table-column prop="setName" label="名称" width="150" show-overflow-tooltip />
      <el-table-column prop="sourceCode" label="数据源" width="130" show-overflow-tooltip />
      <el-table-column prop="setDesc" label="描述" show-overflow-tooltip />
    </el-table>
    <template #footer>
      <el-button @click="visible = false">取 消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">绑 定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
  import { ref, nextTick } from 'vue'
  import { ElMessage } from 'element-plus'
  import { getDataSetAll } from '@/api/report/dataSet'
  import { bindExcelReportDataSets } from '@/api/report/excelReport'

  const props = defineProps({
    reportCode: { type: String, required: true },
    boundSetCodes: { type: Array, default: () => [] }
  })
  const emit = defineEmits(['success'])

  const visible = ref(false)
  const loading = ref(false)
  const submitting = ref(false)
  const dataSets = ref([])
  const selected = ref([])
  const tableRef = ref(null)

  const open = () => {
    visible.value = true
  }

  const loadDataSets = async () => {
    loading.value = true
    try {
      const res = await getDataSetAll()
      if (res.code === 0) {
        dataSets.value = res.data || []
        // 回显当前已绑定
        await nextTick()
        const bound = new Set(props.boundSetCodes)
        dataSets.value.forEach((row) => {
          if (bound.has(row.setCode)) {
            tableRef.value?.toggleRowSelection(row, true)
          }
        })
      }
    } finally {
      loading.value = false
    }
  }

  const onSelectionChange = (rows) => {
    selected.value = rows.map((r) => r.setCode)
  }

  const submit = async () => {
    submitting.value = true
    try {
      const res = await bindExcelReportDataSets({ reportCode: props.reportCode, setCodes: selected.value })
      if (res.code === 0) {
        ElMessage.success('绑定成功')
        visible.value = false
        emit('success')
      }
    } finally {
      submitting.value = false
    }
  }

  defineExpose({ open })
</script>
