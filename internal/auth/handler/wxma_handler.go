package handler

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/auth/model"
	"youlai-gin/internal/auth/service"
	response "youlai-gin/internal/common"
	"youlai-gin/internal/common/validator"
	"youlai-gin/pkg/errs"
)

// WxMaHandler 微信小程序认证接口层
type WxMaHandler struct {
	svc *service.WxMaService
}

// NewWxMaHandler 创建 WxMaHandler 实例
func NewWxMaHandler(svc *service.WxMaService) *WxMaHandler { return &WxMaHandler{svc: svc} }

// RegisterRoutes 注册微信小程序认证路由
func (h *WxMaHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/wxma/auth/silent-login", h.SilentLogin)
	r.POST("/wxma/auth/phone-login", h.PhoneLogin)
	r.POST("/wxma/auth/bind-mobile", h.BindMobile)
}

// SilentLogin 静默登录
// @Summary 静默登录
// @Description 微信小程序静默登录
// @Tags 12.微信小程序认证
// @Accept application/json
// @Produce json
// @Param body body model.WxMaSilentLoginRequest true "登录信息"
// @Success 200 {object} map[string]interface{} "code/msg/data"
// @Router /api/v1/wxma/auth/silent-login [post]
func (h *WxMaHandler) SilentLogin(c *gin.Context) {
	var req model.WxMaSilentLoginRequest
	if err := validator.BindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}

	if req.Code == "" {
		c.Error(errs.BadRequest("code 不能为空"))
		return
	}

	result, err := h.svc.SilentLogin(c.Request.Context(), req.Code)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}

// PhoneLogin 手机号快捷登录
// @Summary 手机号快捷登录
// @Description 微信小程序手机号快捷登录
// @Tags 12.微信小程序认证
// @Accept application/json
// @Produce json
// @Param body body model.WxMaPhoneLoginRequest true "登录信息"
// @Success 200 {object} map[string]interface{} "code/msg/data"
// @Router /api/v1/wxma/auth/phone-login [post]
func (h *WxMaHandler) PhoneLogin(c *gin.Context) {
	var req model.WxMaPhoneLoginRequest
	if err := validator.BindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}

	if req.LoginCode == "" || req.PhoneCode == "" {
		c.Error(errs.BadRequest("loginCode 和 phoneCode 不能为空"))
		return
	}

	result, err := h.svc.PhoneLogin(c.Request.Context(), req.LoginCode, req.PhoneCode)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}

// BindMobile 绑定手机号
// @Summary 绑定手机号
// @Description 微信小程序绑定手机号
// @Tags 12.微信小程序认证
// @Accept application/json
// @Produce json
// @Param body body model.WxMaBindMobileRequest true "绑定信息"
// @Success 200 {object} map[string]interface{} "code/msg/data"
// @Router /api/v1/wxma/auth/bind-mobile [post]
func (h *WxMaHandler) BindMobile(c *gin.Context) {
	var req model.WxMaBindMobileRequest
	if err := validator.BindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}

	if req.OpenID == "" || req.Mobile == "" || req.SmsCode == "" {
		c.Error(errs.BadRequest("openId、mobile 和 smsCode 不能为空"))
		return
	}

	result, err := h.svc.BindMobile(c.Request.Context(), req.OpenID, req.Mobile, req.SmsCode)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}
