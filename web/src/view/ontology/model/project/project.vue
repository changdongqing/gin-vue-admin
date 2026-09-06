<template>
  <div class="model-project">
    <warning-bar title="注：本体项目是建模域顶层容器，持有命名空间基址、IRI 前缀注册与序列化策略；项目归档后只读（编辑/删除/前缀维护全部拒绝）" />
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="openForm('add')">新增本体项目</el-button>
      </template>

      <template #operate="{ row }">
        <el-button icon="edit" type="primary" link :disabled="row.status === 'archived'" @click="openForm('edit', row)">编辑</el-button>
        <el-button icon="link" type="primary" link @click="openPrefix(row)">前缀管理</el-button>
        <el-button icon="delete" type="primary" link :disabled="row.status === 'archived'" @click="deleteRow(row)">删除</el-button>
      </template>
    </GvaGrid>

    <!-- 项目表单抽屉 -->
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
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="项目编码" prop="projectCode">
          <el-input v-model="form.projectCode" placeholder="唯一标识，如 fire-equipment" />
        </el-form-item>
        <el-form-item label="项目名称" prop="name">
          <el-input v-model="form.name" placeholder="如 消防设备本体" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="命名空间基址" prop="namespaceBase">
          <el-input v-model="form.namespaceBase" placeholder="如 http://example.com/onto/fire/" />
        </el-form-item>
        <el-form-item label="默认格式" prop="defaultFormat">
          <el-radio-group v-model="form.defaultFormat">
            <el-radio value="TTL">Turtle</el-radio>
            <el-radio value="OWL_XML">OWL XML</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="序列化策略" prop="serializationStrategy">
          <el-radio-group v-model="form.serializationStrategy">
            <el-radio value="B">方案B（独立副本）</el-radio>
            <el-tooltip content="预留，暂未开放" placement="top">
              <span><el-radio value="A" disabled>方案A（共享属性）</el-radio></span>
            </el-tooltip>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="form.status" placeholder="请选择状态">
            <el-option v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
      </el-form>
    </el-drawer>

    <!-- 前缀管理抽屉 -->
    <el-drawer v-model="prefixVisible" size="600px" :title="`前缀管理 - ${prefixProject.name || ''}`">
      <div class="mb-3">
        <el-button type="primary" icon="plus" @click="openPrefixForm('add')">新增前缀</el-button>
      </div>
      <el-table :data="prefixList" border>
        <el-table-column prop="prefix" label="前缀名" width="120" />
        <el-table-column prop="namespace" label="命名空间" min-width="260" show-overflow-tooltip />
        <el-table-column prop="isDefault" label="默认" width="90" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.isDefault === 1" type="success">默认</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" align="center">
          <template #default="{ row }">
            <el-button icon="edit" type="primary" link @click="openPrefixForm('edit', row)">编辑</el-button>
            <el-popconfirm title="确定删除此前缀？" @confirm="deletePrefix(row)">
              <template #reference>
                <el-button icon="delete" type="primary" link>删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <!-- 前缀表单（内嵌弹窗，避免抽屉叠抽屉） -->
      <el-dialog v-model="prefixFormVisible" :title="prefixFormType === 'add' ? '新增前缀' : '编辑前缀'" width="460px" append-to-body>
        <el-form ref="prefixFormRef" :model="prefixForm" :rules="prefixRules" label-width="90px">
          <el-form-item label="前缀名" prop="prefix">
            <el-input v-model="prefixForm.prefix" placeholder="如 ex / qudt / fire" />
          </el-form-item>
          <el-form-item label="命名空间" prop="namespace">
            <el-input v-model="prefixForm.namespace" placeholder="如 http://example.com/onto/fire/" />
          </el-form-item>
          <el-form-item label="默认前缀" prop="isDefault">
            <el-switch v-model="prefixForm.isDefault" :active-value="1" :inactive-value="0" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="prefixFormVisible = false">取 消</el-button>
          <el-button type="primary" @click="submitPrefixForm">确 定</el-button>
        </template>
      </el-dialog>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    getModelProjectList,
    findModelProject,
    createModelProject,
    updateModelProject,
    deleteModelProject,
    getModelPrefixList,
    createModelPrefix,
    updateModelPrefix,
    deleteModelPrefix
  } from '@/api/ontology/modelProject'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref, reactive } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useAppStore } from '@/pinia'
  import GvaGrid, { useGvaGrid } from '@/components/gvaGrid'

  defineOptions({
    name: 'OntModelProject'
  })

  const appStore = useAppStore()

  const ncNameRe = /^[A-Za-z_][A-Za-z0-9_.-]*$/

  const { gridRef, gridOptions, gridEvents, refresh } = useGvaGrid({
    id: 'ontology-modelProject',
    defaultSort: null,
    api: getModelProjectList,
    searchItems: [
      {
        field: 'keyword',
        title: '关键字',
        span: 6,
        itemRender: { name: 'VxeInput', props: { placeholder: '项目编码或名称', clearable: true } }
      },
      {
        field: 'status',
        title: '状态',
        span: 6,
        itemRender: { name: 'gvaDictSelect', props: { dict: 'ont_model_project_status', placeholder: '请选择', clearable: true } }
      }
    ],
    columns: [
      { field: 'projectCode', title: '项目编码', width: 140 },
      { field: 'name', title: '项目名称', width: 140 },
      { field: 'namespaceBase', title: '命名空间基址', width: 240, showOverflow: true },
      { field: 'defaultFormat', title: '默认格式', width: 90, cellRender: { name: 'gvaDict', props: { dict: 'ont_model_format' } } },
      { field: 'serializationStrategy', title: '策略', width: 90, cellRender: { name: 'gvaDict', props: { dict: 'ont_model_strategy' } } },
      { field: 'status', title: '状态', width: 90, cellRender: { name: 'gvaDict', props: { dict: 'ont_model_project_status' } } },
      { field: 'CreatedAt', title: '创建时间', width: 170, cellRender: { name: 'gvaDate' } },
      { title: '操作', width: 240, fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  // ── 项目表单 ────────────────────────────────
  const statusOptions = [
    { label: '草稿', value: 'draft' },
    { label: '活跃', value: 'active' },
    { label: '归档', value: 'archived' }
  ]

  const formVisible = ref(false)
  const titleForm = ref('新增本体项目')
  const dialogType = ref('add')
  const formRef = ref(null)
  const form = ref({})
  const rules = ref({
    projectCode: [{ required: true, message: '请输入项目编码', trigger: 'blur' }],
    name: [{ required: true, message: '请输入项目名称', trigger: 'blur' }],
    namespaceBase: [{ required: true, message: '请输入命名空间基址', trigger: 'blur' }],
    status: [{ required: true, message: '请选择状态', trigger: 'change' }]
  })

  const initForm = () => {
    formRef.value && formRef.value.resetFields()
    form.value = {
      ID: 0,
      projectCode: '',
      name: '',
      description: '',
      namespaceBase: '',
      defaultFormat: 'TTL',
      serializationStrategy: 'B',
      status: 'draft'
    }
  }
  const closeForm = () => {
    initForm()
    formVisible.value = false
  }
  const openForm = async (type, row) => {
    initForm()
    dialogType.value = type
    titleForm.value = type === 'add' ? '新增本体项目' : '编辑本体项目'
    if (type === 'edit') {
      const res = await findModelProject({ ID: row.ID })
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
      const res = isAdd ? await createModelProject(req) : await updateModelProject(req)
      if (res.code === 0) {
        ElMessage({ type: 'success', message: isAdd ? '创建成功' : '更新成功' })
        await refresh()
        closeForm()
      }
    })
  }
  const deleteRow = (row) => {
    ElMessageBox.confirm(
      `此操作将删除本体项目「${row.name}」，前缀随之废弃，是否继续?`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
      .then(async () => {
        const res = await deleteModelProject({ ID: row.ID })
        if (res.code === 0) {
          ElMessage({ type: 'success', message: '删除成功!' })
          refresh()
        }
      })
      .catch(() => {})
  }

  // ── 前缀管理 ────────────────────────────────
  const prefixVisible = ref(false)
  const prefixProject = reactive({ ID: 0, name: '' })
  const prefixList = ref([])
  const prefixFormVisible = ref(false)
  const prefixFormType = ref('add')
  const prefixFormRef = ref(null)
  const prefixForm = ref({})
  const prefixRules = ref({
    prefix: [
      { required: true, message: '请输入前缀名', trigger: 'blur' },
      { pattern: ncNameRe, message: '须以字母或下划线开头，仅含字母/数字/下划线/点/连字符', trigger: 'blur' }
    ],
    namespace: [{ required: true, message: '请输入命名空间', trigger: 'blur' }]
  })

  const loadPrefixList = async () => {
    const res = await getModelPrefixList({ projectId: prefixProject.ID })
    if (res.code === 0) {
      prefixList.value = res.data || []
    }
  }
  const openPrefix = (row) => {
    prefixProject.ID = row.ID
    prefixProject.name = row.name
    prefixVisible.value = true
    loadPrefixList()
  }
  const initPrefixForm = () => {
    prefixFormRef.value && prefixFormRef.value.resetFields()
    prefixForm.value = { ID: 0, projectId: prefixProject.ID, prefix: '', namespace: '', isDefault: 0 }
  }
  const openPrefixForm = async (type, row) => {
    initPrefixForm()
    prefixFormType.value = type
    if (type === 'edit') {
      prefixForm.value = { ...prefixForm.value, ...row }
      prefixFormRef.value && prefixFormRef.value.clearValidate()
    }
    prefixFormVisible.value = true
  }
  const submitPrefixForm = () => {
    prefixFormRef.value.validate(async (valid) => {
      if (!valid) return
      const req = { ...prefixForm.value, projectId: prefixProject.ID }
      const isAdd = prefixFormType.value === 'add'
      const res = isAdd ? await createModelPrefix(req) : await updateModelPrefix(req)
      if (res.code === 0) {
        ElMessage({ type: 'success', message: isAdd ? '创建成功' : '更新成功' })
        prefixFormVisible.value = false
        loadPrefixList()
      }
    })
  }
  const deletePrefix = async (row) => {
    const res = await deleteModelPrefix({ ID: row.ID, projectId: prefixProject.ID })
    if (res.code === 0) {
      ElMessage({ type: 'success', message: '删除成功!' })
      loadPrefixList()
    }
  }
</script>

<style lang="scss" scoped>
  .model-project {
    .el-select {
      width: 100%;
    }

    .el-input-number {
      margin-left: 0;
    }
  }
</style>
