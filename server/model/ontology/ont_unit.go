package ontology

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/shopspring/decimal"
)

// 与前端 number + precision 15 直连：decimal JSON 默认带引号字符串，这里按数字序列化
func init() {
	decimal.MarshalJSONWithoutQuotes = true
}

// OntQuantityKind 量纲（引用 QUDT，只读；无量纲 CRUD，扩充走迁移种子）
type OntQuantityKind struct {
	global.GVA_MODEL
	QuantityKindCode string `json:"quantityKindCode" form:"quantityKindCode" gorm:"column:quantity_kind_code;size:64;not null;comment:量纲标识"`
	QudtIri          string `json:"qudtIri" form:"qudtIri" gorm:"column:qudt_iri;size:255;not null;comment:QUDT量纲IRI"`
	Label            string `json:"label" form:"label" gorm:"column:label;size:64;not null;comment:英文显示名"`
	LabelCn          string `json:"labelCn" form:"labelCn" gorm:"column:label_cn;size:64;comment:中文显示名"`
	DimensionVector  string `json:"dimensionVector" form:"dimensionVector" gorm:"column:dimension_vector;size:32;comment:量纲向量"`
	Sort             int    `json:"sort" form:"sort" gorm:"column:sort;default:0;comment:排序"`
}

func (OntQuantityKind) TableName() string {
	return "ont_quantity_kinds"
}

// OntUnit 单位（引用 QUDT，含线性换算参数；基准值 = 数值 × multiplier + offset）
type OntUnit struct {
	global.GVA_MODEL
	UnitCode             string          `json:"unitCode" form:"unitCode" gorm:"column:unit_code;size:64;not null;comment:单位标识（唯一）" binding:"required"`
	QudtIri              string          `json:"qudtIri" form:"qudtIri" gorm:"column:qudt_iri;size:255;not null;comment:QUDT单位IRI（唯一）" binding:"required"`
	Symbol               string          `json:"symbol" form:"symbol" gorm:"column:symbol;size:32;comment:符号"`
	Label                string          `json:"label" form:"label" gorm:"column:label;size:64;not null;comment:英文显示名" binding:"required"`
	LabelCn              string          `json:"labelCn" form:"labelCn" gorm:"column:label_cn;size:64;comment:中文显示名"`
	QuantityKindCode     string          `json:"quantityKindCode" form:"quantityKindCode" gorm:"column:quantity_kind_code;size:64;not null;comment:所属量纲code" binding:"required"`
	ConversionMultiplier decimal.Decimal `json:"conversionMultiplier" form:"conversionMultiplier" gorm:"column:conversion_multiplier;type:numeric(30,15);comment:换算乘数"`
	ConversionOffset     decimal.Decimal `json:"conversionOffset" form:"conversionOffset" gorm:"column:conversion_offset;type:numeric(30,15);comment:换算偏移"`
	ScalingOf            string          `json:"scalingOf" form:"scalingOf" gorm:"column:scaling_of;size:255;comment:基准单位IRI"`
	UcumCode             string          `json:"ucumCode" form:"ucumCode" gorm:"column:ucum_code;size:32;comment:UCUM编码"`
	Source               string          `json:"source" form:"source" gorm:"column:source;size:16;not null;default:custom;comment:来源 builtin/custom"`
	SourceRef            string          `json:"sourceRef" form:"sourceRef" gorm:"column:source_ref;size:64;comment:来源本体标识"`
	QudtVersion          string          `json:"qudtVersion" form:"qudtVersion" gorm:"column:qudt_version;size:32;comment:QUDT版本快照"`
	Status               int             `json:"status" form:"status" gorm:"column:status;default:0;comment:启停 0/1"`
	CreatedBy            string          `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy            string          `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (OntUnit) TableName() string {
	return "ont_units"
}

// ConvertUnitResp 换算响应（跨量纲/单位不存在不报错，result=null + reason）
type ConvertUnitResp struct {
	Result *decimal.Decimal `json:"result"` // null=不可换算
	Reason string           `json:"reason,omitempty"`
}
