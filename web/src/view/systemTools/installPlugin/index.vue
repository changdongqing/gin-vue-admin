<template>
  <div class="gva-form-box">
    <el-upload
      drag
      :action="`${getBaseUrl()}/autoCode/installPlugin`"
      :show-file-list="false"
      :on-success="handleSuccess"
      :on-error="handleSuccess"
      :headers="{'x-token': token}"
      name="plug"
    >
      <el-icon class="el-icon--upload"><upload-filled /></el-icon>
      <div class="el-upload__text">拖拽或<em>点击上传</em></div>
      <template #tip>
        <div class="el-upload__tip">请把安装包的zip拖拽至此处上传</div>
      </template>
    </el-upload>

    <!-- Plugin List Table -->
    <div style="margin-top: 20px;">
      <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
        <template #expand="{ row }">
          <div style="padding: 20px;">
            <h3>API 列表</h3>
            <el-table :data="row.apis" border>
              <el-table-column prop="path" label="路径" />
              <el-table-column prop="method" label="方法" />
              <el-table-column prop="description" label="描述" />
              <el-table-column prop="apiGroup" label="APIGROUP" />
            </el-table>
            <h3>菜单列表</h3>
            <el-table :data="row.menus" row-key="name" :tree-props="{children: 'children', hasChildren: 'hasChildren'}" border>
              <el-table-column prop="meta.title" label="标题" />
              <el-table-column prop="name" label="Name" />
              <el-table-column prop="path" label="Path" />
            </el-table>
            <h3>字典列表</h3>
            <el-table :data="row.dictionaries" border>
              <el-table-column prop="name" label="字典名" />
              <el-table-column prop="type" label="字典类型" />
              <el-table-column prop="desc" label="描述" />
            </el-table>
          </div>
        </template>

        <template #pluginType="{ row }">
          {{ typeMap[row.pluginType] || '未知类型' }}
        </template>

        <template #operate="{ row }">
          <el-button type="primary" link icon="delete" @click="deletePlugin(row)">删除</el-button>
        </template>
      </GvaGrid>
    </div>
  </div>
</template>

<script setup>
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { UploadFilled } from '@element-plus/icons-vue'
  import { getBaseUrl } from '@/utils/format'
  import { useUserStore } from "@/pinia";
  import { getPluginList, removePlugin } from '@/api/autoCode'
  import GvaGrid, { useGvaGrid, useGvaGridDelete } from '@/components/gvaGrid'

  const userStore = useUserStore()
  const token = userStore.token

  const typeMap = {
    "server": "后端插件",
    "web": "前端插件",
    "full": "全栈插件"
  }

  const { gridRef, gridOptions, gridEvents, refresh } = useGvaGrid({
    id: 'systemTools-installPlugin',
    pager: false,
    defaultSort: null,
    api: async () => {
      const res = await getPluginList()
      const list = res.data || []
      return { ...res, data: { list, total: list.length } }
    },
    columns: [
      { type: 'expand', width: 50, slots: { content: 'expand' } },
      { field: 'pluginName', title: '插件名称', minWidth: 180 },
      { field: 'pluginType', title: '插件类型', minWidth: 120, slots: { default: 'pluginType' } },
      { title: '操作', width: 120, slots: { default: 'operate' } }
    ]
  })

  const { deleteRow: deletePlugin } = useGvaGridDelete(gridRef, {
    delete: (rows) => removePlugin({ pluginName: rows[0].pluginName, pluginType: rows[0].pluginType }),
    confirmText: '此操作将永久删除该插件及其关联的API、菜单和字典数据, 是否继续?',
    successText: '删除成功'
  })

  const handleSuccess = (res) => {
    if (res.code === 0) {
      let msg = ``
      res.data &&
        res.data.forEach((item, index) => {
          msg += `${index + 1}.${item.msg}\n`
        })
      alert(msg)
      refresh() // Refresh list on success
    } else {
      ElMessage.error(res.msg)
    }
  }
</script>
