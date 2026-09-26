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
	RoleCodeRoot  = "ROOT"  // 超级管理员
	RoleCodeGuest = "GUEST" // 游客（微信等第三方自助注册默认角色）
)

// 角色 ID（对应 sys_role 表固定记录；配合 RoleCodeGuest 使用）
const RoleGuestID int64 = 3

// Redis Key
const (
	RedisKeyRolePerms = "system:role:perms" // 角色权限缓存 Hash key
)
