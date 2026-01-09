package utils

import (
	"fmt"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strings"

	"youlai-gin/pkg/errs"
)

// FileValidator 文件验证器
type FileValidator struct {
	MaxSize        int64    // 最大文件大小（字节）
	AllowedExts    []string // 允许的文件扩展名
	AllowedMimes   []string // 允许的 MIME 类型
	ForbiddenExts  []string // 禁止的文件扩展名
}

// ImageValidator 图片验证器（预设）
var ImageValidator = &FileValidator{
	MaxSize:       5 * 1024 * 1024, // 5MB
	AllowedExts:   []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"},
	AllowedMimes:  []string{"image/jpeg", "image/png", "image/gif", "image/bmp", "image/webp"},
	ForbiddenExts: []string{".exe", ".bat", ".sh", ".php", ".jsp", ".asp"},
}

// DocumentValidator 文档验证器（预设）
var DocumentValidator = &FileValidator{
	MaxSize:       10 * 1024 * 1024, // 10MB
	AllowedExts:   []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt"},
	AllowedMimes:  []string{"application/pdf", "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "text/plain"},
	ForbiddenExts: []string{".exe", ".bat", ".sh", ".php", ".jsp", ".asp", ".js", ".html", ".htm"},
}

// ExcelValidator Excel 文件验证器（预设）
var ExcelValidator = &FileValidator{
	MaxSize:       10 * 1024 * 1024, // 10MB
	AllowedExts:   []string{".xls", ".xlsx"},
	AllowedMimes:  []string{"application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
	ForbiddenExts: []string{".exe", ".bat", ".sh", ".php", ".jsp", ".asp"},
}

// Validate 验证文件
func (v *FileValidator) Validate(file *multipart.FileHeader) error {
	// 1. 检查文件大小
	if v.MaxSize > 0 && file.Size > v.MaxSize {
		return errs.BadRequest(fmt.Sprintf("文件大小超过限制，最大允许 %s", formatFileSize(v.MaxSize)))
	}

	// 2. 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))

	// 2.1 检查是否在禁止列表中
	if len(v.ForbiddenExts) > 0 {
		for _, forbidden := range v.ForbiddenExts {
			if ext == strings.ToLower(forbidden) {
				return errs.BadRequest(fmt.Sprintf("不允许上传 %s 类型的文件", ext))
			}
		}
	}

	// 2.2 检查是否在允许列表中
	if len(v.AllowedExts) > 0 {
		allowed := false
		for _, allowedExt := range v.AllowedExts {
			if ext == strings.ToLower(allowedExt) {
				allowed = true
				break
			}
		}
		if !allowed {
			return errs.BadRequest(fmt.Sprintf("不支持的文件类型 %s，允许的类型：%s", ext, strings.Join(v.AllowedExts, ", ")))
		}
	}

	// 3. 检查 MIME 类型
	if len(v.AllowedMimes) > 0 {
		contentType := strings.TrimSpace(file.Header.Get("Content-Type"))
		if contentType == "" {
			guessed := mime.TypeByExtension(ext)
			if guessed != "" {
				contentType = strings.TrimSpace(strings.SplitN(guessed, ";", 2)[0])
			}
		}
		allowed := false
		for _, allowedMime := range v.AllowedMimes {
			if strings.HasPrefix(contentType, allowedMime) {
				allowed = true
				break
			}
		}
		if !allowed {
			return errs.BadRequest(fmt.Sprintf("不支持的文件 MIME 类型：%s", contentType))
		}
	}

	// 4. 检查文件名安全性
	if err := validateFilename(file.Filename); err != nil {
		return err
	}

	return nil
}

// validateFilename 验证文件名安全性
func validateFilename(filename string) error {
	// 检查文件名长度
	if len(filename) > 255 {
		return errs.BadRequest("文件名过长，最大 255 个字符")
	}

	// 检查危险字符
	dangerousChars := []string{"../", "..\\", "<", ">", ":", "\"", "|", "?", "*", "\x00"}
	for _, char := range dangerousChars {
		if strings.Contains(filename, char) {
			return errs.BadRequest("文件名包含非法字符")
		}
	}

	// 检查是否为空
	if strings.TrimSpace(filename) == "" {
		return errs.BadRequest("文件名不能为空")
	}

	return nil
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// ValidateImage 验证图片文件（快捷方式）
func ValidateImage(file *multipart.FileHeader) error {
	return ImageValidator.Validate(file)
}

// ValidateDocument 验证文档文件（快捷方式）
func ValidateDocument(file *multipart.FileHeader) error {
	return DocumentValidator.Validate(file)
}

// ValidateExcel 验证 Excel 文件（快捷方式）
func ValidateExcel(file *multipart.FileHeader) error {
	return ExcelValidator.Validate(file)
}
