package storage

import (
	"io"
	"time"
)

// Storage 对象存储接口（支持多种云服务商）
type Storage interface {
	// Upload 上传文件
	// path: 存储路径（如：uploads/2024/01/01/file.jpg）
	// file: 文件内容
	// contentType: 文件MIME类型
	Upload(path string, file io.Reader, contentType string) (string, error)
	
	// UploadWithOptions 带选项的上传
	UploadWithOptions(path string, file io.Reader, opts *UploadOptions) (string, error)
	
	// Delete 删除文件
	Delete(path string) error
	
	// GetURL 获取文件访问URL
	// path: 存储路径
	// expires: 过期时间（0表示永久，>0表示生成临时签名URL）
	GetURL(path string, expires time.Duration) (string, error)
	
	// Exists 检查文件是否存在
	Exists(path string) (bool, error)
	
	// GetInfo 获取文件信息
	GetInfo(path string) (*FileInfo, error)
}

// UploadOptions 上传选项
type UploadOptions struct {
	ContentType   string            // MIME类型
	CacheControl  string            // 缓存控制
	ContentDisposition string       // 内容处置
	Metadata      map[string]string // 自定义元数据
	ACL           string            // 访问控制（public-read, private等）
}

// FileInfo 文件信息
type FileInfo struct {
	Path         string    // 文件路径
	Size         int64     // 文件大小（字节）
	ContentType  string    // MIME类型
	LastModified time.Time // 最后修改时间
	ETag         string    // ETag
	URL          string    // 访问URL
}

// Config 存储配置
type Config struct {
	Type      string `mapstructure:"type"`      // 存储类型：local, aliyun
	Endpoint  string `mapstructure:"endpoint"`  // 端点地址
	Bucket    string `mapstructure:"bucket"`    // 存储桶名称
	AccessKey string `mapstructure:"accessKey"` // 访问密钥ID
	SecretKey string `mapstructure:"secretKey"` // 访问密钥Secret
	Region    string `mapstructure:"region"`    // 区域
	Domain    string `mapstructure:"domain"`    // 自定义域名（CDN）
	IsPrivate bool   `mapstructure:"isPrivate"` // 是否私有（影响URL生成）
	BasePath  string `mapstructure:"basePath"`  // 基础路径
}

// StorageType 存储类型常量
const (
	TypeLocal  = "local"
	TypeAliyun = "aliyun"
)
