package system

import (
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"gorm.io/gorm"
)

type AuthorityDataScopeService struct{}

var AuthorityDataScopeServiceApp = new(AuthorityDataScopeService)

// 数据范围档位（与前端 dataScope.vue 常量、设计文档 01 篇 3.2 对齐）
const (
	DataScopeAll          = 1 // 全部数据
	DataScopeCustom       = 2 // 自定义数据
	DataScopeCompany      = 3 // 本公司数据
	DataScopeDeptChildren = 4 // 本部门及以下数据
	DataScopeDept         = 5 // 本部门数据
	DataScopeSelf         = 6 // 仅本人数据
)

// SetAuthorityDataScope 设置角色数据范围（全量覆盖明细）
func (s *AuthorityDataScopeService) SetAuthorityDataScope(authorityId uint, dataScope int, companyIds, departmentIds []uint) error {
	if dataScope < DataScopeAll || dataScope > DataScopeSelf {
		return errors.New("非法的数据范围值")
	}
	var authority system.SysAuthority
	if err := global.GVA_DB.Where("authority_id = ?", authorityId).First(&authority).Error; err != nil {
		return errors.New("角色不存在")
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&system.SysAuthority{}).Where("authority_id = ?", authorityId).
			Update("data_scope", dataScope).Error; err != nil {
			return err
		}
		// 明细全量覆盖：非自定义档位时清空
		if err := tx.Where("authority_id = ?", authorityId).Delete(&system.SysAuthorityDataScope{}).Error; err != nil {
			return err
		}
		if dataScope != DataScopeCustom {
			return nil
		}
		details := make([]system.SysAuthorityDataScope, 0, len(companyIds)+len(departmentIds))
		for _, id := range companyIds {
			details = append(details, system.SysAuthorityDataScope{
				AuthorityId: authorityId, ScopeType: system.ScopeTypeCompany, TargetId: id,
			})
		}
		for _, id := range departmentIds {
			details = append(details, system.SysAuthorityDataScope{
				AuthorityId: authorityId, ScopeType: system.ScopeTypeDepartment, TargetId: id,
			})
		}
		if len(details) > 0 {
			return tx.Create(&details).Error
		}
		return nil
	})
}

// GetAuthorityDataScope 查询角色数据范围及自定义明细
func (s *AuthorityDataScopeService) GetAuthorityDataScope(authorityId uint) (dataScope int, companyIds, departmentIds []uint, err error) {
	var authority system.SysAuthority
	if err = global.GVA_DB.Where("authority_id = ?", authorityId).First(&authority).Error; err != nil {
		return 0, nil, nil, errors.New("角色不存在")
	}
	dataScope = authority.DataScope
	if dataScope != DataScopeCustom {
		return dataScope, []uint{}, []uint{}, nil
	}
	var details []system.SysAuthorityDataScope
	if err = global.GVA_DB.Where("authority_id = ?", authorityId).Find(&details).Error; err != nil {
		return
	}
	companyIds = []uint{}
	departmentIds = []uint{}
	for _, d := range details {
		switch d.ScopeType {
		case system.ScopeTypeCompany:
			companyIds = append(companyIds, d.TargetId)
		case system.ScopeTypeDepartment:
			departmentIds = append(departmentIds, d.TargetId)
		}
	}
	return
}

// CopyAuthorityDataScope 拷贝角色时复制数据范围配置
func (s *AuthorityDataScopeService) CopyAuthorityDataScope(tx *gorm.DB, oldAuthorityId, newAuthorityId uint) error {
	var old system.SysAuthority
	if err := tx.Where("authority_id = ?", oldAuthorityId).First(&old).Error; err != nil {
		return nil // 源角色不存在时跳过
	}
	if err := tx.Model(&system.SysAuthority{}).Where("authority_id = ?", newAuthorityId).
		Update("data_scope", old.DataScope).Error; err != nil {
		return err
	}
	var details []system.SysAuthorityDataScope
	if err := tx.Where("authority_id = ?", oldAuthorityId).Find(&details).Error; err != nil {
		return err
	}
	if len(details) == 0 {
		return nil
	}
	newDetails := make([]system.SysAuthorityDataScope, 0, len(details))
	for _, d := range details {
		d.ID = 0
		d.AuthorityId = newAuthorityId
		newDetails = append(newDetails, d)
	}
	return tx.Create(&newDetails).Error
}

// DeleteByAuthorityId 删除角色的数据范围明细（删角色时调用）
func (s *AuthorityDataScopeService) DeleteByAuthorityId(tx *gorm.DB, authorityId uint) error {
	return tx.Where("authority_id = ?", authorityId).Delete(&system.SysAuthorityDataScope{}).Error
}
