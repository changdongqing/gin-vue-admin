<template>
  <div class="flex gap-3">
    <!-- 左卡：外部模块 -->
    <el-card class="w-[280px] shrink-0" shadow="never">
      <template #header>
        <div class="flex justify-between items-center">
          <span>外部模块</span>
          <el-button type="primary" icon="plus" link @click="openModuleForm()">新增</el-button>
        </div>
      </template>
      <div
        v-for="m in modules"
        :key="m.ID"
        class="p-2 rounded cursor-pointer mb-1"
        :class="m.ID === currentModuleId ? 'bg-blue-50 text-blue-600' : 'hover:bg-gray-50'"
        @click="selectModule(m)"
      >
        <div class="flex justify-between items-center">
          <span class="font-medium">{{ m.name }}</span>
          <el-button v-if="m.ID === currentModuleId" icon="delete" type="primary" link @click.stop="deleteModule(m)" />
        </div>
        <div class="text-xs text-gray-400">{{ m.moduleCode }} · {{ m.tableCount }} 表</div>
      </div>
      <el-empty v-if="!modules.length" description="暂无模块" :image-size="48" />
    </el-card>

    <!-- 右卡：表注册 -->
    <el-card class="flex-1" shadow="never">
      <template #header>
        <div class="flex justify-between items-center">
          <span>表注册{{ currentModule ? `（${currentModule.name}）` : '' }}</span>
          <el-button type="primary" icon="search" :disabled="!currentModuleId" @click="openProbe">注册表（探测）</el-button>
        </div>
      </template>
      <el-table :data="tables" border>
        <el-table-column prop="tableName" label="表名" min-width="180" show-overflow-tooltip />
        <el-table-column prop="displayName" label="显示名" width="140" />
        <el-table-column prop="pkColumn" label="主键列" width="100" />
        <el-table-column prop="updateTimeColumn" label="水位列" width="120" />
        <el-table-column label="增量" width="110" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.supportsIncremental === 1" type="success">是</el-tag>
            <el-tag v-else type="info">否（仅全量）</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
        <el-table-column label="操作" width="80" fixed="right" align="center">
          <template #default="{ row }">
            <el-popconfirm title="有类绑定时删除将被拒绝，确定？" @confirm="deleteTable(row)">
              <template #reference>
                <el-button icon="delete" type="primary" link>删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!currentModuleId" description="请选择左侧模块" :image-size="48" />
    </el-card>

    <!-- 模块表单弹窗 -->
    <el-dialog v-model="moduleFormVisible" :title="moduleFormType === 'add' ? '新增模块' : '编辑模块'" width="460px">
      <el-form ref="moduleFormRef" :model="moduleForm" :rules="moduleRules" label-width="90px">
        <el-form-item label="模块编码" prop="moduleCode">
          <el-input v-model="moduleForm.moduleCode" placeholder="如 system / infra / iot" />
        </el-form-item>
        <el-form-item label="显示名" prop="name">
          <el-input v-model="moduleForm.name" />
        </el-form-item>
        <el-form-item label="所属服务" prop="serviceName">
          <el-input v-model="moduleForm.serviceName" placeholder="可选" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="moduleFormVisible = false">取 消</el-button>
        <el-button type="primary" @click="submitModule">确 定</el-button>
      </template>
    </el-dialog>

    <!-- 探测弹窗 -->
    <el-dialog v-model="probeVisible" title="注册表（探测 public schema）" width="640px">
      <div class="flex gap-2 mb-2">
        <el-input v-model="probeKeyword" placeholder="表名关键字，如 demo" clearable @keyup.enter="loadProbe" />
        <el-button type="primary" @click="loadProbe">搜索</el-button>
      </div>
      <el-table
        ref="probeTableRef"
        :data="probeList"
        border
        max-height="360"
        row-key="tableName"
        @selection-change="(rows) => (probeSelection = rows)"
      >
        <el-table-column type="selection" width="42" :selectable="(row) => !row.registered" />
        <el-table-column prop="tableName" label="表名" min-width="180" show-overflow-tooltip />
        <el-table-column prop="tableComment" label="注释" min-width="180" show-overflow-tooltip />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.registered" type="info">已注册</el-tag>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="probeVisible = false">取 消</el-button>
        <el-button type="primary" :disabled="!probeSelection.length" @click="submitRegister">
          注册选中（{{ probeSelection.length }}）
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import {
    getExtModuleList,
    createExtModule,
    deleteExtModule,
    getExtTableList,
    registerExtTables,
    deleteExtTable,
    probeExtTables
  } from '@/api/ontology/extRegistry'
  import { ref, reactive } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'

  defineOptions({ name: 'ExtRegistry' })

  const modules = ref([])
  const currentModuleId = ref(0)
  const currentModule = ref(null)
  const tables = ref([])

  const loadModules = async () => {
    const res = await getExtModuleList()
    if (res.code === 0) {
      modules.value = res.data || []
      if (currentModuleId.value && !modules.value.some((m) => m.ID === currentModuleId.value)) {
        currentModuleId.value = 0
        tables.value = []
      }
    }
  }
  loadModules()

  const selectModule = async (m) => {
    currentModule.value = m
    currentModuleId.value = m.ID
    const res = await getExtTableList({ moduleId: m.ID })
    if (res.code === 0) {
      tables.value = res.data || []
    }
  }
  const refreshCurrent = () => {
    loadModules()
    if (currentModuleId.value) {
      selectModule(currentModule.value || { ID: currentModuleId.value })
    }
  }

  // 模块表单
  const moduleFormVisible = ref(false)
  const moduleFormType = ref('add')
  const moduleFormRef = ref(null)
  const moduleForm = ref({})
  const moduleRules = ref({
    moduleCode: [{ required: true, message: '请输入模块编码', trigger: 'blur' }],
    name: [{ required: true, message: '请输入显示名', trigger: 'blur' }]
  })
  const openModuleForm = () => {
    moduleFormType.value = 'add'
    moduleForm.value = { moduleCode: '', name: '', serviceName: '' }
    moduleFormVisible.value = true
  }
  const submitModule = () => {
    moduleFormRef.value.validate(async (valid) => {
      if (!valid) return
      const res = await createExtModule({ ...moduleForm.value })
      if (res.code === 0) {
        ElMessage.success('创建成功')
        moduleFormVisible.value = false
        loadModules()
      }
    })
  }
  const deleteModule = (m) => {
    ElMessageBox.confirm(`删除模块「${m.name}」？有注册表时将被拒绝`, '提示', { type: 'warning' })
      .then(async () => {
        const res = await deleteExtModule({ ID: m.ID })
        if (res.code === 0) {
          ElMessage.success('删除成功')
          if (currentModuleId.value === m.ID) {
            currentModuleId.value = 0
            currentModule.value = null
            tables.value = []
          }
          loadModules()
        }
      })
      .catch(() => {})
  }

  // 探测注册
  const probeVisible = ref(false)
  const probeKeyword = ref('')
  const probeList = ref([])
  const probeSelection = ref([])
  const probeTableRef = ref(null)
  const openProbe = () => {
    probeKeyword.value = ''
    probeList.value = []
    probeSelection.value = []
    probeVisible.value = true
  }
  const loadProbe = async () => {
    const res = await probeExtTables({ keyword: probeKeyword.value })
    if (res.code === 0) {
      probeList.value = res.data || []
    }
  }
  const submitRegister = async () => {
    const res = await registerExtTables({
      moduleId: currentModuleId.value,
      items: probeSelection.value.map((r) => ({ tableName: r.tableName, displayName: '', remark: '' }))
    })
    if (res.code === 0) {
      ElMessage.success(`已注册 ${res.data.registered} 张表（已注册表自动跳过）`)
      probeVisible.value = false
      refreshCurrent()
    }
  }
  const deleteTable = async (row) => {
    const res = await deleteExtTable({ ID: row.ID })
    if (res.code === 0) {
      ElMessage.success('删除成功')
      refreshCurrent()
    }
  }
  defineExpose({ refresh: refreshCurrent })
  void reactive
</script>
