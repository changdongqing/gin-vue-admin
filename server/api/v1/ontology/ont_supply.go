package ontology

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	ontReq "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SupplyApi struct{}

// GetPropertyTemplatesForSupply 供给：属性模板列表（FR-5）
// @Tags      本体供给
// @Summary   供给属性模板（只读，默认排除弃用，仅启用）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     kind query string false "属性类型 datatype/object"
// @Param     category query string false "分组"
// @Param     includeDeprecated query bool false "是否包含弃用（默认 false）"
// @Success   200 {object} response.Response{data=[]ontology.OntPropertyTemplate,msg=string}
// @Router    /ontology/supply/v1/propertyTemplates [get]
func (api *SupplyApi) GetPropertyTemplatesForSupply(c *gin.Context) {
	includeDeprecated, _ := strconv.ParseBool(c.Query("includeDeprecated"))
	list, err := PropertyTemplateService.GetPropertyTemplatesForSupply(c.Query("kind"), c.Query("category"), includeDeprecated)
	if err != nil {
		global.GVA_LOG.Error("供给查询失败!", zap.Error(err))
		response.FailWithMessage("供给查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// GetClassTemplateTreeForSupply 供给：分类模板树（FR-5）
// @Tags      本体供给
// @Summary   供给分类模板树扁平列表（只读，默认排除弃用，前端组树渲染）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     treeRoot query string false "分类树标识"
// @Success   200 {object} response.Response{data=[]ontology.OntClassTemplate,msg=string}
// @Router    /ontology/supply/v1/classTemplate/tree [get]
func (api *SupplyApi) GetClassTemplateTreeForSupply(c *gin.Context) {
	list, err := ClassTemplateService.GetClassTemplateTreeForSupply(c.Query("treeRoot"))
	if err != nil {
		global.GVA_LOG.Error("供给查询失败!", zap.Error(err))
		response.FailWithMessage("供给查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// SuggestClassHierarchy 供给：类层级建议（FR-9）
// @Tags      本体供给
// @Summary   类层级建议（本期无镜像数据，返回治理分类树的父模板信息供参考）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     templateCode query string true "模板编码"
// @Success   200 {object} response.Response{data=response.ClassHierarchySuggest,msg=string}
// @Router    /ontology/supply/v1/classHierarchy/suggest [get]
func (api *SupplyApi) SuggestClassHierarchy(c *gin.Context) {
	var req ontReq.SuggestOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	suggest, err := ClassTemplateService.SuggestClassHierarchy(req.TemplateCode)
	if err != nil {
		global.GVA_LOG.Error("供给查询失败!", zap.Error(err))
		response.FailWithMessage("供给查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(suggest, c)
}
