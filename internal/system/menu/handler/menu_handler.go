package handler

import (
	"github.com/gin-gonic/gin"

	response "youlai-gin/internal/common"
	"youlai-gin/internal/common/auth"
	appContext "youlai-gin/internal/common/context"
	"youlai-gin/internal/common/validator"
	"youlai-gin/internal/middleware"
	"youlai-gin/internal/system/menu/model"
	"youlai-gin/internal/system/menu/service"
	"youlai-gin/pkg/enums"
	"youlai-gin/pkg/errs"
)

// Handler 菜单接口层
type Handler struct {
	svc *service.Service
}

// NewHandler 创建 Handler 实例
func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

// RegisterRoutes 注册菜单路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	menus := r.Group("/menus")
	{
		menus.GET("", h.List)
		menus.GET("/options", h.Options)
		menus.GET("/routes", h.UserRoutes)
		menus.POST("", auth.RequirePermission("sys:menu:create"), middleware.OperationLog(enums.LogModuleMenu, enums.ActionTypeInsert), h.Create)
		menus.GET("/:id/form", auth.RequirePermission("sys:menu:update"), h.GetForm)
		menus.PUT("/:id", auth.RequirePermission("sys:menu:update"), middleware.OperationLog(enums.LogModuleMenu, enums.ActionTypeUpdate), h.Update)
		menus.DELETE("/:id", auth.RequirePermission("sys:menu:delete"), middleware.OperationLog(enums.LogModuleMenu, enums.ActionTypeDelete), h.Delete)
	}
}

// @Summary 菜单列表
// @Tags 04.菜单接口
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/menus [get]
func (h *Handler) List(c *gin.Context) {
	var query model.MenuQuery
	if err := validator.BindQuery(c, &query); err != nil {
		c.Error(err)
		return
	}

	list, err := h.svc.List(c.Request.Context(), &query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, list)
}

// @Summary 菜单下拉列表
// @Tags 04.菜单接口
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/menus/options [get]
func (h *Handler) Options(c *gin.Context) {
	onlyParent := c.Query("onlyParent") == "true"

	options, err := h.svc.Options(c.Request.Context(), onlyParent)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, options)
}

// @Summary 获取当前用户路由
// @Tags 04.菜单接口
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/menus/routes [get]
func (h *Handler) UserRoutes(c *gin.Context) {
	userId, err := appContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(errs.Unauthorized("未登录"))
		return
	}

	routes, err := h.svc.UserRoutes(c.Request.Context(), userId)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, routes)
}

// @Summary 新增菜单
// @Tags 04.菜单接口
// @Param body body model.MenuForm true "菜单信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/menus [post]
func (h *Handler) Create(c *gin.Context) {
	var form model.MenuForm
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

// @Summary 获取菜单表单数据
// @Tags 04.菜单接口
// @Param id path int true "菜单ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/menus/{id}/form [get]
func (h *Handler) GetForm(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "菜单")
	if err != nil {
		c.Error(err)
		return
	}

	form, err := h.svc.GetForm(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, form)
}

// @Summary 更新菜单
// @Tags 04.菜单接口
// @Param id path int true "菜单ID"
// @Param body body model.MenuForm true "菜单信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/menus/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "菜单")
	if err != nil {
		c.Error(err)
		return
	}

	var form model.MenuForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.Update(appContext.OperatorCtx(c), id, &form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "修改成功")
}

// @Summary 删除菜单
// @Tags 04.菜单接口
// @Param id path int true "菜单ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/menus/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "菜单")
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.Delete(appContext.OperatorCtx(c), id); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "删除成功")
}
