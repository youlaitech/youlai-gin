package service

import (
	"context"

	"youlai-gin/internal/common/auth"
	permService "youlai-gin/internal/common/permission/service"
	userModel "youlai-gin/internal/system/user/model"
	"youlai-gin/pkg/errs"
)

// TokenIssuer 令牌签发组件：统一「查询角色 → 计算数据权限 → 签发令牌」流程。
// 账号密码、短信、扫码、微信小程序四种登录方式共用，避免各 service 重复实现。
type TokenIssuer struct {
	tm       auth.TokenManager
	userRepo UserRepository
}

// NewTokenIssuer 创建 TokenIssuer 实例
func NewTokenIssuer(tm auth.TokenManager, userRepo UserRepository) *TokenIssuer {
	return &TokenIssuer{tm: tm, userRepo: userRepo}
}

// Issue 为用户签发访问令牌（载荷含角色与数据权限范围）
func (i *TokenIssuer) Issue(ctx context.Context, user *userModel.User) (*auth.AuthenticationToken, error) {
	roles, err := i.userRepo.RoleCodes(ctx, int64(user.ID))
	if err != nil {
		return nil, errs.SystemError("查询用户角色失败")
	}

	dataScopes, err := permService.GetUserDataScopes(int64(user.ID), roles, int64(user.DeptID))
	if err != nil {
		return nil, err
	}

	token, err := i.tm.GenerateToken(&auth.UserDetails{
		UserID:     int64(user.ID),
		Username:   user.Username,
		DeptID:     user.DeptID,
		Avatar:     user.Avatar,
		DataScopes: dataScopes,
		Roles:      roles,
	})
	if err != nil {
		return nil, errs.SystemError("生成令牌失败")
	}

	return token, nil
}
