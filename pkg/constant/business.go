package constant

// 业务常量定义

const (
	// 默认密码
	DefaultPassword = "123456"

	// 导出上限
	ExportMaxLimit = 10000
)

// 角色编码
const (
	RoleCodeRoot = "ROOT" // 超级管理员
)

// Redis Key
const (
	RedisKeyRolePerms = "system:role:perms" // 角色权限缓存 Hash key
)
