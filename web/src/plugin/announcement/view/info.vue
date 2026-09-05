<template>
  <div>
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="openDialog">新增</el-button>
        <el-button icon="delete" :disabled="!selectedRows.length" @click="onDelete(selectedRows)">删除</el-button>
      </template>

      <template #userID="{ row }">
        <span>{{ filterDataSource(dataSource.userID, row.userID) }}</span>
      </template>

      <template #attachments="{ row }">
        <div class="file-list">
          <el-tag v-for="file in row.attachments" :key="file.uid" @click="downloadFile(file.url)">
            {{ file.name }}
          </el-tag>
        </div>
      </template>

      <template #operate="{ row }">
        <el-button type="primary" link icon="edit" @click="updateInfoFunc(row)">变更</el-button>
        <el-button type="primary" link icon="delete" @click="deleteRow(row)">删除</el-button>
      </template>
    </GvaGrid>
    <el-drawer
      v-model="dialogFormVisible"
      destroy-on-close
      size="800"
      :show-close="false"
      :before-close="closeDialog"
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ type === 'create' ? '添加' : '修改' }}</span>
          <div>
            <el-button type="primary" @click="enterDialog"> 确 定 </el-button>
            <el-button @click="closeDialog"> 取 消 </el-button>
          </div>
        </div>
      </template>

      <el-form
        ref="elFormRef"
        :model="formData"
        label-position="top"
        :rules="rule"
        label-width="80px"
      >
        <el-form-item label="标题:" prop="title">
          <el-input
            v-model="formData.title"
            :clearable="true"
            placeholder="请输入标题"
          />
        </el-form-item>
        <el-form-item label="内容:" prop="content">
          <RichEdit v-model="formData.content" />
        </el-form-item>
        <el-form-item label="作者:" prop="userID">
          <el-select
            v-model="formData.userID"
            placeholder="请选择作者"
            style="width: 100%"
            :clearable="true"
          >
            <el-option
              v-for="(item, key) in dataSource.userID"
              :key="key"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="附件:" prop="attachments">
          <SelectFile v-model="formData.attachments" />
        </el-form-item>
      </el-form>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    getInfoDataSource,
    createInfo,
    deleteInfo,
    deleteInfoByIds,
    updateInfo,
    findInfo,
    getInfoList
  } from '@/plugin/announcement/api/info'
  import { getUrl } from '@/utils/image'
  // 富文本组件
  import RichEdit from '@/components/richtext/rich-edit.vue'
  // 文件选择组件
  import SelectFile from '@/components/selectFile/selectFile.vue'

  // 全量引入格式化工具 请按需保留
  import { formatDate, filterDataSource } from '@/utils/format'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { ref, reactive } from 'vue'
  import GvaGrid, { useGvaGrid, useGvaGridDelete } from '@/components/gvaGrid'

  defineOptions({
    name: 'Info'
  })

  // 控制更多查询条件显示/隐藏状态
  const showAllQuery = ref(false)

  // 自动化生成的字典（可能为空）以及字段
  const formData = ref({
    title: '',
    content: '',
    userID: undefined,
    attachments: []
  })
  const dataSource = ref([])
  const getDataSourceFunc = async () => {
    const res = await getInfoDataSource()
    if (res.code === 0) {
      dataSource.value = res.data
    }
  }
  getDataSourceFunc()

  // 验证规则
  const rule = reactive({})

  const elFormRef = ref()
  const { gridRef, gridOptions, gridEvents, selectedRows, refresh } = useGvaGrid({
    id: 'plugin-announcement-info',
    api: getInfoList,
    defaultSort: null,
    checkbox: true,
    searchItems: [
      {
        field: 'createdAt',
        title: '创建日期',
        span: 8,
        itemRender: {
          name: 'gvaDateRange',
          props: { type: 'datetimerange', valueFormat: 'YYYY-MM-DD HH:mm:ss', startField: 'startCreatedAt', endField: 'endCreatedAt' }
        }
      }
    ],
    columns: [
      { field: 'CreatedAt', title: '日期', width: 180, cellRender: { name: 'gvaDate' } },
      { field: 'title', title: '标题', width: 120 },
      { field: 'userID', title: '作者', width: 120, slots: { default: 'userID' } },
      { field: 'attachments', title: '附件', width: 200, slots: { default: 'attachments' } },
      { title: '操作', fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  const { deleteRow, deleteRows: onDelete } = useGvaGridDelete(gridRef, {
    delete: (rows) => {
      if (rows.length === 1) return deleteInfo({ ID: rows[0].ID })
      return deleteInfoByIds({ IDs: rows.map((item) => item.ID) })
    },
    confirmText: '确定要删除吗?',
    successText: '删除成功'
  })

  // 行为控制标记（弹窗内部需要增还是改）
  const type = ref('')

  // 更新行
  const updateInfoFunc = async (row) => {
    const res = await findInfo({ ID: row.ID })
    type.value = 'update'
    if (res.code === 0) {
      formData.value = res.data
      dialogFormVisible.value = true
    }
  }

  // 删除行
  const deleteInfoFunc = async (row) => {
    const res = await deleteInfo({ ID: row.ID })
    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: '删除成功'
      })
      if (tableData.value.length === 1 && page.value > 1) {
        page.value--
      }
      getTableData()
    }
  }

  // 弹窗控制标记
  const dialogFormVisible = ref(false)

  // 打开弹窗
  const openDialog = () => {
    type.value = 'create'
    dialogFormVisible.value = true
  }

  // 关闭弹窗
  const closeDialog = () => {
    dialogFormVisible.value = false
    formData.value = {
      title: '',
      content: '',
      userID: undefined,
      attachments: []
    }
  }
  // 弹窗确定
  const enterDialog = async () => {
    elFormRef.value?.validate(async (valid) => {
      if (!valid) return
      let res
      switch (type.value) {
        case 'create':
          res = await createInfo(formData.value)
          break
        case 'update':
          res = await updateInfo(formData.value)
          break
        default:
          res = await createInfo(formData.value)
          break
      }
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '创建/更改成功'
        })
        closeDialog()
        getTableData()
      }
    })
  }

  const downloadFile = (url) => {
    window.open(getUrl(url), '_blank')
  }
</script>

<style>
  .file-list {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }

  .fileBtn {
    margin-bottom: 10px;
  }

  .fileBtn:last-child {
    margin-bottom: 0;
  }
</style>
