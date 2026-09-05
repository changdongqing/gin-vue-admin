package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CompanyApi struct{}

// CreateCompany
// @Tags      Company
// @Summary   创建公司
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysCompany  true  "公司信息"
// @Success   200   {object}  response.Response{msg=string}
// @Router    /company/createCompany [post]
func (companyApi *CompanyApi) CreateCompany(c *gin.Context) {
	var company system.SysCompany
	if err := c.ShouldBindJSON(&company); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := companyService.CreateCompany(&company); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteCompany
// @Tags      Company
// @Summary   删除公司
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID  query  uint  true  "公司ID"
// @Success   200  {object}  response.Response{msg=string}
// @Router    /company/deleteCompany [delete]
func (companyApi *CompanyApi) DeleteCompany(c *gin.Context) {
	var req struct {
		ID uint `json:"ID" form:"ID"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := companyService.DeleteCompany(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// UpdateCompany
// @Tags      Company
// @Summary   更新公司
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysCompany  true  "公司信息"
// @Success   200   {object}  response.Response{msg=string}
// @Router    /company/updateCompany [put]
func (companyApi *CompanyApi) UpdateCompany(c *gin.Context) {
	var company system.SysCompany
	if err := c.ShouldBindJSON(&company); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := companyService.UpdateCompany(&company); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindCompany
// @Tags      Company
// @Summary   用id查询公司
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID  query  uint  true  "公司ID"
// @Success   200  {object}  response.Response{data=system.SysCompany,msg=string}
// @Router    /company/findCompany [get]
func (companyApi *CompanyApi) FindCompany(c *gin.Context) {
	var req struct {
		ID uint `json:"ID" form:"ID"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	company, err := companyService.GetCompany(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(company, c)
}

// GetCompanyList
// @Tags      Company
// @Summary   获取公司树
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     name  query  string  false  "公司名称关键字(平铺返回命中行)"
// @Success   200   {object}  response.Response{data=[]system.SysCompany,msg=string}
// @Router    /company/getCompanyList [get]
func (companyApi *CompanyApi) GetCompanyList(c *gin.Context) {
	var search systemReq.SysCompanySearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := companyService.GetCompanyTree(search)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list}, "获取成功", c)
}
