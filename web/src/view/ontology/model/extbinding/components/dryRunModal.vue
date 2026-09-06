<template>
  <el-dialog v-model="visible" title="试运行（只读预览，不落库）" width="860px">
    <el-alert v-for="(w, i) in result.warnings" :key="i" :title="w" type="warning" show-icon class="mb-1" :closable="false" />
    <el-table :data="result.rows" border max-height="420" size="small" empty-text="无可预览行">
      <el-table-column prop="bizKey" label="主键" width="90" />
      <el-table-column prop="previewCode" label="预览编码" min-width="140" show-overflow-tooltip />
      <el-table-column prop="previewName" label="预览名称" min-width="120" show-overflow-tooltip />
      <el-table-column label="属性值" min-width="240">
        <template #default="{ row }">
          <div class="flex flex-wrap gap-1">
            <span v-for="(v, i) in row.values" :key="i" class="inline-flex items-center gap-1">
              <el-tag size="small" type="primary">{{ v.label }}</el-tag>
              <span class="text-xs text-gray-500">{{ v.source }} → {{ v.rawValue || '∅' }}</span>
              <el-tag v-if="!v.converted" size="small" type="danger">转换失败</el-tag>
            </span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="关系解析" min-width="180">
        <template #default="{ row }">
          <div class="flex flex-wrap gap-1">
            <span v-for="(r, i) in row.relations" :key="i" class="inline-flex items-center gap-1">
              <el-tag size="small" type="warning">{{ r.label }}</el-tag>
              <el-tag v-if="r.resolved" size="small" type="success">→ {{ r.targetObjectName }}</el-tag>
              <el-tag v-else size="small" type="danger">{{ r.targetObjectName }}</el-tag>
            </span>
          </div>
        </template>
      </el-table-column>
    </el-table>
    <template #footer>
      <el-button type="primary" @click="visible = false">关 闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
  import { dryRunExtSync } from '@/api/ontology/extSync'
  import { ref } from 'vue'
  import { ElMessage } from 'element-plus'

  defineOptions({ name: 'DryRunModal' })

  const visible = ref(false)
  const result = ref({ rows: [], warnings: [] })

  const open = async (bindingId) => {
    visible.value = true
    const res = await dryRunExtSync({ bindingId, limit: 20 })
    if (res.code === 0) {
      result.value = res.data || { rows: [], warnings: [] }
    }
  }
  defineExpose({ open })
  void ElMessage
</script>
