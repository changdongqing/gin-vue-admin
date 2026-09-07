package report

import "errors"

// 报表平台领域 sentinel errors（中文消息直传前端；跨功能引用统一放本文件）
var (
	// ── 01 数据源 ──────────────────────────────────────────
	ErrSourceNotExists      = errors.New("数据源不存在")
	ErrSourceCodeDuplicate  = errors.New("数据源编码已存在")
	ErrSourceTypeInvalid    = errors.New("不支持的数据源类型")
	ErrSourceDriverDisabled = errors.New("该类型数据库驱动未启用，请联系管理员")
	ErrSourceDisabled       = errors.New("数据源已禁用")
	ErrSourceConfigInvalid  = errors.New("数据源连接配置格式错误")
	ErrSourceConnFailed     = errors.New("数据源连接失败")
	ErrSourceReferenced     = errors.New("数据源被数据集引用，无法删除")

	// ── 02 数据集 ──────────────────────────────────────────
	ErrSetNotExists          = errors.New("数据集不存在")
	ErrSetCodeDuplicate      = errors.New("数据集编码已存在")
	ErrSetReferencedByReport = errors.New("数据集被报表引用，无法删除")
	ErrSetSourceMissing      = errors.New("SQL 类型数据集必须选择数据源")
	ErrSetSourceDisabled     = errors.New("数据集关联的数据源已禁用")
	ErrSetSqlInvalid         = errors.New("数据集SQL不合法：仅允许 SELECT 查询")
	ErrSetSqlDangerous       = errors.New("数据集SQL包含危险关键词")
	ErrSetParamRequired      = errors.New("数据集参数为必填项")
	ErrSetParamTypeInvalid   = errors.New("数据集参数类型不合法")
	ErrSetQueryParamInvalid  = errors.New("数据集参数取值不合法")
	ErrSetQueryFailed        = errors.New("数据集查询失败")
	ErrSetHttpFailed         = errors.New("HTTP数据集请求失败")
	ErrSetHttpConfigInvalid  = errors.New("HTTP数据集请求配置格式错误")
	ErrSetTransformFailed    = errors.New("数据集转换执行失败")
	ErrSetResultTooLarge     = errors.New("数据集查询结果超过最大行数限制（50000行）")

	// ── 03 Excel 报表 ──────────────────────────────────────
	ErrExcelReportNotExists        = errors.New("Excel报表不存在")
	ErrExcelReportCodeDuplicate    = errors.New("报表编码已存在")
	ErrExcelTemplateJSONInvalid    = errors.New("报表模板JSON格式错误")
	ErrExcelCopySourceMissing      = errors.New("复制源报表不存在")
	ErrExcelReportDataSetNotExists = errors.New("报表关联的数据集不存在")

	// ── 04 Excel 预览渲染 ──────────────────────────────────
	ErrExcelRenderFailed = errors.New("报表渲染失败")

	// ── 05 分析报表 ────────────────────────────────────────
	ErrAnalysisReportNotExists     = errors.New("分析报表不存在")
	ErrAnalysisReportCodeDuplicate = errors.New("分析报表编码已存在")
	ErrAnalysisConfigMissing       = errors.New("分析报表配置不存在，请先在设计器中保存")
	ErrAnalysisDataSetNotExists    = errors.New("分析报表关联的数据集不存在")
	ErrAnalysisDataSetDisabled     = errors.New("分析报表关联的数据集已禁用")
	ErrAnalysisConfigJSONInvalid   = errors.New("分析报表配置JSON格式错误")
	ErrAnalysisFieldInvalid        = errors.New("配置字段不在数据集字段列表中")
	ErrAnalysisCopySourceMissing   = errors.New("复制源分析报表不存在")
)
