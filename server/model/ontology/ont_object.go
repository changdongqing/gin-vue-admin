package ontology

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// OntObject 本体对象个体（ABox；本期由同步物化，对象管理页后续交付）
type OntObject struct {
	global.GVA_MODEL
	ProjectId    uint       `json:"projectId" gorm:"column:project_id;not null;comment:所属项目ID"` //nolint:stylecheck // 对齐前端
	ClassId      uint       `json:"classId" gorm:"column:class_id;not null;comment:所属类ID"`      //nolint:stylecheck // 对齐前端
	ObjectCode   string     `json:"objectCode" gorm:"column:object_code;size:128;not null;comment:对象编码(项目内唯一)"`
	ObjectName   string     `json:"objectName" gorm:"column:object_name;size:255;comment:对象名称"`
	ParentId     *uint      `json:"parentId" gorm:"column:parent_id;comment:父对象ID"` //nolint:stylecheck // 对齐前端
	State        int        `json:"state" gorm:"column:state;default:1;comment:1在用/2停用/3失联"`
	SourceType   int        `json:"sourceType" gorm:"column:source_type;default:2;comment:1本体系/2业务同步"` //nolint:stylecheck // 对齐前端
	BizTable     string     `json:"bizTable" gorm:"column:biz_table;size:128;comment:来源业务表"`           //nolint:stylecheck // 对齐前端
	BizKey       string     `json:"bizKey" gorm:"column:biz_key;size:128;comment:来源业务主键值"`             //nolint:stylecheck // 对齐前端
	LastSyncTime *time.Time `json:"lastSyncTime" gorm:"column:last_sync_time;comment:最近同步时间"`          //nolint:stylecheck // 对齐前端
	CreatedBy    string     `json:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy    string     `json:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
}

func (OntObject) TableName() string {
	return "ont_objects"
}

// OntObjectAttrValue 对象数据属性值（EAV；对象属性值走 ont_object_relations）
type OntObjectAttrValue struct {
	global.GVA_MODEL
	ObjectId   uint   `json:"objectId" gorm:"column:object_id;not null;comment:对象ID"`       //nolint:stylecheck // 对齐前端
	PropertyId uint   `json:"propertyId" gorm:"column:property_id;not null;comment:数据属性ID"` //nolint:stylecheck // 对齐前端
	ValueJson  string `json:"valueJson" gorm:"column:value_json;type:text;comment:JSON标量"`  //nolint:stylecheck // 对齐前端
}

func (OntObjectAttrValue) TableName() string {
	return "ont_object_attr_values"
}

// OntObjectRelation 对象关系（对象属性实例化）
type OntObjectRelation struct {
	global.GVA_MODEL
	ObjectId         uint `json:"objectId" gorm:"column:object_id;not null;comment:主体对象ID"`                  //nolint:stylecheck // 对齐前端
	ObjectPropertyId uint `json:"objectPropertyId" gorm:"column:object_property_id;not null;comment:对象属性ID"` //nolint:stylecheck // 对齐前端
	TargetObjectId   uint `json:"targetObjectId" gorm:"column:target_object_id;not null;comment:目标对象ID"`     //nolint:stylecheck // 对齐前端
}

func (OntObjectRelation) TableName() string {
	return "ont_object_relations"
}
