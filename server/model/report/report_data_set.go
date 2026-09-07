package report

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// ReportDataSet 报表数据集（主表；set_code 全局唯一，被报表引用后不可改名）
type ReportDataSet struct {
	global.GVA_MODEL
	SetCode     string `json:"setCode" form:"setCode" gorm:"column:set_code;size:50;not null;comment:数据集编码" binding:"required"`
	SetName     string `json:"setName" form:"setName" gorm:"column:set_name;size:100;not null;comment:数据集名称" binding:"required"`
	SetDesc     string `json:"setDesc" form:"setDesc" gorm:"column:set_desc;size:255;comment:描述"`
	SourceCode  string `json:"sourceCode" form:"sourceCode" gorm:"column:source_code;size:50;not null;comment:关联数据源编码(http类型为空)"`
	SetType     string `json:"setType" form:"setType" gorm:"column:set_type;size:10;not null;default:sql;comment:类型 sql/http"`
	DynSentence string `json:"dynSentence" form:"dynSentence" gorm:"column:dyn_sentence;size:4096;not null;comment:查询语句(SQL或HTTP配置JSON)"`
	CaseResult  string `json:"caseResult" form:"caseResult" gorm:"column:case_result;type:text;comment:结果案例JSON(设计器字段来源)"`
	EnableFlag  bool   `json:"enableFlag" form:"enableFlag" gorm:"column:enable_flag;not null;default:true;comment:是否启用"`
	CreatedBy   string `json:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy   string `json:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (ReportDataSet) TableName() string {
	return "report_data_sets"
}

// ReportDataSetParam 数据集参数子表（set_code 逻辑关联）
type ReportDataSetParam struct {
	global.GVA_MODEL
	SetCode         string `json:"setCode" form:"setCode" gorm:"column:set_code;size:50;not null;comment:所属数据集编码"`
	ParamName       string `json:"paramName" form:"paramName" gorm:"column:param_name;size:50;not null;comment:参数名" binding:"required"`
	ParamDesc       string `json:"paramDesc" form:"paramDesc" gorm:"column:param_desc;size:100;comment:参数描述"`
	ParamType       string `json:"paramType" form:"paramType" gorm:"column:param_type;size:20;not null;default:string;comment:参数类型"`
	SampleItem      string `json:"sampleItem" form:"sampleItem" gorm:"column:sample_item;size:1080;comment:示例值"`
	DefaultValue    string `json:"defaultValue" form:"defaultValue" gorm:"column:default_value;size:1080;comment:默认值(直接值或表达式)"`
	DictType        string `json:"dictType" form:"dictType" gorm:"column:dict_type;size:100;comment:字典类型(预留)"`
	CustomOptions   string `json:"customOptions" form:"customOptions" gorm:"column:custom_options;size:2048;comment:自定义选项JSON(预留)"`
	DateFormat      string `json:"dateFormat" form:"dateFormat" gorm:"column:date_format;size:50;comment:日期格式(预留)"`
	RequiredFlag    bool   `json:"requiredFlag" form:"requiredFlag" gorm:"column:required_flag;not null;default:false;comment:是否必填"`
	ValidationRules string `json:"validationRules" form:"validationRules" gorm:"column:validation_rules;size:2048;comment:JS校验规则(预留)"`
	OrderNum        int    `json:"orderNum" form:"orderNum" gorm:"column:order_num;not null;default:0;comment:排序"`
	CreatedBy       string `json:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy       string `json:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (ReportDataSetParam) TableName() string {
	return "report_data_set_params"
}

// ReportDataSetTransform 数据集转换子表（set_code 逻辑关联，按 orderNum 串行执行）
type ReportDataSetTransform struct {
	global.GVA_MODEL
	SetCode         string `json:"setCode" form:"setCode" gorm:"column:set_code;size:50;not null;comment:所属数据集编码"`
	TransformType   string `json:"transformType" form:"transformType" gorm:"column:transform_type;size:50;not null;comment:转换类型 js/dict"`
	TransformScript string `json:"transformScript" form:"transformScript" gorm:"column:transform_script;type:text;comment:转换脚本(js代码或dict映射JSON)"`
	OrderNum        int    `json:"orderNum" form:"orderNum" gorm:"column:order_num;not null;default:0;comment:排序"`
	CreatedBy       string `json:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy       string `json:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (ReportDataSetTransform) TableName() string {
	return "report_data_set_transforms"
}
