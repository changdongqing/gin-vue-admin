package request

// SetAuthorityDataScope 设置角色数据范围
type SetAuthorityDataScope struct {
	AuthorityId   uint   `json:"authorityId" binding:"required"`      // 角色ID
	DataScope     int    `json:"dataScope" binding:"required,min=1,max=6"` // 数据范围 1全部 2自定义 3本公司 4本部门及以下 5本部门 6仅本人
	CompanyIds    []uint `json:"companyIds"`                         // 自定义范围-公司ID列表
	DepartmentIds []uint `json:"departmentIds"`                      // 自定义范围-部门ID列表
}

// GetAuthorityDataScope 查询角色数据范围
type GetAuthorityDataScope struct {
	AuthorityId uint `json:"authorityId" form:"authorityId" binding:"required"` // 角色ID
}
