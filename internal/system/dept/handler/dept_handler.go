package handler

import (
	"github.com/gin-gonic/gin"

	response "youlai-gin/internal/common"
	"youlai-gin/internal/common/auth"
	appContext "youlai-gin/internal/common/context"
	"youlai-gin/internal/common/validator"
	"youlai-gin/internal/middleware"
	"youlai-gin/internal/system/dept/model"
	"youlai-gin/internal/system/dept/service"
	"youlai-gin/pkg/enums"
)

// Handler 部门接口层
type Handler struct {
	svc *service.Service
}

// NewHandler 创建 Handler 实例
func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes 注册部门路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	depts := r.Group("/depts")
	{
		depts.GET("", h.List)
		depts.GET("/options", h.Options)
		depts.POST("", auth.RequirePermission("sys:dept:create"), middleware.OperationLog(enums.LogModuleDept, enums.ActionTypeInsert), h.Create)
		depts.GET("/:id/form", h.GetForm)
		depts.PUT("/:id", auth.RequirePermission("sys:dept:update"), middleware.OperationLog(enums.LogModuleDept, enums.ActionTypeUpdate), h.Update)
		depts.DELETE("/:id", auth.RequirePermission("sys:dept:delete"), middleware.OperationLog(enums.LogModuleDept, enums.ActionTypeDelete), h.Delete)
	}
}

// List 部门列表（树形）
func (h *Handler) List(c *gin.Context) {
	var query model.DeptQuery
	if err := validator.BindQuery(c, &query); err != nil {
		c.Error(err)
		return
	}

	currentUser, err := appContext.GetCurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	list, err := h.svc.List(c.Request.Context(), &query, currentUser)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, list)
}

// Options 部门下拉列表（树形）
func (h *Handler) Options(c *gin.Context) {
	currentUser, err := appContext.GetCurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	options, err := h.svc.Options(c.Request.Context(), currentUser)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, options)
}

// Create 新增部门
func (h *Handler) Create(c *gin.Context) {
	var form model.DeptForm
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

// GetForm 获取部门表单数据
func (h *Handler) GetForm(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "部门")
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

// Update 更新部门
func (h *Handler) Update(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "部门")
	if err != nil {
		c.Error(err)
		return
	}

	var form model.DeptForm
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

// Delete 删除部门
func (h *Handler) Delete(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "部门")
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
