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

type ModelClassApi struct{}

// CreateModelClass 创建本体类
// @Tags      本体建模
// @Summary   创建本体类（IRI 生成/同项目查重/溯源字段强制空）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntModelClass true "本体类"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/modelClass/createModelClass [post]
func (api *ModelClassApi) CreateModelClass(c *gin.Context) {
	var p ontology.OntModelClass
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ModelClassService.CreateModelClass(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateModelClass 更新本体类
// @Tags      本体建模
// @Summary   更新本体类（编辑白名单：label/labelCn/description/icon/color/sortOrder/isInstantiable）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntModelClass true "本体类（含 ID）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/modelClass/updateModelClass [put]
func (api *ModelClassApi) UpdateModelClass(c *gin.Context) {
	var p ontology.OntModelClass
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ModelClassService.UpdateModelClass(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteModelClass 删除本体类
// @Tags      本体建模
// @Summary   删除本体类（三重守卫：归档/被子类引用/存在属性）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "类ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/modelClass/deleteModelClass [delete]
func (api *ModelClassApi) DeleteModelClass(c *gin.Context) {
	var req ontReq.ModelClassOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ModelClassService.DeleteModelClass(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// FindModelClass 本体类详情（单实体，编辑回填）
// @Tags      本体建模
// @Summary   本体类详情
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "类ID"
// @Success   200 {object} response.Response{data=ontology.OntModelClass,msg=string}
// @Router    /ontology/modelClass/findModelClass [get]
func (api *ModelClassApi) FindModelClass(c *gin.Context) {
	var req ontReq.ModelClassOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	p, err := ModelClassService.GetModelClass(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(p, c)
}

// GetModelClassDetail 类详情聚合（属性双清单 + 直接父子类）
// @Tags      本体建模
// @Summary   类详情聚合
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "类ID"
// @Success   200 {object} response.Response{data=ontRes.ClassDetail,msg=string}
// @Router    /ontology/modelClass/getModelClassDetail [get]
func (api *ModelClassApi) GetModelClassDetail(c *gin.Context) {
	var req ontReq.ModelClassOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	detail, err := ModelClassService.GetModelClassDetail(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

// GetModelClassList 分页查询本体类
// @Tags      本体建模
// @Summary   分页查询本体类（projectId/keyword/hasTemplate 三态）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=response.PageResult,msg=string}
// @Router    /ontology/modelClass/getModelClassList [get]
func (api *ModelClassApi) GetModelClassList(c *gin.Context) {
	var info ontReq.SearchModelClass
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
	list, total, err := ModelClassService.GetModelClassList(info)
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

// GetModelClassByProject 项目内全量类
// @Tags      本体建模
// @Summary   项目内全量类（03 外部模块关联选类用）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     projectId query uint true "项目ID"
// @Success   200 {object} response.Response{data=[]ontology.OntModelClass,msg=string}
// @Router    /ontology/modelClass/getModelClassByProject [get]
func (api *ModelClassApi) GetModelClassByProject(c *gin.Context) {
	var req ontReq.ModelClassByProjectOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := ModelClassService.GetModelClassByProject(req.ProjectId)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// CheckModelClassLocalName 类本地名查重
// @Tags      本体建模
// @Summary   类本地名查重（同项目）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     projectId query uint true "项目ID"
// @Param     localName query string true "本地名"
// @Param     excludeId query uint false "排除的类ID"
// @Success   200 {object} response.Response{data=bool,msg=string}
// @Router    /ontology/modelClass/checkModelClassLocalName [get]
func (api *ModelClassApi) CheckModelClassLocalName(c *gin.Context) {
	var req ontReq.CheckModelClassLocalNameOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	exists, err := ModelClassService.CheckModelClassLocalName(req.ProjectId, req.LocalName, req.ExcludeId)
	if err != nil {
		global.GVA_LOG.Error("查重失败!", zap.Error(err))
		response.FailWithMessage("查重失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"exists": exists}, "查重成功", c)
}

// InstantiateModelClass 类级模板实例化
// @Tags      本体建模
// @Summary   分类模板实例化（单事务复制字段值，同类已有属性跳过）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontReq.InstantiateClassOps true "{projectId,classId,templateCode}"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/modelClass/instantiateModelClass [post]
func (api *ModelClassApi) InstantiateModelClass(c *gin.Context) {
	var req ontReq.InstantiateClassOps
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ClassInstantiationService.InstantiateClass(req, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("实例化失败!", zap.Error(err))
		response.FailWithMessage("实例化失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("实例化成功", c)
}

// PreviewInstantiateModelClass 实例化预览
// @Tags      本体建模
// @Summary   分类模板实例化预览（只读不落库）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     projectId query uint true "项目ID"
// @Param     templateCode query string true "分类模板编码"
// @Success   200 {object} response.Response{data=ontRes.InstantiatePreview,msg=string}
// @Router    /ontology/modelClass/previewInstantiateModelClass [get]
func (api *ModelClassApi) PreviewInstantiateModelClass(c *gin.Context) {
	var req ontReq.PreviewInstantiateOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	preview, err := ClassInstantiationService.PreviewInstantiate(req)
	if err != nil {
		global.GVA_LOG.Error("预览失败!", zap.Error(err))
		response.FailWithMessage("预览失败:"+err.Error(), c)
		return
	}
	response.OkWithData(preview, c)
}

type ModelDatatypePropertyApi struct{}

// CreateModelDatatypeProperty 创建数据属性
// @Tags      本体建模
// @Summary   创建数据属性（单位校验/方案B IRI）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntModelDatatypeProperty true "数据属性"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/modelDatatypeProperty/createModelDatatypeProperty [post]
func (api *ModelDatatypePropertyApi) CreateModelDatatypeProperty(c *gin.Context) {
	var p ontology.OntModelDatatypeProperty
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := DatatypePropertyService.CreateModelDatatypeProperty(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateModelDatatypeProperty 更新数据属性
// @Tags      本体建模
// @Summary   更新数据属性（白名单：label/unitRef/enumValues/基数/isIdentifier/sortOrder）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntModelDatatypeProperty true "数据属性（含 ID）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/modelDatatypeProperty/updateModelDatatypeProperty [put]
func (api *ModelDatatypePropertyApi) UpdateModelDatatypeProperty(c *gin.Context) {
	var p ontology.OntModelDatatypeProperty
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := DatatypePropertyService.UpdateModelDatatypeProperty(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteModelDatatypeProperty 删除数据属性
// @Tags      本体建模
// @Summary   删除数据属性
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "属性ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/modelDatatypeProperty/deleteModelDatatypeProperty [delete]
func (api *ModelDatatypePropertyApi) DeleteModelDatatypeProperty(c *gin.Context) {
	var req ontReq.ModelPropertyOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := DatatypePropertyService.DeleteModelDatatypeProperty(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// FindModelDatatypeProperty 数据属性详情
// @Tags      本体建模
// @Summary   数据属性详情（编辑回填）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "属性ID"
// @Success   200 {object} response.Response{data=ontology.OntModelDatatypeProperty,msg=string}
// @Router    /ontology/modelDatatypeProperty/findModelDatatypeProperty [get]
func (api *ModelDatatypePropertyApi) FindModelDatatypeProperty(c *gin.Context) {
	var req ontReq.ModelPropertyOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	p, err := DatatypePropertyService.GetModelDatatypeProperty(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(p, c)
}

// GetModelDatatypePropertyList 分页查询数据属性
// @Tags      本体建模
// @Summary   分页查询数据属性
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=response.PageResult,msg=string}
// @Router    /ontology/modelDatatypeProperty/getModelDatatypePropertyList [get]
func (api *ModelDatatypePropertyApi) GetModelDatatypePropertyList(c *gin.Context) {
	var info ontReq.SearchDatatypeProperty
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
	list, total, err := DatatypePropertyService.GetModelDatatypePropertyList(info)
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

// InstantiateModelDatatypeProperty 数据属性模板挂载
// @Tags      本体建模
// @Summary   从属性库挂载单条数据属性模板（同类已有跳过）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontReq.InstantiatePropertyOps true "{projectId,classId,templateCode}"
// @Success   200 {object} response.Response{data=bool,msg=string}
// @Router    /ontology/modelDatatypeProperty/instantiateModelDatatypeProperty [post]
func (api *ModelDatatypePropertyApi) InstantiateModelDatatypeProperty(c *gin.Context) {
	var req ontReq.InstantiatePropertyOps
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	skipped, err := DatatypePropertyService.InstantiateModelDatatypeProperty(req, utils.GetUserInfo(c).Username)
	if err != nil {
		global.GVA_LOG.Error("挂载失败!", zap.Error(err))
		response.FailWithMessage("挂载失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"skipped": skipped}, "挂载成功", c)
}

type ModelObjectPropertyApi struct{}

// CreateModelObjectProperty 创建对象属性
// @Tags      本体建模
// @Summary   创建对象属性（range 可空；携带 inverseOf 时双向互指）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntModelObjectProperty true "对象属性"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/modelObjectProperty/createModelObjectProperty [post]
func (api *ModelObjectPropertyApi) CreateModelObjectProperty(c *gin.Context) {
	var p ontology.OntModelObjectProperty
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ObjectPropertyService.CreateModelObjectProperty(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateModelObjectProperty 更新对象属性
// @Tags      本体建模
// @Summary   更新对象属性（白名单：label/rangeClassId/基数/inverseOf/sortOrder）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntModelObjectProperty true "对象属性（含 ID）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/modelObjectProperty/updateModelObjectProperty [put]
func (api *ModelObjectPropertyApi) UpdateModelObjectProperty(c *gin.Context) {
	var p ontology.OntModelObjectProperty
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ObjectPropertyService.UpdateModelObjectProperty(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteModelObjectProperty 删除对象属性
// @Tags      本体建模
// @Summary   删除对象属性（同事务清对方 inverseOf 互指）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "属性ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/modelObjectProperty/deleteModelObjectProperty [delete]
func (api *ModelObjectPropertyApi) DeleteModelObjectProperty(c *gin.Context) {
	var req ontReq.ModelPropertyOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ObjectPropertyService.DeleteModelObjectProperty(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// FindModelObjectProperty 对象属性详情
// @Tags      本体建模
// @Summary   对象属性详情
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "属性ID"
// @Success   200 {object} response.Response{data=ontology.OntModelObjectProperty,msg=string}
// @Router    /ontology/modelObjectProperty/findModelObjectProperty [get]
func (api *ModelObjectPropertyApi) FindModelObjectProperty(c *gin.Context) {
	var req ontReq.ModelPropertyOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	p, err := ObjectPropertyService.GetModelObjectProperty(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(p, c)
}

// GetModelObjectPropertyList 分页查询对象属性
// @Tags      本体建模
// @Summary   分页查询对象属性
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=response.PageResult,msg=string}
// @Router    /ontology/modelObjectProperty/getModelObjectPropertyList [get]
func (api *ModelObjectPropertyApi) GetModelObjectPropertyList(c *gin.Context) {
	var info ontReq.SearchObjectProperty
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
	list, total, err := ObjectPropertyService.GetModelObjectPropertyList(info)
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

// InstantiateModelObjectProperty 对象属性模板挂载
// @Tags      本体建模
// @Summary   从属性库挂载单条对象属性模板（同类已有跳过，range 留空待补）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontReq.InstantiatePropertyOps true "{projectId,classId,templateCode}"
// @Success   200 {object} response.Response{data=bool,msg=string}
// @Router    /ontology/modelObjectProperty/instantiateModelObjectProperty [post]
func (api *ModelObjectPropertyApi) InstantiateModelObjectProperty(c *gin.Context) {
	var req ontReq.InstantiatePropertyOps
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	skipped, err := ObjectPropertyService.InstantiateModelObjectProperty(req, utils.GetUserInfo(c).Username)
	if err != nil {
		global.GVA_LOG.Error("挂载失败!", zap.Error(err))
		response.FailWithMessage("挂载失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"skipped": skipped}, "挂载成功", c)
}

// SuggestModelObjectPropertyInverse 反向关系建议
// @Tags      本体建模
// @Summary   反向关系建议（belongsTo；domain/range 对调，不落库）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "对象属性ID"
// @Success   200 {object} response.Response{data=ontRes.SuggestInverseResult,msg=string}
// @Router    /ontology/modelObjectProperty/suggestInverseModelObjectProperty [get]
func (api *ModelObjectPropertyApi) SuggestModelObjectPropertyInverse(c *gin.Context) {
	var req ontReq.ModelPropertyOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	suggestion, err := ObjectPropertyService.SuggestModelObjectPropertyInverse(req.ID)
	if err != nil {
		global.GVA_LOG.Error("建议失败!", zap.Error(err))
		response.FailWithMessage("建议失败:"+err.Error(), c)
		return
	}
	response.OkWithData(suggestion, c)
}
