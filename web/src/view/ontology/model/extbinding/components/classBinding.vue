<template>
  <div class="class-binding">
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="$emit('create')">新建绑定</el-button>
      </template>

      <template #classCell="{ row }">
        {{ row.classLocalName }}（{{ row.classLabelCn || '-' }}）
      </template>

      <template #statusTag="{ row }">
        <el-tag v-if="row.bindingStatus === 1" type="success">生效</el-tag>
        <el-tag v-else-if="row.bindingStatus === 2" type="danger">停用</el-tag>
        <el-tag v-else type="info">草稿</el-tag>
      </template>

      <template #countCell="{ row }">
        属性 {{ row.propertyCount }} / 子表 {{ row.detailCount }}
      </template>

      <template #syncTime="{ row }">
        <span v-if="row.lastSyncTime">{{ row.lastSyncTime }}</span>
        <span v-else>未同步</span>
      </template>

      <template #operate="{ row }">
        <el-button icon="edit" type="primary" link :disabled="row.bindingStatus === 1" @click="$emit('edit', row)">编辑</el-button>
        <el-button v-if="row.bindingStatus !== 1" icon="circle-check" type="primary" link @click="activate(row)">生效</el-button>
        <el-button v-if="row.bindingStatus === 1" icon="turn-off" type="warning" link @click="deactivate(row)">停用</el-button>
        <el-button icon="delete" type="primary" link @click="deleteRow(row)">删除</el-button>
      </template>
    </GvaGrid>
  </div>
</template>

<script setup>
  import { getExtBindingList, changeExtBindingStatus, deleteExtBinding } from '@/api/ontology/extBinding'
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import GvaGrid, { useGvaGrid } from '@/components/gvaGrid'

  defineOptions({ name: 'ClassBinding' })

  const emit = defineEmits(['create', 'edit'])

  const projectOptions = ref([])

  const { gridRef, gridOptions, gridEvents, refresh } = useGvaGrid({
    id: 'ontology-extBinding',
    defaultSort: null,
    api: getExtBindingList,
    searchItems: [
      {
        field: 'projectId',
        title: '项目',
        span: 6,
        itemRender: { name: 'VxeSelect', props: { placeholder: '全部', clearable: true, options: projectOptions } }
      },
      {
        field: 'bindingStatus',
        title: '状态',
        span: 6,
        itemRender: {
          name: 'VxeSelect',
          props: {
            placeholder: '全部',
            clearable: true,
            options: [
              { label: '草稿', value: 0 },
              { label: '生效', value: 1 },
              { label: '停用', value: 2 }
            ]
          }
        }
      }
    ],
    columns: [
      { field: 'projectCode', title: '项目', width: 110 },
      { field: 'classLocalName', title: '本体类', width: 200, slots: { default: 'classCell' } },
      { field: 'tableName', title: '业务表', width: 150 },
      { field: 'bindingStatus', title: '状态', width: 90, slots: { default: 'statusTag' } },
      { field: 'syncMode', title: '同步模式', width: 100, formatter: ({ cellValue }) => (cellValue === 2 ? '定时+手动' : '手动') },
      { field: 'propertyCount', title: '属性/子表', width: 130, slots: { default: 'countCell' } },
      { field: 'lastSyncTime', title: '最近同步', width: 170, slots: { default: 'syncTime' } },
      { field: 'lastSyncSummary', title: '结果摘要', minWidth: 180, showOverflow: true },
      { title: '操作', width: 220, fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  const setProjectOptions = (opts) => {
    projectOptions.value = opts
  }
  defineExpose({ refresh, setProjectOptions })

  const activate = (row) => {
    changeExtBindingStatus({ ID: row.ID, status: 1 }).then((res) => {
      if (res.code === 0) {
        ElMessage.success('绑定已生效')
        refresh()
      }
    })
  }
  const deactivate = (row) => {
    changeExtBindingStatus({ ID: row.ID, status: 2 }).then((res) => {
      if (res.code === 0) {
        ElMessage.success('绑定已停用')
        refresh()
      }
    })
  }
  const deleteRow = (row) => {
    ElMessageBox.confirm('删除绑定？同步日志将保留', '提示', { type: 'warning' })
      .then(async () => {
        const res = await deleteExtBinding({ ID: row.ID })
        if (res.code === 0) {
          ElMessage.success('删除成功')
          refresh()
        }
      })
      .catch(() => {})
  }
</script>
