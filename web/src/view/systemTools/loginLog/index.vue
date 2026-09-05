<template>
  <div>
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button icon="delete" :disabled="!selectedRows.length" @click="onBatchDelete(selectedRows)">
          删除
        </el-button>
      </template>

      <template #status="{ row }">
        <el-tag :type="row.status ? 'success' : 'danger'">
          {{ row.status ? '成功' : '失败' }}
        </el-tag>
      </template>

      <template #detail="{ row }">
        {{ row.status ? '登录成功' : row.errorMessage }}
      </template>

      <template #loginTime="{ row }">
        {{ formatDate(row.CreatedAt) }}
      </template>

      <template #operate="{ row }">
        <el-popover v-model:visible="row.visible" placement="top" width="160">
          <p>确定要删除吗？</p>
          <div style="text-align: right; margin: 0">
            <el-button size="small" type="primary" link @click="row.visible = false">取消</el-button>
            <el-button size="small" type="primary" @click="deleteLoginLogRow(row)">确定</el-button>
          </div>
          <template #reference>
            <el-button icon="delete" type="primary" link @click="row.visible = true">删除</el-button>
          </template>
        </el-popover>
      </template>
    </GvaGrid>
  </div>
</template>

<script setup>
import {
  getLoginLogList,
  deleteLoginLog,
  deleteLoginLogByIds
} from '@/api/sysLoginLog'
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { formatDate } from '@/utils/format'
import GvaGrid, { useGvaGrid, useGvaGridDelete } from '@/components/gvaGrid'

const { gridRef, gridOptions, gridEvents, selectedRows } = useGvaGrid({
  id: 'systemTools-loginLog',
  api: getLoginLogList,
  defaultSort: null,
  checkbox: true,
  searchItems: [
    { field: 'username', title: '用户名', span: 6, itemRender: { name: 'VxeInput', props: { placeholder: '搜索用户名', clearable: true } } },
    {
      field: 'status',
      title: '状态',
      span: 6,
      itemRender: {
        name: 'VxeSelect',
        props: {
          placeholder: '请选择',
          clearable: true,
          options: [
            { label: '成功', value: true },
            { label: '失败', value: false }
          ]
        }
      }
    }
  ],
  columns: [
    { field: 'ID', title: 'ID', width: 80 },
    { field: 'username', title: '用户名', width: 150 },
    { field: 'ip', title: '登录IP', width: 150 },
    { field: 'status', title: '状态', width: 100, slots: { default: 'status' } },
    { field: 'detail', title: '详情', minWidth: 150, slots: { default: 'detail' } },
    { field: 'agent', title: '浏览器/设备', minWidth: 150 },
    { field: 'loginTime', title: '登录时间', width: 180, slots: { default: 'loginTime' } },
    { title: '操作', width: 120, slots: { default: 'operate' } }
  ]
})

const { deleteRows: onBatchDelete } = useGvaGridDelete(gridRef, {
  delete: (rows) => deleteLoginLogByIds({ ids: rows.map((item) => item.ID) }),
  confirmText: '确定要删除吗?',
  successText: '删除成功'
})

const deleteLoginLogRow = async (row) => {
  row.visible = false
  const res = await deleteLoginLog(row)
  if (res.code === 0) {
    ElMessage({
      type: 'success',
      message: '删除成功'
    })
    gridRef.value.commitProxy('query')
  }
}
</script>

<style scoped>
</style>
