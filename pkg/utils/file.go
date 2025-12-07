package utils

import (
	"crypto/md5"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
)

// GenerateFileName 生成文件名（使用MD5 + 时间戳 + 原始扩展名）
func GenerateFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().UnixNano()
	hash := md5.Sum([]byte(fmt.Sprintf("%s-%d", originalName, timestamp)))
	return fmt.Sprintf("%x%s", hash, ext)
}

// GeneratePath 生成存储路径（按日期分类）
// prefix: 路径前缀（如：uploads, images）
// filename: 文件名
// 返回：uploads/2024/01/01/filename.jpg
func GeneratePath(prefix, filename string) string {
	now := time.Now()
	return fmt.Sprintf("%s/%d/%02d/%02d/%s",
		prefix,
		now.Year(),
		now.Month(),
		now.Day(),
		filename,
	)
}

// ValidateFile 验证上传文件
func ValidateFile(file *multipart.FileHeader, maxSize int64, allowedExts []string) error {
	// 验证文件大小
	if maxSize > 0 && file.Size > maxSize {
		return fmt.Errorf("文件大小超过限制：%s（最大：%s）",
			FormatFileSize(file.Size),
			FormatFileSize(maxSize),
		)
	}
	
	// 验证文件扩展名
	if len(allowedExts) > 0 {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		ext = strings.TrimPrefix(ext, ".")
		
		allowed := false
		for _, allowedExt := range allowedExts {
			if ext == strings.ToLower(allowedExt) {
				allowed = true
				break
			}
		}
		
		if !allowed {
			return fmt.Errorf("不允许的文件类型：%s（允许：%s）", ext, strings.Join(allowedExts, ", "))
		}
	}
	
	return nil
}

// FormatFileSize 格式化文件大小
func FormatFileSize(size int64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
	)
	
	s := float64(size)
	switch {
	case s >= TB:
		return fmt.Sprintf("%.2f TB", s/TB)
	case s >= GB:
		return fmt.Sprintf("%.2f GB", s/GB)
	case s >= MB:
		return fmt.Sprintf("%.2f MB", s/MB)
	case s >= KB:
		return fmt.Sprintf("%.2f KB", s/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// GetContentType 根据扩展名获取Content-Type
func GetContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	contentTypes := map[string]string{
		// 图片
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		
		// 文档
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		
		// 视频
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".wmv":  "video/x-ms-wmv",
		".flv":  "video/x-flv",
		
		// 音频
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".ogg":  "audio/ogg",
		
		// 压缩包
		".zip":  "application/zip",
		".rar":  "application/x-rar-compressed",
		".7z":   "application/x-7z-compressed",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
		
		// 文本
		".txt":  "text/plain",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
	}
	
	if ct, ok := contentTypes[ext]; ok {
		return ct
	}
	
	return "application/octet-stream"
}
