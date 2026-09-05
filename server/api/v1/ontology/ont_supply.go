package ontology

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
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
