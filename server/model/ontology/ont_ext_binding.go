package ontology

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// OntExtModule 外部模块注册（业务表逻辑分组）
type OntExtModule struct {
	global.GVA_MODEL
	ModuleCode  string `json:"moduleCode" form:"moduleCode" gorm:"column:module_code;size:64;not null;comment:模块编码(唯一)" binding:"required"`
	Name        string `json:"name" form:"name" gorm:"column:name;size:128;not null;comment:显示名" binding:"required"`
	ServiceName string `json:"serviceName" form:"serviceName" gorm:"column:service_name;size:128;comment:所属服务"`
	Description string `json:"description" form:"description" gorm:"column:description;size:512;comment:描述"`
	Status      int    `json:"status" form:"status" gorm:"column:status;default:0;comment:0启用/1停用"`
	SortOrder   int    `json:"sortOrder" form:"sortOrder" gorm:"column:sort_order;default:0;comment:排序"`
	CreatedBy   string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy   string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (OntExtModule) TableName() string {
	return "ont_ext_modules"
}

// OntExtTable 外部表注册（物理表全局至多注册一次）
type OntExtTable struct {
	global.GVA_MODEL
	ModuleId            uint   `json:"moduleId" form:"moduleId" gorm:"column:module_id;not null;comment:所属模块ID" binding:"required"` //nolint:stylecheck // 对齐前端 moduleId
	Table               string `json:"tableName" form:"tableName" gorm:"column:table_name;size:128;not null;comment:物理表名"`          //nolint:stylecheck // 字段名 Table 避让 TableName() 方法；JSON 契约保持 tableName
	DisplayName         string `json:"displayName" form:"displayName" gorm:"column:display_name;size:128;comment:显示名"`
	PkColumn            string `json:"pkColumn" form:"pkColumn" gorm:"column:pk_column;size:128;not null;default:id;comment:主键列"`
	PkType              string `json:"pkType" form:"pkType" gorm:"column:pk_type;size:64;comment:主键类型"`
	DeletedColumn       string `json:"deletedColumn" form:"deletedColumn" gorm:"column:deleted_column;size:128;comment:软删列(探测回填)"`
	UpdateTimeColumn    string `json:"updateTimeColumn" form:"updateTimeColumn" gorm:"column:update_time_column;size:128;comment:水位列(探测回填)"` //nolint:stylecheck // 对齐前端
	SupportsIncremental int    `json:"supportsIncremental" form:"supportsIncremental" gorm:"column:supports_incremental;default:0;comment:支持增量0/1"`
	Remark              string `json:"remark" form:"remark" gorm:"column:remark;size:512;comment:备注"`
	CreatedBy           string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy           string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (OntExtTable) TableName() string {
	return "ont_ext_tables"
}

// OntExtBinding 类绑定配置（类↔主表；一个类至多一条生效绑定）
type OntExtBinding struct {
	global.GVA_MODEL
	ProjectId           uint       `json:"projectId" form:"projectId" gorm:"column:project_id;not null;comment:所属项目ID"` //nolint:stylecheck // 对齐前端
	ClassId             uint       `json:"classId" form:"classId" gorm:"column:class_id;not null;comment:绑定的类ID"`       //nolint:stylecheck // 对齐前端
	TableId             uint       `json:"tableId" form:"tableId" gorm:"column:table_id;not null;comment:主表注册ID"`       //nolint:stylecheck // 对齐前端
	KeyColumn           string     `json:"keyColumn" form:"keyColumn" gorm:"column:key_column;size:128;not null;default:id;comment:主键列"`
	CodeColumn          string     `json:"codeColumn" form:"codeColumn" gorm:"column:code_column;size:128;comment:编码列(空=生成 类名-主键)"`
	NameColumn          string     `json:"nameColumn" form:"nameColumn" gorm:"column:name_column;size:128;comment:名称列"`
	ParentColumn        string     `json:"parentColumn" form:"parentColumn" gorm:"column:parent_column;size:128;comment:父列(树形)"`
	StatusColumn        string     `json:"statusColumn" form:"statusColumn" gorm:"column:status_column;size:128;comment:状态列"`
	StatusActiveValue   string     `json:"statusActiveValue" form:"statusActiveValue" gorm:"column:status_active_value;size:64;comment:状态在用值"`
	SyncMode            int        `json:"syncMode" form:"syncMode" gorm:"column:sync_mode;default:1;comment:1手动/2定时+手动"`
	SyncOrder           int        `json:"syncOrder" form:"syncOrder" gorm:"column:sync_order;default:100;comment:编排顺序"`
	IncludeDisabled     int        `json:"includeDisabled" form:"includeDisabled" gorm:"column:include_disabled;default:0;comment:含停用行0/1"`
	ConflictStrategy    int        `json:"conflictStrategy" form:"conflictStrategy" gorm:"column:conflict_strategy;default:2;comment:1仅报告/2覆盖"`
	MissingTargetPolicy int        `json:"missingTargetPolicy" form:"missingTargetPolicy" gorm:"column:missing_target_policy;default:1;comment:1跳过记issue/2整行失败"`
	BindingStatus       int        `json:"bindingStatus" form:"bindingStatus" gorm:"column:binding_status;default:0;comment:0草稿/1生效/2停用"`
	TelemetryEnabled    int        `json:"telemetryEnabled" form:"telemetryEnabled" gorm:"column:telemetry_enabled;default:0;comment:遥测直读(预留)"`
	Watermark           *time.Time `json:"watermark" form:"watermark" gorm:"column:watermark;comment:增量水位"`
	LastSyncTime        *time.Time `json:"lastSyncTime" form:"lastSyncTime" gorm:"column:last_sync_time;comment:最近同步时间"`
	LastSyncSummary     string     `json:"lastSyncSummary" form:"lastSyncSummary" gorm:"column:last_sync_summary;size:512;comment:结果摘要"`
	Remark              string     `json:"remark" form:"remark" gorm:"column:remark;size:512;comment:备注"`
	CreatedBy           string     `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy           string     `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (OntExtBinding) TableName() string {
	return "ont_ext_bindings"
}

// OntExtBindingDetail 子表绑定（宽表一对一/窄表 EAV）
type OntExtBindingDetail struct {
	global.GVA_MODEL
	BindingId        uint   `json:"bindingId" form:"bindingId" gorm:"column:binding_id;not null;comment:绑定ID"` //nolint:stylecheck // 对齐前端
	TableId          uint   `json:"tableId" form:"tableId" gorm:"column:table_id;not null;comment:子表注册ID"`     //nolint:stylecheck // 对齐前端
	DetailKind       int    `json:"detailKind" form:"detailKind" gorm:"column:detail_kind;default:1;comment:1宽表/2窄表"`
	JoinColumn       string `json:"joinColumn" form:"joinColumn" gorm:"column:join_column;size:128;comment:回连外键列"`
	IdentifierColumn string `json:"identifierColumn" form:"identifierColumn" gorm:"column:identifier_column;size:128;comment:窄表标识列"`
	ValueColumn      string `json:"valueColumn" form:"valueColumn" gorm:"column:value_column;size:128;comment:窄表值列"`
	TsColumn         string `json:"tsColumn" form:"tsColumn" gorm:"column:ts_column;size:128;comment:时间列"`
	SortOrder        int    `json:"sortOrder" form:"sortOrder" gorm:"column:sort_order;default:0;comment:排序"`
	Remark           string `json:"remark" form:"remark" gorm:"column:remark;size:512;comment:备注"`
}

func (OntExtBindingDetail) TableName() string {
	return "ont_ext_binding_details"
}

// OntExtBindingProperty 属性绑定（类属性↔业务列）
type OntExtBindingProperty struct {
	global.GVA_MODEL
	BindingId       uint   `json:"bindingId" form:"bindingId" gorm:"column:binding_id;not null;comment:绑定ID"` //nolint:stylecheck // 对齐前端
	BindingType     int    `json:"bindingType" form:"bindingType" gorm:"column:binding_type;default:1;comment:1数据/2对象"`
	PropertyId      uint   `json:"propertyId" form:"propertyId" gorm:"column:property_id;not null;comment:类属性ID" binding:"required"` //nolint:stylecheck // 对齐前端
	DetailBindingId *uint  `json:"detailBindingId" form:"detailBindingId" gorm:"column:detail_binding_id;comment:值来源子表(NULL=主表)"`    //nolint:stylecheck // 对齐前端
	BizColumn       string `json:"bizColumn" form:"bizColumn" gorm:"column:biz_column;size:128;comment:业务列/窄表匹配值"`
	ValueConverter  int    `json:"valueConverter" form:"valueConverter" gorm:"column:value_converter;default:0;comment:值转换器(预留)"`
	ConstValue      string `json:"constValue" form:"constValue" gorm:"column:const_value;size:512;comment:常量值(预留)"`
	Enabled         int    `json:"enabled" form:"enabled" gorm:"column:enabled;default:1;comment:启用0/1"`
	SortOrder       int    `json:"sortOrder" form:"sortOrder" gorm:"column:sort_order;default:0;comment:排序"`
	Remark          string `json:"remark" form:"remark" gorm:"column:remark;size:512;comment:备注"`
}

func (OntExtBindingProperty) TableName() string {
	return "ont_ext_binding_properties"
}

// OntExtSyncLog 外部同步日志（一次触发一条；无软删列——日志不删）
type OntExtSyncLog struct {
	ID          uint       `json:"ID" gorm:"primarykey"`
	BindingId   uint       `json:"bindingId" gorm:"column:binding_id;not null;comment:绑定ID"` //nolint:stylecheck // 对齐前端
	ClassId     uint       `json:"classId" gorm:"column:class_id;not null;comment:类ID"`      //nolint:stylecheck // 对齐前端
	TriggerType int        `json:"triggerType" gorm:"column:trigger_type;default:2;comment:1反向/2手动/3定时/4试运行"`
	SyncScope   int        `json:"syncScope" gorm:"column:sync_scope;default:1;comment:1全量/2增量"` //nolint:stylecheck // 对齐前端 syncScope
	TotalRows   int        `json:"totalRows" gorm:"column:total_rows;default:0;comment:扫描行数"`
	Created     int        `json:"created" gorm:"column:created;default:0;comment:新建数"`
	Updated     int        `json:"updated" gorm:"column:updated;default:0;comment:更新数"`
	Skipped     int        `json:"skipped" gorm:"column:skipped;default:0;comment:跳过数"`
	Failed      int        `json:"failed" gorm:"column:failed;default:0;comment:失败数"`
	IssuesCount int        `json:"issuesCount" gorm:"column:issues_count;default:0;comment:问题数"`
	Status      int        `json:"status" gorm:"column:status;default:1;comment:1成功/2部分失败/3失败"`
	DurationMs  int        `json:"durationMs" gorm:"column:duration_ms;default:0;comment:耗时ms"`
	Summary     string     `json:"summary" gorm:"column:summary;size:1000;comment:结果摘要"`
	SkipReasons string     `json:"skipReasons" gorm:"column:skip_reasons;type:text;comment:跳过原因JSON数组"`
	StartTime   time.Time  `json:"startTime" gorm:"column:start_time;comment:开始时间"` //nolint:stylecheck // 对齐前端 startTime
	EndTime     *time.Time `json:"endTime" gorm:"column:end_time;comment:结束时间"`
}

func (OntExtSyncLog) TableName() string {
	return "ont_ext_sync_logs"
}
