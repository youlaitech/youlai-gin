package domain

import (
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/types"
)

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
	ID         types.BigInt   `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     types.BigInt   `gorm:"column:user_id;not null" json:"userId"`
	Platform   SocialPlatform `gorm:"column:platform;not null;size:20" json:"platform"`
	OpenID     string         `gorm:"column:openid;not null;size:64" json:"openid"`
	UnionID    string         `gorm:"column:unionid;size:64" json:"unionid"`
	Nickname   string         `gorm:"column:nickname;size:64" json:"nickname"`
	Avatar     string         `gorm:"column:avatar;size:255" json:"avatar"`
	SessionKey string         `gorm:"column:session_key;size:128" json:"sessionKey"`
	Verified   int            `gorm:"column:verified;default:1" json:"verified"` // 1-已验证 0-未验证
	common.BaseEntity
}

func (UserSocial) TableName() string {
	return "sys_user_social"
}
