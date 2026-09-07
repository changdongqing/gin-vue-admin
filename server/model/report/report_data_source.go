package report

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// ReportDataSource 报表数据源（取数底座：外部数据库连接配置 + 连接池管理）
// 唯一业务索引用迁移 SQL 的部分唯一索引（WHERE deleted_at IS NULL）+ Service 查重共同保证，
// 此处不声明 uniqueIndex tag（AutoMigrate 对 PG 部分索引支持有限，结构以迁移为准）
type ReportDataSource struct {
	global.GVA_MODEL
	SourceCode   string `json:"sourceCode" form:"sourceCode" gorm:"column:source_code;size:50;not null;comment:数据源编码" binding:"required"`
	SourceName   string `json:"sourceName" form:"sourceName" gorm:"column:source_name;size:100;not null;comment:数据源名称" binding:"required"`
	SourceType   string `json:"sourceType" form:"sourceType" gorm:"column:source_type;size:50;not null;comment:数据源类型" binding:"required"`
	SourceDesc   string `json:"sourceDesc" form:"sourceDesc" gorm:"column:source_desc;size:255;comment:描述"`
	SourceConfig string `json:"sourceConfig" form:"sourceConfig" gorm:"column:source_config;size:2048;not null;comment:连接配置JSON(密码加密)" binding:"required"`
	EnableFlag   bool   `json:"enableFlag" form:"enableFlag" gorm:"column:enable_flag;not null;default:true;comment:是否启用"`
	CreatedBy    string `json:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy    string `json:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (ReportDataSource) TableName() string {
	return "report_data_sources"
}
