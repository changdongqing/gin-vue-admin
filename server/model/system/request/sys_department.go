package request

// SysDepartmentSearch 部门树搜索（按公司过滤时返回该公司部门树，按名称过滤时平铺返回）
type SysDepartmentSearch struct {
	CompanyId uint   `json:"companyId" form:"companyId"` // 挂靠公司ID 0为全部
	Name      string `json:"name" form:"name"`           // 部门名称关键字
}
