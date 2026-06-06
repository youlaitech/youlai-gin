package model

import (
	common "youlai-gin/pkg/model"
	"youlai-gin/pkg/types"
)

// User 用户实体
type User struct {
	ID       types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"` // 主键
	Username string       `gorm:"column:username;not null" json:"username"` // 用户名
	Nickname string       `gorm:"column:nickname;not null" json:"nickname"` // 昵称
	Gender   int          `gorm:"column:gender;default:0" json:"gender"` // 0-保密 1-男 2-女
	Password string       `gorm:"column:password;not null" json:"-"` // 密码
	DeptID   types.BigInt `gorm:"column:dept_id" json:"deptId"` // 部门ID
	Avatar   string       `gorm:"column:avatar" json:"avatar"` // 头像
	Mobile   string       `gorm:"column:mobile" json:"mobile"` // 手机号
	Status   int          `gorm:"column:status;default:1" json:"status"` // 0-禁用 1-正常
	Email    string       `gorm:"column:email" json:"email"` // 邮箱
	common.BaseEntity
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

// SocialPlatform 社交平台类型
type SocialPlatform string

const (
	PlatformWechatMini SocialPlatform = "WECHAT_MINI"
	PlatformWechatMP   SocialPlatform = "WECHAT_MP"
	PlatformAlipay     SocialPlatform = "ALIPAY"
	PlatformQQ         SocialPlatform = "QQ"
	PlatformApple      SocialPlatform = "APPLE"
)

// UserSocial 用户第三方账号绑定
type UserSocial struct {
	ID         types.BigInt   `gorm:"primaryKey;autoIncrement" json:"id"` // 主键
	UserID     types.BigInt   `gorm:"column:user_id;not null" json:"userId"`
	Platform   SocialPlatform `gorm:"column:platform;not null;size:20" json:"platform"`
	OpenID     string         `gorm:"column:openid;not null;size:64" json:"openid"`
	UnionID    string         `gorm:"column:unionid;size:64" json:"unionid"`
	Nickname   string         `gorm:"column:nickname;size:64" json:"nickname"` // 昵称
	Avatar     string         `gorm:"column:avatar;size:255" json:"avatar"` // 头像
	SessionKey string         `gorm:"column:session_key;size:128" json:"sessionKey"`
	Verified   int            `gorm:"column:verified;default:1" json:"verified"` // 1-已验证 0-未验证
	common.BaseEntity
}

func (UserSocial) TableName() string {
	return "sys_user_social"
}
