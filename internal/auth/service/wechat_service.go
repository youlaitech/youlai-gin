package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	authModel "youlai-gin/internal/auth/model"
	"youlai-gin/internal/system/user/domain"
	userRepo "youlai-gin/internal/system/user/repository"
	"youlai-gin/pkg/auth"
	"youlai-gin/pkg/config"
	"youlai-gin/pkg/database"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/redis"
)

// wechatConfig 微信小程序配置
type wechatConfig struct {
	AppID     string
	AppSecret string
}

var wechatCfg wechatConfig

// InitWechatConfig 初始化微信配置
func InitWechatConfig() {
	if config.Cfg == nil {
		slog.Error("配置未初始化，无法获取微信配置")
		return
	}
	
	wechatCfg = wechatConfig{
		AppID:     config.Cfg.Wechat.Miniapp.AppID,
		AppSecret: config.Cfg.Wechat.Miniapp.AppSecret,
	}
	
	slog.Info("微信配置初始化完成", "appId", wechatCfg.AppID)
}

// WechatSessionResponse 微信会话响应
type WechatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// WechatPhoneResponse 微信手机号响应
type WechatPhoneResponse struct {
	ErrCode   int `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	PhoneInfo struct {
		PhoneNumber string `json:"phoneNumber"`
	} `json:"phone_info"`
}

// WechatTokenResponse 微信AccessToken响应
type WechatTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// SilentLogin 静默登录
func SilentLogin(code string) (*authModel.WechatMiniappLoginResult, error) {
	session, err := getJsCodeSession(code)
	if err != nil {
		return nil, err
	}

	openID := session.OpenID
	if openID == "" {
		return nil, errs.BusinessError("微信登录失败：无法获取用户标识")
	}

	// 查找是否已绑定用户
	var social domain.UserSocial
	err = database.DB.Where("platform = ? AND openid = ?", domain.PlatformWechatMini, openID).First(&social).Error

	if err == nil {
		// 已绑定用户，直接登录
		token, err := generateTokenByUserID(social.UserID)
		if err != nil {
			return nil, err
		}
		return &authModel.WechatMiniappLoginResult{
			NeedBindMobile: false,
			AccessToken:    token.AccessToken,
			RefreshToken:   token.RefreshToken,
			ExpiresIn:      token.ExpiresIn,
			TokenType:      token.TokenType,
		}, nil
	}

	if err != gorm.ErrRecordNotFound {
		slog.Error("查询用户绑定失败", "error", err)
		return nil, errs.SystemError("查询用户绑定失败")
	}

	// 未绑定用户，返回需要绑定手机号
	slog.Info("微信小程序静默登录：用户未绑定手机号", "openId", openID)
	return &authModel.WechatMiniappLoginResult{
		NeedBindMobile: true,
		OpenID:         openID,
	}, nil
}

// PhoneLogin 手机号快捷登录
func PhoneLogin(loginCode, phoneCode string) (*authModel.AuthenticationToken, error) {
	// 获取微信会话信息
	session, err := getJsCodeSession(loginCode)
	if err != nil {
		return nil, err
	}

	// 获取手机号
	mobile, err := getPhoneNumber(phoneCode)
	if err != nil {
		return nil, err
	}

	slog.Info("微信小程序手机号快捷登录", "openId", session.OpenID, "mobile", mobile)

	// 查询或创建用户
	user, err := findOrCreateUser(mobile)
	if err != nil {
		return nil, err
	}

	// 绑定微信 openid
	bindWechatOpenID(user.ID, session.OpenID, session.UnionID, session.SessionKey)

	// 生成认证令牌
	return generateTokenByUser(user)
}

// BindMobile 绑定手机号
func BindMobile(openID, mobile, smsCode string) (*authModel.AuthenticationToken, error) {
	// 验证短信验证码
	if err := validateSmsCode(mobile, smsCode); err != nil {
		return nil, err
	}

	// 查询或创建用户
	user, err := findOrCreateUser(mobile)
	if err != nil {
		return nil, err
	}

	// 绑定微信 openid
	bindWechatOpenID(user.ID, openID, "", "")

	slog.Info("微信小程序绑定手机号成功", "mobile", mobile, "openId", openID)

	// 生成认证令牌
	return generateTokenByUser(user)
}

// getJsCodeSession 获取微信会话信息
func getJsCodeSession(code string) (*WechatSessionResponse, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		wechatCfg.AppID, wechatCfg.AppSecret, code)

	resp, err := http.Get(url)
	if err != nil {
		slog.Error("获取微信会话信息失败", "code", code, "error", err)
		return nil, errs.BusinessError("微信登录失败：" + err.Error())
	}
	defer resp.Body.Close()

	var result WechatSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errs.BusinessError("解析微信响应失败")
	}

	if result.ErrCode != 0 {
		slog.Error("获取微信会话信息失败", "code", code, "errcode", result.ErrCode, "errmsg", result.ErrMsg)
		return nil, errs.BusinessError("微信登录失败：" + result.ErrMsg)
	}

	return &result, nil
}

// getPhoneNumber 获取微信手机号
func getPhoneNumber(phoneCode string) (string, error) {
	accessToken, err := getAccessToken()
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s&code=%s", accessToken, phoneCode)

	resp, err := http.Get(url)
	if err != nil {
		slog.Error("获取微信手机号失败", "phoneCode", phoneCode, "error", err)
		return "", errs.BusinessError("获取手机号失败：" + err.Error())
	}
	defer resp.Body.Close()

	var result WechatPhoneResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", errs.BusinessError("解析微信响应失败")
	}

	if result.ErrCode != 0 {
		slog.Error("获取微信手机号失败", "phoneCode", phoneCode, "errcode", result.ErrCode, "errmsg", result.ErrMsg)
		return "", errs.BusinessError("获取手机号失败：" + result.ErrMsg)
	}

	return result.PhoneInfo.PhoneNumber, nil
}

// getAccessToken 获取微信 AccessToken
func getAccessToken() (string, error) {
	cacheKey := fmt.Sprintf("wechat:access_token:%s", wechatCfg.AppID)

	// 先从缓存获取
	cached, err := redis.Client().Get(context.Background(), cacheKey).Result()
	if err == nil {
		return cached, nil
	}

	// 请求新 token
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		wechatCfg.AppID, wechatCfg.AppSecret)

	resp, err := http.Get(url)
	if err != nil {
		return "", errs.BusinessError("获取微信AccessToken失败：" + err.Error())
	}
	defer resp.Body.Close()

	var result WechatTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", errs.BusinessError("解析微信响应失败")
	}

	if result.ErrCode != 0 {
		return "", errs.BusinessError("获取微信AccessToken失败：" + result.ErrMsg)
	}

	// 缓存 token（提前5分钟过期）
	expiresIn := max(result.ExpiresIn-300, 60)
	redis.Client().Set(context.Background(), cacheKey, result.AccessToken, time.Duration(expiresIn)*time.Second)

	return result.AccessToken, nil
}

// findOrCreateUser 查询或创建用户
func findOrCreateUser(mobile string) (*domain.User, error) {
	user, err := userRepo.FindByMobile(mobile)
	if err == nil {
		return user, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, errs.SystemError("查询用户失败")
	}

	// 创建新用户
	user = &domain.User{
		Username: "wx_" + uuid.New().String()[:8],
		Nickname: "微信用户",
		Mobile:   mobile,
		Status:   1,
	}

	tx := database.DB.Begin()
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, errs.BusinessError("创建用户失败：" + err.Error())
	}

	// 分配 GUEST 角色（角色ID=3）
	if err := tx.Exec("INSERT INTO sys_user_role (user_id, role_id) VALUES (?, ?)", user.ID, 3).Error; err != nil {
		tx.Rollback()
		return nil, errs.BusinessError("分配角色失败：" + err.Error())
	}

	tx.Commit()
	slog.Info("微信小程序登录：创建新用户", "mobile", mobile, "userId", user.ID)
	return user, nil
}

// bindWechatOpenID 绑定微信 openid
func bindWechatOpenID(userID int64, openID, unionID, sessionKey string) {
	var social domain.UserSocial
	err := database.DB.Where("platform = ? AND openid = ?", domain.PlatformWechatMini, openID).First(&social).Error

	if err == nil {
		// 更新绑定
		database.DB.Model(&social).Updates(map[string]interface{}{
			"user_id":     userID,
			"unionid":     unionID,
			"session_key": sessionKey,
		})
		return
	}

	if err == gorm.ErrRecordNotFound {
		// 新增绑定
		social = domain.UserSocial{
			UserID:     userID,
			Platform:   domain.PlatformWechatMini,
			OpenID:     openID,
			UnionID:    unionID,
			SessionKey: sessionKey,
			Verified:   1,
		}
		database.DB.Create(&social)
	}
}

// validateSmsCode 验证短信验证码
func validateSmsCode(mobile, smsCode string) error {
	cacheKey := fmt.Sprintf("sms:login:%s", mobile)
	cached, err := redis.Client().Get(context.Background(), cacheKey).Result()
	if err != nil {
		return errs.BusinessError("验证码已过期")
	}

	if cached != smsCode {
		return errs.BusinessError("验证码错误")
	}

	// 验证成功后删除验证码
	redis.Client().Del(context.Background(), cacheKey)
	return nil
}

// generateTokenByUserID 根据用户ID生成Token
func generateTokenByUserID(userID int64) (*authModel.AuthenticationToken, error) {
	user, err := userRepo.FindByID(userID)
	if err != nil {
		return nil, errs.BusinessError("用户不存在")
	}
	return generateTokenByUser(user)
}

// generateTokenByUser 根据用户生成Token
func generateTokenByUser(user *domain.User) (*authModel.AuthenticationToken, error) {
	token, err := tokenManager.GenerateToken(&auth.UserAuthInfo{
		UserID: user.ID,
	})
	if err != nil {
		return nil, errs.SystemError("生成令牌失败")
	}

	return &authModel.AuthenticationToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresIn,
		TokenType:    "Bearer",
	}, nil
}
