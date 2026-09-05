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

type ClassTemplateApi struct{}

// CreateClassTemplate 创建分类模板
// @Tags      本体分类模板
// @Summary   创建分类模板（校验 + 编码生成/校验 + 骨架落库）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntClassTemplate true "分类模板（含 ontClassTemplateRefs 骨架）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/classTemplate/createClassTemplate [post]
func (api *ClassTemplateApi) CreateClassTemplate(c *gin.Context) {
	var t ontology.OntClassTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ClassTemplateService.CreateClassTemplate(&t, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateClassTemplate 更新分类模板
// @Tags      本体分类模板
// @Summary   更新分类模板（骨架 diff + 改父平移子树）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntClassTemplate true "分类模板（含 ID 与 ontClassTemplateRefs 骨架）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/classTemplate/updateClassTemplate [put]
func (api *ClassTemplateApi) UpdateClassTemplate(c *gin.Context) {
	var t ontology.OntClassTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ClassTemplateService.UpdateClassTemplate(&t, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteClassTemplate 删除分类模板
// @Tags      本体分类模板
// @Summary   删除分类模板（有子拒绝；级联软删骨架）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "分类模板ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/classTemplate/deleteClassTemplate [delete]
func (api *ClassTemplateApi) DeleteClassTemplate(c *gin.Context) {
	var req ontReq.ClassTemplateOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ClassTemplateService.DeleteClassTemplate(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DisableClassTemplate 弃用/取消弃用（预留，本期无 UI）
// @Tags      本体分类模板
// @Summary   弃用/取消弃用分类模板（幂等切换）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "分类模板ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/classTemplate/disableClassTemplate [put]
func (api *ClassTemplateApi) DisableClassTemplate(c *gin.Context) {
	var req ontReq.ClassTemplateOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ClassTemplateService.DisableClassTemplate(req.ID, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("弃用切换失败!", zap.Error(err))
		response.FailWithMessage("弃用切换失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}

// FindClassTemplate 分类模板详情
// @Tags      本体分类模板
// @Summary   分类模板详情（含骨架子表，供编辑回填）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "分类模板ID"
// @Success   200 {object} response.Response{data=ontology.OntClassTemplate,msg=string}
// @Router    /ontology/classTemplate/findClassTemplate [get]
func (api *ClassTemplateApi) FindClassTemplate(c *gin.Context) {
	var req ontReq.ClassTemplateOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	t, err := ClassTemplateService.GetClassTemplate(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(t, c)
}

// GetClassTemplateList 扁平全量
// @Tags      本体分类模板
// @Summary   扁平全量分类模板（?treeRoot 过滤，前端组树/树选择器用）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     treeRoot query string false "分类树标识"
// @Success   200 {object} response.Response{data=[]ontology.OntClassTemplate,msg=string}
// @Router    /ontology/classTemplate/getClassTemplateList [get]
func (api *ClassTemplateApi) GetClassTemplateList(c *gin.Context) {
	list, err := ClassTemplateService.GetClassTemplateList(c.Query("treeRoot"))
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// GetClassTemplatePage 分页富化
// @Tags      本体分类模板
// @Summary   分页查询分类模板（富化 parentLabel/treeRootLabel/hasChildren）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     page query int false "页码"
// @Param     pageSize query int false "每页大小"
// @Success   200 {object} response.Response{data=response.PageResult,msg=string}
// @Router    /ontology/classTemplate/getClassTemplatePage [get]
func (api *ClassTemplateApi) GetClassTemplatePage(c *gin.Context) {
	var info ontReq.SearchClassTemplate
	if err := c.ShouldBindQuery(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := ClassTemplateService.GetClassTemplatePage(info)
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

// GetClassTemplateTreeRoots 分类树下拉
// @Tags      本体分类模板
// @Summary   分类树下拉（distinct treeRoot + 根节点 label）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=[]response.ClassTreeRootItem,msg=string}
// @Router    /ontology/classTemplate/getClassTemplateTreeRoots [get]
func (api *ClassTemplateApi) GetClassTemplateTreeRoots(c *gin.Context) {
	list, err := ClassTemplateService.GetClassTemplateTreeRoots()
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// GetClassTemplateRefList 骨架子表回填
// @Tags      本体分类模板
// @Summary   骨架引用列表（按 classTemplateId）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     classTemplateId query uint true "分类模板ID"
// @Success   200 {object} response.Response{data=[]ontology.OntClassTemplateRef,msg=string}
// @Router    /ontology/classTemplate/getClassTemplateRefList [get]
func (api *ClassTemplateApi) GetClassTemplateRefList(c *gin.Context) {
	var req ontReq.RefListOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := ClassTemplateService.GetClassTemplateRefList(req.ClassTemplateId)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// GetClassTemplateInherited 继承视图（预留，本期无 UI）
// @Tags      本体分类模板
// @Summary   分类模板继承视图（父链骨架归并 + 外观继承）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     templateCode query string true "模板编码"
// @Success   200 {object} response.Response{data=response.ClassTemplateInheritedView,msg=string}
// @Router    /ontology/classTemplate/getClassTemplateInherited [get]
func (api *ClassTemplateApi) GetClassTemplateInherited(c *gin.Context) {
	var req ontReq.InheritedOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	view, err := ClassTemplateService.GetClassTemplateInherited(req.TemplateCode)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(view, c)
}

// PreviewClassificationCode 编码预览
// @Tags      本体分类模板
// @Summary   分类编码预览（不落库）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     parentId query uint false "父节点ID（0=根级）"
// @Param     treeRoot query string true "分类树标识"
// @Success   200 {object} response.Response{data=string,msg=string}
// @Router    /ontology/classTemplate/previewClassificationCode [get]
func (api *ClassTemplateApi) PreviewClassificationCode(c *gin.Context) {
	var req ontReq.PreviewCodeOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	code, err := ClassTemplateService.PreviewClassificationCode(req.ParentId, req.TreeRoot)
	if err != nil {
		global.GVA_LOG.Error("编码预览失败!", zap.Error(err))
		response.FailWithMessage("编码预览失败:"+err.Error(), c)
		return
	}
	response.OkWithData(code, c)
}

type ClassificationRuleApi struct{}

// FindClassificationRule 查询编码规则
// @Tags      本体分类编码规则
// @Summary   查询分类编码规则（无则返回默认值）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     treeRoot query string true "分类树标识"
// @Success   200 {object} response.Response{data=ontology.OntClassificationRule,msg=string}
// @Router    /ontology/classificationRule/findClassificationRule [get]
func (api *ClassificationRuleApi) FindClassificationRule(c *gin.Context) {
	var req ontReq.RuleOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	rule, err := ClassificationRuleService.FindClassificationRule(req.TreeRoot)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(rule, c)
}

// SaveClassificationRule 保存编码规则
// @Tags      本体分类编码规则
// @Summary   保存分类编码规则（upsert；仅影响新节点，存量编码冻结）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntClassificationRule true "编码规则"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/classificationRule/saveClassificationRule [put]
func (api *ClassificationRuleApi) SaveClassificationRule(c *gin.Context) {
	var rule ontology.OntClassificationRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ClassificationRuleService.SaveClassificationRule(&rule); err != nil {
		global.GVA_LOG.Error("保存失败!", zap.Error(err))
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}
