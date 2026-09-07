<template>
  <div class="field-panel">
    <div class="mb-2 text-sm">
      <div class="font-bold truncate" :title="dataSetName">{{ dataSetName || '未关联数据集' }}</div>
      <div class="text-xs text-gray-400 truncate">{{ setCode }}</div>
    </div>
    <el-empty v-if="!fieldMeta.length" description="暂无字段" :image-size="60" />
    <template v-else>
      <div class="text-xs text-gray-400 mb-1 px-1">维度（string / date）</div>
      <div
        v-for="f in dimensionFields"
        :key="f.name"
        class="field-row"
        draggable="true"
        @dragstart="onDragStart($event, f.name)"
      >
        <span class="flex-1 truncate" :title="f.name">{{ f.name }}</span>
        <span v-if="usedIn(f.name)" class="text-xs text-gray-300" :title="usedIn(f.name)">已用于 {{ usedIn(f.name) }}</span>
        <template v-else>
          <el-button size="small" link type="primary" @click="addToRows(f.name)">+行头</el-button>
          <el-button size="small" link type="primary" @click="addToColumns(f.name)">+列头</el-button>
        </template>
      </div>

      <div class="text-xs text-gray-400 mt-3 mb-1 px-1">度量（number）<span class="ml-1 text-gray-300">SQL 已预聚合字段请选 NONE</span></div>
      <div
        v-for="f in measureFields"
        :key="f.name"
        class="field-row"
        draggable="true"
        @dragstart="onDragStart($event, f.name)"
      >
        <span class="flex-1 truncate" :title="f.name">{{ f.name }}</span>
        <span v-if="usedIn(f.name)" class="text-xs text-gray-300">已用于 {{ usedIn(f.name) }}</span>
        <el-button v-else size="small" link type="primary" @click="addToValues(f.name)">+数值</el-button>
      </div>
    </template>
  </div>
</template>

<script setup>
  // 左栏：维度/度量分区（点击添加为主、HTML5 拖拽为辅；已在某区域置灰并标注位置）
  import { computed } from 'vue'

  const props = defineProps({
    fieldMeta: { type: Array, default: () => [] }, // [{name, type}]
    config: { type: Object, required: true },
    setCode: { type: String, default: '' },
    dataSetName: { type: String, default: '' }
  })
  const emit = defineEmits(['add-row', 'add-column', 'add-value'])

  const dimensionFields = computed(() => props.fieldMeta.filter((f) => f.type === 'string' || f.type === 'date'))
  const measureFields = computed(() => props.fieldMeta.filter((f) => f.type === 'number'))

  const usedIn = (name) => {
    const { rows = [], columns = [], values = [] } = props.config.fields
    if (rows.includes(name)) return '行头'
    if (columns.includes(name)) return '列头'
    if (values.some((v) => v.field === name)) return '数值'
    return ''
  }

  const addToRows = (name) => emit('add-row', name)
  const addToColumns = (name) => emit('add-column', name)
  const addToValues = (name) => emit('add-value', name)

  const onDragStart = (event, field) => {
    event.dataTransfer.setData('application/json', JSON.stringify({ field }))
    event.dataTransfer.effectAllowed = 'copy'
  }
</script>

<style lang="scss" scoped>
  .field-panel {
    height: 100%;
    overflow: auto;
    padding-right: 4px;

    .field-row {
      display: flex;
      align-items: center;
      gap: 2px;
      padding: 3px 6px;
      margin: 2px 0;
      border: 1px solid var(--el-border-color-lighter);
      border-radius: 4px;
      font-size: 13px;
      cursor: grab;
      background: var(--el-bg-color);

      &:hover {
        border-color: var(--el-color-primary-light-5);
        background: var(--el-color-primary-light-9);
      }
    }
  }
</style>
