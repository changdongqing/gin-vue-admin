package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PostApi struct{}

// CreatePost
// @Tags      Post
// @Summary   创建岗位
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysPost  true  "岗位信息"
// @Success   200   {object}  response.Response{msg=string}
// @Router    /post/createPost [post]
func (postApi *PostApi) CreatePost(c *gin.Context) {
	var post system.SysPost
	if err := c.ShouldBindJSON(&post); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := postService.CreatePost(&post); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeletePost
// @Tags      Post
// @Summary   删除岗位（有关联员工时拒绝）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID  query  uint  true  "岗位ID"
// @Success   200  {object}  response.Response{msg=string}
// @Router    /post/deletePost [delete]
func (postApi *PostApi) DeletePost(c *gin.Context) {
	var req struct {
		ID uint `json:"ID" form:"ID"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := postService.DeletePost(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// UpdatePost
// @Tags      Post
// @Summary   更新岗位
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      system.SysPost  true  "岗位信息"
// @Success   200   {object}  response.Response{msg=string}
// @Router    /post/updatePost [put]
func (postApi *PostApi) UpdatePost(c *gin.Context) {
	var post system.SysPost
	if err := c.ShouldBindJSON(&post); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := postService.UpdatePost(&post); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindPost
// @Tags      Post
// @Summary   用id查询岗位
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID  query  uint  true  "岗位ID"
// @Success   200  {object}  response.Response{data=system.SysPost,msg=string}
// @Router    /post/findPost [get]
func (postApi *PostApi) FindPost(c *gin.Context) {
	var req struct {
		ID uint `json:"ID" form:"ID"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	post, err := postService.GetPost(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(post, c)
}

// GetPostList
// @Tags      Post
// @Summary   分页获取岗位列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      systemReq.SysPostSearch  true  "页码, 每页大小, 关键字, 岗位类型, 状态"
// @Success   200   {object}  response.Response{data=response.PageResult,msg=string}
// @Router    /post/getPostList [post]
func (postApi *PostApi) GetPostList(c *gin.Context) {
	var search systemReq.SysPostSearch
	if err := c.ShouldBindJSON(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := postService.GetPostList(search)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     search.Page,
		PageSize: search.PageSize,
	}, "获取成功", c)
}

// GetPostListAll
// @Tags      Post
// @Summary   获取全部启用岗位（不分页，下拉/多选用）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200  {object}  response.Response{data=object,msg=string}
// @Router    /post/getPostListAll [get]
func (postApi *PostApi) GetPostListAll(c *gin.Context) {
	list, err := postService.GetPostListAll()
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list}, "获取成功", c)
}

// GetPostUsers
// @Tags      Post
// @Summary   获取岗位已关联员工（全量，分配员工弹窗回显）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     postId  query  uint  true  "岗位ID"
// @Success   200  {object}  response.Response{data=object,msg=string}
// @Router    /post/getPostUsers [get]
func (postApi *PostApi) GetPostUsers(c *gin.Context) {
	var req struct {
		PostId uint `json:"postId" form:"postId"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := postService.GetPostUsers(req.PostId)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list}, "获取成功", c)
}

// SetPostUsers
// @Tags      Post
// @Summary   岗位绑定员工（全量覆盖）
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      systemReq.SetPostUsersReq  true  "岗位ID + 用户ID列表"
// @Success   200  {object}  response.Response{msg=string}
// @Router    /post/setPostUsers [post]
func (postApi *PostApi) SetPostUsers(c *gin.Context) {
	var req systemReq.SetPostUsersReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := postService.SetPostUsers(req.PostId, req.UserIds); err != nil {
		global.GVA_LOG.Error("绑定失败!", zap.Error(err))
		response.FailWithMessage("绑定失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("绑定成功", c)
}

// GetUserPosts
// @Tags      Post
// @Summary   获取员工已关联岗位（全量，员工侧回显）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     userId  query  uint  true  "用户ID"
// @Success   200  {object}  response.Response{data=object,msg=string}
// @Router    /post/getUserPosts [get]
func (postApi *PostApi) GetUserPosts(c *gin.Context) {
	var req struct {
		UserId uint `json:"userId" form:"userId"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := postService.GetUserPosts(req.UserId)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list}, "获取成功", c)
}

// SetUserPosts
// @Tags      Post
// @Summary   员工绑定岗位（全量覆盖）
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      systemReq.SetUserPostsReq  true  "用户ID + 岗位ID列表"
// @Success   200  {object}  response.Response{msg=string}
// @Router    /post/setUserPosts [post]
func (postApi *PostApi) SetUserPosts(c *gin.Context) {
	var req systemReq.SetUserPostsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := postService.SetUserPosts(req.UserId, req.PostIds); err != nil {
		global.GVA_LOG.Error("绑定失败!", zap.Error(err))
		response.FailWithMessage("绑定失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("绑定成功", c)
}
