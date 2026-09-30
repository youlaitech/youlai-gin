package handler

import (
	"github.com/gin-gonic/gin"

	response "youlai-gin/internal/common"
	"youlai-gin/internal/file/service"
	"youlai-gin/pkg/errs"
)

// FileHandler 文件管理接口层
type FileHandler struct {
	svc *service.FileService
}

// NewFileHandler 创建 FileHandler 实例
func NewFileHandler(svc *service.FileService) *FileHandler { return &FileHandler{svc: svc} }

// RegisterRoutes 注册文件管理路由，基础路径 /api/v1/files
func (h *FileHandler) RegisterRoutes(r *gin.RouterGroup) {
	fileGroup := r.Group("/files")
	fileGroup.POST("", h.UploadFile)   // 单文件上传
	fileGroup.DELETE("", h.DeleteFile) // 删除文件
}

// UploadFile 单文件上传
// @Summary 文件上传
// @Tags 10.文件接口
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件"
// @Param path formData string false "存储路径前缀"
// @Success 200 {object} response.Result{data=service.UploadResult}
// @Router /api/v1/files [post]
func (h *FileHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.Error(errs.BadRequest("未选择文件"))
		return
	}

	result, err := h.svc.Upload(file, c.DefaultPostForm("path", "uploads"))
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, result)
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Tags 10.文件接口
// @Produce json
// @Param path query string true "文件路径"
// @Success 200 {object} response.Result
// @Router /api/v1/files [delete]
func (h *FileHandler) DeleteFile(c *gin.Context) {
	if err := h.svc.Delete(c.Query("path")); err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, nil)
}
