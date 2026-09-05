<template>
  <div>
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="msg('新增点击')">新增</el-button>
        <el-button
          type="danger"
          icon="delete"
          :disabled="!selectedRows.length"
          @click="onDeleteRows(selectedRows)"
        >
          批量删除 ({{ selectedRows.length }})
        </el-button>
      </template>

      <template #status="{ row }">
        <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
          {{ row.status === 1 ? '启用' : '禁用' }}
        </el-tag>
      </template>

      <template #operate="{ row }">
        <el-button type="primary" link icon="edit" @click="msg('编辑 ' + row.id)">编辑</el-button>
        <el-button type="primary" link icon="delete" @click="onDeleteRow(row)">删除</el-button>
      </template>    </GvaGrid>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import GvaGrid, { useGvaGrid, useGvaGridDelete } from '@/components/gvaGrid'

defineOptions({ name: 'SmokeGrid' })

const msg = (text) => ElMessage.info(text)
const searchCount = ref(0)

// —— 假数据 + 假接口（模拟 gin-vue-admin 协议：page/pageSize/orderKey/desc → list/total）——
const ALL_ROWS = Array.from({ length: 128 }, (_, i) => ({
  ID: i + 1,
  name: `用户${String(i + 1).padStart(3, '0')}`,
  nickName: `昵称${i + 1}`,
  status: i % 3 === 0 ? 2 : 1,
  level: i % 4,
  avatar: `https://dummyimage.com/80x80/4d70ff/fff&text=${i + 1}`,
  CreatedAt: new Date(Date.now() - i * 86400000).toISOString()
}))

function fakeApi(params) {
  const { page = 1, pageSize = 10, orderKey = 'id', desc = true, ...rest } = params
  searchCount.value++
  window.__queryLog = window.__queryLog || []
  return new Promise((resolve) => {
    setTimeout(() => {
      let list = ALL_ROWS.filter((row) =>
        rest.name ? row.name.includes(rest.name) : true
      ).filter((row) => (rest.status !== undefined && rest.status !== '' && rest.status !== null ? row.status === Number(rest.status) : true))
      list = [...list].sort((a, b) => {
        const key = orderKey === 'ID' ? 'ID' : orderKey
        const va = a[key]
        const vb = b[key]
        const cmp = va > vb ? 1 : va < vb ? -1 : 0
        return desc ? -cmp : cmp
      })
      const start = (page - 1) * pageSize
      window.__queryLog.push({ page, pageSize, total: list.length, ids: list.slice(start, start + pageSize).map((r) => r.ID) })
      resolve({
        code: 0,
        data: { list: list.slice(start, start + pageSize), total: list.length, page, pageSize }
      })
    }, 300)
  })
}

const { gridRef, gridOptions, gridEvents, selectedRows } = useGvaGrid({
  id: 'smoke-grid',
  api: fakeApi,
  checkbox: true,
  defaultSort: { field: 'ID', order: 'desc' },
  searchItems: [
    { field: 'name', title: '用户名', span: 6, itemRender: { name: 'VxeInput', props: { placeholder: '请输入用户名', clearable: true } } },
    { field: 'nickName', title: '昵称', span: 6, itemRender: { name: 'VxeInput', props: { placeholder: '请输入昵称', clearable: true } } },
    {
      field: 'status',
      title: '状态',
      span: 6,
      itemRender: {
        name: 'gvaDictSelect',
        props: {
          options: [
            { label: '启用', value: 1 },
            { label: '禁用', value: 2 }
          ],
          placeholder: '请选择状态'
        }
      }
    },
    {
      field: 'createdAt',
      title: '创建时间',
      span: 6,
      itemRender: { name: 'gvaDateRange', props: { type: 'daterange' } }
    }
  ],
  columns: [
    { field: 'ID', title: 'ID', width: 80, sortable: true },
    { field: 'avatar', title: '头像', width: 72, cellRender: { name: 'gvaImage', props: { round: true } } },
    { field: 'CreatedAt', title: '创建时间', width: 170, sortable: true, cellRender: { name: 'gvaDate' } },
    { field: 'name', title: '用户名', minWidth: 120, sortable: true },
    { field: 'nickName', title: '昵称', minWidth: 120 },
    { field: 'level', title: '等级', width: 90, cellRender: { name: 'gvaDict', props: { options: [
      { label: '普通', value: 0 }, { label: '青铜', value: 1 }, { label: '白银', value: 2 }, { label: '黄金', value: 3 }
    ] } } },
    { field: 'status', title: '状态', width: 90, slots: { default: 'status' } },
    { title: '操作', fixed: 'right', slots: { default: 'operate' } }
  ]
})

const { deleteRow: onDeleteRow, deleteRows: onDeleteRows } = useGvaGridDelete(gridRef, {
  delete: (rows) => {
    // 冒烟测试：真实移除数据，用于验证“末页删空回退”
    rows.forEach((row) => {
      const idx = ALL_ROWS.findIndex((r) => r.ID === row.ID)
      if (idx >= 0) ALL_ROWS.splice(idx, 1)
    })
    return Promise.resolve({ code: 0 })
  },
  confirmText: '冒烟测试将真实移除假数据，确认继续？'
})
</script>
