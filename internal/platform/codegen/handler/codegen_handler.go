package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"youlai-gin/internal/platform/codegen/model"
	"youlai-gin/internal/platform/codegen/service"
	"youlai-gin/pkg/response"
)

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

func GetGenConfig(c *gin.Context) {
	tableName := c.Param("tableName")
	result, err := service.GetGenConfig(tableName)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, result)
}

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

func DeleteGenConfig(c *gin.Context) {
	tableName := c.Param("tableName")
	if err := service.DeleteGenConfig(tableName); err != nil {
		c.Error(err)
		return
	}
	response.OkMsg(c, "删除成功")
}

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
