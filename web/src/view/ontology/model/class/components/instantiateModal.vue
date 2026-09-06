<template>
  <el-dialog v-model="visible" title="从分类模板实例化" width="720px" :close-on-click-modal="false">
    <div class="flex gap-3">
      <div class="w-[260px] shrink-0 border border-gray-200 rounded p-2 h-[380px] overflow-auto">
        <el-tree
          ref="treeRef"
          :data="treeData"
          :props="{ label: 'label', children: 'children' }"
          node-key="templateCode"
          default-expand-all
          highlight-current
          @node-click="onNodeClick"
        />
        <el-empty v-if="!treeData.length" description="暂无分类模板" :image-size="48" />
      </div>
      <div class="flex-1 min-w-0">
        <div v-if="preview.templateCode" class="mb-2 flex items-center gap-2 flex-wrap">
          <el-tag type="primary">{{ preview.templateCode }}</el-tag>
          <el-tag v-if="preview.classificationCode">分类编码：{{ preview.classificationCode }}</el-tag>
          <span class="text-sm text-gray-500">将挂载 {{ preview.properties?.length || 0 }} 个属性</span>
        </div>
        <el-table :data="preview.properties || []" border max-height="340" size="small" empty-text="请在左侧选择分类模板">
          <el-table-column prop="propertyTemplateCode" label="属性模板" min-width="140" show-overflow-tooltip />
          <el-table-column prop="kind" label="种类" width="90" />
          <el-table-column prop="label" label="显示名" width="120" />
          <el-table-column prop="type" label="类型" width="110" />
          <el-table-column prop="source" label="来源" width="100" />
        </el-table>
      </div>
    </div>
    <template #footer>
      <el-button @click="visible = false">取 消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">确认实例化</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
  import { getClassTemplateTreeForSupply } from '@/api/ontology/classTemplate'
  import { previewInstantiateModelClass, instantiateModelClass } from '@/api/ontology/modelClass'
  import { ref, reactive } from 'vue'
  import { ElMessage } from 'element-plus'

  defineOptions({ name: 'InstantiateModal' })

  const emit = defineEmits(['success'])

  const visible = ref(false)
  const submitting = ref(false)
  const treeRef = ref(null)
  const treeData = ref([])
  const selectedCode = ref('')
  const preview = reactive({ templateCode: '', classificationCode: '', properties: [] })
  let ctx = { projectId: 0, classId: 0 }

  // 供给树为扁平列表，按 parentId 组树（node-key=templateCode）
  const buildTree = (nodes, parentId) => {
    return nodes
      .filter((item) => item.parentId === parentId)
      .map((item) => {
        const children = buildTree(nodes, item.ID)
        return children.length ? { ...item, children } : { ...item }
      })
  }

  const open = (projectId, classId) => {
    ctx = { projectId, classId }
    selectedCode.value = ''
    preview.templateCode = ''
    preview.properties = []
    visible.value = true
    getClassTemplateTreeForSupply().then((res) => {
      if (res.code === 0) {
        treeData.value = buildTree(res.data || [], 0)
      }
    })
  }
  defineExpose({ open })

  const onNodeClick = async (node) => {
    selectedCode.value = node.templateCode
    const res = await previewInstantiateModelClass({ projectId: ctx.projectId, templateCode: node.templateCode })
    if (res.code === 0) {
      Object.assign(preview, res.data)
    }
  }

  const submit = async () => {
    if (!selectedCode.value) {
      ElMessage.warning('请先在左侧选择分类模板')
      return
    }
    submitting.value = true
    try {
      const res = await instantiateModelClass({
        projectId: ctx.projectId,
        classId: ctx.classId,
        templateCode: selectedCode.value
      })
      if (res.code === 0) {
        ElMessage.success('实例化成功')
        visible.value = false
        emit('success')
      }
    } finally {
      submitting.value = false
    }
  }
</script>
