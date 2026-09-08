<template>
  <el-drawer v-model="visible" size="560px" :title="`RunLog 实时调试 —— ${chainId}`">
    <div class="mb-2 text-sm text-gray-500">
      直连 rulego RunLog WebSocket（?token= 认证）；触发一次采集即可看到节点级日志流。
    </div>
    <div class="mb-2">
      <el-button size="small" type="primary" :disabled="connected" @click="connect">{{ connected ? '已连接' : '连接' }}</el-button>
      <el-button size="small" @click="disconnect">断开</el-button>
      <el-button size="small" @click="logs = []">清空</el-button>
    </div>
    <div ref="logBox" class="runlog-box bg-gray-900 text-green-300 text-xs font-mono rounded p-2 overflow-auto">
      <div v-for="(l, i) in logs" :key="i" class="whitespace-pre-wrap break-all">{{ l }}</div>
      <div v-if="!logs.length" class="text-gray-500">暂无日志…</div>
    </div>
  </el-drawer>
</template>

<script setup>
import { ref, watch, nextTick, onBeforeUnmount } from 'vue'
import { useUserStore } from '@/pinia/modules/user'

// RunLog 调试抽屉（03 文档 §7.4）：WS 直连 rulego /logs/ws/:chainId/:clientId?token=
const visible = ref(false)
const props = defineProps({ modelValue: Boolean, chainId: String })
const emit = defineEmits(['update:modelValue'])

const userStore = useUserStore()
const logs = ref([])
const connected = ref(false)
let ws = null
let clientId = ''
const logBox = ref(null)

watch(() => props.modelValue, (v) => {
  visible.value = v
  if (v && !connected.value) connect()
})
watch(visible, (v) => emit('update:modelValue', v))
watch(logs, () => nextTick(() => {
  if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight
}))

const connect = () => {
  if (!props.chainId) return
  disconnect()
  clientId = `gva-${Date.now()}`
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  const base = import.meta.env.VITE_BASE_API || ''
  const wsBase = base.replace(/^http/, 'ws')
  const url = `${wsBase}/rulego/logs/ws/${props.chainId}/${clientId}?token=${userStore.token}`
  ws = new WebSocket(url)
  ws.onopen = () => {
    connected.value = true
    logs.value.push(`[连接] ${props.chainId}`)
  }
  ws.onmessage = (e) => {
    logs.value.push(e.data)
    if (logs.value.length > 500) logs.value.splice(0, logs.value.length - 500)
  }
  ws.onclose = () => {
    connected.value = false
    logs.value.push('[断开]')
  }
  ws.onerror = () => logs.value.push('[错误] 连接异常（检查通道是否已部署）')
}

const disconnect = () => {
  if (ws) { ws.close(); ws = null }
  connected.value = false
}

onBeforeUnmount(disconnect)
</script>

<style scoped>
.runlog-box {
  height: calc(100vh - 220px);
}
</style>
