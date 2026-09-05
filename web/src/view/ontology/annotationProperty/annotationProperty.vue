<template>
  <div class="annotation-property">
    <warning-bar title="注：注释属性注册表是建模侧序列化的驱动数据（新增一行即自动支持 ont:xxx）；本表无内置保护，删除后建模侧将不再序列化对应注释" />
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="openForm('add')">新增注释属性</el-button>
        <el-button type="primary" icon="download" @click="exportExcel">导出</el-button>
      </template>

      <template #operate="{ row }">
        <el-button icon="edit" type="primary" link @click="openForm('edit', row)">编辑</el-button>
        <el-button icon="delete" type="primary" link @click="deleteRow(row)">删除</el-button>
      </template>
    </GvaGrid>

    <el-drawer v-model="formVisible" :size="appStore.drawerSize" :show-close="false">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ titleForm }}</span>
          <div>
            <el-button @click="closeForm">取 消</el-button>
            <el-button type="primary" @click="submitForm">确 定</el-button>
          </div>
        </div>
      </template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="属性名" prop="localName">
          <el-input v-model="form.localName" placeholder="唯一标识，如 icon" />
        </el-form-item>
        <el-form-item label="显示名" prop="label">
          <el-input v-model="form.label" placeholder="如 图标" />
        </el-form-item>
        <el-form-item label="值域类型" prop="rangeXsd">
          <el-input v-model="form.rangeXsd" placeholder="如 xsd:string" />
        </el-form-item>
        <el-form-item label="作用对象" prop="appliesTo">
          <el-select v-model="form.appliesTo" placeholder="请选择作用对象">
            <el-option v-for="item in appliesToOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
      </el-form>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    getAnnotationPropertyList,
    findAnnotationProperty,
    createAnnotationProperty,
    updateAnnotationProperty,
    deleteAnnotationProperty,
    exportAnnotationPropertyExcel
  } from '@/api/ontology/annotationProperty'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useAppStore } from '@/pinia'
  import GvaGrid, { useGvaGrid } from '@/components/gvaGrid'

  defineOptions({
    name: 'AnnotationProperty'
  })

  const appStore = useAppStore()

  const appliesToOptions = [
    { label: '类', value: 'class' },
    { label: '数据属性', value: 'datatypeProperty' },
    { label: '对象属性', value: 'objectProperty' },
    { label: '个体', value: 'individual' },
    { label: '全部', value: 'all' }
  ]

  const { gridRef, gridOptions, gridEvents, refresh } = useGvaGrid({
    id: 'ontology-annotationProperty',
    defaultSort: null,
    api: getAnnotationPropertyList,
    searchItems: [
      {
        field: 'keyword',
        title: '关键字',
        span: 6,
        itemRender: { name: 'VxeInput', props: { placeholder: '属性名或显示名', clearable: true } }
      },
      {
        field: 'appliesTo',
        title: '作用对象',
        span: 6,
        itemRender: { name: 'gvaDictSelect', props: { dict: 'ont_applies_to', placeholder: '请选择', clearable: true } }
      },
      {
        field: 'startTime',
        title: '创建时间',
        span: 8,
        itemRender: {
          name: 'gvaDateRange',
          props: { type: 'datetimerange', startField: 'startTime', endField: 'endTime', valueFormat: 'YYYY-MM-DD HH:mm:ss' }
        }
      }
    ],
    columns: [
      { field: 'localName', title: '属性名', width: 140 },
      { field: 'label', title: '显示名', width: 120 },
      { field: 'rangeXsd', title: '值域类型', width: 120 },
      { field: 'appliesTo', title: '作用对象', width: 120, cellRender: { name: 'gvaDict', props: { dict: 'ont_applies_to' } } },
      { field: 'description', title: '描述', minWidth: 200, showOverflow: true },
      { field: 'sort', title: '排序', width: 70 },
      { field: 'createdBy', title: '创建者', width: 100 },
      { field: 'CreatedAt', title: '创建时间', width: 170, cellRender: { name: 'gvaDate' } },
      { field: 'updatedBy', title: '更新者', width: 100 },
      { field: 'UpdatedAt', title: '更新时间', width: 170, cellRender: { name: 'gvaDate' } },
      { title: '操作', width: 160, fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  // 新增/编辑表单
  const formVisible = ref(false)
  const titleForm = ref('新增注释属性')
  const dialogType = ref('add')
  const formRef = ref(null)
  const form = ref({})
  const rules = ref({
    localName: [{ required: true, message: '请输入属性名', trigger: 'blur' }],
    label: [{ required: true, message: '请输入显示名', trigger: 'blur' }],
    appliesTo: [{ required: true, message: '请选择作用对象', trigger: 'change' }]
  })

  const initForm = () => {
    formRef.value && formRef.value.resetFields()
    form.value = {
      ID: 0,
      localName: '',
      label: '',
      rangeXsd: '',
      appliesTo: 'all',
      description: '',
      sort: 0
    }
  }
  const closeForm = () => {
    initForm()
    formVisible.value = false
  }
  const openForm = async (type, row) => {
    initForm()
    dialogType.value = type
    titleForm.value = type === 'add' ? '新增注释属性' : '编辑注释属性'
    if (type === 'edit') {
      const res = await findAnnotationProperty({ ID: row.ID })
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
      const req = { ...form.value }
      const isAdd = dialogType.value === 'add'
      const res = isAdd ? await createAnnotationProperty(req) : await updateAnnotationProperty(req)
      if (res.code === 0) {
        ElMessage({ type: 'success', message: isAdd ? '创建成功' : '更新成功' })
        await refresh()
        closeForm()
      }
    })
  }
  const deleteRow = (row) => {
    ElMessageBox.confirm(
      `此操作将删除注释属性「${row.label}」，删除后建模侧将不再序列化该 ont:${row.localName} 注释, 是否继续?`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
      .then(async () => {
        const res = await deleteAnnotationProperty({ ID: row.ID })
        if (res.code === 0) {
          ElMessage({ type: 'success', message: '删除成功!' })
          refresh()
        }
      })
      .catch(() => {})
  }
  const getSearchParams = () => {
    const data = gridRef.value ? gridRef.value.getFormData() : {}
    const params = {}
    Object.keys(data).forEach((key) => {
      if (data[key] !== undefined && data[key] !== null && data[key] !== '') {
        params[key] = data[key]
      }
    })
    return params
  }
  const exportExcel = async () => {
    // 拦截器对 blob 响应返回原始 response，取 res.data 下载
    const res = await exportAnnotationPropertyExcel(getSearchParams())
    const blob = new Blob([res.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = 'annotation_properties.xlsx'
    link.click()
    URL.revokeObjectURL(url)
    ElMessage({ type: 'success', message: '导出成功' })
  }
</script>

<style lang="scss" scoped>
  .annotation-property {
    .el-select {
      width: 100%;
    }

    .el-input-number {
      margin-left: 0;
    }
  }
</style>
