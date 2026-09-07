<template>
  <div class="param-bar" v-if="params.length">
    <div class="flex items-center gap-2 cursor-pointer py-1" @click="expanded = !expanded">
      <el-icon><component :is="expanded ? 'ArrowDown' : 'ArrowRight'" /></el-icon>
      <span class="text-sm font-bold">查询参数</span>
      <el-button size="small" type="primary" icon="refresh" link @click.stop="emit('reload')">刷新数据</el-button>
    </div>
    <div v-show="expanded" class="pb-1">
      <param-form ref="paramFormRef" :params="params" />
    </div>
  </div>
</template>

<script setup>
  // 参数栏（可折叠）：复用 04 的公共参数表单组件；参数变化由父组件重新取数
  import { ref } from 'vue'
  import ParamForm from '@/view/report/components/paramForm.vue'

  const props = defineProps({
    params: { type: Array, default: () => [] }
  })
  const emit = defineEmits(['reload'])

  const expanded = ref((props.params || []).length > 0)
  const paramFormRef = ref(null)

  const buildParamValues = () => paramFormRef.value?.buildParamValues() || {}
  const resetValues = () => paramFormRef.value?.resetValues()

  defineExpose({ buildParamValues, resetValues })
</script>

<style lang="scss" scoped>
  .param-bar {
    border-bottom: 1px solid var(--el-border-color-lighter);
    padding: 0 4px;
  }
</style>
