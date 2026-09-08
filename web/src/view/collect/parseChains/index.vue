<template>
  <div class="collect-parse-chains">
    <warning-bar title="子流程库：开发标准化交付的解析规则链（输入 string/json/bytes 报文 → 输出 {device, points:[{name,value,quality}]}）；发布后可被报文型设备类型引用" />
    <div class="mb-3">
      <el-button type="primary" icon="plus" @click="openForm()">新建子流程</el-button>
      <el-button icon="refresh" @click="load">刷新</el-button>
    </div>
    <el-table :data="list" size="small" border>
      <el-table-column prop="name" label="名称" min-width="160" />
      <el-table-column prop="version" label="版本" width="70" />
      <el-table-column prop="status" label="状态" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="row.status === 'published' ? 'success' : 'info'">
            {{ row.status === 'published' ? '已发布' : '草稿' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="inputContract" label="输入契约" min-width="140" show-overflow-tooltip />
      <el-table-column label="操作" width="230" fixed="right">
        <template #default="{ row }">
          <el-button icon="edit" type="primary" link :disabled="row.status === 'published'" @click="openForm(row)">编辑</el-button>
          <el-button icon="video-play" type="primary" link @click="doTest(row)">回放</el-button>
          <el-button icon="upload" type="primary" link :disabled="row.status === 'published'" @click="doPublish(row)">发布</el-button>
          <el-button icon="delete" type="primary" link @click="doDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-drawer v-model="formVisible" size="640px" :show-close="false">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ form.ID ? '编辑子流程' : '新建子流程' }}</span>
          <div>
            <el-button @click="formVisible = false">取 消</el-button>
            <el-button type="primary" @click="submit">确 定</el-button>
          </div>
        </div>
      </template>
      <el-form :model="form" label-width="90px">
        <el-form-item label="名称" required><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="输入契约"><el-input v-model="form.inputContract" placeholder="如：JSON 报文，含 deviceId 与测点字段" /></el-form-item>
        <el-form-item label="DSL" required>
          <el-input v-model="form.dsl" type="textarea" :rows="18" placeholder='{"ruleChain":{...},"metadata":{...}}' class="font-mono" />
        </el-form-item>
        <el-form-item label="测试报文">
          <el-input v-model="form.testPayload" type="textarea" :rows="4" placeholder='回放测试用的样例报文' />
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
    </el-drawer>

    <el-dialog v-model="testVisible" title="回放输出" width="700px">
      <pre class="bg-gray-900 text-green-300 text-xs p-3 rounded overflow-auto max-h-96">{{ testOut }}</pre>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import WarningBar from '@/components/warningBar/warningBar.vue'
import {
  getParseChainList, createParseChain, updateParseChain, deleteParseChain,
  publishParseChain, testParseChain
} from '@/api/collect'

defineOptions({ name: 'CollectParseChains' })

const list = ref([])
const formVisible = ref(false)
const form = ref({})
const testVisible = ref(false)
const testOut = ref('')

const load = async () => {
  const r = await getParseChainList()
  list.value = r.data || []
}
const openForm = (row) => {
  form.value = row ? { ...row } : { status: 'draft' }
  formVisible.value = true
}
const submit = async () => {
  if (form.value.ID) await updateParseChain(form.value)
  else await createParseChain(form.value)
  ElMessage.success('保存成功')
  formVisible.value = false
  load()
}
const doPublish = async (row) => {
  await ElMessageBox.confirm('发布后子流程链部署到引擎，可被报文型设备类型引用，确认？', '提示', { type: 'info' })
  await publishParseChain(row.ID)
  ElMessage.success('发布成功')
  load()
}
const doTest = async (row) => {
  const r = await testParseChain(row.ID)
  testOut.value = typeof r.data === 'string' ? r.data : JSON.stringify(r.data, null, 2)
  testVisible.value = true
}
const doDelete = async (row) => {
  await ElMessageBox.confirm(`删除子流程「${row.name}」，确认？`, '警告', { type: 'warning' })
  await deleteParseChain({ id: row.ID })
  ElMessage.success('删除成功')
  load()
}

onMounted(load)
</script>
