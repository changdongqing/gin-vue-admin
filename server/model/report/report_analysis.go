package report

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// ReportAnalysisReport 分析报表元数据（配置内容按 report_code 1:1 存于 ReportAnalysisConfig）
type ReportAnalysisReport struct {
	global.GVA_MODEL
	ReportCode  string `json:"reportCode" form:"reportCode" gorm:"column:report_code;size:100;not null;comment:报表编码(唯一,创建后不可改)" binding:"required"`
	ReportName  string `json:"reportName" form:"reportName" gorm:"column:report_name;size:100;not null;comment:报表名称" binding:"required"`
	ReportGroup string `json:"reportGroup" form:"reportGroup" gorm:"column:report_group;size:100;comment:报表分组"`
	ReportDesc  string `json:"reportDesc" form:"reportDesc" gorm:"column:report_desc;size:255;comment:描述"`
	SetCode     string `json:"setCode" form:"setCode" gorm:"column:set_code;size:50;not null;comment:关联数据集编码(单选)"`
	Status      int    `json:"status" form:"status" gorm:"column:status;not null;default:0;comment:0启用/1禁用"`
	CreatedBy   string `json:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy   string `json:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (ReportAnalysisReport) TableName() string {
	return "report_analysis_reports"
}

// ReportAnalysisConfig 分析报表配置内容（S2 配置 JSON：fields/options/theme 三段）
type ReportAnalysisConfig struct {
	global.GVA_MODEL
	ReportCode string `json:"reportCode" form:"reportCode" gorm:"column:report_code;size:100;not null;comment:报表编码(唯一)"`
	ConfigJson string `json:"configJson" form:"configJson" gorm:"column:config_json;type:text;comment:S2配置JSON"`
	SetParam   string `json:"setParam" form:"setParam" gorm:"column:set_param;type:text;comment:参数默认值JSON(预留)"`
	CreatedBy  string `json:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy  string `json:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (ReportAnalysisConfig) TableName() string {
	return "report_analysis_configs"
}
