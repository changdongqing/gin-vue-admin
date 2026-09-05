<template>
  <div class="department">
    <warning-bar title="注：部门须挂靠公司，父级部门只能选同公司部门；用户的主属部门在「用户管理」中设置" />
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="addDepartment(0)">新增部门</el-button>
      </template>

      <template #companyName="{ row }">
        {{ companyNameMap[row.companyId] || row.companyId }}
      </template>

      <template #status="{ row }">
        <el-tag :type="row.status === 1 ? 'success' : 'danger'">
          {{ row.status === 1 ? '启用' : '停用' }}
        </el-tag>
      </template>

      <template #operate="{ row }">
        <el-button icon="plus" type="primary" link @click="addDepartment(row.ID, row.companyId)">新增子级</el-button>
        <el-button icon="edit" type="primary" link @click="editDepartment(row)">编辑</el-button>
        <el-button icon="delete" type="primary" link @click="deleteDepartmentFunc(row)">删除</el-button>
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
        <el-form-item label="所属公司" prop="companyId">
          <el-tree-select
            v-model="form.companyId"
            :data="companyTreeData"
            node-key="ID"
            :props="{ label: 'name', children: 'children' }"
            check-strictly
            :render-after-expand="false"
            default-expand-all
            style="width: 100%"
            placeholder="请选择所属公司"
          />
        </el-form-item>
        <el-form-item label="父级部门" prop="parentId">
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
            placeholder="不选则为该公司一级部门"
          />
        </el-form-item>
        <el-form-item label="部门名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入部门名称" />
        </el-form-item>
        <el-form-item label="部门编码" prop="code">
          <el-input v-model="form.code" placeholder="请输入部门编码(唯一)" />
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
    getDepartmentList,
    createDepartment,
    updateDepartment,
    deleteDepartment
  } from '@/api/department'
  import { getCompanyList } from '@/api/company'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useAppStore } from '@/pinia'
  import GvaGrid, { useGvaGrid } from '@/components/gvaGrid'

  defineOptions({
    name: 'Department'
  })

  const appStore = useAppStore()

  // 公司数据（树 + 平铺映射）
  const companyTreeData = ref([])
  const companyNameMap = ref({})
  const flattenCompany = (tree, level, out) => {
    tree &&
      tree.forEach((item) => {
        companyNameMap.value[item.ID] = item.name
        out.push({ label: `${'　'.repeat(level)}${item.name}`, value: item.ID })
        if (item.children && item.children.length) {
          flattenCompany(item.children, level + 1, out)
        }
      })
  }
  const companySelectOptions = ref([])
  const loadCompanies = async () => {
    const res = await getCompanyList()
    companyTreeData.value = res.data.list || []
    companyNameMap.value = {}
    companySelectOptions.value = [{ label: '全部公司', value: 0 }]
    flattenCompany(companyTreeData.value, 0, companySelectOptions.value)
  }
  loadCompanies()

  // 部门树
  const departmentTree = ref([])
  const { gridRef, gridOptions, gridEvents, refresh } = useGvaGrid({
    id: 'superAdmin-department',
    pager: false,
    defaultSort: null,
    api: async (data) => {
      const res = await getDepartmentList({
        companyId: (data && data.companyId) || 0,
        name: (data && data.name) || ''
      })
      departmentTree.value = res.data.list || []
      return { ...res, data: { list: departmentTree.value, total: departmentTree.value.length } }
    },
    gridConfig: {
      treeConfig: { rowField: 'ID', children: 'children' }
    },
    searchItems: [
      {
        field: 'companyId',
        title: '所属公司',
        span: 6,
        itemRender: {
          name: 'VxeSelect',
          props: { options: companySelectOptions, placeholder: '全部公司', clearable: true }
        }
      },
      {
        field: 'name',
        title: '部门名称',
        span: 6,
        itemRender: { name: 'VxeInput', props: { placeholder: '部门名称', clearable: true } }
      }
    ],
    columns: [
      { field: 'name', title: '部门名称', minWidth: 200, treeNode: true },
      { field: 'code', title: '部门编码', minWidth: 140 },
      { field: 'companyId', title: '所属公司', minWidth: 140, slots: { default: 'companyName' } },
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
  const titleForm = ref('新增部门')
  const dialogType = ref('add')
  const formRef = ref(null)
  const form = ref({
    ID: 0,
    companyId: null,
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
    companyId: [{ required: true, message: '请选择所属公司', trigger: 'blur' }],
    name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }],
    code: [{ required: true, message: '请输入部门编码', trigger: 'blur' }]
  })

  // 父级部门选项：仅同公司部门
  const parentOptions = ref([{ ID: 0, name: '一级部门' }])
  const buildDeptOptions = (tree, options, disabled) => {
    tree &&
      tree.forEach((item) => {
        if (item.companyId !== form.value.companyId) return
        const selfDisabled = disabled || item.ID === form.value.ID
        const option = { ID: item.ID, name: item.name, disabled: selfDisabled }
        if (item.children && item.children.length) {
          option.children = []
          buildDeptOptions(item.children, option.children, selfDisabled)
        }
        options.push(option)
      })
  }
  const setParentOptions = () => {
    parentOptions.value = [{ ID: 0, name: '一级部门' }]
    buildDeptOptions(departmentTree.value, parentOptions.value, false)
  }

  const initForm = () => {
    formRef.value && formRef.value.resetFields()
    form.value = {
      ID: 0,
      companyId: null,
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
  const addDepartment = (parentId, companyId) => {
    initForm()
    dialogType.value = 'add'
    titleForm.value = '新增部门'
    form.value.parentId = parentId
    form.value.companyId = companyId || null
    setParentOptions()
    formVisible.value = true
  }
  const editDepartment = (row) => {
    initForm()
    dialogType.value = 'edit'
    titleForm.value = '编辑部门'
    form.value = {
      ID: row.ID,
      companyId: row.companyId,
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
      const req = { ...form.value, parentId: form.value.parentId || 0 }
      if (dialogType.value === 'add') {
        const res = await createDepartment(req)
        if (res.code === 0) {
          ElMessage({ type: 'success', message: '创建成功' })
          await refresh()
          closeForm()
        }
      } else {
        const res = await updateDepartment(req)
        if (res.code === 0) {
          ElMessage({ type: 'success', message: '更新成功' })
          await refresh()
          closeForm()
        }
      }
    })
  }
  const deleteDepartmentFunc = (row) => {
    ElMessageBox.confirm(`此操作将删除部门「${row.name}」, 是否继续?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        const res = await deleteDepartment(row.ID)
        if (res.code === 0) {
          ElMessage({ type: 'success', message: '删除成功!' })
          refresh()
        }
      })
      .catch(() => {})
  }
</script>

<style lang="scss" scoped>
  .department {
    .el-input-number {
      margin-left: 0;
    }
  }
</style>
