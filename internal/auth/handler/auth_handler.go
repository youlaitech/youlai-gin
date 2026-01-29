package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"youlai-gin/internal/auth/model"
	"youlai-gin/internal/auth/service"
	pkgAuth "youlai-gin/pkg/auth"
	"youlai-gin/pkg/response"
	"youlai-gin/pkg/validator"
)

// RegisterAuthRoutes 注册认证相关 HTTP 路由
func RegisterAuthRoutes(r *gin.RouterGroup) {
	r.GET("/auth/captcha", GetCaptcha)
	r.POST("/auth/login", Login)
	r.POST("/auth/login/sms", LoginBySms)
	r.POST("/auth/sms/code", SendSmsCode)
	r.POST("/auth/login/wechat", LoginByWechat)
	r.POST("/auth/wx/miniapp/code-login", LoginByWxMiniAppCode)
	r.POST("/auth/wx/miniapp/phone-login", LoginByWxMiniAppPhone)
	r.DELETE("/auth/logout", Logout)
	r.POST("/auth/refresh-token", RefreshToken)
}

// GetCaptcha 获取验证码
// @Summary 获取验证码
// @Description 获取图形验证码
// @Tags 01.认证接口
// @Produce json
// @Success 200 {object} map[string]interface{} "code/msg/data，data 为 CaptchaVO"
// @Router /api/v1/auth/captcha [get]
func GetCaptcha(c *gin.Context) {
	captcha, err := service.GetCaptcha()
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, captcha)
}

// Login 账号密码登录
// @Summary 账号密码登录
// @Description 用户名密码登录，返回访问令牌和刷新令牌
// @Tags 01.认证接口
// @Accept application/json
// @Produce json
// @Param body body model.LoginRequest true "登录信息"
// @Success 200 {object} map[string]interface{} "code/msg/data，data 为 AuthenticationToken"
// @Router /api/v1/auth/login [post]
func Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if req.Username == "" || req.Password == "" {
		response.BadRequest(c, "用户名和密码不能为空")
		return
	}

	token, err := service.Login(&req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, token)
}

// Logout 退出登录
// @Summary 退出登录
// @Description 使当前访问令牌失效
// @Tags 01.认证接口
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "code/msg"
// @Router /api/v1/auth/logout [delete]
func Logout(c *gin.Context) {
	// 从 Header 中获取 Token
	authHeader := c.GetHeader(pkgAuth.AuthorizationHeader)
	token := strings.TrimPrefix(authHeader, pkgAuth.BearerPrefix)

	if err := service.Logout(token); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "退出成功")
}

// RefreshToken 刷新令牌
// @Summary 刷新令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 01.认证接口
// @Accept json
// @Produce json
// @Param body body map[string]string true "刷新令牌信息 {\"refreshToken\":\"刷新令牌\"}"
// @Success 200 {object} map[string]interface{} "code/msg/data，data 为 AuthenticationToken"
// @Router /api/v1/auth/refresh-token [post]
func RefreshToken(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	refreshToken := req["refreshToken"]
	if refreshToken == "" {
		response.BadRequest(c, "刷新令牌不能为空")
		return
	}

	token, err := service.RefreshToken(refreshToken)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, token)
}

// SendSmsCode 发送登录短信验证码
// @Summary 发送登录短信验证码
// @Description 发送短信验证码到指定手机号
// @Tags 01.认证接口
// @Accept json
// @Produce json
// @Param body body map[string]string true "手机号信息 {\"mobile\":\"手机号\"}"
// @Success 200 {object} map[string]interface{} "code/msg"
// @Router /api/v1/auth/sms/code [post]
func SendSmsCode(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	mobile := req["mobile"]
	if mobile == "" {
		response.BadRequest(c, "手机号不能为空")
		return
	}

	err := service.SendSmsLoginCode(mobile)
	if err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "验证码已发送")
}

// LoginBySms 短信验证码登录
// @Summary 短信验证码登录
// @Description 使用手机号和短信验证码登录
// @Tags 01.认证接口
// @Accept json
// @Produce json
// @Param body body model.SmsLoginRequest true "短信登录信息"
// @Success 200 {object} map[string]interface{} "code/msg/data，data 为 AuthenticationToken"
// @Router /api/v1/auth/login/sms [post]
func LoginBySms(c *gin.Context) {
	var req model.SmsLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if req.Mobile == "" || req.Code == "" {
		response.BadRequest(c, "手机号和验证码不能为空")
		return
	}

	token, err := service.LoginBySms(&req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, token)
}

// LoginByWechat 微信授权登录(Web)
// @Summary 微信授权登录
// @Description 使用微信授权码登录
// @Tags 01.认证接口
// @Accept json
// @Produce json
// @Param body body map[string]string true "微信授权信息 {\"code\":\"授权码\"}"
// @Success 200 {object} map[string]interface{} "code/msg/data，data 为 AuthenticationToken"
// @Router /api/v1/auth/login/wechat [post]
func LoginByWechat(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	code := req["code"]
	if code == "" {
		response.BadRequest(c, "微信授权码不能为空")
		return
	}

	token, err := service.LoginByWechat(code)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, token)
}

// LoginByWxMiniAppCode 微信小程序登录(Code)
// @Summary 微信小程序Code登录
// @Description 使用微信小程序Code登录
// @Tags 01.认证接口
// @Accept json
// @Produce json
// @Param body body model.WxMiniAppCodeLoginRequest true "微信小程序Code登录信息"
// @Success 200 {object} map[string]interface{} "code/msg/data，data 为 AuthenticationToken"
// @Router /api/v1/auth/wx/miniapp/code-login [post]
func LoginByWxMiniAppCode(c *gin.Context) {
	var req model.WxMiniAppCodeLoginRequest
	if err := validator.BindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}

	token, err := service.LoginByWxMiniAppCode(&req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, token)
}

// LoginByWxMiniAppPhone 微信小程序登录(手机号)
// @Summary 微信小程序手机号登录
// @Description 使用微信小程序获取手机号登录
// @Tags 01.认证接口
// @Accept json
// @Produce json
// @Param body body model.WxMiniAppPhoneLoginRequest true "微信小程序手机号登录信息"
// @Success 200 {object} map[string]interface{} "code/msg/data，data 为 AuthenticationToken"
// @Router /api/v1/auth/wx/miniapp/phone-login [post]
func LoginByWxMiniAppPhone(c *gin.Context) {
	var req model.WxMiniAppPhoneLoginRequest
	if err := validator.BindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}

	token, err := service.LoginByWxMiniAppPhone(&req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, token)
}
