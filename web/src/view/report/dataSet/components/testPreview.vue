<template>
  <div>
    <div class="flex items-center gap-2 mb-3">
      <el-button type="primary" icon="video-play" :loading="loading" @click="run(1, pagination.pageSize)">执行查询</el-button>
      <span class="text-xs text-gray-400">按示例值预填参数；服务端分页（默认 20 行/页）；测试预览不写入结果案例</span>
    </div>

    <el-alert v-if="errorMsg" type="error" :title="errorMsg" :closable="true" class="mb-3" />

    <el-table v-if="columns.length" :data="rows" border size="small" max-height="480">
      <el-table-column v-for="col in columns" :key="col" :prop="col" :label="col" min-width="120" show-overflow-tooltip>
        <template #default="{ row }">
          {{ formatCell(row[col]) }}
        </template>
      </el-table-column>
    </el-table>
    <el-empty v-else-if="!loading && executed" description="无数据" />

    <div v-if="columns.length" class="mt-3 flex justify-end">
      <el-pagination
        v-model:current-page="pagination.pageNo"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="(p) => run(p, pagination.pageSize)"
        @size-change="(s) => run(1, s)"
      />
    </div>
  </div>
</template>

<script setup>
  import { ref, reactive } from 'vue'
  import { testDataSetPreview } from '@/api/report/dataSet'

  const props = defineProps({
    // 父组件提供当前基础信息+参数（编辑中即时测试）
    getPayload: { type: Function, required: true }
  })

  const loading = ref(false)
  const executed = ref(false)
  const errorMsg = ref('')
  const columns = ref([])
  const rows = ref([])
  const pagination = reactive({ pageNo: 1, pageSize: 20, total: 0 })

  // dateRange 参数示例值拼 "起,止" 提交（后端拆分 _start/_end）
  const buildParamValues = (params) => {
    const values = {}
    for (const p of params || []) {
      if (p.sampleItem !== undefined && p.sampleItem !== null && p.sampleItem !== '') {
        values[p.paramName] = p.sampleItem
      }
    }
    return values
  }

  const run = async (pageNo, pageSize) => {
    loading.value = true
    errorMsg.value = ''
    try {
      const payload = props.getPayload()
      const req = {
        pageNo,
        pageSize,
        paramValues: buildParamValues(payload.params)
      }
      if (payload.setCode) {
        // 已保存数据集场景：后端以库中配置优先
        req.setCode = payload.setCode
      } else {
        req.setType = payload.setType
        req.sourceCode = payload.sourceCode
        req.dynSentence = payload.dynSentence
      }
      const res = await testDataSetPreview(req)
      if (res.code === 0) {
        columns.value = res.data.columns || []
        rows.value = res.data.rows || []
        pagination.total = res.data.total || 0
        pagination.pageNo = res.data.pageNo || pageNo
        pagination.pageSize = res.data.pageSize || pageSize
        executed.value = true
      } else {
        errorMsg.value = res.msg
        columns.value = []
        rows.value = []
        executed.value = true
      }
    } finally {
      loading.value = false
    }
  }

  const formatCell = (v) => {
    if (v === null || v === undefined) return ''
    if (typeof v === 'object') return JSON.stringify(v)
    return String(v)
  }

  defineExpose({ run })
</script>
