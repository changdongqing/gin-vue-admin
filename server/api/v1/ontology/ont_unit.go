package ontology

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
	ontReq "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type UnitApi struct{}

// CreateUnit 创建单位
// @Tags      本体单位
// @Summary   创建单位（量纲存在 + 双唯一校验，source 强制 custom）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntUnit true "单位"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/unit/createUnit [post]
func (api *UnitApi) CreateUnit(c *gin.Context) {
	var u ontology.OntUnit
	if err := c.ShouldBindJSON(&u); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := UnitService.CreateUnit(&u, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateUnit 更新单位
// @Tags      本体单位
// @Summary   更新单位（builtin 拒绝）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntUnit true "单位（含 ID）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/unit/updateUnit [put]
func (api *UnitApi) UpdateUnit(c *gin.Context) {
	var u ontology.OntUnit
	if err := c.ShouldBindJSON(&u); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := UnitService.UpdateUnit(&u, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteUnit 删除单位
// @Tags      本体单位
// @Summary   删除单位（builtin 拒绝，软删除）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "单位ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/unit/deleteUnit [delete]
func (api *UnitApi) DeleteUnit(c *gin.Context) {
	var req ontReq.UnitOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := UnitService.DeleteUnit(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DisableUnit 停用/启用
// @Tags      本体单位
// @Summary   停用/启用单位（幂等切换，builtin 可停用）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "单位ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/unit/disableUnit [put]
func (api *UnitApi) DisableUnit(c *gin.Context) {
	var req ontReq.UnitOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := UnitService.DisableUnit(req.ID, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("停用切换失败!", zap.Error(err))
		response.FailWithMessage("停用切换失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}

// FindUnit 单位详情
// @Tags      本体单位
// @Summary   单位详情
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "单位ID"
// @Success   200 {object} response.Response{data=ontology.OntUnit,msg=string}
// @Router    /ontology/unit/findUnit [get]
func (api *UnitApi) FindUnit(c *gin.Context) {
	var req ontReq.UnitOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	u, err := UnitService.GetUnit(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(u, c)
}

// GetUnitPage 分页查询单位
// @Tags      本体单位
// @Summary   分页查询单位（keyword 匹配编码/名称/中文名/符号）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=response.PageResult,msg=string}
// @Router    /ontology/unit/getUnitPage [get]
func (api *UnitApi) GetUnitPage(c *gin.Context) {
	var info ontReq.SearchUnit
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
	list, total, err := UnitService.GetUnitPage(info)
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

// GetUnitAll 全量下拉
// @Tags      本体单位
// @Summary   全量单位（仅启用；换算面板/属性模板预设单位用）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     quantityKindCode query string false "量纲标识"
// @Success   200 {object} response.Response{data=[]ontology.OntUnit,msg=string}
// @Router    /ontology/unit/getUnitAll [get]
func (api *UnitApi) GetUnitAll(c *gin.Context) {
	list, err := UnitService.GetUnitAll(c.Query("quantityKindCode"))
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// ConvertUnit 单位换算
// @Tags      本体单位
// @Summary   单位换算（跨量纲/单位不存在返回 result=null + reason，不报错）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     value query string true "数值（字符串承载，避免浮点损失）"
// @Param     fromIri query string true "源单位 QUDT IRI"
// @Param     toIri query string true "目标单位 QUDT IRI"
// @Success   200 {object} response.Response{data=ontology.ConvertUnitResp,msg=string}
// @Router    /ontology/unit/convertUnit [get]
func (api *UnitApi) ConvertUnit(c *gin.Context) {
	var req ontReq.ConvertUnitReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	value, err := decimal.NewFromString(req.Value)
	if err != nil {
		response.FailWithMessage("数值格式不合法:"+err.Error(), c)
		return
	}
	resp := ontology.ConvertUnitResp{}
	resp.Result, resp.Reason, err = UnitService.ConvertValue(value, req.FromIri, req.ToIri)
	if err != nil {
		global.GVA_LOG.Error("换算失败!", zap.Error(err))
		response.FailWithMessage("换算失败:"+err.Error(), c)
		return
	}
	response.OkWithData(resp, c)
}

type QuantityKindApi struct{}

// GetQuantityKindList 量纲列表
// @Tags      本体量纲
// @Summary   量纲全量列表（只读，左树/下拉数据源）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=[]ontology.OntQuantityKind,msg=string}
// @Router    /ontology/quantityKind/getQuantityKindList [get]
func (api *QuantityKindApi) GetQuantityKindList(c *gin.Context) {
	list, err := QuantityKindService.GetQuantityKindList()
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}
