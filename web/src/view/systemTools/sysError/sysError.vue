<template>
  <div>
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button icon="delete" :disabled="!selectedRows.length" @click="onBatchDelete(selectedRows)">
          删除
        </el-button>
      </template>

      <template #level="{ row }">
        <el-tag effect="dark" :type="levelTagMap[row.level] || 'info'">
          {{ levelLabelMap[row.level] || defaultLevelLabel }}
        </el-tag>
      </template>

      <template #status="{ row }">
        <el-tag effect="light" :type="statusTagMap[row.status] || 'info'">
          {{ statusLabelMap[row.status] || defaultStatusLabel }}
        </el-tag>
      </template>

      <template #operate="{ row }">
        <el-button v-if="row.status !== '处理中'" type="primary" link @click="getSolution(row.ID)">
          <el-icon><ai-gva /></el-icon>方案
        </el-button>
        <el-button type="primary" link icon="info-filled" @click="getDetails(row)">查看</el-button>
        <el-button type="primary" link icon="delete" @click="deleteRow(row)">删除</el-button>
      </template>
    </GvaGrid>

    <el-drawer
      destroy-on-close
      :size="appStore.drawerSize"
      v-model="detailShow"
      :show-close="true"
      :before-close="closeDetailShow"
      title="查看"
    >
      <el-descriptions :column="2" border direction="vertical">
        <el-descriptions-item label="错误来源">
          {{ detailForm.form }}
        </el-descriptions-item>
        <el-descriptions-item label="错误等级">
          <el-tag effect="dark" :type="levelTagMap[detailForm.level] || 'info'">
            {{ levelLabelMap[detailForm.level] || defaultLevelLabel }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="处理状态">
          <el-tag effect="light" :type="statusTagMap[detailForm.status] || 'info'">
            {{ statusLabelMap[detailForm.status] || defaultStatusLabel }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="错误内容" :span="2">
          <pre class="whitespace-pre-wrap break-words">{{ detailForm.info }}</pre>
        </el-descriptions-item>
        <el-descriptions-item label="解决方案" :span="2">
          <pre class="whitespace-pre-wrap break-words">{{ detailForm.solution }}</pre>
        </el-descriptions-item>
      </el-descriptions>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    deleteSysError,
    deleteSysErrorByIds,
    findSysError,
    getSysErrorList,
    getSysErrorSolution
  } from '@/api/system/sysError'

  import { ElMessage, ElMessageBox } from 'element-plus'
  import { ref } from 'vue'
  import { useAppStore } from '@/pinia'
  import GvaGrid, { useGvaGrid, useGvaGridDelete } from '@/components/gvaGrid'

  defineOptions({
    name: 'SysError'
  })

  const appStore = useAppStore()

  const { gridRef, gridOptions, gridEvents, selectedRows, refresh } = useGvaGrid({
    id: 'systemTools-sysError',
    api: getSysErrorList,
    defaultSort: null,
    checkbox: true,
    searchItems: [
      {
        field: 'createdAtRange',
        title: '创建日期',
        span: 8,
        itemRender: { name: 'gvaDateRange', props: { type: 'datetimerange', valueFormat: 'YYYY-MM-DD HH:mm:ss' } }
      },
      { field: 'form', title: '错误来源', span: 6, itemRender: { name: 'VxeInput', props: { placeholder: '搜索条件', clearable: true } } },
      { field: 'info', title: '错误内容', span: 6, itemRender: { name: 'VxeInput', props: { placeholder: '搜索条件', clearable: true } } }
    ],
    columns: [
      { field: 'CreatedAt', title: '日期', width: 180, sortable: true, cellRender: { name: 'gvaDate' } },
      { field: 'form', title: '错误来源', width: 120 },
      { field: 'level', title: '错误等级', width: 120, slots: { default: 'level' } },
      { field: 'status', title: '处理状态', width: 140, slots: { default: 'status' } },
      { field: 'info', title: '错误内容', minWidth: 240 },
      { field: 'solution', title: '解决方案', width: 120 },
      { title: '操作', fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  const { deleteRow, deleteRows: onBatchDelete } = useGvaGridDelete(gridRef, {
    delete: (rows) => {
      if (rows.length === 1) return deleteSysError({ ID: rows[0].ID })
      return deleteSysErrorByIds({ IDs: rows.map((item) => item.ID) })
    },
    confirmText: '确定要删除吗?',
    successText: '删除成功'
  })

  const getSolution = async (id) => {
    const confirmed = await ElMessageBox.confirm(
      '日志将通过 AI-PATH 传输至 GVA AI 用于错误分析，并在 GVA 官方平台短暂存储作为 AI 上下文。是否确认进行 AI 处理？（此功能仅向授权用户开放）',
      '提示(Beta)',
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning'
      }
    ).catch(() => false)
    if (!confirmed) return
    const res = await getSysErrorSolution({ id })
    if (res.code === 0) {
      ElMessage({ type: 'success', message: res.msg || '处理已提交，1分钟后完成' })
      refresh()
    }
  }

  const detailForm = ref({})

  // 查看详情控制标记
  const detailShow = ref(false)

  // 打开详情弹窗
  const openDetailShow = () => {
    detailShow.value = true
  }

  // 打开详情
  const getDetails = async (row) => {
    // 打开弹窗
    const res = await findSysError({ ID: row.ID })
    if (res.code === 0) {
      detailForm.value = res.data
      openDetailShow()
    }
  }

  // 关闭详情弹窗
  const closeDetailShow = () => {
    detailShow.value = false
    detailForm.value = {}
  }

  const statusLabelMap = {
    未处理: '未处理',
    处理中: '处理中',
    处理完成: '处理完成',
    处理失败: '处理失败'
  }
  const statusTagMap = {
    未处理: 'info',
    处理中: 'warning',
    处理完成: 'success',
    处理失败: 'danger'
  }
  const defaultStatusLabel = '未处理'

  const levelLabelMap = {
    fatal: '致命错误',
    error: '一般错误'
  }
  const levelTagMap = {
    fatal: 'danger',
    error: 'warning'
  }
  const defaultLevelLabel = '一般错误'
</script>
