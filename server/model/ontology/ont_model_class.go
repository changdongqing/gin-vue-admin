package ontology

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// OntModelClass 本体类实体（owl:Class，模板实例化/空白创建）
// 项目内 classIri 唯一由迁移 SQL 部分唯一索引 + Service 查重共同保证
type OntModelClass struct {
	global.GVA_MODEL
	ProjectId          uint   `json:"projectId" form:"projectId" gorm:"column:project_id;not null;comment:所属项目ID"` //nolint:stylecheck // 对齐前端 projectId；required 由 Service 校验（更新按白名单部分提交）
	ClassIri           string `json:"classIri" form:"classIri" gorm:"column:class_iri;size:255;not null;comment:类IRI"`
	LocalName          string `json:"localName" form:"localName" gorm:"column:local_name;size:128;not null;comment:本地名(不可改)"`
	Label              string `json:"label" form:"label" gorm:"column:label;size:128;comment:英文标签"`
	LabelCn            string `json:"labelCn" form:"labelCn" gorm:"column:label_cn;size:128;comment:中文标签"`
	Description        string `json:"description" form:"description" gorm:"column:description;size:512;comment:描述"`
	TemplateCode       string `json:"templateCode" form:"templateCode" gorm:"column:template_code;size:64;comment:溯源:分类模板编码"`
	ClassificationCode string `json:"classificationCode" form:"classificationCode" gorm:"column:classification_code;size:64;comment:溯源:分类编码"`
	Icon               string `json:"icon" form:"icon" gorm:"column:icon;size:100;comment:图标"`
	Color              string `json:"color" form:"color" gorm:"column:color;size:50;comment:颜色"`
	SortOrder          int    `json:"sortOrder" form:"sortOrder" gorm:"column:sort_order;default:0;comment:排序"`
	IsInstantiable     int    `json:"isInstantiable" form:"isInstantiable" gorm:"column:is_instantiable;default:0;comment:实例化0/1"`
	Table              string `json:"tableName" form:"tableName" gorm:"column:table_name;size:128;comment:预留:对象域表名"` //nolint:stylecheck // 字段名 Table 避让 TableName() 方法；JSON 契约保持 tableName 对齐前端
	CreatedBy          string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy          string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (OntModelClass) TableName() string {
	return "ont_model_classes"
}

// OntModelDatatypeProperty 本体数据属性（owl:DatatypeProperty，按类独立副本/方案B）
type OntModelDatatypeProperty struct {
	global.GVA_MODEL
	ProjectId      uint   `json:"projectId" form:"projectId" gorm:"column:project_id;not null;comment:所属项目ID"` //nolint:stylecheck // 对齐前端 projectId
	ClassId        uint   `json:"classId" form:"classId" gorm:"column:class_id;not null;comment:所属类ID"`
	PropertyIri    string `json:"propertyIri" form:"propertyIri" gorm:"column:property_iri;size:255;not null;comment:属性IRI"`
	LocalName      string `json:"localName" form:"localName" gorm:"column:local_name;size:128;not null;comment:本地名(不可改)"`
	Label          string `json:"label" form:"label" gorm:"column:label;size:128;comment:显示名"`
	TemplateCode   string `json:"templateCode" form:"templateCode" gorm:"column:template_code;size:64;comment:溯源:属性模板编码"`
	XsdType        string `json:"xsdType" form:"xsdType" gorm:"column:xsd_type;size:64;not null;default:xsd:string;comment:XSD类型" binding:"omitempty,oneof=xsd:string xsd:integer xsd:decimal xsd:boolean xsd:datetime"`
	UnitRef        string `json:"unitRef" form:"unitRef" gorm:"column:unit_ref;size:255;comment:单位引用(QUDT IRI)"`
	EnumValues     string `json:"enumValues" form:"enumValues" gorm:"column:enum_values;type:text;comment:枚举值(JSON数组字符串)"`
	MinCardinality int    `json:"minCardinality" form:"minCardinality" gorm:"column:min_cardinality;default:0;comment:最小基数"`
	MaxCardinality int    `json:"maxCardinality" form:"maxCardinality" gorm:"column:max_cardinality;default:-1;comment:最大基数(-1=无限制)"`
	IsIdentifier   int    `json:"isIdentifier" form:"isIdentifier" gorm:"column:is_identifier;default:0;comment:是否标识0/1"`
	SortOrder      int    `json:"sortOrder" form:"sortOrder" gorm:"column:sort_order;default:0;comment:排序"`
	CreatedBy      string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy      string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (OntModelDatatypeProperty) TableName() string {
	return "ont_model_datatype_properties"
}

// OntModelObjectProperty 本体对象属性（owl:ObjectProperty，domain=所属类）
type OntModelObjectProperty struct {
	global.GVA_MODEL
	ProjectId      uint   `json:"projectId" form:"projectId" gorm:"column:project_id;not null;comment:所属项目ID"`                  //nolint:stylecheck // 对齐前端 projectId
	DomainClassId  uint   `json:"domainClassId" form:"domainClassId" gorm:"column:domain_class_id;not null;comment:域类ID(=所属类)"` //nolint:stylecheck // 对齐前端 domainClassId
	RangeClassId   *uint  `json:"rangeClassId" form:"rangeClassId" gorm:"column:range_class_id;comment:值域类ID(可空)"`              //nolint:stylecheck // 对齐前端 rangeClassId
	PropertyIri    string `json:"propertyIri" form:"propertyIri" gorm:"column:property_iri;size:255;not null;comment:属性IRI"`
	LocalName      string `json:"localName" form:"localName" gorm:"column:local_name;size:128;not null;comment:本地名(不可改)"`
	Label          string `json:"label" form:"label" gorm:"column:label;size:128;comment:显示名"`
	TemplateCode   string `json:"templateCode" form:"templateCode" gorm:"column:template_code;size:64;comment:溯源:属性模板编码"`
	MinCardinality int    `json:"minCardinality" form:"minCardinality" gorm:"column:min_cardinality;default:0;comment:最小基数"`
	MaxCardinality int    `json:"maxCardinality" form:"maxCardinality" gorm:"column:max_cardinality;default:-1;comment:最大基数(-1=无限制)"`
	InverseOf      *uint  `json:"inverseOf" form:"inverseOf" gorm:"column:inverse_of;comment:反向属性ID"`
	SortOrder      int    `json:"sortOrder" form:"sortOrder" gorm:"column:sort_order;default:0;comment:排序"`
	CreatedBy      string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy      string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (OntModelObjectProperty) TableName() string {
	return "ont_model_object_properties"
}

// OntModelSubclassOf subClassOf 类层级关系（建模域权威源）
// 本期无写入入口：仅用于类删除守卫与类详情「层级」只读聚合；层级编辑/镜像回推为后续需求
type OntModelSubclassOf struct {
	global.GVA_MODEL
	ProjectId          uint       `json:"projectId" form:"projectId" gorm:"column:project_id;not null;comment:所属项目ID"`            //nolint:stylecheck // 对齐前端 projectId
	ChildClassId       uint       `json:"childClassId" form:"childClassId" gorm:"column:child_class_id;not null;comment:子类ID"`    //nolint:stylecheck // 对齐前端 childClassId
	ParentClassId      uint       `json:"parentClassId" form:"parentClassId" gorm:"column:parent_class_id;not null;comment:父类ID"` //nolint:stylecheck // 对齐前端 parentClassId
	SourceTemplateCode string     `json:"sourceTemplateCode" form:"sourceTemplateCode" gorm:"column:source_template_code;size:64;comment:来源分类模板编码"`
	SyncStatus         int        `json:"syncStatus" form:"syncStatus" gorm:"column:sync_status;default:0;comment:镜像回推状态0/1/2"`
	SyncTime           *time.Time `json:"syncTime" form:"syncTime" gorm:"column:sync_time;comment:回推时间"`
	RetryCount         int        `json:"retryCount" form:"retryCount" gorm:"column:retry_count;default:0;comment:重试次数"`
	CreatedBy          string     `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy          string     `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (OntModelSubclassOf) TableName() string {
	return "ont_model_subclassofs"
}
