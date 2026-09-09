package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	authModel "youlai-gin/internal/auth/model"
	"youlai-gin/internal/auth/service"
	response "youlai-gin/internal/common"
	pkgAuth "youlai-gin/internal/common/auth"
	appContext "youlai-gin/internal/common/context"
	"youlai-gin/internal/common/validator"
	"youlai-gin/pkg/errs"
)

// QrCodeHandler 扫码登录接口层
type QrCodeHandler struct {
	svc *service.QrCodeService
}

// NewQrCodeHandler 创建 QrCodeHandler 实例
func NewQrCodeHandler(svc *service.QrCodeService) *QrCodeHandler { return &QrCodeHandler{svc: svc} }

// RegisterRoutes 注册扫码登录路由，基础路径 /api/v1/auth/qr-code。
// generate/status/login 免登录；scan/confirm/cancel 需 APP 登录态，套 JWT 鉴权中间件。
func (h *QrCodeHandler) RegisterRoutes(r *gin.RouterGroup, tokenManager pkgAuth.TokenManager) {
	qr := r.Group("/auth/qr-code")
	qr.POST("/generate", h.Generate) // 免登录
	qr.GET("/status", h.Status)      // 免登录
	qr.POST("/login", h.Login)       // 免登录（PC 端）

	authQr := r.Group("/auth/qr-code")
	authQr.Use(pkgAuth.Middleware(tokenManager))
	authQr.POST("/scan", h.Scan)       // 需 APP 登录态
	authQr.POST("/confirm", h.Confirm) // 需 APP 登录态
	authQr.POST("/cancel", h.Cancel)   // 需 APP 登录态
}

// Generate 生成扫码登录票据
// @Summary 生成扫码登录票据
// @Tags 02.扫码登录
// @Produce json
// @Success 200 {object} map[string]interface{} "data 为 QrCodeGenerateVO"
// @Router /api/v1/auth/qr-code/generate [post]
func (h *QrCodeHandler) Generate(c *gin.Context) {
	vo, err := h.svc.Generate(c.Request.Context(), c.ClientIP())
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, vo)
}

// Status 查询扫码状态
// @Summary 查询扫码状态
// @Tags 02.扫码登录
// @Produce json
// @Param ticket query string true "票据号"
// @Success 200 {object} map[string]interface{} "data 为 QrCodeStatusVO"
// @Router /api/v1/auth/qr-code/status [get]
func (h *QrCodeHandler) Status(c *gin.Context) {
	vo, err := h.svc.Status(c.Request.Context(), c.Query("ticket"))
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, vo)
}

// Scan APP 标记已扫码
// @Summary APP 标记已扫码
// @Tags 02.扫码登录
// @Accept json
// @Produce json
// @Param body body model.QrCodeTicketForm true "票据号"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "data 为 QrCodeStatusVO"
// @Router /api/v1/auth/qr-code/scan [post]
func (h *QrCodeHandler) Scan(c *gin.Context) {
	vo, err := h.handleTicketAction(c, h.svc.Scan)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, vo)
}

// Confirm APP 确认登录
// @Summary APP 确认登录
// @Tags 02.扫码登录
// @Accept json
// @Produce json
// @Param body body model.QrCodeTicketForm true "票据号"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "data 为 QrCodeStatusVO"
// @Router /api/v1/auth/qr-code/confirm [post]
func (h *QrCodeHandler) Confirm(c *gin.Context) {
	vo, err := h.handleTicketAction(c, h.svc.Confirm)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, vo)
}

// Cancel APP 取消登录
// @Summary APP 取消登录
// @Tags 02.扫码登录
// @Accept json
// @Produce json
// @Param body body model.QrCodeTicketForm true "票据号"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "data 为 QrCodeStatusVO"
// @Router /api/v1/auth/qr-code/cancel [post]
func (h *QrCodeHandler) Cancel(c *gin.Context) {
	vo, err := h.handleTicketAction(c, h.svc.Cancel)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, vo)
}

// Login PC 端换取会话令牌
// @Summary PC 端换取会话令牌
// @Tags 02.扫码登录
// @Accept json
// @Produce json
// @Param body body model.QrCodeTicketForm true "票据号"
// @Success 200 {object} map[string]interface{} "data 为 AuthenticationToken"
// @Router /api/v1/auth/qr-code/login [post]
func (h *QrCodeHandler) Login(c *gin.Context) {
	var form authModel.QrCodeTicketForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	token, err := h.svc.Login(c.Request.Context(), form.Ticket)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, token)
}

// handleTicketAction 绑定票据表单并解析当前登录用户后执行扫码动作（scan/confirm/cancel 同构）
func (h *QrCodeHandler) handleTicketAction(
	c *gin.Context,
	action func(ctx context.Context, ticket string, userID int64) (*authModel.QrCodeStatusVO, error),
) (*authModel.QrCodeStatusVO, error) {
	var form authModel.QrCodeTicketForm
	if err := validator.BindJSON(c, &form); err != nil {
		return nil, err
	}

	userID, err := appContext.GetCurrentUserID(c)
	if err != nil {
		return nil, errs.TokenInvalid()
	}

	return action(c.Request.Context(), form.Ticket, userID)
}
