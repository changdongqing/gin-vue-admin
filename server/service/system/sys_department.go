package system

import (
	"errors"
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"gorm.io/gorm"
)

type DepartmentService struct{}

var DepartmentServiceApp = new(DepartmentService)

// departmentSelfPath 计算部门节点自身的物化路径前缀（含自身，如 "0,1,3,"）
func departmentSelfPath(d *system.SysDepartment) string {
	return d.ParentIds + strconv.Itoa(int(d.ID)) + ","
}

// resolveDepartmentParent 解析并校验父节点（父必须与自身同属一家公司；parentId=0 返回 nil, nil）
func resolveDepartmentParent(db *gorm.DB, parentId, selfId, companyId uint) (*system.SysDepartment, error) {
	if parentId == 0 {
		return nil, nil
	}
	if parentId == selfId && selfId != 0 {
		return nil, errors.New("父级部门不能是自身")
	}
	var parent system.SysDepartment
	if err := db.Where("id = ?", parentId).First(&parent).Error; err != nil {
		return nil, errors.New("父级部门不存在")
	}
	if selfId != 0 {
		if parent.CompanyId != companyId {
			return nil, errors.New("父级部门与当前部门不属于同一家公司")
		}
		if parent.ParentId != 0 && len(parent.ParentIds) > 0 && containsPathToken(parent.ParentIds, selfId) {
			return nil, errors.New("不能将部门挂靠到自身的子级下")
		}
	}
	return &parent, nil
}

// containsPathToken 判断物化路径 "0,1,3," 中是否包含指定ID节点
func containsPathToken(parentIds string, id uint) bool {
	target := strconv.Itoa(int(id))
	for _, seg := range strings.Split(parentIds, ",") {
		if seg == target {
			return true
		}
	}
	return false
}

// CreateDepartment 新建部门
func (departmentService *DepartmentService) CreateDepartment(d *system.SysDepartment) error {
	var company system.SysCompany
	if err := global.GVA_DB.Where("id = ?", d.CompanyId).First(&company).Error; err != nil {
		return errors.New("所属公司不存在")
	}
	var count int64
	global.GVA_DB.Model(&system.SysDepartment{}).Where("code = ?", d.Code).Count(&count)
	if count > 0 {
		return errors.New("部门编码已存在")
	}
	parent, err := resolveDepartmentParent(global.GVA_DB, d.ParentId, 0, d.CompanyId)
	if err != nil {
		return err
	}
	if parent == nil {
		d.ParentId = 0
		d.ParentIds = "0,"
		d.TreeLevel = 0
	} else {
		d.ParentIds = parent.ParentIds + strconv.Itoa(int(parent.ID)) + ","
		d.TreeLevel = parent.TreeLevel + 1
	}
	return global.GVA_DB.Create(d).Error
}

// UpdateDepartment 更新部门（支持移动父级，同步重算子孙路径；公司变更时子孙一并切换）
func (departmentService *DepartmentService) UpdateDepartment(d *system.SysDepartment) error {
	var old system.SysDepartment
	if err := global.GVA_DB.Where("id = ?", d.ID).First(&old).Error; err != nil {
		return errors.New("部门不存在")
	}
	var count int64
	global.GVA_DB.Model(&system.SysDepartment{}).Where("code = ? AND id <> ?", d.Code, d.ID).Count(&count)
	if count > 0 {
		return errors.New("部门编码已存在")
	}
	var company system.SysCompany
	if err := global.GVA_DB.Where("id = ?", d.CompanyId).First(&company).Error; err != nil {
		return errors.New("所属公司不存在")
	}
	parent, err := resolveDepartmentParent(global.GVA_DB, d.ParentId, d.ID, d.CompanyId)
	if err != nil {
		return err
	}
	var newParentIds string
	var newLevel int
	if parent == nil {
		d.ParentId = 0
		newParentIds = "0,"
		newLevel = 0
	} else {
		newParentIds = parent.ParentIds + strconv.Itoa(int(parent.ID)) + ","
		newLevel = parent.TreeLevel + 1
	}
	oldSelfPath := departmentSelfPath(&old)
	newSelfPath := newParentIds + strconv.Itoa(int(d.ID)) + ","
	levelDelta := newLevel - old.TreeLevel

	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&system.SysDepartment{}).Where("id = ?", d.ID).Updates(map[string]interface{}{
			"company_id": d.CompanyId,
			"name":       d.Name,
			"code":       d.Code,
			"parent_id":  d.ParentId,
			"parent_ids": newParentIds,
			"tree_level": newLevel,
			"sort":       d.Sort,
			"leader":     d.Leader,
			"phone":      d.Phone,
			"email":      d.Email,
			"address":    d.Address,
			"remarks":    d.Remarks,
			"status":     d.Status,
		}).Error; err != nil {
			return err
		}
		if oldSelfPath == newSelfPath && old.CompanyId == d.CompanyId {
			return nil
		}
		updates := map[string]interface{}{
			"parent_ids": gorm.Expr("REPLACE(parent_ids, ?, ?)", oldSelfPath, newSelfPath),
			"tree_level": gorm.Expr("tree_level + ?", levelDelta),
		}
		// 公司整体切换时，子孙部门一并归属新公司
		if old.CompanyId != d.CompanyId {
			updates["company_id"] = d.CompanyId
		}
		return tx.Model(&system.SysDepartment{}).
			Where("parent_ids LIKE ?", oldSelfPath+"%").
			Updates(updates).Error
	})
}

// DeleteDepartment 删除部门（存在子部门或用户时拒绝）
func (departmentService *DepartmentService) DeleteDepartment(id uint) error {
	var department system.SysDepartment
	if err := global.GVA_DB.Where("id = ?", id).First(&department).Error; err != nil {
		return errors.New("部门不存在")
	}
	var childCount int64
	global.GVA_DB.Model(&system.SysDepartment{}).Where("parent_id = ?", id).Count(&childCount)
	if childCount > 0 {
		return errors.New("存在子部门，请先处理子部门后再删除")
	}
	var userCount int64
	global.GVA_DB.Model(&system.SysUser{}).Where("department_id = ?", id).Count(&userCount)
	if userCount > 0 {
		return errors.New("该部门下存在用户，请先转移用户后再删除")
	}
	return global.GVA_DB.Delete(&system.SysDepartment{}, "id = ?", id).Error
}

// GetDepartment 根据ID获取部门
func (departmentService *DepartmentService) GetDepartment(id uint) (d system.SysDepartment, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&d).Error
	return
}

// GetDepartmentTree 部门树
// companyId > 0：仅返回该公司部门构成的树；name 非空：平铺返回命中行
func (departmentService *DepartmentService) GetDepartmentTree(info systemReq.SysDepartmentSearch) (list []system.SysDepartment, err error) {
	db := global.GVA_DB.Model(&system.SysDepartment{})
	if info.CompanyId > 0 {
		db = db.Where("company_id = ?", info.CompanyId)
	}
	if info.Name != "" {
		db = db.Where("name LIKE ?", "%"+info.Name+"%")
		err = db.Order("parent_ids, sort, id").Find(&list).Error
		return
	}
	var all []system.SysDepartment
	err = db.Order("parent_ids, sort, id").Find(&all).Error
	if err != nil {
		return
	}
	// companyId 过滤时根为该公司一级部门，否则为全局一级部门
	list = buildDepartmentTree(all, 0)
	return
}

func buildDepartmentTree(all []system.SysDepartment, parentId uint) []system.SysDepartment {
	children := make([]system.SysDepartment, 0)
	for _, item := range all {
		if item.ParentId == parentId {
			node := item
			node.Children = buildDepartmentTree(all, item.ID)
			children = append(children, node)
		}
	}
	return children
}

// GetDepartmentSubTreeIds 返回部门子树全部ID（含自身）
func (departmentService *DepartmentService) GetDepartmentSubTreeIds(id uint) ([]uint, error) {
	var department system.SysDepartment
	if err := global.GVA_DB.Select("id, parent_ids").Where("id = ?", id).First(&department).Error; err != nil {
		return nil, err
	}
	selfPath := departmentSelfPath(&department)
	var ids []uint
	err := global.GVA_DB.Model(&system.SysDepartment{}).
		Where("id = ? OR parent_ids LIKE ?", id, selfPath+"%").
		Pluck("id", &ids).Error
	return ids, err
}

// GetDepartmentIdsByCompany 返回公司全部部门ID
func (departmentService *DepartmentService) GetDepartmentIdsByCompany(companyId uint) ([]uint, error) {
	var ids []uint
	err := global.GVA_DB.Model(&system.SysDepartment{}).Where("company_id = ?", companyId).Pluck("id", &ids).Error
	return ids, err
}
