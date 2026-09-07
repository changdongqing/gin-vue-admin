package report

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/report"
	repReq "github.com/flipped-aurora/gin-vue-admin/server/model/report/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ExcelReportApi struct{}

// CreateExcelReport 创建 Excel 报表元数据
// @Tags      报表平台
// @Summary   创建 Excel 报表（编码查重）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body report.ReportExcelReport true "Excel 报表"
// @Success   200 {object} response.Response{msg=string}
// @Router    /report/excelReport/createExcelReport [post]
func (api *ExcelReportApi) CreateExcelReport(c *gin.Context) {
	var p report.ReportExcelReport
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ExcelReportService.CreateReport(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateExcelReport 更新 Excel 报表元数据
// @Tags      报表平台
// @Summary   更新 Excel 报表（编码不可改）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body report.ReportExcelReport true "Excel 报表（含 ID）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /report/excelReport/updateExcelReport [put]
func (api *ExcelReportApi) UpdateExcelReport(c *gin.Context) {
	var p report.ReportExcelReport
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ExcelReportService.UpdateReport(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteExcelReport 删除 Excel 报表
// @Tags      报表平台
// @Summary   删除 Excel 报表（级联软删模板）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "报表ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /report/excelReport/deleteExcelReport [delete]
func (api *ExcelReportApi) DeleteExcelReport(c *gin.Context) {
	var req repReq.ExcelReportOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ExcelReportService.DeleteReport(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// FindExcelReport 报表详情（元数据+模板）
// @Tags      报表平台
// @Summary   Excel 报表详情（含 setCodes/setParam/jsonStr）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "报表ID"
// @Success   200 {object} response.Response{data=report.ExcelReportDetail,msg=string}
// @Router    /report/excelReport/findExcelReport [get]
func (api *ExcelReportApi) FindExcelReport(c *gin.Context) {
	var req repReq.ExcelReportOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	detail, err := ExcelReportService.GetReport(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

// GetExcelReportList 分页查询 Excel 报表
// @Tags      报表平台
// @Summary   分页查询（keyword/reportGroup；reportCode 精确过滤；不含 jsonStr 大字段）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=response.PageResult,msg=string}
// @Router    /report/excelReport/getExcelReportList [get]
func (api *ExcelReportApi) GetExcelReportList(c *gin.Context) {
	var info repReq.SearchExcelReport
	if err := c.ShouldBindQuery(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if info.PageSize <= 0 {
		info.PageSize = 10
	}
	if info.Page <= 0 {
		info.Page = 1
	}
	list, total, err := ExcelReportService.GetReportList(info)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     info.Page,
		PageSize: info.PageSize,
	}, "获取成功", c)
}

// GetExcelReportAll 已启用 Excel 报表全量
// @Tags      报表平台
// @Summary   已启用 Excel 报表全量下拉（预留）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=[]report.ReportExcelReport,msg=string}
// @Router    /report/excelReport/getExcelReportAll [get]
func (api *ExcelReportApi) GetExcelReportAll(c *gin.Context) {
	list, err := ExcelReportService.GetEnabledReportAll()
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// CopyExcelReport 复制报表
// @Tags      报表平台
// @Summary   复制报表（元数据+模板深拷贝，新编码必填）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body repReq.CopyExcelReportReq true "复制请求"
// @Success   200 {object} response.Response{data=uint,msg=string}
// @Router    /report/excelReport/copyExcelReport [post]
func (api *ExcelReportApi) CopyExcelReport(c *gin.Context) {
	var req repReq.CopyExcelReportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	id, err := ExcelReportService.CopyReport(&req, utils.GetUserInfo(c).Username)
	if err != nil {
		global.GVA_LOG.Error("复制失败!", zap.Error(err))
		response.FailWithMessage("复制失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(id, "复制成功", c)
}

// SaveExcelTemplate 保存模板
// @Tags      报表平台
// @Summary   保存模板（仅 jsonStr/setParam；setCodes 由 bind 接口独立维护）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body repReq.SaveExcelTemplateReq true "模板保存请求"
// @Success   200 {object} response.Response{msg=string}
// @Router    /report/excelReport/saveExcelTemplate [post]
func (api *ExcelReportApi) SaveExcelTemplate(c *gin.Context) {
	var req repReq.SaveExcelTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ExcelReportService.SaveTemplate(&req, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("保存失败!", zap.Error(err))
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// BindExcelReportDataSets 关联数据集
// @Tags      报表平台
// @Summary   关联数据集（仅写模板 set_codes；数据集不存在显式报错）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body repReq.BindDataSetsReq true "关联数据集请求"
// @Success   200 {object} response.Response{msg=string}
// @Router    /report/excelReport/bindExcelReportDataSets [post]
func (api *ExcelReportApi) BindExcelReportDataSets(c *gin.Context) {
	var req repReq.BindDataSetsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ExcelReportService.BindDataSets(&req, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("绑定失败!", zap.Error(err))
		response.FailWithMessage("绑定失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("绑定成功", c)
}

// GetExcelReportDataSetFields 数据集字段列表
// @Tags      报表平台
// @Summary   设计器左栏字段列表（模板 set_codes → 各数据集字段，来自结果案例首行 keys）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     reportCode query string true "报表编码"
// @Success   200 {object} response.Response{data=[]report.DataSetFields,msg=string}
// @Router    /report/excelReport/getExcelReportDataSetFields [get]
func (api *ExcelReportApi) GetExcelReportDataSetFields(c *gin.Context) {
	reportCode := c.Query("reportCode")
	if reportCode == "" {
		response.FailWithMessage("reportCode 不能为空", c)
		return
	}
	fields, err := ExcelReportService.GetReportDataSetFields(reportCode)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(fields, c)
}

// PreviewExcelReport 分页渲染预览
// @Tags      报表平台
// @Summary   Excel 报表预览渲染（主数据集服务端分页；snapshot 为渲染后快照对象）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body repReq.PreviewExcelReportReq true "预览请求（paramValues 平铺）"
// @Success   200 {object} response.Response{data=object,msg=string}
// @Router    /report/excelReport/previewExcelReport [post]
func (api *ExcelReportApi) PreviewExcelReport(c *gin.Context) {
	var req repReq.PreviewExcelReportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, err := RenderServiceApp.Render(req.ReportCode, req.ParamValues, req.PageNo, req.PageSize)
	if err != nil {
		global.GVA_LOG.Error("预览失败!", zap.Error(err))
		response.FailWithMessage("预览失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"snapshot": result.Snapshot, "total": result.Total}, "获取成功", c)
}

// GetExcelReportParamDefs 关联数据集参数定义聚合
// @Tags      报表平台
// @Summary   报表关联的全部数据集参数平铺返回（跨数据集按参数名去重，同名保留先出现者）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     reportCode query string true "报表编码"
// @Success   200 {object} response.Response{data=[]report.ReportParamDef,msg=string}
// @Router    /report/excelReport/getExcelReportParamDefs [get]
func (api *ExcelReportApi) GetExcelReportParamDefs(c *gin.Context) {
	reportCode := c.Query("reportCode")
	if reportCode == "" {
		response.FailWithMessage("reportCode 不能为空", c)
		return
	}
	defs, err := RenderServiceApp.GetParamDefs(reportCode)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(defs, c)
}
