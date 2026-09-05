<template>
  <div class="data-scope">
    <warning-bar
      title="注：数据范围决定角色在受控业务数据（如用户管理列表）中可见的行；多角色用户取并集（就宽）；勾选部门含其子级，勾选公司含其全部部门"
    />
    <el-form label-width="90px">
      <el-form-item label="数据范围">
        <el-radio-group v-model="dataScope" @change="onChange">
          <el-radio v-for="item in scopeOptions" :key="item.value" :value="item.value">
            {{ item.label }}
          </el-radio>
        </el-radio-group>
      </el-form-item>
    </el-form>

    <template v-if="dataScope === 2">
      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane label="按公司" name="company">
          <div class="tree-content">
            <el-tree
              ref="companyTreeRef"
              :data="companyTree"
              node-key="ID"
              show-checkbox
              check-strictly
              default-expand-all
              :props="{ label: 'name', children: 'children' }"
              @check="onChange"
            >
              <template #default="{ data }">
                <span>{{ data.name }}</span>
                <span class="tree-tip">（含全部部门）</span>
              </template>
            </el-tree>
          </div>
        </el-tab-pane>
        <el-tab-pane label="按部门" name="department">
          <div class="tree-content">
            <el-tree
              ref="departmentTreeRef"
              :data="departmentTree"
              node-key="ID"
              show-checkbox
              default-expand-all
              :props="{ label: 'name', children: 'children' }"
              @check="onChange"
            >
              <template #default="{ data }">
                <span>{{ data.name }}</span>
                <span class="tree-tip">（含子级）</span>
              </template>
            </el-tree>
          </div>
        </el-tab-pane>
      </el-tabs>
    </template>

    <div class="mt-4 text-right">
      <el-button type="primary" @click="enterAndNext">保 存</el-button>
    </div>
  </div>
</template>

<script setup>
  import { getDataScope, setDataScope } from '@/api/authority'
  import { getCompanyList } from '@/api/company'
  import { getDepartmentList } from '@/api/department'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { nextTick, onMounted, ref } from 'vue'
  import { ElMessage } from 'element-plus'

  defineOptions({
    name: 'DataScope'
  })

  const props = defineProps({
    row: {
      default: function () {
        return {}
      },
      type: Object
    }
  })

  // 与后端 service/system/sys_authority_data_scope.go 常量对齐
  const scopeOptions = [
    { value: 1, label: '全部数据' },
    { value: 2, label: '自定义数据' },
    { value: 3, label: '本公司数据' },
    { value: 4, label: '本部门及以下数据' },
    { value: 5, label: '本部门数据' },
    { value: 6, label: '仅本人数据' }
  ]

  const dataScope = ref(1)
  const activeTab = ref('company')
  const companyTree = ref([])
  const departmentTree = ref([])
  const companyTreeRef = ref(null)
  const departmentTreeRef = ref(null)
  const needConfirm = ref(false)

  const loadTrees = async () => {
    const [companyRes, departmentRes] = await Promise.all([
      getCompanyList(),
      getDepartmentList()
    ])
    companyTree.value = companyRes.data.list || []
    departmentTree.value = departmentRes.data.list || []
  }

  const loadScope = async () => {
    const res = await getDataScope(props.row.authorityId)
    if (res.code === 0 && res.data) {
      dataScope.value = res.data.dataScope || 1
      await nextTick()
      if (dataScope.value === 2) {
        companyTreeRef.value &&
          companyTreeRef.value.setCheckedKeys(res.data.companyIds || [], false)
        departmentTreeRef.value &&
          departmentTreeRef.value.setCheckedKeys(res.data.departmentIds || [], false)
      }
    }
  }

  onMounted(async () => {
    await loadTrees()
    await nextTick()
    await loadScope()
  })

  const onChange = () => {
    needConfirm.value = true
  }

  // 保存（同时暴露给外层 Tab 切换拦截）
  const enterAndNext = async () => {
    const req = {
      authorityId: props.row.authorityId,
      dataScope: dataScope.value
    }
    if (dataScope.value === 2) {
      req.companyIds = companyTreeRef.value
        ? companyTreeRef.value.getCheckedKeys(false)
        : []
      req.departmentIds = departmentTreeRef.value
        ? departmentTreeRef.value.getCheckedKeys(false)
        : []
    }
    const res = await setDataScope(req)
    if (res.code === 0) {
      ElMessage({ type: 'success', message: '数据范围设置成功' })
      needConfirm.value = false
    }
  }

  defineExpose({
    enterAndNext,
    needConfirm
  })
</script>

<style lang="scss" scoped>
  .data-scope {
    .tree-content {
      margin-top: 10px;
      height: calc(100vh - 420px);
      overflow: auto;
    }
    .tree-tip {
      color: var(--el-text-color-secondary);
      font-size: 12px;
      margin-left: 4px;
    }
  }
</style>
