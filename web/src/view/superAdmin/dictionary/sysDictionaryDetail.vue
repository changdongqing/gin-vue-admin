<template>
  <div class="h-full">
    <!-- fit 测量模式： splitter 面板无确定高度链，按 .gva-container2 实测定高；下方无 BottomInfo，预留 0 -->
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents" height="fit" :bottom-reserve="0">
      <template #toolbar-buttons>
        <span class="text font-bold">字典详细内容</span>
      </template>

      <template #toolbar-tools>
        <el-input
          placeholder="搜索展示值"
          v-model="searchName"
          clearable
          class="!w-52"
          @clear="clearSearchInput"
          :prefix-icon="Search"
          @keydown="handleInputKeyDown"
        >
          <template #append>
            <el-button :type="searchName ? 'primary' : 'info'" @click="applySearch">搜索</el-button>
          </template>
        </el-input>
        <el-button type="primary" icon="plus" @click="openDrawer">新增字典项</el-button>
      </template>

      <template #status="{ row }">
        {{ formatBoolean(row.status) }}
      </template>

      <template #operate="{ row }">
        <el-button type="primary" link icon="plus" @click="addChildNode(row)">添加子项</el-button>
        <el-button type="primary" link icon="edit" @click="updateSysDictionaryDetailFunc(row)">变更</el-button>
        <el-button type="primary" link icon="delete" @click="deleteRow(row)">删除</el-button>
      </template>
    </GvaGrid>

    <el-drawer
      v-model="drawerFormVisible"
      :size="appStore.drawerSize"
      :show-close="false"
      :before-close="closeDrawer"
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ type === 'create' ? '添加字典项' : '修改字典项' }}</span>
          <div>
            <el-button @click="closeDrawer"> 取 消 </el-button>
            <el-button type="primary" @click="enterDrawer"> 确 定 </el-button>
          </div>
        </div>
      </template>
      <el-form ref="drawerForm" :model="formData" :rules="rules" label-width="110px">
        <el-form-item label="父级字典项" prop="parentID">
          <el-cascader
            v-model="formData.parentID"
            :options="[rootOption, ...treeData]"
            :props="cascadeProps"
            placeholder="请选择父级字典项（可选）"
            clearable
            filterable
            :style="{ width: '100%' }"
            @change="handleParentChange"
          />
        </el-form-item>
        <el-form-item label="展示值" prop="label">
          <el-input v-model="formData.label" placeholder="请输入展示值" clearable :style="{ width: '100%' }" />
        </el-form-item>
        <el-form-item label="字典值" prop="value">
          <el-input v-model="formData.value" placeholder="请输入字典值" clearable :style="{ width: '100%' }" />
        </el-form-item>
        <el-form-item label="扩展值" prop="extend">
          <el-input v-model="formData.extend" placeholder="请输入扩展值" clearable :style="{ width: '100%' }" />
        </el-form-item>
        <el-form-item label="启用状态" prop="status" required>
          <el-switch v-model="formData.status" active-text="开启" inactive-text="停用" />
        </el-form-item>
        <el-form-item label="排序标记" prop="sort">
          <el-input-number v-model.number="formData.sort" placeholder="排序标记" />
        </el-form-item>
      </el-form>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    createSysDictionaryDetail,
    deleteSysDictionaryDetail,
    updateSysDictionaryDetail,
    findSysDictionaryDetail,
    getDictionaryTreeList
  } from '@/api/sysDictionaryDetail' // 此处请自行替换地址
  import { ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import { formatBoolean } from '@/utils/format'
  import { useAppStore } from '@/pinia'
  import { Search } from '@element-plus/icons-vue'
  import GvaGrid, { useGvaGrid, useGvaGridDelete } from '@/components/gvaGrid'

  defineOptions({
    name: 'SysDictionaryDetail'
  })

  const appStore = useAppStore()
  const searchName = ref('')

  const props = defineProps({
    sysDictionaryID: {
      type: Number,
      default: 0
    }
  })

  const formData = ref({
    label: null,
    value: null,
    status: true,
    sort: null,
    parentID: null
  })

  const rules = ref({
    label: [
      {
        required: true,
        message: '请输入展示值',
        trigger: 'blur'
      }
    ],
    value: [
      {
        required: true,
        message: '请输入字典值',
        trigger: 'blur'
      }
    ],
    sort: [
      {
        required: true,
        message: '排序标记',
        trigger: 'blur'
      }
    ]
  })

  const treeData = ref([])

  // 级联选择器配置
  const cascadeProps = {
    value: 'ID',
    label: 'label',
    children: 'children',
    checkStrictly: true, // 允许选择任意级别
    emitPath: false // 只返回选中节点的值
  }

  const normalizeSearch = (value) => (value ?? '').toString().toLowerCase()

  const filterTree = (nodes, keyword) => {
    const trimmed = normalizeSearch(keyword).trim()
    if (!trimmed) {
      return nodes
    }
    const walk = (list) => {
      const result = []
      for (const node of list) {
        const label = normalizeSearch(node.label)
        const children = Array.isArray(node.children) ? walk(node.children) : []
        if (label.includes(trimmed) || children.length > 0) {
          result.push({
            ...node,
            children
          })
        }
      }
      return result
    }
    return walk(nodes)
  }

  // 字典树：无分页；搜索为客户端过滤（由 api 包装层应用）
  const { gridRef, gridOptions, gridEvents, refresh } = useGvaGrid({
    id: 'superAdmin-sysDictionaryDetail',
    pager: false,
    checkbox: true,
    defaultSort: null,
    api: async () => {
      if (!props.sysDictionaryID) return { code: 0, data: { list: [], total: 0 } }
      const res = await getDictionaryTreeList({ sysDictionaryID: props.sysDictionaryID })
      treeData.value = (res.data && res.data.list) || []
      const list = filterTree(treeData.value, searchName.value)
      return { ...res, data: { list, total: list.length } }
    },
    gridConfig: {
      treeConfig: { rowField: 'ID', children: 'children', expandAll: true }
    },
    columns: [
      { field: 'label', title: '展示值', minWidth: 100, treeNode: true },
      { field: 'value', title: '字典值', minWidth: 100 },
      { field: 'extend', title: '扩展值', minWidth: 100 },
      { field: 'level', title: '层级', width: 80 },
      { field: 'status', title: '启用状态', width: 100, slots: { default: 'status' } },
      { field: 'sort', title: '排序标记', width: 100 },
      { title: '操作', fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  const { deleteRow } = useGvaGridDelete(gridRef, {
    delete: (rows) => deleteSysDictionaryDetail({ ID: rows[0].ID }),
    confirmText: '确定要删除吗?',
    successText: '删除成功'
  })

  const rootOption = {
    ID: null,
    label: '无父级（根级）'
  }

  const applySearch = () => {
    refresh()
  }

  const type = ref('')
  const drawerFormVisible = ref(false)

  const updateSysDictionaryDetailFunc = async (row) => {
    drawerForm.value && drawerForm.value.clearValidate()
    const res = await findSysDictionaryDetail({ ID: row.ID })
    type.value = 'update'
    if (res.code === 0) {
      formData.value = res.data.reSysDictionaryDetail
      drawerFormVisible.value = true
    }
  }

  // 添加子节点
  const addChildNode = (parentNode) => {
    type.value = 'create'
    formData.value = {
      label: null,
      value: null,
      status: true,
      sort: null,
      parentID: parentNode.ID,
      sysDictionaryID: props.sysDictionaryID
    }
    drawerForm.value && drawerForm.value.clearValidate()
    drawerFormVisible.value = true
  }

  // 处理父级选择变化
  const handleParentChange = (value) => {
    formData.value.parentID = value
  }

  const closeDrawer = () => {
    drawerFormVisible.value = false
    formData.value = {
      label: null,
      value: null,
      status: true,
      sort: null,
      parentID: null,
      sysDictionaryID: props.sysDictionaryID
    }
  }

  const drawerForm = ref(null)
  const enterDrawer = async () => {
    drawerForm.value.validate(async (valid) => {
      formData.value.sysDictionaryID = props.sysDictionaryID
      if (!valid) return
      let res
      switch (type.value) {
        case 'create':
          res = await createSysDictionaryDetail(formData.value)
          break
        case 'update':
          res = await updateSysDictionaryDetail(formData.value)
          break
        default:
          res = await createSysDictionaryDetail(formData.value)
          break
      }
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '创建/更改成功'
        })
        closeDrawer()
        refresh()
      }
    })
  }

  const openDrawer = () => {
    type.value = 'create'
    formData.value.parentID = null
    drawerForm.value && drawerForm.value.clearValidate()
    drawerFormVisible.value = true
  }

  const clearSearchInput = () => {
    searchName.value = ''
    applySearch()
  }

  const handleInputKeyDown = (e) => {
    if (e.key === 'Enter') {
      applySearch()
    }
  }

  watch(
    () => props.sysDictionaryID,
    () => {
      refresh()
    }
  )
</script>

<style scoped></style>
