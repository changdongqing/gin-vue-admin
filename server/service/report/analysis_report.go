package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/report"
	repReq "github.com/flipped-aurora/gin-vue-admin/server/model/report/request"
	"gorm.io/gorm"
)

type AnalysisReportService struct{}

// AnalysisReportDetail 详情聚合（元数据 + 配置 + 参数定义 + 数据集名）
type AnalysisReportDetail struct {
	report.ReportAnalysisReport
	SetName    string                      `json:"setName"`
	ConfigJson string                      `json:"configJson"`
	SetParam   string                      `json:"setParam"`
	Params     []report.ReportDataSetParam `json:"params"`
}

var dateHeadPattern = regexp.MustCompile(`^\d{4}[-/]\d{1,2}[-/]\d{1,2}.*$`)

// CreateAnalysisReport 新增元数据（编码查重；数据集存在且启用）
func (s *AnalysisReportService) CreateAnalysisReport(p *repReq.SaveAnalysisReq, operator string) (uint, error) {
	if err := s.checkCodeUnique(global.GVA_DB, p.ReportCode, 0); err != nil {
		return 0, err
	}
	if err := s.checkDataSetUsable(p.SetCode); err != nil {
		return 0, err
	}
	status := 0
	if p.Status != nil {
		status = *p.Status
	}
	main := report.ReportAnalysisReport{
		ReportCode:  p.ReportCode,
		ReportName:  p.ReportName,
		ReportGroup: p.ReportGroup,
		ReportDesc:  p.ReportDesc,
		SetCode:     p.SetCode,
		Status:      status,
		CreatedBy:   operator,
		UpdatedBy:   operator,
	}
	if err := global.GVA_DB.Create(&main).Error; err != nil {
		return 0, err
	}
	return main.ID, nil
}

// UpdateAnalysisReport 更新元数据（编码不可改；更换数据集前端强提示配置失效，后端不做级联清理）
func (s *AnalysisReportService) UpdateAnalysisReport(p *repReq.SaveAnalysisReq, operator string) error {
	var old report.ReportAnalysisReport
	if err := global.GVA_DB.First(&old, p.ID).Error; err != nil {
		return ErrAnalysisReportNotExists
	}
	if p.ReportCode != old.ReportCode {
		return fmt.Errorf("报表编码不允许修改")
	}
	if err := s.checkDataSetUsable(p.SetCode); err != nil {
		return err
	}
	status := old.Status
	if p.Status != nil {
		status = *p.Status
	}
	return global.GVA_DB.Model(&report.ReportAnalysisReport{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
		"report_name":  p.ReportName,
		"report_group": p.ReportGroup,
		"report_desc":  p.ReportDesc,
		"set_code":     p.SetCode,
		"status":       status,
		"updated_by":   operator,
	}).Error
}

// DeleteAnalysisReport 删除（级联删配置）
func (s *AnalysisReportService) DeleteAnalysisReport(id uint) error {
	var old report.ReportAnalysisReport
	if err := global.GVA_DB.First(&old, id).Error; err != nil {
		return ErrAnalysisReportNotExists
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&report.ReportAnalysisReport{}, id).Error; err != nil {
			return err
		}
		return tx.Where("report_code = ?", old.ReportCode).Delete(&report.ReportAnalysisConfig{}).Error
	})
}

// GetAnalysisReport 详情（一个接口拉齐：元数据+setName+configJson/setParam+params）
func (s *AnalysisReportService) GetAnalysisReport(id uint) (*AnalysisReportDetail, error) {
	var main report.ReportAnalysisReport
	if err := global.GVA_DB.First(&main, id).Error; err != nil {
		return nil, ErrAnalysisReportNotExists
	}
	return s.assembleDetail(&main)
}

// GetAnalysisReportByCode 按编码详情（预览页/设计器使用）
func (s *AnalysisReportService) GetAnalysisReportByCode(code string) (*AnalysisReportDetail, error) {
	var main report.ReportAnalysisReport
	if err := global.GVA_DB.Where("report_code = ?", code).First(&main).Error; err != nil {
		return nil, ErrAnalysisReportNotExists
	}
	return s.assembleDetail(&main)
}

func (s *AnalysisReportService) assembleDetail(main *report.ReportAnalysisReport) (*AnalysisReportDetail, error) {
	detail := &AnalysisReportDetail{ReportAnalysisReport: *main, Params: []report.ReportDataSetParam{}}
	if main.SetCode != "" {
		var set report.ReportDataSet
		if err := global.GVA_DB.Where("set_code = ?", main.SetCode).First(&set).Error; err == nil {
			detail.SetName = set.SetName
			params, perr := DataSetServiceApp.GetParams(main.SetCode)
			if perr == nil {
				detail.Params = params
			}
		}
	}
	var cfg report.ReportAnalysisConfig
	if err := global.GVA_DB.Where("report_code = ?", main.ReportCode).First(&cfg).Error; err == nil {
		detail.ConfigJson, detail.SetParam = cfg.ConfigJson, cfg.SetParam
	}
	return detail, nil
}

// GetAnalysisReportList 分页（keyword 匹配编码/名称；不返回 config 大字段）
func (s *AnalysisReportService) GetAnalysisReportList(info repReq.SearchAnalysis) (list []report.ReportAnalysisReport, total int64, err error) {
	db := global.GVA_DB.Model(&report.ReportAnalysisReport{})
	if info.ReportGroup != "" {
		db = db.Where("report_group = ?", info.ReportGroup)
	}
	if info.SetCode != "" {
		db = db.Where("set_code = ?", info.SetCode)
	}
	if info.Status != nil {
		db = db.Where("status = ?", *info.Status)
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

// CopyAnalysisReport 复制（元数据+配置深拷贝）
func (s *AnalysisReportService) CopyAnalysisReport(r *repReq.CopyAnalysisReq, operator string) (uint, error) {
	var source report.ReportAnalysisReport
	if err := global.GVA_DB.Where("report_code = ?", r.SourceReportCode).First(&source).Error; err != nil {
		return 0, ErrAnalysisCopySourceMissing
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
		var cfg report.ReportAnalysisConfig
		if err := tx.Where("report_code = ?", source.ReportCode).First(&cfg).Error; err == nil {
			cloneCfg := cfg
			cloneCfg.ID = 0
			cloneCfg.ReportCode = r.ReportCode
			cloneCfg.CreatedBy, cloneCfg.UpdatedBy = operator, operator
			return tx.Create(&cloneCfg).Error
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

// SaveConfig 三层校验 + upsert 配置：
// ① JSON 可解析 → ② sheetType/aggregation 枚举合法 → ③ 字段白名单（case_result 首行 keys；为空跳过——纵深防御）
func (s *AnalysisReportService) SaveConfig(r *repReq.SaveAnalysisConfigReq, operator string) error {
	var main report.ReportAnalysisReport
	if err := global.GVA_DB.Where("report_code = ?", r.ReportCode).First(&main).Error; err != nil {
		return ErrAnalysisReportNotExists
	}
	var cfgMap map[string]interface{}
	if err := json.Unmarshal([]byte(r.ConfigJson), &cfgMap); err != nil {
		return ErrAnalysisConfigJSONInvalid
	}
	// 枚举校验
	sheetType, _ := cfgMap["sheetType"].(string)
	if sheetType != "pivot" && sheetType != "table" {
		return fmt.Errorf("%w: sheetType 须为 pivot/table", ErrAnalysisConfigJSONInvalid)
	}
	fields, _ := cfgMap["fields"].(map[string]interface{})
	if fields == nil {
		return fmt.Errorf("%w: 缺少 fields", ErrAnalysisConfigJSONInvalid)
	}
	values, _ := fields["values"].([]interface{})
	for _, v := range values {
		vm, _ := v.(map[string]interface{})
		agg, _ := vm["aggregation"].(string)
		switch agg {
		case "SUM", "COUNT", "AVG", "MIN", "MAX", "NONE":
		default:
			return fmt.Errorf("%w: aggregation 非法 %q", ErrAnalysisConfigJSONInvalid, agg)
		}
	}
	// 字段白名单（纵深防御）
	var set report.ReportDataSet
	if err := global.GVA_DB.Where("set_code = ?", main.SetCode).First(&set).Error; err == nil {
		whitelist := extractFieldsFromCaseResult(set.CaseResult)
		if len(whitelist) > 0 {
			allowed := map[string]bool{}
			for _, f := range whitelist {
				allowed[f] = true
			}
			checkField := (func(field string) error {
				field = strings.TrimSpace(field)
				if field == "" {
					return nil
				}
				if !allowed[field] {
					return fmt.Errorf("%w: %s", ErrAnalysisFieldInvalid, field)
				}
				return nil
			})
			for _, key := range []string{"rows", "columns"} {
				arr, _ := fields[key].([]interface{})
				for _, f := range arr {
					fs, _ := f.(string)
					if err := checkField(fs); err != nil {
						return err
					}
				}
			}
			for _, v := range values {
				vm, _ := v.(map[string]interface{})
				fs, _ := vm["field"].(string)
				if err := checkField(fs); err != nil {
					return err
				}
			}
		}
	}
	// upsert
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var cfg report.ReportAnalysisConfig
		err := tx.Where("report_code = ?", r.ReportCode).First(&cfg).Error
		if err == gorm.ErrRecordNotFound {
			cfg = report.ReportAnalysisConfig{
				ReportCode: r.ReportCode,
				ConfigJson: r.ConfigJson,
				SetParam:   r.SetParam,
				CreatedBy:  operator,
				UpdatedBy:  operator,
			}
			return tx.Create(&cfg).Error
		}
		if err != nil {
			return err
		}
		return tx.Model(&report.ReportAnalysisConfig{}).Where("id = ?", cfg.ID).Updates(map[string]interface{}{
			"config_json": r.ConfigJson,
			"set_param":   r.SetParam,
			"updated_by":  operator,
		}).Error
	})
}

// PreviewAnalysisResp 预览取数响应（明细 + 类型推断 + 截断标志）
type PreviewAnalysisResp struct {
	Columns   []repReq.ColumnMeta      `json:"columns"`
	Rows      []map[string]interface{} `json:"rows"`
	Total     int64                    `json:"total"`
	Truncated bool                     `json:"truncated"`
}

// Preview 预览取数：返回单数据集明细（不聚合）；超限（>50000）回退 QueryPage 截断置 truncated
func (s *AnalysisReportService) Preview(r *repReq.PreviewAnalysisReq) (*PreviewAnalysisResp, error) {
	var main report.ReportAnalysisReport
	if err := global.GVA_DB.Where("report_code = ?", r.ReportCode).First(&main).Error; err != nil {
		return nil, ErrAnalysisReportNotExists
	}
	// 数据集存在且启用（禁用报表允许设计器调试，预览页由前端先校验状态）
	set, transforms, params, err := DataSetServiceApp.LoadForQuery(main.SetCode)
	if err != nil {
		return nil, err
	}
	if !set.EnableFlag {
		return nil, ErrAnalysisDataSetDisabled
	}
	resolved, err := ResolveSetParam(params, r.ParamValues)
	if err != nil {
		return nil, err
	}
	// 全量明细（受 50000 上限）；超限回退分页截断
	result, err := QueryServiceApp.Execute(set.SetType, set.SourceCode, set.DynSentence, resolved, transforms)
	truncated := false
	if err != nil {
		if !errors.Is(err, ErrSetResultTooLarge) {
			return nil, err
		}
		// 含转换器数据集超限时 QueryPage 同样全量触顶（已知限制，直接上抛）
		result, _, err = QueryServiceApp.QueryPage(set.SetType, set.SourceCode, set.DynSentence, resolved, 1, maxQueryRows, transforms)
		if err != nil {
			return nil, err
		}
		truncated = true
	} else {
		result, err = RunTransforms(result, transforms)
		if err != nil {
			return nil, err
		}
	}
	return &PreviewAnalysisResp{
		Columns:   InferColumns(result.Columns, result.Rows),
		Rows:      result.Rows,
		Total:     int64(len(result.Rows)),
		Truncated: truncated,
	}, nil
}

// InferColumns 字段类型推断（v1.2 终态语义）：
// number：前 50 行非空样本中 ≥80% 可转数字（数值类型直判；字符串 strconv.ParseFloat，容忍空值）；
// date：非 number 且首行值匹配日期头；其余 string
func InferColumns(columns []string, rows []map[string]interface{}) []repReq.ColumnMeta {
	metas := make([]repReq.ColumnMeta, 0, len(columns))
	for _, col := range columns {
		metas = append(metas, repReq.ColumnMeta{Name: col, Type: inferColumnType(col, rows)})
	}
	return metas
}

func inferColumnType(col string, rows []map[string]interface{}) string {
	sample := make([]interface{}, 0, 50)
	for _, row := range rows {
		if v, ok := row[col]; ok && v != nil && fmt.Sprintf("%v", v) != "" {
			sample = append(sample, v)
		}
		if len(sample) >= 50 {
			break
		}
	}
	if len(sample) == 0 {
		return "string"
	}
	numberCount := 0
	for _, v := range sample {
		if isNumberValue(v) {
			numberCount++
		}
	}
	if float64(numberCount)/float64(len(sample)) >= 0.8 {
		return "number"
	}
	first := fmt.Sprintf("%v", sample[0])
	if dateHeadPattern.MatchString(first) {
		return "date"
	}
	return "string"
}

func isNumberValue(v interface{}) bool {
	switch t := v.(type) {
	case int, int64, float64, float32:
		return true
	case string:
		_, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return err == nil
	default:
		return false
	}
}

// IsSetCodeReferencedByAnalysis 数据集删除引用校验（等值匹配，挂 02 DeleteDataSet）
func (s *AnalysisReportService) IsSetCodeReferencedByAnalysis(setCode string) (bool, error) {
	var count int64
	if err := global.GVA_DB.Model(&report.ReportAnalysisReport{}).
		Where("set_code = ?", setCode).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *AnalysisReportService) checkCodeUnique(db *gorm.DB, reportCode string, excludeID uint) error {
	var count int64
	q := db.Model(&report.ReportAnalysisReport{}).Where("report_code = ?", reportCode)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrAnalysisReportCodeDuplicate
	}
	return nil
}

func (s *AnalysisReportService) checkDataSetUsable(setCode string) error {
	var set report.ReportDataSet
	if err := global.GVA_DB.Where("set_code = ?", setCode).First(&set).Error; err != nil {
		return ErrAnalysisDataSetNotExists
	}
	if !set.EnableFlag {
		return ErrAnalysisDataSetDisabled
	}
	return nil
}
