<template>
  <div class="class-template">
    <warning-bar title="注：分类模板为多棵分类树（物化路径存储），节点携带外观与结构骨架（引用属性模板）；层级最多四级，规范分类编码可自动生成或手动覆盖" />
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="openForm('add')">新增分类模板</el-button>
      </template>

      <template #operate="{ row }">
        <el-button v-if="row.treeLevel < 3" icon="plus" type="primary" link @click="openForm('add', null, { parentId: row.ID, treeRoot: row.treeRoot })">新增下级</el-button>
        <el-button v-if="row.source !== 'builtin'" icon="edit" type="primary" link @click="openForm('edit', row)">编辑</el-button>
        <el-button v-if="row.source !== 'builtin'" icon="delete" type="primary" link :disabled="row.hasChildren" @click="deleteRow(row)">删除</el-button>
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
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="模板编码" prop="templateCode">
          <el-input v-model="form.templateCode" placeholder="唯一标识，如 spray-pump" />
        </el-form-item>
        <el-form-item label="分类编码" prop="classificationCode">
          <el-input v-model="form.classificationCode" placeholder="空则自动生成（如 30-01-01）" />
        </el-form-item>
        <el-form-item label="显示名" prop="label">
          <el-input v-model="form.label" placeholder="如 喷淋泵" />
        </el-form-item>
        <el-form-item label="中文名" prop="labelCn">
          <el-input v-model="form.labelCn" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="父节点" prop="parentId">
          <el-tree-select
            v-model="form.parentId"
            style="width: 100%"
            :data="parentOptions"
            node-key="ID"
            :props="{ label: 'label', children: 'children' }"
            check-strictly
            default-expand-all
            :render-after-expand="false"
            @change="onParentChange"
          />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="图标" prop="icon">
          <el-input v-model="form.icon" placeholder="emoji 或图标类名，如 💧" />
        </el-form-item>
        <el-form-item label="颜色" prop="color">
          <el-color-picker v-model="form.color" />
        </el-form-item>
        <el-form-item label="继承父外观" prop="inheritAppearance">
          <el-switch v-model="form.inheritAppearance" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    getClassTemplateList,
    getClassTemplatePage,
    getClassTemplateTreeRoots,
    findClassTemplate,
    createClassTemplate,
    updateClassTemplate,
    deleteClassTemplate,
    previewClassificationCode
  } from '@/api/ontology/classTemplate'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useAppStore } from '@/pinia'
  import GvaGrid, { useGvaGrid } from '@/components/gvaGrid'

  defineOptions({
    name: 'ClassTemplate'
  })

  const appStore = useAppStore()

  const levelLabels = ['一级', '二级', '三级', '四级']

  // 分类树下拉数据源（搜索表单）
  const treeRootOptions = ref([])
  const loadTreeRoots = async () => {
    const res = await getClassTemplateTreeRoots()
    if (res.code === 0) {
      treeRootOptions.value = (res.data || []).map((item) => ({ label: item.label, value: item.treeRoot }))
    }
  }
  onMounted(loadTreeRoots)

  const { gridRef, gridOptions, gridEvents, refresh, reload } = useGvaGrid({
    id: 'ontology-classTemplate',
    defaultSort: null,
    pageSize: 100,
    pageSizes: [50, 100, 200, 500],
    api: getClassTemplatePage,
    searchItems: [
      {
        field: 'classificationCode',
        title: '分类编码',
        span: 6,
        itemRender: { name: 'VxeInput', props: { placeholder: '向右模糊，如 30-01', clearable: true } }
      },
      {
        field: 'name',
        title: '名称',
        span: 6,
        itemRender: { name: 'VxeInput', props: { placeholder: '模糊匹配显示名/中文名', clearable: true } }
      },
      {
        field: 'treeRoot',
        title: '分类树',
        span: 6,
        itemRender: { name: 'VxeSelect', props: { placeholder: '请选择', clearable: true, options: treeRootOptions } }
      }
    ],
    columns: [
      { field: 'label', title: '名称', minWidth: 200 },
      { field: 'classificationCode', title: '编码', width: 120 },
      { field: 'templateCode', title: '模板编码', width: 140 },
      { field: 'treeLevel', title: '层级', width: 80, formatter: ({ cellValue }) => levelLabels[cellValue] || `第${cellValue + 1}级` },
      { field: 'treeRoot', title: '分类树', width: 120, formatter: ({ row }) => row.treeRootLabel || row.treeRoot },
      { field: 'parentId', title: '父级', width: 140, formatter: ({ row }) => row.parentLabel || '—' },
      { field: 'icon', title: '图标', width: 60 },
      { field: 'source', title: '来源', width: 80, cellRender: { name: 'gvaDict', props: { dict: 'ont_source' } } },
      { field: 'CreatedAt', title: '创建时间', width: 170, cellRender: { name: 'gvaDate' } },
      { title: '操作', width: 220, fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  // 新增/编辑表单
  const formVisible = ref(false)
  const titleForm = ref('新增分类模板')
  const dialogType = ref('add')
  const formRef = ref(null)
  const form = ref({})
  const rules = ref({
    templateCode: [{ required: true, message: '请输入模板编码', trigger: 'blur' }],
    label: [{ required: true, message: '请输入显示名', trigger: 'blur' }],
    parentId: [{ required: true, message: '请选择父节点', trigger: 'change' }]
  })

  // 全量扁平数据 + 前端组树（父节点选择器数据源）
  const flatList = ref([])
  const parentOptions = ref([])

  const buildTree = (nodes, parentId) => {
    return nodes
      .filter((item) => item.parentId === parentId)
      .map((item) => ({
        ID: item.ID,
        label: `${item.classificationCode ? item.classificationCode + ' ' : ''}${item.label}`,
        children: buildTree(nodes, item.ID)
      }))
  }
  const setParentOptions = () => {
    // 头部插入虚拟根；组树时排除编辑中自身（防止把节点挂到自己/后代下）
    const editable = dialogType.value === 'edit' ? flatList.value.filter((item) => item.ID !== form.value.ID) : flatList.value
    parentOptions.value = [{ ID: 0, label: '顶级（根节点）', children: buildTree(editable, 0) }]
  }
  const loadFlatList = async () => {
    const res = await getClassTemplateList()
    if (res.code === 0) {
      flatList.value = res.data || []
    }
    setParentOptions()
  }

  const initForm = () => {
    formRef.value && formRef.value.resetFields()
    form.value = {
      ID: 0,
      templateCode: '',
      classificationCode: '',
      label: '',
      labelCn: '',
      description: '',
      treeRoot: 'equipment',
      parentId: 0,
      sort: 0,
      icon: '',
      color: '',
      inheritAppearance: 1,
      ontClassTemplateRefs: []
    }
  }
  const closeForm = () => {
    initForm()
    formVisible.value = false
  }
  // type: add/edit；row: 编辑行；preset: 新增下级时预置 {parentId, treeRoot}
  const openForm = async (type, row, preset) => {
    initForm()
    dialogType.value = type
    titleForm.value = type === 'add' ? '新增分类模板' : '编辑分类模板'
    if (type === 'edit') {
      const res = await findClassTemplate({ ID: row.ID })
      if (res.code === 0) {
        form.value = { ...form.value, ...res.data }
      }
      formRef.value && formRef.value.clearValidate()
    } else if (preset) {
      form.value.parentId = preset.parentId
      form.value.treeRoot = preset.treeRoot
    }
    await loadFlatList()
    formVisible.value = true
  }
  // 新增且选了父节点时，若编码为空则预览预填
  const onParentChange = async () => {
    if (dialogType.value === 'add' && form.value.parentId !== 0 && !form.value.classificationCode) {
      const res = await previewClassificationCode({ parentId: form.value.parentId, treeRoot: form.value.treeRoot })
      if (res.code === 0 && res.data) {
        form.value.classificationCode = res.data
      }
    }
  }
  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      const req = { ...form.value }
      const isAdd = dialogType.value === 'add'
      const res = isAdd ? await createClassTemplate(req) : await updateClassTemplate(req)
      if (res.code === 0) {
        ElMessage({ type: 'success', message: isAdd ? '创建成功' : '更新成功' })
        await refresh()
        loadTreeRoots()
        closeForm()
      }
    })
  }
  const deleteRow = (row) => {
    ElMessageBox.confirm(`此操作将删除分类模板「${row.label}」, 是否继续?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        const res = await deleteClassTemplate({ ID: row.ID })
        if (res.code === 0) {
          ElMessage({ type: 'success', message: '删除成功!' })
          reload()
          loadTreeRoots()
        }
      })
      .catch(() => {})
  }
</script>

<style lang="scss" scoped>
  .class-template {
    .el-input-number {
      margin-left: 0;
    }
  }
</style>
