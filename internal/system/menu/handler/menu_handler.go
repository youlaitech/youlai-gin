package handler

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/system/menu/model"
	"youlai-gin/internal/system/menu/service"
	"youlai-gin/internal/common/auth"
	appContext "youlai-gin/internal/common/context"
	"youlai-gin/pkg/enums"
	"youlai-gin/pkg/errs"
	"youlai-gin/internal/middleware"
	response "youlai-gin/internal/common"
	"youlai-gin/pkg/types"
	"youlai-gin/internal/common/validator"
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
		menus.GET("", h.GetMenuList)
		menus.GET("/options", h.GetMenuOptions)
		menus.GET("/routes", h.GetCurrentUserRoutes)
		menus.POST("", auth.RequirePermission("sys:menu:create"), middleware.OperationLog(enums.LogModuleMenu, enums.ActionTypeInsert), h.SaveMenu)
		menus.GET("/:id/form", auth.RequirePermission("sys:menu:update"), h.GetMenuForm)
		menus.PUT("/:id", auth.RequirePermission("sys:menu:update"), middleware.OperationLog(enums.LogModuleMenu, enums.ActionTypeUpdate), h.UpdateMenu)
		menus.DELETE("/:id", auth.RequirePermission("sys:menu:delete"), middleware.OperationLog(enums.LogModuleMenu, enums.ActionTypeDelete), h.DeleteMenu)
	}
}

// @Summary 菜单列表


// @Summary 菜单列表
// @Tags 04.菜单接口
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/menus [get]
func (h *Handler) GetMenuList(c *gin.Context) {
	var query model.MenuQuery
	if err := validator.BindQuery(c, &query); err != nil {
		c.Error(err)
		return
	}

	list, err := h.svc.GetMenuList(&query)
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
func (h *Handler) GetMenuOptions(c *gin.Context) {
	onlyParent := c.Query("onlyParent") == "true"

	options, err := h.svc.GetMenuOptions(onlyParent)
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
func (h *Handler) GetCurrentUserRoutes(c *gin.Context) {
	userId, err := appContext.GetUserIDMust(c)
	if err != nil {
		c.Error(errs.Unauthorized("未登录"))
		return
	}

	routes, err := h.svc.GetCurrentUserRoutes(userId)
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
func (h *Handler) SaveMenu(c *gin.Context) {
	var form model.MenuForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.SaveMenu(&form); err != nil {
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
func (h *Handler) GetMenuForm(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "菜单")
	if err != nil {
		c.Error(err)
		return
	}

	form, err := h.svc.GetMenuForm(id)
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
func (h *Handler) UpdateMenu(c *gin.Context) {
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

	form.ID = types.BigInt(id)
	if err := h.svc.SaveMenu(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "更新成功")
}

// @Summary 删除菜单
// @Tags 04.菜单接口
// @Param id path int true "菜单ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/menus/{id} [delete]
func (h *Handler) DeleteMenu(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "菜单")
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.DeleteMenu(id); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "删除成功")
}
