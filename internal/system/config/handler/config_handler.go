package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"youlai-gin/internal/system/config/model"
	"youlai-gin/internal/system/config/service"
	"youlai-gin/pkg/response"
)

// RegisterRoutes 注册配置管理路由
func RegisterRoutes(r *gin.RouterGroup) {
	// 使用复数形式
	config := r.Group("/configs")
	{
		config.GET("", GetConfigPage)
		config.GET("/:id/form", GetConfigForm)
		config.GET("/:id", GetConfigByID)
		config.GET("/key/:key", GetConfigByKey)
		config.POST("", SaveConfig)
		config.PUT("/:id", UpdateConfig)
		config.DELETE("/:ids", DeleteConfigs)
		config.POST("/refresh/:key", RefreshConfigCache)
		config.POST("/refresh", RefreshAllConfigCache)
	}
}

// GetConfigPage 获取配置分页列表
// @Summary 配置分页
// @Tags 08.系统配置
// @Router /api/v1/configs [get]
func GetConfigPage(c *gin.Context) {
	var query model.ConfigQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	result, err := service.GetConfigPage(&query)
	if err != nil {
		c.Error(err)
		return
	}

	response.OkPaged(c, result)
}

// GetConfigForm 获取配置表单数据
// @Summary 配置表单
// @Tags 08.系统配置
// @Param id path int true "配置ID"
// @Router /api/v1/configs/{id}/form [get]
func GetConfigForm(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Fail(c, "ID格式错误")
		return
	}

	formData, err := service.GetConfigFormData(id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, formData)
}

// GetConfigByID 根据ID获取配置
// @Summary 配置详情
// @Tags 08.系统配置
// @Param id path int true "配置ID"
// @Router /api/v1/configs/{id} [get]
func GetConfigByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Fail(c, "ID格式错误")
		return
	}

	config, err := service.GetConfigByID(id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, config)
}

// GetConfigByKey 根据Key获取配置
// @Summary 根据键获取配置
// @Tags 08.系统配置
// @Param key path string true "配置键"
// @Router /api/v1/configs/key/{key} [get]
func GetConfigByKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.Fail(c, "配置Key不能为空")
		return
	}

	config, err := service.GetConfigByKey(key)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, config)
}

// SaveConfig 保存配置（新增）
// @Summary 新增配置
// @Tags 08.系统配置
// @Router /api/v1/configs [post]
func SaveConfig(c *gin.Context) {
	var form model.ConfigForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	if err := service.SaveConfig(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "保存成功")
}

// UpdateConfig 更新配置
// @Summary 更新配置
// @Tags 08.系统配置
// @Param id path int true "配置ID"
// @Router /api/v1/configs/{id} [put]
func UpdateConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Fail(c, "ID格式错误")
		return
	}

	var form model.ConfigForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	form.ID = id
	if err := service.SaveConfig(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "更新成功")
}

// DeleteConfigs 删除配置（支持批量）
// @Summary 删除配置
// @Tags 08.系统配置
// @Param ids path string true "配置ID列表"
// @Router /api/v1/configs/{ids} [delete]
func DeleteConfigs(c *gin.Context) {
	idsStr := c.Param("ids")
	if idsStr == "" {
		response.Fail(c, "ID不能为空")
		return
	}

	// 支持批量删除，ids格式：1,2,3
	idStrArr := strings.Split(idsStr, ",")
	if len(idStrArr) == 1 {
		// 单个删除
		id, err := strconv.ParseInt(idsStr, 10, 64)
		if err != nil {
			response.Fail(c, "ID格式错误")
			return
		}
		if err := service.DeleteConfig(id); err != nil {
			c.Error(err)
			return
		}
	} else {
		// 批量删除
		ids := make([]int64, 0, len(idStrArr))
		for _, idStr := range idStrArr {
			id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
			if err != nil {
				response.Fail(c, "ID格式错误")
				return
			}
			ids = append(ids, id)
		}
		if err := service.BatchDeleteConfig(ids); err != nil {
			c.Error(err)
			return
		}
	}

	response.OkMsg(c, "删除成功")
}

// RefreshConfigCache 刷新指定配置缓存
// @Summary 刷新配置缓存
// @Tags 08.系统配置
// @Param key path string true "配置键"
// @Router /api/v1/configs/refresh/{key} [post]
func RefreshConfigCache(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.Fail(c, "配置Key不能为空")
		return
	}

	if err := service.RefreshConfigCache(key); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "刷新成功")
}

// RefreshAllConfigCache 刷新所有配置缓存
// @Summary 刷新全部配置缓存
// @Tags 08.系统配置
// @Router /api/v1/configs/refresh [post]
func RefreshAllConfigCache(c *gin.Context) {
	service.ClearAllConfigCache()
	response.OkMsg(c, "刷新成功")
}
