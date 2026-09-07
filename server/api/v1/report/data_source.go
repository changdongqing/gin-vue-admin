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

type DataSourceApi struct{}

// CreateDataSource 创建数据源
// @Tags      报表平台
// @Summary   创建数据源（类型合法/编码查重/密码AES加密）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body report.ReportDataSource true "数据源"
// @Success   200 {object} response.Response{msg=string}
// @Router    /report/dataSource/createDataSource [post]
func (api *DataSourceApi) CreateDataSource(c *gin.Context) {
	var p report.ReportDataSource
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := DataSourceService.CreateDataSource(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateDataSource 更新数据源
// @Tags      报表平台
// @Summary   更新数据源（类型不可变/密码留空或******保留原密文/销毁旧连接池）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body report.ReportDataSource true "数据源（含 ID）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /report/dataSource/updateDataSource [put]
func (api *DataSourceApi) UpdateDataSource(c *gin.Context) {
	var p report.ReportDataSource
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := DataSourceService.UpdateDataSource(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteDataSource 删除数据源
// @Tags      报表平台
// @Summary   删除数据源（被数据集引用拒绝/销毁连接池）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "数据源ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /report/dataSource/deleteDataSource [delete]
func (api *DataSourceApi) DeleteDataSource(c *gin.Context) {
	var req repReq.DataSourceOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := DataSourceService.DeleteDataSource(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// FindDataSource 数据源详情
// @Tags      报表平台
// @Summary   数据源详情（密码脱敏返回）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "数据源ID"
// @Success   200 {object} response.Response{data=report.ReportDataSource,msg=string}
// @Router    /report/dataSource/findDataSource [get]
func (api *DataSourceApi) FindDataSource(c *gin.Context) {
	var req repReq.DataSourceOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	p, err := DataSourceService.GetDataSource(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(p, c)
}

// GetDataSourceList 分页查询数据源
// @Tags      报表平台
// @Summary   分页查询数据源（keyword/sourceType/enableFlag；密码脱敏）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=response.PageResult,msg=string}
// @Router    /report/dataSource/getDataSourceList [get]
func (api *DataSourceApi) GetDataSourceList(c *gin.Context) {
	var info repReq.SearchDataSource
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
	list, total, err := DataSourceService.GetDataSourceList(info)
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

// GetDataSourceAll 已启用数据源全量
// @Tags      报表平台
// @Summary   已启用数据源全量下拉（不含连接配置）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=[]report.ReportDataSource,msg=string}
// @Router    /report/dataSource/getDataSourceAll [get]
func (api *DataSourceApi) GetDataSourceAll(c *gin.Context) {
	list, err := DataSourceService.GetEnabledDataSourceAll()
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// TestConnection 测试连接
// @Tags      报表平台
// @Summary   测试连接（临时连接执行验证SQL，不写池缓存，不保存）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body repReq.TestConnectionReq true "测试连接请求"
// @Success   200 {object} response.Response{msg=string}
// @Router    /report/dataSource/testDataSourceConnection [post]
func (api *DataSourceApi) TestConnection(c *gin.Context) {
	var req repReq.TestConnectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := DataSourceService.TestConnection(req.SourceType, req.SourceConfig); err != nil {
		response.FailWithMessage("连接失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("连接成功", c)
}
