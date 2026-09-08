<template>
  <div class="collect-config">
    <warning-bar title="采集平台：通道→设备→测点三级配置为唯一事实源，保存后点「部署」编译为规则链并立即生效；也可用「快速导入」批量上数" />
    <div class="flex gap-4">
      <!-- 左：三级树 -->
      <el-card class="w-80 shrink-0" shadow="never">
        <template #header>
          <div class="flex justify-between items-center">
            <span>通道-设备-测点</span>
            <div>
              <el-button type="primary" icon="plus" link @click="openChannelForm()">通道</el-button>
              <el-button icon="refresh" link @click="loadTree" />
            </div>
          </div>
        </template>
        <el-tree
          ref="treeRef"
          :data="tree"
          node-key="nodeKey"
          :props="{ label: 'label', children: 'children' }"
          highlight-current
          @node-click="onNodeClick"
        >
          <template #default="{ data }">
            <span class="flex-1 flex justify-between items-center pr-1">
              <span>
                <el-tag v-if="data.kind === 'channel'" size="small" :type="deployBadge(data.raw).type" class="mr-1">
                  {{ deployBadge(data.raw).text }}
                </el-tag>
                {{ data.label }}
              </span>
              <span @click.stop>
                <el-dropdown trigger="click" @command="(cmd) => onNodeCommand(cmd, data)">
                  <el-button icon="more" link />
                  <template #dropdown>
                    <el-dropdown-menu>
                      <template v-if="data.kind === 'channel'">
                        <el-dropdown-item command="edit">编辑通道</el-dropdown-item>
                        <el-dropdown-item command="device">新增设备</el-dropdown-item>
                        <el-dropdown-item command="deploy" divided>部署</el-dropdown-item>
                        <el-dropdown-item command="undeploy">下线</el-dropdown-item>
                        <el-dropdown-item command="runlog">RunLog 调试</el-dropdown-item>
                        <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                      </template>
                      <template v-else-if="data.kind === 'device'">
                        <el-dropdown-item command="edit">编辑设备</el-dropdown-item>
                        <el-dropdown-item command="variable">新增测点</el-dropdown-item>
                        <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                      </template>
                      <template v-else>
                        <el-dropdown-item command="edit">编辑测点</el-dropdown-item>
                        <el-dropdown-item command="delete">删除</el-dropdown-item>
                      </template>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </span>
            </span>
          </template>
        </el-tree>
      </el-card>

      <!-- 右：详情表格 -->
      <el-card class="flex-1" shadow="never">
        <template #header>
          <div class="flex justify-between items-center">
            <span>{{ panelTitle }}</span>
            <div>
              <el-button icon="upload" @click="wizardVisible = true">快速导入</el-button>
              <el-button icon="download" @click="doExport">导出配置</el-button>
              <el-button icon="refresh" @click="doRebuildAll">全量重编译</el-button>
              <el-button v-if="selected.kind === 'device'" type="primary" icon="plus" @click="openVarForm()">新增测点</el-button>
            </div>
          </div>
        </template>

        <el-table v-if="selected.kind === 'device'" :data="selected.raw.variables || []" size="small" border>
          <el-table-column prop="name" label="测点" min-width="110" />
          <el-table-column prop="addr" label="地址" min-width="90" />
          <el-table-column prop="dataType" label="类型" width="80" />
          <el-table-column prop="scale" label="系数" width="70" />
          <el-table-column prop="offset" label="偏移" width="70" />
          <el-table-column prop="endian" label="字节序" width="70" />
          <el-table-column prop="unit" label="单位" width="70" />
          <el-table-column prop="rw" label="读写" width="60" />
          <el-table-column label="使能" width="70">
            <template #default="{ row }">
              <el-tag size="small" :type="row.enable ? 'success' : 'info'">{{ row.enable ? '启用' : '停用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button icon="edit" type="primary" link @click="openVarForm(row)">编辑</el-button>
              <el-button icon="delete" type="primary" link @click="deleteVar(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-table v-else-if="selected.kind === 'channel'" :data="selected.raw.devices || []" size="small" border>
          <el-table-column prop="name" label="设备" min-width="120" />
          <el-table-column prop="deviceKind" label="类别" width="90">
            <template #default="{ row }">
              <el-tag size="small">{{ row.deviceKind === 'register' ? '寄存器型' : '报文型' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="unitId" label="站号" width="70" />
          <el-table-column prop="pollInterval" label="周期ms" width="90" />
          <el-table-column label="测点数" width="80">
            <template #default="{ row }">{{ (row.variables || []).length }}</template>
          </el-table-column>
          <el-table-column label="使能" width="70">
            <template #default="{ row }">
              <el-tag size="small" :type="row.enable ? 'success' : 'info'">{{ row.enable ? '启用' : '停用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button icon="edit" type="primary" link @click="openDeviceForm(row)">编辑</el-button>
              <el-button icon="delete" type="primary" link @click="removeDevice(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-else description="左侧选择通道/设备查看详情；「快速导入」可批量上数" />
      </el-card>
    </div>

    <!-- 通道表单 -->
    <el-drawer v-model="channelVisible" :size="appStore.drawerSize" :show-close="false">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ channelForm.ID ? '编辑通道' : '新增通道' }}</span>
          <div>
            <el-button @click="channelVisible = false">取 消</el-button>
            <el-button type="primary" @click="submitChannel">确 定</el-button>
          </div>
        </div>
      </template>
      <el-form :model="channelForm" label-width="110px">
        <el-form-item label="通道名称" required><el-input v-model="channelForm.name" /></el-form-item>
        <el-form-item label="接入方式" required>
          <el-select v-model="channelForm.accessMode" style="width: 100%">
            <el-option label="轮询采集（poll）" value="poll" />
            <el-option label="主动上报（report/MQTT）" value="report" />
          </el-select>
        </el-form-item>
        <el-form-item label="驱动" required>
          <el-select v-model="channelForm.driver" style="width: 100%" allow-create filterable>
            <el-option v-for="d in drivers" :key="d" :label="d" :value="d" />
          </el-select>
        </el-form-item>
        <el-form-item label="连接配置JSON" required>
          <el-input v-model="channelForm.connConfig" type="textarea" :rows="4" placeholder='{"server":"tcp://192.168.1.100:502"}' />
        </el-form-item>
        <el-form-item v-if="channelForm.accessMode === 'poll'" label="轮询周期ms">
          <el-input-number v-model="channelForm.pollInterval" :min="1000" :step="500" />
        </el-form-item>
        <el-form-item v-if="channelForm.accessMode === 'report'" label="MQTT server">
          <el-input v-model="reportConn.server" placeholder="127.0.0.1:1883" />
        </el-form-item>
        <el-form-item v-if="channelForm.accessMode === 'report'" label="订阅topic">
          <el-input v-model="reportConn.topic" placeholder="sensors/+/data" />
        </el-form-item>
        <el-form-item label="使能"><el-switch v-model="channelForm.enable" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="channelForm.remark" type="textarea" /></el-form-item>
      </el-form>
    </el-drawer>

    <!-- 设备表单 -->
    <el-drawer v-model="deviceVisible" :size="appStore.drawerSize" :show-close="false">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ deviceForm.ID ? '编辑设备' : '新增设备' }}</span>
          <div>
            <el-button @click="deviceVisible = false">取 消</el-button>
            <el-button type="primary" @click="submitDevice">确 定</el-button>
          </div>
        </div>
      </template>
      <el-form :model="deviceForm" label-width="110px">
        <el-form-item label="所属通道">{{ deviceForm.channelName }}</el-form-item>
        <el-form-item label="设备名称" required><el-input v-model="deviceForm.name" /></el-form-item>
        <el-form-item label="设备类别" required>
          <el-select v-model="deviceForm.deviceKind" style="width: 100%">
            <el-option label="寄存器型（点表直采）" value="register" />
            <el-option label="报文型（子流程解析）" value="report" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="deviceForm.deviceKind === 'report'" label="设备类型">
          <el-select v-model="deviceForm.deviceTypeId" style="width: 100%">
            <el-option v-for="t in deviceTypes" :key="t.ID" :label="t.name" :value="t.ID" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="deviceForm.deviceKind === 'register'" label="站号"><el-input-number v-model="deviceForm.unitId" :min="0" :max="255" /></el-form-item>
        <el-form-item label="采集周期ms"><el-input-number v-model="deviceForm.pollInterval" :min="0" :step="500" /></el-form-item>
        <el-form-item label="使能"><el-switch v-model="deviceForm.enable" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="deviceForm.remark" type="textarea" /></el-form-item>
      </el-form>
    </el-drawer>

    <!-- 测点表单 -->
    <el-drawer v-model="varVisible" :size="appStore.drawerSize" :show-close="false">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ varForm.ID ? '编辑测点' : '新增测点' }}</span>
          <div>
            <el-button @click="varVisible = false">取 消</el-button>
            <el-button type="primary" @click="submitVar">确 定</el-button>
          </div>
        </div>
      </template>
      <el-form :model="varForm" label-width="110px">
        <el-form-item label="测点名称" required><el-input v-model="varForm.name" /></el-form-item>
        <el-form-item label="地址" required><el-input v-model="varForm.addr" placeholder="40001 / ai:1 / DB1.DBD0" /></el-form-item>
        <el-form-item label="数据类型" required>
          <el-select v-model="varForm.dataType" style="width: 100%">
            <el-option v-for="t in ['INT16','UINT16','INT32','UINT32','FLOAT32','FLOAT64','BOOL','STRING']" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="系数"><el-input-number v-model="varForm.scale" :precision="6" :step="0.1" /></el-form-item>
        <el-form-item label="偏移"><el-input-number v-model="varForm.offset" :precision="6" :step="0.1" /></el-form-item>
        <el-form-item label="字节序">
          <el-select v-model="varForm.endian" style="width: 100%" clearable>
            <el-option v-for="e in ['ABCD','CDAB','BADC','DCBA']" :key="e" :label="e" :value="e" />
          </el-select>
        </el-form-item>
        <el-form-item label="采集分组"><el-input v-model="varForm.collectGroup" /></el-form-item>
        <el-form-item label="读写"><el-select v-model="varForm.rw" style="width:100%"><el-option label="R" value="R" /><el-option label="RW" value="RW" /></el-select></el-form-item>
        <el-form-item label="单位"><el-input v-model="varForm.unit" /></el-form-item>
        <el-form-item label="使能"><el-switch v-model="varForm.enable" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="varForm.remark" type="textarea" /></el-form-item>
      </el-form>
    </el-drawer>

    <!-- 导入向导 -->
    <import-wizard v-model="wizardVisible" @imported="loadTree" />

    <!-- RunLog 调试 -->
    <runlog-drawer v-model="runlogVisible" :chain-id="runlogChainId" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { UploadFilled } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import WarningBar from '@/components/warningBar/warningBar.vue'
import { useAppStore } from '@/pinia'
import {
  getCollectTree, createChannel, updateChannel, deleteChannel,
  createDevice, updateDevice, deleteDevice,
  createVariable, updateVariable, deleteVariable,
  deployChannel, undeployChannel, rebuildAll, exportConfig, getDeviceTypeList
} from '@/api/collect'
import ImportWizard from './importWizard.vue'
import RunlogDrawer from '../components/runlogDrawer.vue'

defineOptions({ name: 'CollectConfig' })

const appStore = useAppStore()
const tree = ref([])
const selected = ref({ kind: 'none', raw: {} })
const drivers = ['modbus', 'bacnet', 's7', 'opcua', 'snmp', 'fins', 'mc', 'iec104', 'dlt645', 'eip']
const deviceTypes = ref([])
const wizardVisible = ref(false)
const runlogVisible = ref(false)
const runlogChainId = ref('')

const panelTitle = computed(() => {
  if (selected.value.kind === 'channel') return `通道「${selected.value.raw.name}」的设备`
  if (selected.value.kind === 'device') return `设备「${selected.value.raw.name}」的测点`
  return '详情'
})

const loadTree = async () => {
  const res = await getCollectTree()
  const decorate = (list) =>
    (list || []).map((n) => ({
      ...n,
      nodeKey: `${n.ID}`,
      label: n.name,
      children: (n.devices || []).map((d) => ({
        ...d,
        kind: 'device',
        nodeKey: `d${d.ID}`,
        label: d.name,
        children: (d.variables || []).map((v) => ({
          ...v, kind: 'variable', nodeKey: `v${v.ID}`, label: `${v.name}（${v.addr}）`
        }))
      }))
    }))
  tree.value = decorate((res.data || []).map((c) => ({ ...c, kind: 'channel' })))
  try {
    const t = await getDeviceTypeList()
    deviceTypes.value = t.data || []
  } catch (e) { /* 设备类型接口失败不阻塞 */ }
}

const deployBadge = (ch) => {
  // 树上通道部署徽标：以最新 deployment 状态为准（由 status 接口懒加载可优化，一期用 enable 简化）
  return ch.enable ? { text: '已启用', type: 'success' } : { text: '已停用', type: 'info' }
}

const onNodeClick = (data) => { selected.value = { kind: data.kind, raw: data } }

const onNodeCommand = async (cmd, data) => {
  const node = data.kind === 'channel' ? data : data
  switch (cmd) {
    case 'edit':
      if (data.kind === 'channel') openChannelForm(data)
      else if (data.kind === 'device') openDeviceForm(data)
      else openVarForm(data)
      break
    case 'device': openDeviceForm(null, data.ID, data.name); break
    case 'variable': openVarForm(null, data.ID, data.name); break
    case 'deploy':
      await deployChannel(data.ID)
      ElMessage.success(data.deploySkipped ? '配置无变更，跳过部署' : '部署成功')
      break
    case 'undeploy':
      await ElMessageBox.confirm('下线后停止采集，确认？', '提示', { type: 'warning' })
      await undeployChannel(data.ID)
      ElMessage.success('已下线')
      break
    case 'runlog':
      runlogChainId.value = `collect_ch_${data.ID}`
      runlogVisible.value = true
      break
    case 'delete':
      if (data.kind === 'channel') {
        await ElMessageBox.confirm(`删除通道「${data.name}」及其全部设备/测点，并下线规则链，确认？`, '警告', { type: 'warning' })
        await deleteChannel({ id: data.ID })
      } else if (data.kind === 'device') {
        await ElMessageBox.confirm(`删除设备「${data.name}」及其测点，确认？`, '警告', { type: 'warning' })
        await deleteDevice({ id: data.ID })
      } else {
        await deleteVariable({ id: data.ID })
      }
      ElMessage.success('删除成功')
      loadTree()
      break
  }
}

// 通道表单
const channelVisible = ref(false)
const channelForm = ref({})
const reportConn = ref({ server: '', topic: '' })
const openChannelForm = (row) => {
  if (row) {
    channelForm.value = { ...row }
    try {
      const conn = JSON.parse(row.connConfig || '{}')
      reportConn.value = { server: conn.server || '', topic: conn.topic || '' }
    } catch (e) { reportConn.value = { server: '', topic: '' } }
  } else {
    channelForm.value = { accessMode: 'poll', driver: 'modbus', enable: true, pollInterval: 5000 }
    reportConn.value = { server: '', topic: '' }
  }
  channelVisible.value = true
}
const submitChannel = async () => {
  const f = { ...channelForm.value }
  if (f.accessMode === 'report') {
    const base = f.ID ? safeParse(f.connConfig) : {}
    base.server = reportConn.value.server
    base.topic = reportConn.value.topic
    f.connConfig = JSON.stringify(base)
  }
  if (f.ID) await updateChannel(f)
  else await createChannel(f)
  ElMessage.success('保存成功，部署后生效')
  channelVisible.value = false
  loadTree()
}
const safeParse = (s) => { try { return JSON.parse(s || '{}') } catch (e) { return {} } }

// 设备表单
const deviceVisible = ref(false)
const deviceForm = ref({})
const openDeviceForm = (row, channelId, channelName) => {
  if (row) deviceForm.value = { ...row, channelName: findChannelName(row.channelId) }
  else deviceForm.value = { channelId, channelName, deviceKind: 'register', enable: true, unitId: 1 }
  deviceVisible.value = true
}
const findChannelName = (id) => {
  const node = tree.value.find((c) => c.ID === id)
  return node ? node.name : ''
}
const submitDevice = async () => {
  if (deviceForm.value.ID) await updateDevice(deviceForm.value)
  else await createDevice(deviceForm.value)
  ElMessage.success('保存成功')
  deviceVisible.value = false
  loadTree()
}
const removeDevice = async (row) => {
  await ElMessageBox.confirm(`删除设备「${row.name}」及其测点，确认？`, '警告', { type: 'warning' })
  await deleteDevice({ id: row.ID })
  ElMessage.success('删除成功')
  loadTree()
}

// 测点表单
const varVisible = ref(false)
const varForm = ref({})
const openVarForm = (row) => {
  const dev = selected.value.raw
  if (!dev || selected.value.kind !== 'device') {
    ElMessage.warning('请先在左侧选择设备')
    return
  }
  varForm.value = row ? { ...row } : { deviceId: dev.ID, dataType: 'FLOAT32', rw: 'R', enable: true }
  varVisible.value = true
}
const submitVar = async () => {
  if (varForm.value.ID) await updateVariable(varForm.value)
  else await createVariable(varForm.value)
  ElMessage.success('保存成功，重新部署通道后生效')
  varVisible.value = false
  loadTree()
}
const deleteVar = async (row) => {
  await deleteVariable({ id: row.ID })
  ElMessage.success('删除成功')
  loadTree()
}

const doExport = async () => {
  const res = await exportConfig()
  downloadBlob(res, 'collect-export.xlsx')
}
const doRebuildAll = async () => {
  const res = await rebuildAll()
  const d = res.data
  ElMessage.success(`重编译完成：成功 ${d.ok} / 跳过 ${d.skipped} / 失败 ${Object.keys(d.failures || {}).length}`)
}

const downloadBlob = (res, filename) => {
  const blob = new Blob([res.data || res], { type: 'application/octet-stream' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}
defineExpose({ downloadBlob })

onMounted(loadTree)
</script>
