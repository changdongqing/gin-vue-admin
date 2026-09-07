package report

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	repModel "github.com/flipped-aurora/gin-vue-admin/server/model/report"
	repReq "github.com/flipped-aurora/gin-vue-admin/server/model/report/request"
	repService "github.com/flipped-aurora/gin-vue-admin/server/service/report"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DataSetApi struct{}

// CreateDataSet 创建数据集
// @Tags      报表平台
// @Summary   创建数据集（主子表事务保存/SQL防注入校验/编码查重）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body repReq.SaveDataSetReq true "数据集（含参数/转换）"
// @Success   200 {object} response.Response{data=uint,msg=string}
// @Router    /report/dataSet/createDataSet [post]
func (api *DataSetApi) CreateDataSet(c *gin.Context) {
	var req repReq.SaveDataSetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	id, err := DataSetService.CreateDataSet(&req, utils.GetUserInfo(c).Username)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(id, "创建成功", c)
}

// UpdateDataSet 更新数据集
// @Tags      报表平台
// @Summary   更新数据集（主表更新+参数/转换先删后插全量重建）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body repReq.SaveDataSetReq true "数据集（含 ID 与参数/转换）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /report/dataSet/updateDataSet [put]
func (api *DataSetApi) UpdateDataSet(c *gin.Context) {
	var req repReq.SaveDataSetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := DataSetService.UpdateDataSet(&req, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteDataSet 删除数据集
// @Tags      报表平台
// @Summary   删除数据集（被报表引用拒绝/级联子表）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "数据集ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /report/dataSet/deleteDataSet [delete]
func (api *DataSetApi) DeleteDataSet(c *gin.Context) {
	var req repReq.DataSetOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := DataSetService.DeleteDataSet(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// FindDataSet 数据集详情
// @Tags      报表平台
// @Summary   数据集详情（含参数与转换）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "数据集ID"
// @Success   200 {object} response.Response{data=report.DataSetDetail,msg=string}
// @Router    /report/dataSet/findDataSet [get]
func (api *DataSetApi) FindDataSet(c *gin.Context) {
	var req repReq.DataSetOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	detail, err := DataSetService.GetDataSet(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

// GetDataSetList 分页查询数据集
// @Tags      报表平台
// @Summary   分页查询数据集（keyword/setType/enableFlag）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=response.PageResult,msg=string}
// @Router    /report/dataSet/getDataSetList [get]
func (api *DataSetApi) GetDataSetList(c *gin.Context) {
	var info repReq.SearchDataSet
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
	list, total, err := DataSetService.GetDataSetList(info)
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

// GetDataSetAll 已启用数据集全量
// @Tags      报表平台
// @Summary   已启用数据集全量下拉（03/05 绑定用）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=[]report.ReportDataSet,msg=string}
// @Router    /report/dataSet/getDataSetAll [get]
func (api *DataSetApi) GetDataSetAll(c *gin.Context) {
	list, err := DataSetService.GetEnabledDataSetAll()
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// TestDataSetPreview 测试预览
// @Tags      报表平台
// @Summary   测试预览（两场景：已保存数据集/编辑中即时测试；服务端分页）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body repReq.TestPreviewReq true "测试预览请求"
// @Success   200 {object} response.Response{data=object,msg=string}
// @Router    /report/dataSet/testDataSetPreview [post]
func (api *DataSetApi) TestDataSetPreview(c *gin.Context) {
	var req repReq.TestPreviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.PageNo < 1 {
		req.PageNo = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	var (
		setType, sourceCode, dynSentence string
		transforms                       []repModel.ReportDataSetTransform
		paramValues                      map[string]interface{}
	)
	if req.SetCode != "" {
		set, tfs, params, err := DataSetService.LoadForQuery(req.SetCode)
		if err != nil {
			response.FailWithMessage("测试失败:"+err.Error(), c)
			return
		}
		setType, sourceCode, dynSentence, transforms = set.SetType, set.SourceCode, set.DynSentence, tfs
		resolved, err := repService.ResolveSetParam(params, req.ParamValues)
		if err != nil {
			response.FailWithMessage("测试失败:"+err.Error(), c)
			return
		}
		paramValues = resolved
	} else {
		// 编辑中即时测试：前端按 sampleItem 构建 paramValues，后端直用
		if req.SetType == "" || req.DynSentence == "" {
			response.FailWithMessage("测试失败:setType 与 dynSentence 不能为空", c)
			return
		}
		setType, sourceCode, dynSentence = req.SetType, req.SourceCode, req.DynSentence
		paramValues = req.ParamValues
		if paramValues == nil {
			paramValues = map[string]interface{}{}
		}
	}
	result, total, err := QueryServiceApp.QueryPage(setType, sourceCode, dynSentence, paramValues, req.PageNo, req.PageSize, transforms)
	if err != nil {
		response.FailWithMessage("测试失败:"+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{
		"columns":  result.Columns,
		"rows":     result.Rows,
		"total":    total,
		"pageNo":   req.PageNo,
		"pageSize": req.PageSize,
	}, c)
}
