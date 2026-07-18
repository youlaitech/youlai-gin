package handler

import (
	"github.com/gin-gonic/gin"

	authModel "youlai-gin/internal/auth/model"
	appContext "youlai-gin/internal/common/context"
	"youlai-gin/internal/auth/service"
	response "youlai-gin/internal/common"
	pkgAuth "youlai-gin/internal/common/auth"
	"youlai-gin/internal/common/validator"
	"youlai-gin/pkg/errs"
)

// RegisterQrCodeRoutes 注册扫码登录路由，基础路径 /api/v1/auth/qr-code
// generate/status/login 免登录；scan/confirm/cancel 需 APP 登录态，套 JWT 鉴权中间件。
func RegisterQrCodeRoutes(r *gin.RouterGroup, tokenManager pkgAuth.TokenManager) {
	qr := r.Group("/auth/qr-code")
	qr.POST("/generate", QrCodeGenerate) // 免登录
	qr.GET("/status", QrCodeStatus)      // 免登录
	qr.POST("/login", QrCodeLogin)       // 免登录（PC 端）

	authQr := r.Group("/auth/qr-code")
	authQr.Use(pkgAuth.Middleware(tokenManager))
	authQr.POST("/scan", QrCodeScan)     // 需 APP 登录态
	authQr.POST("/confirm", QrCodeConfirm) // 需 APP 登录态
	authQr.POST("/cancel", QrCodeCancel) // 需 APP 登录态
}

// QrCodeGenerate 生成扫码登录票据
// @Summary 生成扫码登录票据
// @Tags 02.扫码登录
// @Produce json
// @Success 200 {object} map[string]interface{} "data 为 QrCodeGenerateVO"
// @Router /api/v1/auth/qr-code/generate [post]
func QrCodeGenerate(c *gin.Context) {
	vo, err := service.QrCodeGenerate(c.ClientIP())
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, vo)
}

// QrCodeStatus 查询扫码状态
// @Summary 查询扫码状态
// @Tags 02.扫码登录
// @Produce json
// @Param ticket query string true "票据号"
// @Success 200 {object} map[string]interface{} "data 为 QrCodeStatusVO"
// @Router /api/v1/auth/qr-code/status [get]
func QrCodeStatus(c *gin.Context) {
	vo, err := service.QrCodeStatus(c.Query("ticket"))
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, vo)
}

// QrCodeScan APP 标记已扫码
// @Summary APP 标记已扫码
// @Tags 02.扫码登录
// @Accept json
// @Produce json
// @Param body body model.QrCodeTicketForm true "票据号"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "data 为 QrCodeStatusVO"
// @Router /api/v1/auth/qr-code/scan [post]
func QrCodeScan(c *gin.Context) {
	var form authModel.QrCodeTicketForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}
	userID, err := appContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(errs.TokenInvalid())
		return
	}
	vo, err := service.QrCodeScan(form.Ticket, userID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, vo)
}

// QrCodeConfirm APP 确认登录
// @Summary APP 确认登录
// @Tags 02.扫码登录
// @Accept json
// @Produce json
// @Param body body model.QrCodeTicketForm true "票据号"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "data 为 QrCodeStatusVO"
// @Router /api/v1/auth/qr-code/confirm [post]
func QrCodeConfirm(c *gin.Context) {
	var form authModel.QrCodeTicketForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}
	userID, err := appContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(errs.TokenInvalid())
		return
	}
	vo, err := service.QrCodeConfirm(form.Ticket, userID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, vo)
}

// QrCodeCancel APP 取消登录
// @Summary APP 取消登录
// @Tags 02.扫码登录
// @Accept json
// @Produce json
// @Param body body model.QrCodeTicketForm true "票据号"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "data 为 QrCodeStatusVO"
// @Router /api/v1/auth/qr-code/cancel [post]
func QrCodeCancel(c *gin.Context) {
	var form authModel.QrCodeTicketForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}
	userID, err := appContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(errs.TokenInvalid())
		return
	}
	vo, err := service.QrCodeCancel(form.Ticket, userID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, vo)
}

// QrCodeLogin PC 端换取会话令牌
// @Summary PC 端换取会话令牌
// @Tags 02.扫码登录
// @Accept json
// @Produce json
// @Param body body model.QrCodeTicketForm true "票据号"
// @Success 200 {object} map[string]interface{} "data 为 AuthenticationToken"
// @Router /api/v1/auth/qr-code/login [post]
func QrCodeLogin(c *gin.Context) {
	var form authModel.QrCodeTicketForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}
	token, err := service.QrCodeLogin(form.Ticket)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, token)
}
