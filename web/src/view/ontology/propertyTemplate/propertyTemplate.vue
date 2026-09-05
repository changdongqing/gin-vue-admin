<template>
  <div class="property-template">
    <warning-bar title="注：属性模板是本体治理的基础资产，供建模侧经供给接口拉取挂载；内置模板不可编辑/删除，仅可弃用" />
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="openForm('add')">新增属性模板</el-button>
      </template>

      <template #deprecated="{ row }">
        <el-tag :type="row.deprecated === 1 ? 'info' : 'success'" size="small">
          {{ row.deprecated === 1 ? '已弃用' : '正常' }}
        </el-tag>
      </template>

      <template #operate="{ row }">
        <el-button v-if="row.source !== 'builtin'" icon="edit" type="primary" link @click="openForm('edit', row)">编辑</el-button>
        <el-button type="primary" link @click="toggleDeprecated(row)">{{ row.deprecated === 1 ? '取消弃用' : '弃用' }}</el-button>
        <el-button v-if="row.source !== 'builtin'" icon="delete" type="primary" link @click="deleteRow(row)">删除</el-button>
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
        <el-form-item label="模板编码" prop="templateCode">
          <el-input
            v-model="form.templateCode"
            :disabled="dialogType === 'edit' && form.source === 'builtin'"
            placeholder="唯一标识，如 weight"
            @blur="checkCode"
          />
          <div v-if="codeDuplicate" class="text-red-500 text-xs">编码已存在</div>
        </el-form-item>
        <el-form-item label="属性类型" prop="kind">
          <el-select v-model="form.kind" placeholder="请选择属性类型" @change="onKindChange">
            <el-option label="数据属性" value="datatype" />
            <el-option label="对象属性" value="object" />
          </el-select>
        </el-form-item>
        <el-form-item label="显示名" prop="label">
          <el-input v-model="form.label" placeholder="如 重量" />
        </el-form-item>
        <el-form-item label="同义别名" prop="alias">
          <el-input v-model="form.alias" type="textarea" :rows="2" placeholder="逗号分隔，如 温度值,实时温度（供搜索命中与治理合并留痕）" />
        </el-form-item>
        <el-form-item label="业务说明" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="分组" prop="category">
          <el-select v-model="form.category" clearable placeholder="请选择分组">
            <el-option v-for="item in categoryOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <template v-if="form.kind === 'datatype'">
          <el-form-item label="数据类型" prop="type">
            <el-select v-model="form.type" clearable placeholder="请选择数据类型">
              <el-option v-for="item in typeOptions" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="是否标识" prop="isIdentifier">
            <el-switch v-model="form.isIdentifier" :active-value="1" :inactive-value="0" />
          </el-form-item>
          <el-form-item label="预设单位" prop="unitRef">
            <el-input v-model="form.unitRef" placeholder="http://qudt.org/vocab/unit/KiloGM（可选，单位注册表就绪后联动）" />
          </el-form-item>
          <el-form-item label="枚举值" prop="values">
            <el-input v-model="form.values" placeholder="逗号分隔，如 HIGH,MEDIUM,LOW" />
          </el-form-item>
        </template>
        <el-form-item v-if="form.kind === 'object'" label="默认基数" prop="defaultCardinality">
          <el-select v-model="form.defaultCardinality" clearable placeholder="请选择默认基数">
            <el-option v-for="item in cardinalityOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
      </el-form>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    getPropertyTemplateList,
    findPropertyTemplate,
    createPropertyTemplate,
    updatePropertyTemplate,
    deletePropertyTemplate,
    disablePropertyTemplate,
    checkPropertyTemplateCode
  } from '@/api/ontology/propertyTemplate'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useAppStore } from '@/pinia'
  import GvaGrid, { useGvaGrid } from '@/components/gvaGrid'

  defineOptions({
    name: 'PropertyTemplate'
  })

  const appStore = useAppStore()

  const categoryOptions = [
    { label: '基础', value: 'basic' },
    { label: '联系', value: 'contact' },
    { label: '金额', value: 'monetary' },
    { label: '时间', value: 'temporal' },
    { label: '状态', value: 'status' },
    { label: '包含', value: 'containment' },
    { label: '归属', value: 'attribution' },
    { label: '额定', value: 'rated' },
    { label: '监测', value: 'monitoring' },
    { label: '运行', value: 'running' },
    { label: '管理', value: 'management' },
    { label: '运维', value: 'operation' }
  ]
  const typeOptions = [
    { label: '字符串', value: 'string' },
    { label: '整数', value: 'integer' },
    { label: '小数', value: 'decimal' },
    { label: '布尔', value: 'boolean' },
    { label: '日期时间', value: 'datetime' }
  ]
  const cardinalityOptions = [
    { label: '一对多', value: 'one-to-many' },
    { label: '多对一', value: 'many-to-one' },
    { label: '一对一', value: 'one-to-one' },
    { label: '多对多', value: 'many-to-many' }
  ]

  const { gridRef, gridOptions, gridEvents, refresh, reload } = useGvaGrid({
    id: 'ontology-propertyTemplate',
    defaultSort: { field: 'ID', order: 'desc' },
    api: getPropertyTemplateList,
    searchItems: [
      {
        field: 'keyword',
        title: '关键字',
        span: 6,
        itemRender: { name: 'VxeInput', props: { placeholder: '模板编码、显示名或同义别名', clearable: true } }
      },
      {
        field: 'kind',
        title: '属性类型',
        span: 6,
        itemRender: { name: 'gvaDictSelect', props: { dict: 'ont_property_kind', placeholder: '请选择', clearable: true } }
      },
      {
        field: 'category',
        title: '分组',
        span: 6,
        itemRender: { name: 'VxeSelect', props: { placeholder: '请选择', clearable: true, options: categoryOptions } }
      },
      {
        field: 'source',
        title: '来源',
        span: 6,
        itemRender: { name: 'gvaDictSelect', props: { dict: 'ont_source', placeholder: '请选择', clearable: true } }
      }
    ],
    columns: [
      { field: 'templateCode', title: '模板编码', width: 140 },
      { field: 'kind', title: '属性类型', width: 100, cellRender: { name: 'gvaDict', props: { dict: 'ont_property_kind' } } },
      { field: 'label', title: '显示名', width: 120 },
      { field: 'alias', title: '同义别名', width: 140, showOverflow: true },
      { field: 'category', title: '分组', width: 90 },
      { field: 'type', title: '数据类型', width: 90 },
      { field: 'unitRef', title: '预设单位', width: 160, showOverflow: true },
      { field: 'defaultCardinality', title: '基数', width: 100 },
      { field: 'source', title: '来源', width: 80, cellRender: { name: 'gvaDict', props: { dict: 'ont_source' } } },
      { field: 'deprecated', title: '弃用', width: 80, slots: { default: 'deprecated' } },
      { field: 'CreatedAt', title: '创建时间', width: 170, cellRender: { name: 'gvaDate' } },
      { title: '操作', width: 200, fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  // 新增/编辑表单
  const formVisible = ref(false)
  const titleForm = ref('新增属性模板')
  const dialogType = ref('add')
  const formRef = ref(null)
  const codeDuplicate = ref(false)
  const form = ref({})
  const rules = ref({
    templateCode: [{ required: true, message: '请输入模板编码', trigger: 'blur' }],
    kind: [{ required: true, message: '请选择属性类型', trigger: 'change' }],
    label: [{ required: true, message: '请输入显示名', trigger: 'blur' }]
  })

  const initForm = () => {
    formRef.value && formRef.value.resetFields()
    form.value = {
      ID: 0,
      templateCode: '',
      kind: 'datatype',
      label: '',
      alias: '',
      description: '',
      category: '',
      type: '',
      isIdentifier: 0,
      unitRef: '',
      values: '',
      defaultCardinality: '',
      source: 'custom'
    }
    codeDuplicate.value = false
  }
  const closeForm = () => {
    initForm()
    formVisible.value = false
  }
  const openForm = async (type, row) => {
    initForm()
    dialogType.value = type
    titleForm.value = type === 'add' ? '新增属性模板' : '编辑属性模板'
    if (type === 'edit') {
      // 行数据可能不含全字段，编辑前拉详情回填
      const res = await findPropertyTemplate({ ID: row.ID })
      if (res.code === 0) {
        form.value = { ...form.value, ...res.data }
      }
      formRef.value && formRef.value.clearValidate()
    }
    formVisible.value = true
  }
  const onKindChange = () => {
    // kind 切换时清空对方专属字段，避免错位脏数据
    if (form.value.kind === 'datatype') {
      form.value.defaultCardinality = ''
    } else {
      form.value.type = ''
      form.value.isIdentifier = 0
      form.value.unitRef = ''
      form.value.values = ''
    }
  }
  const checkCode = async () => {
    const code = (form.value.templateCode || '').trim()
    if (!code) {
      codeDuplicate.value = false
      return
    }
    const res = await checkPropertyTemplateCode({
      templateCode: code,
      excludeId: dialogType.value === 'edit' ? form.value.ID : undefined
    })
    codeDuplicate.value = res.code === 0 && res.data && res.data.usable === false
  }
  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      const req = { ...form.value }
      const isAdd = dialogType.value === 'add'
      const res = isAdd ? await createPropertyTemplate(req) : await updatePropertyTemplate(req)
      if (res.code === 0) {
        ElMessage({ type: 'success', message: isAdd ? '创建成功' : '更新成功' })
        await refresh()
        closeForm()
      }
    })
  }
  const toggleDeprecated = (row) => {
    disablePropertyTemplate({ ID: row.ID }).then((res) => {
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: row.deprecated === 1 ? `已取消弃用「${row.label}」` : `已弃用「${row.label}」`
        })
        refresh()
      }
    })
  }
  const deleteRow = (row) => {
    ElMessageBox.confirm(`此操作将删除属性模板「${row.label}」, 是否继续?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        const res = await deletePropertyTemplate({ ID: row.ID })
        if (res.code === 0) {
          ElMessage({ type: 'success', message: '删除成功!' })
          reload()
        }
      })
      .catch(() => {})
  }
</script>

<style lang="scss" scoped>
  .property-template {
    .el-select {
      width: 100%;
    }
  }
</style>
