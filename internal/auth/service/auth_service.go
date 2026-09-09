package service

import (
	"context"
	"errors"
	"image/color"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mojocn/base64Captcha"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	authModel "youlai-gin/internal/auth/model"
	"youlai-gin/internal/common/auth"
	"youlai-gin/internal/common/redis"
	userModel "youlai-gin/internal/system/user/model"
	"youlai-gin/pkg/errs"
)

// UserRepository 用户数据访问子集（认证场景所需查询，由 user.Repository 实现）
type UserRepository interface {
	Get(ctx context.Context, id int64) (*userModel.User, error)
	GetByUsername(ctx context.Context, username string) (*userModel.User, error)
	GetByMobile(ctx context.Context, mobile string) (*userModel.User, error)
	RoleCodes(ctx context.Context, userID int64) ([]string, error)
}

// captchaStore 图形验证码内存回退存储
var captchaStore = base64Captcha.DefaultMemStore

// AuthService 账号密码 / 短信验证码登录业务逻辑层
type AuthService struct {
	tm       auth.TokenManager // 令牌生命周期操作（登出/刷新）；签发统一走 TokenIssuer
	issuer   *TokenIssuer
	userRepo UserRepository
}

// NewAuthService 创建 AuthService 实例
func NewAuthService(tm auth.TokenManager, issuer *TokenIssuer, userRepo UserRepository) *AuthService {
	return &AuthService{tm: tm, issuer: issuer, userRepo: userRepo}
}

// GetCaptcha 获取图形验证码（存 Redis，失败回退内存存储）
func (s *AuthService) GetCaptcha(ctx context.Context) (*authModel.CaptchaVO, error) {
	// 清新亮色验证码配置（浅色背景 + 无干扰线/噪点）
	bgColor := &color.RGBA{R: 240, G: 248, B: 255, A: 255}
	driver := base64Captcha.NewDriverString(
		44,         // 高度 44px
		140,        // 宽度 140px
		0,          // 噪点数量 0（清爽）
		0,          // 无干扰线
		4,          // 验证码长度 4位
		"23456789", // 只使用清晰数字
		bgColor,    // 浅色背景
		nil,        // 默认字体
	)

	captcha := base64Captcha.NewCaptcha(driver, captchaStore)
	id, b64s, err := captcha.Generate()
	if err != nil {
		return nil, errs.SystemError("生成验证码失败")
	}

	answer := captchaStore.Get(id, false)
	captchaID := uuid.New().String()

	// 验证码存 Redis（5分钟过期），失败回退内存存储
	redisKey := redis.CaptchaImagePrefix + captchaID
	if err := redis.Client.Set(ctx, redisKey, answer, 5*time.Minute).Err(); err != nil {
		captchaStore.Set(captchaID, answer)
	}
	captchaStore.Set(id, "")

	return &authModel.CaptchaVO{
		CaptchaID:     captchaID,
		CaptchaBase64: b64s,
		CaptchaCode:   answer,
	}, nil
}

// Login 账号密码登录
func (s *AuthService) Login(ctx context.Context, req *authModel.LoginRequest) (*auth.AuthenticationToken, int64, error) {
	if err := validateImageCaptcha(ctx, req.CaptchaID, req.CaptchaCode); err != nil {
		return nil, 0, err
	}

	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errs.UserPasswordError()
		}
		return nil, 0, errs.SystemError("查询用户失败")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, 0, errs.UserPasswordError()
	}

	if user.Status != 1 {
		return nil, 0, errs.BadRequest("用户已被禁用")
	}

	token, err := s.issuer.Issue(ctx, user)
	if err != nil {
		return nil, 0, err
	}

	return token, int64(user.ID), nil
}

// LoginBySms 短信验证码登录
func (s *AuthService) LoginBySms(ctx context.Context, req *authModel.SmsLoginRequest) (*auth.AuthenticationToken, int64, error) {
	redisKey := redis.CaptchaSmsPrefix + req.Mobile
	cachedCode, err := redis.Client.Get(ctx, redisKey).Result()
	if err != nil {
		return nil, 0, errs.BadRequest("验证码已过期或不存在")
	}

	if cachedCode != req.Code {
		return nil, 0, errs.BadRequest("验证码错误")
	}

	user, err := s.userRepo.GetByMobile(ctx, req.Mobile)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errs.BadRequest("用户不存在")
		}
		return nil, 0, errs.SystemError("查询用户失败")
	}

	if user.Status != 1 {
		return nil, 0, errs.BadRequest("用户已被禁用")
	}

	redis.Client.Del(ctx, redisKey)

	token, err := s.issuer.Issue(ctx, user)
	if err != nil {
		return nil, 0, err
	}

	return token, int64(user.ID), nil
}

// SendSmsLoginCode 发送登录短信验证码
func (s *AuthService) SendSmsLoginCode(ctx context.Context, mobile string) error {
	// 生成验证码（开发环境固定值，生产环境接入短信服务后改为随机码）
	code := "1234"

	// 缓存验证码至 Redis（5分钟过期）
	redisKey := redis.CaptchaSmsPrefix + mobile
	if err := redis.Client.Set(ctx, redisKey, code, 5*time.Minute).Err(); err != nil {
		return errs.SystemError("发送短信验证码失败")
	}

	// TODO: 接入短信服务商并发送验证码
	// smsService.SendSMS(mobile, code)

	return nil
}

// Logout 退出登录（令牌为空时静默成功）
func (s *AuthService) Logout(token string) error {
	if token == "" {
		return nil
	}
	return s.tm.InvalidateToken(token)
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(refreshToken string) (*auth.AuthenticationToken, error) {
	if refreshToken == "" {
		return nil, errs.BadRequest("刷新令牌不能为空")
	}

	token, err := s.tm.RefreshToken(refreshToken)
	if err != nil {
		return nil, errs.RefreshTokenInvalid()
	}

	return token, nil
}

// validateImageCaptcha 图形验证码校验（Redis 优先，回退内存存储）
func validateImageCaptcha(ctx context.Context, captchaID, captchaCode string) error {
	if captchaID == "" || captchaCode == "" {
		return errs.BadRequest("验证码不能为空")
	}

	redisKey := redis.CaptchaImagePrefix + captchaID
	answer, err := redis.Client.Get(ctx, redisKey).Result()
	if err == nil {
		redis.Client.Del(ctx, redisKey)
		if strings.EqualFold(answer, captchaCode) {
			return nil
		}
		return errs.BadRequest("验证码错误")
	}

	answer = captchaStore.Get(captchaID, true)
	if answer == "" {
		return errs.BadRequest("验证码已过期")
	}
	if !strings.EqualFold(answer, captchaCode) {
		return errs.BadRequest("验证码错误")
	}
	return nil
}
