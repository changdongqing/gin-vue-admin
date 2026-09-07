package report

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// ReportExcelReport Excel 报表元数据（模板内容按 report_code 1:1 存于 ReportExcelTemplate）
type ReportExcelReport struct {
	global.GVA_MODEL
	ReportCode  string `json:"reportCode" form:"reportCode" gorm:"column:report_code;size:100;not null;comment:报表编码(唯一,创建后不可改)" binding:"required"`
	ReportName  string `json:"reportName" form:"reportName" gorm:"column:report_name;size:100;not null;comment:报表名称" binding:"required"`
	ReportGroup string `json:"reportGroup" form:"reportGroup" gorm:"column:report_group;size:100;comment:报表分组"`
	ReportDesc  string `json:"reportDesc" form:"reportDesc" gorm:"column:report_desc;size:255;comment:描述"`
	CreatedBy   string `json:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy   string `json:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (ReportExcelReport) TableName() string {
	return "report_excel_reports"
}

// ReportExcelTemplate Excel 报表模板内容（set_codes 与 json_str 两路独立维护：
// bind-datasets 只写 set_codes；save-template 只写 json_str/set_param）
type ReportExcelTemplate struct {
	global.GVA_MODEL
	ReportCode string `json:"reportCode" form:"reportCode" gorm:"column:report_code;size:100;not null;comment:报表编码(唯一)"`
	SetCodes   string `json:"setCodes" form:"setCodes" gorm:"column:set_codes;size:500;comment:关联数据集编码(|分隔)"`
	SetParam   string `json:"setParam" form:"setParam" gorm:"column:set_param;type:text;comment:参数默认值JSON(预留)"`
	JsonStr    string `json:"jsonStr" form:"jsonStr" gorm:"column:json_str;type:text;comment:Univer工作簿快照JSON"`
	CreatedBy  string `json:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy  string `json:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (ReportExcelTemplate) TableName() string {
	return "report_excel_templates"
}
