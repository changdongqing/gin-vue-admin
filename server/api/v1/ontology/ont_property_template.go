package ontology

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
	ontReq "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PropertyTemplateApi struct{}

// CreatePropertyTemplate 创建属性模板
// @Tags      本体属性模板
// @Summary   创建属性模板
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntPropertyTemplate true "属性模板（source 由服务端强制 custom）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/propertyTemplate/createPropertyTemplate [post]
func (api *PropertyTemplateApi) CreatePropertyTemplate(c *gin.Context) {
	var t ontology.OntPropertyTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := PropertyTemplateService.CreatePropertyTemplate(&t, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdatePropertyTemplate 更新属性模板
// @Tags      本体属性模板
// @Summary   更新属性模板（builtin 拒绝）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntPropertyTemplate true "属性模板（含 ID）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/propertyTemplate/updatePropertyTemplate [put]
func (api *PropertyTemplateApi) UpdatePropertyTemplate(c *gin.Context) {
	var t ontology.OntPropertyTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := PropertyTemplateService.UpdatePropertyTemplate(&t, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeletePropertyTemplate 删除属性模板
// @Tags      本体属性模板
// @Summary   删除属性模板（builtin 拒绝，软删除）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "属性模板ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/propertyTemplate/deletePropertyTemplate [delete]
func (api *PropertyTemplateApi) DeletePropertyTemplate(c *gin.Context) {
	var req ontReq.PropertyTemplateOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := PropertyTemplateService.DeletePropertyTemplate(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DisablePropertyTemplate 弃用/取消弃用
// @Tags      本体属性模板
// @Summary   弃用/取消弃用属性模板（幂等切换，builtin 也可弃用）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "属性模板ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/propertyTemplate/disablePropertyTemplate [put]
func (api *PropertyTemplateApi) DisablePropertyTemplate(c *gin.Context) {
	var req ontReq.PropertyTemplateOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := PropertyTemplateService.DisablePropertyTemplate(req.ID, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("弃用切换失败!", zap.Error(err))
		response.FailWithMessage("弃用切换失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}

// FindPropertyTemplate 查询属性模板详情
// @Tags      本体属性模板
// @Summary   查询属性模板详情
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "属性模板ID"
// @Success   200 {object} response.Response{data=ontology.OntPropertyTemplate,msg=string}
// @Router    /ontology/propertyTemplate/findPropertyTemplate [get]
func (api *PropertyTemplateApi) FindPropertyTemplate(c *gin.Context) {
	var req ontReq.PropertyTemplateOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	t, err := PropertyTemplateService.GetPropertyTemplate(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(t, c)
}

// GetPropertyTemplateList 分页查询属性模板
// @Tags      本体属性模板
// @Summary   分页查询属性模板（keyword 匹配编码/显示名/别名）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     page query int false "页码"
// @Param     pageSize query int false "每页大小"
// @Param     keyword query string false "关键字"
// @Success   200 {object} response.Response{data=response.PageResult,msg=string}
// @Router    /ontology/propertyTemplate/getPropertyTemplateList [get]
func (api *PropertyTemplateApi) GetPropertyTemplateList(c *gin.Context) {
	var info ontReq.SearchPropertyTemplate
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
	list, total, err := PropertyTemplateService.GetPropertyTemplateList(info)
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

// GetPropertyTemplateAll 全量下拉
// @Tags      本体属性模板
// @Summary   全量属性模板（下拉/选择器，排除弃用）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     kind query string false "属性类型 datatype/object"
// @Success   200 {object} response.Response{data=[]ontology.OntPropertyTemplate,msg=string}
// @Router    /ontology/propertyTemplate/getPropertyTemplateAll [get]
func (api *PropertyTemplateApi) GetPropertyTemplateAll(c *gin.Context) {
	kind := c.Query("kind")
	list, err := PropertyTemplateService.GetPropertyTemplateAll(kind)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// CheckPropertyTemplateCode 编码查重
// @Tags      本体属性模板
// @Summary   属性模板编码查重
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     templateCode query string true "模板编码"
// @Param     excludeId query uint false "排除的模板ID（编辑时传自身）"
// @Success   200 {object} response.Response{data=map[string]bool,msg=string}
// @Router    /ontology/propertyTemplate/checkPropertyTemplateCode [get]
func (api *PropertyTemplateApi) CheckPropertyTemplateCode(c *gin.Context) {
	var req ontReq.CheckPropertyTemplateCode
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	usable, err := PropertyTemplateService.IsTemplateCodeUnique(req.TemplateCode, req.ExcludeID)
	if err != nil {
		global.GVA_LOG.Error("编码查重失败!", zap.Error(err))
		response.FailWithMessage("编码查重失败:"+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"usable": usable}, c)
}

// PromotePropertyTemplate 提升（FR-6 预留，供建模侧调用）
// @Tags      本体属性模板
// @Summary   属性快照提升为自定义模板（重复编码返回业务错误不覆盖）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontReq.PromotePropertyTemplate true "属性快照"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/propertyTemplate/promotePropertyTemplate [post]
func (api *PropertyTemplateApi) PromotePropertyTemplate(c *gin.Context) {
	var req ontReq.PromotePropertyTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := PropertyTemplateService.PromotePropertyTemplate(req, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("提升失败!", zap.Error(err))
		response.FailWithMessage("提升失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("提升成功", c)
}
