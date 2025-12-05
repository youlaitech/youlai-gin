package model

// User 实体，对应 youlai_boot.sys_user 表（示例字段最小化）
// 实际项目可以根据 youlai-boot 的 SysUser 表补充手机号、性别、邮箱等字段
type User struct {
	ID       uint64 `gorm:"primaryKey" json:"id"` // 主键 ID
	Username string `json:"username" validate:"required,min=3,max=20"` // 用户名（登录账号）
	Nickname string `json:"nickname" validate:"required,min=2,max=30"` // 显示昵称
}

// TableName 显式指定表名
func (User) TableName() string {
	return "sys_user"
}

// CreateUserRequest 创建用户请求（带密码）
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Nickname string `json:"nickname" validate:"required,min=2,max=30"`
	Password string `json:"password" validate:"required,min=6,max=20"`
}
