package handler

import (
	"strconv"
	"strings"
	"youlai-gin/internal/system/dict/model"
	"youlai-gin/internal/system/dict/service"
	"youlai-gin/pkg/response"
	"youlai-gin/pkg/types"
	"youlai-gin/pkg/validator"

	"github.com/gin-gonic/gin"
)

// RegisterDictRoutes 注册字典模块路由
func RegisterDictRoutes(r *gin.RouterGroup) {
	dicts := r.Group("/dicts")
	{
		// 字典分页查询（对齐 youlai-boot：GET /dicts）
		dicts.GET("", GetDictPage)
		// 字典下拉列表（对齐 youlai-boot：GET /dicts/options）
		dicts.GET("/options", GetDictList)
		dicts.POST("", SaveDict)

		// RESTful风格的字典项路由（使用统一的参数名 :id 来避免冲突）
		dicts.GET("/:id/items", GetDictItemPageByCode)
		dicts.GET("/:id/items/options", GetDictItemsByCode)
		dicts.POST("/:id/items", SaveDictItemByCode)
		dicts.GET("/:id/items/:itemId/form", GetDictItemFormByCode)
		dicts.PUT("/:id/items/:itemId", UpdateDictItemByCode)
		dicts.DELETE("/:id/items/:itemIds", DeleteDictItemsByCode)

		// 字典CRUD操作
		dicts.GET("/:id/form", GetDictForm)
		dicts.PUT("/:id", UpdateDict)
		dicts.DELETE("/:id", DeleteDict)
	}
}

// GetDictPage 字典分页列表
// @Summary 字典分页列表
// @Tags 字典管理
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/dicts [get]
func GetDictPage(c *gin.Context) {
	var query model.DictQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	result, err := service.GetDictPage(&query)
	if err != nil {
		c.Error(err)
		return
	}

	response.OkPaged(c, result)
}

func GetDictItemPageByCode(c *gin.Context) {
	dictCode := c.Param("id")

	var query model.DictItemQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	query.DictCode = dictCode

	result, err := service.GetDictItemPage(&query)
	if err != nil {
		c.Error(err)
		return
	}

	response.OkPaged(c, result)
}

// GetDictList 获取字典下拉列表
// @Summary 字典下拉列表
// @Tags 字典管理
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/dicts/options [get]
func GetDictList(c *gin.Context) {
	list, err := service.GetDictList()
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, list)
}

// SaveDict 新增字典
// @Summary 新增字典
// @Tags 字典管理
// @Param body body model.DictForm true "字典信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/dicts [post]
func SaveDict(c *gin.Context) {
	var form model.DictForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := service.SaveDict(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "保存成功")
}

// GetDictForm 获取字典表单数据
// @Summary 获取字典表单数据
// @Tags 字典管理
// @Param id path int true "字典ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/dicts/{id}/form [get]
func GetDictForm(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的字典ID")
		return
	}

	form, err := service.GetDictForm(id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, form)
}

// UpdateDict 更新字典
// @Summary 更新字典
// @Tags 字典管理
// @Param id path int true "字典ID"
// @Param body body model.DictForm true "字典信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/dicts/{id} [put]
func UpdateDict(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的字典ID")
		return
	}

	var form model.DictForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	form.ID = types.BigInt(id)
	if err := service.SaveDict(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "更新成功")
}

// DeleteDict 删除字典
// @Summary 删除字典
// @Tags 字典管理
// @Param id path int true "字典ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/dicts/{id} [delete]
func DeleteDict(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的字典ID")
		return
	}

	if err := service.DeleteDict(id); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "删除成功")
}

// ========== RESTful风格的字典项接口（匹配Java版本）==========

// GetDictItemsByCode 字典项列表（RESTful）
// @Summary 字典项列表（RESTful）
// @Tags 字典管理
// @Param id path string true "字典编码"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/dicts/{id}/items [get]
func GetDictItemsByCode(c *gin.Context) {
	dictCode := c.Param("id") // 从 :id 参数获取字典编码

	items, err := service.GetDictItems(dictCode)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, items)
}

// SaveDictItemByCode 新增字典项（RESTful）
// @Summary 新增字典项（RESTful）
// @Tags 字典管理
// @Param id path string true "字典编码"
// @Param body body model.DictItemForm true "字典项信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/dicts/{id}/items [post]
func SaveDictItemByCode(c *gin.Context) {
	dictCode := c.Param("id")

	var form model.DictItemForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	form.DictCode = dictCode

	if err := service.SaveDictItem(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "新增成功")
}

// GetDictItemFormByCode 字典项表单数据（RESTful）
// @Summary 字典项表单数据（RESTful）
// @Tags 字典管理
// @Param id path string true "字典编码"
// @Param itemId path int true "字典项ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/dicts/{id}/items/{itemId}/form [get]
func GetDictItemFormByCode(c *gin.Context) {
	itemIdStr := c.Param("itemId")
	itemId, err := strconv.ParseInt(itemIdStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的字典项ID")
		return
	}

	form, err := service.GetDictItemForm(itemId)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, form)
}

// UpdateDictItemByCode 修改字典项（RESTful）
// @Summary 修改字典项（RESTful）
// @Tags 字典管理
// @Param id path string true "字典编码"
// @Param itemId path int true "字典项ID"
// @Param body body model.DictItemForm true "字典项信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/dicts/{id}/items/{itemId} [put]
func UpdateDictItemByCode(c *gin.Context) {
	dictCode := c.Param("id")
	itemIdStr := c.Param("itemId")
	itemId, err := strconv.ParseInt(itemIdStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的字典项ID")
		return
	}

	var form model.DictItemForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	form.ID = types.BigInt(itemId)
	form.DictCode = dictCode

	if err := service.SaveDictItem(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "更新成功")
}

// DeleteDictItemsByCode 删除字典项（RESTful）
// @Summary 删除字典项（RESTful）
// @Tags 字典管理
// @Param id path string true "字典编码"
// @Param itemIds path string true "字典项ID（多个用逗号分隔）"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/dicts/{id}/items/{itemIds} [delete]
func DeleteDictItemsByCode(c *gin.Context) {
	itemIdsStr := c.Param("itemIds")
	if itemIdsStr == "" {
		response.BadRequest(c, "无效的字典项ID")
		return
	}

	idStrArr := strings.Split(itemIdsStr, ",")
	ids := make([]int64, 0, len(idStrArr))
	for _, idStr := range idStrArr {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "无效的字典项ID")
			return
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		response.BadRequest(c, "无效的字典项ID")
		return
	}

	if err := service.BatchDeleteDictItems(ids); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "删除成功")
}
