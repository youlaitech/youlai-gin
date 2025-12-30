package model

import (
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/types"
)

// User 用户实体
type User struct {
	ID       types.BigInt `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string       `gorm:"column:username;not null;uniqueIndex:uk_username" json:"username"`
	Nickname string       `gorm:"column:nickname;not null" json:"nickname"`
	Gender   int          `gorm:"column:gender;default:0" json:"gender"` // 0-保密 1-男 2-女
	Password string       `gorm:"column:password;not null" json:"-"`
	DeptID   types.BigInt `gorm:"column:dept_id" json:"deptId"`
	Avatar   string       `gorm:"column:avatar" json:"avatar"`
	Mobile   string       `gorm:"column:mobile" json:"mobile"`
	Status   int          `gorm:"column:status;default:1" json:"status"` // 0-禁用 1-正常
	Email    string       `gorm:"column:email" json:"email"`
	Openid   string       `gorm:"column:openid" json:"openid"`
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
