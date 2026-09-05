<template>
  <div>
    <warning-bar title="注：右上角头像下拉可切换角色" />
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="addUser">新增用户</el-button>
      </template>

      <template #headerImg="{ row }">
        <CustomPic style="margin-top: 8px" :pic-src="row.headerImg" />
      </template>

      <template #authorities="{ row }">
        <el-cascader
          v-model="row.authorityIds"
          :options="authOptions"
          :show-all-levels="false"
          collapse-tags
          :props="{
            multiple: true,
            checkStrictly: true,
            label: 'authorityName',
            value: 'authorityId',
            disabled: 'disabled',
            emitPath: false
          }"
          :clearable="false"
          @visible-change="
            (flag) => {
              changeAuthority(row, flag, 0)
            }
          "
          @remove-tag="
            (removeAuth) => {
              changeAuthority(row, false, removeAuth)
            }
          "
        />
      </template>

      <template #enable="{ row }">
        <el-switch
          v-model="row.enable"
          inline-prompt
          :active-value="1"
          :inactive-value="2"
          @change="
            () => {
              switchEnable(row)
            }
          "
        />
      </template>

      <template #operate="{ row }">
        <el-button type="primary" link icon="delete" @click="deleteRow(row)">删除</el-button>
        <el-button type="primary" link icon="edit" @click="openEdit(row)">编辑</el-button>
        <el-button type="primary" link icon="magic-stick" @click="resetPasswordFunc(row)">
          重置密码
        </el-button>
      </template>
    </GvaGrid>

    <!-- 重置密码对话框 -->
    <el-dialog
      v-model="resetPwdDialog"
      title="重置密码"
      width="500px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
    >
      <el-form :model="resetPwdInfo" ref="resetPwdForm" label-width="100px">
        <el-form-item label="用户账号">
          <el-input v-model="resetPwdInfo.userName" disabled />
        </el-form-item>
        <el-form-item label="用户昵称">
          <el-input v-model="resetPwdInfo.nickName" disabled />
        </el-form-item>
        <el-form-item label="新密码">
          <div class="flex w-full">
            <el-input class="flex-1" v-model="resetPwdInfo.password" placeholder="请输入新密码" show-password />
            <el-button type="primary" @click="generateRandomPassword" style="margin-left: 10px">
              生成随机密码
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeResetPwdDialog">取 消</el-button>
          <el-button type="primary" @click="confirmResetPassword">确 定</el-button>
        </div>
      </template>
    </el-dialog>

    <el-drawer v-model="addUserDialog" :size="appStore.drawerSize" :show-close="false">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">用户</span>
          <div>
            <el-button @click="closeAddUserDialog">取 消</el-button>
            <el-button type="primary" @click="enterAddUserDialog">确 定</el-button>
          </div>
        </div>
      </template>

      <el-form ref="userForm" :rules="rules" :model="userInfo" label-width="80px">
        <el-form-item v-if="dialogFlag === 'add'" label="用户名" prop="userName">
          <el-input v-model="userInfo.userName" />
        </el-form-item>
        <el-form-item v-if="dialogFlag === 'add'" label="密码" prop="password">
          <el-input v-model="userInfo.password" />
        </el-form-item>
        <el-form-item label="昵称" prop="nickName">
          <el-input v-model="userInfo.nickName" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="userInfo.phone" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userInfo.email" />
        </el-form-item>
        <el-form-item label="用户角色" prop="authorityId">
          <el-cascader
            v-model="userInfo.authorityIds"
            style="width: 100%"
            :options="authOptions"
            :show-all-levels="false"
            :props="{
              multiple: true,
              checkStrictly: true,
              label: 'authorityName',
              value: 'authorityId',
              disabled: 'disabled',
              emitPath: false
            }"
            :clearable="false"
          />
        </el-form-item>
        <el-form-item label="启用" prop="disabled">
          <el-switch v-model="userInfo.enable" inline-prompt :active-value="1" :inactive-value="2" />
        </el-form-item>
        <el-form-item label="头像" label-width="80px">
          <SelectImage v-model="userInfo.headerImg" />
        </el-form-item>
      </el-form>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    getUserList,
    setUserAuthorities,
    register,
    deleteUser,
    setUserInfo,
    resetPassword
  } from '@/api/user'

  import { getAuthorityList } from '@/api/authority'
  import GvaGrid, { useGvaGrid, useGvaGridDelete } from '@/components/gvaGrid'
  import CustomPic from '@/components/customPic/index.vue'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { nextTick, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import SelectImage from '@/components/selectImage/selectImage.vue'
  import { useAppStore } from '@/pinia'

  defineOptions({
    name: 'User'
  })

  const appStore = useAppStore()

  const { gridRef, gridOptions, gridEvents, refresh } = useGvaGrid({
    id: 'superAdmin-user',
    api: getUserList,
    defaultSort: { field: 'ID', order: 'desc' },
    searchItems: [
      { field: 'username', title: '用户名', span: 6, itemRender: { name: 'VxeInput', props: { placeholder: '用户名', clearable: true } } },
      { field: 'nickname', title: '昵称', span: 6, itemRender: { name: 'VxeInput', props: { placeholder: '昵称', clearable: true } } },
      { field: 'phone', title: '手机号', span: 6, itemRender: { name: 'VxeInput', props: { placeholder: '手机号', clearable: true } } },
      { field: 'email', title: '邮箱', span: 6, itemRender: { name: 'VxeInput', props: { placeholder: '邮箱', clearable: true } } }
    ],
    columns: [
      { field: 'headerImg', title: '头像', width: 80, slots: { default: 'headerImg' } },
      { field: 'ID', title: 'ID', width: 80, sortable: true },
      { field: 'userName', title: '用户名', minWidth: 150 },
      { field: 'nickName', title: '昵称', minWidth: 150 },
      { field: 'phone', title: '手机号', minWidth: 180 },
      { field: 'email', title: '邮箱', minWidth: 180 },
      { field: 'authorities', title: '用户角色', minWidth: 200, slots: { default: 'authorities' } },
      { field: 'enable', title: '启用', minWidth: 100, slots: { default: 'enable' } },
      { title: '操作', fixed: 'right', slots: { default: 'operate' } }
    ],
    // 数据后处理：行内角色级联需要 authorityIds 数组（原 watch(tableData) 逻辑）
    afterQuery: (res) => {
      const list = (res.data && res.data.list) || []
      list.forEach((user) => {
        user.authorityIds =
          user.authorities && user.authorities.map((i) => i.authorityId)
      })
    }
  })

  const { deleteRow } = useGvaGridDelete(gridRef, {
    delete: (rows) => deleteUser({ id: rows[0].ID }),
    confirmText: '确定要删除吗?'
  })

  // 角色选项（初始化相关）
  const setAuthorityOptions = (AuthorityData, optionsData) => {
    AuthorityData &&
      AuthorityData.forEach((item) => {
        if (item.children && item.children.length) {
          const option = {
            authorityId: item.authorityId,
            authorityName: item.authorityName,
            children: []
          }
          setAuthorityOptions(item.children, option.children)
          optionsData.push(option)
        } else {
          const option = {
            authorityId: item.authorityId,
            authorityName: item.authorityName
          }
          optionsData.push(option)
        }
      })
  }

  const authOptions = ref([])
  const setOptions = (authData) => {
    authOptions.value = []
    setAuthorityOptions(authData, authOptions.value)
  }

  const initPage = async () => {
    const res = await getAuthorityList()
    setOptions(res.data)
  }

  initPage()

  // 重置密码对话框相关
  const resetPwdDialog = ref(false)
  const resetPwdForm = ref(null)
  const resetPwdInfo = ref({
    ID: '',
    userName: '',
    nickName: '',
    password: ''
  })

  // 生成随机密码
  const generateRandomPassword = () => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*'
    let password = ''
    for (let i = 0; i < 12; i++) {
      password += chars.charAt(Math.floor(Math.random() * chars.length))
    }
    resetPwdInfo.value.password = password
    // 复制到剪贴板
    navigator.clipboard.writeText(password).then(() => {
      ElMessage({
        type: 'success',
        message: '密码已复制到剪贴板'
      })
    }).catch(() => {
      ElMessage({
        type: 'error',
        message: '复制失败，请手动复制'
      })
    })
  }

  // 打开重置密码对话框
  const resetPasswordFunc = (row) => {
    resetPwdInfo.value.ID = row.ID
    resetPwdInfo.value.userName = row.userName
    resetPwdInfo.value.nickName = row.nickName
    resetPwdInfo.value.password = ''
    resetPwdDialog.value = true
  }

  // 确认重置密码
  const confirmResetPassword = async () => {
    if (!resetPwdInfo.value.password) {
      ElMessage({
        type: 'warning',
        message: '请输入或生成密码'
      })
      return
    }

    const res = await resetPassword({
      ID: resetPwdInfo.value.ID,
      password: resetPwdInfo.value.password
    })

    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: res.msg || '密码重置成功'
      })
      resetPwdDialog.value = false
    } else {
      ElMessage({
        type: 'error',
        message: res.msg || '密码重置失败'
      })
    }
  }

  // 关闭重置密码对话框
  const closeResetPwdDialog = () => {
    resetPwdInfo.value.password = ''
    resetPwdDialog.value = false
  }

  // 弹窗相关
  const userInfo = ref({
    userName: '',
    password: '',
    nickName: '',
    headerImg: '',
    authorityId: '',
    authorityIds: [],
    enable: 1
  })

  const rules = ref({
    userName: [
      { required: true, message: '请输入用户名', trigger: 'blur' },
      { min: 5, message: '最低5位字符', trigger: 'blur' }
    ],
    password: [
      { required: true, message: '请输入用户密码', trigger: 'blur' },
      { min: 6, message: '最低6位字符', trigger: 'blur' }
    ],
    nickName: [{ required: true, message: '请输入用户昵称', trigger: 'blur' }],
    phone: [
      {
        pattern: /^1([38][0-9]|4[014-9]|[59][0-35-9]|6[2567]|7[0-8])\d{8}$/,
        message: '请输入合法手机号',
        trigger: 'blur'
      }
    ],
    email: [
      {
        pattern: /^([0-9A-Za-z\-_.]+)@([0-9a-z]+\.[a-z]{2,3}(\.[a-z]{2})?)$/g,
        message: '请输入正确的邮箱',
        trigger: 'blur'
      }
    ],
    authorityId: [{ required: true, message: '请选择用户角色', trigger: 'blur' }]
  })
  const userForm = ref(null)
  const enterAddUserDialog = async () => {
    userInfo.value.authorityId = userInfo.value.authorityIds[0]
    userForm.value.validate(async (valid) => {
      if (valid) {
        const req = {
          ...userInfo.value
        }
        if (dialogFlag.value === 'add') {
          const res = await register(req)
          if (res.code === 0) {
            ElMessage({ type: 'success', message: '创建成功' })
            await refresh()
            closeAddUserDialog()
          }
        }
        if (dialogFlag.value === 'edit') {
          const res = await setUserInfo(req)
          if (res.code === 0) {
            ElMessage({ type: 'success', message: '编辑成功' })
            await refresh()
            closeAddUserDialog()
          }
        }
      }
    })
  }

  const addUserDialog = ref(false)
  const closeAddUserDialog = () => {
    userForm.value.resetFields()
    userInfo.value.headerImg = ''
    userInfo.value.authorityIds = []
    addUserDialog.value = false
  }

  const dialogFlag = ref('add')

  const addUser = () => {
    dialogFlag.value = 'add'
    addUserDialog.value = true
  }

  const tempAuth = {}
  const changeAuthority = async (row, flag, removeAuth) => {
    if (flag) {
      if (!removeAuth) {
        tempAuth[row.ID] = [...row.authorityIds]
      }
      return
    }
    await nextTick()
    const res = await setUserAuthorities({
      ID: row.ID,
      authorityIds: row.authorityIds
    })
    if (res.code === 0) {
      ElMessage({ type: 'success', message: '角色设置成功' })
    } else {
      if (!removeAuth) {
        row.authorityIds = [...tempAuth[row.ID]]
        delete tempAuth[row.ID]
      } else {
        row.authorityIds = [removeAuth, ...row.authorityIds]
      }
    }
  }

  const openEdit = (row) => {
    dialogFlag.value = 'edit'
    userInfo.value = JSON.parse(JSON.stringify(row))
    addUserDialog.value = true
  }

  const switchEnable = async (row) => {
    userInfo.value = JSON.parse(JSON.stringify(row))
    await nextTick()
    const req = {
      ...userInfo.value
    }
    const res = await setUserInfo(req)
    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: `${req.enable === 2 ? '禁用' : '启用'}成功`
      })
      await refresh()
      userInfo.value.headerImg = ''
      userInfo.value.authorityIds = []
    }
  }
</script>

<style lang="scss">
  .header-img-box {
    @apply w-52 h-52 border border-solid border-gray-300 rounded-xl flex justify-center items-center cursor-pointer;
  }
</style>
