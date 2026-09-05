package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type CompanyRouter struct{}

// InitCompanyRouter 初始化公司管理路由
func (c *CompanyRouter) InitCompanyRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	companyRouter := Router.Group("company").Use(middleware.OperationRecord())
	companyRouterWithoutRecord := Router.Group("company")
	{
		companyRouter.POST("createCompany", companyApi.CreateCompany) // 新建公司
		companyRouter.PUT("updateCompany", companyApi.UpdateCompany)  // 更新公司
		companyRouter.DELETE("deleteCompany", companyApi.DeleteCompany) // 删除公司
	}
	{
		companyRouterWithoutRecord.GET("findCompany", companyApi.FindCompany)       // 根据ID获取公司
		companyRouterWithoutRecord.GET("getCompanyList", companyApi.GetCompanyList) // 获取公司树
	}
}
