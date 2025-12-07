package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LocalStorage 本地存储
type LocalStorage struct {
	basePath string // 基础路径（如：./uploads）
	baseURL  string // 访问基础URL（如：http://localhost:8080/uploads）
}

// NewLocalStorage 创建本地存储
func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	// 确保目录存在
	os.MkdirAll(basePath, 0755)
	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}
}

// Upload 上传文件
func (s *LocalStorage) Upload(path string, file io.Reader, contentType string) (string, error) {
	return s.UploadWithOptions(path, file, &UploadOptions{
		ContentType: contentType,
	})
}

// UploadWithOptions 带选项的上传
func (s *LocalStorage) UploadWithOptions(path string, file io.Reader, opts *UploadOptions) (string, error) {
	fullPath := filepath.Join(s.basePath, path)
	
	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}
	
	// 创建文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()
	
	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}
	
	// 返回访问URL
	return s.GetURL(path, 0)
}

// Delete 删除文件
func (s *LocalStorage) Delete(path string) error {
	fullPath := filepath.Join(s.basePath, path)
	return os.Remove(fullPath)
}

// GetURL 获取文件访问URL
func (s *LocalStorage) GetURL(path string, expires time.Duration) (string, error) {
	// 本地存储不支持签名URL，直接返回公开URL
	return fmt.Sprintf("%s/%s", s.baseURL, path), nil
}

// Exists 检查文件是否存在
func (s *LocalStorage) Exists(path string) (bool, error) {
	fullPath := filepath.Join(s.basePath, path)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// GetInfo 获取文件信息
func (s *LocalStorage) GetInfo(path string) (*FileInfo, error) {
	fullPath := filepath.Join(s.basePath, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	
	url, _ := s.GetURL(path, 0)
	
	return &FileInfo{
		Path:         path,
		Size:         info.Size(),
		LastModified: info.ModTime(),
		URL:          url,
	}, nil
}
