package handler

import (
	"github.com/gin-gonic/gin"

	response "youlai-gin/internal/common"
	"youlai-gin/internal/common/auth"
	appContext "youlai-gin/internal/common/context"
	"youlai-gin/internal/common/validator"
	"youlai-gin/internal/middleware"
	"youlai-gin/internal/system/role/model"
	"youlai-gin/internal/system/role/service"
	"youlai-gin/pkg/enums"
)

// Handler 角色管理 HTTP 处理器
type Handler struct {
	svc *service.Service
}

// NewHandler 创建 Handler 实例
func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

// RegisterRoutes 注册角色路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	roles := r.Group("/roles")
	{
		roles.GET("", middleware.OperationLog(enums.LogModuleRole, enums.ActionTypeList), h.Page)
		roles.GET("/options", h.Options)
		roles.POST("", auth.RequirePermission("sys:role:create"), middleware.OperationLog(enums.LogModuleRole, enums.ActionTypeInsert), h.Create)
		roles.GET("/:id/form", auth.RequirePermission("sys:role:update"), h.GetForm)
		roles.PUT("/:id", auth.RequirePermission("sys:role:update"), middleware.OperationLog(enums.LogModuleRole, enums.ActionTypeUpdate), h.Update)
		roles.DELETE("/:id", auth.RequirePermission("sys:role:delete"), middleware.OperationLog(enums.LogModuleRole, enums.ActionTypeDelete), h.Delete)
		roles.GET("/:id/menu-ids", auth.RequirePermission("sys:role:update"), h.MenuIds)
		roles.PUT("/:id/menus", auth.RequirePermission("sys:role:assign"), middleware.OperationLog(enums.LogModuleRole, enums.ActionTypeGrant), h.AssignMenus)
		roles.GET("/:id/dept-ids", h.DeptIds)
		roles.PUT("/:id/depts", auth.RequirePermission("sys:role:update"), middleware.OperationLog(enums.LogModuleRole, enums.ActionTypeGrant), h.AssignDepts)
	}
}

// Page 角色分页列表
// @Summary 角色分页列表
// @Tags 03.角色接口
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keywords query string false "关键字"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles [get]
func (h *Handler) Page(c *gin.Context) {
	var query model.RoleQuery
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

// Options 角色下拉列表
// @Summary 角色下拉列表
// @Tags 03.角色接口
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/options [get]
func (h *Handler) Options(c *gin.Context) {
	options, err := h.svc.Options(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, options)
}

// Create 新增角色
// @Summary 新增角色
// @Tags 03.角色接口
// @Param body body model.RoleForm true "角色信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles [post]
func (h *Handler) Create(c *gin.Context) {
	var form model.RoleForm
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

// GetForm 获取角色表单数据
// @Summary 获取角色表单数据
// @Tags 03.角色接口
// @Param id path int true "角色ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/{id}/form [get]
func (h *Handler) GetForm(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "角色")
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

// Update 更新角色
// @Summary 更新角色
// @Tags 03.角色接口
// @Param id path int true "角色ID"
// @Param body body model.RoleForm true "角色信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "角色")
	if err != nil {
		c.Error(err)
		return
	}

	var form model.RoleForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.Update(appContext.OperatorCtx(c), id, &form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "更新成功")
}

// Delete 删除角色
// @Summary 删除角色
// @Tags 03.角色接口
// @Param id path int true "角色ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "角色")
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

// MenuIds 获取角色菜单ID列表
// @Summary 获取角色菜单ID列表
// @Tags 03.角色接口
// @Param id path int true "角色ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/{id}/menu-ids [get]
func (h *Handler) MenuIds(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "角色")
	if err != nil {
		c.Error(err)
		return
	}

	menuIds, err := h.svc.MenuIds(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, menuIds)
}

// AssignMenus 分配菜单权限
// @Summary 分配菜单权限
// @Tags 03.角色接口
// @Param id path int true "角色ID"
// @Param body body []int64 true "菜单ID列表"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/{id}/menus [put]
func (h *Handler) AssignMenus(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "角色")
	if err != nil {
		c.Error(err)
		return
	}

	var menuIds []int64
	if err := validator.BindJSON(c, &menuIds); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.AssignMenus(appContext.OperatorCtx(c), id, menuIds); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "分配成功")
}

// DeptIds 获取角色自定义部门ID列表
// @Summary 获取角色自定义部门ID列表
// @Tags 03.角色接口
// @Param id path int true "角色ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/{id}/dept-ids [get]
func (h *Handler) DeptIds(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "角色")
	if err != nil {
		c.Error(err)
		return
	}

	deptIds, err := h.svc.DeptIds(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, deptIds)
}

// AssignDepts 分配角色自定义部门
// @Summary 分配角色自定义部门
// @Tags 03.角色接口
// @Param id path int true "角色ID"
// @Param body body []int64 true "部门ID列表"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/{id}/depts [put]
func (h *Handler) AssignDepts(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "角色")
	if err != nil {
		c.Error(err)
		return
	}

	var deptIds []int64
	if err := validator.BindJSON(c, &deptIds); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.AssignDepts(appContext.OperatorCtx(c), id, deptIds); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "分配成功")
}
