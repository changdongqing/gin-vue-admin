package report

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	repReq "github.com/flipped-aurora/gin-vue-admin/server/model/report/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AnalysisReportApi struct{}

// CreateAnalysisReport 创建分析报表
// @Tags      报表平台
// @Summary   创建分析报表（编码查重/数据集存在且启用）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body repReq.SaveAnalysisReq true "分析报表"
// @Success   200 {object} response.Response{data=uint,msg=string}
// @Router    /report/analysisReport/createAnalysisReport [post]
func (api *AnalysisReportApi) CreateAnalysisReport(c *gin.Context) {
	var req repReq.SaveAnalysisReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	id, err := AnalysisReportService.CreateAnalysisReport(&req, utils.GetUserInfo(c).Username)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(id, "创建成功", c)
}

// UpdateAnalysisReport 更新分析报表
// @Tags      报表平台
// @Summary   更新分析报表（编码不可改/数据集可用性）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body repReq.SaveAnalysisReq true "分析报表（含 ID）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /report/analysisReport/updateAnalysisReport [put]
func (api *AnalysisReportApi) UpdateAnalysisReport(c *gin.Context) {
	var req repReq.SaveAnalysisReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := AnalysisReportService.UpdateAnalysisReport(&req, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteAnalysisReport 删除分析报表
// @Tags      报表平台
// @Summary   删除分析报表（级联删配置）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "报表ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /report/analysisReport/deleteAnalysisReport [delete]
func (api *AnalysisReportApi) DeleteAnalysisReport(c *gin.Context) {
	var req repReq.AnalysisReportOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := AnalysisReportService.DeleteAnalysisReport(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// FindAnalysisReport 详情
// @Tags      报表平台
// @Summary   分析报表详情（元数据+setName+configJson/setParam+params）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "报表ID"
// @Success   200 {object} response.Response{data=report.AnalysisReportDetail,msg=string}
// @Router    /report/analysisReport/findAnalysisReport [get]
func (api *AnalysisReportApi) FindAnalysisReport(c *gin.Context) {
	var req repReq.AnalysisReportOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	detail, err := AnalysisReportService.GetAnalysisReport(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

// GetAnalysisReportByCode 按编码详情
// @Tags      报表平台
// @Summary   按编码查询分析报表详情（预览页/设计器）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     reportCode query string true "报表编码"
// @Success   200 {object} response.Response{data=report.AnalysisReportDetail,msg=string}
// @Router    /report/analysisReport/getAnalysisReportByCode [get]
func (api *AnalysisReportApi) GetAnalysisReportByCode(c *gin.Context) {
	reportCode := c.Query("reportCode")
	if reportCode == "" {
		response.FailWithMessage("reportCode 不能为空", c)
		return
	}
	detail, err := AnalysisReportService.GetAnalysisReportByCode(reportCode)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

// GetAnalysisReportList 分页查询
// @Tags      报表平台
// @Summary   分页查询分析报表（keyword/分组/数据集/状态）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=response.PageResult,msg=string}
// @Router    /report/analysisReport/getAnalysisReportList [get]
func (api *AnalysisReportApi) GetAnalysisReportList(c *gin.Context) {
	var info repReq.SearchAnalysis
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
	list, total, err := AnalysisReportService.GetAnalysisReportList(info)
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

// CopyAnalysisReport 复制分析报表
// @Tags      报表平台
// @Summary   复制分析报表（元数据+配置深拷贝）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body repReq.CopyAnalysisReq true "复制请求"
// @Success   200 {object} response.Response{data=uint,msg=string}
// @Router    /report/analysisReport/copyAnalysisReport [post]
func (api *AnalysisReportApi) CopyAnalysisReport(c *gin.Context) {
	var req repReq.CopyAnalysisReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	id, err := AnalysisReportService.CopyAnalysisReport(&req, utils.GetUserInfo(c).Username)
	if err != nil {
		global.GVA_LOG.Error("复制失败!", zap.Error(err))
		response.FailWithMessage("复制失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(id, "复制成功", c)
}

// SaveAnalysisConfig 保存设计器配置
// @Tags      报表平台
// @Summary   保存分析报表配置（JSON/枚举/字段白名单三层校验 + upsert）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body repReq.SaveAnalysisConfigReq true "配置保存请求"
// @Success   200 {object} response.Response{msg=string}
// @Router    /report/analysisReport/saveAnalysisConfig [post]
func (api *AnalysisReportApi) SaveAnalysisConfig(c *gin.Context) {
	var req repReq.SaveAnalysisConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := AnalysisReportService.SaveConfig(&req, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("保存失败!", zap.Error(err))
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// PreviewAnalysisReport 预览取数
// @Tags      报表平台
// @Summary   分析报表预览取数（明细不聚合/类型推断/超限截断）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body repReq.PreviewAnalysisReq true "预览请求"
// @Success   200 {object} response.Response{data=object,msg=string}
// @Router    /report/analysisReport/previewAnalysisReport [post]
func (api *AnalysisReportApi) PreviewAnalysisReport(c *gin.Context) {
	var req repReq.PreviewAnalysisReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, err := AnalysisReportService.Preview(&req)
	if err != nil {
		global.GVA_LOG.Error("预览失败!", zap.Error(err))
		response.FailWithMessage("预览失败:"+err.Error(), c)
		return
	}
	response.OkWithData(result, c)
}
