package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
)

// SearchExtModule 模块查询（分页参数兼容，实际全量返回）
type SearchExtModule struct {
	request.PageInfo
}

// ExtModuleOps 单条操作
type ExtModuleOps struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// ExtTableListOps 模块下注册表清单
type ExtTableListOps struct {
	ModuleId uint `form:"moduleId" binding:"required"` //nolint:stylecheck // 对齐前端 moduleId
}

// ProbeExtTablesOps 探测物理表
type ProbeExtTablesOps struct {
	Keyword string `form:"keyword"`
}

// ExtTableColumnsOps 探测表列
type ExtTableColumnsOps struct {
	TableName string `form:"tableName" binding:"required"` //nolint:stylecheck // 对齐前端 tableName
}

// RegisterExtTableItem 批量注册行
type RegisterExtTableItem struct {
	TableName   string `json:"tableName" binding:"required"` //nolint:stylecheck // 对齐前端
	DisplayName string `json:"displayName"`                  //nolint:stylecheck // 对齐前端
	Remark      string `json:"remark"`
}

// RegisterExtTablesOps 批量注册
type RegisterExtTablesOps struct {
	ModuleId uint                   `json:"moduleId" binding:"required"` //nolint:stylecheck // 对齐前端
	Items    []RegisterExtTableItem `json:"items" binding:"required,min=1"`
}

// SearchExtBinding 绑定分页
type SearchExtBinding struct {
	request.PageInfo
	ProjectId     uint `form:"projectId"`     //nolint:stylecheck // 对齐前端
	ClassId       uint `form:"classId"`       //nolint:stylecheck // 对齐前端
	BindingStatus int  `form:"bindingStatus"` //nolint:stylecheck // 对齐前端；-1/0=全部
}

// ExtBindingOps 单条操作
type ExtBindingOps struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// ChangeExtBindingStatusOps 生效/停用（1生效/2停用）
type ChangeExtBindingStatusOps struct {
	ID     uint `form:"ID" binding:"required"`
	Status int  `form:"status" binding:"required,oneof=1 2"`
}

// TriggerSyncOps 触发同步（1全量/2增量）
type TriggerSyncOps struct {
	BindingId uint `json:"bindingId" form:"bindingId" binding:"required"` //nolint:stylecheck // 对齐前端
	Scope     int  `json:"scope" form:"scope" binding:"required,oneof=1 2"`
}

// DryRunOps 试运行（limit 默认 20）
type DryRunOps struct {
	BindingId uint `form:"bindingId" binding:"required"` //nolint:stylecheck // 对齐前端
	Limit     int  `form:"limit"`
}

// SaveExtBinding 保存载荷：绑定主表 + 子表绑定 + 属性绑定（整体替换式保存）
type SaveExtBinding struct {
	ontology.OntExtBinding
	Details    []ontology.OntExtBindingDetail   `json:"details"`
	Properties []ontology.OntExtBindingProperty `json:"properties"`
}

// SearchExtSyncLog 同步日志分页
type SearchExtSyncLog struct {
	request.PageInfo
	BindingId uint `form:"bindingId"` //nolint:stylecheck // 对齐前端
	ClassId   uint `form:"classId"`   //nolint:stylecheck // 对齐前端
}
