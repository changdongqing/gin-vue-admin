package report

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/report"
	repReq "github.com/flipped-aurora/gin-vue-admin/server/model/report/request"
	"gorm.io/gorm"
)

type DataSetService struct{}

// DataSetServiceApp 领域内单例（query 服务经它访问数据集装载/参数/转换）
var DataSetServiceApp = &DataSetService{}

var setCodePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{1,50}$`)

var paramTypeSet = map[string]bool{
	"string": true, "number": true, "date": true, "datetime": true,
	"dateRange": true, "select": true, "multipleSelect": true,
}

var transformTypeSet = map[string]bool{"js": true, "dict": true}

// CreateDataSet 新增数据集（校验 → 事务：主表 Create → 子表批量插入）
func (s *DataSetService) CreateDataSet(p *repReq.SaveDataSetReq, operator string) (uint, error) {
	enableFlag := true
	if p.EnableFlag != nil {
		enableFlag = *p.EnableFlag
	}
	if err := s.validateSave(p); err != nil {
		return 0, err
	}
	if err := s.checkCodeUnique(global.GVA_DB, p.SetCode, 0); err != nil {
		return 0, err
	}
	var id uint
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		main := report.ReportDataSet{
			SetCode:     p.SetCode,
			SetName:     p.SetName,
			SetDesc:     p.SetDesc,
			SourceCode:  p.SourceCode,
			SetType:     p.SetType,
			DynSentence: p.DynSentence,
			EnableFlag:  enableFlag,
			CreatedBy:   operator,
			UpdatedBy:   operator,
		}
		if err := tx.Create(&main).Error; err != nil {
			return err
		}
		id = main.ID
		return s.insertChildren(tx, p, operator)
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateDataSet 更新数据集（主表 Updates → 子表先删后插全量重建；数据集不含连接配置，无需触碰连接池）
func (s *DataSetService) UpdateDataSet(p *repReq.SaveDataSetReq, operator string) error {
	var old report.ReportDataSet
	if err := global.GVA_DB.First(&old, p.ID).Error; err != nil {
		return ErrSetNotExists
	}
	if err := s.validateSave(p); err != nil {
		return err
	}
	if err := s.checkCodeUnique(global.GVA_DB, p.SetCode, p.ID); err != nil {
		return err
	}
	if p.SetType != old.SetType {
		return fmt.Errorf("数据集类型不允许修改")
	}
	enableFlag := true
	if p.EnableFlag != nil {
		enableFlag = *p.EnableFlag
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&report.ReportDataSet{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
			"set_code":     p.SetCode,
			"set_name":     p.SetName,
			"set_desc":     p.SetDesc,
			"source_code":  p.SourceCode,
			"dyn_sentence": p.DynSentence,
			"enable_flag":  enableFlag,
			"updated_by":   operator,
		}).Error; err != nil {
			return err
		}
		if err := s.deleteChildren(tx, old.SetCode); err != nil {
			return err
		}
		return s.insertChildren(tx, p, operator)
	})
}

// DeleteDataSet 删除数据集：引用校验（03/05 落地）→ 事务：主表软删 + 两子表删除
func (s *DataSetService) DeleteDataSet(id uint) error {
	var old report.ReportDataSet
	if err := global.GVA_DB.First(&old, id).Error; err != nil {
		return ErrSetNotExists
	}
	if err := s.validateNotReferencedByReport(old.SetCode); err != nil {
		return err
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&report.ReportDataSet{}, id).Error; err != nil {
			return err
		}
		return s.deleteChildren(tx, old.SetCode)
	})
}

// DataSetDetail 详情聚合（主表 + 参数 + 转换）
type DataSetDetail struct {
	report.ReportDataSet
	Params     []report.ReportDataSetParam     `json:"params"`
	Transforms []report.ReportDataSetTransform `json:"transforms"`
}

// GetDataSet 详情
func (s *DataSetService) GetDataSet(id uint) (*DataSetDetail, error) {
	var main report.ReportDataSet
	if err := global.GVA_DB.First(&main, id).Error; err != nil {
		return nil, ErrSetNotExists
	}
	params, transforms, err := s.loadChildren(main.SetCode)
	if err != nil {
		return nil, err
	}
	return &DataSetDetail{ReportDataSet: main, Params: params, Transforms: transforms}, nil
}

// GetDataSetByCode 按编码查询主表
func (s *DataSetService) GetDataSetByCode(code string) (p report.ReportDataSet, err error) {
	err = global.GVA_DB.Where("set_code = ?", code).First(&p).Error
	if err == gorm.ErrRecordNotFound {
		return report.ReportDataSet{}, ErrSetNotExists
	}
	return
}

// GetDataSetList 分页查询
func (s *DataSetService) GetDataSetList(info repReq.SearchDataSet) (list []report.ReportDataSet, total int64, err error) {
	db := global.GVA_DB.Model(&report.ReportDataSet{})
	if info.SetType != "" {
		db = db.Where("set_type = ?", info.SetType)
	}
	if info.SourceCode != "" {
		db = db.Where("source_code = ?", info.SourceCode)
	}
	if info.EnableFlag != nil {
		db = db.Where("enable_flag = ?", *info.EnableFlag)
	}
	if info.Keyword != "" {
		kw := "%" + info.Keyword + "%"
		db = db.Where("set_code LIKE ? OR set_name LIKE ?", kw, kw)
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Order("id DESC").Limit(info.PageSize).Offset((info.Page - 1) * info.PageSize).Find(&list).Error
	return
}

// GetEnabledDataSetAll 已启用全量下拉（03/05 绑定用）
func (s *DataSetService) GetEnabledDataSetAll() (list []report.ReportDataSet, err error) {
	err = global.GVA_DB.Where("enable_flag = ?", true).Order("id ASC").Find(&list).Error
	return
}

// GetParams 参数列表（04 预览参数表单 / 05 参数栏复用）
func (s *DataSetService) GetParams(setCode string) (list []report.ReportDataSetParam, err error) {
	err = global.GVA_DB.Where("set_code = ?", setCode).Order("order_num ASC, id ASC").Find(&list).Error
	return
}

// GetTransforms 转换列表（04 渲染复用）
func (s *DataSetService) GetTransforms(setCode string) (list []report.ReportDataSetTransform, err error) {
	err = global.GVA_DB.Where("set_code = ?", setCode).Order("order_num ASC, id ASC").Find(&list).Error
	return
}

// LoadForQuery 查询前置装载（主表存在/启用校验 + 转换 + 参数）
func (s *DataSetService) LoadForQuery(setCode string) (set report.ReportDataSet, transforms []report.ReportDataSetTransform, params []report.ReportDataSetParam, err error) {
	set, err = s.GetDataSetByCode(setCode)
	if err != nil {
		return
	}
	transforms, err = s.GetTransforms(setCode)
	if err != nil {
		return
	}
	params, err = s.GetParams(setCode)
	return
}

// extractFieldsFromCaseResult 结果案例首行 keys（03 设计器字段来源 / 05 白名单来源）
func extractFieldsFromCaseResult(caseResult string) []string {
	if strings.TrimSpace(caseResult) == "" {
		return nil
	}
	var rows []map[string]interface{}
	if err := jsonUnmarshal([]byte(caseResult), &rows); err != nil {
		return nil
	}
	if len(rows) == 0 {
		return nil
	}
	fields := make([]string, 0, len(rows[0]))
	for k := range rows[0] {
		fields = append(fields, k)
	}
	return fields
}

// validateNotReferencedByReport 数据集删除引用校验（03/05 落地）：
// Excel 模板 set_codes LIKE 粗筛 + Go 拆分精确比对；分析报表 set_code 等值
func (s *DataSetService) validateNotReferencedByReport(setCode string) error {
	if global.GVA_DB.Migrator().HasTable("report_excel_templates") {
		var setCodes []string
		if err := global.GVA_DB.Table("report_excel_templates").
			Where("set_codes LIKE ? AND deleted_at IS NULL", "%"+setCode+"%").
			Pluck("set_codes", &setCodes).Error; err != nil {
			return err
		}
		for _, joined := range setCodes {
			for _, code := range strings.Split(joined, "|") { // LIKE 粗筛后 Go 精确比对，防子串误判
				if strings.TrimSpace(code) == setCode {
					return ErrSetReferencedByReport
				}
			}
		}
	}
	if global.GVA_DB.Migrator().HasTable("report_analysis_reports") {
		var count int64
		if err := global.GVA_DB.Table("report_analysis_reports").
			Where("set_code = ? AND deleted_at IS NULL", setCode).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrSetReferencedByReport
		}
	}
	return nil
}

// validateSave 保存校验（编码格式/SQL 防注入/参数名唯一/类型合法/转换类型合法/数据源存在启用）
func (s *DataSetService) validateSave(p *repReq.SaveDataSetReq) error {
	if !setCodePattern.MatchString(p.SetCode) {
		return fmt.Errorf("数据集编码不合法（字母/数字/下划线，≤50字符）")
	}
	if p.SetType == "sql" {
		if strings.TrimSpace(p.SourceCode) == "" {
			return ErrSetSourceMissing
		}
		// 语法与危险词校验（数据源存在性/启用性在执行时校验——创建期数据源可能尚未配好）
		if err := ValidateSQL(p.DynSentence); err != nil {
			return err
		}
	} else {
		if _, err := parseHTTPConfig(p.DynSentence); err != nil {
			return err
		}
	}
	seenParams := map[string]bool{}
	for _, param := range p.Params {
		if strings.TrimSpace(param.ParamName) == "" {
			return fmt.Errorf("参数名不能为空")
		}
		if !paramNamePattern.MatchString(param.ParamName) {
			return fmt.Errorf("参数名「%s」不合法（字母/数字/下划线）", param.ParamName)
		}
		if seenParams[param.ParamName] {
			return fmt.Errorf("参数名「%s」重复", param.ParamName)
		}
		seenParams[param.ParamName] = true
		if !paramTypeSet[param.ParamType] {
			return fmt.Errorf("%w: %s", ErrSetParamTypeInvalid, param.ParamType)
		}
	}
	for _, tf := range p.Transforms {
		if !transformTypeSet[tf.TransformType] {
			return fmt.Errorf("%w: %s", ErrSetTransformFailed, "不支持的转换类型 "+tf.TransformType)
		}
	}
	return nil
}

// paramNamePattern 参数名（与 ${} 正则一致）
var paramNamePattern = regexp.MustCompile(`^\w+$`)

func (s *DataSetService) checkCodeUnique(db *gorm.DB, setCode string, excludeID uint) error {
	var count int64
	q := db.Model(&report.ReportDataSet{}).Where("set_code = ?", setCode)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrSetCodeDuplicate
	}
	return nil
}

func (s *DataSetService) insertChildren(tx *gorm.DB, p *repReq.SaveDataSetReq, operator string) error {
	for i := range p.Params {
		p.Params[i].ID = 0
		p.Params[i].SetCode = p.SetCode
		p.Params[i].CreatedBy = operator
		p.Params[i].UpdatedBy = operator
	}
	for i := range p.Transforms {
		p.Transforms[i].ID = 0
		p.Transforms[i].SetCode = p.SetCode
		p.Transforms[i].CreatedBy = operator
		p.Transforms[i].UpdatedBy = operator
	}
	if len(p.Params) > 0 {
		if err := tx.Create(&p.Params).Error; err != nil {
			return err
		}
	}
	if len(p.Transforms) > 0 {
		if err := tx.Create(&p.Transforms).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *DataSetService) deleteChildren(tx *gorm.DB, setCode string) error {
	if err := tx.Where("set_code = ?", setCode).Delete(&report.ReportDataSetParam{}).Error; err != nil {
		return err
	}
	return tx.Where("set_code = ?", setCode).Delete(&report.ReportDataSetTransform{}).Error
}

func (s *DataSetService) loadChildren(setCode string) ([]report.ReportDataSetParam, []report.ReportDataSetTransform, error) {
	params, err := s.GetParams(setCode)
	if err != nil {
		return nil, nil, err
	}
	transforms, err := s.GetTransforms(setCode)
	if err != nil {
		return nil, nil, err
	}
	return params, transforms, nil
}
