<template>
  <el-dialog v-model="visible" title="添加到菜单" width="560px" :close-on-click-modal="false">
    <el-alert type="info" :closable="false" class="mb-3"
      title="将当前报表的预览页注册为系统菜单项（携带报表编码）。菜单路由需刷新/重新登录后生效" />
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="上级菜单" prop="parentId">
        <el-cascader
          v-model="form.parentId"
          :options="menuOptions"
          :props="{ value: 'id', label: 'label', children: 'children', checkStrictly: true, emitPath: false }"
          placeholder="不选则为顶级菜单"
          clearable
          class="w-full"
        />
      </el-form-item>
      <el-form-item label="菜单名称" prop="title">
        <el-input v-model="form.title" placeholder="默认为报表名" />
      </el-form-item>
      <el-form-item label="路由路径" prop="path">
        <el-input v-model="form.path" placeholder="reportpreview">
          <template #append>?reportCode={{ reportCode }}</template>
        </el-input>
      </el-form-item>
      <el-form-item label="排序" prop="sort">
        <el-input-number v-model="form.sort" :min="0" :step="1" step-strictly controls-position="right" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取 消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">创 建</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
  // 「添加到菜单」：创建指向预览页并携带 reportCode 的菜单。
  // GVA 适配：菜单 component 不支持 query → path 内嵌 query（reportpreview?reportCode=x），
  // 由 asyncRouter 注册路由时拆分（向后兼容）
  import { ref, computed } from 'vue'
  import { ElMessage } from 'element-plus'
  import { getMenuList, addBaseMenu } from '@/api/menu'

  const props = defineProps({
    reportCode: { type: String, required: true },
    reportName: { type: String, default: '' }
  })

  const visible = ref(false)
  const submitting = ref(false)
  const formRef = ref(null)
  const menuTree = ref([])
  const form = ref({})

  const rules = {
    title: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }],
    path: [{ required: true, message: '请输入路由路径', trigger: 'blur' }]
  }

  const menuOptions = computed(() => {
    const walk = (nodes) =>
      (nodes || []).map((n) => ({
        id: n.ID,
        label: n.meta?.title || n.name,
        children: n.children?.length ? walk(n.children) : undefined
      }))
    return walk(menuTree.value)
  })

  const open = async () => {
    form.value = {
      parentId: undefined,
      title: props.reportName || props.reportCode,
      path: 'reportpreview',
      sort: 1
    }
    visible.value = true
    const res = await getMenuList()
    if (res.code === 0) {
      menuTree.value = res.data || []
    }
  }

  const submit = () => {
    formRef.value.validate(async (valid) => {
      if (!valid) return
      // 前端预检：同父级下菜单名查重
      const parent = form.value.parentId || 0
      const findDup = (nodes) => {
        for (const n of nodes || []) {
          if (n.ID === parent) return (n.children || []).some((c) => c.meta?.title === form.value.title)
          const r = findDup(n.children)
          if (r) return true
        }
        return false
      }
      if (findDup(menuTree.value)) {
        ElMessage.error('同级菜单下已存在同名菜单，请更换名称')
        return
      }
      const rawPath = `${form.value.path || 'reportpreview'}?reportCode=${props.reportCode}`
      const path = parent === 0 ? `/${rawPath}` : rawPath
      submitting.value = true
      try {
        const res = await addBaseMenu({
          ParentId: parent,
          Path: path,
          Name: `reportPreview${props.reportCode}`,
          Hidden: false,
          Component: 'view/report/excel/preview/preview.vue',
          Sort: form.value.sort,
          Meta: {
            Title: form.value.title,
            Icon: 'document',
            KeepAlive: false
          }
        })
        if (res.code === 0) {
          // 授权给当前角色（GVA addMenuAuthority 为全量替换语义：先取现有列表再追加，避免覆盖丢失）
          let authorized = false
          try {
            const authorityId = userStore.userInfo?.authority?.authorityId || userStore.userInfo?.authorityId
            if (authorityId) {
              const listRes = await getMenuList()
              const authRes = await getMenuAuthority({ authorityId })
              if (listRes.code === 0 && authRes.code === 0) {
                const findMenu = (nodes) => {
                  for (const n of nodes || []) {
                    if (n.name === `reportPreview${props.reportCode}`) return n
                    const r = findMenu(n.children)
                    if (r) return r
                  }
                  return null
                }
                const created = findMenu(listRes.data || [])
                // getMenuAuthority 返回 { menus: [...] }（全量替换语义：取现有列表追加新菜单回写）
                const menus = authRes.data?.menus || authRes.data || []
                if (created && Array.isArray(menus) && !menus.some((m) => m.ID === created.ID)) {
                  menus.push(created)
                }
                const saveRes = await addMenuAuthority({ menus, authorityId })
                authorized = saveRes.code === 0
              }
            }
          } catch (e) {
            /* 授权失败降级提示，不影响菜单创建结果 */
          }
          ElMessage.success(authorized ? '菜单已创建并授权，刷新页面后可在侧边栏查看' : '菜单已创建，请在「角色管理-菜单权限」中授权后查看')
          visible.value = false
        }
      } finally {
        submitting.value = false
      }
    })
  }

  defineExpose({ open })
</script>
