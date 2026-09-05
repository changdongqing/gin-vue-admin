<template>
  <div>
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button icon="delete" :disabled="!selectedRows.length" @click="onBatchDelete(selectedRows)">
          删除
        </el-button>
      </template>

      <template #user="{ row }">
        <div>
          {{ row.user.userName }}({{ row.user.nickName }})
        </div>
      </template>

      <template #status="{ row }">
        <div>
          <el-tag type="success">{{ row.status }}</el-tag>
        </div>
      </template>

      <template #body="{ row }">
        <div>
          <el-popover v-if="row.body" placement="left-start" :width="444">
            <div class="popover-box">
              <pre>{{ fmtBody(row.body) }}</pre>
            </div>
            <template #reference>
              <el-icon style="cursor: pointer"><warning /></el-icon>
            </template>
          </el-popover>

          <span v-else>无</span>
        </div>
      </template>

      <template #resp="{ row }">
        <div>
          <el-popover v-if="row.resp" placement="left-start" :width="444">
            <div class="popover-box">
              <pre>{{ fmtBody(row.resp) }}</pre>
            </div>
            <template #reference>
              <el-icon style="cursor: pointer"><warning /></el-icon>
            </template>
          </el-popover>
          <span v-else>无</span>
        </div>
      </template>

      <template #operate="{ row }">
        <el-button icon="delete" type="primary" link @click="deleteRow(row)">删除</el-button>
      </template>
    </GvaGrid>
  </div>
</template>

<script setup>
  import {
    deleteSysOperationRecord,
    getSysOperationRecordList,
    deleteSysOperationRecordByIds
  } from '@/api/sysOperationRecord' // 此处请自行替换地址
  import GvaGrid, { useGvaGrid, useGvaGridDelete } from '@/components/gvaGrid'

  defineOptions({
    name: 'SysOperationRecord'
  })

  const { gridRef, gridOptions, gridEvents, selectedRows } = useGvaGrid({
    id: 'superAdmin-sysOperationRecord',
    api: getSysOperationRecordList,
    defaultSort: null,
    checkbox: true,
    searchItems: [
      { field: 'method', title: '请求方法', span: 6, itemRender: { name: 'VxeInput', props: { placeholder: '搜索条件', clearable: true } } },
      { field: 'path', title: '请求路径', span: 6, itemRender: { name: 'VxeInput', props: { placeholder: '搜索条件', clearable: true } } },
      { field: 'status', title: '结果状态码', span: 6, itemRender: { name: 'VxeInput', props: { placeholder: '搜索条件', clearable: true } } }
    ],
    columns: [
      { field: 'user', title: '操作人', width: 140, slots: { default: 'user' } },
      { field: 'CreatedAt', title: '日期', width: 180, cellRender: { name: 'gvaDate' } },
      { field: 'status', title: '状态码', width: 120, slots: { default: 'status' } },
      { field: 'ip', title: '请求IP', width: 120 },
      { field: 'method', title: '请求方法', width: 120 },
      { field: 'path', title: '请求路径', width: 240 },
      { field: 'body', title: '请求', width: 80, slots: { default: 'body' } },
      { field: 'resp', title: '响应', width: 80, slots: { default: 'resp' } },
      { title: '操作', minWidth: 100, slots: { default: 'operate' } }
    ]
  })

  const { deleteRow, deleteRows: onBatchDelete } = useGvaGridDelete(gridRef, {
    delete: (rows) => {
      if (rows.length === 1) return deleteSysOperationRecord({ ID: rows[0].ID })
      return deleteSysOperationRecordByIds({ ids: rows.map((item) => item.ID) })
    },
    confirmText: '确定要删除吗?',
    successText: '删除成功'
  })

  const fmtBody = (value) => {
    try {
      return JSON.parse(value)
    } catch (_) {
      return value
    }
  }
</script>

<style lang="scss">
  .table-expand {
    padding-left: 60px;
    font-size: 0;
    label {
      width: 90px;
      color: #99a9bf;
      .el-form-item {
        margin-right: 0;
        margin-bottom: 0;
        width: 50%;
      }
    }
  }
  .popover-box {
    background: #112435;
    color: #f08047;
    height: 600px;
    width: 420px;
    overflow: auto;
  }
  .popover-box::-webkit-scrollbar {
    display: none; /* Chrome Safari */
  }
</style>
