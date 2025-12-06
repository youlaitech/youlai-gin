package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	authModel "youlai-gin/internal/auth/model"
	userRepo "youlai-gin/internal/user/repository"
	"youlai-gin/pkg/auth"
	"youlai-gin/pkg/errs"
)

// tokenManager 全局 TokenManager 实例
var tokenManager auth.TokenManager

// InitTokenManager 初始化 TokenManager（由 main 或 router 调用）
func InitTokenManager(tm auth.TokenManager) {
	tokenManager = tm
}

// Login 账号密码登录
func Login(req *authModel.LoginRequest) (*auth.AuthenticationToken, error) {
	// 1. 根据用户名查询用户
	user, err := userRepo.FindByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.BadRequest("用户名或密码错误")
		}
		return nil, errs.SystemError("查询用户失败")
	}

	// 2. 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errs.BadRequest("用户名或密码错误")
	}

	// 3. 检查用户状态
	if user.Status != 1 {
		return nil, errs.BadRequest("用户已被禁用")
	}

	// 4. 生成 Token
	userDetails := &auth.UserDetails{
		UserID:    int64(user.ID),
		Username:  user.Username,
		DeptID:    user.DeptID,
		DataScope: 0, // TODO: 从角色权限获取
		Roles:     []string{}, // TODO: 从用户角色关联表获取
	}

	token, err := tokenManager.GenerateToken(userDetails)
	if err != nil {
		return nil, errs.SystemError("生成令牌失败")
	}

	return token, nil
}

// Logout 退出登录
func Logout(token string) error {
	if token == "" {
		return nil
	}
	return tokenManager.InvalidateToken(token)
}

// RefreshToken 刷新令牌
func RefreshToken(refreshToken string) (*auth.AuthenticationToken, error) {
	if refreshToken == "" {
		return nil, errs.BadRequest("刷新令牌不能为空")
	}

	token, err := tokenManager.RefreshToken(refreshToken)
	if err != nil {
		return nil, errs.TokenInvalid()
	}

	return token, nil
}
