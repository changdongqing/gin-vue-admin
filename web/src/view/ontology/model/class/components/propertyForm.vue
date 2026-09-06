<template>
  <el-dialog v-model="visible" :title="dialogTitle" width="560px" :close-on-click-modal="false">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
      <el-form-item label="本地名" prop="localName">
        <el-input v-model="form.localName" :disabled="mode === 'edit'" placeholder="如 ratedPower" />
      </el-form-item>
      <el-form-item label="显示名" prop="label">
        <el-input v-model="form.label" placeholder="如 额定功率" />
      </el-form-item>

      <!-- 数据属性专属 -->
      <template v-if="kind === 'datatype'">
        <el-form-item label="XSD 类型" prop="xsdType">
          <el-select v-model="form.xsdType" :disabled="mode === 'edit'" placeholder="请选择类型">
            <el-option v-for="item in xsdOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="isNumericXsd" label="单位" prop="unitRef">
          <el-select v-model="form.unitRef" filterable allow-create clearable placeholder="选择或输入 QUDT 单位 IRI">
            <el-option v-for="item in unitOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="枚举值" prop="enumText">
          <el-input v-model="form.enumText" placeholder="逗号分隔，如 低压,中压,高压" />
        </el-form-item>
        <el-form-item label="标识符" prop="isIdentifier">
          <el-switch v-model="form.isIdentifier" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </template>

      <!-- 对象属性专属 -->
      <template v-else>
        <el-form-item label="Range 类" prop="rangeClassId">
          <el-select v-model="form.rangeClassId" clearable filterable placeholder="可空，选择目标类">
            <el-option
              v-for="item in classList"
              :key="item.ID"
              :label="item.localName + (item.labelCn ? `（${item.labelCn}）` : '')"
              :value="item.ID"
            />
          </el-select>
        </el-form-item>
      </template>

      <el-form-item label="最小基数" prop="minCardinality">
        <el-input-number v-model="form.minCardinality" :min="0" />
      </el-form-item>
      <el-form-item label="最大基数" prop="maxCardinality">
        <el-input-number v-model="form.maxCardinality" :min="-1" />
        <span class="ml-2 text-xs text-gray-400">-1 = 无限制（n）</span>
      </el-form-item>
      <el-form-item label="排序" prop="sortOrder">
        <el-input-number v-model="form.sortOrder" :min="0" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button v-if="kind === 'object' && mode === 'edit'" type="warning" plain @click="onSuggestInverse">建议反向</el-button>
      <el-button @click="visible = false">取 消</el-button>
      <el-button type="primary" @click="submit">确 定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
  import {
    findModelDatatypeProperty,
    createModelDatatypeProperty,
    updateModelDatatypeProperty,
    findModelObjectProperty,
    createModelObjectProperty,
    updateModelObjectProperty,
    suggestInverseModelObjectProperty
  } from '@/api/ontology/modelProperty'
  import { getUnitsForSupply } from '@/api/ontology/unit'
  import { ref, computed, reactive } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'

  defineOptions({ name: 'PropertyForm' })

  const emit = defineEmits(['success'])

  const props = defineProps({
    classList: { type: Array, default: () => [] } // 同项目类清单（对象属性 Range 下拉）
  })

  const visible = ref(false)
  const mode = ref('add')
  const kind = ref('datatype')
  const formRef = ref(null)
  const unitOptions = ref([])
  const form = ref({})
  const xsdOptions = [
    { label: '字符串', value: 'xsd:string' },
    { label: '整数', value: 'xsd:integer' },
    { label: '小数', value: 'xsd:decimal' },
    { label: '布尔', value: 'xsd:boolean' },
    { label: '日期时间', value: 'xsd:datetime' }
  ]
  const rules = ref({
    localName: [
      { required: true, message: '请输入本地名', trigger: 'blur' },
      { pattern: /^[A-Za-z][A-Za-z0-9_-]*$/, message: '须以字母开头，仅含字母/数字/下划线/连字符', trigger: 'blur' }
    ],
    xsdType: [{ required: true, message: '请选择XSD类型', trigger: 'change' }]
  })

  let ctx = { projectId: 0, classId: 0, propertyId: 0 }

  const isNumericXsd = computed(() => form.value.xsdType === 'xsd:integer' || form.value.xsdType === 'xsd:decimal')
  const dialogTitle = computed(() => {
    const action = mode.value === 'add' ? '新增' : '编辑'
    return `${action}${kind.value === 'object' ? '对象属性' : '数据属性'}`
  })

  const initForm = () => ({
    ID: 0,
    localName: '',
    label: '',
    xsdType: 'xsd:string',
    unitRef: '',
    enumText: '',
    rangeClassId: null,
    minCardinality: 0,
    maxCardinality: -1,
    isIdentifier: 0,
    sortOrder: 0
  })

  // JSON 数组字符串 → 逗号分隔（编辑回填）
  const jsonToText = (jsonStr) => {
    if (!jsonStr) return ''
    try {
      const arr = JSON.parse(jsonStr)
      return Array.isArray(arr) ? arr.join(',') : jsonStr
    } catch (e) {
      return jsonStr
    }
  }

  const open = async (projectId, classId, propertyKind, propertyId = 0) => {
    ctx = { projectId, classId, propertyId }
    kind.value = propertyKind
    mode.value = propertyId ? 'edit' : 'add'
    formRef.value && formRef.value.resetFields()
    form.value = initForm()
    visible.value = true
    if (propertyKind === 'datatype') {
      getUnitsForSupply().then((res) => {
        if (res.code === 0) {
          unitOptions.value = (res.data || []).map((u) => ({ label: u.label || u.symbol || u.unitRef, value: u.unitRef }))
        }
      })
    }
    if (mode.value === 'edit') {
      if (propertyKind === 'datatype') {
        const res = await findModelDatatypeProperty({ ID: propertyId })
        if (res.code === 0) {
          form.value = { ...form.value, ...res.data, enumText: jsonToText(res.data.enumValues) }
        }
      } else {
        const res = await findModelObjectProperty({ ID: propertyId })
        if (res.code === 0) {
          form.value = { ...form.value, ...res.data }
        }
      }
      formRef.value && formRef.value.clearValidate()
    }
  }
  defineExpose({ open })

  const submit = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      let res
      if (kind.value === 'datatype') {
        const req = {
          ...form.value,
          projectId: ctx.projectId,
          classId: ctx.classId,
          enumValues: toEnumJson(form.value.enumText)
        }
        delete req.enumText
        res = mode.value === 'add' ? await createModelDatatypeProperty(req) : await updateModelDatatypeProperty(req)
      } else {
        const req = { ...form.value, projectId: ctx.projectId, domainClassId: ctx.classId }
        res = mode.value === 'add' ? await createModelObjectProperty(req) : await updateModelObjectProperty(req)
      }
      if (res.code === 0) {
        ElMessage.success(mode.value === 'add' ? '创建成功' : '更新成功')
        visible.value = false
        emit('success')
      }
    })
  }

  // 逗号分隔 → JSON 数组字符串（AC-12.3：a,b,c → ["a","b","c"]）
  const toEnumJson = (text) => {
    const t = (text || '').trim()
    if (!t) return ''
    return JSON.stringify(t.split(',').map((s) => s.trim()).filter(Boolean))
  }

  // 建议反向：取建议 → 确认 → 创建反向属性并双向 inverseOf 互指
  const onSuggestInverse = async () => {
    const res = await suggestInverseModelObjectProperty({ ID: ctx.propertyId })
    if (res.code !== 0) return
    const s = res.data
    try {
      await ElMessageBox.confirm(
        `将创建反向属性「${s.suggestedLocalName}（${s.suggestedLabel}）」，domain/range 对调并与当前属性 inverseOf 互指，是否继续?`,
        '建议反向',
        { confirmButtonText: '创建', cancelButtonText: '取消', type: 'info' }
      )
    } catch (e) {
      return
    }
    const createRes = await createModelObjectProperty({
      projectId: ctx.projectId,
      domainClassId: s.suggestedDomainClassId,
      rangeClassId: s.suggestedRangeClassId || null,
      localName: s.suggestedLocalName,
      label: s.suggestedLabel,
      inverseOf: ctx.propertyId,
      minCardinality: 0,
      maxCardinality: -1
    })
    if (createRes.code === 0) {
      ElMessage.success('反向属性已创建并互指')
      visible.value = false
      emit('success')
    }
  }
</script>
