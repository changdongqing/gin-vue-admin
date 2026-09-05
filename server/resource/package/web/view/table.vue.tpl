{{- $global := . }}
{{- $templateID := printf "%s_%s" .Package .StructName }}
{{- if .IsAdd }}

// 请在 useGvaGrid 的 searchItems 中增加如下代码
{{- range .Fields}}
    {{- if .FieldSearchType}}
{{ GenerateGridSearchItem .}}
    {{ end }}
{{ end }}


// 表格增加如下列代码（columns 数组）

{{- range .Fields}}
    {{- if .Table}}
       {{ GenerateGridColumnSchema . }}
    {{- end }}
{{- end }}

// 复杂列的插槽片段（放入 <GvaGrid> 内）

{{- range .Fields}}
    {{- if .Table}}
       {{ GenerateGridColumnSlot . }}
    {{- end }}
{{- end }}

// 查询项插槽片段（放入 <GvaGrid> 内）

{{- range .Fields}}
    {{- if .FieldSearchType}}
       {{ GenerateGridSearchItemSlot . }}
    {{- end }}
{{- end }}

// 新增表单中增加如下代码
{{- range .Fields}}
   {{- if .Form}}
     {{ GenerateFormItem . }}
   {{- end }}
{{- end }}

// 查看抽屉中增加如下代码

{{- range .Fields}}
              {{- if .Desc }}
    {{ GenerateDescriptionItem . }}
              {{- end }}
            {{- end }}

// 字典增加如下代码
    {{- range $index, $element := .DictTypes}}
const {{ $element }}Options = ref([])
    {{- end }}

// setOptions方法中增加如下调用

{{- range $index, $element := .DictTypes }}
    {{ $element }}Options.value = await getDictFunc('{{$element}}')
{{- end }}

// 基础formData结构（变量处和关闭表单处）增加如下字段
{{- range .Fields}}
          {{- if .Form}}
            {{ GenerateDefaultFormValue . }}
          {{- end }}
        {{- end }}
// 验证规则中增加如下字段

{{- range .Fields }}
        {{- if .Form }}
            {{- if eq .Require true }}
{{.FieldJson }} : [{
    required: true,
    message: '{{ .ErrorText }}',
    trigger: ['input','blur'],
},
               {{- if eq .FieldType "string" }}
{
    whitespace: true,
    message: '不能只输入空格',
    trigger: ['input', 'blur'],
}
              {{- end }}
],
            {{- end }}
        {{- end }}
    {{- end }}



{{- if .HasDataSource }}
// 请引用
get{{.StructName}}DataSource,

//  获取数据源
const dataSource = ref({})
const getDataSourceFunc = async()=>{
  const res = await get{{.StructName}}DataSource()
  if (res.code === 0) {
    dataSource.value = res.data
  }
}
getDataSourceFunc()
{{- end }}

{{- else }}

{{- if not .OnlyTemplate}}
<template>
  <div>
    <GvaGrid ref="gridRef" v-bind="gridOptions" v-on="gridEvents">
      <template #toolbar-buttons>
        <el-button {{ if $global.AutoCreateBtnAuth }}v-auth="btnAuth.add"{{ end }} type="primary" icon="plus" @click="openDialog()">新增</el-button>
        <el-button {{ if $global.AutoCreateBtnAuth }}v-auth="btnAuth.batchDelete"{{ end }} icon="delete" :disabled="!selectedRows.length" @click="onDelete(selectedRows)">删除</el-button>
        {{ if .HasExcel -}}
        <ExportTemplate {{ if $global.AutoCreateBtnAuth }}v-auth="btnAuth.exportTemplate"{{ end }} template-id="{{$templateID}}" />
        <ExportExcel {{ if $global.AutoCreateBtnAuth }}v-auth="btnAuth.exportExcel"{{ end }} template-id="{{$templateID}}" filterDeleted/>
        <ImportExcel {{ if $global.AutoCreateBtnAuth }}v-auth="btnAuth.importExcel"{{ end }} template-id="{{$templateID}}" @on-success="refresh" />
        {{- end }}
      </template>

      {{- range .Fields}}
      {{- if .FieldSearchType}}
      {{ GenerateGridSearchItemSlot . }}
      {{- end }}
      {{- end }}

      {{- range .Fields}}
      {{- if .Table}}
      {{ GenerateGridColumnSlot . }}
      {{- end }}
      {{- end }}

      <template #operate="{ row }">
      {{- if .IsTree }}
        <el-button {{ if $global.AutoCreateBtnAuth }}v-auth="btnAuth.add"{{ end }} type="primary" link icon="plus" @click="openDialog(row)"><el-icon style="margin-right: 5px"><InfoFilled /></el-icon>新增子节点</el-button>
      {{- end }}
        <el-button {{ if $global.AutoCreateBtnAuth }}v-auth="btnAuth.info"{{ end }} type="primary" link icon="info-filled" @click="getDetails(row)"><el-icon style="margin-right: 5px"><InfoFilled /></el-icon>查看</el-button>
        <el-button {{ if $global.AutoCreateBtnAuth }}v-auth="btnAuth.edit"{{ end }} type="primary" link icon="edit" @click="update{{.StructName}}Func(row)">编辑</el-button>
        <el-button {{ if .IsTree }}v-if="!row.children?.length" {{ end }} {{if $global.AutoCreateBtnAuth }}v-auth="btnAuth.delete"{{ end }} type="primary" link icon="delete" @click="deleteRow(row)">删除</el-button>
      </template>
    </GvaGrid>

    <el-drawer destroy-on-close :size="appStore.drawerSize" v-model="dialogFormVisible" :show-close="false" :before-close="closeDialog">
       <template #header>
              <div class="flex justify-between items-center">
                <span class="text-lg">{{"{{"}}type==='create'?'新增':'编辑'{{"}}"}}</span>
                <div>
                  <el-button :loading="btnLoading" type="primary" @click="enterDialog">确 定</el-button>
                  <el-button @click="closeDialog">取 消</el-button>
                </div>
              </div>
            </template>

          <el-form :model="formData" label-position="top" ref="elFormRef" :rules="rule" label-width="80px">
          {{- if .IsTree }}
            <el-form-item label="父节点:" prop="parentID" >
                <el-tree-select
                    v-model="formData.parentID"
                    :data="[rootNode,...tableData]"
                    check-strictly
                    :render-after-expand="false"
                    :props="defaultProps"
                    clearable
                    style="width: 240px"
                    placeholder="根节点"
                />
            </el-form-item>
          {{- end }}
        {{- range .Fields}}
          {{- if .Form}}
            {{ GenerateFormItem . }}
          {{- end }}
          {{- end }}
          </el-form>
    </el-drawer>

    <el-drawer destroy-on-close :size="appStore.drawerSize" v-model="detailShow" :show-close="true" :before-close="closeDetailShow" title="查看">
            <el-descriptions :column="1" border>
            {{- if .IsTree }}
            <el-descriptions-item label="父节点">
                <el-tree-select
                  v-model="detailForm.parentID"
                  :data="[rootNode,...tableData]"
                  check-strictly
                  disabled
                  :render-after-expand="false"
                  :props="defaultProps"
                  clearable
                  style="width: 240px"
                  placeholder="根节点"
                />
            </el-descriptions-item>
            {{- end }}
            {{- range .Fields}}
              {{- if .Desc }}
                    {{ GenerateDescriptionItem . }}
              {{- end }}
            {{- end }}
            </el-descriptions>
        </el-drawer>

  </div>
</template>

<script setup>
import {
  {{- if .HasDataSource }}
    get{{.StructName}}DataSource,
  {{- end }}
  create{{.StructName}},
  delete{{.StructName}},
  delete{{.StructName}}ByIds,
  update{{.StructName}},
  find{{.StructName}},
  get{{.StructName}}List
} from '@/api/{{.Package}}/{{.PackageName}}'

{{- if or .HasPic .HasFile}}
import { getUrl } from '@/utils/image'
{{- end }}
{{- if .HasPic }}
// 图片选择组件
import SelectImage from '@/components/selectImage/selectImage.vue'
{{- end }}

{{- if .HasRichText }}
// 富文本组件
import RichEdit from '@/components/richtext/rich-edit.vue'
import RichView from '@/components/richtext/rich-view.vue'
{{- end }}

{{- if .HasFile }}
// 文件选择组件
import SelectFile from '@/components/selectFile/selectFile.vue'
{{- end }}

{{- if .HasArray}}
// 数组控制组件
import ArrayCtrl from '@/components/arrayCtrl/arrayCtrl.vue'
{{- end }}

// 全量引入格式化工具 请按需保留
import { getDictFunc, formatDate, formatBoolean, filterDict ,filterDataSource, returnArrImg, onDownloadFile } from '@/utils/format'
import { ElMessage } from 'element-plus'
import { ref, reactive } from 'vue'
{{- if .AutoCreateBtnAuth }}
// 引入按钮权限标识
import { useBtnAuth } from '@/utils/btnAuth'
{{- end }}
import { useAppStore } from "@/pinia"
import GvaGrid, { useGvaGrid, useGvaGridDelete } from '@/components/gvaGrid'

{{if .HasExcel -}}
// 导出组件
import ExportExcel from '@/components/exportExcel/exportExcel.vue'
// 导入组件
import ImportExcel from '@/components/exportExcel/importExcel.vue'
// 导出模板组件
import ExportTemplate from '@/components/exportExcel/exportTemplate.vue'
{{- end}}


defineOptions({
    name: '{{.StructName}}'
})

{{- if .AutoCreateBtnAuth }}
// 按钮权限实例化
    const btnAuth = useBtnAuth()
{{- end }}

// 提交按钮loading
const btnLoading = ref(false)
const appStore = useAppStore()

// 自动化生成的字典（可能为空）以及字段
    {{- range $index, $element := .DictTypes}}
const {{ $element }}Options = ref([])
    {{- end }}
const formData = ref({
        {{- if .IsTree }}
            parentID:undefined,
        {{- end }}
        {{- range .Fields}}
          {{- if .Form}}
            {{ GenerateDefaultFormValue . }}
          {{- end }}
        {{- end }}
        })

{{- if .HasDataSource }}
  const dataSource = ref([])
  const getDataSourceFunc = async()=>{
    const res = await get{{.StructName}}DataSource()
    if (res.code === 0) {
      dataSource.value = res.data
    }
  }
  getDataSourceFunc()
{{- end }}



// 验证规则
const rule = reactive({
    {{- range .Fields }}
        {{- if .Form }}
            {{- if eq .Require true }}
               {{.FieldJson }} : [{
                   required: true,
                   message: '{{ .ErrorText }}',
                   trigger: ['input','blur'],
               },
               {{- if eq .FieldType "string" }}
               {
                   whitespace: true,
                   message: '不能只输入空格',
                   trigger: ['input', 'blur'],
              }
              {{- end }}
              ],
            {{- end }}
        {{- end }}
    {{- end }}
})

const elFormRef = ref()

{{- if .IsTree }}
// 树选择器配置
const defaultProps = {
  children: "children",
  label: "{{ .TreeJson }}",
  value: "{{ .PrimaryField.FieldJson }}"
}

const rootNode = {
  {{ .PrimaryField.FieldJson }}: 0,
  {{ .TreeJson }}: '根节点',
  children: []
}
{{- end }}

// =========== GvaGrid 表格配置 ===========
const tableData = ref([])

const { gridRef, gridOptions, gridEvents, selectedRows, refresh } = useGvaGrid({
    id: '{{.Package}}-{{.PackageName}}',
    api: async (params) => {
        const res = await get{{.StructName}}List(params)
        {{- if .IsTree }}
        tableData.value = res.data || []
        return { ...res, data: { list: tableData.value, total: tableData.value.length } }
        {{- else }}
        return res
        {{- end }}
    },
    {{- if .IsTree }}
    pager: false,
    {{- end }}
    defaultSort: null,
    {{- if .NeedSort }}
    sortProtocol: 'autoCode',
    {{- end }}
    checkbox: true,
    {{- if not .IsTree }}
    searchItems: [
      {{- if .GvaModel }}
      { field: 'createdAtRange', title: '创建日期', span: 8, itemRender: { name: 'gvaDateRange', props: { type: 'datetimerange', valueFormat: 'YYYY-MM-DD HH:mm:ss' } } },
      {{- end }}
      {{- range .Fields}}{{- if .FieldSearchType}}{{- if not .FieldSearchHide }}
      {{ GenerateGridSearchItem .}}
      {{- end }}{{- end }}{{- end }}
    ],
    {{- end }}
    gridConfig: {
      {{- if .IsTree }}
      treeConfig: { rowField: '{{.PrimaryField.FieldJson}}', children: 'children', expandAll: false }
      {{- else }}
      customConfig: { storage: true }
      {{- end }}
    },
    columns: [
      {{- if .GvaModel }}
      { field: 'CreatedAt', title: '日期', width: 180, {{- if .NeedSort }} sortable: true,{{- end }} cellRender: { name: 'gvaDate' } },
      {{- end }}
      {{- range .Fields}}
      {{- if .Table}}
      {{ GenerateGridColumnSchema . }}
      {{- end }}
      {{- end }}
      { title: '操作', fixed: 'right', slots: { default: 'operate' } }
    ]
})

{{- if .IsTree }}
// 树表首列显示树形展开
if (gridOptions.columns && gridOptions.columns.length) {
  const firstFieldColumn = gridOptions.columns.find((c) => !c.type && c.field !== 'operate')
  if (firstFieldColumn) firstFieldColumn.treeNode = true
}
{{- end }}

// ============== GvaGrid 表格配置结束 ===============

// 获取需要的字典 可能为空 按需保留
const setOptions = async () =>{
{{- range $index, $element := .DictTypes }}
    {{ $element }}Options.value = await getDictFunc('{{$element}}')
{{- end }}
}

// 获取需要的字典 可能为空 按需保留
setOptions()

// 删除（单行/批量，含确认框、成功提示、刷新、末页删空回退）
const { deleteRow, deleteRows: onDelete } = useGvaGridDelete(gridRef, {
    delete: (rows) => {
      if (rows.length === 1) {
        return delete{{.StructName}}({ {{.PrimaryField.FieldJson}}: rows[0].{{.PrimaryField.FieldJson}} })
      }
      return delete{{.StructName}}ByIds({ {{.PrimaryField.FieldJson}}s: rows.map((item) => item.{{.PrimaryField.FieldJson}}) })
    },
    confirmText: '确定要删除吗?',
    successText: '删除成功'
})

// 行为控制标记（弹窗内部需要增还是改）
const type = ref('')

// 更新行
const update{{.StructName}}Func = async(row) => {
    const res = await find{{.StructName}}({ {{.PrimaryField.FieldJson}}: row.{{.PrimaryField.FieldJson}} })
    type.value = 'update'
    if (res.code === 0) {
        formData.value = res.data
        dialogFormVisible.value = true
    }
}

// 弹窗控制标记
const dialogFormVisible = ref(false)

// 打开弹窗
const openDialog = ({{- if .IsTree -}}row{{- end -}}) => {
    type.value = 'create'
    {{- if .IsTree }}
    formData.value.parentID = row ? row.{{.PrimaryField.FieldJson}} : undefined
    {{- end }}
    dialogFormVisible.value = true
}

// 关闭弹窗
const closeDialog = () => {
    dialogFormVisible.value = false
    formData.value = {
    {{- range .Fields}}
      {{- if .Form}}
        {{ GenerateDefaultFormValue . }}
      {{- end }}
    {{- end }}
        }
}
// 弹窗确定
const enterDialog = async () => {
     btnLoading.value = true
     elFormRef.value?.validate( async (valid) => {
             if (!valid) return btnLoading.value = false
              let res
              switch (type.value) {
                case 'create':
                  res = await create{{.StructName}}(formData.value)
                  break
                case 'update':
                  res = await update{{.StructName}}(formData.value)
                  break
                default:
                  res = await create{{.StructName}}(formData.value)
                  break
              }
              btnLoading.value = false
              if (res.code === 0) {
                ElMessage({
                  type: 'success',
                  message: '创建/更改成功'
                })
                closeDialog()
                refresh()
              }
      })
}

const detailForm = ref({})

// 查看详情控制标记
const detailShow = ref(false)


// 打开详情弹窗
const openDetailShow = () => {
  detailShow.value = true
}


// 打开详情
const getDetails = async (row) => {
  // 打开弹窗
  const res = await find{{.StructName}}({ {{.PrimaryField.FieldJson}}: row.{{.PrimaryField.FieldJson}} })
  if (res.code === 0) {
    detailForm.value = res.data
    openDetailShow()
  }
}


// 关闭详情弹窗
const closeDetailShow = () => {
  detailShow.value = false
  detailForm.value = {}
}


</script>

<style>
{{if .HasFile }}
.file-list{
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.fileBtn{
  margin-bottom: 10px;
}

.fileBtn:last-child{
  margin-bottom: 0;
}
{{end}}
</style>
{{- else}}
<template>
<div>form</div>
</template>
<script setup>
defineOptions({
  name: '{{.StructName}}'
})
</script>
<style>
</style>
{{- end }}

{{- end }}
