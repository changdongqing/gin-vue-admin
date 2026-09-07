<template>
  <el-drawer v-model="visible" size="720px" :title="`类详情 - ${detail.localName || ''}`">
    <el-descriptions :column="1" border size="small">
      <el-descriptions-item label="类 IRI">{{ detail.classIri }}</el-descriptions-item>
      <el-descriptions-item label="本地名">{{ detail.localName }}</el-descriptions-item>
      <el-descriptions-item label="中文名">{{ detail.labelCn || '-' }}</el-descriptions-item>
      <el-descriptions-item label="英文名">{{ detail.label || '-' }}</el-descriptions-item>
      <el-descriptions-item label="描述">{{ detail.description || '-' }}</el-descriptions-item>
      <el-descriptions-item label="模板溯源">
        <el-tag v-if="detail.templateCode" type="primary">模板：{{ detail.templateCode }}</el-tag>
        <el-tag v-else type="info">空白类</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="分类编码">{{ detail.classificationCode || '-' }}</el-descriptions-item>
      <el-descriptions-item label="可实例化">{{ detail.isInstantiable === 1 ? '是' : '否' }}</el-descriptions-item>
    </el-descriptions>

    <el-tabs v-model="activeTab" class="mt-4">
      <el-tab-pane :label="`数据属性 (${detail.datatypeProperties?.length || 0})`" name="datatype">
        <div class="mb-2">
          <el-button type="primary" icon="plus" @click="$emit('select-templates', 'datatype')">从属性库选择</el-button>
          <el-button type="primary" icon="plus" plain @click="$emit('add-property', 'datatype')">新增数据属性</el-button>
        </div>
        <el-table :data="detail.datatypeProperties || []" border max-height="320">
          <el-table-column prop="propertyIri" label="属性 IRI" min-width="220" show-overflow-tooltip />
          <el-table-column prop="localName" label="本地名" width="130" />
          <el-table-column prop="label" label="显示名" width="120" />
          <el-table-column prop="typeOrRange" label="类型" width="130" />
          <el-table-column prop="templateCode" label="模板" width="120" />
          <el-table-column label="操作" width="150" fixed="right" align="center">
            <template #default="{ row }">
              <el-button icon="edit" type="primary" link @click="$emit('edit-property', 'datatype', row)">编辑</el-button>
              <el-popconfirm title="确定删除该数据属性？" @confirm="deleteProperty('datatype', row)">
                <template #reference>
                  <el-button icon="delete" type="primary" link>删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane :label="`对象属性 (${detail.objectProperties?.length || 0})`" name="object">
        <div class="mb-2">
          <el-button type="primary" icon="plus" @click="$emit('select-templates', 'object')">从属性库选择</el-button>
          <el-button type="primary" icon="plus" plain @click="$emit('add-property', 'object')">新增对象属性</el-button>
        </div>
        <el-table :data="detail.objectProperties || []" border max-height="320">
          <el-table-column prop="propertyIri" label="属性 IRI" min-width="220" show-overflow-tooltip />
          <el-table-column prop="localName" label="本地名" width="130" />
          <el-table-column prop="label" label="显示名" width="110" />
          <el-table-column prop="typeOrRange" label="Range" width="150" />
          <el-table-column prop="templateCode" label="模板" width="120" />
          <el-table-column label="操作" width="150" fixed="right" align="center">
            <template #default="{ row }">
              <el-button icon="edit" type="primary" link @click="$emit('edit-property', 'object', row)">编辑</el-button>
              <el-popconfirm title="确定删除该对象属性？" @confirm="deleteProperty('object', row)">
                <template #reference>
                  <el-button icon="delete" type="primary" link>删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="层级" name="hierarchy">
        <div class="text-sm mb-1 font-bold">直接父类</div>
        <el-table v-if="detail.parentClasses?.length" :data="detail.parentClasses" border size="small">
          <el-table-column prop="classIri" label="类 IRI" min-width="240" show-overflow-tooltip />
          <el-table-column prop="label" label="名称" width="180" />
        </el-table>
        <el-empty v-else description="无父类（顶层类）" :image-size="48" />
        <div class="text-sm mb-1 mt-4 font-bold">直接子类</div>
        <el-table v-if="detail.childClasses?.length" :data="detail.childClasses" border size="small">
          <el-table-column prop="classIri" label="类 IRI" min-width="240" show-overflow-tooltip />
          <el-table-column prop="label" label="名称" width="180" />
        </el-table>
        <el-empty v-else description="无子类" :image-size="48" />
      </el-tab-pane>
    </el-tabs>
  </el-drawer>
</template>

<script setup>
  import { getModelClassDetail } from '@/api/ontology/modelClass'
  import { deleteModelDatatypeProperty, deleteModelObjectProperty } from '@/api/ontology/modelProperty'
  import { ref, reactive } from 'vue'
  import { ElMessage } from 'element-plus'

  defineOptions({ name: 'ClassDetail' })

  const emit = defineEmits(['select-templates', 'add-property', 'edit-property', 'changed'])

  const visible = ref(false)
  const activeTab = ref('datatype')
  const detail = reactive({
    localName: '',
    classIri: '',
    label: '',
    labelCn: '',
    description: '',
    templateCode: '',
    classificationCode: '',
    isInstantiable: 0,
    datatypeProperties: [],
    objectProperties: [],
    parentClasses: [],
    childClasses: []
  })
  let currentId = 0

  const open = (row) => {
    currentId = row.ID
    activeTab.value = 'datatype'
    visible.value = true
    load()
  }
  const load = async () => {
    const res = await getModelClassDetail({ ID: currentId })
    if (res.code === 0) {
      Object.assign(detail, res.data)
    }
  }
  defineExpose({ open, reload: load })

  const deleteProperty = async (kind, row) => {
    const res =
      kind === 'datatype'
        ? await deleteModelDatatypeProperty({ ID: row.ID })
        : await deleteModelObjectProperty({ ID: row.ID })
    if (res.code === 0) {
      ElMessage.success('删除成功')
      load()
      emit('changed')
    }
  }
</script>
