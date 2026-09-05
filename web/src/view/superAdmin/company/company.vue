<template>
  <div class="company">
    <warning-bar title="注：公司为经营主体维度，部门挂靠在公司下；数据权限（本公司/本部门等）基于该组织树生效" />
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="addCompany(0)">新增公司</el-button>
      </template>

      <template #status="{ row }">
        <el-tag :type="row.status === 1 ? 'success' : 'danger'">
          {{ row.status === 1 ? '启用' : '停用' }}
        </el-tag>
      </template>

      <template #operate="{ row }">
        <el-button icon="plus" type="primary" link @click="addCompany(row.ID)">新增子级</el-button>
        <el-button icon="edit" type="primary" link @click="editCompany(row)">编辑</el-button>
        <el-button icon="delete" type="primary" link @click="deleteCompanyFunc(row)">删除</el-button>
      </template>
    </GvaGrid>

    <el-drawer v-model="formVisible" :size="appStore.drawerSize" :show-close="false">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ titleForm }}</span>
          <div>
            <el-button @click="closeForm">取 消</el-button>
            <el-button type="primary" @click="submitForm">确 定</el-button>
          </div>
        </div>
      </template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="父级公司" prop="parentId">
          <el-cascader
            v-model="form.parentId"
            style="width: 100%"
            :options="parentOptions"
            :props="{
              checkStrictly: true,
              label: 'name',
              value: 'ID',
              disabled: 'disabled',
              emitPath: false
            }"
            :show-all-levels="false"
            filterable
          />
        </el-form-item>
        <el-form-item label="公司名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入公司名称" />
        </el-form-item>
        <el-form-item label="公司编码" prop="code">
          <el-input v-model="form.code" placeholder="请输入公司编码(唯一)" />
        </el-form-item>
        <el-form-item label="负责人" prop="leader">
          <el-input v-model="form.leader" placeholder="请输入负责人" />
        </el-form-item>
        <el-form-item label="联系电话" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="地址" prop="address">
          <el-input v-model="form.address" placeholder="请输入地址" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch v-model="form.status" inline-prompt :active-value="1" :inactive-value="2" />
        </el-form-item>
        <el-form-item label="备注" prop="remarks">
          <el-input v-model="form.remarks" type="textarea" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    getCompanyList,
    createCompany,
    updateCompany,
    deleteCompany
  } from '@/api/company'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useAppStore } from '@/pinia'
  import GvaGrid, { useGvaGrid } from '@/components/gvaGrid'

  defineOptions({
    name: 'Company'
  })

  const appStore = useAppStore()

  // 公司树：无分页无搜索，接口直接返回树数组
  const companyTree = ref([])
  const { gridRef, gridOptions, gridEvents, refresh } = useGvaGrid({
    id: 'superAdmin-company',
    pager: false,
    defaultSort: null,
    api: async (data) => {
      const res = await getCompanyList({ name: (data && data.name) || '' })
      companyTree.value = res.data.list || []
      return { ...res, data: { list: companyTree.value, total: companyTree.value.length } }
    },
    gridConfig: {
      treeConfig: { rowField: 'ID', children: 'children' }
    },
    searchItems: [
      {
        field: 'name',
        title: '公司名称',
        span: 6,
        itemRender: { name: 'VxeInput', props: { placeholder: '公司名称', clearable: true } }
      }
    ],
    columns: [
      { field: 'name', title: '公司名称', minWidth: 220, treeNode: true },
      { field: 'code', title: '公司编码', minWidth: 140 },
      { field: 'leader', title: '负责人', minWidth: 100 },
      { field: 'phone', title: '联系电话', minWidth: 130 },
      { field: 'sort', title: '排序', width: 80 },
      { field: 'status', title: '状态', width: 90, slots: { default: 'status' } },
      { field: 'remarks', title: '备注', minWidth: 120 },
      { title: '操作', width: 240, fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  // 新增/编辑表单
  const formVisible = ref(false)
  const titleForm = ref('新增公司')
  const dialogType = ref('add')
  const formRef = ref(null)
  const form = ref({
    ID: 0,
    name: '',
    code: '',
    parentId: 0,
    sort: 0,
    leader: '',
    phone: '',
    email: '',
    address: '',
    remarks: '',
    status: 1
  })
  const rules = ref({
    name: [{ required: true, message: '请输入公司名称', trigger: 'blur' }],
    code: [{ required: true, message: '请输入公司编码', trigger: 'blur' }],
    parentId: [{ required: true, message: '请选择父级公司', trigger: 'blur' }]
  })

  const parentOptions = ref([{ ID: 0, name: '根公司' }])
  // 递归构建父级选项，编辑时禁用自身及子孙
  const buildOptions = (tree, options, disabled) => {
    tree &&
      tree.forEach((item) => {
        const selfDisabled = disabled || item.ID === form.value.ID
        const option = { ID: item.ID, name: item.name, disabled: selfDisabled }
        if (item.children && item.children.length) {
          option.children = []
          buildOptions(item.children, option.children, selfDisabled)
        }
        options.push(option)
      })
  }
  const setParentOptions = () => {
    parentOptions.value = [{ ID: 0, name: '根公司' }]
    buildOptions(companyTree.value, parentOptions.value, false)
  }

  const initForm = () => {
    formRef.value && formRef.value.resetFields()
    form.value = {
      ID: 0,
      name: '',
      code: '',
      parentId: 0,
      sort: 0,
      leader: '',
      phone: '',
      email: '',
      address: '',
      remarks: '',
      status: 1
    }
  }
  const closeForm = () => {
    initForm()
    formVisible.value = false
  }
  const addCompany = (parentId) => {
    initForm()
    dialogType.value = 'add'
    titleForm.value = '新增公司'
    form.value.parentId = parentId
    setParentOptions()
    formVisible.value = true
  }
  const editCompany = (row) => {
    initForm()
    dialogType.value = 'edit'
    titleForm.value = '编辑公司'
    form.value = {
      ID: row.ID,
      name: row.name,
      code: row.code,
      parentId: row.parentId,
      sort: row.sort,
      leader: row.leader,
      phone: row.phone,
      email: row.email,
      address: row.address,
      remarks: row.remarks,
      status: row.status
    }
    setParentOptions()
    formRef.value && formRef.value.clearValidate()
    formVisible.value = true
  }
  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      const req = { ...form.value }
      if (dialogType.value === 'add') {
        const res = await createCompany(req)
        if (res.code === 0) {
          ElMessage({ type: 'success', message: '创建成功' })
          await refresh()
          closeForm()
        }
      } else {
        const res = await updateCompany(req)
        if (res.code === 0) {
          ElMessage({ type: 'success', message: '更新成功' })
          await refresh()
          closeForm()
        }
      }
    })
  }
  const deleteCompanyFunc = (row) => {
    ElMessageBox.confirm(`此操作将删除公司「${row.name}」, 是否继续?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        const res = await deleteCompany(row.ID)
        if (res.code === 0) {
          ElMessage({ type: 'success', message: '删除成功!' })
          refresh()
        }
      })
      .catch(() => {})
  }
</script>

<style lang="scss" scoped>
  .company {
    .el-input-number {
      margin-left: 0;
    }
  }
</style>
