<template>
  <el-tree-select
    v-model="proxyValue"
    :data="treeData"
    node-key="ID"
    :props="treeProps"
    :multiple="multiple"
    :check-strictly="checkStrictly"
    filterable
    default-expand-all
    :render-after-expand="false"
    :disabled="disabled"
    :clearable="clearable"
    :placeholder="placeholder"
    :size="size"
    style="width: 100%"
    @change="onChange"
  />
</template>

<script setup>
  import { computed, onMounted } from 'vue'
  import { ensureDepts, state } from './cache'

  defineOptions({
    name: 'GvaChooseDept'
  })

  const props = defineProps({
    // 单选 number / 多选 number[]（切换 multiple 由调用方自行清值）
    modelValue: { type: [Number, Array], default: undefined },
    multiple: { type: Boolean, default: false },
    // 任意层级可选（默认）；false 收紧为仅叶子
    checkStrictly: { type: Boolean, default: true },
    // { companyId } 静态裁剪：缓存内过滤并保留祖先链
    params: { type: Object, default: () => ({}) },
    disabled: { type: Boolean, default: false },
    placeholder: { type: String, default: '请选择部门' },
    clearable: { type: Boolean, default: true },
    size: { type: String, default: undefined }
  })

  const emit = defineEmits(['update:modelValue', 'change'])

  const proxyValue = computed({
    get: () => props.modelValue,
    set: (v) => emit('update:modelValue', v)
  })

  // 停用部门（status=2）可见但不可选——保证历史值回显不丢
  const treeProps = {
    label: 'name',
    children: 'children',
    disabled: (data) => data.status === 2
  }

  const treeData = computed(() => {
    const companyId = props.params && props.params.companyId
    if (!companyId) return state.deptTree
    // 保留命中公司的部门及其祖先链
    const parentById = {}
    state.departments.forEach((d) => {
      parentById[d.ID] = d.parentId
    })
    const keep = new Set()
    state.departments.forEach((d) => {
      if (Number(d.companyId) === Number(companyId)) {
        let cur = d.ID
        while (cur && !keep.has(cur)) {
          keep.add(cur)
          cur = parentById[cur]
        }
      }
    })
    const prune = (nodes) =>
      (nodes || [])
        .filter((n) => keep.has(n.ID))
        .map((n) => ({ ...n, children: prune(n.children) }))
    return prune(state.deptTree)
  })

  const onChange = (value) => {
    const ids = toIds(value)
    const rows = ids.map((id) => state.deptMap[id]).filter(Boolean)
    emit('change', value, rows)
  }

  function toIds(value) {
    if (Array.isArray(value)) return value
    if (value === null || value === undefined || value === '') return []
    return [value]
  }

  onMounted(() => {
    ensureDepts()
  })

  defineExpose({
    refresh: () => ensureDepts(true)
  })
</script>
