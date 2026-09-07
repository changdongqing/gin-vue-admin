<template>
  <div class="dataset-panel">
    <div class="mb-2">
      <el-button type="primary" size="small" icon="link" class="w-full" @click="emit('bind')">关联数据集</el-button>
    </div>
    <el-empty v-if="!dataSetFields.length" description="尚未关联数据集" :image-size="60" />

    <el-collapse v-model="expanded">
      <el-collapse-item v-for="ds in dataSetFields" :key="ds.setCode" :name="ds.setCode">
        <template #title>
          <span class="font-bold">{{ ds.setName }}</span>
          <span class="ml-1 text-xs text-gray-400">（{{ ds.setCode }}）</span>
        </template>
        <el-alert
          v-if="!ds.fields.length"
          type="warning"
          :closable="false"
          title="无字段：请先在数据集管理准备结果案例"
          :style="{ fontSize: '12px' }"
        />
        <div
          v-for="field in ds.fields"
          :key="field"
          class="field-row"
          draggable="true"
          @dragstart="onDragStart($event, ds.setCode, field)"
        >
          <span class="truncate flex-1" :title="field">{{ field }}</span>
          <el-button type="primary" link size="small" icon="plus" @click.stop="emit('insert', ds.setCode, field)" />
        </div>
      </el-collapse-item>
    </el-collapse>
  </div>
</template>

<script setup>
  // 左栏：关联数据集折叠面板；字段行可拖拽（dataTransfer 写入 JSON）或点「+」插入
  import { ref } from 'vue'

  const props = defineProps({
    dataSetFields: { type: Array, default: () => [] }
  })
  const emit = defineEmits(['bind', 'insert'])

  const expanded = ref([])

  const onDragStart = (event, setCode, field) => {
    event.dataTransfer.setData('application/json', JSON.stringify({ setCode, fieldName: field }))
    event.dataTransfer.effectAllowed = 'copy'
  }
</script>

<style lang="scss" scoped>
  .dataset-panel {
    height: 100%;
    overflow: auto;
    padding-right: 4px;

    .field-row {
      display: flex;
      align-items: center;
      gap: 4px;
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
