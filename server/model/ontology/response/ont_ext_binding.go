package response

import (
	ontModel "github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
)

// ExtModulePageItem 模块列表项（含注册表数）
type ExtModulePageItem struct {
	ontModel.OntExtModule
	TableCount int64 `json:"tableCount"`
}

// ProbeExtTableItem 物理表探测行
type ProbeExtTableItem struct {
	TableName    string `json:"tableName"` //nolint:stylecheck // 对齐前端
	TableComment string `json:"tableComment"`
	Registered   bool   `json:"registered"`
}

// ExtTableColumnItem 表列探测行
type ExtTableColumnItem struct {
	ColumnName    string `json:"columnName"` //nolint:stylecheck // 对齐前端
	DataType      string `json:"dataType"`   //nolint:stylecheck // 对齐前端
	ColumnComment string `json:"columnComment"`
}

// ExtBindingPageItem 绑定列表富化行
type ExtBindingPageItem struct {
	ontModel.OntExtBinding
	ProjectCode    string `json:"projectCode"`
	ClassLocalName string `json:"classLocalName"`
	ClassLabelCn   string `json:"classLabelCn"`
	TableName      string `json:"tableName"` //nolint:stylecheck // 对齐前端
	PropertyCount  int64  `json:"propertyCount"`
	DetailCount    int64  `json:"detailCount"`
}

// DryRunValue 试运行值转换行
type DryRunValue struct {
	Label     string `json:"label"`
	Source    string `json:"source"` // 主表 / 子表:xxx / 窄表:xxx(identifier)
	RawValue  string `json:"rawValue"`
	Converted bool   `json:"converted"`
}

// DryRunRelation 试运行关系解析行
type DryRunRelation struct {
	Label            string `json:"label"`
	BizColumnValue   string `json:"bizColumnValue"`
	Resolved         bool   `json:"resolved"`
	TargetObjectName string `json:"targetObjectName"`
}

// DryRunRow 试运行预览行
type DryRunRow struct {
	BizKey      string           `json:"bizKey"`
	PreviewCode string           `json:"previewCode"`
	PreviewName string           `json:"previewName"`
	Values      []DryRunValue    `json:"values"`
	Relations   []DryRunRelation `json:"relations"`
}

// DryRunResult 试运行结果（零写入零日志）
type DryRunResult struct {
	Rows     []DryRunRow `json:"rows"`
	Warnings []string    `json:"warnings"`
}

// SyncResult 同步结果（同时落一条日志）
type SyncResult struct {
	BindingId   uint     `json:"bindingId"`
	ClassId     uint     `json:"classId"`
	TriggerType int      `json:"triggerType"`
	SyncScope   int      `json:"syncScope"`
	TotalRows   int      `json:"totalRows"`
	Created     int      `json:"created"`
	Updated     int      `json:"updated"`
	Skipped     int      `json:"skipped"`
	Failed      int      `json:"failed"`
	IssuesCount int      `json:"issuesCount"`
	DurationMs  int      `json:"durationMs"`
	Status      int      `json:"status"`
	SkipReasons []string `json:"skipReasons"`
}
