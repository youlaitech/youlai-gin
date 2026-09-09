package handler

import (
	"github.com/gin-gonic/gin"

	response "youlai-gin/internal/common"
	"youlai-gin/internal/common/auth"
	appContext "youlai-gin/internal/common/context"
	"youlai-gin/internal/common/validator"
	"youlai-gin/internal/middleware"
	"youlai-gin/internal/system/dict/model"
	"youlai-gin/internal/system/dict/service"
	"youlai-gin/pkg/enums"
)

// Handler 字典接口层
type Handler struct {
	svc *service.Service
}

// NewHandler 创建 Handler 实例
func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

// RegisterRoutes 注册字典模块路由（路径参数 :id 为字典编码）
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	dicts := r.Group("/dicts")
	{
		dicts.GET("", middleware.OperationLog(enums.LogModuleDict, enums.ActionTypeList), h.Page)
		dicts.GET("/options", h.Options)
		dicts.POST("", auth.RequirePermission("sys:dict:create"), middleware.OperationLog(enums.LogModuleDict, enums.ActionTypeInsert), h.Create)
		dicts.GET("/:id/form", h.GetForm)
		dicts.PUT("/:id", auth.RequirePermission("sys:dict:update"), middleware.OperationLog(enums.LogModuleDict, enums.ActionTypeUpdate), h.Update)
		dicts.DELETE("/:id", auth.RequirePermission("sys:dict:delete"), middleware.OperationLog(enums.LogModuleDict, enums.ActionTypeDelete), h.Delete)
		dicts.GET("/:id/items", h.ItemPage)
		dicts.GET("/:id/items/options", h.Items)
		dicts.POST("/:id/items", auth.RequirePermission("sys:dict-item:create"), middleware.OperationLog(enums.LogModuleDict, enums.ActionTypeInsert), h.CreateItem)
		dicts.GET("/:id/items/:itemId/form", h.GetItemForm)
		dicts.PUT("/:id/items/:itemId", auth.RequirePermission("sys:dict-item:update"), middleware.OperationLog(enums.LogModuleDict, enums.ActionTypeUpdate), h.UpdateItem)
		dicts.DELETE("/:id/items/:itemIds", auth.RequirePermission("sys:dict-item:delete"), middleware.OperationLog(enums.LogModuleDict, enums.ActionTypeDelete), h.BatchDeleteItems)
	}
}

// Page 字典分页列表
func (h *Handler) Page(c *gin.Context) {
	var query model.DictQuery
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

// Options 字典下拉选项
func (h *Handler) Options(c *gin.Context) {
	options, err := h.svc.Options(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, options)
}

// Create 新增字典
func (h *Handler) Create(c *gin.Context) {
	var form model.DictForm
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

// GetForm 获取字典表单数据
func (h *Handler) GetForm(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "字典")
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

// Update 更新字典
func (h *Handler) Update(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "字典")
	if err != nil {
		c.Error(err)
		return
	}

	var form model.DictForm
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

// Delete 删除字典
func (h *Handler) Delete(c *gin.Context) {
	id, err := appContext.ParsePathParam(c, "id", "字典")
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

// ItemPage 字典项分页列表（字典编码取自路径）
func (h *Handler) ItemPage(c *gin.Context) {
	var query model.DictItemQuery
	if err := validator.BindQuery(c, &query); err != nil {
		c.Error(err)
		return
	}
	query.DictCode = c.Param("id")

	result, err := h.svc.ItemPage(c.Request.Context(), &query)
	if err != nil {
		c.Error(err)
		return
	}

	response.OkPaged(c, result)
}

// Items 字典项列表（字典编码取自路径）
func (h *Handler) Items(c *gin.Context) {
	items, err := h.svc.Items(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, items)
}

// CreateItem 新增字典项（字典编码取自路径）
func (h *Handler) CreateItem(c *gin.Context) {
	var form model.DictItemForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}
	form.DictCode = c.Param("id")

	if err := h.svc.CreateItem(appContext.OperatorCtx(c), &form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "保存成功")
}

// GetItemForm 获取字典项表单数据
func (h *Handler) GetItemForm(c *gin.Context) {
	itemId, err := appContext.ParsePathParam(c, "itemId", "字典项")
	if err != nil {
		c.Error(err)
		return
	}

	form, err := h.svc.GetItemForm(c.Request.Context(), itemId)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, form)
}

// UpdateItem 更新字典项（字典编码取自路径）
func (h *Handler) UpdateItem(c *gin.Context) {
	itemId, err := appContext.ParsePathParam(c, "itemId", "字典项")
	if err != nil {
		c.Error(err)
		return
	}

	var form model.DictItemForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}
	form.DictCode = c.Param("id")

	if err := h.svc.UpdateItem(appContext.OperatorCtx(c), itemId, &form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "修改成功")
}

// BatchDeleteItems 批量删除字典项
func (h *Handler) BatchDeleteItems(c *gin.Context) {
	ids, err := appContext.ParseIntList(c.Param("itemIds"), "字典项")
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.BatchDeleteItems(appContext.OperatorCtx(c), ids); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "删除成功")
}
