<template>
  <el-drawer v-model="visible" size="80%" :show-close="false" @closed="onClosed">
    <template #header>
      <div class="flex justify-between items-center">
        <span class="text-lg">{{ type === 'add' ? '新增数据集' : '编辑数据集' }}</span>
        <div>
          <el-button @click="close">取 消</el-button>
          <el-button type="primary" @click="submit">保 存</el-button>
        </div>
      </div>
    </template>

    <el-tabs v-model="activeTab" type="border-card">
      <el-tab-pane label="基础信息" name="basic">
        <basic-info ref="basicRef" :form="form" :edit-mode="type === 'edit'" />
      </el-tab-pane>
      <el-tab-pane label="参数配置" name="params">
        <param-config ref="paramRef" :params="form.params" />
      </el-tab-pane>
      <el-tab-pane label="数据转换" name="transforms">
        <transform-config ref="transformRef" :transforms="form.transforms" />
      </el-tab-pane>
      <el-tab-pane label="测试预览" name="preview" v-if="visited.preview">
        <test-preview :get-payload="getPayload" />
      </el-tab-pane>
    </el-tabs>
  </el-drawer>
</template>

<script setup>
  import { ref, reactive, watch, nextTick } from 'vue'
  import { ElMessage } from 'element-plus'
  import { findDataSet, createDataSet, updateDataSet } from '@/api/report/dataSet'
  import BasicInfo from './basicInfo.vue'
  import ParamConfig from './paramConfig.vue'
  import TransformConfig from './transformConfig.vue'
  import TestPreview from './testPreview.vue'

  const props = defineProps({
    modelValue: { type: Boolean, default: false },
    type: { type: String, default: 'add' },
    rowId: { type: Number, default: 0 }
  })
  const emit = defineEmits(['update:modelValue', 'saved'])

  const visible = ref(false)
  watch(
    () => props.modelValue,
    (v) => {
      visible.value = v
      if (v) init()
    },
    { immediate: true }
  )
  watch(visible, (v) => emit('update:modelValue', v))

  const activeTab = ref('basic')
  const visited = reactive({ preview: false })
  const basicRef = ref(null)
  const paramRef = ref(null)
  const transformRef = ref(null)
  const form = ref({})
  const savedId = ref(0)

  const initForm = () => ({
    ID: 0,
    setCode: '',
    setName: '',
    setDesc: '',
    sourceCode: '',
    setType: 'sql',
    dynSentence: '',
    enableFlag: true,
    params: [],
    transforms: []
  })

  const init = async () => {
    activeTab.value = 'basic'
    visited.preview = false
    form.value = initForm()
    savedId.value = 0
    if (props.type === 'edit' && props.rowId) {
      const res = await findDataSet({ ID: props.rowId })
      if (res.code === 0) {
        form.value = { ...form.value, ...res.data }
        savedId.value = res.data.ID
      }
    }
    await nextTick()
    basicRef.value && basicRef.value.clearValidate()
  }

  const onClosed = () => {
    form.value = initForm()
  }

  const close = () => {
    visible.value = false
  }

  // 供测试预览 Tab 取当前基础信息 + 参数配置（编辑中即时测试场景）
  const getPayload = () => {
    const basic = basicRef.value ? basicRef.value.collect() : {}
    return {
      setType: basic.setType,
      sourceCode: basic.sourceCode,
      dynSentence: basic.dynSentence,
      params: form.value.params || [],
      setCode: savedId.value && props.type === 'edit' ? basic.setCode : ''
    }
  }

  // 逐 Tab 校验，失败切到对应 Tab
  const submit = async () => {
    activeTab.value = 'basic'
    const basic = basicRef.value ? await basicRef.value.validate() : null
    if (!basic) return
    const params = paramRef.value ? paramRef.value.collect() : []
    if (paramRef.value && !paramRef.value.validate()) {
      activeTab.value = 'params'
      return
    }
    const transforms = transformRef.value ? transformRef.value.collect() : []
    const req = {
      ID: savedId.value,
      setCode: basic.setCode,
      setName: basic.setName,
      setDesc: basic.setDesc,
      sourceCode: basic.sourceCode,
      setType: basic.setType,
      dynSentence: basic.dynSentence,
      enableFlag: basic.enableFlag,
      params,
      transforms
    }
    const isAdd = props.type === 'add'
    const res = isAdd ? await createDataSet(req) : await updateDataSet(req)
    if (res.code === 0) {
      ElMessage.success(isAdd ? '创建成功' : '更新成功')
      visible.value = false
      emit('saved')
    }
  }

  // 切到测试预览 Tab 时标记已访问（懒渲染，切过即保留）
  watch(activeTab, (tab) => {
    if (tab === 'preview') visited.preview = true
  })
</script>
