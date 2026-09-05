package system

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"gorm.io/gorm"
)

type CompanyService struct{}

var CompanyServiceApp = new(CompanyService)

// selfPath 计算节点自身的物化路径前缀（含自身，如 "0,1,3,"）
func companySelfPath(c *system.SysCompany) string {
	return c.ParentIds + strconv.Itoa(int(c.ID)) + ","
}

// resolveCompanyParent 解析并校验父节点，返回父记录（parentId=0 时返回 nil, nil）
func resolveCompanyParent(db *gorm.DB, parentId, selfId uint) (*system.SysCompany, error) {
	if parentId == 0 {
		return nil, nil
	}
	if parentId == selfId && selfId != 0 {
		return nil, errors.New("父级公司不能是自身")
	}
	var parent system.SysCompany
	if err := db.Where("id = ?", parentId).First(&parent).Error; err != nil {
		return nil, errors.New("父级公司不存在")
	}
	if selfId != 0 {
		// 目标父节点的路径包含自身路径，说明把节点挂到了自己子孙下
		selfPrefix := strconv.Itoa(int(selfId)) + ","
		if parent.ParentId != 0 && strings.Contains(","+parent.ParentIds, ","+selfPrefix) {
			return nil, errors.New("不能将公司挂靠到自身的子级下")
		}
	}
	return &parent, nil
}

// CreateCompany 新建公司
func (companyService *CompanyService) CreateCompany(c *system.SysCompany) error {
	var count int64
	global.GVA_DB.Model(&system.SysCompany{}).Where("code = ?", c.Code).Count(&count)
	if count > 0 {
		return errors.New("公司编码已存在")
	}
	parent, err := resolveCompanyParent(global.GVA_DB, c.ParentId, 0)
	if err != nil {
		return err
	}
	if parent == nil {
		c.ParentId = 0
		c.ParentIds = "0,"
		c.TreeLevel = 0
	} else {
		c.ParentIds = parent.ParentIds + strconv.Itoa(int(parent.ID)) + ","
		c.TreeLevel = parent.TreeLevel + 1
	}
	return global.GVA_DB.Create(c).Error
}

// UpdateCompany 更新公司（支持移动父级，同步重算子孙路径）
func (companyService *CompanyService) UpdateCompany(c *system.SysCompany) error {
	var old system.SysCompany
	if err := global.GVA_DB.Where("id = ?", c.ID).First(&old).Error; err != nil {
		return errors.New("公司不存在")
	}
	var count int64
	global.GVA_DB.Model(&system.SysCompany{}).Where("code = ? AND id <> ?", c.Code, c.ID).Count(&count)
	if count > 0 {
		return errors.New("公司编码已存在")
	}
	parent, err := resolveCompanyParent(global.GVA_DB, c.ParentId, c.ID)
	if err != nil {
		return err
	}
	var newParentIds string
	var newLevel int
	if parent == nil {
		c.ParentId = 0
		newParentIds = "0,"
		newLevel = 0
	} else {
		newParentIds = parent.ParentIds + strconv.Itoa(int(parent.ID)) + ","
		newLevel = parent.TreeLevel + 1
	}
	oldSelfPath := companySelfPath(&old)
	newSelfPath := newParentIds + strconv.Itoa(int(c.ID)) + ","
	levelDelta := newLevel - old.TreeLevel

	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		// 更新自身（Updates 结构体忽略零值，改为指定列）
		if err := tx.Model(&system.SysCompany{}).Where("id = ?", c.ID).Updates(map[string]interface{}{
			"name":       c.Name,
			"code":       c.Code,
			"parent_id":  c.ParentId,
			"parent_ids": newParentIds,
			"tree_level": newLevel,
			"sort":       c.Sort,
			"leader":     c.Leader,
			"phone":      c.Phone,
			"email":      c.Email,
			"address":    c.Address,
			"remarks":    c.Remarks,
			"status":     c.Status,
		}).Error; err != nil {
			return err
		}
		if oldSelfPath == newSelfPath {
			return nil
		}
		// 祖先变动：前缀替换子孙路径，层级平移
		if err := tx.Model(&system.SysCompany{}).
			Where("parent_ids LIKE ?", oldSelfPath+"%").
			Updates(map[string]interface{}{
				"parent_ids": gorm.Expr("REPLACE(parent_ids, ?, ?)", oldSelfPath, newSelfPath),
				"tree_level": gorm.Expr("tree_level + ?", levelDelta),
			}).Error; err != nil {
			return err
		}
		return nil
	})
}

// DeleteCompany 删除公司（存在子公司或挂靠部门时拒绝）
func (companyService *CompanyService) DeleteCompany(id uint) error {
	var company system.SysCompany
	if err := global.GVA_DB.Where("id = ?", id).First(&company).Error; err != nil {
		return errors.New("公司不存在")
	}
	var childCount int64
	global.GVA_DB.Model(&system.SysCompany{}).Where("parent_id = ?", id).Count(&childCount)
	if childCount > 0 {
		return errors.New("存在子公司，请先处理子公司后再删除")
	}
	var deptCount int64
	global.GVA_DB.Model(&system.SysDepartment{}).Where("company_id = ?", id).Count(&deptCount)
	if deptCount > 0 {
		return errors.New("该公司下存在部门，请先处理部门后再删除")
	}
	return global.GVA_DB.Delete(&system.SysCompany{}, "id = ?", id).Error
}

// GetCompany 根据ID获取公司
func (companyService *CompanyService) GetCompany(id uint) (c system.SysCompany, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&c).Error
	return
}

// GetCompanyTree 公司树：name 过滤时平铺返回命中行，否则返回整棵森林
func (companyService *CompanyService) GetCompanyTree(info systemReq.SysCompanySearch) (list []system.SysCompany, err error) {
	db := global.GVA_DB.Model(&system.SysCompany{})
	if info.Name != "" {
		db = db.Where("name LIKE ?", "%"+info.Name+"%").Order("parent_ids, sort, id")
		err = db.Find(&list).Error
		return
	}
	var all []system.SysCompany
	err = db.Order("parent_ids, sort, id").Find(&all).Error
	if err != nil {
		return
	}
	list = buildCompanyTree(all, 0)
	return
}

// buildCompanyTree 由平铺列表建树（parentId=0 的为根）
func buildCompanyTree(all []system.SysCompany, parentId uint) []system.SysCompany {
	children := make([]system.SysCompany, 0)
	for _, item := range all {
		if item.ParentId == parentId {
			node := item
			node.Children = buildCompanyTree(all, item.ID)
			children = append(children, node)
		}
	}
	return children
}

// GetCompanySubTreeIds 返回公司子树全部ID（含自身）
func (companyService *CompanyService) GetCompanySubTreeIds(id uint) ([]uint, error) {
	var company system.SysCompany
	if err := global.GVA_DB.Select("id, parent_ids").Where("id = ?", id).First(&company).Error; err != nil {
		return nil, fmt.Errorf("公司不存在: %w", err)
	}
	selfPath := companySelfPath(&company)
	var ids []uint
	err := global.GVA_DB.Model(&system.SysCompany{}).
		Where("id = ? OR parent_ids LIKE ?", id, selfPath+"%").
		Pluck("id", &ids).Error
	return ids, err
}
