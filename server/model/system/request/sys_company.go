package request

// SysCompanySearch 公司搜索（树接口支持名称过滤，命中行平铺返回）
type SysCompanySearch struct {
	Name string `json:"name" form:"name"` // 公司名称关键字
}
