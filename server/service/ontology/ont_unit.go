package ontology

import (
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
	ontReq "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/request"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type QuantityKindService struct{}

// GetQuantityKindList 量纲全量（Order sort），只读
func (s *QuantityKindService) GetQuantityKindList() (list []ontology.OntQuantityKind, err error) {
	err = global.GVA_DB.Order("sort ASC, id ASC").Find(&list).Error
	return
}

type UnitService struct{}

// CreateUnit 创建（量纲存在 + 双唯一校验；强制 source=custom）
func (s *UnitService) CreateUnit(u *ontology.OntUnit, operator string) error {
	if err := s.validateUnitUnique(global.GVA_DB, u, 0); err != nil {
		return err
	}
	if err := s.validateQuantityKindExists(global.GVA_DB, u.QuantityKindCode); err != nil {
		return err
	}
	u.Source = "custom"
	u.Status = 0
	u.CreatedBy, u.UpdatedBy = operator, operator
	return global.GVA_DB.Create(u).Error
}

// UpdateUnit 更新（builtin 拒绝；量纲存在 + 双唯一校验）
func (s *UnitService) UpdateUnit(u *ontology.OntUnit, operator string) error {
	db := global.GVA_DB
	var existing ontology.OntUnit
	if err := db.First(&existing, u.ID).Error; err != nil {
		return errors.New("单位不存在")
	}
	if existing.Source == "builtin" {
		return errors.New("内置单位不可编辑")
	}
	if err := s.validateUnitUnique(db, u, u.ID); err != nil {
		return err
	}
	if err := s.validateQuantityKindExists(db, u.QuantityKindCode); err != nil {
		return err
	}
	u.UpdatedBy = operator
	return db.Model(&ontology.OntUnit{}).Where("id = ?", u.ID).Updates(map[string]interface{}{
		"unit_code":             u.UnitCode,
		"qudt_iri":              u.QudtIri,
		"symbol":                u.Symbol,
		"label":                 u.Label,
		"label_cn":              u.LabelCn,
		"quantity_kind_code":    u.QuantityKindCode,
		"conversion_multiplier": u.ConversionMultiplier,
		"conversion_offset":     u.ConversionOffset,
		"scaling_of":            u.ScalingOf,
		"ucum_code":             u.UcumCode,
		"qudt_version":          u.QudtVersion,
		"updated_by":            operator,
	}).Error
}

// DeleteUnit 删除（builtin 拒绝；软删除）
func (s *UnitService) DeleteUnit(id uint) error {
	var existing ontology.OntUnit
	if err := global.GVA_DB.First(&existing, id).Error; err != nil {
		return errors.New("单位不存在")
	}
	if existing.Source == "builtin" {
		return errors.New("内置单位不可删除")
	}
	return global.GVA_DB.Delete(&ontology.OntUnit{}, id).Error
}

// DisableUnit 停用/启用（幂等双向切换；builtin 可停用）
func (s *UnitService) DisableUnit(id uint, operator string) error {
	var existing ontology.OntUnit
	if err := global.GVA_DB.First(&existing, id).Error; err != nil {
		return errors.New("单位不存在")
	}
	target := 1 - existing.Status
	return global.GVA_DB.Model(&ontology.OntUnit{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": target, "updated_by": operator}).Error
}

// GetUnit 详情
func (s *UnitService) GetUnit(id uint) (u ontology.OntUnit, err error) {
	err = global.GVA_DB.First(&u, id).Error
	return
}

// GetUnitPage 分页查询（keyword 匹配 unitCode/label/labelCn/symbol 四字段）
func (s *UnitService) GetUnitPage(info ontReq.SearchUnit) (list []ontology.OntUnit, total int64, err error) {
	db := s.buildPageQuery(info)
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Order("id DESC").Limit(info.PageSize).Offset((info.Page - 1) * info.PageSize).Find(&list).Error
	return
}

func (s *UnitService) buildPageQuery(info ontReq.SearchUnit) *gorm.DB {
	db := global.GVA_DB.Model(&ontology.OntUnit{})
	if info.QuantityKindCode != "" {
		db = db.Where("quantity_kind_code = ?", info.QuantityKindCode)
	}
	if info.Source != "" {
		db = db.Where("source = ?", info.Source)
	}
	if info.Status != nil {
		db = db.Where("status = ?", *info.Status)
	}
	if info.Keyword != "" {
		kw := "%" + info.Keyword + "%"
		db = db.Where("unit_code LIKE ? OR label LIKE ? OR label_cn LIKE ? OR symbol LIKE ?", kw, kw, kw, kw)
	}
	return db
}

// GetUnitAll 全量下拉（仅启用；可选按量纲过滤）
func (s *UnitService) GetUnitAll(quantityKindCode string) (list []ontology.OntUnit, err error) {
	db := global.GVA_DB.Where("status = 0")
	if quantityKindCode != "" {
		db = db.Where("quantity_kind_code = ?", quantityKindCode)
	}
	err = db.Order("quantity_kind_code ASC, id ASC").Find(&list).Error
	return
}

// GetUnitsForSupply 供给（仅启用，含换算系数）
func (s *UnitService) GetUnitsForSupply(quantityKindCode string) (list []ontology.OntUnit, err error) {
	return s.GetUnitAll(quantityKindCode)
}

// ConvertValue 单位换算（FR-3 核心）
// 同量纲：result = (value × from.mult + from.offset − to.offset) ÷ to.mult（HALF_UP 保留 12 位小数）；
// 单位不存在/跨量纲：返回 (nil, reason, nil)，HTTP 仍 200 不报错。
// 舍入取 12 位小数：DEG_F 乘数 0.555555555555555 的倒数残差约 2e-13，15 位舍入会带出
// 212.000000000000212 类脏尾数，12 位可让 AC 验算值（0/373.15/212）精确成立。
func (s *UnitService) ConvertValue(value decimal.Decimal, fromIri, toIri string) (*decimal.Decimal, string, error) {
	var from, to ontology.OntUnit
	if err := global.GVA_DB.Where("qudt_iri = ?", fromIri).First(&from).Error; err != nil {
		return nil, "quantity kind mismatch or unit not exists", nil
	}
	if err := global.GVA_DB.Where("qudt_iri = ?", toIri).First(&to).Error; err != nil {
		return nil, "quantity kind mismatch or unit not exists", nil
	}
	if from.QuantityKindCode != to.QuantityKindCode {
		return nil, "quantity kind mismatch or unit not exists", nil
	}
	if fromIri == toIri {
		return &value, "", nil
	}
	fm, fo := nullSafeOne(from.ConversionMultiplier), nullSafeZero(from.ConversionOffset)
	tm, tof := nullSafeOne(to.ConversionMultiplier), nullSafeZero(to.ConversionOffset)
	if tm.IsZero() {
		return nil, "invalid multiplier", nil
	}
	base := value.Mul(fm).Add(fo)
	res := base.Sub(tof).DivRound(tm, 12)
	return &res, "", nil
}

func nullSafeOne(d decimal.Decimal) decimal.Decimal {
	if d.IsZero() {
		return decimal.NewFromInt(1)
	}
	return d
}

func nullSafeZero(d decimal.Decimal) decimal.Decimal {
	return d
}

func (s *UnitService) validateUnitUnique(db *gorm.DB, u *ontology.OntUnit, excludeID uint) error {
	var count int64
	q := db.Model(&ontology.OntUnit{}).Where("unit_code = ?", u.UnitCode)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("单位编码已存在")
	}
	q2 := db.Model(&ontology.OntUnit{}).Where("qudt_iri = ?", u.QudtIri)
	if excludeID > 0 {
		q2 = q2.Where("id <> ?", excludeID)
	}
	if err := q2.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("QUDT IRI 已存在")
	}
	return nil
}

func (s *UnitService) validateQuantityKindExists(db *gorm.DB, code string) error {
	var count int64
	if err := db.Model(&ontology.OntQuantityKind{}).Where("quantity_kind_code = ?", code).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("量纲不存在: " + code)
	}
	return nil
}
