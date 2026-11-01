package model

import (
	"time"

	baseModel "youlai-gin/pkg/model"
	"youlai-gin/pkg/types"
)

// User 用户实体
type User struct {
	ID       types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"`       // 主键
	Username string       `gorm:"column:username;not null" json:"username"` // 用户名
	Nickname string       `gorm:"column:nickname;not null" json:"nickname"` // 昵称
	Gender   int          `gorm:"column:gender;default:0" json:"gender"`    // 0-保密 1-男 2-女
	Password string       `gorm:"column:password;not null" json:"-"`        // 密码
	DeptID   types.BigInt `gorm:"column:dept_id" json:"deptId"`             // 部门ID
	Avatar   string       `gorm:"column:avatar" json:"avatar"`              // 头像
	Mobile   string       `gorm:"column:mobile" json:"mobile"`              // 手机号
	Status   int          `gorm:"column:status;default:1" json:"status"`    // 0-禁用 1-正常
	Email    string       `gorm:"column:email" json:"email"`                // 邮箱
	baseModel.BaseEntity
}

func (User) TableName() string {
	return "sys_user"
}

// UserRole 用户角色关联
type UserRole struct {
	UserID types.BigInt `gorm:"column:user_id;not null;primaryKey" json:"userId"`
	RoleID types.BigInt `gorm:"column:role_id;not null;primaryKey" json:"roleId"`
}

func (UserRole) TableName() string {
	return "sys_user_role"
}

// WechatMiniProvider 微信小程序的认证源标识，与配置一致
const WechatMiniProvider = "wechat-mini-main"

// UserIdentity 本地用户与外部账号的关联
type UserIdentity struct {
	ID                types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"` // 关联记录ID
	UserID            types.BigInt `gorm:"column:user_id;not null" json:"userId"`
	Provider          string       `gorm:"column:provider;not null;size:64" json:"provider"`                      // 具体认证源标识
	ProviderAccountID string       `gorm:"column:provider_account_id;not null;size:255" json:"providerAccountId"` // 来源方账号ID
	DisplayName       string       `gorm:"column:display_name;size:128" json:"displayName"`                       // 外部账号展示名
	AvatarURL         string       `gorm:"column:avatar_url;size:512" json:"avatarUrl"`                           // 外部账号头像
	LastLoginTime     *time.Time   `gorm:"column:last_login_time" json:"lastLoginTime"`                           // 最近登录时间
	baseModel.BaseEntity
}

func (UserIdentity) TableName() string {
	return "sys_user_identity"
}
