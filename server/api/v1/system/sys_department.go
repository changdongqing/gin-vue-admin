package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DepartmentApi struct{}

// CreateDepartment
// @Tags      Department
// @Summary   创建部门
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysDepartment  true  "部门信息"
// @Success   200   {object}  response.Response{msg=string}
// @Router    /department/createDepartment [post]
func (departmentApi *DepartmentApi) CreateDepartment(c *gin.Context) {
	var department system.SysDepartment
	if err := c.ShouldBindJSON(&department); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := departmentService.CreateDepartment(&department); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteDepartment
// @Tags      Department
// @Summary   删除部门
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID  query  uint  true  "部门ID"
// @Success   200  {object}  response.Response{msg=string}
// @Router    /department/deleteDepartment [delete]
func (departmentApi *DepartmentApi) DeleteDepartment(c *gin.Context) {
	var req struct {
		ID uint `json:"ID" form:"ID"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := departmentService.DeleteDepartment(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// UpdateDepartment
// @Tags      Department
// @Summary   更新部门
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysDepartment  true  "部门信息"
// @Success   200   {object}  response.Response{msg=string}
// @Router    /department/updateDepartment [put]
func (departmentApi *DepartmentApi) UpdateDepartment(c *gin.Context) {
	var department system.SysDepartment
	if err := c.ShouldBindJSON(&department); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := departmentService.UpdateDepartment(&department); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindDepartment
// @Tags      Department
// @Summary   用id查询部门
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID  query  uint  true  "部门ID"
// @Success   200  {object}  response.Response{data=system.SysDepartment,msg=string}
// @Router    /department/findDepartment [get]
func (departmentApi *DepartmentApi) FindDepartment(c *gin.Context) {
	var req struct {
		ID uint `json:"ID" form:"ID"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	department, err := departmentService.GetDepartment(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(department, c)
}

// GetDepartmentList
// @Tags      Department
// @Summary   获取部门树
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     companyId  query  uint    false  "挂靠公司ID(过滤该公司部门树)"
// @Param     name       query  string  false  "部门名称关键字(平铺返回命中行)"
// @Success   200   {object}  response.Response{data=[]system.SysDepartment,msg=string}
// @Router    /department/getDepartmentList [get]
func (departmentApi *DepartmentApi) GetDepartmentList(c *gin.Context) {
	var search systemReq.SysDepartmentSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := departmentService.GetDepartmentTree(search)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list}, "获取成功", c)
}
