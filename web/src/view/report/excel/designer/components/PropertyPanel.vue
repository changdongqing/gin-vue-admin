<template>
  <div class="property-panel">
    <el-empty v-if="!selectedCell" description="点击单元格查看属性" :image-size="60" />
    <template v-else>
      <el-descriptions :column="1" border size="small">
        <el-descriptions-item label="Sheet">{{ selectedCell.sheetName }}</el-descriptions-item>
        <el-descriptions-item label="单元格">
          {{ a1Notation(selectedCell.row, selectedCell.col) }}（行 {{ selectedCell.row }} / 列 {{ selectedCell.col }}）
        </el-descriptions-item>
      </el-descriptions>

      <el-form label-width="60px" class="mt-3" @submit.prevent>
        <el-form-item label="值">
          <el-input
            v-model="editValue"
            type="textarea"
            :rows="3"
            @change="(v) => emit('update-value', v)"
          />
        </el-form-item>
      </el-form>

      <el-descriptions v-if="placeholder" :column="1" border size="small" title="数据绑定占位符">
        <el-descriptions-item label="完整占位">
          <el-tag type="success" size="small">是</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="数据集">{{ placeholder.set }}</el-descriptions-item>
        <el-descriptions-item label="字段">{{ placeholder.field }}</el-descriptions-item>
      </el-descriptions>
      <el-alert v-else-if="(editValue || '').includes('#{')" type="warning" :closable="false"
        title="包含片段占位符（渲染时文本内替换）" :style="{ fontSize: '12px' }" />
    </template>
  </div>
</template>

<script setup>
  // 右栏：单元格属性 / 占位符识别（宽匹配） / 值编辑回写
  import { ref, watch, computed } from 'vue'

  const props = defineProps({
    selectedCell: { type: Object, default: null }
  })
  const emit = defineEmits(['update-value'])

  const editValue = ref('')

  watch(
    () => props.selectedCell,
    (cell) => {
      editValue.value = cell ? String(cell.value ?? '') : ''
    },
    { immediate: true }
  )

  // 完整占位符：^#\{([^{}]+)\.([^{}]+)}$（字段名宽匹配，支持中文/连字符）
  const fullPlaceholderRe = /^#\{([^{}]+)\.([^{}]+)}$/

  const placeholder = computed(() => {
    const v = editValue.value || ''
    const m = v.match(fullPlaceholderRe)
    return m ? { set: m[1], field: m[2] } : null
  })

  const a1Notation = (row, col) => {
    let s = ''
    let n = col + 1
    while (n > 0) {
      const rem = (n - 1) % 26
      s = String.fromCharCode(65 + rem) + s
      n = Math.floor((n - 1) / 26)
    }
    return `${s}${row + 1}`
  }
</script>

<style lang="scss" scoped>
  .property-panel {
    height: 100%;
    overflow: auto;
    padding-right: 4px;
  }
</style>
