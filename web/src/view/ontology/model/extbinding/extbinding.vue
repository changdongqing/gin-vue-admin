<template>
  <div class="ext-binding">
    <warning-bar title="注：外部模块关联把既有业务表接入本体（注册 → 绑定 → 试运行 → 生效 → 同步物化对象）；业务表只读零改动，v1 单向同步" />
    <el-tabs v-model="activeTab">
      <el-tab-pane label="类绑定配置" name="binding">
        <class-binding ref="bindingTabRef" @create="openForm()" @edit="(row) => openForm(row)" />
      </el-tab-pane>
      <el-tab-pane label="模块与表" name="registry">
        <registry ref="registryTabRef" />
      </el-tab-pane>
      <el-tab-pane label="同步中心" name="sync">
        <sync-center ref="syncTabRef" />
      </el-tab-pane>
    </el-tabs>

    <binding-form ref="formRef" @saved="onSaved" @dry-run="openDryRun" />
    <dry-run-modal ref="dryRunRef" />
  </div>
</template>

<script setup>
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref } from 'vue'
  import ClassBinding from './components/classBinding.vue'
  import Registry from './components/registry.vue'
  import BindingForm from './components/bindingForm.vue'
  import DryRunModal from './components/dryRunModal.vue'
  import SyncCenter from './components/syncCenter.vue'

  defineOptions({
    name: 'OntExtBinding'
  })

  const activeTab = ref('binding')
  const bindingTabRef = ref(null)
  const registryTabRef = ref(null)
  const syncTabRef = ref(null)
  const formRef = ref(null)
  const dryRunRef = ref(null)

  void registryTabRef
  void syncTabRef

  const openForm = (row = null) => {
    formRef.value.open(row)
  }
  const onSaved = () => {
    bindingTabRef.value && bindingTabRef.value.refresh()
    syncTabRef.value && syncTabRef.value.refresh()
  }
  const openDryRun = (bindingId) => {
    dryRunRef.value.open(bindingId)
  }
</script>
