<template>
  <div>
    <warning-bar
      href="https://www.bilibili.com/video/BV1kv4y1g7nT?p=3"
      title="此功能为开发环境使用，不建议发布到生产，具体使用效果请看视频https://www.bilibili.com/video/BV1kv4y1g7nT?p=3"
    />
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="openDialog('addApi')">新增</el-button>
      </template>

      <template #operate="{ row }">
        <el-button icon="delete" type="primary" link @click="deleteApiFunc(row)">删除</el-button>
      </template>
    </GvaGrid>

    <el-drawer v-model="dialogFormVisible" size="40%" :show-close="false">
      <warning-bar title="模板package会创建集成于项目本体中的代码包，模板plugin会创建插件包" />
      <el-form ref="pkgForm" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="包名" prop="packageName">
          <el-input v-model="form.packageName" autocomplete="off" />
        </el-form-item>
        <el-form-item label="模板" prop="template">
          <el-select v-model="form.template">
            <el-option
              v-for="template in templatesOptions"
              :label="template"
              :value="template"
              :key="template"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="展示名" prop="label">
          <el-input v-model="form.label" autocomplete="off" />
        </el-form-item>
        <el-form-item label="描述" prop="desc">
          <el-input v-model="form.desc" autocomplete="off" />
        </el-form-item>
      </el-form>
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">创建Package</span>
          <div>
            <el-button @click="closeDialog"> 取 消 </el-button>
            <el-button type="primary" @click="enterDialog"> 确 定 </el-button>
          </div>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    createPackageApi,
    getPackageApi,
    deletePackageApi,
    getTemplatesApi
  } from '@/api/autoCode'
  import { ref } from 'vue'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import GvaGrid, { useGvaGrid, useGvaGridDelete } from '@/components/gvaGrid'

  defineOptions({
    name: 'AutoPkg'
  })

  const form = ref({
    packageName: '',
    template: '',
    label: '',
    desc: ''
  })
  const templatesOptions = ref([])

  const getTemplates = async () => {
    const res = await getTemplatesApi()
    if (res.code === 0) {
      templatesOptions.value = res.data
    }
  }

  getTemplates()

  const validateData = (rule, value, callback) => {
    if (/[\u4E00-\u9FA5]/g.test(value)) {
      callback(new Error('不能为中文'))
    } else if (/^\d+$/.test(value[0])) {
      callback(new Error('不能够以数字开头'))
    } else if (!/^[a-zA-Z0-9_]+$/.test(value)) {
      callback(new Error('只能包含英文字母、数字和下划线'))
    } else {
      callback()
    }
  }

  const rules = ref({
    packageName: [
      { required: true, message: '请输入包名', trigger: 'blur' },
      { validator: validateData, trigger: 'blur' }
    ],
    template: [
      { required: true, message: '请选择模板', trigger: 'change' },
      { validator: validateData, trigger: 'blur' }
    ]
  })

  const dialogFormVisible = ref(false)
  const openDialog = () => {
    dialogFormVisible.value = true
  }

  const closeDialog = () => {
    dialogFormVisible.value = false
    form.value = {
      packageName: '',
      template: '',
      label: '',
      desc: ''
    }
  }

  const pkgForm = ref(null)
  const enterDialog = async () => {
    pkgForm.value.validate(async (valid) => {
      if (valid) {
        const res = await createPackageApi(form.value)
        if (res.code === 0) {
          ElMessage({
            type: 'success',
            message: '添加成功',
            showClose: true
          })
        }
        refresh()
        closeDialog()
      }
    })
  }

  const { gridRef, gridOptions, gridEvents, refresh } = useGvaGrid({
    id: 'systemTools-autoPkg',
    pager: false,
    defaultSort: null,
    api: async () => {
      const res = await getPackageApi()
      const list = (res.data && res.data.pkgs) || []
      return { ...res, data: { list, total: list.length } }
    },
    columns: [
      { field: 'ID', title: 'id', width: 120 },
      { field: 'packageName', title: '包名', width: 150 },
      { field: 'template', title: '模板', width: 150 },
      { field: 'label', title: '展示名', width: 150 },
      { field: 'desc', title: '描述', minWidth: 150 },
      { title: '操作', width: 200, slots: { default: 'operate' } }
    ]
  })

  const { deleteRow: deleteApiFunc } = useGvaGridDelete(gridRef, {
    delete: (rows) => deletePackageApi(rows[0]),
    confirmText: '此操作仅删除数据库中的pkg存储，后端相应目录结构请自行删除与数据库保持一致！',
    successText: '删除成功!'
  })
</script>
