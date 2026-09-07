<template>
  <div class="max-w-4xl">
    <el-alert type="info" :closable="false" class="mb-3"
      title="转换按排序串行执行；JS 脚本输入 data（行数组），须返回新数组；字典映射未命中的值保持原样" />
    <div class="mb-2">
      <el-button type="primary" icon="plus" @click="addRow">添加转换</el-button>
    </div>
    <el-empty v-if="!list.length" description="暂无数据转换（结果将原样输出）" />

    <el-card v-for="(row, idx) in list" :key="idx" class="mb-3" shadow="never">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="font-bold">转换 {{ idx + 1 }}</span>
          <div class="flex items-center gap-3">
            <el-select v-model="row.transformType" style="width: 140px" placeholder="转换类型">
              <el-option v-for="o in TRANSFORM_TYPE_OPTIONS" :key="o.value" :label="o.label" :value="o.value" />
            </el-select>
            <span class="text-xs text-gray-400">排序</span>
            <el-input-number v-model="row.orderNum" :min="0" :step="1" step-strictly controls-position="right" style="width: 100px" />
            <el-button icon="delete" type="danger" link @click="list.splice(idx, 1)">删除</el-button>
          </div>
        </div>
      </template>
      <el-input
        v-model="row.transformScript"
        type="textarea"
        :rows="6"
        class="font-mono"
        :placeholder="placeholderFor(row.transformType)"
      />
      <div v-if="row.transformType === 'dict'" class="mt-1">
        <el-tag v-if="dictValid(row.transformScript)" type="success" size="small">JSON 格式正确</el-tag>
        <el-tag v-else type="danger" size="small">JSON 格式错误</el-tag>
      </div>
    </el-card>
  </div>
</template>

<script setup>
  import { defineExpose } from 'vue'
  import { TRANSFORM_TYPE_OPTIONS } from '@/api/report/dataSet'

  const props = defineProps({
    transforms: { type: Array, default: () => [] }
  })

  const list = props.transforms

  const addRow = () => {
    list.push({ transformType: 'js', transformScript: '', orderNum: list.length + 1 })
  }

  const placeholderFor = (type) =>
    type === 'dict'
      ? '{"field":"status","mapping":{"0":"停用","1":"启用"}}'
      : '// 输入 data 为行数组，返回处理后的新数组\nreturn data.map(r => ({ ...r, qty: Number(r.qty) * 10 }))'

  const dictValid = (script) => {
    if (!script || !script.trim()) return false
    try {
      const v = JSON.parse(script)
      return typeof v === 'object' && v !== null
    } catch (e) {
      return false
    }
  }

  const validate = () => true

  const collect = () =>
    list.map((r, i) => ({
      transformType: r.transformType || 'js',
      transformScript: r.transformScript || '',
      orderNum: r.orderNum ?? i + 1
    }))

  defineExpose({ validate, collect, addRow })
</script>

<style lang="scss" scoped>
  .font-mono :deep(textarea) {
    font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
    font-size: 12px;
  }
</style>
