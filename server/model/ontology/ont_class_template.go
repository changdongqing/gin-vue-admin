package ontology

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// OntClassTemplate 本体分类模板（树表，物化路径；外观+结构骨架+父子继承+编码）
// 唯一编码/分类编码由迁移 SQL 部分唯一索引 + Service 校验共同保证，不声明 uniqueIndex tag（以迁移为准）
type OntClassTemplate struct {
	global.GVA_MODEL
	TemplateCode         string                `json:"templateCode" form:"templateCode" gorm:"column:template_code;size:64;not null;comment:模板编码（唯一）" binding:"required"` //nolint:stylecheck // 驼峰 JSON 契约
	ClassificationCode   string                `json:"classificationCode" form:"classificationCode" gorm:"column:classification_code;size:64;comment:规范分类编码（空则自动生成）"`
	Label                string                `json:"label" form:"label" gorm:"column:label;size:128;not null;comment:显示名" binding:"required"`
	LabelCn              string                `json:"labelCn" form:"labelCn" gorm:"column:label_cn;size:128;comment:中文名"`
	Description          string                `json:"description" form:"description" gorm:"column:description;size:512;comment:描述"`
	TreeRoot             string                `json:"treeRoot" form:"treeRoot" gorm:"column:tree_root;size:64;not null;comment:分类树标识" binding:"required"`
	ParentId             uint                  `json:"parentId" form:"parentId" gorm:"column:parent_id;default:0;comment:父节点ID 根=0"`
	ParentPath           string                `json:"parentPath" form:"parentPath" gorm:"column:parent_path;size:767;comment:物化路径 含父链不含自身"`
	TreeLevel            int                   `json:"treeLevel" form:"treeLevel" gorm:"column:tree_level;default:0;comment:层级 根=0 上限3"`
	Sort                 int                   `json:"sort" form:"sort" gorm:"column:sort;default:0;comment:排序"`
	Icon                 string                `json:"icon" form:"icon" gorm:"column:icon;size:100;comment:图标"`
	Color                string                `json:"color" form:"color" gorm:"column:color;size:50;comment:颜色hex"`
	InheritAppearance    int                   `json:"inheritAppearance" form:"inheritAppearance" gorm:"column:inherit_appearance;default:1;comment:继承父外观 0/1"`
	Source               string                `json:"source" form:"source" gorm:"column:source;size:16;not null;default:custom;comment:来源 builtin/custom"`
	SourceRef            string                `json:"sourceRef" form:"sourceRef" gorm:"column:source_ref;size:64;comment:来源本体标识"`
	Deprecated           int                   `json:"deprecated" form:"deprecated" gorm:"column:deprecated;default:0;comment:弃用 0/1"`
	Status               int                   `json:"status" form:"status" gorm:"column:status;default:0;comment:启停 0/1"`
	CreatedBy            string                `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:64;comment:创建人"`
	UpdatedBy            string                `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:64;comment:更新人"`
	Children             []OntClassTemplate    `json:"children" gorm:"-"`             // 前端组树用
	OntClassTemplateRefs []OntClassTemplateRef `json:"ontClassTemplateRefs" gorm:"-"` // 骨架子表随表单提交/详情返回
}

func (OntClassTemplate) TableName() string {
	return "ont_class_templates"
}

// OntClassTemplateRef 结构骨架引用（子表）
type OntClassTemplateRef struct {
	global.GVA_MODEL
	ClassTemplateId      uint   `json:"classTemplateId" form:"classTemplateId" gorm:"column:class_template_id;not null;comment:所属分类模板ID"` //nolint:stylecheck // 对齐前端 classTemplateId
	PropertyTemplateCode string `json:"propertyTemplateCode" form:"propertyTemplateCode" gorm:"column:property_template_code;size:64;not null;comment:属性模板编码" binding:"required"`
	RefType              string `json:"refType" form:"refType" gorm:"column:ref_type;size:16;default:property;comment:引用类型 property/relationship"`
	SortOrder            int    `json:"sortOrder" form:"sortOrder" gorm:"column:sort_order;default:0;comment:注入顺序"`
}

func (OntClassTemplateRef) TableName() string {
	return "ont_class_template_refs"
}

// OntClassificationRule 分类编码规则（每棵树一条，FR-8）
type OntClassificationRule struct {
	global.GVA_MODEL
	TreeRoot    string `json:"treeRoot" form:"treeRoot" gorm:"column:tree_root;size:64;not null;comment:分类树标识（唯一）" binding:"required"`
	Separator   string `json:"separator" form:"separator" gorm:"column:separator;size:4;default:-;comment:分隔符"`
	LevelDigits int    `json:"levelDigits" form:"levelDigits" gorm:"column:level_digits;default:2;comment:每级位数"`
	BaseNumber  int    `json:"baseNumber" form:"baseNumber" gorm:"column:base_number;default:30;comment:根级编码基数"`
	ZeroPad     int    `json:"zeroPad" form:"zeroPad" gorm:"column:zero_pad;default:1;comment:零填充 0/1"`
	Description string `json:"description" form:"description" gorm:"column:description;size:512;comment:描述"`
}

func (OntClassificationRule) TableName() string {
	return "ont_classification_rules"
}

// OntClassHierarchy 类层级镜像（本期建表预留，不实现读写）
type OntClassHierarchy struct {
	global.GVA_MODEL
	ChildClassIri      string     `json:"childClassIri" gorm:"column:child_class_iri;size:255;comment:子类IRI（建模侧）"`
	ParentClassIri     string     `json:"parentClassIri" gorm:"column:parent_class_iri;size:255;comment:父类IRI"`
	SourceTemplateCode string     `json:"sourceTemplateCode" gorm:"column:source_template_code;size:64;comment:溯源模板编码"`
	TreeRoot           string     `json:"treeRoot" gorm:"column:tree_root;size:64;comment:所属类树"`
	SyncStatus         int        `json:"syncStatus" gorm:"column:sync_status;default:0;comment:0待同步/1已同步/2已失效"`
	SyncTime           *time.Time `json:"syncTime" gorm:"column:sync_time;comment:最近同步时间"`
}

func (OntClassHierarchy) TableName() string {
	return "ont_class_hierarchies"
}
