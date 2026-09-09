package handler

import (
	"github.com/gin-gonic/gin"

	response "youlai-gin/internal/common"
	"youlai-gin/internal/common/auth"
	appContext "youlai-gin/internal/common/context"
	"youlai-gin/internal/common/validator"
	"youlai-gin/internal/middleware"
	"youlai-gin/internal/system/config/model"
	"youlai-gin/internal/system/config/service"
	"youlai-gin/pkg/enums"
	"youlai-gin/pkg/errs"
)

// Handler 配置管理 HTTP 处理器
type Handler struct {
	svc *service.Service
}

// NewHandler 创建 Handler 实例
func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

// RegisterRoutes 注册配置管理路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	config := r.Group("/configs")
	{
		config.GET("", auth.RequirePermission("sys:config:list"), middleware.OperationLog(enums.LogModuleConfig, enums.ActionTypeList), h.Page)
		config.GET("/:id/form", h.GetForm)
		config.GET("/:id", h.Get)
		config.GET("/key/:key", h.GetByKey)
		config.POST("", auth.RequirePermission("sys:config:create"), middleware.OperationLog(enums.LogModuleConfig, enums.ActionTypeInsert), h.Create)
		config.PUT("/:id", auth.RequirePermission("sys:config:update"), middleware.OperationLog(enums.LogModuleConfig, enums.ActionTypeUpdate), h.Update)
		config.DELETE("/:ids", auth.RequirePermission("sys:config:delete"), middleware.OperationLog(enums.LogModuleConfig, enums.ActionTypeDelete), h.Delete)
		config.POST("/refresh/:key", auth.RequirePermission("sys:config:refresh"), h.RefreshCache)
		config.POST("/refresh", auth.RequirePermission("sys:config:refresh"), h.RefreshAllCache)
	}
}

// Page 配置分页列表
// @Summary 配置分页
// @Tags 07.系统配置
// @Router /api/v1/configs [get]
func (h *Handler) Page(c *gin.Context) {
	var query model.ConfigQuery
	if err := validator.BindQuery(c, &query); err != nil {
		c.Error(err)
		return
	}

	result, err := h.svc.Page(c.Request.Context(), &query)
	if err != nil {
		c.Error(err)
		return
	}

	response.OkPaged(c, result)
}

// GetForm 获取配置表单数据
// @Summary 配置表单
// @Tags 07.系统配置
// @Param id path int true "配置ID"
// @Router /api/v1/configs/{id}/form [get]
func (h *Handler) GetForm(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "配置")
	if err != nil {
		c.Error(err)
		return
	}

	formData, err := h.svc.GetForm(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, formData)
}

// Get 根据 ID 获取配置详情
// @Summary 配置详情
// @Tags 07.系统配置
// @Param id path int true "配置ID"
// @Router /api/v1/configs/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "配置")
	if err != nil {
		c.Error(err)
		return
	}

	config, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, config)
}

// GetByKey 根据 Key 获取配置
// @Summary 根据键获取配置
// @Tags 07.系统配置
// @Param key path string true "配置键"
// @Router /api/v1/configs/key/{key} [get]
func (h *Handler) GetByKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.Error(errs.BadRequest("配置Key不能为空"))
		return
	}

	config, err := h.svc.GetByKey(c.Request.Context(), key)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, config)
}

// Create 新增配置
// @Summary 新增配置
// @Tags 07.系统配置
// @Router /api/v1/configs [post]
func (h *Handler) Create(c *gin.Context) {
	var form model.ConfigForm
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

// Update 更新配置
// @Summary 更新配置
// @Tags 07.系统配置
// @Param id path int true "配置ID"
// @Router /api/v1/configs/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "配置")
	if err != nil {
		c.Error(err)
		return
	}

	var form model.ConfigForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	form.ID = id
	if err := h.svc.Update(appContext.OperatorCtx(c), &form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "更新成功")
}

// Delete 删除配置（支持批量）
// @Summary 删除配置
// @Tags 07.系统配置
// @Param ids path string true "配置ID列表"
// @Router /api/v1/configs/{ids} [delete]
func (h *Handler) Delete(c *gin.Context) {
	ids, err := appContext.ParseIntList(c.Param("ids"), "配置")
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.Delete(appContext.OperatorCtx(c), ids); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "删除成功")
}

// RefreshCache 刷新指定配置缓存
// @Summary 刷新配置缓存
// @Tags 07.系统配置
// @Param key path string true "配置键"
// @Router /api/v1/configs/refresh/{key} [post]
func (h *Handler) RefreshCache(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.Error(errs.BadRequest("配置Key不能为空"))
		return
	}

	if err := h.svc.RefreshCache(c.Request.Context(), key); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "刷新成功")
}

// RefreshAllCache 刷新所有配置缓存
// @Summary 刷新全部配置缓存
// @Tags 07.系统配置
// @Router /api/v1/configs/refresh [post]
func (h *Handler) RefreshAllCache(c *gin.Context) {
	h.svc.ClearAllCache(c.Request.Context())
	response.OkMsg(c, "刷新成功")
}
