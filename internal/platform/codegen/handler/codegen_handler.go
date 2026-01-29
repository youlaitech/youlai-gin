package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"youlai-gin/internal/platform/codegen/model"
	"youlai-gin/internal/platform/codegen/service"
	"youlai-gin/pkg/response"
)

// GetTablePage 数据表分页
// @Summary 数据表分页
// @Tags 13.代码生成
// @Router /api/v1/codegen/table [get]
func GetTablePage(c *gin.Context) {
	var query model.TableQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	result, err := service.GetTablePage(&query)
	if err != nil {
		c.Error(err)
		return
	}

	response.OkPaged(c, result)
}

// GetGenConfig 获取生成配置
// @Summary 获取生成配置
// @Tags 13.代码生成
// @Param tableName path string true "表名"
// @Router /api/v1/codegen/{tableName}/config [get]
func GetGenConfig(c *gin.Context) {
	tableName := c.Param("tableName")
	result, err := service.GetGenConfig(tableName)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, result)
}

// SaveGenConfig 保存生成配置
// @Summary 保存生成配置
// @Tags 13.代码生成
// @Param tableName path string true "表名"
// @Router /api/v1/codegen/{tableName}/config [post]
func SaveGenConfig(c *gin.Context) {
	tableName := c.Param("tableName")
	var body model.GenConfigFormDto
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	if err := service.SaveGenConfig(tableName, &body); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "保存成功")
}

// DeleteGenConfig 删除生成配置
// @Summary 删除生成配置
// @Tags 13.代码生成
// @Param tableName path string true "表名"
// @Router /api/v1/codegen/{tableName}/config [delete]
func DeleteGenConfig(c *gin.Context) {
	tableName := c.Param("tableName")
	if err := service.DeleteGenConfig(tableName); err != nil {
		c.Error(err)
		return
	}
	response.OkMsg(c, "删除成功")
}

// GetPreview 预览代码
// @Summary 预览代码
// @Tags 13.代码生成
// @Param tableName path string true "表名"
// @Router /api/v1/codegen/{tableName}/preview [get]
func GetPreview(c *gin.Context) {
	tableName := c.Param("tableName")
	pageType := c.DefaultQuery("pageType", "classic")
	typeParam := c.DefaultQuery("type", "ts")

	list, err := service.GetPreview(tableName, pageType, typeParam)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, list)
}

// Download 下载代码
// @Summary 下载代码
// @Tags 13.代码生成
// @Param tableName path string true "表名（可逗号分隔）"
// @Router /api/v1/codegen/{tableName}/download [get]
func Download(c *gin.Context) {
	tableName := c.Param("tableName")
	pageType := c.DefaultQuery("pageType", "classic")
	typeParam := c.DefaultQuery("type", "ts")

	tableNames := make([]string, 0)
	for _, t := range strings.Split(tableName, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			tableNames = append(tableNames, t)
		}
	}
	if len(tableNames) == 0 {
		response.Fail(c, "参数错误")
		return
	}

	fileName, data, err := service.DownloadZip(tableNames, pageType, typeParam)
	if err != nil {
		c.Error(err)
		return
	}

	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	c.Data(http.StatusOK, "application/octet-stream", data)
}
