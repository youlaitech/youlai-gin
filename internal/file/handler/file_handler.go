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
	fileGroup.POST("", h.UploadFile)       // 单文件上传
	fileGroup.POST("/batch", h.UploadFiles) // 批量上传
	fileGroup.POST("/image", h.UploadImage) // 图片上传
	fileGroup.DELETE("", h.DeleteFile)      // 删除文件
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

// UploadFiles 批量文件上传
// @Summary 批量上传
// @Tags 10.文件接口
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "文件列表"
// @Param path formData string false "存储路径前缀"
// @Success 200 {object} response.Result{data=[]service.UploadResult}
// @Router /api/v1/files/batch [post]
func (h *FileHandler) UploadFiles(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.Error(errs.BadRequest("未选择文件"))
		return
	}
	if len(form.File["files"]) == 0 {
		c.Error(errs.BadRequest("未选择文件"))
		return
	}

	pathPrefix := c.DefaultPostForm("path", "uploads")
	results := make([]*service.UploadResult, 0, len(form.File["files"]))
	for _, file := range form.File["files"] {
		// 跳过校验或上传失败的文件（批量接口幂等，不中断整体）
		if result, err := h.svc.Upload(file, pathPrefix); err == nil {
			results = append(results, result)
		}
	}

	response.Ok(c, results)
}

// UploadImage 图片上传（带限制）
// @Summary 图片上传
// @Tags 10.文件接口
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "图片文件"
// @Success 200 {object} response.Result{data=service.UploadResult}
// @Router /api/v1/files/image [post]
func (h *FileHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.Error(errs.BadRequest("未选择文件"))
		return
	}

	result, err := h.svc.UploadImage(file)
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