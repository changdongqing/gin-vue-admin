<template>
  <div class="model-class">
    <warning-bar title="注：本体类是建模域核心，可从分类模板实例化（复制字段值+溯源）或空白创建；「可实例化」类才能被外部模块关联绑定" />
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="openForm('add')">新增本体类</el-button>
      </template>

      <template #templateTag="{ row }">
        <el-tag v-if="row.templateCode" type="primary">模板</el-tag>
        <el-tag v-else type="info">空白</el-tag>
      </template>

      <template #operate="{ row }">
        <el-button icon="edit" type="primary" link @click="openForm('edit', row)">编辑</el-button>
        <el-button icon="magic-stick" type="primary" link @click="openInstantiate(row)">实例化</el-button>
        <el-button icon="view" type="primary" link @click="openDetail(row)">详情</el-button>
        <el-button icon="delete" type="primary" link @click="deleteRow(row)">删除</el-button>
      </template>
    </GvaGrid>

    <!-- 类编辑抽屉 -->
    <el-drawer v-model="formVisible" size="780px" :show-close="false">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ titleForm }}</span>
          <div>
            <el-button @click="closeForm">取 消</el-button>
            <el-button type="primary" @click="submitForm">确 定</el-button>
          </div>
        </div>
      </template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="本地名" prop="localName">
          <el-input v-model="form.localName" :disabled="dialogType === 'edit'" placeholder="如 FirePump" />
          <div class="w-full text-xs text-gray-400 leading-5">
            将生成 IRI = 项目命名空间 + 本地名；创建后不可修改
          </div>
        </el-form-item>
        <el-form-item label="英文名" prop="label">
          <el-input v-model="form.label" placeholder="英文标签" />
        </el-form-item>
        <el-form-item label="中文名" prop="labelCn">
          <el-input v-model="form.labelCn" placeholder="中文标签" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="图标" prop="icon">
          <el-input v-model="form.icon" placeholder="如 ep:set-up（可被模板实例化继承）" />
        </el-form-item>
        <el-form-item label="颜色" prop="color">
          <el-input v-model="form.color" placeholder="如 #1890ff" />
        </el-form-item>
        <el-form-item label="排序" prop="sortOrder">
          <el-input-number v-model="form.sortOrder" :min="0" />
        </el-form-item>
        <el-form-item label="可实例化" prop="isInstantiable">
          <el-switch v-model="form.isInstantiable" :active-value="1" :inactive-value="0" />
          <div class="w-full text-xs text-gray-400 leading-5">
            开启表示该类会实例化为主数据个体（如设备），外部模块关联仅绑定此类
          </div>
        </el-form-item>
      </el-form>
    </el-drawer>

    <instantiate-modal ref="instantiateRef" @success="refresh" />
    <class-detail
      ref="detailRef"
      @select-templates="(kind) => openTemplateSelect(kind)"
      @add-property="(kind) => openPropertyForm(kind)"
      @edit-property="(kind, row) => openPropertyForm(kind, row)"
      @changed="reloadList"
    />
    <property-template-select ref="templateSelectRef" @success="reloadDetail" />
    <property-form ref="propertyFormRef" :class-list="classList" @success="reloadDetail" />
  </div>
</template>

<script setup>
  import {
    getModelClassList,
    findModelClass,
    createModelClass,
    updateModelClass,
    deleteModelClass,
    getModelProjectAll,
    getModelClassByProject
  } from '@/api/ontology/modelClass'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useAppStore } from '@/pinia'
  import GvaGrid, { useGvaGrid } from '@/components/gvaGrid'
  import ClassDetail from './components/classDetail.vue'
  import InstantiateModal from './components/instantiateModal.vue'
  import PropertyTemplateSelect from './components/propertyTemplateSelect.vue'
  import PropertyForm from './components/propertyForm.vue'

  defineOptions({
    name: 'OntModelClass'
  })

  const appStore = useAppStore()

  // 项目下拉（搜索区 + 新增前置校验）
  const projectOptions = ref([])
  const loadProjects = async () => {
    const res = await getModelProjectAll()
    if (res.code === 0) {
      projectOptions.value = (res.data || []).map((item) => ({ label: item.name, value: item.ID }))
    }
  }
  onMounted(loadProjects)

  const { gridRef, gridOptions, gridEvents, refresh, reload } = useGvaGrid({
    id: 'ontology-modelClass',
    defaultSort: null,
    api: getModelClassList,
    searchItems: [
      {
        field: 'projectId',
        title: '本体项目',
        span: 6,
        itemRender: { name: 'VxeSelect', props: { placeholder: '不选则查全部', clearable: true, options: projectOptions } }
      },
      {
        field: 'keyword',
        title: '关键字',
        span: 6,
        itemRender: { name: 'VxeInput', props: { placeholder: 'label / 中文名 / 本地名', clearable: true } }
      },
      {
        field: 'hasTemplate',
        title: '模板溯源',
        span: 6,
        itemRender: {
          name: 'VxeSelect',
          props: {
            placeholder: '请选择',
            clearable: true,
            options: [
              { label: '仅模板创建', value: 'true' },
              { label: '仅空白类', value: 'false' }
            ]
          }
        }
      }
    ],
    columns: [
      { field: 'classIri', title: '类 IRI', width: 280, showOverflow: true },
      { field: 'localName', title: '本地名', width: 140 },
      { field: 'label', title: '英文名', width: 110 },
      { field: 'labelCn', title: '中文名', width: 110 },
      { field: 'templateCode', title: '模板溯源', width: 110, slots: { default: 'templateTag' } },
      { field: 'classificationCode', title: '分类编码', width: 120 },
      { field: 'isInstantiable', title: '实例化', width: 80, formatter: ({ cellValue }) => (cellValue === 1 ? '是' : '否') },
      { field: 'sortOrder', title: '排序', width: 70 },
      { field: 'CreatedAt', title: '创建时间', width: 170, cellRender: { name: 'gvaDate' } },
      { title: '操作', width: 280, fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  // ── 类编辑抽屉 ──────────────────────────────
  const formVisible = ref(false)
  const titleForm = ref('新增本体类')
  const dialogType = ref('add')
  const formRef = ref(null)
  const form = ref({})
  const rules = ref({
    localName: [
      { required: true, message: '请输入本地名', trigger: 'blur' },
      { pattern: /^[A-Za-z][A-Za-z0-9_-]*$/, message: '须以字母开头，仅含字母/数字/下划线/连字符', trigger: 'blur' }
    ]
  })

  const initForm = (projectId) => {
    formRef.value && formRef.value.resetFields()
    form.value = {
      ID: 0,
      projectId,
      localName: '',
      label: '',
      labelCn: '',
      description: '',
      icon: '',
      color: '',
      sortOrder: 0,
      isInstantiable: 0
    }
  }
  const closeForm = () => {
    formVisible.value = false
  }
  const openForm = async (type, row) => {
    const searchForm = gridRef.value ? gridRef.value.getFormData() : {}
    if (type === 'add') {
      if (!searchForm.projectId) {
        ElMessage.warning('请先选择本体项目')
        return
      }
      initForm(searchForm.projectId)
      titleForm.value = '新增本体类'
      dialogType.value = 'add'
      formVisible.value = true
      return
    }
    initForm(row.projectId)
    titleForm.value = '编辑本体类'
    dialogType.value = 'edit'
    const res = await findModelClass({ ID: row.ID })
    if (res.code === 0) {
      form.value = { ...form.value, ...res.data }
    }
    formRef.value && formRef.value.clearValidate()
    formVisible.value = true
  }
  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      const req = { ...form.value }
      const isAdd = dialogType.value === 'add'
      const res = isAdd ? await createModelClass(req) : await updateModelClass(req)
      if (res.code === 0) {
        ElMessage({ type: 'success', message: isAdd ? '创建成功' : '更新成功' })
        await refresh()
        closeForm()
      }
    })
  }
  const deleteRow = (row) => {
    ElMessageBox.confirm(`此操作将删除本体类「${row.localName}」，是否继续?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        const res = await deleteModelClass({ ID: row.ID })
        if (res.code === 0) {
          ElMessage({ type: 'success', message: '删除成功!' })
          refresh()
        }
      })
      .catch(() => {})
  }

  // ── 实例化 / 详情 / 属性维护 ────────────────
  const instantiateRef = ref(null)
  const detailRef = ref(null)
  const templateSelectRef = ref(null)
  const propertyFormRef = ref(null)
  const classList = ref([]) // 当前操作类所在项目的类清单（对象属性 Range 下拉）
  let currentClass = null

  const openInstantiate = (row) => {
    instantiateRef.value.open(row.projectId, row.ID)
  }
  const openDetail = (row) => {
    currentClass = row
    loadClassList(row.projectId)
    detailRef.value.open(row)
  }
  const loadClassList = async (projectId) => {
    const res = await getModelClassByProject({ projectId })
    if (res.code === 0) {
      classList.value = res.data || []
    }
  }
  const openTemplateSelect = (kind) => {
    templateSelectRef.value.open(currentClass.projectId, currentClass.ID, kind)
  }
  const openPropertyForm = (kind, row = null) => {
    propertyFormRef.value.open(currentClass.projectId, currentClass.ID, kind, row ? row.ID : 0)
  }
  const reloadDetail = () => {
    detailRef.value && detailRef.value.reload()
    refresh()
  }
  const reloadList = () => {
    refresh()
  }
</script>

<style lang="scss" scoped>
  .model-class {
    .el-select {
      width: 100%;
    }

    .el-input-number {
      margin-left: 0;
    }
  }
</style>
