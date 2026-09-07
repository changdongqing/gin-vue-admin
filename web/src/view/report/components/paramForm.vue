<template>
  <el-form inline class="param-form">
    <el-form-item v-for="p in params" :key="p.paramName" :label="paramLabel(p)">
      <!-- 日期范围 -->
      <el-date-picker
        v-if="p.paramType === 'dateRange'"
        v-model="innerValues[p.paramName + '__range']"
        type="daterange"
        value-format="YYYY-MM-DD"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        clearable
        @change="emitUpdate"
      />
      <!-- 日期 -->
      <el-date-picker
        v-else-if="p.paramType === 'date'"
        v-model="innerValues[p.paramName]"
        type="date"
        :format="dateFormatOf(p) || 'YYYY-MM-DD'"
        :value-format="dateFormatOf(p) || 'YYYY-MM-DD'"
        clearable
        @change="emitUpdate"
      />
      <!-- 日期时间 -->
      <el-date-picker
        v-else-if="p.paramType === 'datetime'"
        v-model="innerValues[p.paramName]"
        type="datetime"
        :format="dateFormatOf(p) || 'YYYY-MM-DD HH:mm:ss'"
        :value-format="dateFormatOf(p) || 'YYYY-MM-DD HH:mm:ss'"
        clearable
        @change="emitUpdate"
      />
      <!-- 下拉多选 -->
      <el-select
        v-else-if="p.paramType === 'multipleSelect'"
        v-model="innerValues[p.paramName]"
        multiple
        clearable
        collapse-tags
        :placeholder="`请选择${p.paramDesc || p.paramName}`"
        style="min-width: 180px"
        @change="emitUpdate"
      >
        <el-option v-for="o in optionsOf(p)" :key="String(o.value)" :label="o.label" :value="o.value" />
      </el-select>
      <!-- 下拉单选 -->
      <el-select
        v-else-if="p.paramType === 'select'"
        v-model="innerValues[p.paramName]"
        clearable
        :placeholder="`请选择${p.paramDesc || p.paramName}`"
        style="min-width: 160px"
        @change="emitUpdate"
      >
        <el-option v-for="o in optionsOf(p)" :key="String(o.value)" :label="o.label" :value="o.value" />
      </el-select>
      <!-- 数字 -->
      <el-input-number
        v-else-if="p.paramType === 'number'"
        v-model="innerValues[p.paramName]"
        :controls="false"
        style="width: 170px"
        @change="emitUpdate"
      />
      <!-- 文本（默认） -->
      <el-input v-else v-model="innerValues[p.paramName]" clearable style="width: 190px" @change="emitUpdate" />
    </el-form-item>
  </el-form>
</template>

<script setup>
  // 参数表单公共组件（04 Excel 预览页 / 05 分析报表共用）：
  // 类型→组件映射、sampleItem 预填（日期类转 Date；默认值表达式不在前端解析，空值不提交交由后端）
  import { reactive, watch, defineExpose } from 'vue'
  import { getDict } from '@/utils/dictionary'

  const props = defineProps({
    params: { type: Array, default: () => [] }
  })
  const emit = defineEmits(['update'])

  const innerValues = reactive({})

  const paramLabel = (p) => {
    const desc = p.paramDesc || p.paramName
    return p.requiredFlag ? `${desc} *` : desc
  }

  const dateFormatOf = (p) => (p.dateFormat ? p.dateFormat : '')

  // 下拉参数选项：字典优先 → 自定义 JSON（label/value，兼容 text/name）→ 空
  const optionsCache = reactive({})
  const optionsOf = (p) => {
    if (optionsCache[p.paramName]) return optionsCache[p.paramName]
    if (p.customOptions) {
      try {
        const arr = JSON.parse(p.customOptions)
        if (Array.isArray(arr)) {
          optionsCache[p.paramName] = arr.map((o) => ({
            label: o.label ?? o.text ?? o.name ?? String(o.value ?? o),
            value: o.value ?? o.id ?? o
          }))
          return optionsCache[p.paramName]
        }
      } catch (e) {
        /* 忽略非法 JSON */
      }
    }
    return []
  }

  const loadDictOptions = async () => {
    for (const p of props.params || []) {
      if ((p.paramType === 'select' || p.paramType === 'multipleSelect') && p.dictType && !optionsCache[p.paramName]) {
        const list = await getDict(p.dictType)
        if (list?.length) {
          optionsCache[p.paramName] = list.map((d) => ({ label: d.label, value: d.value }))
        }
      }
    }
  }

  // sampleItem 预填：dateRange 拆两段；date/datetime 原样（YYYY-MM-DD 字符串可直接给 value-format 的 picker）
  const initFormValues = () => {
    Object.keys(innerValues).forEach((k) => delete innerValues[k])
    for (const p of props.params || []) {
      if (p.sampleItem === undefined || p.sampleItem === null || p.sampleItem === '') continue
      if (p.paramType === 'dateRange') {
        const [s, e] = String(p.sampleItem).split(',')
        innerValues[p.paramName + '__range'] = e !== undefined ? [s, e] : [s, s]
      } else if (p.paramType === 'multipleSelect') {
        innerValues[p.paramName] = String(p.sampleItem).split(',').filter((x) => x !== '')
      } else {
        innerValues[p.paramName] = p.sampleItem
      }
    }
    loadDictOptions()
  }

  watch(() => props.params, initFormValues, { immediate: true, deep: false })

  const emitUpdate = () => emit('update', innerValues)

  // buildParamValues：空值跳过（交给后端默认值）；dateRange 拼提交；multipleSelect join(',')
  const buildParamValues = () => {
    const out = {}
    for (const p of props.params || []) {
      if (p.paramType === 'dateRange') {
        const range = innerValues[p.paramName + '__range']
        if (Array.isArray(range) && range[0] && range[1]) {
          out[p.paramName] = `${range[0]},${range[1]}`
        }
        continue
      }
      const v = innerValues[p.paramName]
      if (v === undefined || v === null || v === '') continue
      if (p.paramType === 'multipleSelect' && Array.isArray(v)) {
        if (v.length) out[p.paramName] = v.join(',')
        continue
      }
      out[p.paramName] = v
    }
    return out
  }

  // 重置为示例值并清空已填
  const resetValues = () => {
    initFormValues()
    emit('update', innerValues)
  }

  defineExpose({ buildParamValues, resetValues })
</script>

<style scoped>
  .param-form :deep(.el-form-item) {
    margin-bottom: 8px;
  }
</style>
