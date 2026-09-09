package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	authModel "youlai-gin/internal/auth/model"
	"youlai-gin/internal/common/auth"
	"youlai-gin/internal/common/config"
	"youlai-gin/internal/common/logger"
	"youlai-gin/internal/common/redis"
	userModel "youlai-gin/internal/system/user/model"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/types"
)

// WxMaRepository 微信小程序登录所需数据访问（用户 + 第三方绑定，由 user.Repository 实现）
type WxMaRepository interface {
	UserRepository
	GetSocial(ctx context.Context, platform userModel.SocialPlatform, openID string) (*userModel.UserSocial, error)
	UpsertSocial(ctx context.Context, social *userModel.UserSocial) error
	CreateUserWithGuestRole(ctx context.Context, user *userModel.User) error
}

// wechatConfig 微信小程序配置
type wechatConfig struct {
	AppID     string
	AppSecret string
}

var wechatCfg wechatConfig

// wechatHTTPClient 微信 API 专用客户端：显式超时防止微信侧抖动拖死协程（默认 http.Get 无超时）
var wechatHTTPClient = &http.Client{Timeout: 10 * time.Second}

// InitWechatConfig 初始化微信配置（由 router 启动时调用）
func InitWechatConfig() {
	if config.Cfg == nil {
		logger.Log.Sugar().Errorw("配置未初始化，无法获取微信配置")
		return
	}

	wechatCfg = wechatConfig{
		AppID:     config.Cfg.Wechat.Miniapp.AppID,
		AppSecret: config.Cfg.Wechat.Miniapp.AppSecret,
	}

	logger.Log.Sugar().Infow("微信配置初始化完成", "appId", wechatCfg.AppID)
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
	ErrCode   int    `json:"errcode"`
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

// WxMaService 微信小程序认证业务逻辑层
type WxMaService struct {
	issuer   *TokenIssuer
	userRepo WxMaRepository
}

// NewWxMaService 创建 WxMaService 实例
func NewWxMaService(issuer *TokenIssuer, userRepo WxMaRepository) *WxMaService {
	return &WxMaService{issuer: issuer, userRepo: userRepo}
}

// SilentLogin 静默登录：已绑定手机号直接签发令牌，未绑定返回 openId 引导绑定
func (s *WxMaService) SilentLogin(ctx context.Context, code string) (*authModel.WxMaLoginResult, error) {
	session, err := getJsCodeSession(code)
	if err != nil {
		return nil, err
	}

	openID := session.OpenID
	if openID == "" {
		return nil, errs.BadRequest("微信登录失败：无法获取用户标识")
	}

	// 查找是否已绑定用户
	social, err := s.userRepo.GetSocial(ctx, userModel.PlatformWechatMini, openID)

	if err == nil {
		// 已绑定用户，直接登录
		token, err := s.generateTokenByUserID(ctx, int64(social.UserID))
		if err != nil {
			return nil, err
		}
		return &authModel.WxMaLoginResult{
			NeedBindMobile: false,
			AccessToken:    token.AccessToken,
			RefreshToken:   token.RefreshToken,
			ExpiresIn:      int64(token.ExpiresIn),
			TokenType:      token.TokenType,
		}, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Log.Sugar().Errorw("查询用户绑定失败", "error", err)
		return nil, errs.SystemError("查询用户绑定失败")
	}

	// 未绑定用户，返回需要绑定手机号
	logger.Log.Sugar().Infow("微信小程序静默登录：用户未绑定手机号", "openId", openID)
	return &authModel.WxMaLoginResult{
		NeedBindMobile: true,
		OpenID:         openID,
	}, nil
}

// PhoneLogin 手机号快捷登录：取微信会话与手机号，查询或创建用户后绑定并签发令牌
func (s *WxMaService) PhoneLogin(ctx context.Context, loginCode, phoneCode string) (*auth.AuthenticationToken, error) {
	session, err := getJsCodeSession(loginCode)
	if err != nil {
		return nil, err
	}

	mobile, err := getPhoneNumber(phoneCode)
	if err != nil {
		return nil, err
	}

	logger.Log.Sugar().Infow("微信小程序手机号快捷登录", "openId", session.OpenID, "mobile", mobile)

	user, err := s.findOrCreateUser(ctx, mobile)
	if err != nil {
		return nil, err
	}

	s.bindWechatOpenID(ctx, int64(user.ID), session.OpenID, session.UnionID, session.SessionKey)

	return s.issuer.Issue(ctx, user)
}

// BindMobile 绑定手机号：校验短信验证码，查询或创建用户后绑定并签发令牌
func (s *WxMaService) BindMobile(ctx context.Context, openID, mobile, smsCode string) (*auth.AuthenticationToken, error) {
	if err := validateSmsCode(ctx, mobile, smsCode); err != nil {
		return nil, err
	}

	user, err := s.findOrCreateUser(ctx, mobile)
	if err != nil {
		return nil, err
	}

	s.bindWechatOpenID(ctx, int64(user.ID), openID, "", "")

	logger.Log.Sugar().Infow("微信小程序绑定手机号成功", "mobile", mobile, "openId", openID)

	return s.issuer.Issue(ctx, user)
}

// findOrCreateUser 按手机号查用户，不存在则创建新用户（分配 GUEST 角色）
func (s *WxMaService) findOrCreateUser(ctx context.Context, mobile string) (*userModel.User, error) {
	user, err := s.userRepo.GetByMobile(ctx, mobile)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.SystemError("查询用户失败")
	}

	// 创建新用户并分配 GUEST 角色（事务由仓储层保证原子性）
	user = &userModel.User{
		Username: "wx_" + uuid.New().String()[:8],
		Nickname: "微信用户",
		Mobile:   mobile,
		Status:   1,
	}

	if err := s.userRepo.CreateUserWithGuestRole(ctx, user); err != nil {
		return nil, errs.BadRequest("创建用户失败：" + err.Error())
	}

	logger.Log.Sugar().Infow("微信小程序登录：创建新用户", "mobile", mobile, "userId", user.ID)
	return user, nil
}

// bindWechatOpenID 绑定微信 openid（已存在则更新绑定的用户与会话，否则新增）
func (s *WxMaService) bindWechatOpenID(ctx context.Context, userID int64, openID, unionID, sessionKey string) {
	social := &userModel.UserSocial{
		UserID:     types.BigInt(userID),
		Platform:   userModel.PlatformWechatMini,
		OpenID:     openID,
		UnionID:    unionID,
		SessionKey: sessionKey,
		Verified:   1,
	}

	if err := s.userRepo.UpsertSocial(ctx, social); err != nil {
		logger.Log.Sugar().Errorw("绑定微信 openid 失败", "openId", openID, "error", err)
	}
}

// generateTokenByUserID 根据用户ID生成令牌
func (s *WxMaService) generateTokenByUserID(ctx context.Context, userID int64) (*auth.AuthenticationToken, error) {
	user, err := s.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, errs.BadRequest("用户不存在")
	}
	return s.issuer.Issue(ctx, user)
}

// getJsCodeSession 获取微信会话信息
func getJsCodeSession(code string) (*WechatSessionResponse, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		wechatCfg.AppID, wechatCfg.AppSecret, code)

	resp, err := wechatHTTPClient.Get(url)
	if err != nil {
		logger.Log.Sugar().Errorw("获取微信会话信息失败", "code", code, "error", err)
		return nil, errs.BadRequest("微信登录失败，请稍后重试")
	}
	defer resp.Body.Close()

	var result WechatSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errs.BadRequest("解析微信响应失败")
	}

	if result.ErrCode != 0 {
		logger.Log.Sugar().Errorw("获取微信会话信息失败", "code", code, "errcode", result.ErrCode, "errmsg", result.ErrMsg)
		return nil, errs.BadRequest("微信登录失败，请稍后重试")
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

	resp, err := wechatHTTPClient.Get(url)
	if err != nil {
		logger.Log.Sugar().Errorw("获取微信手机号失败", "phoneCode", phoneCode, "error", err)
		return "", errs.BadRequest("获取手机号失败，请稍后重试")
	}
	defer resp.Body.Close()

	var result WechatPhoneResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", errs.BadRequest("解析微信响应失败")
	}

	if result.ErrCode != 0 {
		logger.Log.Sugar().Errorw("获取微信手机号失败", "phoneCode", phoneCode, "errcode", result.ErrCode, "errmsg", result.ErrMsg)
		return "", errs.BadRequest("获取手机号失败，请稍后重试")
	}

	return result.PhoneInfo.PhoneNumber, nil
}

// getAccessToken 获取微信 AccessToken（Redis 缓存，提前5分钟过期）
func getAccessToken() (string, error) {
	cacheKey := redis.WechatAccessTokenPrefix + wechatCfg.AppID

	cached, err := redis.Client.Get(context.Background(), cacheKey).Result()
	if err == nil {
		return cached, nil
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		wechatCfg.AppID, wechatCfg.AppSecret)

	resp, err := wechatHTTPClient.Get(url)
	if err != nil {
		return "", errs.BadRequest("获取微信AccessToken失败：" + err.Error())
	}
	defer resp.Body.Close()

	var result WechatTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", errs.BadRequest("解析微信响应失败")
	}

	if result.ErrCode != 0 {
		return "", errs.BadRequest("获取微信AccessToken失败：" + result.ErrMsg)
	}

	expiresIn := max(result.ExpiresIn-300, 60)
	redis.Client.Set(context.Background(), cacheKey, result.AccessToken, time.Duration(expiresIn)*time.Second)

	return result.AccessToken, nil
}

// validateSmsCode 验证短信验证码（成功后删除）
func validateSmsCode(ctx context.Context, mobile, smsCode string) error {
	// 与 AuthService.SendSmsLoginCode 共用同一 Key（历史上此处曾误用 "sms:login:" 前缀导致校验必失败）
	cacheKey := redis.CaptchaSmsPrefix + mobile
	cached, err := redis.Client.Get(ctx, cacheKey).Result()
	if err != nil {
		return errs.BadRequest("验证码已过期")
	}

	if cached != smsCode {
		return errs.BadRequest("验证码错误")
	}

	redis.Client.Del(ctx, cacheKey)
	return nil
}
