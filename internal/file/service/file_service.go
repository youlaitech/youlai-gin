package service

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"youlai-gin/internal/common/config"
	"youlai-gin/internal/common/storage"
	"youlai-gin/internal/common/utils"
	"youlai-gin/pkg/errs"
)

// UploadResult 上传结果
type UploadResult struct {
	Name string `json:"name"` // 原始文件名
	URL  string `json:"url"`  // 访问URL
	Path string `json:"path"` // 存储路径
	Size int64  `json:"size"` // 文件大小
}

// FileService 文件上传/删除业务逻辑层（封装存储适配与参数校验）
type FileService struct{}

// NewFileService 创建 FileService 实例
func NewFileService() *FileService { return &FileService{} }

// Upload 单文件上传到指定路径前缀，失败统一转换为规范的 API 错误
func (s *FileService) Upload(file *multipart.FileHeader, pathPrefix string) (*UploadResult, error) {
	if pathPrefix == "" {
		pathPrefix = "uploads"
	}
	if err := validateUpload(file); err != nil {
		return nil, errs.BadRequest(err.Error())
	}

	result, err := doUpload(file, pathPrefix)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UploadImage 上传图片（复用严格图片校验，固定 images 路径前缀）
func (s *FileService) UploadImage(file *multipart.FileHeader) (*UploadResult, error) {
	if err := utils.ValidateImage(file); err != nil {
		return nil, errs.BadRequest(err.Error())
	}
	return doUpload(file, "images")
}

// Delete 删除存储中指定路径的文件
func (s *FileService) Delete(path string) error {
	if path == "" {
		return errs.BadRequest("文件路径不能为空")
	}
	if err := storage.DefaultStorage.Delete(path); err != nil {
		return errs.SystemError("删除失败")
	}
	return nil
}

// doUpload 统一打开文件并写入存储
func doUpload(file *multipart.FileHeader, pathPrefix string) (*UploadResult, error) {
	filename := utils.GenerateFileName(file.Filename)
	path := utils.GeneratePath(pathPrefix, filename)

	src, err := file.Open()
	if err != nil {
		return nil, errs.SystemError("打开文件失败")
	}
	defer src.Close()

	contentType := utils.GetContentType(file.Filename)
	url, err := storage.DefaultStorage.Upload(path, src, contentType)
	if err != nil {
		return nil, errs.SystemError("上传失败")
	}

	return &UploadResult{
		Name: file.Filename,
		URL:  url,
		Path: path,
		Size: file.Size,
	}, nil
}

// validateUpload 按 storage 配置校验上传文件（大小上限 + 扩展名白名单）
func validateUpload(file *multipart.FileHeader) error {
	upload := config.Cfg.FileStorage.Upload

	if max := upload.MaxFileSizeBytes(); max > 0 && file.Size > max {
		return fmt.Errorf("文件大小超过限制，最大允许 %s", utils.FormatFileSize(max))
	}

	if len(upload.AllowedExtensions) > 0 {
		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Filename), "."))
		allowed := false
		for _, e := range upload.AllowedExtensions {
			if strings.EqualFold(ext, e) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("不支持的文件类型：.%s", ext)
		}
	}
	return nil
}
