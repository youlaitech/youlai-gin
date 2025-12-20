package model

// UserPermissionsVO 用户权限信息（供权限校验、数据权限过滤使用）
type UserPermissionsVO struct {
	UserID    int64   `json:"userId"`
	Roles     []string `json:"roles"`
	Perms     []string `json:"perms"`
	DataScope int     `json:"dataScope"`
	DeptID    int64   `json:"deptId"`
	DeptIds   []int64 `json:"deptIds"`
}
