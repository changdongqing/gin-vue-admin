<template>
  <div>
    <warning-bar title="获取参数且缓存方法已在前端utils/params 已经封装完成 不必自己书写 使用方法查看文件内注释" />
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="openDialog">新增</el-button>
        <el-button icon="delete" :disabled="!selectedRows.length" @click="onBatchDelete(selectedRows)">
          删除
        </el-button>
      </template>

      <template #operate="{ row }">
        <el-button type="primary" link icon="info-filled" @click="getDetails(row)">查看详情</el-button>
        <el-button type="primary" link icon="edit" @click="updateSysParamsFunc(row)">变更</el-button>
        <el-button type="primary" link icon="delete" @click="deleteRow(row)">删除</el-button>
      </template>
    </GvaGrid>

    <el-drawer
      destroy-on-close
      size="800"
      v-model="dialogFormVisible"
      :show-close="false"
      :before-close="closeDialog"
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ type === 'create' ? '添加' : '修改' }}</span>
          <div>
            <el-button type="primary" @click="enterDialog">确 定</el-button>
            <el-button @click="closeDialog">取 消</el-button>
          </div>
        </div>
      </template>

      <el-form :model="formData" label-position="top" ref="elFormRef" :rules="rule" label-width="80px">
        <el-form-item label="参数名称:" prop="name">
          <el-input v-model="formData.name" :clearable="true" placeholder="请输入参数名称" />
        </el-form-item>
        <el-form-item label="参数键:" prop="key">
          <el-input v-model="formData.key" :clearable="true" placeholder="请输入参数键" />
        </el-form-item>
        <el-form-item label="参数值:" prop="value">
          <el-input
            type="textarea"
            :rows="5"
            v-model="formData.value"
            :clearable="true"
            placeholder="请输入参数值"
          />
        </el-form-item>
        <el-form-item label="参数说明:" prop="desc">
          <el-input v-model="formData.desc" :clearable="true" placeholder="请输入参数说明" />
        </el-form-item>
      </el-form>

      <div class="usage-instructions bg-gray-100 border border-gray-300 rounded-lg p-4 mt-5">
        <h3 class="mb-3 text-lg text-gray-800">使用说明</h3>
        <p class="mb-2 text-sm text-gray-600">
          前端可以通过引入
          <code class="bg-blue-100 px-1 py-0.5 rounded">import { getParams } from '@/utils/params'</code>
          然后通过
          <code class="bg-blue-100 px-1 py-0.5 rounded">await getParams("{{ formData.key }}")</code>
          来获取对应的参数。
        </p>
        <p class="text-sm text-gray-600">
          后端需要提前
          <code class="bg-blue-100 px-1 py-0.5 rounded">
            import "github.com/flipped-aurora/gin-vue-admin/server/service/system"
          </code>
        </p>
        <p class="mb-2 text-sm text-gray-600">
          然后调用
          <code class="bg-blue-100 px-1 py-0.5 rounded">
            new(system.SysParamsService).GetSysParam("{{ formData.key }}")
          </code>
          来获取对应的 value 值。
        </p>
      </div>
    </el-drawer>

    <el-drawer destroy-on-close size="800" v-model="detailShow" :show-close="true" :before-close="closeDetailShow">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="参数名称">
          {{ detailForm.name }}
        </el-descriptions-item>
        <el-descriptions-item label="参数键">
          {{ detailForm.key }}
        </el-descriptions-item>
        <el-descriptions-item label="参数值">
          {{ detailForm.value }}
        </el-descriptions-item>
        <el-descriptions-item label="参数说明">
          {{ detailForm.desc }}
        </el-descriptions-item>
      </el-descriptions>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    createSysParams,
    deleteSysParams,
    deleteSysParamsByIds,
    updateSysParams,
    findSysParams,
    getSysParamsList
  } from '@/api/sysParams'

  import { ElMessage } from 'element-plus'
  import { ref, reactive } from 'vue'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import GvaGrid, { useGvaGrid, useGvaGridDelete } from '@/components/gvaGrid'

  defineOptions({
    name: 'SysParams'
  })

  // 自动化生成的字典（可能为空）以及字段
  const formData = ref({
    name: '',
    key: '',
    value: '',
    desc: ''
  })

  // 验证规则
  const rule = reactive({
    name: [
      {
        required: true,
        message: '',
        trigger: ['input', 'blur']
      },
      {
        whitespace: true,
        message: '不能只输入空格',
        trigger: ['input', 'blur']
      }
    ],
    key: [
      {
        required: true,
        message: '',
        trigger: ['input', 'blur']
      },
      {
        whitespace: true,
        message: '不能只输入空格',
        trigger: ['input', 'blur']
      }
    ],
    value: [
      {
        required: true,
        message: '',
        trigger: ['input', 'blur']
      },
      {
        whitespace: true,
        message: '不能只输入空格',
        trigger: ['input', 'blur']
      }
    ]
  })

  const elFormRef = ref()

  const { gridRef, gridOptions, gridEvents, selectedRows, refresh } = useGvaGrid({
    id: 'superAdmin-sysParams',
    api: getSysParamsList,
    defaultSort: null,
    checkbox: true,
    searchItems: [
      {
        field: 'createdAt',
        title: '创建日期',
        span: 8,
        itemRender: {
          name: 'gvaDateRange',
          props: { type: 'datetimerange', valueFormat: 'YYYY-MM-DD HH:mm:ss', startField: 'startCreatedAt', endField: 'endCreatedAt' }
        }
      },
      { field: 'name', title: '参数名称', span: 6, itemRender: { name: 'VxeInput', props: { placeholder: '搜索条件', clearable: true } } },
      { field: 'key', title: '参数键', span: 6, itemRender: { name: 'VxeInput', props: { placeholder: '搜索条件', clearable: true } } }
    ],
    columns: [
      { field: 'CreatedAt', title: '日期', width: 180, cellRender: { name: 'gvaDate' } },
      { field: 'name', title: '参数名称', width: 120 },
      { field: 'key', title: '参数键', width: 120 },
      { field: 'value', title: '参数值', width: 120 },
      { field: 'desc', title: '参数说明', width: 120 },
      { title: '操作', fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  const { deleteRow, deleteRows: onBatchDelete } = useGvaGridDelete(gridRef, {
    delete: (rows) => {
      if (rows.length === 1) return deleteSysParams({ ID: rows[0].ID })
      return deleteSysParamsByIds({ IDs: rows.map((item) => item.ID) })
    },
    confirmText: '确定要删除吗?',
    successText: '删除成功'
  })

  // 行为控制标记（弹窗内部需要增还是改）
  const type = ref('')

  // 更新行
  const updateSysParamsFunc = async (row) => {
    const res = await findSysParams({ ID: row.ID })
    type.value = 'update'
    if (res.code === 0) {
      formData.value = res.data
      dialogFormVisible.value = true
    }
  }

  // 弹窗控制标记
  const dialogFormVisible = ref(false)

  // 打开弹窗
  const openDialog = () => {
    type.value = 'create'
    dialogFormVisible.value = true
  }

  // 关闭弹窗
  const closeDialog = () => {
    dialogFormVisible.value = false
    formData.value = {
      name: '',
      key: '',
      value: '',
      desc: ''
    }
  }
  // 弹窗确定
  const enterDialog = async () => {
    elFormRef.value?.validate(async (valid) => {
      if (!valid) return
      let res
      switch (type.value) {
        case 'create':
          res = await createSysParams(formData.value)
          break
        case 'update':
          res = await updateSysParams(formData.value)
          break
        default:
          res = await createSysParams(formData.value)
          break
      }
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '创建/更改成功'
        })
        closeDialog()
        refresh()
      }
    })
  }

  const detailForm = ref({})

  // 查看详情控制标记
  const detailShow = ref(false)

  // 打开详情弹窗
  const openDetailShow = () => {
    detailShow.value = true
  }

  // 打开详情
  const getDetails = async (row) => {
    // 打开弹窗
    const res = await findSysParams({ ID: row.ID })
    if (res.code === 0) {
      detailForm.value = res.data
      openDetailShow()
    }
  }

  // 关闭详情弹窗
  const closeDetailShow = () => {
    detailShow.value = false
    detailForm.value = {}
  }
</script>

<style></style>
