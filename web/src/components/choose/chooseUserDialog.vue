<template>
  <el-dialog
    :model-value="visible"
    title="选择用户"
    width="860px"
    append-to-body
    destroy-on-close
    :close-on-click-modal="false"
    @update:model-value="(v) => emit('update:visible', v)"
  >
    <div class="flex gap-3">
      <div class="w-52 shrink-0">
        <div class="mb-2 text-sm text-gray-500">按部门过滤（任意层级，含子部门）</div>
        <el-tree-select
          v-model="deptId"
          :data="state.deptTree"
          node-key="ID"
          :props="{ label: 'name', children: 'children' }"
          check-strictly
          :render-after-expand="false"
          default-expand-all
          filterable
          clearable
          style="width: 100%"
          placeholder="全部部门"
        />
      </div>
      <div class="flex-1 min-w-0">
        <div class="mb-2 flex gap-2">
          <el-input v-model="keyword" placeholder="搜索昵称/用户名" clearable class="flex-1" />
          <el-select
            v-model="postFilterIds"
            multiple
            collapse-tags
            collapse-tags-tooltip
            clearable
            placeholder="全部岗位"
            style="width: 190px"
          >
            <el-option v-for="p in state.posts" :key="p.ID" :label="p.postName" :value="p.ID" />
          </el-select>
        </div>
        <el-table
          ref="tableRef"
          :data="filteredUsers"
          v-loading="loading"
          row-key="ID"
          max-height="400"
          :highlight-current-row="!multiple"
          @selection-change="handleSelectionChange"
          @current-change="handleCurrentChange"
        >
          <el-table-column v-if="multiple" type="selection" width="45" />
          <el-table-column prop="nickName" label="昵称" min-width="110" show-overflow-tooltip />
          <el-table-column prop="userName" label="用户名" min-width="110" show-overflow-tooltip />
          <el-table-column label="部门" min-width="120" show-overflow-tooltip>
            <template #default="{ row }">{{ row.departmentName || '未分配' }}</template>
          </el-table-column>
          <el-table-column label="岗位" min-width="150">
            <template #default="{ row }">
              <el-tag
                v-for="name in resolvePostLabels(row.postIds)"
                :key="name"
                size="small"
                class="mr-1 mb-0.5"
              >
                {{ name }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
    <div class="mt-3">
      <span class="text-sm text-gray-500">已选（{{ selectedRows.length }}）：</span>
      <el-tag v-for="u in selectedRows" :key="u.ID" class="mr-2 mb-1" closable @close="removeSelected(u.ID)">
        {{ u.nickName || u.userName }}
      </el-tag>
    </div>
    <template #footer>
      <el-button @click="close">取 消</el-button>
      <el-button type="primary" @click="confirm">确 定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
  import { computed, nextTick, ref, watch } from 'vue'
  import { ensurePosts, ensureUsers, filterUsers, resolvePostLabels, state } from './cache'

  defineOptions({
    name: 'GvaChooseUserDialog'
  })

  const props = defineProps({
    visible: { type: Boolean, default: false },
    // 单选 number / 多选 number[]
    modelValue: { type: [Number, Array], default: undefined },
    multiple: { type: Boolean, default: true },
    // { departmentId, postId } 作为打开时的初始过滤条件
    params: { type: Object, default: () => ({}) }
  })

  const emit = defineEmits(['update:visible', 'update:modelValue', 'change'])

  const tableRef = ref(null)
  const loading = ref(false)
  const deptId = ref(undefined) // undefined = 全部部门
  const keyword = ref('')
  const postFilterIds = ref([])
  // 唯一事实源：过滤条件切换/清空不丢已选，已选标签可单独移除
  const selectedMap = ref({})
  // 过滤导致表格数据更换时，el-table 会清空勾选并触发事件——置位期间跳过同步，再从 selectedMap 回填
  const syncing = ref(false)

  const filteredUsers = computed(() =>
    filterUsers({
      keyword: keyword.value,
      departmentId: deptId.value,
      postIds: postFilterIds.value
    })
  )

  const selectedRows = computed(() => Object.values(selectedMap.value))

  const open = async () => {
    const ids = toIds(props.modelValue)
    keyword.value = ''
    deptId.value = (props.params && props.params.departmentId) || undefined
    postFilterIds.value = props.params && props.params.postId ? [props.params.postId] : []
    selectedMap.value = {}
    loading.value = true
    try {
      await ensureUsers(true) // 打开强刷：人员调动尽快可见
      ensurePosts()
    } finally {
      loading.value = false
    }
    ids.forEach((id) => {
      const u = state.userMap[id]
      if (u) selectedMap.value[id] = u
    })
    applySelection()
  }

  watch(
    () => props.visible,
    (v) => {
      if (v) open()
    }
  )

  // 数据变化（三维过滤/缓存到位）→ 同步表格勾选态
  watch(filteredUsers, () => applySelection())

  const applySelection = async () => {
    syncing.value = true
    await nextTick()
    const table = tableRef.value
    if (table) {
      if (props.multiple) {
        filteredUsers.value.forEach((u) => {
          if (selectedMap.value[u.ID]) table.toggleRowSelection(u, true)
        })
      } else {
        const hit = filteredUsers.value.find((u) => selectedMap.value[u.ID])
        if (hit) {
          table.setCurrentRow(hit)
        } else {
          table.setCurrentRow()
        }
      }
    }
    await nextTick()
    syncing.value = false
  }

  const handleSelectionChange = (rows) => {
    if (syncing.value) return
    const checkedIds = new Set(rows.map((r) => r.ID))
    filteredUsers.value.forEach((u) => {
      if (checkedIds.has(u.ID)) {
        selectedMap.value[u.ID] = u
      } else {
        delete selectedMap.value[u.ID]
      }
    })
  }

  const handleCurrentChange = (row) => {
    if (syncing.value) return
    selectedMap.value = {}
    if (row) selectedMap.value[row.ID] = row
  }

  const removeSelected = (id) => {
    delete selectedMap.value[id]
    const table = tableRef.value
    if (!table) return
    const row = filteredUsers.value.find((u) => Number(u.ID) === Number(id))
    if (row) {
      if (props.multiple) {
        table.toggleRowSelection(row, false)
      } else {
        table.setCurrentRow()
      }
    }
  }

  // 确定：value 即最终态；取消不改动 modelValue
  const confirm = () => {
    const rows = Object.values(selectedMap.value)
    const value = props.multiple ? rows.map((r) => r.ID) : rows.length ? rows[0].ID : undefined
    emit('update:modelValue', value)
    emit('change', value, rows)
    close()
  }

  const close = () => emit('update:visible', false)

  function toIds(value) {
    if (Array.isArray(value)) return value
    if (value === null || value === undefined || value === '') return []
    return [value]
  }
</script>
