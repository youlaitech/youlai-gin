package handler

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/system/dept/model"
	"youlai-gin/internal/system/dept/service"
	"youlai-gin/internal/common/auth"
	appContext "youlai-gin/internal/common/context"
	"youlai-gin/internal/middleware"
	response "youlai-gin/internal/common"
	"youlai-gin/pkg/enums"
	"youlai-gin/pkg/types"
	"youlai-gin/internal/common/validator"
)

// Handler 部门接口层
type Handler struct {
	svc *service.Service
}

// NewHandler Wire Provider
func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes 注册部门路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	depts := r.Group("/depts")
	{
		depts.GET("", h.GetDeptList)
		depts.GET("/options", h.GetDeptOptions)
		depts.POST("", auth.RequirePermission("sys:dept:create"), middleware.OperationLog(enums.LogModuleDept, enums.ActionTypeInsert), h.SaveDept)
		depts.GET("/:id/form", h.GetDeptForm)
		depts.PUT("/:id", auth.RequirePermission("sys:dept:update"), middleware.OperationLog(enums.LogModuleDept, enums.ActionTypeUpdate), h.UpdateDept)
		depts.DELETE("/:id", auth.RequirePermission("sys:dept:delete"), middleware.OperationLog(enums.LogModuleDept, enums.ActionTypeDelete), h.DeleteDept)
	}
}

// GetDeptList 部门列表
func (h *Handler) GetDeptList(c *gin.Context) {
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

	list, err := h.svc.GetDeptList(&query, currentUser)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, list)
}

// GetDeptOptions 部门下拉列表
func (h *Handler) GetDeptOptions(c *gin.Context) {
	currentUser, err := appContext.GetCurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	options, err := h.svc.GetDeptOptions(currentUser)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, options)
}

// SaveDept 新增部门
func (h *Handler) SaveDept(c *gin.Context) {
	var form model.DeptForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.SaveDept(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "保存成功")
}

// GetDeptForm 获取部门表单数据
func (h *Handler) GetDeptForm(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "部门")
	if err != nil {
		c.Error(err)
		return
	}

	form, err := h.svc.GetDeptForm(id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, form)
}

// UpdateDept 更新部门
func (h *Handler) UpdateDept(c *gin.Context) {
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

	form.ID = types.BigInt(id)
	if err := h.svc.SaveDept(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "更新成功")
}

// DeleteDept 删除部门
func (h *Handler) DeleteDept(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "部门")
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.DeleteDept(id); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "删除成功")
}
