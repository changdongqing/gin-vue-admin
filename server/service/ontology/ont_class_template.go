package ontology

import (
	"errors"
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ontology"
	ontReq "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/request"
	ontRes "github.com/flipped-aurora/gin-vue-admin/server/model/ontology/response"
	"gorm.io/gorm"
)

const MaxTreeLevel = 3 // 层级上限四级（treeLevel 0~3）

var errClassCodeDuplicated = errors.New("分类模板编码已存在")

type ClassTemplateService struct{}

// —— 树维护 ——

// validateAndFillTree 校验父链并填充 parent_path/tree_level（参照 sys_company 维护算法）
func (s *ClassTemplateService) validateAndFillTree(db *gorm.DB, t *ontology.OntClassTemplate) error {
	if t.ParentId == 0 { // 顶级
		t.ParentPath, t.TreeLevel = "0,", 0
		return nil
	}
	if t.ID > 0 && t.ParentId == t.ID {
		return errors.New("不能以自身为父节点")
	}
	var parent ontology.OntClassTemplate
	if err := db.First(&parent, t.ParentId).Error; err != nil {
		return errors.New("父分类模板不存在")
	}
	if parent.TreeRoot != t.TreeRoot {
		return errors.New("父节点属于其他分类树（" + parent.TreeRoot + "），不可跨树挂载")
	}
	// 环路检测：新父的祖先链（含父）不能包含自身 —— O(1) 物化路径包含判断
	selfID := "," + strconv.FormatUint(uint64(t.ID), 10) + ","
	if t.ID > 0 && strings.Contains(","+parent.ParentPath, selfID) {
		return errors.New("不能将自身或后代节点设为父节点（成环）")
	}
	if parent.TreeLevel+1 > MaxTreeLevel {
		return errors.New("分类模板层级最多四级")
	}
	t.ParentPath = parent.ParentPath + strconv.FormatUint(uint64(parent.ID), 10) + ","
	t.TreeLevel = parent.TreeLevel + 1
	return nil
}

// moveSubtree 改父后批量平移子树（一条 UPDATE：parent_path 前缀 REPLACE + tree_level 平移）
// oldSelf/newSelf 为移动前/后「含自身」的完整路径；delta 为子树层级平移量
func (s *ClassTemplateService) moveSubtree(tx *gorm.DB, oldSelf, newSelf string, delta int) error {
	return tx.Model(&ontology.OntClassTemplate{}).
		Where("parent_path LIKE ?", oldSelf+"%").
		Updates(map[string]interface{}{
			"parent_path": gorm.Expr("REPLACE(parent_path, ?, ?)", oldSelf, newSelf),
			"tree_level":  gorm.Expr("tree_level + ?", delta),
		}).Error
}

// —— 校验 ——

func (s *ClassTemplateService) validateTemplateCodeUnique(db *gorm.DB, code string, excludeID uint) error {
	var count int64
	q := db.Model(&ontology.OntClassTemplate{}).Where("template_code = ?", code)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errClassCodeDuplicated
	}
	return nil
}

// validateClassificationCode 手动编码校验：(treeRoot, code) 唯一 + 与父编码前缀一致（根节点免前缀校验）
func (s *ClassTemplateService) validateClassificationCode(db *gorm.DB, t *ontology.OntClassTemplate) error {
	var count int64
	q := db.Model(&ontology.OntClassTemplate{}).
		Where("tree_root = ? AND classification_code = ?", t.TreeRoot, t.ClassificationCode)
	if t.ID > 0 {
		q = q.Where("id <> ?", t.ID)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("分类编码在该分类树下已存在")
	}
	if t.ParentId == 0 {
		return nil
	}
	var parent ontology.OntClassTemplate
	if err := db.First(&parent, t.ParentId).Error; err != nil {
		return errors.New("父分类模板不存在")
	}
	if parent.ClassificationCode == "" { // 父无编码时免前缀校验
		return nil
	}
	rule, err := new(ClassificationRuleService).FindClassificationRule(t.TreeRoot)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(t.ClassificationCode, parent.ClassificationCode+rule.Separator) {
		return errors.New("分类编码须与父编码前缀一致（" + parent.ClassificationCode + rule.Separator + "…）")
	}
	return nil
}

// validateRefCodes 骨架引用校验：propertyTemplateCode 必须存在于属性模板库（含 builtin/custom、含弃用）
func (s *ClassTemplateService) validateRefCodes(db *gorm.DB, refs []ontology.OntClassTemplateRef) error {
	for _, r := range refs {
		if r.PropertyTemplateCode == "" {
			continue
		}
		var count int64
		if err := db.Model(&ontology.OntPropertyTemplate{}).
			Where("template_code = ?", r.PropertyTemplateCode).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("引用的属性模板不存在: " + r.PropertyTemplateCode)
		}
	}
	return nil
}

// syncRefs 骨架子表增量同步（按 ID diff：更新/新增/删除，禁止 delete-then-recreate）
func (s *ClassTemplateService) syncRefs(tx *gorm.DB, templateID uint, newList []ontology.OntClassTemplateRef) error {
	var oldList []ontology.OntClassTemplateRef
	if err := tx.Where("class_template_id = ?", templateID).Find(&oldList).Error; err != nil {
		return err
	}
	newIDs := make(map[uint]bool)
	for i := range newList {
		n := &newList[i]
		n.ClassTemplateId = templateID
		if n.ID > 0 { // 更新既有行
			newIDs[n.ID] = true
			if err := tx.Model(&ontology.OntClassTemplateRef{}).Where("id = ?", n.ID).
				Updates(map[string]interface{}{
					"property_template_code": n.PropertyTemplateCode,
					"ref_type":               n.RefType,
					"sort_order":             n.SortOrder,
				}).Error; err != nil {
				return err
			}
		} else { // 新增行
			if err := tx.Create(n).Error; err != nil {
				return err
			}
			newIDs[n.ID] = true
		}
	}
	for _, o := range oldList { // 删除被移除的行（软删）
		if !newIDs[o.ID] {
			if err := tx.Delete(&ontology.OntClassTemplateRef{}, o.ID).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// —— CRUD ——

// CreateClassTemplate 创建（校验 + 编码生成/校验 + 骨架落库）
func (s *ClassTemplateService) CreateClassTemplate(t *ontology.OntClassTemplate, operator string) error {
	db := global.GVA_DB
	if err := s.validateTemplateCodeUnique(db, t.TemplateCode, 0); err != nil {
		return err
	}
	if err := s.validateAndFillTree(db, t); err != nil {
		return err
	}
	if err := s.validateRefCodes(db, t.OntClassTemplateRefs); err != nil {
		return err
	}
	if t.ClassificationCode == "" {
		code, err := new(ClassificationRuleService).GenerateClassificationCode(db, t.ParentId, t.TreeRoot)
		if err != nil {
			return err
		}
		t.ClassificationCode = code
	} else if err := s.validateClassificationCode(db, t); err != nil {
		return err
	}
	t.Source, t.Deprecated, t.Status = "custom", 0, 0
	t.CreatedBy, t.UpdatedBy = operator, operator
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(t).Error; err != nil {
			return err
		}
		return s.syncRefs(tx, t.ID, t.OntClassTemplateRefs)
	})
}

// UpdateClassTemplate 更新（builtin 拒绝；treeRoot 不可变；骨架 diff；改父平移子树）
func (s *ClassTemplateService) UpdateClassTemplate(t *ontology.OntClassTemplate, operator string) error {
	db := global.GVA_DB
	var existing ontology.OntClassTemplate
	if err := db.First(&existing, t.ID).Error; err != nil {
		return errors.New("分类模板不存在")
	}
	if existing.Source == "builtin" {
		return errors.New("内置分类模板不可编辑")
	}
	if t.TreeRoot != existing.TreeRoot {
		return errors.New("分类树不可变更（换树请新建并废弃旧节点）")
	}
	if err := s.validateTemplateCodeUnique(db, t.TemplateCode, t.ID); err != nil {
		return err
	}
	if err := s.validateAndFillTree(db, t); err != nil {
		return err
	}
	if err := s.validateRefCodes(db, t.OntClassTemplateRefs); err != nil {
		return err
	}
	if t.ClassificationCode != "" {
		if err := s.validateClassificationCode(db, t); err != nil {
			return err
		}
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&ontology.OntClassTemplate{}).Where("id = ?", t.ID).Updates(map[string]interface{}{
			"template_code":       t.TemplateCode,
			"classification_code": t.ClassificationCode,
			"label":               t.Label,
			"label_cn":            t.LabelCn,
			"description":         t.Description,
			"parent_id":           t.ParentId,
			"parent_path":         t.ParentPath,
			"tree_level":          t.TreeLevel,
			"sort":                t.Sort,
			"icon":                t.Icon,
			"color":               t.Color,
			"inherit_appearance":  t.InheritAppearance,
			"deprecated":          t.Deprecated,
			"updated_by":          operator,
		}).Error; err != nil {
			return err
		}
		if err := s.syncRefs(tx, t.ID, t.OntClassTemplateRefs); err != nil {
			return err
		}
		if existing.ParentId != t.ParentId { // 改父：批量平移子树（delta 用移动前自身 level）
			oldSelf := existing.ParentPath + strconv.FormatUint(uint64(t.ID), 10) + ","
			newSelf := t.ParentPath + strconv.FormatUint(uint64(t.ID), 10) + "," // t.ParentPath/TreeLevel 已由 validateAndFillTree 按新父填充
			if err := s.moveSubtree(tx, oldSelf, newSelf, t.TreeLevel-existing.TreeLevel); err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteClassTemplate 删除（builtin/有子拒绝；级联软删骨架子表）
func (s *ClassTemplateService) DeleteClassTemplate(id uint) error {
	db := global.GVA_DB
	var existing ontology.OntClassTemplate
	if err := db.First(&existing, id).Error; err != nil {
		return errors.New("分类模板不存在")
	}
	if existing.Source == "builtin" {
		return errors.New("内置分类模板不可删除")
	}
	var childCount int64
	db.Model(&ontology.OntClassTemplate{}).Where("parent_id = ?", id).Count(&childCount)
	if childCount > 0 {
		return errors.New("存在子节点，不可直接删除")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&ontology.OntClassTemplate{}, id).Error; err != nil {
			return err
		}
		return tx.Where("class_template_id = ?", id).Delete(&ontology.OntClassTemplateRef{}).Error
	})
}

// DisableClassTemplate 弃用/取消弃用（幂等切换）
func (s *ClassTemplateService) DisableClassTemplate(id uint, operator string) error {
	var existing ontology.OntClassTemplate
	if err := global.GVA_DB.First(&existing, id).Error; err != nil {
		return errors.New("分类模板不存在")
	}
	target := 1 - existing.Deprecated
	return global.GVA_DB.Model(&ontology.OntClassTemplate{}).Where("id = ?", id).
		Updates(map[string]interface{}{"deprecated": target, "updated_by": operator}).Error
}

// —— 查询 ——

// GetClassTemplate 详情（含骨架子表，供编辑回填）
func (s *ClassTemplateService) GetClassTemplate(id uint) (t ontology.OntClassTemplate, err error) {
	if err = global.GVA_DB.First(&t, id).Error; err != nil {
		return
	}
	err = global.GVA_DB.Where("class_template_id = ?", id).Order("sort_order ASC, id ASC").Find(&t.OntClassTemplateRefs).Error
	return
}

// GetClassTemplateList 扁平全量（?treeRoot 过滤，前端组树/树选择器用）
func (s *ClassTemplateService) GetClassTemplateList(treeRoot string) (list []ontology.OntClassTemplate, err error) {
	db := global.GVA_DB.Model(&ontology.OntClassTemplate{})
	if treeRoot != "" {
		db = db.Where("tree_root = ?", treeRoot)
	}
	err = db.Order("tree_root ASC, parent_path ASC, sort ASC, id ASC").Find(&list).Error
	return
}

// GetClassTemplateRefList 骨架子表回填
func (s *ClassTemplateService) GetClassTemplateRefList(classTemplateId uint) (list []ontology.OntClassTemplateRef, err error) {
	err = global.GVA_DB.Where("class_template_id = ?", classTemplateId).Order("sort_order ASC, id ASC").Find(&list).Error
	return
}

// GetClassTemplatePage 分页 + 富化（parentLabel/treeRootLabel/hasChildren，三次批查避免 N+1）
func (s *ClassTemplateService) GetClassTemplatePage(info ontReq.SearchClassTemplate) (list []ontRes.ClassTemplatePageItem, total int64, err error) {
	db := global.GVA_DB.Model(&ontology.OntClassTemplate{})
	if info.TreeRoot != "" {
		db = db.Where("tree_root = ?", info.TreeRoot)
	}
	if info.ClassificationCode != "" {
		db = db.Where("classification_code LIKE ?", info.ClassificationCode+"%") // 右模糊
	}
	if info.Name != "" {
		n := "%" + info.Name + "%"
		db = db.Where("label LIKE ? OR label_cn LIKE ?", n, n)
	}
	if info.PageSize <= 0 {
		info.PageSize = 100
	}
	if info.Page <= 0 {
		info.Page = 1
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	var rows []ontology.OntClassTemplate
	if err = db.Order("tree_root ASC, parent_path ASC, sort ASC, id ASC").
		Limit(info.PageSize).Offset((info.Page - 1) * info.PageSize).Find(&rows).Error; err != nil {
		return
	}

	parentIds := make([]uint, 0)
	ids := make([]uint, 0)
	treeRoots := make([]string, 0)
	rootSeen := make(map[string]bool)
	for _, r := range rows {
		if r.ParentId > 0 {
			parentIds = append(parentIds, r.ParentId)
		}
		ids = append(ids, r.ID)
		if !rootSeen[r.TreeRoot] {
			rootSeen[r.TreeRoot] = true
			treeRoots = append(treeRoots, r.TreeRoot)
		}
	}

	parentMap := make(map[uint]string)
	if len(parentIds) > 0 {
		var parents []ontology.OntClassTemplate
		if err = global.GVA_DB.Select("id, label").Where("id IN ?", parentIds).Find(&parents).Error; err != nil {
			return
		}
		for _, p := range parents {
			parentMap[p.ID] = p.Label
		}
	}
	childMap := make(map[uint]int64)
	if len(ids) > 0 {
		type cntRow struct {
			ParentId uint
			Cnt      int64
		}
		var cnts []cntRow
		if err = global.GVA_DB.Model(&ontology.OntClassTemplate{}).
			Select("parent_id, count(*) AS cnt").Where("parent_id IN ?", ids).
			Group("parent_id").Find(&cnts).Error; err != nil {
			return
		}
		for _, c := range cnts {
			childMap[c.ParentId] = c.Cnt
		}
	}
	rootMap := make(map[string]string) // 每棵树取 tree_level=0 且排序最前的根节点 label
	if len(treeRoots) > 0 {
		var roots []ontology.OntClassTemplate
		if err = global.GVA_DB.Select("tree_root, label").Where("tree_level = 0 AND tree_root IN ?", treeRoots).
			Order("tree_root ASC, sort ASC, id ASC").Find(&roots).Error; err != nil {
			return
		}
		for _, r := range roots {
			if _, ok := rootMap[r.TreeRoot]; !ok {
				rootMap[r.TreeRoot] = r.Label
			}
		}
	}

	list = make([]ontRes.ClassTemplatePageItem, 0, len(rows))
	for _, r := range rows {
		item := ontRes.ClassTemplatePageItem{
			OntClassTemplate: r,
			ParentLabel:      parentMap[r.ParentId],
			TreeRootLabel:    rootMap[r.TreeRoot],
			HasChildren:      childMap[r.ID] > 0,
		}
		list = append(list, item)
	}
	return
}

// GetClassTemplateTreeRoots 分类树下拉：distinct tree_root + 该树根节点 label
func (s *ClassTemplateService) GetClassTemplateTreeRoots() (list []ontRes.ClassTreeRootItem, err error) {
	var roots []ontology.OntClassTemplate
	if err = global.GVA_DB.Select("tree_root, label").
		Where("tree_level = 0 AND tree_root <> ''").
		Order("tree_root ASC, sort ASC, id ASC").Find(&roots).Error; err != nil {
		return
	}
	seen := make(map[string]bool)
	list = make([]ontRes.ClassTreeRootItem, 0)
	for _, r := range roots {
		if seen[r.TreeRoot] {
			continue
		}
		seen[r.TreeRoot] = true
		list = append(list, ontRes.ClassTreeRootItem{TreeRoot: r.TreeRoot, Label: r.Label})
	}
	return
}

// GetClassTemplateInherited 继承视图（FR-2 核心）：祖先链一次 IN + 全链 refs 一次 IN，无 N+1
func (s *ClassTemplateService) GetClassTemplateInherited(templateCode string) (view ontRes.ClassTemplateInheritedView, err error) {
	db := global.GVA_DB
	var self ontology.OntClassTemplate
	if err = db.Where("template_code = ?", templateCode).First(&self).Error; err != nil {
		return view, errors.New("分类模板不存在: " + templateCode)
	}
	// 1) parent_path 解析祖先 id 链（根→自身）
	chain := make([]ontology.OntClassTemplate, 0, 4)
	parts := strings.Split(self.ParentPath, ",")
	ids := make([]uint, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" || p == "0" {
			continue
		}
		id, pErr := strconv.ParseUint(p, 10, 64)
		if pErr != nil {
			continue
		}
		ids = append(ids, uint(id))
	}
	if len(ids) > 0 {
		var ancestors []ontology.OntClassTemplate
		if err = db.Where("id IN ?", ids).Find(&ancestors).Error; err != nil {
			return
		}
		byID := make(map[uint]ontology.OntClassTemplate, len(ancestors))
		for _, a := range ancestors {
			byID[a.ID] = a
		}
		for _, id := range ids { // 保持根→自身顺序
			if a, ok := byID[id]; ok {
				chain = append(chain, a)
			}
		}
	}
	chain = append(chain, self)

	// 2) 全链 refs 一次 IN，按 (层级序, sort_order) 归并（同名覆盖）
	chainIDs := make([]uint, 0, len(chain))
	for _, n := range chain {
		chainIDs = append(chainIDs, n.ID)
	}
	var allRefs []ontology.OntClassTemplateRef
	if err = db.Where("class_template_id IN ?", chainIDs).Find(&allRefs).Error; err != nil {
		return
	}
	refsByNode := make(map[uint][]ontology.OntClassTemplateRef, len(chain))
	for _, r := range allRefs {
		refsByNode[r.ClassTemplateId] = append(refsByNode[r.ClassTemplateId], r)
	}
	for _, refs := range refsByNode { // 节点内按 sort_order 稳定排序
		for i := 1; i < len(refs); i++ {
			for j := i; j > 0 && refs[j].SortOrder < refs[j-1].SortOrder; j-- {
				refs[j], refs[j-1] = refs[j-1], refs[j]
			}
		}
	}

	indexByCode := make(map[string]int)
	properties := make([]ontRes.InheritedRefItem, 0)
	for nodeIdx, node := range chain {
		isSelf := nodeIdx == len(chain)-1
		for _, r := range refsByNode[node.ID] {
			item := ontRes.InheritedRefItem{
				PropertyTemplateCode: r.PropertyTemplateCode,
				RefType:              r.RefType,
				SortOrder:            r.SortOrder,
				Source:               "inherited",
			}
			if isSelf {
				if idx, ok := indexByCode[r.PropertyTemplateCode]; ok {
					item.Source = "overridden" // 本节点覆盖父级同名
					properties[idx] = item
					continue
				}
				item.Source = "node"
			} else if idx, ok := indexByCode[r.PropertyTemplateCode]; ok {
				// 祖先链上同名：近覆盖远，保持原位置
				properties[idx] = item
				continue
			}
			indexByCode[item.PropertyTemplateCode] = len(properties)
			properties = append(properties, item)
		}
	}

	// 3) 外观继承：inherit_appearance=1 → 自近及远取最近非空父 icon/color
	view = ontRes.ClassTemplateInheritedView{
		TemplateCode:     self.TemplateCode,
		Label:            self.Label,
		Properties:       properties,
		Icon:             self.Icon,
		Color:            self.Color,
		AppearanceSource: "node",
	}
	if self.InheritAppearance == 1 {
		for i := len(chain) - 2; i >= 0; i-- { // 自近及远（不含自身）
			if chain[i].Icon != "" || chain[i].Color != "" {
				view.Icon = chain[i].Icon
				view.Color = chain[i].Color
				view.AppearanceSource = "inherited"
				break
			}
		}
	}
	return view, nil
}

// PreviewClassificationCode 编码预览（不落库）
func (s *ClassTemplateService) PreviewClassificationCode(parentId uint, treeRoot string) (string, error) {
	return new(ClassificationRuleService).GenerateClassificationCode(global.GVA_DB, parentId, treeRoot)
}

// —— 供给 ——

// GetClassTemplateTreeForSupply 供给：分类模板树扁平列表（默认排除弃用，仅启用）
func (s *ClassTemplateService) GetClassTemplateTreeForSupply(treeRoot string) (list []ontology.OntClassTemplate, err error) {
	db := global.GVA_DB.Where("deprecated = 0 AND status = 0")
	if treeRoot != "" {
		db = db.Where("tree_root = ?", treeRoot)
	}
	err = db.Order("tree_root ASC, parent_path ASC, sort ASC, id ASC").Find(&list).Error
	return
}

// SuggestClassHierarchy 类层级建议（FR-9）：本期无镜像数据，返回父分类模板信息供参考；根节点返回空对象
func (s *ClassTemplateService) SuggestClassHierarchy(templateCode string) (ontRes.ClassHierarchySuggest, error) {
	var self ontology.OntClassTemplate
	if err := global.GVA_DB.Where("template_code = ?", templateCode).First(&self).Error; err != nil {
		return ontRes.ClassHierarchySuggest{}, errors.New("分类模板不存在: " + templateCode)
	}
	if self.ParentId == 0 {
		return ontRes.ClassHierarchySuggest{}, nil
	}
	var parent ontology.OntClassTemplate
	if err := global.GVA_DB.First(&parent, self.ParentId).Error; err != nil {
		return ontRes.ClassHierarchySuggest{}, nil
	}
	return ontRes.ClassHierarchySuggest{
		SuggestedParentTemplateCode: parent.TemplateCode,
		SuggestedParentLabel:        parent.Label,
		Note:                        "依据治理分类树的父子关系推荐（ont_class_hierarchies 镜像未启用）",
	}, nil
}
