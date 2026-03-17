package handler

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/auth/model"
	"youlai-gin/internal/auth/service"
	"youlai-gin/pkg/response"
)

// RegisterWechatRoutes 注册微信小程序认证路由
func RegisterWechatRoutes(r *gin.RouterGroup) {
	r.POST("/wechat/miniapp/auth/silent-login", WechatSilentLogin)
	r.POST("/wechat/miniapp/auth/phone-login", WechatPhoneLogin)
	r.POST("/wechat/miniapp/auth/bind-mobile", WechatBindMobile)
}

// WechatSilentLogin 静默登录
// @Summary 静默登录
// @Description 微信小程序静默登录
// @Tags 13.微信小程序认证
// @Accept application/json
// @Produce json
// @Param body body model.WechatSilentLoginRequest true "登录信息"
// @Success 200 {object} map[string]interface{} "code/msg/data"
// @Router /api/v1/wechat/miniapp/auth/silent-login [post]
func WechatSilentLogin(c *gin.Context) {
	var req model.WechatSilentLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if req.Code == "" {
		response.BadRequest(c, "code不能为空")
		return
	}

	result, err := service.SilentLogin(req.Code)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}

// WechatPhoneLogin 手机号快捷登录
// @Summary 手机号快捷登录
// @Description 微信小程序手机号快捷登录
// @Tags 13.微信小程序认证
// @Accept application/json
// @Produce json
// @Param body body model.WechatPhoneLoginRequest true "登录信息"
// @Success 200 {object} map[string]interface{} "code/msg/data"
// @Router /api/v1/wechat/miniapp/auth/phone-login [post]
func WechatPhoneLogin(c *gin.Context) {
	var req model.WechatPhoneLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if req.LoginCode == "" || req.PhoneCode == "" {
		response.BadRequest(c, "loginCode和phoneCode不能为空")
		return
	}

	result, err := service.PhoneLogin(req.LoginCode, req.PhoneCode)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}

// WechatBindMobile 绑定手机号
// @Summary 绑定手机号
// @Description 微信小程序绑定手机号
// @Tags 13.微信小程序认证
// @Accept application/json
// @Produce json
// @Param body body model.WechatBindMobileRequest true "绑定信息"
// @Success 200 {object} map[string]interface{} "code/msg/data"
// @Router /api/v1/wechat/miniapp/auth/bind-mobile [post]
func WechatBindMobile(c *gin.Context) {
	var req model.WechatBindMobileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if req.OpenID == "" || req.Mobile == "" || req.SmsCode == "" {
		response.BadRequest(c, "openId、mobile和smsCode不能为空")
		return
	}

	result, err := service.BindMobile(req.OpenID, req.Mobile, req.SmsCode)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}
