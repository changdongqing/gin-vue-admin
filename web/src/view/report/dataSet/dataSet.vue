<template>
  <div class="report-data-set">
    <warning-bar title="注：数据集是报表取数的核心桥梁，支持 SQL（${param} 参数化 + <if param> 条件语法）与 HTTP 两种类型；测试预览走服务端分页" />
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="openForm('add')">新增数据集</el-button>
      </template>

      <template #setType="{ row }">
        <el-tag :type="row.setType === 'http' ? 'warning' : 'primary'" size="small">{{ setTypeLabel(row.setType) }}</el-tag>
      </template>

      <template #enableFlag="{ row }">
        <el-tag :type="row.enableFlag ? 'success' : 'info'" size="small">{{ row.enableFlag ? '启用' : '禁用' }}</el-tag>
      </template>

      <template #operate="{ row }">
        <el-button icon="edit" type="primary" link @click="openForm('edit', row)">编辑</el-button>
        <el-button icon="delete" type="primary" link @click="deleteRow(row)">删除</el-button>
      </template>
    </GvaGrid>

    <!-- 编辑抽屉（四 Tab：基础信息/参数配置/数据转换/测试预览） -->
    <data-set-form v-model="formVisible" :type="dialogType" :row-id="rowId" @saved="refresh()" />
  </div>
</template>

<script setup>
  import { getDataSetList, deleteDataSet, setTypeLabel } from '@/api/report/dataSet'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import GvaGrid, { useGvaGrid } from '@/components/gvaGrid'
  import DataSetForm from './components/dataSetForm.vue'

  defineOptions({
    name: 'ReportDataSet'
  })

  const { gridRef, gridOptions, gridEvents, refresh } = useGvaGrid({
    id: 'report-dataSet',
    defaultSort: null,
    api: getDataSetList,
    searchItems: [
      {
        field: 'keyword',
        title: '关键字',
        span: 6,
        itemRender: { name: 'VxeInput', props: { placeholder: '数据集编码或名称', clearable: true } }
      },
      {
        field: 'setType',
        title: '类型',
        span: 6,
        itemRender: {
          name: 'VxeSelect',
          props: {
            placeholder: '不选则查全部',
            clearable: true,
            options: [
              { label: 'SQL', value: 'sql' },
              { label: 'HTTP', value: 'http' }
            ]
          }
        }
      },
      {
        field: 'enableFlag',
        title: '状态',
        span: 6,
        itemRender: {
          name: 'VxeSelect',
          props: {
            placeholder: '不选则查全部',
            clearable: true,
            options: [
              { label: '启用', value: 'true' },
              { label: '禁用', value: 'false' }
            ]
          }
        }
      }
    ],
    columns: [
      { field: 'setCode', title: '数据集编码', width: 150 },
      { field: 'setName', title: '数据集名称', width: 150 },
      { field: 'setType', title: '类型', width: 80, slots: { default: 'setType' } },
      { field: 'sourceCode', title: '数据源', width: 120 },
      { field: 'setDesc', title: '描述', minWidth: 160, showOverflow: true },
      { field: 'enableFlag', title: '状态', width: 80, slots: { default: 'enableFlag' } },
      { field: 'CreatedAt', title: '创建时间', width: 170, cellRender: { name: 'gvaDate' } },
      { title: '操作', width: 140, fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  const formVisible = ref(false)
  const dialogType = ref('add')
  const rowId = ref(0)

  const openForm = (type, row) => {
    dialogType.value = type
    rowId.value = row ? row.ID : 0
    formVisible.value = true
  }

  const deleteRow = (row) => {
    ElMessageBox.confirm(`此操作将删除数据集「${row.setName}」及其参数与转换配置，是否继续?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        const res = await deleteDataSet({ ID: row.ID })
        if (res.code === 0) {
          ElMessage({ type: 'success', message: '删除成功!' })
          refresh()
        }
      })
      .catch(() => {})
  }
</script>
