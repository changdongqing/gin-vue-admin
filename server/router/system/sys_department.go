package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type DepartmentRouter struct{}

// InitDepartmentRouter 初始化部门管理路由
func (d *DepartmentRouter) InitDepartmentRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	departmentRouter := Router.Group("department").Use(middleware.OperationRecord())
	departmentRouterWithoutRecord := Router.Group("department")
	{
		departmentRouter.POST("createDepartment", departmentApi.CreateDepartment) // 新建部门
		departmentRouter.PUT("updateDepartment", departmentApi.UpdateDepartment)  // 更新部门
		departmentRouter.DELETE("deleteDepartment", departmentApi.DeleteDepartment) // 删除部门
	}
	{
		departmentRouterWithoutRecord.GET("findDepartment", departmentApi.FindDepartment)       // 根据ID获取部门
		departmentRouterWithoutRecord.GET("getDepartmentList", departmentApi.GetDepartmentList) // 获取部门树
	}
}
