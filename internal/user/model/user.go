package model

// User 实体，对应 youlai_boot.sys_user 表（示例字段最小化）
// 实际项目可以根据 youlai-boot 的 SysUser 表补充手机号、性别、邮箱等字段
type User struct {
	ID       uint64 `gorm:"primaryKey" json:"id"` // 主键 ID
	Username string `json:"username"`             // 用户名（登录账号）
	Nickname string `json:"nickname"`             // 显示昵称
}

// TableName 显式指定表名
func (User) TableName() string {
	return "sys_user"
}
