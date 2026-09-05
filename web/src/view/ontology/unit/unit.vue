<template>
  <div class="unit-page">
    <div class="unit-main">
      <div class="left-tree">
        <el-card shadow="never">
          <template #header><span>量纲</span></template>
          <el-tree
            ref="treeRef"
            :data="qkTree"
            node-key="id"
            :props="{ label: 'label', children: 'children' }"
            default-expand-all
            highlight-current
            @node-click="onNodeClick"
          />
        </el-card>
      </div>
      <div class="right-list">
        <warning-bar title="注：单位以 QUDT IRI 为规范身份，乘数/偏移相对量纲内隐式基准（如长度=m、温度=摄氏度）；内置单位不可编辑/删除，可停用" />
        <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
          <template #toolbar-buttons>
            <el-button type="primary" icon="plus" @click="openForm('add')">新增单位</el-button>
          </template>

          <template #status="{ row }">
            <el-tag :type="row.status === 1 ? 'info' : 'success'" size="small">
              {{ row.status === 1 ? '停用' : '正常' }}
            </el-tag>
          </template>

          <template #operate="{ row }">
            <el-button v-if="row.source !== 'builtin'" icon="edit" type="primary" link @click="openForm('edit', row)">编辑</el-button>
            <el-button type="primary" link @click="toggleStatus(row)">{{ row.status === 1 ? '启用' : '停用' }}</el-button>
            <el-button v-if="row.source !== 'builtin'" icon="delete" type="primary" link @click="deleteRow(row)">删除</el-button>
          </template>
        </GvaGrid>
      </div>
    </div>

    <el-card shadow="never" class="convert-panel">
      <span class="title">单位换算试算</span>
      <el-input-number v-model="conv.value" :precision="15" :step="1" style="width: 180px" />
      <el-select v-model="conv.fromIri" filterable placeholder="选择源单位" style="width: 240px">
        <el-option v-for="u in unitOptions" :key="u.qudtIri" :label="`${u.label}（${u.symbol || u.unitCode}）`" :value="u.qudtIri" />
      </el-select>
      <span class="arrow">→</span>
      <el-select v-model="conv.toIri" filterable placeholder="选择目标单位" style="width: 240px">
        <el-option v-for="u in unitOptions" :key="u.qudtIri" :label="`${u.label}（${u.symbol || u.unitCode}）`" :value="u.qudtIri" />
      </el-select>
      <span class="result">= {{ conv.result }}</span>
      <el-text v-if="conv.reason" type="danger">{{ conv.reason }}</el-text>
    </el-card>

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
        <el-form-item label="单位编码" prop="unitCode">
          <el-input v-model="form.unitCode" :disabled="dialogType === 'edit' && form.source === 'builtin'" placeholder="唯一标识，如 KiloGM" />
        </el-form-item>
        <el-form-item label="QUDT IRI" prop="qudtIri">
          <el-input v-model="form.qudtIri" placeholder="http://qudt.org/vocab/unit/KiloGM" />
        </el-form-item>
        <el-form-item label="符号" prop="symbol">
          <el-input v-model="form.symbol" placeholder="如 kg" />
        </el-form-item>
        <el-form-item label="英文名" prop="label">
          <el-input v-model="form.label" placeholder="如 Kilogram" />
        </el-form-item>
        <el-form-item label="中文名" prop="labelCn">
          <el-input v-model="form.labelCn" placeholder="如 千克" />
        </el-form-item>
        <el-form-item label="所属量纲" prop="quantityKindCode">
          <el-select v-model="form.quantityKindCode" filterable placeholder="请选择量纲">
            <el-option v-for="k in quantityKinds" :key="k.quantityKindCode" :label="`${k.labelCn || k.label}（${k.quantityKindCode}）`" :value="k.quantityKindCode" />
          </el-select>
        </el-form-item>
        <el-form-item label="换算乘数" prop="conversionMultiplier">
          <el-input-number v-model="form.conversionMultiplier" :precision="15" :step="0.000001" style="width: 100%" placeholder="相对基准单位（如 0.01）" />
        </el-form-item>
        <el-form-item label="换算偏移" prop="conversionOffset">
          <el-input-number v-model="form.conversionOffset" :precision="15" :step="0.000001" style="width: 100%" placeholder="温度等有偏移单位（默认 0）" />
        </el-form-item>
        <el-form-item label="基准单位" prop="scalingOf">
          <el-input v-model="form.scalingOf" placeholder="基准单位 IRI（可选）" />
        </el-form-item>
        <el-form-item label="UCUM 编码" prop="ucumCode">
          <el-input v-model="form.ucumCode" placeholder="如 kg" />
        </el-form-item>
        <el-form-item label="QUDT 版本" prop="qudtVersion">
          <el-input v-model="form.qudtVersion" placeholder="如 3.1.5" />
        </el-form-item>
      </el-form>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    getQuantityKindList,
    getUnitPage,
    getUnitAll,
    findUnit,
    createUnit,
    updateUnit,
    deleteUnit,
    disableUnit,
    convertUnit
  } from '@/api/ontology/unit'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref, reactive, watch, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useAppStore } from '@/pinia'
  import GvaGrid, { useGvaGrid } from '@/components/gvaGrid'

  defineOptions({
    name: 'Unit'
  })

  const appStore = useAppStore()

  // —— 左侧量纲树 ——
  const treeRef = ref(null)
  const quantityKinds = ref([])
  const selectedQk = ref('__all__')
  const qkTree = ref([{ id: '__all__', label: '全部量纲', children: [] }])
  const loadQuantityKinds = async () => {
    const res = await getQuantityKindList()
    if (res.code === 0) {
      quantityKinds.value = res.data || []
      qkTree.value = [
        {
          id: '__all__',
          label: '全部量纲',
          children: quantityKinds.value.map((k) => ({ id: k.quantityKindCode, label: k.labelCn || k.label }))
        }
      ]
    }
  }
  const onNodeClick = (node) => {
    selectedQk.value = node.id
    reload()
  }

  const { gridRef, gridOptions, gridEvents, refresh, reload } = useGvaGrid({
    id: 'ontology-unit',
    defaultSort: { field: 'ID', order: 'desc' },
    api: async (params) => {
      // 左树量纲优先于搜索表单同名字段
      const qk = selectedQk.value !== '__all__' ? selectedQk.value : undefined
      return getUnitPage({ ...params, quantityKindCode: qk })
    },
    searchItems: [
      {
        field: 'keyword',
        title: '关键字',
        span: 6,
        itemRender: { name: 'VxeInput', props: { placeholder: '编码/名称/符号', clearable: true } }
      },
      {
        field: 'source',
        title: '来源',
        span: 6,
        itemRender: { name: 'gvaDictSelect', props: { dict: 'ont_source', placeholder: '请选择', clearable: true } }
      },
      {
        field: 'status',
        title: '状态',
        span: 6,
        itemRender: {
          name: 'VxeSelect',
          props: {
            placeholder: '请选择',
            clearable: true,
            options: [
              { label: '正常', value: 0 },
              { label: '停用', value: 1 }
            ]
          }
        }
      }
    ],
    columns: [
      { field: 'unitCode', title: '编码', width: 100 },
      { field: 'symbol', title: '符号', width: 70 },
      { field: 'label', title: '名称', width: 120 },
      { field: 'labelCn', title: '中文名', width: 100 },
      { field: 'quantityKindCode', title: '量纲', width: 100 },
      { field: 'conversionMultiplier', title: '换算系数', width: 120 },
      { field: 'conversionOffset', title: '偏移', width: 100 },
      { field: 'scalingOf', title: '基准单位', width: 160, showOverflow: true },
      { field: 'ucumCode', title: 'UCUM', width: 80 },
      { field: 'source', title: '来源', width: 80, cellRender: { name: 'gvaDict', props: { dict: 'ont_source' } } },
      { field: 'status', title: '状态', width: 80, slots: { default: 'status' } },
      { title: '操作', width: 200, fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  // —— 换算试算面板（防抖 300ms）——
  const unitOptions = ref([])
  const loadUnitOptions = async () => {
    const res = await getUnitAll()
    if (res.code === 0) {
      unitOptions.value = res.data || []
    }
  }
  const conv = reactive({ value: 1, fromIri: '', toIri: '', result: '--', reason: '' })
  let convTimer = null
  const trimZeros = (v) => {
    const s = String(v)
    return s.includes('.') ? s.replace(/0+$/, '').replace(/\.$/, '') : s
  }
  const doConvert = async () => {
    if (conv.value === undefined || conv.value === null || !conv.fromIri || !conv.toIri) {
      conv.result = '--'
      conv.reason = ''
      return
    }
    const res = await convertUnit({ value: String(conv.value), fromIri: conv.fromIri, toIri: conv.toIri })
    if (res.code === 0 && res.data) {
      if (res.data.result !== null && res.data.result !== undefined) {
        conv.result = trimZeros(res.data.result)
        conv.reason = ''
      } else {
        conv.result = '无法换算'
        conv.reason = res.data.reason || ''
      }
    }
  }
  watch(
    () => [conv.value, conv.fromIri, conv.toIri],
    () => {
      if (convTimer) clearTimeout(convTimer)
      convTimer = setTimeout(doConvert, 300)
    }
  )

  // —— 新增/编辑表单 ——
  const formVisible = ref(false)
  const titleForm = ref('新增单位')
  const dialogType = ref('add')
  const formRef = ref(null)
  const form = ref({})
  const rules = ref({
    unitCode: [{ required: true, message: '请输入单位编码', trigger: 'blur' }],
    qudtIri: [{ required: true, message: '请输入 QUDT IRI', trigger: 'blur' }],
    label: [{ required: true, message: '请输入英文名', trigger: 'blur' }],
    quantityKindCode: [{ required: true, message: '请选择量纲', trigger: 'change' }]
  })

  const initForm = () => {
    formRef.value && formRef.value.resetFields()
    form.value = {
      ID: 0,
      unitCode: '',
      qudtIri: '',
      symbol: '',
      label: '',
      labelCn: '',
      quantityKindCode: selectedQk.value !== '__all__' ? selectedQk.value : '',
      conversionMultiplier: 1,
      conversionOffset: 0,
      scalingOf: '',
      ucumCode: '',
      qudtVersion: '',
      source: 'custom',
      status: 0
    }
  }
  const closeForm = () => {
    initForm()
    formVisible.value = false
  }
  const openForm = async (type, row) => {
    initForm()
    dialogType.value = type
    titleForm.value = type === 'add' ? '新增单位' : '编辑单位'
    if (type === 'edit') {
      const res = await findUnit({ ID: row.ID })
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
      const res = isAdd ? await createUnit(req) : await updateUnit(req)
      if (res.code === 0) {
        ElMessage({ type: 'success', message: isAdd ? '创建成功' : '更新成功' })
        await refresh()
        loadUnitOptions()
        closeForm()
      }
    })
  }
  const toggleStatus = (row) => {
    disableUnit({ ID: row.ID }).then((res) => {
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: row.status === 1 ? `已启用「${row.label}」` : `已停用「${row.label}」`
        })
        refresh()
        loadUnitOptions()
      }
    })
  }
  const deleteRow = (row) => {
    ElMessageBox.confirm(`此操作将删除单位「${row.label}」, 是否继续?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        const res = await deleteUnit({ ID: row.ID })
        if (res.code === 0) {
          ElMessage({ type: 'success', message: '删除成功!' })
          refresh()
          loadUnitOptions()
        }
      })
      .catch(() => {})
  }

  onMounted(() => {
    loadQuantityKinds()
    loadUnitOptions()
  })
</script>

<style lang="scss" scoped>
  .unit-page {
    .unit-main {
      display: flex;
      gap: 12px;

      .left-tree {
        width: 220px;
        flex-shrink: 0;
      }

      .right-list {
        flex: 1;
        min-width: 0;
      }
    }

    .convert-panel {
      margin-top: 12px;

      .title {
        font-weight: bold;
        margin-right: 16px;
      }

      .arrow {
        margin: 0 8px;
      }

      .result {
        margin-left: 16px;
        font-weight: bold;
      }

      .el-text {
        margin-left: 12px;
      }
    }
  }
</style>
