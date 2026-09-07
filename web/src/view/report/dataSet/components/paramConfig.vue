<template>
  <div class="max-w-4xl">
    <el-alert type="info" :closable="false" class="mb-3"
      title="参数在查询语句中以 ${paramName} 引用；下拉类参数的字典/自定义选项编辑入口列后续迭代，当前可先维护示例值与默认值" />
    <div class="mb-2">
      <el-button type="primary" icon="plus" @click="addRow">添加参数</el-button>
    </div>
    <el-table :data="list" border size="small">
      <el-table-column type="index" label="序号" width="55" align="center" />
      <el-table-column label="参数名" width="180">
        <template #default="{ row }">
          <el-input v-model="row.paramName" placeholder="如 username" />
        </template>
      </el-table-column>
      <el-table-column label="描述" min-width="150">
        <template #default="{ row }">
          <el-input v-model="row.paramDesc" />
        </template>
      </el-table-column>
      <el-table-column label="类型" width="150">
        <template #default="{ row }">
          <el-select v-model="row.paramType">
            <el-option-group v-for="g in PARAM_TYPE_GROUPS" :key="g.label" :label="g.label">
              <el-option v-for="o in g.options" :key="o.value" :label="o.label" :value="o.value" />
            </el-option-group>
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="示例值" min-width="150">
        <template #default="{ row }">
          <el-input v-model="row.sampleItem" :placeholder="row.paramType === 'dateRange' ? '起,止' : ''" />
        </template>
      </el-table-column>
      <el-table-column label="默认值/表达式" min-width="170">
        <template #default="{ row }">
          <el-select v-model="row.defaultValue" filterable allow-create default-first-option clearable placeholder="表达式或固定值">
            <el-option v-for="o in DEFAULT_VALUE_EXPR_OPTIONS" :key="o.value" :label="o.label" :value="o.value" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="必填" width="70" align="center">
        <template #default="{ row }">
          <el-switch v-model="row.requiredFlag" />
        </template>
      </el-table-column>
      <el-table-column label="排序" width="120" align="center">
        <template #default="{ row }">
          <el-input-number v-model="row.orderNum" :min="0" :step="1" step-strictly controls-position="right" style="width: 90px" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="70" align="center">
        <template #default="{ $index }">
          <el-button icon="delete" type="primary" link @click="list.splice($index, 1)" />
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
  import { defineExpose } from 'vue'
  import { ElMessage } from 'element-plus'
  import { PARAM_TYPE_GROUPS, DEFAULT_VALUE_EXPR_OPTIONS } from '@/api/report/dataSet'

  const props = defineProps({
    params: { type: Array, default: () => [] }
  })

  // props.params 数组直接就地编辑（父组件提交时收集）
  const list = props.params

  const addRow = () => {
    list.push({
      paramName: '',
      paramDesc: '',
      paramType: 'string',
      sampleItem: '',
      defaultValue: '',
      requiredFlag: false,
      orderNum: list.length + 1
    })
  }

  const validate = () => {
    const seen = new Set()
    for (const row of list) {
      if (!row.paramName || !row.paramName.trim()) {
        ElMessage.error('参数名不能为空')
        return false
      }
      if (!/^\w+$/.test(row.paramName)) {
        ElMessage.error(`参数名「${row.paramName}」不合法（字母/数字/下划线）`)
        return false
      }
      if (seen.has(row.paramName)) {
        ElMessage.error(`参数名「${row.paramName}」重复`)
        return false
      }
      seen.add(row.paramName)
    }
    return true
  }

  const collect = () =>
    list.map((r, i) => ({
      paramName: r.paramName,
      paramDesc: r.paramDesc || '',
      paramType: r.paramType || 'string',
      sampleItem: r.sampleItem || '',
      defaultValue: r.defaultValue || '',
      requiredFlag: !!r.requiredFlag,
      orderNum: r.orderNum ?? i + 1
    }))

  defineExpose({ validate, collect, addRow })
</script>
