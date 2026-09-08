<template>
  <el-dialog v-model="visible" title="快速导入点表" width="860px" :close-on-click-modal="false" @closed="reset">
    <el-steps :active="step" align-center finish-status="success" class="mb-4">
      <el-step title="上传" />
      <el-step title="预览确认" />
      <el-step title="执行结果" />
    </el-steps>

    <!-- Step1 上传 -->
    <div v-if="step === 0" class="text-center py-6">
      <el-upload drag :auto-upload="false" :limit="1" accept=".xlsx" :on-change="onFileChange" :on-remove="() => (file = null)">
        <el-icon class="el-icon--upload"><upload-filled /></el-icon>
        <div class="el-upload__text">拖拽 xlsx 到此处，或 <em>点击选择</em></div>
      </el-upload>
      <div class="mt-3">
        <el-button icon="download" link type="primary" @click="downloadTemplate">下载导入模板</el-button>
      </div>
    </div>

    <!-- Step2 预览 -->
    <div v-else-if="step === 1">
      <el-alert v-if="errors.length" type="warning" :closable="false" class="mb-3"
        :title="`校验发现 ${errors.length} 个错误行（已剔除，不参与导入）`" />
      <el-alert v-else type="success" :closable="false" class="mb-3" title="校验通过" />
      <el-tabs v-model="tab">
        <el-tab-pane :label="`通道（${channels.length}）`" name="ch">
          <el-table :data="channels" size="small" border max-height="320">
            <el-table-column prop="name" label="名称" />
            <el-table-column prop="accessMode" label="接入" width="80" />
            <el-table-column prop="driver" label="驱动" width="90" />
            <el-table-column label="动作" width="80">
              <template #default="{ row }">
                <el-tag size="small" :type="row.ID ? 'primary' : 'success'">{{ row.ID ? '更新' : '新增' }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane :label="`设备（${devices.length}）`" name="dev">
          <el-table :data="devices" size="small" border max-height="320">
            <el-table-column prop="channelName" label="通道" min-width="110" />
            <el-table-column prop="name" label="设备" min-width="110" />
            <el-table-column prop="deviceKind" label="类别" width="90" />
            <el-table-column label="动作" width="80">
              <template #default="{ row }">
                <el-tag size="small" :type="row.ID ? 'primary' : 'success'">{{ row.ID ? '更新' : '新增' }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane :label="`测点（${variables.length}）`" name="var">
          <el-table :data="variables" size="small" border max-height="320">
            <el-table-column label="所属设备" min-width="140">
              <template #default="{ row }">{{ row.channelName }}/{{ row.deviceName }}</template>
            </el-table-column>
            <el-table-column prop="name" label="测点" min-width="100" />
            <el-table-column prop="addr" label="地址" min-width="80" />
            <el-table-column prop="dataType" label="类型" width="80" />
            <el-table-column prop="scale" label="系数" width="70" />
            <el-table-column label="动作" width="80">
              <template #default="{ row }">
                <el-tag size="small" :type="row.ID ? 'primary' : 'success'">{{ row.ID ? '更新' : '新增' }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane v-if="errors.length" :label="`错误（${errors.length}）`" name="err">
          <el-table :data="errors" size="small" border max-height="320">
            <el-table-column prop="sheet" label="sheet" width="80" />
            <el-table-column prop="row" label="行号" width="70" />
            <el-table-column prop="reason" label="原因" min-width="220" />
          </el-table>
        </el-tab-pane>
      </el-tabs>
      <div class="mt-3 flex items-center gap-3">
        <el-checkbox v-model="autoDeploy">自动编译部署变更通道（导入即上数）</el-checkbox>
      </div>
    </div>

    <!-- Step3 结果 -->
    <div v-else class="py-4">
      <el-result icon="success" title="导入完成"
        :sub-title="`通道 ${res.channels} / 设备 ${res.devices} / 测点 ${res.variables} 条已入库；部署成功 ${res.deployed}，跳过 ${res.skipped}`" />
      <el-alert v-if="Object.keys(res.deployFailures || {}).length" type="error" :closable="false"
        title="部分通道部署失败（可在采集配置页对单通道重试部署）">
        <div v-for="(msg, id) in res.deployFailures" :key="id">通道 {{ id }}：{{ msg }}</div>
      </el-alert>
    </div>

    <template #footer>
      <el-button v-if="step === 0" @click="visible = false">取 消</el-button>
      <el-button v-if="step === 1" @click="step = 0">上 一 步</el-button>
      <el-button v-if="step === 1" type="primary" :loading="committing" @click="doCommit">确认导入</el-button>
      <el-button v-if="step === 2" type="primary" @click="visible = false">完 成</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { importPreview, importCommit, downloadImportTemplate } from '@/api/collect'

const visible = ref(false)
const step = ref(0)
const tab = ref('ch')
const file = ref(null)
const channels = ref([])
const devices = ref([])
const variables = ref([])
const errors = ref([])
const autoDeploy = ref(true)
const committing = ref(false)
const res = ref({})
const emit = defineEmits(['imported'])

const onFileChange = async (uploadFile) => {
  file.value = uploadFile.raw
  const fd = new FormData()
  fd.append('file', uploadFile.raw)
  const r = await importPreview(fd)
  channels.value = (r.data.data && r.data.data.channels) || []
  devices.value = (r.data.data && r.data.data.devices) || []
  variables.value = (r.data.data && r.data.data.variables) || []
  errors.value = r.data.errors || []
  step.value = 1
}

const doCommit = async () => {
  committing.value = true
  try {
    const r = await importCommit({
      data: {
        channels: channels.value, devices: devices.value, variables: variables.value
      },
      autoDeploy: autoDeploy.value
    })
    res.value = r.data
    step.value = 2
    emit('imported')
    ElMessage.success('导入完成')
  } finally {
    committing.value = false
  }
}

const downloadTemplate = async () => {
  const r = await downloadImportTemplate()
  const blob = new Blob([r.data || r], { type: 'application/octet-stream' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'collect-import-template.xlsx'
  a.click()
  URL.revokeObjectURL(url)
}

const reset = () => {
  step.value = 0
  file.value = null
  channels.value = []
  devices.value = []
  variables.value = []
  errors.value = []
  res.value = {}
}
</script>
