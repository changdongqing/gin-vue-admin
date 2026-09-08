package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type PostRouter struct{}

// InitPostRouter 初始化岗位管理路由
func (p *PostRouter) InitPostRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	postRouter := Router.Group("post").Use(middleware.OperationRecord())
	postRouterWithoutRecord := Router.Group("post")
	{
		postRouter.POST("createPost", postApi.CreatePost)     // 新建岗位
		postRouter.PUT("updatePost", postApi.UpdatePost)      // 更新岗位
		postRouter.DELETE("deletePost", postApi.DeletePost)   // 删除岗位
		postRouter.POST("setPostUsers", postApi.SetPostUsers) // 岗位绑定员工（全量覆盖）
		postRouter.POST("setUserPosts", postApi.SetUserPosts) // 员工绑定岗位（全量覆盖）
	}
	{
		postRouterWithoutRecord.GET("findPost", postApi.FindPost)             // 根据ID获取岗位
		postRouterWithoutRecord.POST("getPostList", postApi.GetPostList)      // 分页获取岗位列表
		postRouterWithoutRecord.GET("getPostListAll", postApi.GetPostListAll) // 全部启用岗位（下拉）
		postRouterWithoutRecord.GET("getPostUsers", postApi.GetPostUsers)     // 岗位已关联员工
		postRouterWithoutRecord.GET("getUserPosts", postApi.GetUserPosts)     // 员工已关联岗位
	}
}
