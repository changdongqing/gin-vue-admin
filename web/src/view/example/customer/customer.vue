<template>
  <div>
    <warning-bar
      title="在资源权限中将此角色的资源权限清空 或者不包含创建者的角色 即可屏蔽此客户资源的显示"
    />
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button type="primary" icon="plus" @click="openDrawer">新增</el-button>
      </template>

      <template #createdAt="{ row }">
        <span>{{ formatDate(row.CreatedAt) }}</span>
      </template>

      <template #operate="{ row }">
        <el-button type="primary" link icon="edit" @click="updateCustomer(row)">变更</el-button>
        <el-button type="primary" link icon="delete" @click="deleteCustomerRow(row)">删除</el-button>
      </template>
    </GvaGrid>

    <el-drawer v-model="drawerFormVisible" :before-close="closeDrawer" :show-close="false">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">客户</span>
          <div>
            <el-button @click="closeDrawer">取 消</el-button>
            <el-button type="primary" @click="enterDrawer">确 定</el-button>
          </div>
        </div>
      </template>
      <el-form :inline="true" :model="form" label-width="80px">
        <el-form-item label="客户名">
          <el-input v-model="form.customerName" autocomplete="off" />
        </el-form-item>
        <el-form-item label="客户电话">
          <el-input v-model="form.customerPhoneData" autocomplete="off" />
        </el-form-item>
      </el-form>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    createExaCustomer,
    updateExaCustomer,
    deleteExaCustomer,
    getExaCustomer,
    getExaCustomerList
  } from '@/api/customer'
  import WarningBar from '@/components/warningBar/warningBar.vue'
  import { ref } from 'vue'
  import { formatDate } from '@/utils/format'
  import GvaGrid, { useGvaGrid, useGvaGridDelete } from '@/components/gvaGrid'

  defineOptions({
    name: 'Customer'
  })

  const form = ref({
    customerName: '',
    customerPhoneData: ''
  })

  const { gridRef, gridOptions, gridEvents, refresh } = useGvaGrid({
    id: 'example-customer',
    api: getExaCustomerList,
    defaultSort: null,
    checkbox: true,
    columns: [
      { field: 'CreatedAt', title: '接入日期', width: 180, cellRender: { name: 'gvaDate' } },
      { field: 'customerName', title: '姓名', width: 120 },
      { field: 'customerPhoneData', title: '电话', width: 120 },
      { field: 'sysUserId', title: '接入人ID', width: 120 },
      { title: '操作', fixed: 'right', slots: { default: 'operate' } }
    ]
  })

  const { deleteRow: deleteCustomerRow } = useGvaGridDelete(gridRef, {
    delete: (rows) => deleteExaCustomer({ ID: rows[0].ID }),
    confirmText: '确定要删除吗?',
    successText: '删除成功'
  })

  const drawerFormVisible = ref(false)
  const type = ref('')
  const updateCustomer = async (row) => {
    const res = await getExaCustomer({ ID: row.ID })
    type.value = 'update'
    if (res.code === 0) {
      form.value = res.data.customer
      drawerFormVisible.value = true
    }
  }
  const closeDrawer = () => {
    drawerFormVisible.value = false
    form.value = {
      customerName: '',
      customerPhoneData: ''
    }
  }
  const enterDrawer = async () => {
    let res
    switch (type.value) {
      case 'create':
        res = await createExaCustomer(form.value)
        break
      case 'update':
        res = await updateExaCustomer(form.value)
        break
      default:
        res = await createExaCustomer(form.value)
        break
    }

    if (res.code === 0) {
      closeDrawer()
      refresh()
    }
  }
  const openDrawer = () => {
    type.value = 'create'
    drawerFormVisible.value = true
  }
</script>

<style></style>
