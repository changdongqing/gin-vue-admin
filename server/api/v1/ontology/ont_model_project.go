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

type ModelProjectApi struct{}

// CreateModelProject 创建本体项目
// @Tags      本体建模
// @Summary   创建本体项目（编码查重/默认值兜底/方案A拒绝）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntModelProject true "本体项目"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/modelProject/createModelProject [post]
func (api *ModelProjectApi) CreateModelProject(c *gin.Context) {
	var p ontology.OntModelProject
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ModelProjectService.CreateModelProject(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateModelProject 更新本体项目
// @Tags      本体建模
// @Summary   更新本体项目（archived 拒绝/状态机单向/编码查重排除自身）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntModelProject true "本体项目（含 ID）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/modelProject/updateModelProject [put]
func (api *ModelProjectApi) UpdateModelProject(c *gin.Context) {
	var p ontology.OntModelProject
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ModelProjectService.UpdateModelProject(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteModelProject 删除本体项目
// @Tags      本体建模
// @Summary   删除本体项目（archived 拒绝/项目下存在本体类拒绝）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "项目ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/modelProject/deleteModelProject [delete]
func (api *ModelProjectApi) DeleteModelProject(c *gin.Context) {
	var req ontReq.ModelProjectOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ModelProjectService.DeleteModelProject(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// FindModelProject 本体项目详情
// @Tags      本体建模
// @Summary   本体项目详情
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "项目ID"
// @Success   200 {object} response.Response{data=ontology.OntModelProject,msg=string}
// @Router    /ontology/modelProject/findModelProject [get]
func (api *ModelProjectApi) FindModelProject(c *gin.Context) {
	var req ontReq.ModelProjectOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	p, err := ModelProjectService.GetModelProject(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(p, c)
}

// GetModelProjectList 分页查询本体项目
// @Tags      本体建模
// @Summary   分页查询本体项目（keyword 匹配编码/名称 + status）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=response.PageResult,msg=string}
// @Router    /ontology/modelProject/getModelProjectList [get]
func (api *ModelProjectApi) GetModelProjectList(c *gin.Context) {
	var info ontReq.SearchModelProject
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
	list, total, err := ModelProjectService.GetModelProjectList(info)
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

// GetModelProjectAll 全量本体项目
// @Tags      本体建模
// @Summary   全量本体项目（类建模/外部模块关联下拉数据源）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=[]ontology.OntModelProject,msg=string}
// @Router    /ontology/modelProject/getModelProjectAll [get]
func (api *ModelProjectApi) GetModelProjectAll(c *gin.Context) {
	list, err := ModelProjectService.GetModelProjectAll()
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// CheckModelProjectCode 项目编码查重
// @Tags      本体建模
// @Summary   项目编码查重（data=true 表示已存在）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     projectCode query string true "项目编码"
// @Param     excludeId query uint false "排除的项目ID（编辑态）"
// @Success   200 {object} response.Response{data=bool,msg=string}
// @Router    /ontology/modelProject/checkModelProjectCode [get]
func (api *ModelProjectApi) CheckModelProjectCode(c *gin.Context) {
	var req ontReq.CheckModelProjectCodeOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	exists, err := ModelProjectService.CheckModelProjectCode(req.ProjectCode, req.ExcludeId)
	if err != nil {
		global.GVA_LOG.Error("查重失败!", zap.Error(err))
		response.FailWithMessage("查重失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"exists": exists}, "查重成功", c)
}

type ModelPrefixApi struct{}

// GetModelPrefixList 项目前缀列表
// @Tags      本体建模
// @Summary   项目 IRI 前缀列表（默认前缀置顶）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     projectId query uint true "项目ID"
// @Success   200 {object} response.Response{data=[]ontology.OntModelPrefix,msg=string}
// @Router    /ontology/modelPrefix/getModelPrefixList [get]
func (api *ModelPrefixApi) GetModelPrefixList(c *gin.Context) {
	var req ontReq.PrefixOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := ModelPrefixService.GetModelPrefixList(req.ProjectId)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// CreateModelPrefix 新增项目前缀
// @Tags      本体建模
// @Summary   新增项目前缀（NCName/同项目唯一/默认前缀唯一）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntModelPrefix true "前缀（含 projectId）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/modelPrefix/createModelPrefix [post]
func (api *ModelPrefixApi) CreateModelPrefix(c *gin.Context) {
	var p ontology.OntModelPrefix
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ModelPrefixService.CreateModelPrefix(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateModelPrefix 更新项目前缀
// @Tags      本体建模
// @Summary   更新项目前缀（归属/NCName/唯一/默认唯一校验）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntModelPrefix true "前缀（含 ID+projectId）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/modelPrefix/updateModelPrefix [put]
func (api *ModelPrefixApi) UpdateModelPrefix(c *gin.Context) {
	var p ontology.OntModelPrefix
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ModelPrefixService.UpdateModelPrefix(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteModelPrefix 删除项目前缀
// @Tags      本体建模
// @Summary   删除项目前缀（归属/归档校验）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "前缀ID"
// @Param     projectId query uint true "项目ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/modelPrefix/deleteModelPrefix [delete]
func (api *ModelPrefixApi) DeleteModelPrefix(c *gin.Context) {
	var req ontReq.PrefixOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.ID == 0 {
		response.FailWithMessage("前缀ID不能为空", c)
		return
	}
	if err := ModelPrefixService.DeleteModelPrefix(req.ID, req.ProjectId); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
