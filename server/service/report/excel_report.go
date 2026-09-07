package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/report"
	repReq "github.com/flipped-aurora/gin-vue-admin/server/model/report/request"
	"gorm.io/gorm"
)

type ExcelReportService struct{}

// ExcelReportDetail 详情聚合（元数据 + 模板三字段）
type ExcelReportDetail struct {
	report.ReportExcelReport
	SetCodes string `json:"setCodes"`
	SetParam string `json:"setParam"`
	JsonStr  string `json:"jsonStr"`
}

// DataSetFields 设计器左栏数据集字段
type DataSetFields struct {
	SetCode string   `json:"setCode"`
	SetName string   `json:"setName"`
	Fields  []string `json:"fields"`
}

// CreateReport 新增元数据
func (s *ExcelReportService) CreateReport(p *report.ReportExcelReport, operator string) error {
	if err := s.checkCodeUnique(global.GVA_DB, p.ReportCode, 0); err != nil {
		return err
	}
	p.CreatedBy, p.UpdatedBy = operator, operator
	return global.GVA_DB.Create(p).Error
}

// UpdateReport 更新元数据（编码不可改）
func (s *ExcelReportService) UpdateReport(p *report.ReportExcelReport, operator string) error {
	var old report.ReportExcelReport
	if err := global.GVA_DB.First(&old, p.ID).Error; err != nil {
		return ErrExcelReportNotExists
	}
	if p.ReportCode != old.ReportCode {
		return fmt.Errorf("报表编码不允许修改")
	}
	return global.GVA_DB.Model(&report.ReportExcelReport{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
		"report_name":  p.ReportName,
		"report_group": p.ReportGroup,
		"report_desc":  p.ReportDesc,
		"updated_by":   operator,
	}).Error
}

// DeleteReport 删除报表（级联软删模板）
func (s *ExcelReportService) DeleteReport(id uint) error {
	var old report.ReportExcelReport
	if err := global.GVA_DB.First(&old, id).Error; err != nil {
		return ErrExcelReportNotExists
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&report.ReportExcelReport{}, id).Error; err != nil {
			return err
		}
		return tx.Where("report_code = ?", old.ReportCode).Delete(&report.ReportExcelTemplate{}).Error
	})
}

// GetReport 详情（元数据+模板）
func (s *ExcelReportService) GetReport(id uint) (*ExcelReportDetail, error) {
	var main report.ReportExcelReport
	if err := global.GVA_DB.First(&main, id).Error; err != nil {
		return nil, ErrExcelReportNotExists
	}
	return s.assembleDetail(main)
}

// GetReportByCode 按编码查询（04 预览页使用）
func (s *ExcelReportService) GetReportByCode(code string) (*ExcelReportDetail, error) {
	var main report.ReportExcelReport
	if err := global.GVA_DB.Where("report_code = ?", code).First(&main).Error; err != nil {
		return nil, ErrExcelReportNotExists
	}
	return s.assembleDetail(main)
}

func (s *ExcelReportService) assembleDetail(main report.ReportExcelReport) (*ExcelReportDetail, error) {
	detail := &ExcelReportDetail{ReportExcelReport: main}
	var tpl report.ReportExcelTemplate
	if err := global.GVA_DB.Where("report_code = ?", main.ReportCode).First(&tpl).Error; err == nil {
		detail.SetCodes, detail.SetParam, detail.JsonStr = tpl.SetCodes, tpl.SetParam, tpl.JsonStr
	}
	return detail, nil
}

// GetReportList 分页（keyword 匹配编码/名称；reportCode 精确过滤供设计器/预览页；不返回 json_str）
func (s *ExcelReportService) GetReportList(info repReq.SearchExcelReport) (list []report.ReportExcelReport, total int64, err error) {
	db := global.GVA_DB.Model(&report.ReportExcelReport{}).Omit("json_str")
	_ = db // Omit 在 Model 上生效
	if info.ReportCode != "" {
		db = db.Where("report_code = ?", info.ReportCode)
	}
	if info.ReportGroup != "" {
		db = db.Where("report_group = ?", info.ReportGroup)
	}
	if info.Keyword != "" {
		kw := "%" + info.Keyword + "%"
		db = db.Where("report_code LIKE ? OR report_name LIKE ?", kw, kw)
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Order("id DESC").Limit(info.PageSize).Offset((info.Page - 1) * info.PageSize).Find(&list).Error
	return
}

// GetEnabledReportAll 已启用全量（预留）
func (s *ExcelReportService) GetEnabledReportAll() (list []report.ReportExcelReport, err error) {
	err = global.GVA_DB.Order("id ASC").Find(&list).Error
	return
}

// SaveTemplate upsert 模板行【仅 json_str/set_param】：行不存在则 insert（set_codes 置空），存在则不动 set_codes
func (s *ExcelReportService) SaveTemplate(r *repReq.SaveExcelTemplateReq, operator string) error {
	var main report.ReportExcelReport
	if err := global.GVA_DB.Where("report_code = ?", r.ReportCode).First(&main).Error; err != nil {
		return ErrExcelReportNotExists
	}
	if !json.Valid([]byte(r.JsonStr)) {
		return ErrExcelTemplateJSONInvalid
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var tpl report.ReportExcelTemplate
		err := tx.Where("report_code = ?", r.ReportCode).First(&tpl).Error
		if err == gorm.ErrRecordNotFound {
			tpl = report.ReportExcelTemplate{
				ReportCode: r.ReportCode,
				SetParam:   r.SetParam,
				JsonStr:    r.JsonStr,
				CreatedBy:  operator,
				UpdatedBy:  operator,
			}
			return tx.Create(&tpl).Error
		}
		if err != nil {
			return err
		}
		return tx.Model(&report.ReportExcelTemplate{}).Where("id = ?", tpl.ID).Updates(map[string]interface{}{
			"set_param":  r.SetParam,
			"json_str":   r.JsonStr,
			"updated_by": operator,
		}).Error
	})
}

// BindDataSets upsert 模板行【仅 set_codes】：逐个校验数据集存在（不存在显式抛错）
func (s *ExcelReportService) BindDataSets(r *repReq.BindDataSetsReq, operator string) error {
	var main report.ReportExcelReport
	if err := global.GVA_DB.Where("report_code = ?", r.ReportCode).First(&main).Error; err != nil {
		return ErrExcelReportNotExists
	}
	setCodes := ""
	if len(r.SetCodes) > 0 {
		seen := map[string]bool{}
		codes := make([]string, 0, len(r.SetCodes))
		for _, code := range r.SetCodes {
			code = strings.TrimSpace(code)
			if code == "" || seen[code] {
				continue
			}
			var count int64
			if err := global.GVA_DB.Model(&report.ReportDataSet{}).Where("set_code = ?", code).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return fmt.Errorf("%w: %s", ErrExcelReportDataSetNotExists, code)
			}
			seen[code] = true
			codes = append(codes, code)
		}
		setCodes = strings.Join(codes, "|")
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var tpl report.ReportExcelTemplate
		err := tx.Where("report_code = ?", r.ReportCode).First(&tpl).Error
		if err == gorm.ErrRecordNotFound {
			tpl = report.ReportExcelTemplate{
				ReportCode: r.ReportCode,
				SetCodes:   setCodes,
				CreatedBy:  operator,
				UpdatedBy:  operator,
			}
			return tx.Create(&tpl).Error
		}
		if err != nil {
			return err
		}
		return tx.Model(&report.ReportExcelTemplate{}).Where("id = ?", tpl.ID).Updates(map[string]interface{}{
			"set_codes":  setCodes,
			"updated_by": operator,
		}).Error
	})
}

// CopyReport 复制（元数据+模板深拷贝，新编码）
func (s *ExcelReportService) CopyReport(r *repReq.CopyExcelReportReq, operator string) (uint, error) {
	var source report.ReportExcelReport
	if err := global.GVA_DB.Where("report_code = ?", r.SourceReportCode).First(&source).Error; err != nil {
		return 0, ErrExcelCopySourceMissing
	}
	if err := s.checkCodeUnique(global.GVA_DB, r.ReportCode, 0); err != nil {
		return 0, err
	}
	var newID uint
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		clone := source
		clone.ID = 0
		clone.ReportCode = r.ReportCode
		clone.ReportName = r.ReportName
		clone.CreatedBy, clone.UpdatedBy = operator, operator
		if err := tx.Create(&clone).Error; err != nil {
			return err
		}
		newID = clone.ID
		var tpl report.ReportExcelTemplate
		if err := tx.Where("report_code = ?", source.ReportCode).First(&tpl).Error; err == nil {
			cloneTpl := tpl
			cloneTpl.ID = 0
			cloneTpl.ReportCode = r.ReportCode
			cloneTpl.CreatedBy, cloneTpl.UpdatedBy = operator, operator
			return tx.Create(&cloneTpl).Error
		} else if err != gorm.ErrRecordNotFound {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return newID, nil
}

// GetReportDataSetFields 设计器左栏：set_codes 拆分 → 每数据集 {setCode,setName,fields}。
// fields = case_result 首行 keys；数据集不存在显式抛错（脏引用不静默跳过）
func (s *ExcelReportService) GetReportDataSetFields(reportCode string) ([]DataSetFields, error) {
	var tpl report.ReportExcelTemplate
	if err := global.GVA_DB.Where("report_code = ?", reportCode).First(&tpl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []DataSetFields{}, nil
		}
		return nil, err
	}
	result := make([]DataSetFields, 0)
	for _, code := range strings.Split(tpl.SetCodes, "|") {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		var set report.ReportDataSet
		if err := global.GVA_DB.Where("set_code = ?", code).First(&set).Error; err != nil {
			return nil, fmt.Errorf("%w: %s", ErrExcelReportDataSetNotExists, code)
		}
		fields := extractFieldsFromCaseResult(set.CaseResult)
		if fields == nil {
			fields = []string{}
		}
		result = append(result, DataSetFields{SetCode: code, SetName: set.SetName, Fields: fields})
	}
	return result, nil
}

// IsSetCodeReferenced 数据集删除引用校验（挂 02 DeleteDataSet 的 validateNotReferencedByReport）
func (s *ExcelReportService) IsSetCodeReferenced(setCode string) (bool, error) {
	var templates []report.ReportExcelTemplate
	if err := global.GVA_DB.Where("set_codes LIKE ?", "%"+setCode+"%").Find(&templates).Error; err != nil {
		return false, err
	}
	for _, tpl := range templates {
		for _, code := range strings.Split(tpl.SetCodes, "|") {
			if strings.TrimSpace(code) == setCode {
				return true, nil
			}
		}
	}
	return false, nil
}

func (s *ExcelReportService) checkCodeUnique(db *gorm.DB, reportCode string, excludeID uint) error {
	var count int64
	q := db.Model(&report.ReportExcelReport{}).Where("report_code = ?", reportCode)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrExcelReportCodeDuplicate
	}
	return nil
}
