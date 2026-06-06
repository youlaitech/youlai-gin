package handler

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/system/dict/model"
	"youlai-gin/internal/system/dict/service"
	appContext "youlai-gin/internal/common/context"
	"youlai-gin/internal/common/auth"
	"youlai-gin/pkg/enums"
	"youlai-gin/internal/middleware"
	response "youlai-gin/internal/common"
	"youlai-gin/pkg/types"
	"youlai-gin/internal/common/validator"
)

// Handler 字典接口层
type Handler struct {
	svc *service.Service
}

// NewHandler 创建 Handler 实例
func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

// RegisterRoutes 注册字典模块路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	dicts := r.Group("/dicts")
	{
		dicts.GET("", middleware.OperationLog(enums.LogModuleDict, enums.ActionTypeList), h.GetDictPage)
		dicts.GET("/options", h.GetDictList)
		dicts.POST("", auth.RequirePermission("sys:dict:create"), middleware.OperationLog(enums.LogModuleDict, enums.ActionTypeInsert), h.SaveDict)
		dicts.GET("/:id/items", h.GetDictItemPageByCode)
		dicts.GET("/:id/items/options", h.GetDictItemsByCode)
		dicts.POST("/:id/items", auth.RequirePermission("sys:dict-item:create"), middleware.OperationLog(enums.LogModuleDict, enums.ActionTypeInsert), h.SaveDictItemByCode)
		dicts.GET("/:id/items/:itemId/form", h.GetDictItemFormByCode)
		dicts.PUT("/:id/items/:itemId", auth.RequirePermission("sys:dict-item:update"), middleware.OperationLog(enums.LogModuleDict, enums.ActionTypeUpdate), h.UpdateDictItemByCode)
		dicts.DELETE("/:id/items/:itemIds", auth.RequirePermission("sys:dict-item:delete"), middleware.OperationLog(enums.LogModuleDict, enums.ActionTypeDelete), h.DeleteDictItemsByCode)
		dicts.GET("/:id/form", h.GetDictForm)
		dicts.PUT("/:id", auth.RequirePermission("sys:dict:update"), middleware.OperationLog(enums.LogModuleDict, enums.ActionTypeUpdate), h.UpdateDict)
		dicts.DELETE("/:id", auth.RequirePermission("sys:dict:delete"), middleware.OperationLog(enums.LogModuleDict, enums.ActionTypeDelete), h.DeleteDict)
	}
}

func (h *Handler) GetDictPage(c *gin.Context) {
	var query model.DictQuery
	if err := validator.BindQuery(c, &query); err != nil { c.Error(err); return }
	r, e := h.svc.GetDictPage(&query)
	if e != nil { c.Error(e); return }
	response.OkPaged(c, r)
}
func (h *Handler) GetDictList(c *gin.Context) {
	l, e := h.svc.GetDictList()
	if e != nil { c.Error(e); return }
	response.Ok(c, l)
}
func (h *Handler) SaveDict(c *gin.Context) {
	var f model.DictForm
	if e := validator.BindJSON(c, &f); e != nil { c.Error(e); return }
	if e := h.svc.SaveDict(&f); e != nil { c.Error(e); return }
	response.OkMsg(c, "保存成功")
}
func (h *Handler) GetDictForm(c *gin.Context) {
	id, e := appContext.ParsePathParam(c, "id", "字典")
	if e != nil { c.Error(e); return }
	f, e := h.svc.GetDictForm(id)
	if e != nil { c.Error(e); return }
	response.Ok(c, f)
}
func (h *Handler) UpdateDict(c *gin.Context) {
	id, e := appContext.ParsePathParam(c, "id", "字典")
	if e != nil { c.Error(e); return }
	var f model.DictForm
	if e := validator.BindJSON(c, &f); e != nil { c.Error(e); return }
	f.ID = types.BigInt(id)
	if e := h.svc.SaveDict(&f); e != nil { c.Error(e); return }
	response.OkMsg(c, "更新成功")
}
func (h *Handler) DeleteDict(c *gin.Context) {
	id, e := appContext.ParsePathParam(c, "id", "字典")
	if e != nil { c.Error(e); return }
	if e := h.svc.DeleteDict(id); e != nil { c.Error(e); return }
	response.OkMsg(c, "删除成功")
}
func (h *Handler) GetDictItemsByCode(c *gin.Context) {
	i, e := h.svc.GetDictItems(c.Param("id"))
	if e != nil { c.Error(e); return }
	response.Ok(c, i)
}
func (h *Handler) GetDictItemPageByCode(c *gin.Context) {
	var q model.DictItemQuery
	if e := validator.BindQuery(c, &q); e != nil { c.Error(e); return }
	q.DictCode = c.Param("id")
	r, e := h.svc.GetDictItemPage(&q)
	if e != nil { c.Error(e); return }
	response.OkPaged(c, r)
}
func (h *Handler) SaveDictItemByCode(c *gin.Context) {
	var f model.DictItemForm
	if e := validator.BindJSON(c, &f); e != nil { c.Error(e); return }
	f.DictCode = c.Param("id")
	if e := h.svc.SaveDictItem(&f); e != nil { c.Error(e); return }
	response.OkMsg(c, "新增成功")
}
func (h *Handler) GetDictItemFormByCode(c *gin.Context) {
	id, e := appContext.ParsePathParam(c, "itemId", "字典项")
	if e != nil { c.Error(e); return }
	f, e := h.svc.GetDictItemForm(id)
	if e != nil { c.Error(e); return }
	response.Ok(c, f)
}
func (h *Handler) UpdateDictItemByCode(c *gin.Context) {
	id, e := appContext.ParsePathParam(c, "itemId", "字典项")
	if e != nil { c.Error(e); return }
	var f model.DictItemForm
	if e := validator.BindJSON(c, &f); e != nil { c.Error(e); return }
	f.ID = types.BigInt(id)
	f.DictCode = c.Param("id")
	if e := h.svc.SaveDictItem(&f); e != nil { c.Error(e); return }
	response.OkMsg(c, "更新成功")
}
func (h *Handler) DeleteDictItemsByCode(c *gin.Context) {
	ids, e := appContext.ParseIntList(c.Param("itemIds"), "字典项")
	if e != nil { c.Error(e); return }
	if e := h.svc.BatchDeleteDictItems(ids); e != nil { c.Error(e); return }
	response.OkMsg(c, "删除成功")
}
