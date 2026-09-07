<template>
  <div class="report-data-source">
    <warning-bar title="注：数据源是报表平台取数底座，管理外部数据库连接配置；密码加密存储、回显脱敏；编辑态编码/类型不可修改" />
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="openForm('add')">新增数据源</el-button>
      </template>

      <template #enableFlag="{ row }">
        <el-tag :type="row.enableFlag ? 'success' : 'info'" size="small">{{ row.enableFlag ? '启用' : '禁用' }}</el-tag>
      </template>

      <template #operate="{ row }">
        <el-button icon="edit" type="primary" link @click="openForm('edit', row)">编辑</el-button>
        <el-button icon="delete" type="primary" link @click="deleteRow(row)">删除</el-button>
      </template>
    </GvaGrid>

    <!-- 数据源表单抽屉 -->
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
        <el-form-item label="数据源编码" prop="sourceCode">
          <el-input v-model="form.sourceCode" :disabled="dialogType === 'edit'" placeholder="唯一标识，如 biz_pg" />
        </el-form-item>
        <el-form-item label="数据源名称" prop="sourceName">
          <el-input v-model="form.sourceName" placeholder="如 业务数据库" />
        </el-form-item>
        <el-form-item label="数据源类型" prop="sourceType">
          <el-select v-model="form.sourceType" :disabled="dialogType === 'edit'" placeholder="请选择类型" @change="handleTypeChange">
            <el-option-group v-for="group in DATA_SOURCE_TYPE_GROUPS" :key="group" :label="group">
              <el-option
                v-for="item in typeOptionsByGroup(group)"
                :key="item.value"
                :label="item.enabled ? item.label : `${item.label}（驱动未启用）`"
                :value="item.value"
                :disabled="!item.enabled"
              />
            </el-option-group>
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="sourceDesc">
          <el-input v-model="form.sourceDesc" type="textarea" :rows="2" />
        </el-form-item>
        <template v-if="form.sourceType !== 'http'">
          <el-form-item label="连接配置" prop="sourceConfig">
            <el-input
              v-model="form.sourceConfig"
              type="textarea"
              :rows="6"
              maxlength="2048"
              show-word-limit
              class="font-mono"
              placeholder='{"dsn":"postgres://user:pass@127.0.0.1:5432/your_db","username":"","password":""}'
            />
          </el-form-item>
          <el-form-item v-if="form.sourceType" label=" ">
            <el-button icon="magic-stick" @click="fillTemplate">填充模板</el-button>
            <span class="ml-2 text-xs text-gray-400">独立 username/password 优先于 dsn 内嵌凭据；编辑态密码留空或 ****** 表示保留原密码</span>
          </el-form-item>
        </template>
        <el-form-item v-else label=" ">
          <el-alert type="info" :closable="false" title="HTTP 数据集无需连接配置，直接在数据集中配置请求地址" />
        </el-form-item>
        <el-form-item label="是否启用" prop="enableFlag">
          <el-switch v-model="form.enableFlag" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="flex justify-end">
          <el-button type="primary" plain :loading="testing" @click="handleTestConnection">测试连接</el-button>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    getDataSourceList,
    findDataSource,
    createDataSource,
    updateDataSource,
    deleteDataSource,
    testDataSourceConnection,
    DATA_SOURCE_TYPE_OPTIONS,
    DATA_SOURCE_TYPE_GROUPS
  } from '@/api/report/dataSource'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useAppStore } from '@/pinia'
  import GvaGrid, { useGvaGrid } from '@/components/gvaGrid'

  defineOptions({
    name: 'ReportDataSource'
  })

  const appStore = useAppStore()

  const typeLabel = (value) => DATA_SOURCE_TYPE_OPTIONS.find((o) => o.value === value)?.label || value
  const typeOptionsByGroup = (group) => DATA_SOURCE_TYPE_OPTIONS.filter((o) => o.group === group)

  const { gridRef, gridOptions, gridEvents, refresh } = useGvaGrid({
    id: 'report-dataSource',
    defaultSort: null,
    api: getDataSourceList,
    searchItems: [
      {
        field: 'keyword',
        title: '关键字',
        span: 6,
        itemRender: { name: 'VxeInput', props: { placeholder: '数据源编码或名称', clearable: true } }
      },
      {
        field: 'sourceType',
        title: '类型',
        span: 6,
        itemRender: {
          name: 'VxeSelect',
          props: {
            placeholder: '不选则查全部',
            clearable: true,
            options: DATA_SOURCE_TYPE_OPTIONS.map((o) => ({ label: o.label, value: o.value }))
          }
        }
      },
      {
        field: 'enableFlag',
        title: '状态',
        span: 6,
        itemRender: {
          name: 'VxeSelect',
          props: {
            placeholder: '不选则查全部',
            clearable: true,
            options: [
              { label: '启用', value: 'true' },
              { label: '禁用', value: 'false' }
            ]
          }
        }
      }
    ],
    columns: [
      { field: 'sourceCode', title: '数据源编码', width: 140 },
      { field: 'sourceName', title: '数据源名称', width: 150 },
      { field: 'sourceType', title: '类型', width: 110, formatter: ({ cellValue }) => typeLabel(cellValue) },
      { field: 'sourceDesc', title: '描述', minWidth: 160, showOverflow: true },
      { field: 'enableFlag', title: '状态', width: 80, slots: { default: 'enableFlag' } },
      { field: 'CreatedAt', title: '创建时间', width: 170, cellRender: { name: 'gvaDate' } },
      { title: '操作', width: 140, fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  // ── 表单 ────────────────────────────────
  const formVisible = ref(false)
  const titleForm = ref('新增数据源')
  const dialogType = ref('add')
  const formRef = ref(null)
  const form = ref({})
  const testing = ref(false)

  // 连接配置校验：http 类型跳过；其余须为合法 JSON 且含 dsn
  const validateSourceConfig = (rule, value, callback) => {
    if (form.value.sourceType === 'http') return callback()
    if (!value) return callback(new Error('请填写连接配置（可用「填充模板」生成骨架）'))
    try {
      const cfg = JSON.parse(value)
      if (!cfg.dsn || !String(cfg.dsn).trim()) return callback(new Error('连接配置须包含 dsn 字段'))
      callback()
    } catch (e) {
      callback(new Error('连接配置必须是合法 JSON'))
    }
  }

  const rules = ref({
    sourceCode: [
      { required: true, message: '请输入数据源编码', trigger: 'blur' },
      { pattern: /^[a-zA-Z0-9_]+$/, message: '仅允许字母/数字/下划线', trigger: 'blur' }
    ],
    sourceName: [{ required: true, message: '请输入数据源名称', trigger: 'blur' }],
    sourceType: [{ required: true, message: '请选择数据源类型', trigger: 'change' }],
    sourceConfig: [{ validator: validateSourceConfig, trigger: 'blur' }]
  })

  const initForm = () => {
    formRef.value && formRef.value.resetFields()
    form.value = {
      ID: 0,
      sourceCode: '',
      sourceName: '',
      sourceType: '',
      sourceDesc: '',
      sourceConfig: '',
      enableFlag: true
    }
  }

  const handleTypeChange = (type) => {
    if (type === 'http') {
      form.value.sourceConfig = ''
    }
  }

  // 按当前类型生成 {dsn, username, password} 配置骨架
  const fillTemplate = () => {
    const opt = DATA_SOURCE_TYPE_OPTIONS.find((o) => o.value === form.value.sourceType)
    if (!opt || !opt.dsnTemplate) return
    form.value.sourceConfig = JSON.stringify({ dsn: opt.dsnTemplate, username: '', password: '' }, null, 2)
  }

  const closeForm = () => {
    initForm()
    formVisible.value = false
  }

  const openForm = async (type, row) => {
    initForm()
    dialogType.value = type
    titleForm.value = type === 'add' ? '新增数据源' : '编辑数据源'
    if (type === 'edit') {
      const res = await findDataSource({ ID: row.ID })
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
      if (form.value.sourceType !== 'http' && !form.value.sourceConfig) {
        // 兼容未挂规则校验的场景（类型为空时跳过 JSON 校验）
        if (!tryParseConfig()) return
      }
      const req = { ...form.value }
      const isAdd = dialogType.value === 'add'
      const res = isAdd ? await createDataSource(req) : await updateDataSource(req)
      if (res.code === 0) {
        ElMessage({ type: 'success', message: isAdd ? '创建成功' : '更新成功' })
        await refresh()
        closeForm()
      }
    })
  }

  const tryParseConfig = () => {
    try {
      const cfg = JSON.parse(form.value.sourceConfig)
      if (!cfg.dsn || !String(cfg.dsn).trim()) {
        ElMessage.error('连接配置须包含 dsn 字段')
        return false
      }
      return true
    } catch (e) {
      ElMessage.error('连接配置必须是合法 JSON')
      return false
    }
  }

  // 测试连接：请求体直接传当前表单值（不先保存）；测试不通过不阻断保存
  const handleTestConnection = async () => {
    if (form.value.sourceType !== 'http' && !tryParseConfig()) return
    testing.value = true
    try {
      const res = await testDataSourceConnection({
        sourceType: form.value.sourceType,
        sourceConfig: form.value.sourceType === 'http' ? '' : form.value.sourceConfig
      })
      if (res.code === 0) {
        ElMessage.success('连接成功')
      }
    } finally {
      testing.value = false
    }
  }

  const deleteRow = (row) => {
    ElMessageBox.confirm(`此操作将删除数据源「${row.sourceName}」，连接池随之释放，是否继续?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        const res = await deleteDataSource({ ID: row.ID })
        if (res.code === 0) {
          ElMessage({ type: 'success', message: '删除成功!' })
          refresh()
        }
      })
      .catch(() => {})
  }
</script>

<style lang="scss" scoped>
  .report-data-source {
    .el-select {
      width: 100%;
    }

    .font-mono :deep(textarea) {
      font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
      font-size: 12px;
    }
  }
</style>
