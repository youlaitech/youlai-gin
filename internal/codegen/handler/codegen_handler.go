package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"youlai-gin/internal/codegen/model"
	"youlai-gin/internal/codegen/service"
	response "youlai-gin/internal/common"
	"youlai-gin/internal/common/validator"
	"youlai-gin/pkg/errs"
)

// CodegenHandler 代码生成接口层（数据表 / 配置 / 预览 / 下载）
type CodegenHandler struct {
	svc *service.CodegenService
}

// NewCodegenHandler 创建 CodegenHandler 实例
func NewCodegenHandler(svc *service.CodegenService) *CodegenHandler { return &CodegenHandler{svc: svc} }

// RegisterRoutes 注册代码生成路由，基础路径 /api/v1/codegen
func (h *CodegenHandler) RegisterRoutes(r *gin.RouterGroup) {
	codegenGroup := r.Group("/codegen")
	codegenGroup.GET("/table", h.GetTablePage)
	codegenGroup.GET("/:tableName/config", h.GetGenConfig)
	codegenGroup.POST("/:tableName/config", h.SaveGenConfig)
	codegenGroup.DELETE("/:tableName/config", h.DeleteGenConfig)
	codegenGroup.GET("/:tableName/preview", h.GetPreview)
	codegenGroup.GET("/:tableName/download", h.Download)
}

// GetTablePage 数据表分页
// @Summary 数据表分页
// @Tags 11.代码生成
// @Router /api/v1/codegen/table [get]
func (h *CodegenHandler) GetTablePage(c *gin.Context) {
	var query model.TableQuery
	if err := validator.BindQuery(c, &query); err != nil {
		c.Error(err)
		return
	}

	result, err := h.svc.GetTablePage(c.Request.Context(), &query)
	if err != nil {
		c.Error(err)
		return
	}

	response.OkPaged(c, result)
}

// GetGenConfig 获取生成配置
// @Summary 获取生成配置
// @Tags 11.代码生成
// @Param tableName path string true "表名"
// @Router /api/v1/codegen/{tableName}/config [get]
func (h *CodegenHandler) GetGenConfig(c *gin.Context) {
	tableName := c.Param("tableName")
	result, err := h.svc.GetGenConfig(c.Request.Context(), tableName)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, result)
}

// SaveGenConfig 保存生成配置
// @Summary 保存生成配置
// @Tags 11.代码生成
// @Param tableName path string true "表名"
// @Router /api/v1/codegen/{tableName}/config [post]
func (h *CodegenHandler) SaveGenConfig(c *gin.Context) {
	tableName := c.Param("tableName")
	var body model.GenConfigForm
	if err := validator.BindJSON(c, &body); err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.SaveGenConfig(c.Request.Context(), tableName, &body); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "保存成功")
}

// DeleteGenConfig 删除生成配置
// @Summary 删除生成配置
// @Tags 11.代码生成
// @Param tableName path string true "表名"
// @Router /api/v1/codegen/{tableName}/config [delete]
func (h *CodegenHandler) DeleteGenConfig(c *gin.Context) {
	tableName := c.Param("tableName")
	if err := h.svc.DeleteGenConfig(c.Request.Context(), tableName); err != nil {
		c.Error(err)
		return
	}
	response.OkMsg(c, "删除成功")
}

// GetPreview 预览代码
// @Summary 预览代码
// @Tags 11.代码生成
// @Param tableName path string true "表名"
// @Router /api/v1/codegen/{tableName}/preview [get]
func (h *CodegenHandler) GetPreview(c *gin.Context) {
	tableName := c.Param("tableName")
	pageType := c.DefaultQuery("pageType", "classic")
	typeParam := c.DefaultQuery("type", "ts")

	list, err := h.svc.GetPreview(c.Request.Context(), tableName, pageType, typeParam)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, list)
}

// Download 下载代码
// @Summary 下载代码
// @Tags 11.代码生成
// @Param tableName path string true "表名（可逗号分隔）"
// @Router /api/v1/codegen/{tableName}/download [get]
func (h *CodegenHandler) Download(c *gin.Context) {
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
		c.Error(errs.BadRequest("请选择要生成的表"))
		return
	}

	fileName, data, err := h.svc.DownloadZip(c.Request.Context(), tableNames, pageType, typeParam)
	if err != nil {
		c.Error(err)
		return
	}

	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	c.Data(http.StatusOK, "application/octet-stream", data)
}