<template>
  <el-form ref="formRef" :model="form" :rules="rules" label-width="110px" class="max-w-3xl">
    <el-form-item label="数据集编码" prop="setCode">
      <el-input v-model="form.setCode" :disabled="editMode" placeholder="唯一标识，如 sys_user_list" />
    </el-form-item>
    <el-form-item label="数据集名称" prop="setName">
      <el-input v-model="form.setName" placeholder="如 系统用户清单" />
    </el-form-item>
    <el-form-item label="描述" prop="setDesc">
      <el-input v-model="form.setDesc" type="textarea" :rows="2" />
    </el-form-item>
    <el-form-item label="类型" prop="setType">
      <el-radio-group v-model="form.setType" :disabled="editMode">
        <el-radio value="sql">SQL 查询</el-radio>
        <el-radio value="http">HTTP 请求</el-radio>
      </el-radio-group>
    </el-form-item>
    <el-form-item v-if="form.setType === 'sql'" label="数据源" prop="sourceCode">
      <el-select v-model="form.sourceCode" placeholder="请选择已启用的数据源" filterable>
        <el-option v-for="item in dataSourceOptions" :key="item.sourceCode" :label="`${item.sourceName}（${item.sourceCode}）`" :value="item.sourceCode" />
      </el-select>
    </el-form-item>

    <template v-if="form.setType === 'sql'">
      <el-form-item label="查询语句" prop="dynSentence">
        <el-input
          v-model="form.dynSentence"
          type="textarea"
          :rows="8"
          class="font-mono"
          placeholder="仅允许 SELECT/WITH 查询；${paramName} 为参数占位符，<if param=&quot;x&quot;>…</if> 为条件片段（x 为空时整段剥离）"
        />
      </el-form-item>
      <el-form-item label=" ">
        <div class="text-xs text-gray-400 leading-5">
          示例：SELECT * FROM sys_users WHERE 1=1<br />
          &lt;if param="username"&gt; AND username LIKE CONCAT('%', ${username}, '%')&lt;/if&gt;<br />
          多选参数值以英文逗号分隔，自动展开 IN (?,?)
        </div>
      </el-form-item>
    </template>

    <template v-else>
      <el-form-item label="请求地址" prop="apiUrl">
        <el-input v-model="apiUrl" placeholder="https://api.example.com/data?date=${date}" />
      </el-form-item>
      <el-form-item label="请求方法" prop="method">
        <el-select v-model="method" style="width: 160px">
          <el-option label="GET" value="GET" />
          <el-option label="POST" value="POST" />
        </el-select>
      </el-form-item>
      <el-form-item label="请求头">
        <div class="w-full">
          <div v-for="(h, idx) in headers" :key="idx" class="flex gap-2 mb-2">
            <el-input v-model="h.key" placeholder="Header 名" style="width: 240px" />
            <el-input v-model="h.value" placeholder="值（支持 ${param}）" />
            <el-button icon="delete" @click="headers.splice(idx, 1)" />
          </div>
          <el-button icon="plus" @click="headers.push({ key: '', value: '' })">添加请求头</el-button>
        </div>
      </el-form-item>
      <el-form-item label="请求体(JSON)">
        <el-input
          v-model="bodyText"
          type="textarea"
          :rows="5"
          class="font-mono"
          placeholder='{"keyword": "${username}", "page": {"size": 100}}'
        />
      </el-form-item>
      <el-form-item label=" ">
        <div class="text-xs text-gray-400">URL/请求头做字符串替换，请求体做深度替换；响应解析优先级：JSON 数组 → 对象的 data 数组 → 对象本身单行</div>
      </el-form-item>
    </template>

    <el-form-item label="是否启用" prop="enableFlag">
      <el-switch v-model="form.enableFlag" />
    </el-form-item>
  </el-form>
</template>

<script setup>
  import { ref, watch, onMounted, defineExpose } from 'vue'
  import { getDataSourceAll } from '@/api/report/dataSource'

  const props = defineProps({
    form: { type: Object, required: true },
    editMode: { type: Boolean, default: false }
  })

  const formRef = ref(null)
  const dataSourceOptions = ref([])

  const rules = {
    setCode: [
      { required: true, message: '请输入数据集编码', trigger: 'blur' },
      { pattern: /^[a-zA-Z0-9_]{1,50}$/, message: '仅允许字母/数字/下划线（≤50字符）', trigger: 'blur' }
    ],
    setName: [{ required: true, message: '请输入数据集名称', trigger: 'blur' }],
    setType: [{ required: true, message: '请选择类型', trigger: 'change' }],
    sourceCode: [{ required: true, message: 'SQL 类型必须选择数据源', trigger: 'change' }]
  }

  // HTTP 配置（form.dynSentence 存 JSON；apiUrl/method/headers/body 为视图态）
  const apiUrl = ref('')
  const method = ref('GET')
  const headers = ref([])
  const bodyText = ref('')

  const buildDynSentence = () => {
    if (props.form.setType !== 'http') return props.form.dynSentence
    let body = {}
    try {
      if (bodyText.value.trim()) body = JSON.parse(bodyText.value)
    } catch (e) {
      body = { __invalid__: bodyText.value }
    }
    const headerMap = {}
    headers.value.forEach((h) => {
      if (h.key) headerMap[h.key] = h.value
    })
    return JSON.stringify({ apiUrl: apiUrl.value, method: method.value, headers: headerMap, body })
  }

  const loadHttpView = () => {
    if (props.form.setType !== 'http' || !props.form.dynSentence) return
    try {
      const cfg = JSON.parse(props.form.dynSentence)
      apiUrl.value = cfg.apiUrl || ''
      method.value = cfg.method || 'GET'
      headers.value = Object.entries(cfg.headers || {}).map(([key, value]) => ({ key, value }))
      bodyText.value = cfg.body && Object.keys(cfg.body).length ? JSON.stringify(cfg.body, null, 2) : ''
    } catch (e) {
      /* 保留空视图态 */
    }
  }

  watch(() => props.form.setType, loadHttpView, { immediate: true })

  onMounted(async () => {
    const res = await getDataSourceAll()
    if (res.code === 0) {
      dataSourceOptions.value = res.data || []
    }
  })

  const validate = async () => {
    let valid = true
    try {
      await formRef.value.validate()
    } catch (e) {
      valid = false
    }
    // HTTP 场景即时校验：请求地址必填、请求体须为合法 JSON
    if (valid && props.form.setType === 'http') {
      if (!apiUrl.value || !apiUrl.value.trim()) {
        ElMessage.error('请填写请求地址')
        return null
      }
      if (bodyText.value.trim()) {
        try {
          JSON.parse(bodyText.value)
        } catch (e) {
          ElMessage.error('请求体必须是合法 JSON')
          return null
        }
      }
    }
    if (!valid) return null
    return { ...props.form, dynSentence: buildDynSentence() }
  }

  const collect = () => ({ ...props.form, dynSentence: buildDynSentence() })
  const clearValidate = () => formRef.value && formRef.value.clearValidate()

  defineExpose({ validate, collect, clearValidate })
</script>

<style lang="scss" scoped>
  .font-mono :deep(textarea) {
    font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
    font-size: 12px;
  }
</style>
