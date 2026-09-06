<template>
  <el-drawer v-model="visible" size="960px" title="绑定编辑（保存为草稿，生效请在列表操作）">
    <!-- 第一段：基本信息（类 ↔ 主表） -->
    <el-divider content-position="left">基本信息</el-divider>
    <el-form label-width="120px" inline>
      <el-form-item label="本体项目" required>
        <el-select v-model="form.projectId" filterable placeholder="选择项目" style="width: 220px" @change="onProjectChange">
          <el-option v-for="p in projects" :key="p.ID" :label="p.projectCode" :value="p.ID" />
        </el-select>
      </el-form-item>
      <el-form-item label="本体类（可实例化）" required>
        <el-select v-model="form.classId" filterable placeholder="先选项目" style="width: 240px" @change="onClassChange">
          <el-option
            v-for="c in classList"
            :key="c.ID"
            :label="`${c.localName}（${c.labelCn || c.label || '-'}）`"
            :value="c.ID"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="业务主表" required>
        <el-select v-model="form.tableId" filterable placeholder="选择注册表" style="width: 220px" @change="onTableChange">
          <el-option v-for="t in allTables" :key="t.ID" :label="t.tableName" :value="t.ID" />
        </el-select>
      </el-form-item>
      <el-form-item label="主键列">
        <el-input v-model="form.keyColumn" style="width: 140px" />
      </el-form-item>
      <el-form-item label="编码列">
        <el-select v-model="form.codeColumn" clearable filterable style="width: 140px" placeholder="空=类名-主键">
          <el-option v-for="c in mainColumns" :key="c" :label="c" :value="c" />
        </el-select>
      </el-form-item>
      <el-form-item label="名称列">
        <el-select v-model="form.nameColumn" clearable filterable style="width: 140px">
          <el-option v-for="c in mainColumns" :key="c" :label="c" :value="c" />
        </el-select>
      </el-form-item>
      <el-form-item label="父列">
        <el-select v-model="form.parentColumn" clearable filterable style="width: 140px">
          <el-option v-for="c in mainColumns" :key="c" :label="c" :value="c" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态列">
        <el-select v-model="form.statusColumn" clearable filterable style="width: 140px">
          <el-option v-for="c in mainColumns" :key="c" :label="c" :value="c" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态在用值">
        <el-input v-model="form.statusActiveValue" style="width: 140px" placeholder="如 0" />
      </el-form-item>
      <el-form-item label="同步模式">
        <el-select v-model="form.syncMode" style="width: 140px">
          <el-option label="手动" :value="1" />
          <el-option label="定时+手动" :value="2" />
        </el-select>
      </el-form-item>
      <el-form-item label="编排顺序">
        <el-input-number v-model="form.syncOrder" :min="1" style="width: 140px" />
      </el-form-item>
      <el-form-item label="含停用行">
        <el-switch v-model="form.includeDisabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item label="值漂移策略">
        <el-select v-model="form.conflictStrategy" style="width: 140px">
          <el-option label="业务表覆盖" :value="2" />
          <el-option label="仅报告" :value="1" />
        </el-select>
      </el-form-item>
      <el-form-item label="目标缺失策略">
        <el-select v-model="form.missingTargetPolicy" style="width: 140px">
          <el-option label="跳过并记issue" :value="1" />
          <el-option label="整行失败" :value="2" />
        </el-select>
      </el-form-item>
    </el-form>

    <!-- 第二段：子表绑定 -->
    <el-divider content-position="left">子表绑定（宽表一对一 / 窄表 EAV）</el-divider>
    <el-button type="primary" icon="plus" plain size="small" class="mb-2" @click="addDetail">添加子表绑定</el-button>
    <el-table :data="details" border size="small">
      <el-table-column label="子表" min-width="160">
        <template #default="{ row }">
          <el-select v-model="row.tableId" filterable placeholder="注册表">
            <el-option v-for="t in allTables" :key="t.ID" :label="t.tableName" :value="t.ID" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="形态" width="140">
        <template #default="{ row }">
          <el-select v-model="row.detailKind">
            <el-option label="宽表一对一" :value="1" />
            <el-option label="窄表EAV" :value="2" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="回连外键列" min-width="140">
        <template #default="{ row }">
          <el-select v-model="row.joinColumn" clearable filterable placeholder="该子表列">
            <el-option v-for="c in detailColumns(row.tableId)" :key="c" :label="c" :value="c" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="行键列（窄表）" min-width="130">
        <template #default="{ row }">
          <el-select v-if="row.detailKind === 2" v-model="row.identifierColumn" clearable filterable>
            <el-option v-for="c in detailColumns(row.tableId)" :key="c" :label="c" :value="c" />
          </el-select>
          <span v-else>—</span>
        </template>
      </el-table-column>
      <el-table-column label="值列（窄表）" min-width="130">
        <template #default="{ row }">
          <el-select v-if="row.detailKind === 2" v-model="row.valueColumn" clearable filterable>
            <el-option v-for="c in detailColumns(row.tableId)" :key="c" :label="c" :value="c" />
          </el-select>
          <span v-else>—</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="70" align="center">
        <template #default="{ $index }">
          <el-button icon="delete" type="primary" link @click="details.splice($index, 1)" />
        </template>
      </el-table-column>
    </el-table>

    <!-- 第三段：属性绑定 -->
    <el-divider content-position="left">属性绑定（选类后自动补全）</el-divider>
    <el-button type="primary" icon="refresh" plain size="small" class="mb-2" :disabled="!form.classId" @click="autoFillProperties">
      按类属性重建行
    </el-button>
    <el-table :data="properties" border size="small" max-height="280">
      <el-table-column label="属性" min-width="150">
        <template #default="{ row }">{{ row.propertyLabel }}</template>
      </el-table-column>
      <el-table-column label="类型" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="row.bindingType === 2 ? 'warning' : 'primary'" size="small">
            {{ row.bindingType === 2 ? '对象' : '数据' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="值来源" min-width="170">
        <template #default="{ row }">
          <el-select v-if="row.bindingType === 1" v-model="row.detailBindingId" placeholder="主表或宽表子表">
            <el-option :label="`主表（${mainTableName || '-'}）`" :value="0" />
            <el-option
              v-for="d in wideDetails"
              :key="d.ID"
              :label="`宽表（${detailTableName(d.tableId)}）`"
              :value="d.ID"
            />
          </el-select>
          <span v-else>主表（对象属性仅限主表外键列）</span>
        </template>
      </el-table-column>
      <el-table-column label="业务列/匹配值" min-width="160">
        <template #default="{ row }">
          <el-select
            v-if="!(row.bindingType === 1 && isNarrowSource(row))"
            v-model="row.bizColumn"
            clearable
            filterable
            allow-create
            placeholder="选择列"
          >
            <el-option v-for="c in sourceColumns(row)" :key="c.column" :label="`${c.column}（${c.dataType}）`" :value="c.column" />
          </el-select>
          <el-input v-else v-model="row.bizColumn" placeholder="identifier 匹配值，如 temperature" />
        </template>
      </el-table-column>
      <el-table-column label="启用" width="70" align="center">
        <template #default="{ row }">
          <el-checkbox v-model="row.enabledBool" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="70" align="center">
        <template #default="{ $index }">
          <el-button icon="delete" type="primary" link @click="properties.splice($index, 1)" />
        </template>
      </el-table-column>
    </el-table>

    <template #footer>
      <div class="flex justify-between">
        <el-button type="warning" plain :disabled="!savedId" @click="$emit('dry-run', savedId)">试运行（预览前 20 行）</el-button>
        <div>
          <el-button @click="visible = false">取 消</el-button>
          <el-button type="primary" :loading="saving" @click="save">保存（草稿）</el-button>
        </div>
      </div>
    </template>
  </el-drawer>
</template>

<script setup>
  import {
    getExtModuleList,
    getExtTableList,
    getExtTableColumns
  } from '@/api/ontology/extRegistry'
  import { getModelProjectAll, getModelClassByProject, getModelClassDetail } from '@/api/ontology/modelClass'
  import { createExtBinding, updateExtBinding, findExtBinding } from '@/api/ontology/extBinding'
  import { ref, computed } from 'vue'
  import { ElMessage } from 'element-plus'

  defineOptions({ name: 'ExtBindingForm' })

  const emit = defineEmits(['saved', 'dry-run'])

  // xsd ↔ PG 兼容组（与后端附录A一致；列下拉过滤用）
  const COMPAT = {
    'xsd:string': ['character varying', 'character', 'text', 'jsonb', 'uuid'],
    'xsd:integer': ['smallint', 'integer', 'bigint'],
    'xsd:decimal': ['numeric', 'real', 'double precision', 'money'],
    'xsd:boolean': ['boolean', 'smallint'],
    'xsd:datetime': ['timestamp without time zone', 'timestamp with time zone', 'date']
  }

  const visible = ref(false)
  const saving = ref(false)
  const savedId = ref(0)
  const projects = ref([])
  const classList = ref([])
  const allTables = ref([])
  const mainColumns = ref([])
  const columnCache = ref({}) // tableId → [{column,dataType}]
  const details = ref([])
  const properties = ref([])
  const form = ref({})
  const mainTableName = computed(() => allTables.value.find((t) => t.ID === form.value.tableId)?.tableName || '')
  const wideDetails = computed(() => details.value.filter((d) => d.detailKind === 1 && d.ID))

  const initForm = () => ({
    ID: 0,
    projectId: null,
    classId: null,
    tableId: null,
    keyColumn: 'id',
    codeColumn: '',
    nameColumn: '',
    parentColumn: '',
    statusColumn: '',
    statusActiveValue: '',
    syncMode: 1,
    syncOrder: 100,
    includeDisabled: 0,
    conflictStrategy: 2,
    missingTargetPolicy: 1
  })

  const loadTables = async () => {
    const res = await getExtModuleList()
    const tables = []
    if (res.code === 0) {
      for (const m of res.data || []) {
        const r = await getExtTableList({ moduleId: m.ID })
        if (r.code === 0) tables.push(...(r.data || []))
      }
    }
    allTables.value = tables
  }

  const open = async (row = null) => {
    form.value = initForm()
    details.value = []
    properties.value = []
    classList.value = []
    mainColumns.value = []
    columnCache.value = {}
    savedId.value = 0
    visible.value = true
    const [prjRes] = await Promise.all([getModelProjectAll(), loadTables()])
    if (prjRes.code === 0) projects.value = prjRes.data || []
    if (row) {
      const res = await findExtBinding({ ID: row.ID })
      if (res.code === 0) {
        form.value = { ...form.value, ...res.data.binding }
        details.value = (res.data.details || []).map((d) => ({ ...d, enabledBool: true }))
        for (const d of details.value) loadColumns(d.tableId)
        properties.value = (res.data.properties || []).map((p) => ({ ...p, enabledBool: p.enabled === 1 }))
        savedId.value = row.ID
        await onProjectChange(form.value.projectId, true)
        await onClassChange(form.value.classId, true)
        await onTableChange(form.value.tableId, true)
      }
    }
  }
  defineExpose({ open })

  const onProjectChange = async (projectId, silent = false) => {
    if (!projectId) return
    const res = await getModelClassByProject({ projectId })
    if (res.code === 0) {
      classList.value = (res.data || []).filter((c) => c.isInstantiable === 1)
    }
    if (!silent) form.value.classId = null
  }

  const onClassChange = async (classId, silent = false) => {
    if (!classId) return
    const res = await getModelClassDetail({ ID: classId })
    if (res.code === 0) {
      classPropSource.value = res.data
    }
    if (!silent) autoFillProperties()
  }

  const classPropSource = ref(null)

  const onTableChange = async (tableId, silent = false) => {
    if (!tableId) return
    const t = allTables.value.find((x) => x.ID === tableId)
    if (t) form.value.keyColumn = t.pkColumn || 'id'
    await loadColumns(tableId)
    const cols = columnCache.value[tableId] || []
    mainColumns.value = cols.map((c) => c.column)
    if (!silent) return
  }

  const loadColumns = async (tableId) => {
    if (!tableId || columnCache.value[tableId]) return
    const t = allTables.value.find((x) => x.ID === tableId)
    if (!t) return
    const res = await getExtTableColumns({ tableName: t.tableName })
    if (res.code === 0) {
      columnCache.value = { ...columnCache.value, [tableId]: res.data || [] }
    }
  }
  const detailColumns = (tableId) => (columnCache.value[tableId] || []).map((c) => c.column)
  const detailTableName = (tableId) => allTables.value.find((t) => t.ID === tableId)?.tableName || '-'

  const addDetail = async () => {
    const d = {
      ID: 0,
      tableId: null,
      detailKind: 1,
      joinColumn: '',
      identifierColumn: '',
      valueColumn: '',
      tsColumn: ''
    }
    details.value.push(d)
  }
  // 行编辑时子表列缓存按需补拉（detailColumns 依赖 columnCache）
  void (async () => {})()

  const isNarrowSource = (row) => {
    if (!row.detailBindingId) return false
    const d = details.value.find((x) => x.ID === row.detailBindingId)
    return d && d.detailKind === 2
  }
  const sourceColumns = (row) => {
    let tableId = form.value.tableId
    if (row.detailBindingId) {
      const d = details.value.find((x) => x.ID === row.detailBindingId)
      if (d) tableId = d.tableId
    }
    let cols = columnCache.value[tableId] || []
    if (row.bindingType === 1 && row.xsdType && COMPAT[row.xsdType]) {
      const ok = COMPAT[row.xsdType]
      cols = cols.filter((c) => ok.includes(c.dataType))
    }
    return cols
  }

  // 按类属性重建绑定行（保留已有配置）
  const autoFillProperties = () => {
    const src = classPropSource.value
    if (!src) return
    const existing = new Map(properties.value.map((p) => [p.propertyId, p]))
    const rows = []
    for (const p of src.datatypeProperties || []) {
      const old = existing.get(p.ID)
      rows.push({
        propertyId: p.ID,
        bindingType: 1,
        propertyLabel: p.label || p.localName,
        xsdType: p.typeOrRange,
        detailBindingId: old ? old.detailBindingId : null,
        bizColumn: old ? old.bizColumn : '',
        enabledBool: old ? old.enabledBool : true
      })
    }
    for (const p of src.objectProperties || []) {
      const old = existing.get(p.ID)
      rows.push({
        propertyId: p.ID,
        bindingType: 2,
        propertyLabel: p.label || p.localName,
        xsdType: '',
        detailBindingId: null,
        bizColumn: old ? old.bizColumn : '',
        enabledBool: old ? old.enabledBool : true
      })
    }
    properties.value = rows
  }

  const save = async () => {
    if (!form.value.classId || !form.value.tableId) {
      ElMessage.warning('请先选择本体项目、类与业务主表')
      return
    }
    saving.value = true
    try {
      const req = {
        ...form.value,
        projectId: form.value.projectId,
        details: details.value.filter((d) => d.tableId > 0).map((d) => ({ ...d, enabledBool: undefined })),
        properties: properties.value
          .filter((p) => p.bizColumn && p.bizColumn.trim())
          .map((p) => ({
            propertyId: p.propertyId,
            bindingType: p.bindingType,
            detailBindingId: p.detailBindingId || null,
            bizColumn: p.bizColumn,
            enabled: p.enabledBool ? 1 : 0
          }))
      }
      let res
      if (form.value.ID) {
        res = await updateExtBinding(req)
        savedId.value = form.value.ID
      } else {
        res = await createExtBinding(req)
        if (res.code === 0) savedId.value = res.data.ID
      }
      if (res.code === 0) {
        ElMessage.success('绑定已保存（草稿态，生效请在列表页操作）')
        form.value.ID = savedId.value
        emit('saved')
      }
    } finally {
      saving.value = false
    }
  }
</script>
