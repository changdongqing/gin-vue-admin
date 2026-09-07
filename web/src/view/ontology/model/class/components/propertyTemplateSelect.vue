<template>
  <el-dialog v-model="visible" :title="`从属性库选择（${kind === 'object' ? '对象属性' : '数据属性'}）`" width="760px" :close-on-click-modal="false">
    <el-input
      v-model="keyword"
      placeholder="搜索模板编码 / 显示名 / 别名"
      clearable
      class="mb-2"
      :prefix-icon="'Search'"
    />
    <el-table ref="tableRef" :data="pagedList" border max-height="360" row-key="templateCode" @selection-change="onSelectionChange">
      <el-table-column type="selection" width="42" :selectable="(row) => !attachedSet[row.templateCode]" />
      <el-table-column prop="templateCode" label="模板编码" width="190" show-overflow-tooltip />
      <el-table-column prop="label" label="显示名" width="130" />
      <el-table-column prop="category" label="分类" width="100" />
      <el-table-column label="类型/基数" width="110">
        <template #default="{ row }">
          {{ kind === 'object' ? row.defaultCardinality || '-' : row.type || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="unitRef" label="单位" min-width="140" show-overflow-tooltip />
      <el-table-column prop="values" label="枚举值" min-width="120" show-overflow-tooltip />
    </el-table>
    <div class="flex justify-between items-center mt-2">
      <span class="text-xs text-gray-400">已挂载的属性（灰显）不可重复选择</span>
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="filteredList.length"
        :page-sizes="[8, 15, 30]"
        layout="total, sizes, prev, pager, next"
        small
      />
    </div>
    <template #footer>
      <el-button @click="visible = false">取 消</el-button>
      <el-button type="primary" :loading="submitting" :disabled="!selection.length" @click="submit">
        挂载选中（{{ selection.length }}）
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
  import { getPropertyTemplatesForSupply } from '@/api/ontology/propertyTemplate'
  import { getModelClassDetail } from '@/api/ontology/modelClass'
  import { instantiateModelDatatypeProperty, instantiateModelObjectProperty } from '@/api/ontology/modelProperty'
  import { ref, computed } from 'vue'
  import { ElMessage } from 'element-plus'

  defineOptions({ name: 'PropertyTemplateSelect' })

  const emit = defineEmits(['success'])

  const visible = ref(false)
  const submitting = ref(false)
  const kind = ref('datatype')
  const keyword = ref('')
  const page = ref(1)
  const pageSize = ref(8)
  const tableRef = ref(null)
  const list = ref([])
  const selection = ref([])
  const attachedSet = ref({})
  let ctx = { projectId: 0, classId: 0 }

  const filteredList = computed(() => {
    const kw = keyword.value.trim().toLowerCase()
    if (!kw) return list.value
    return list.value.filter(
      (item) =>
        (item.templateCode || '').toLowerCase().includes(kw) ||
        (item.label || '').toLowerCase().includes(kw) ||
        (item.alias || '').toLowerCase().includes(kw)
    )
  })
  const pagedList = computed(() => filteredList.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value))

  // 已挂载属性编码集合（同 KIND 两表均查，灰显禁选）
  const loadAttached = async () => {
    const res = await getModelClassDetail({ ID: ctx.classId })
    if (res.code === 0) {
      const set = {}
      for (const p of res.data.datatypeProperties || []) set[p.localName] = true
      for (const p of res.data.objectProperties || []) set[p.localName] = true
      attachedSet.value = set
    }
  }

  const open = async (projectId, classId, propertyKind) => {
    ctx = { projectId, classId }
    kind.value = propertyKind
    keyword.value = ''
    page.value = 1
    selection.value = []
    visible.value = true
    loadAttached()
    const res = await getPropertyTemplatesForSupply({ kind: propertyKind, includeDeprecated: false })
    if (res.code === 0) {
      list.value = res.data || []
    }
  }
  defineExpose({ open })

  const onSelectionChange = (rows) => {
    selection.value = rows
  }

  const submit = async () => {
    submitting.value = true
    let success = 0
    const failures = []
    try {
      for (const row of selection.value) {
        const fn =
          kind.value === 'object' ? instantiateModelObjectProperty : instantiateModelDatatypeProperty
        const res = await fn({ projectId: ctx.projectId, classId: ctx.classId, templateCode: row.templateCode })
        if (res.code === 0) {
          success++
        } else {
          failures.push(`${row.templateCode}（${res.msg || '失败'}）`)
        }
      }
      if (failures.length) {
        ElMessage.warning(`成功挂载 ${success} 条属性；失败 ${failures.length} 条：${failures.join('、')}`)
      } else {
        ElMessage.success(`成功挂载 ${success} 条属性`)
      }
      visible.value = false
      emit('success')
    } finally {
      submitting.value = false
    }
  }
</script>
