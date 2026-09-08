<template>
  <div class="post">
    <warning-bar title="岗位为全局共享字典：一个岗位可分配多名员工，一名员工可挂多个岗位；部门归属在「用户管理」中设置" />
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="openForm()">新增岗位</el-button>
      </template>

      <template #status="{ row }">
        <el-switch
          v-model="row.status"
          inline-prompt
          :active-value="1"
          :inactive-value="2"
          @change="switchStatus(row)"
        />
      </template>

      <template #operate="{ row }">
        <el-tooltip v-if="row.status === 2" content="岗位已停用" placement="top">
          <span>
            <el-button icon="user" type="primary" link disabled>分配员工</el-button>
          </span>
        </el-tooltip>
        <el-button v-else icon="user" type="primary" link @click="openAssign(row)">分配员工</el-button>
        <el-button icon="edit" type="primary" link @click="openForm(row)">编辑</el-button>
        <el-button icon="delete" type="primary" link @click="deleteRow(row)">删除</el-button>
      </template>
    </GvaGrid>

    <!-- 新增/编辑岗位 -->
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
        <el-form-item label="岗位编码" prop="postCode">
          <el-input v-model="form.postCode" placeholder="请输入岗位编码(唯一，如 PM)" />
        </el-form-item>
        <el-form-item label="岗位名称" prop="postName">
          <el-input v-model="form.postName" placeholder="请输入岗位名称" />
        </el-form-item>
        <el-form-item label="岗位类型" prop="postType">
          <el-select v-model="form.postType" placeholder="请选择岗位类型" style="width: 100%">
            <el-option v-for="item in postTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
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

    <!-- 分配员工（岗位 → 员工，全量覆盖） -->
    <el-drawer v-model="assignVisible" :size="appStore.drawerSize" :show-close="false">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">分配员工{{ currentPost ? ` —— ${currentPost.postName}（${currentPost.postCode}）` : '' }}</span>
          <div>
            <el-button @click="assignVisible = false">取 消</el-button>
            <el-button type="primary" @click="saveAssign">确 定</el-button>
          </div>
        </div>
      </template>
      <div class="flex gap-4">
        <div class="w-52 shrink-0">
          <div class="mb-2 text-sm text-gray-500">按部门过滤（含子部门）</div>
          <el-tree-select
            v-model="deptId"
            :data="departmentTree"
            node-key="ID"
            :props="{ label: 'name', children: 'children' }"
            check-strictly
            :render-after-expand="false"
            default-expand-all
            filterable
            clearable
            style="width: 100%"
            placeholder="全部部门"
            @change="onFilterChange"
          />
        </div>
        <div class="flex-1 min-w-0">
          <div class="flex mb-3">
            <el-input v-model="keyword" placeholder="按昵称搜索" clearable class="flex-1" @keyup.enter="onSearch" @clear="onSearch" />
            <el-button type="primary" class="ml-2" @click="onSearch">搜索</el-button>
          </div>
          <el-table
            ref="userTableRef"
            :data="userRows"
            v-loading="userLoading"
            row-key="ID"
            max-height="480"
            @selection-change="handleSelectionChange"
          >
            <el-table-column type="selection" width="45" />
            <el-table-column prop="userName" label="用户名" min-width="120" show-overflow-tooltip />
            <el-table-column prop="nickName" label="昵称" min-width="120" show-overflow-tooltip />
            <el-table-column label="部门" min-width="130" show-overflow-tooltip>
              <template #default="{ row }">
                {{ row.department && row.department.name ? row.department.name : '未分配' }}
              </template>
            </el-table-column>
          </el-table>
          <el-pagination
            class="mt-3 justify-end"
            layout="total, prev, pager, next, sizes"
            :total="userTotal"
            v-model:current-page="userPage"
            v-model:page-size="userPageSize"
            :page-sizes="[20, 50, 100]"
            @current-change="loadUsers"
            @size-change="onSearch"
          />
        </div>
      </div>
      <div class="mt-4">
        <span class="text-sm text-gray-500">已选（{{ selectedUsers.length }}）：</span>
        <el-tag
          v-for="u in selectedUsers"
          :key="u.ID"
          class="mr-2 mb-1"
          closable
          @close="removeSelected(u.ID)"
        >
          {{ u.nickName || u.userName }}
        </el-tag>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    createPost,
    updatePost,
    deletePost,
    getPostList,
    getPostUsers,
    setPostUsers
  } from '@/api/post'
  import { getUserList } from '@/api/user'
  import { getDepartmentList } from '@/api/department'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import GvaGrid, { useGvaGrid, useGvaGridDelete } from '@/components/gvaGrid'
  import { getDict } from '@/utils/dictionary'
  import { computed, nextTick, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useAppStore } from '@/pinia'

  defineOptions({
    name: 'Post'
  })

  const appStore = useAppStore()

  const { gridRef, gridOptions, gridEvents, refresh } = useGvaGrid({
    id: 'superAdmin-post',
    api: getPostList,
    defaultSort: null,
    searchItems: [
      {
        field: 'keyword',
        title: '关键字',
        span: 6,
        itemRender: { name: 'VxeInput', props: { placeholder: '岗位编码/名称', clearable: true } }
      },
      {
        field: 'postType',
        title: '岗位类型',
        span: 6,
        itemRender: { name: 'gvaDictSelect', props: { dict: 'post_type', placeholder: '全部类型', clearable: true } }
      },
      {
        field: 'status',
        title: '状态',
        span: 6,
        itemRender: {
          name: 'VxeSelect',
          props: {
            options: [
              { label: '启用', value: 1 },
              { label: '停用', value: 2 }
            ],
            placeholder: '全部状态',
            clearable: true
          }
        }
      }
    ],
    columns: [
      { field: 'postCode', title: '岗位编码', minWidth: 120 },
      { field: 'postName', title: '岗位名称', minWidth: 160 },
      {
        field: 'postType',
        title: '岗位类型',
        minWidth: 110,
        cellRender: { name: 'gvaDict', props: { dict: 'post_type' } }
      },
      { field: 'userCount', title: '关联员工数', width: 110 },
      { field: 'sort', title: '排序', width: 80 },
      { field: 'status', title: '状态', width: 90, slots: { default: 'status' } },
      { field: 'remarks', title: '备注', minWidth: 140, showOverflow: true },
      { title: '操作', width: 240, fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  const { deleteRow } = useGvaGridDelete(gridRef, {
    delete: (rows) => deletePost({ ID: rows[0].ID }),
    confirmText: '确定要删除该岗位吗?'
  })

  // 岗位类型字典选项
  const postTypeOptions = ref([])
  getDict('post_type', { depth: 1 }).then((list) => {
    postTypeOptions.value = (list || []).map((d) => ({ label: d.label, value: d.value }))
  })

  // 新增/编辑表单
  const formVisible = ref(false)
  const titleForm = ref('新增岗位')
  const dialogType = ref('add')
  const formRef = ref(null)
  const form = ref({
    ID: 0,
    postCode: '',
    postName: '',
    postType: '',
    sort: 0,
    status: 1,
    remarks: ''
  })
  const rules = ref({
    postCode: [{ required: true, message: '请输入岗位编码', trigger: 'blur' }],
    postName: [{ required: true, message: '请输入岗位名称', trigger: 'blur' }],
    postType: [{ required: true, message: '请选择岗位类型', trigger: 'change' }]
  })

  const openForm = (row) => {
    if (row) {
      dialogType.value = 'edit'
      titleForm.value = '编辑岗位'
      form.value = JSON.parse(JSON.stringify(row))
    } else {
      dialogType.value = 'add'
      titleForm.value = '新增岗位'
      form.value = {
        ID: 0,
        postCode: '',
        postName: '',
        postType: '',
        sort: 0,
        status: 1,
        remarks: ''
      }
    }
    formVisible.value = true
  }

  const closeForm = () => {
    formVisible.value = false
  }

  const submitForm = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      const isAdd = dialogType.value === 'add'
      const res = isAdd ? await createPost(form.value) : await updatePost(form.value)
      if (res.code === 0) {
        ElMessage.success(isAdd ? '创建成功' : '更新成功')
        formVisible.value = false
        await refresh()
      }
    })
  }

  const switchStatus = async (row) => {
    const res = await updatePost(row)
    if (res.code === 0) {
      ElMessage.success(row.status === 1 ? '启用成功' : '停用成功')
    } else {
      row.status = row.status === 1 ? 2 : 1
    }
  }

  // 分配员工抽屉
  const assignVisible = ref(false)
  const currentPost = ref(null)
  const departmentTree = ref([])
  const deptId = ref(0)
  const keyword = ref('')
  const userRows = ref([])
  const userTotal = ref(0)
  const userPage = ref(1)
  const userPageSize = ref(20)
  const userTableRef = ref(null)
  const userLoading = ref(false)
  const selectedMap = ref({}) // ID -> { ID, userName, nickName, departmentId, departmentName }
  const echoedIds = ref(new Set()) // 已做过已选回显勾选的用户ID（只回显一次，尊重用户手动取消）

  const selectedUsers = computed(() => Object.values(selectedMap.value))

  const loadDepartments = async () => {
    const res = await getDepartmentList()
    departmentTree.value = res.data.list || []
  }

  const openAssign = async (row) => {
    currentPost.value = row
    selectedMap.value = {}
    echoedIds.value = new Set()
    deptId.value = 0
    keyword.value = ''
    userPage.value = 1
    const res = await getPostUsers({ postId: row.ID })
    ;(res.data.list || []).forEach((u) => {
      selectedMap.value[u.ID] = u
    })
    assignVisible.value = true
    loadDepartments()
    loadUsers()
  }

  const loadUsers = async () => {
    userLoading.value = true
    try {
      const res = await getUserList({
        page: userPage.value,
        pageSize: userPageSize.value,
        nickName: keyword.value || undefined,
        departmentId: deptId.value || 0
      })
      userRows.value = res.data.list || []
      userTotal.value = res.data.total || 0
      await nextTick()
      // 回显已选：仅对该行首次出现时勾选一次，之后尊重用户手动取消
      userRows.value.forEach((u) => {
        if (selectedMap.value[u.ID] && !echoedIds.value.has(u.ID)) {
          userTableRef.value && userTableRef.value.toggleRowSelection(u, true)
          echoedIds.value.add(u.ID)
        }
      })
    } finally {
      userLoading.value = false
    }
  }

  const onSearch = () => {
    userPage.value = 1
    loadUsers()
  }

  const onFilterChange = () => {
    onSearch()
  }

  const handleSelectionChange = (rows) => {
    const pageIds = new Set(userRows.value.map((u) => String(u.ID)))
    const checkedIds = new Set(rows.map((u) => String(u.ID)))
    // 当前页勾选同步进已选集合（翻页时 el-table 清空勾选触发 [] 事件，
    // 旧页数据已不在 pageIds 中，已选集合不受影响）
    pageIds.forEach((id) => {
      if (checkedIds.has(id)) {
        const u = userRows.value.find((r) => String(r.ID) === id)
        selectedMap.value[id] = {
          ID: u.ID,
          userName: u.userName,
          nickName: u.nickName,
          departmentId: u.departmentId,
          departmentName: (u.department && u.department.name) || ''
        }
      } else {
        delete selectedMap.value[id]
      }
    })
  }

  const removeSelected = (id) => {
    delete selectedMap.value[id]
    const row = userRows.value.find((u) => String(u.ID) === String(id))
    if (row) {
      userTableRef.value && userTableRef.value.toggleRowSelection(row, false)
    }
  }

  const saveAssign = async () => {
    const userIds = selectedUsers.value.map((u) => u.ID)
    const res = await setPostUsers({ postId: currentPost.value.ID, userIds })
    if (res.code === 0) {
      ElMessage.success('分配成功')
      assignVisible.value = false
      await refresh()
    }
  }
</script>

<style scoped lang="scss"></style>
