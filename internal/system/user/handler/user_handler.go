package handler

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	response "youlai-gin/internal/common"
	"youlai-gin/internal/common/auth"
	appContext "youlai-gin/internal/common/context"
	"youlai-gin/internal/common/utils"
	"youlai-gin/internal/common/validator"
	"youlai-gin/internal/middleware"
	"youlai-gin/internal/system/user/model"
	"youlai-gin/internal/system/user/service"
	"youlai-gin/pkg/enums"
	"youlai-gin/pkg/errs"
)

// Handler 用户接口处理器
type Handler struct {
	svc *service.Service
}

// NewHandler 创建 Handler 实例
func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes 注册用户路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// 写操作 - 需要权限
	r.POST("/users", auth.RequirePermission("sys:user:create"), middleware.OperationLog(enums.LogModuleUser, enums.ActionTypeInsert), h.Create)
	r.PUT("/users/:userId", auth.RequirePermission("sys:user:update"), middleware.OperationLog(enums.LogModuleUser, enums.ActionTypeUpdate), h.Update)
	r.DELETE("/users/:ids", auth.RequirePermission("sys:user:delete"), middleware.OperationLog(enums.LogModuleUser, enums.ActionTypeDelete), h.Delete)
	r.PATCH("/users/:userId/status", auth.RequirePermission("sys:user:update"), middleware.OperationLog(enums.LogModuleUser, enums.ActionTypeUpdate), h.UpdateStatus)
	r.PUT("/users/:userId/password/reset", auth.RequirePermission("sys:user:reset-password"), middleware.OperationLog(enums.LogModuleUser, enums.ActionTypeResetPassword), h.ResetPassword)
	r.POST("/users/import", auth.RequirePermission("sys:user:import"), middleware.OperationLog(enums.LogModuleUser, enums.ActionTypeImport), h.Import)

	// 读操作 - 无需权限
	r.GET("/users", middleware.OperationLog(enums.LogModuleUser, enums.ActionTypeList), h.Page)
	r.GET("/users/:userId/form", auth.RequirePermission("sys:user:update"), h.GetForm)
	r.GET("/users/export", auth.RequirePermission("sys:user:export"), h.Export)
	r.GET("/users/template", h.Template)
	r.GET("/users/options", h.Options)

	// 个人操作 - 无需权限
	r.GET("/users/me", h.CurrentUser)
	r.GET("/users/profile", h.Profile)
	r.PUT("/users/profile", middleware.OperationLog(enums.LogModuleUser, enums.ActionTypeUpdate), h.UpdateProfile)
	r.PUT("/users/password", middleware.OperationLog(enums.LogModuleUser, enums.ActionTypeChangePassword), h.ChangePassword)
	r.POST("/users/mobile/code", h.SendMobileCode)
	r.PUT("/users/mobile", middleware.OperationLog(enums.LogModuleUser, enums.ActionTypeUpdate), h.BindOrChangeMobile)
	r.DELETE("/users/mobile", middleware.OperationLog(enums.LogModuleUser, enums.ActionTypeUpdate), h.UnbindMobile)
	r.POST("/users/email/code", h.SendEmailCode)
	r.PUT("/users/email", middleware.OperationLog(enums.LogModuleUser, enums.ActionTypeUpdate), h.BindOrChangeEmail)
	r.DELETE("/users/email", middleware.OperationLog(enums.LogModuleUser, enums.ActionTypeUpdate), h.UnbindEmail)
}

// Page 用户分页列表
// @Summary 用户分页列表
// @Tags 02.用户接口
// @Router /api/v1/users [get]
func (h *Handler) Page(c *gin.Context) {
	var query model.UserQuery
	if err := validator.BindQuery(c, &query); err != nil {
		c.Error(err)
		return
	}

	currentUser, err := appContext.GetCurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := h.svc.Page(c.Request.Context(), &query, currentUser)
	if err != nil {
		c.Error(err)
		return
	}

	response.OkPaged(c, result)
}

// Create 新增用户
// @Summary 保存用户
// @Tags 02.用户接口
// @Router /api/v1/users [post]
func (h *Handler) Create(c *gin.Context) {
	var form model.UserForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.Create(appContext.OperatorCtx(c), &form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "保存成功")
}

// GetForm 获取用户表单数据
// @Summary 获取用户表单
// @Tags 02.用户接口
// @Param userId path int true "用户ID"
// @Router /api/v1/users/{userId}/form [get]
func (h *Handler) GetForm(c *gin.Context) {
	userId, err := appContext.ParsePathParam(c, "userId", "用户")
	if err != nil {
		c.Error(err)
		return
	}

	formData, err := h.svc.GetForm(c.Request.Context(), userId)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, formData)
}

// Update 修改用户
// @Summary 更新用户
// @Tags 02.用户接口
// @Param userId path int true "用户ID"
// @Router /api/v1/users/{userId} [put]
func (h *Handler) Update(c *gin.Context) {
	userId, err := appContext.ParsePathParam(c, "userId", "用户")
	if err != nil {
		c.Error(err)
		return
	}

	var form model.UserForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.Update(appContext.OperatorCtx(c), userId, &form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "修改成功")
}

// Delete 批量删除用户
// @Summary 删除用户
// @Tags 02.用户接口
// @Param ids path string true "用户ID列表"
// @Router /api/v1/users/{ids} [delete]
func (h *Handler) Delete(c *gin.Context) {
	ids := c.Param("ids")

	if err := h.svc.Delete(appContext.OperatorCtx(c), ids); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "删除成功")
}

// UpdateStatus 修改用户状态
// @Summary 修改用户状态
// @Tags 02.用户接口
// @Param userId path int true "用户ID"
// @Router /api/v1/users/{userId}/status [patch]
func (h *Handler) UpdateStatus(c *gin.Context) {
	userId, err := appContext.ParsePathParam(c, "userId", "用户")
	if err != nil {
		c.Error(err)
		return
	}

	statusStr := c.Query("status")
	if statusStr == "" {
		c.Error(errs.BadRequest("状态值不能为空"))
		return
	}
	status, err := strconv.Atoi(statusStr)
	if err != nil {
		c.Error(errs.BadRequest("无效的状态值"))
		return
	}

	if err := h.svc.UpdateStatus(appContext.OperatorCtx(c), userId, status); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "修改成功")
}

// CurrentUser 获取当前登录用户信息
// @Summary 当前登录用户
// @Tags 02.用户接口
// @Router /api/v1/users/me [get]
func (h *Handler) CurrentUser(c *gin.Context) {
	// 角色信息取自 token，用于查询权限
	userDetails, err := appContext.GetCurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	currentUser, err := h.svc.CurrentUser(c.Request.Context(), userDetails.UserID, userDetails.Roles)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, currentUser)
}

// Profile 获取个人中心用户信息
// @Summary 个人中心信息
// @Tags 02.用户接口
// @Router /api/v1/users/profile [get]
func (h *Handler) Profile(c *gin.Context) {
	userId, err := appContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	profile, err := h.svc.Profile(c.Request.Context(), userId)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, profile)
}

// UpdateProfile 个人中心修改用户信息
// @Summary 更新个人中心信息
// @Tags 02.用户接口
// @Router /api/v1/users/profile [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	userId, err := appContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var form model.UserProfileForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.UpdateProfile(appContext.OperatorCtx(c), userId, &form); err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, true)
}

// ResetPassword 重置指定用户密码
// @Summary 重置用户密码
// @Tags 02.用户接口
// @Param userId path int true "用户ID"
// @Router /api/v1/users/{userId}/password/reset [put]
func (h *Handler) ResetPassword(c *gin.Context) {
	userId, err := appContext.ParsePathParam(c, "userId", "用户")
	if err != nil {
		c.Error(err)
		return
	}

	var form model.PasswordResetForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.ResetPassword(appContext.OperatorCtx(c), userId, form.Password); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "重置成功")
}

// ChangePassword 当前用户修改密码
// @Summary 修改当前用户密码
// @Tags 02.用户接口
// @Router /api/v1/users/password [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	userId, err := appContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var form model.PasswordForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.ChangePassword(appContext.OperatorCtx(c), userId, &form); err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, true)
}

// SendMobileCode 发送短信验证码
// @Summary 发送手机号验证码
// @Tags 02.用户接口
// @Router /api/v1/users/mobile/code [post]
func (h *Handler) SendMobileCode(c *gin.Context) {
	mobile := c.Query("mobile")
	if mobile == "" {
		c.Error(errs.BadRequest("手机号不能为空"))
		return
	}

	if err := h.svc.SendMobileCode(c.Request.Context(), mobile); err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, true)
}

// BindOrChangeMobile 绑定或更换手机号
// @Summary 绑定或更换手机号
// @Tags 02.用户接口
// @Router /api/v1/users/mobile [put]
func (h *Handler) BindOrChangeMobile(c *gin.Context) {
	userId, err := appContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var form model.MobileBindingForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.BindOrChangeMobile(appContext.OperatorCtx(c), userId, &form); err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, true)
}

// UnbindMobile 解绑手机号
// @Summary 解绑手机号
// @Tags 02.用户接口
// @Router /api/v1/users/mobile [delete]
func (h *Handler) UnbindMobile(c *gin.Context) {
	userId, err := appContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var form model.PasswordVerifyForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.UnbindMobile(appContext.OperatorCtx(c), userId, &form); err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, true)
}

// SendEmailCode 发送邮箱验证码
// @Summary 发送邮箱验证码
// @Tags 02.用户接口
// @Router /api/v1/users/email/code [post]
func (h *Handler) SendEmailCode(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.Error(errs.BadRequest("邮箱不能为空"))
		return
	}

	if err := h.svc.SendEmailCode(c.Request.Context(), email); err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, true)
}

// BindOrChangeEmail 绑定或更换邮箱
// @Summary 绑定或更换邮箱
// @Tags 02.用户接口
// @Router /api/v1/users/email [put]
func (h *Handler) BindOrChangeEmail(c *gin.Context) {
	userId, err := appContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var form model.EmailBindingForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.BindOrChangeEmail(appContext.OperatorCtx(c), userId, &form); err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, true)
}

// UnbindEmail 解绑邮箱
// @Summary 解绑邮箱
// @Tags 02.用户接口
// @Router /api/v1/users/email [delete]
func (h *Handler) UnbindEmail(c *gin.Context) {
	userId, err := appContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var form model.PasswordVerifyForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.UnbindEmail(appContext.OperatorCtx(c), userId, &form); err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, true)
}

// Options 获取用户下拉选项
// @Summary 用户下拉选项
// @Tags 02.用户接口
// @Router /api/v1/users/options [get]
func (h *Handler) Options(c *gin.Context) {
	options, err := h.svc.Options(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, options)
}

// Export 导出用户列表
// @Summary 导出用户
// @Tags 02.用户接口
// @Router /api/v1/users/export [get]
func (h *Handler) Export(c *gin.Context) {
	var query model.UserQuery
	if err := validator.BindQuery(c, &query); err != nil {
		c.Error(err)
		return
	}

	currentUser, err := appContext.GetCurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	exporter, err := h.svc.ExportToExcel(c.Request.Context(), &query, currentUser)
	if err != nil {
		c.Error(err)
		return
	}
	defer exporter.Close()

	filename := fmt.Sprintf("用户列表_%s.xlsx", time.Now().Format("20060102150405"))
	encodedFilename := url.QueryEscape(filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", encodedFilename))
	c.Header("Content-Transfer-Encoding", "binary")

	if err := exporter.Write(c.Writer); err != nil {
		c.Error(errs.SystemError("导出文件失败"))
		return
	}
}

// Template 下载用户导入模板
// @Summary 下载用户导入模板
// @Tags 02.用户接口
// @Router /api/v1/users/template [get]
func (h *Handler) Template(c *gin.Context) {
	exporter, err := h.svc.GenerateTemplate()
	if err != nil {
		c.Error(err)
		return
	}
	defer exporter.Close()

	filename := "用户导入模板.xlsx"
	encodedFilename := url.QueryEscape(filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", encodedFilename))
	c.Header("Content-Transfer-Encoding", "binary")

	if err := exporter.Write(c.Writer); err != nil {
		c.Error(errs.SystemError("生成模板失败"))
		return
	}
}

// Import 导入用户数据
// @Summary 导入用户
// @Tags 02.用户接口
// @Router /api/v1/users/import [post]
func (h *Handler) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.Error(errs.BadRequest("请选择要导入的文件"))
		return
	}

	if err := utils.ValidateExcel(file); err != nil {
		c.Error(err)
		return
	}

	f, err := file.Open()
	if err != nil {
		c.Error(errs.SystemError("文件打开失败"))
		return
	}
	defer f.Close()

	result, err := h.svc.ImportFromExcel(appContext.OperatorCtx(c), f)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}
