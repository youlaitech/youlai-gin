package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mojocn/base64Captcha"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	authModel "youlai-gin/internal/auth/model"
	userRepo "youlai-gin/internal/system/user/repository"
	"youlai-gin/pkg/auth"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/redis"
)

// tokenManager 全局 TokenManager 实例
var tokenManager auth.TokenManager

// captchaStore 验证码存储
var captchaStore = base64Captcha.DefaultMemStore

// InitTokenManager 初始化 TokenManager（由 main 或 router 调用）
func InitTokenManager(tm auth.TokenManager) {
	tokenManager = tm
}

// GetCaptcha 获取验证码
func GetCaptcha() (*authModel.CaptchaVO, error) {
	// 现代化清爽风格验证码配置
	// 使用字符串验证码，简洁美观
	driver := base64Captcha.NewDriverString(
		40,                              // 高度 40px
		120,                             // 宽度 120px
		0,                               // 噪点数量 0（完全无干扰，最清爽）
		base64Captcha.OptionShowSlimeLine, // 只显示细线条（可选）
		4,                               // 验证码长度 4位
		"0123456789",                    // 只使用数字，简单易读
		nil,                             // 使用默认背景色（白色）
		nil,                             // 使用默认字体
	)
	
	// 生成验证码
	captcha := base64Captcha.NewCaptcha(driver, captchaStore)
	id, b64s, err := captcha.Generate()
	if err != nil {
		return nil, errs.SystemError("生成验证码失败")
	}
	
	// 获取验证码答案
	answer := captchaStore.Get(id, false)
	
	// 生成验证码 Key
	captchaKey := uuid.New().String()
	
	// 将验证码存储到 Redis（5分钟过期）
	redisKey := fmt.Sprintf("captcha:image:%s", captchaKey)
	ctx := context.Background()
	err = redis.Client.Set(ctx, redisKey, answer, 5*time.Minute).Err()
	if err != nil {
		// Redis 失败回退到内存存储
		captchaStore.Set(captchaKey, answer)
	}
	
	// 清理内存中的验证码ID（我们使用自己的key）
	captchaStore.Set(id, "")
	
	return &authModel.CaptchaVO{
		CaptchaKey:    captchaKey,
		CaptchaBase64: b64s,
	}, nil
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

	// 4. 获取用户角色
	roles, err := userRepo.GetUserRoles(int64(user.ID))
	if err != nil {
		return nil, errs.SystemError("查询用户角色失败")
	}

	// 5. 生成 Token
	userDetails := &auth.UserDetails{
		UserID:    int64(user.ID),
		Username:  user.Username,
		DeptID:    user.DeptID,
		DataScope: 0, // TODO: 从角色权限获取
		Roles:     roles,
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

// SendSmsLoginCode 发送登录短信验证码
func SendSmsLoginCode(mobile string) error {
	// 生成验证码
	// code := fmt.Sprintf("%04d", rand.Intn(10000))
	// TODO: 为了方便测试，验证码固定为 1234，实际开发中在配置了厂商短信服务后，可以使用上面的随机验证码
	code := "1234"

	// 缓存验证码至 Redis（5分钟过期）
	redisKey := fmt.Sprintf("captcha:sms:login:%s", mobile)
	ctx := context.Background()
	err := redis.Client.Set(ctx, redisKey, code, 5*time.Minute).Err()
	if err != nil {
		return errs.SystemError("发送短信验证码失败")
	}

	// TODO: 实际开发中对接短信服务商（阿里云、腾讯云等）
	// 示例：smsService.SendSMS(mobile, code)

	return nil
}

// LoginBySms 短信验证码登录
func LoginBySms(req *authModel.SmsLoginRequest) (*auth.AuthenticationToken, error) {
	// 1. 验证短信验证码
	redisKey := fmt.Sprintf("captcha:sms:login:%s", req.Mobile)
	ctx := context.Background()
	cachedCode, err := redis.Client.Get(ctx, redisKey).Result()
	if err != nil {
		return nil, errs.BadRequest("验证码已过期或不存在")
	}
	
	if cachedCode != req.Code {
		return nil, errs.BadRequest("验证码错误")
	}
	
	// 2. 根据手机号查询用户
	user, err := userRepo.GetUserByMobile(req.Mobile)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.BadRequest("用户不存在")
		}
		return nil, errs.SystemError("查询用户失败")
	}
	
	// 3. 检查用户状态
	if user.Status != 1 {
		return nil, errs.BadRequest("用户已被禁用")
	}
	
	// 4. 获取用户角色
	roles, err := userRepo.GetUserRoles(int64(user.ID))
	if err != nil {
		return nil, errs.SystemError("查询用户角色失败")
	}

	// 5. 验证成功后删除验证码
	redis.Client.Del(ctx, redisKey)

	// 6. 生成 Token
	userDetails := &auth.UserDetails{
		UserID:    int64(user.ID),
		Username:  user.Username,
		DeptID:    user.DeptID,
		DataScope: 0, // TODO: 从角色权限获取
		Roles:     roles,
	}
	
	token, err := tokenManager.GenerateToken(userDetails)
	if err != nil {
		return nil, errs.SystemError("生成令牌失败")
	}
	
	return token, nil
}

// LoginByWechat 微信授权登录(Web)
func LoginByWechat(code string) (*auth.AuthenticationToken, error) {
	// TODO: 实现微信网页授权登录
	// 1. 使用 code 调用微信接口获取 access_token 和 openid
	// 2. 根据 openid 查询或创建用户
	// 3. 生成 JWT Token
	return nil, errs.SystemError("微信登录功能待实现")
}

// LoginByWxMiniAppCode 微信小程序登录(Code)
func LoginByWxMiniAppCode(req *authModel.WxMiniAppCodeLoginRequest) (*auth.AuthenticationToken, error) {
	// TODO: 实现微信小程序 Code 登录
	// 1. 使用 code 调用微信接口获取 session_key 和 openid
	// 2. 根据 openid 查询或创建用户
	// 3. 生成 JWT Token
	return nil, errs.SystemError("微信小程序Code登录功能待实现")
}

// LoginByWxMiniAppPhone 微信小程序登录(手机号)
func LoginByWxMiniAppPhone(req *authModel.WxMiniAppPhoneLoginRequest) (*auth.AuthenticationToken, error) {
	// TODO: 实现微信小程序手机号登录
	// 1. 使用 code 获取 session_key
	// 2. 解密 encryptedData 获取手机号
	// 3. 根据手机号查询或创建用户
	// 4. 生成 JWT Token
	return nil, errs.SystemError("微信小程序手机号登录功能待实现")
}
