package ontology

import (
	"net/http"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
	ontReq "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type AnnotationPropertyApi struct{}

// CreateAnnotationProperty 创建注释属性
// @Tags      本体注释属性
// @Summary   创建注释属性（localName 唯一校验）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntAnnotationProperty true "注释属性"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/annotationProperty/createAnnotationProperty [post]
func (api *AnnotationPropertyApi) CreateAnnotationProperty(c *gin.Context) {
	var p ontology.OntAnnotationProperty
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := AnnotationPropertyService.CreateAnnotationProperty(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateAnnotationProperty 更新注释属性
// @Tags      本体注释属性
// @Summary   更新注释属性（唯一校验排除自身；无 builtin 保护）
// @Security  ApiKeyAuth
// @Accept    application/json
// @Produce   application/json
// @Param     data body ontology.OntAnnotationProperty true "注释属性（含 ID）"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/annotationProperty/updateAnnotationProperty [put]
func (api *AnnotationPropertyApi) UpdateAnnotationProperty(c *gin.Context) {
	var p ontology.OntAnnotationProperty
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := AnnotationPropertyService.UpdateAnnotationProperty(&p, utils.GetUserInfo(c).Username); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteAnnotationProperty 删除注释属性
// @Tags      本体注释属性
// @Summary   删除注释属性（无 builtin 保护；删除后建模侧不再序列化对应 ont:xxx）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "注释属性ID"
// @Success   200 {object} response.Response{msg=string}
// @Router    /ontology/annotationProperty/deleteAnnotationProperty [delete]
func (api *AnnotationPropertyApi) DeleteAnnotationProperty(c *gin.Context) {
	var req ontReq.AnnotationPropertyOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := AnnotationPropertyService.DeleteAnnotationProperty(req.ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// FindAnnotationProperty 注释属性详情
// @Tags      本体注释属性
// @Summary   注释属性详情
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     ID query uint true "注释属性ID"
// @Success   200 {object} response.Response{data=ontology.OntAnnotationProperty,msg=string}
// @Router    /ontology/annotationProperty/findAnnotationProperty [get]
func (api *AnnotationPropertyApi) FindAnnotationProperty(c *gin.Context) {
	var req ontReq.AnnotationPropertyOps
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	p, err := AnnotationPropertyService.GetAnnotationProperty(req.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(p, c)
}

// GetAnnotationPropertyList 分页查询注释属性
// @Tags      本体注释属性
// @Summary   分页查询注释属性（keyword/appliesTo/创建时间范围）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=response.PageResult,msg=string}
// @Router    /ontology/annotationProperty/getAnnotationPropertyList [get]
func (api *AnnotationPropertyApi) GetAnnotationPropertyList(c *gin.Context) {
	var info ontReq.SearchAnnotationProperty
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
	list, total, err := AnnotationPropertyService.GetAnnotationPropertyList(info)
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

// GetAnnotationPropertyAll 全量查询
// @Tags      本体注释属性
// @Summary   全量注释属性（同分页条件去分页，下拉/导出数据源）
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200 {object} response.Response{data=[]ontology.OntAnnotationProperty,msg=string}
// @Router    /ontology/annotationProperty/getAnnotationPropertyAll [get]
func (api *AnnotationPropertyApi) GetAnnotationPropertyAll(c *gin.Context) {
	var info ontReq.SearchAnnotationProperty
	if err := c.ShouldBindQuery(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := AnnotationPropertyService.GetAnnotationPropertyAll(info)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// ExportAnnotationPropertyExcel 导出注释属性清单（按当前搜索条件全量）
// @Tags      本体注释属性
// @Summary   导出 xlsx（列：编号/注释属性名/显示名/值域类型/作用对象/描述）
// @Security  ApiKeyAuth
// @Produce   application/octet-stream
// @Router    /ontology/annotationProperty/exportAnnotationPropertyExcel [get]
func (api *AnnotationPropertyApi) ExportAnnotationPropertyExcel(c *gin.Context) {
	var info ontReq.SearchAnnotationProperty
	if err := c.ShouldBindQuery(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := AnnotationPropertyService.GetAnnotationPropertyAll(info)
	if err != nil {
		global.GVA_LOG.Error("导出失败!", zap.Error(err))
		response.FailWithMessage("导出失败:"+err.Error(), c)
		return
	}
	f := excelize.NewFile()
	sheet := "注释属性清单"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"编号", "注释属性名", "显示名", "值域类型", "作用对象", "描述"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for i, item := range list {
		r := i + 2
		values := []interface{}{item.ID, item.LocalName, item.Label, item.RangeXsd, item.AppliesTo, item.Description}
		for col, v := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, r)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="annotation_properties.xlsx"`)
	c.Header("success", "true")
	if err := f.Write(c.Writer); err != nil {
		global.GVA_LOG.Error("导出失败!", zap.Error(err))
		response.FailWithMessage("导出失败:"+err.Error(), c)
		return
	}
	c.Status(http.StatusOK)
}
