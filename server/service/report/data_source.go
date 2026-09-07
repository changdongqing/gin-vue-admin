package report

import (
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/report"
	repReq "github.com/flipped-aurora/gin-vue-admin/server/model/report/request"
	"gorm.io/gorm"
)

type DataSourceService struct{}

// CreateDataSource 创建数据源：类型合法 → 编码查重 → 密码加密 → 审计 → 落库
func (s *DataSourceService) CreateDataSource(p *report.ReportDataSource, operator string) error {
	if !IsSourceTypeValid(p.SourceType) {
		return ErrSourceTypeInvalid
	}
	if err := s.checkCodeUnique(global.GVA_DB, p.SourceCode, 0); err != nil {
		return err
	}
	encrypted, err := EncryptSourceConfig(p.SourceConfig)
	if err != nil {
		return err
	}
	p.SourceConfig = encrypted
	p.CreatedBy, p.UpdatedBy = operator, operator
	return global.GVA_DB.Create(p).Error
}

// UpdateDataSource 更新数据源：存在性 → 编码查重排除自身 → 类型不可变 → 密码三态 → 落库销毁旧池
func (s *DataSourceService) UpdateDataSource(p *report.ReportDataSource, operator string) error {
	var old report.ReportDataSource
	if err := global.GVA_DB.First(&old, p.ID).Error; err != nil {
		return ErrSourceNotExists
	}
	if err := s.checkCodeUnique(global.GVA_DB, p.SourceCode, p.ID); err != nil {
		return err
	}
	if p.SourceType != old.SourceType {
		return errors.New("数据源类型不允许修改，如需变更请新建数据源")
	}
	if !IsSourceTypeValid(p.SourceType) {
		return ErrSourceTypeInvalid
	}
	encrypted, err := HandlePasswordOnUpdate(p.SourceConfig, old.SourceConfig)
	if err != nil {
		return err
	}
	if err := global.GVA_DB.Model(&report.ReportDataSource{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
		"source_code":   p.SourceCode,
		"source_name":   p.SourceName,
		"source_desc":   p.SourceDesc,
		"source_config": encrypted,
		"enable_flag":   p.EnableFlag,
		"updated_by":    operator,
	}).Error; err != nil {
		return err
	}
	GlobalPoolManager.Remove(p.SourceCode) // 配置变更销毁旧池，下次取数自动重建
	return nil
}

// DeleteDataSource 删除数据源：存在性 → 引用校验（02 落地 report_data_sets 计数）→ 销毁旧池 → 软删
func (s *DataSourceService) DeleteDataSource(id uint) error {
	var old report.ReportDataSource
	if err := global.GVA_DB.First(&old, id).Error; err != nil {
		return ErrSourceNotExists
	}
	if err := s.validateNotReferenced(old.SourceCode); err != nil {
		return err
	}
	if err := global.GVA_DB.Delete(&report.ReportDataSource{}, id).Error; err != nil {
		return err
	}
	GlobalPoolManager.Remove(old.SourceCode)
	return nil
}

// GetDataSource 详情（脱敏返回）
func (s *DataSourceService) GetDataSource(id uint) (p report.ReportDataSource, err error) {
	if err = global.GVA_DB.First(&p, id).Error; err != nil {
		return report.ReportDataSource{}, ErrSourceNotExists
	}
	p.SourceConfig = MaskSourceConfig(p.SourceConfig)
	return
}

// GetDataSourceByCode 按编码查询（内部使用：配置保持密文，由取数方解密）
func (s *DataSourceService) GetDataSourceByCode(code string) (p report.ReportDataSource, err error) {
	err = global.GVA_DB.Where("source_code = ?", code).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return report.ReportDataSource{}, ErrSourceNotExists
	}
	return
}

// GetDataSourceList 分页查询（keyword 匹配编码/名称 + 类型/启用状态过滤；脱敏返回）
func (s *DataSourceService) GetDataSourceList(info repReq.SearchDataSource) (list []report.ReportDataSource, total int64, err error) {
	db := global.GVA_DB.Model(&report.ReportDataSource{})
	if info.SourceType != "" {
		db = db.Where("source_type = ?", info.SourceType)
	}
	if info.EnableFlag != nil {
		db = db.Where("enable_flag = ?", *info.EnableFlag)
	}
	if info.Keyword != "" {
		kw := "%" + info.Keyword + "%"
		db = db.Where("source_code LIKE ? OR source_name LIKE ?", kw, kw)
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	if err = db.Order("id DESC").Limit(info.PageSize).Offset((info.Page - 1) * info.PageSize).Find(&list).Error; err != nil {
		return
	}
	for i := range list {
		list[i].SourceConfig = MaskSourceConfig(list[i].SourceConfig)
	}
	return
}

// GetEnabledDataSourceAll 已启用全量下拉（02 数据集编辑选择；置空连接配置，仅编码/名称/类型）
func (s *DataSourceService) GetEnabledDataSourceAll() (list []report.ReportDataSource, err error) {
	if err = global.GVA_DB.Where("enable_flag = ?", true).Order("id ASC").Find(&list).Error; err != nil {
		return
	}
	for i := range list {
		list[i].SourceConfig = ""
	}
	return
}

// TestConnection 测试连接（前端明文表单值，不经加密；不写入池缓存）
func (s *DataSourceService) TestConnection(sourceType, sourceConfig string) error {
	if !IsSourceTypeValid(sourceType) {
		return ErrSourceTypeInvalid
	}
	if _, err := parseSourceConfig(sourceConfig); err != nil {
		return err
	}
	return GlobalPoolManager.TestConnection(sourceType, sourceConfig)
}

func (s *DataSourceService) checkCodeUnique(db *gorm.DB, sourceCode string, excludeID uint) error {
	var count int64
	q := db.Model(&report.ReportDataSource{}).Where("source_code = ?", sourceCode)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrSourceCodeDuplicate
	}
	return nil
}

// validateNotReferenced 删除引用校验：被 report_data_sets.source_code 引用时拒绝删除（02 落地）
func (s *DataSourceService) validateNotReferenced(sourceCode string) error {
	var count int64
	if err := global.GVA_DB.Model(&report.ReportDataSet{}).
		Where("source_code = ?", sourceCode).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrSourceReferenced
	}
	return nil
}
