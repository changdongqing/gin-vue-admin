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

type ExtModuleApi struct{}

// GetExtModuleList 模块列表（含 tableCount）
// @Tags      本体建模
// @Summary   外部模块列表
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=[]ontRes.ExtModulePageItem,msg=string}
// @Router    /ontology/extModule/getExtModuleList [get]
func (api *ExtModuleApi) GetExtModuleList(c *gin.Context) {
	list, err := ExtModuleService.GetExtModuleList()
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// CreateExtModule 新增模块
// @Tags      本体建模
// @Summary   新增外部模块（moduleCode 唯一）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntExtModule true "模块"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/extModule/createExtModule [post]
func (api *ExtModuleApi) CreateExtModule(c *gin.Context) {
	var p ontology.OntExtModule
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ExtModuleService.CreateExtModule(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateExtModule 更新模块
// @Tags      本体建模
// @Summary   更新外部模块
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntExtModule true "模块（含 ID）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/extModule/updateExtModule [put]
func (api *ExtModuleApi) UpdateExtModule(c *gin.Context) {
	var p ontology.OntExtModule
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ExtModuleService.UpdateExtModule(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteExtModule 删除模块
// @Tags      本体建模
// @Summary   删除外部模块（有注册表拒绝）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "模块ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/extModule/deleteExtModule [delete]
func (api *ExtModuleApi) DeleteExtModule(c *gin.Context) {
	var req ontReq.ExtModuleOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ExtModuleService.DeleteExtModule(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

type ExtTableApi struct{}

// GetExtTableList 模块下注册表清单
// @Tags      本体建模
// @Summary   模块下注册表清单
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     moduleId query uint true "模块ID"
// @Success   200 {object} response.Response{data=[]ontology.OntExtTable,msg=string}
// @Router    /ontology/extTable/getExtTableList [get]
func (api *ExtTableApi) GetExtTableList(c *gin.Context) {
	var req ontReq.ExtTableListOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := ExtTableService.GetExtTableList(req.ModuleId)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// RegisterExtTables 探测勾选批量注册
// @Tags      本体建模
// @Summary   批量注册物理表（探测回填软删/水位列，已注册跳过）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontReq.RegisterExtTablesOps true "{moduleId, items:[{tableName,displayName,remark}]}"
// @Success   200 {object} response.Response{data=int,msg=string}
// @Router    /ontology/extTable/registerExtTables [post]
func (api *ExtTableApi) RegisterExtTables(c *gin.Context) {
	var req ontReq.RegisterExtTablesOps
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	registered, err := ExtTableService.RegisterExtTables(req, utils.GetUserInfo(c).Username)
	if err != nil {
		global.GVA_LOG.Error("注册失败!", zap.Error(err))
		response.FailWithMessage("注册失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"registered": registered}, "注册成功", c)
}

// DeleteExtTable 删除注册表
// @Tags      本体建模
// @Summary   删除注册表（被绑定引用拒绝）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "注册表ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/extTable/deleteExtTable [delete]
func (api *ExtTableApi) DeleteExtTable(c *gin.Context) {
	var req ontReq.ExtModuleOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ExtTableService.DeleteExtTable(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ProbeExtTables 探测物理表
// @Tags      本体建模
// @Summary   探测 public schema 物理表（注释/已注册标注）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     keyword query string false "表名关键字"
// @Success   200 {object} response.Response{data=[]ontRes.ProbeExtTableItem,msg=string}
// @Router    /ontology/extTable/probeExtTables [get]
func (api *ExtTableApi) ProbeExtTables(c *gin.Context) {
	var req ontReq.ProbeExtTablesOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := ExtTableService.ProbeExtTables(req.Keyword)
	if err != nil {
		global.GVA_LOG.Error("探测失败!", zap.Error(err))
		response.FailWithMessage("探测失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// GetExtTableColumns 探测表列
// @Tags      本体建模
// @Summary   探测表列（绑定表单列下拉）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     tableName query string true "表名"
// @Success   200 {object} response.Response{data=[]ontRes.ExtTableColumnItem,msg=string}
// @Router    /ontology/extTable/getExtTableColumns [get]
func (api *ExtTableApi) GetExtTableColumns(c *gin.Context) {
	var req ontReq.ExtTableColumnsOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := ExtTableService.GetExtTableColumns(req.TableName)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

type ExtBindingApi struct{}

// GetExtBindingList 绑定分页（富化）
// @Tags      本体建模
// @Summary   绑定分页（项目/类/表富化 + 行数计数）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=response.PageResult,msg=string}
// @Router    /ontology/extBinding/getExtBindingList [get]
func (api *ExtBindingApi) GetExtBindingList(c *gin.Context) {
	var info ontReq.SearchExtBinding
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
	list, total, err := ExtBindingService.GetExtBindingList(info)
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

// FindExtBinding 绑定详情（含双子表）
// @Tags      本体建模
// @Summary   绑定详情（主表+子表绑定+属性绑定，编辑回填）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "绑定ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/extBinding/findExtBinding [get]
func (api *ExtBindingApi) FindExtBinding(c *gin.Context) {
	var req ontReq.ExtBindingOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	binding, details, properties, err := ExtBindingService.GetExtBinding(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"binding": binding, "details": details, "properties": properties}, c)
}

// CreateExtBinding 保存草稿
// @Tags      本体建模
// @Summary   新建绑定草稿（返回 ID 供试运行）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontReq.SaveExtBinding true "绑定+子表+属性"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/extBinding/createExtBinding [post]
func (api *ExtBindingApi) CreateExtBinding(c *gin.Context) {
	var req ontReq.SaveExtBinding
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	id, err := ExtBindingService.CreateExtBinding(&req, utils.GetUserInfo(c).Username)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"ID": id}, "创建成功", c)
}

// UpdateExtBinding 更新草稿
// @Tags      本体建模
// @Summary   更新绑定草稿（生效态不可改；整体替换式保存）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontReq.SaveExtBinding true "绑定+子表+属性（含 ID）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/extBinding/updateExtBinding [post]
func (api *ExtBindingApi) UpdateExtBinding(c *gin.Context) {
	var req ontReq.SaveExtBinding
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ExtBindingService.UpdateExtBinding(&req, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteExtBinding 删除绑定
// @Tags      本体建模
// @Summary   删除绑定（同步日志保留）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "绑定ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/extBinding/deleteExtBinding [delete]
func (api *ExtBindingApi) DeleteExtBinding(c *gin.Context) {
	var req ontReq.ExtBindingOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ExtBindingService.DeleteExtBinding(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ChangeExtBindingStatus 生效/停用
// @Tags      本体建模
// @Summary   绑定状态变更（生效走校验链/停用直接置停用）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "绑定ID"
// @Param     status query int true "1生效/2停用"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/extBinding/changeExtBindingStatus [put]
func (api *ExtBindingApi) ChangeExtBindingStatus(c *gin.Context) {
	var req ontReq.ChangeExtBindingStatusOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var err error
	if req.Status == 1 {
		err = ExtBindingService.ActivateExtBinding(req.ID)
	} else {
		err = ExtBindingService.DeactivateExtBinding(req.ID)
	}
	if err != nil {
		global.GVA_LOG.Error("状态变更失败!", zap.Error(err))
		response.FailWithMessage("状态变更失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("状态变更成功", c)
}

type ExtSyncApi struct{}

// DryRunExtSync 试运行
// @Tags      本体建模
// @Summary   试运行（只读预览前 N 行，零写入零日志）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     bindingId query uint true "绑定ID"
// @Param     limit query int false "行数（默认20）"
// @Success   200 {object} response.Response{data=ontRes.DryRunResult,msg=string}
// @Router    /ontology/extSync/dryRunExtSync [post]
func (api *ExtSyncApi) DryRunExtSync(c *gin.Context) {
	var req ontReq.DryRunOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, err := ExtSyncService.DryRunExtBinding(req.BindingId, req.Limit)
	if err != nil {
		global.GVA_LOG.Error("试运行失败!", zap.Error(err))
		response.FailWithMessage("试运行失败:"+err.Error(), c)
		return
	}
	response.OkWithData(result, c)
}

// TriggerExtSync 触发同步
// @Tags      本体建模
// @Summary   触发同步（1全量含孤儿检测/2增量水位）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontReq.TriggerSyncOps true "{bindingId,scope}"
// @Success   200 {object} response.Response{data=ontRes.SyncResult,msg=string}
// @Router    /ontology/extSync/triggerExtSync [post]
func (api *ExtSyncApi) TriggerExtSync(c *gin.Context) {
	var req ontReq.TriggerSyncOps
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, err := ExtSyncService.TriggerExtSync(req, utils.GetUserInfo(c).Username)
	if err != nil {
		global.GVA_LOG.Error("同步失败!", zap.Error(err))
		response.FailWithMessage("同步失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(result, "同步完成", c)
}

// GetExtSyncLogList 同步日志分页
// @Tags      本体建模
// @Summary   同步日志分页
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=response.PageResult,msg=string}
// @Router    /ontology/extSync/getExtSyncLogList [get]
func (api *ExtSyncApi) GetExtSyncLogList(c *gin.Context) {
	var info ontReq.SearchExtSyncLog
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
	list, total, err := ExtSyncService.GetExtSyncLogList(info)
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
