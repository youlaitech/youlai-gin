package storage

import (
	"fmt"
	"io"
	"time"
	
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// AliyunOSS 阿里云OSS存储
type AliyunOSS struct {
	client    *oss.Client
	bucket    *oss.Bucket
	bucketName string
	endpoint   string
	domain     string    // 自定义域名（CDN）
	isPrivate  bool      // 是否私有
}

// NewAliyunOSS 创建阿里云OSS存储
// 使用前需要安装: go get github.com/aliyun/aliyun-oss-go-sdk/oss
func NewAliyunOSS(config *Config) (*AliyunOSS, error) {
	client, err := oss.New(config.Endpoint, config.AccessKey, config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("创建OSS客户端失败: %w", err)
	}
	
	bucket, err := client.Bucket(config.Bucket)
	if err != nil {
		return nil, fmt.Errorf("获取Bucket失败: %w", err)
	}
	
	domain := config.Domain
	if domain == "" {
		// 默认域名：<bucket>.<endpoint>
		domain = fmt.Sprintf("https://%s.%s", config.Bucket, config.Endpoint)
	}
	
	return &AliyunOSS{
		client:     client,
		bucket:     bucket,
		bucketName: config.Bucket,
		endpoint:   config.Endpoint,
		domain:     domain,
		isPrivate:  config.IsPrivate,
	}, nil
}

// Upload 上传文件
func (s *AliyunOSS) Upload(path string, file io.Reader, contentType string) (string, error) {
	return s.UploadWithOptions(path, file, &UploadOptions{
		ContentType: contentType,
	})
}

// UploadWithOptions 带选项的上传
func (s *AliyunOSS) UploadWithOptions(path string, file io.Reader, opts *UploadOptions) (string, error) {
	var options []oss.Option
	
	if opts.ContentType != "" {
		options = append(options, oss.ContentType(opts.ContentType))
	}
	
	if opts.CacheControl != "" {
		options = append(options, oss.CacheControl(opts.CacheControl))
	}
	
	if opts.ContentDisposition != "" {
		options = append(options, oss.ContentDisposition(opts.ContentDisposition))
	}
	
	// 设置ACL
	if opts.ACL != "" {
		var aclType oss.ACLType
		switch opts.ACL {
		case "public-read":
			aclType = oss.ACLPublicRead
		case "private":
			aclType = oss.ACLPrivate
		default:
			aclType = oss.ACLDefault
		}
		options = append(options, oss.ObjectACL(aclType))
	}
	
	// 上传文件
	if err := s.bucket.PutObject(path, file, options...); err != nil {
		return "", fmt.Errorf("上传文件失败: %w", err)
	}
	
	// 返回访问URL
	return s.GetURL(path, 0)
}

// Delete 删除文件
func (s *AliyunOSS) Delete(path string) error {
	return s.bucket.DeleteObject(path)
}

// GetURL 获取文件访问URL
func (s *AliyunOSS) GetURL(path string, expires time.Duration) (string, error) {
	// 私有文件生成签名URL
	if s.isPrivate && expires > 0 {
		return s.bucket.SignURL(path, oss.HTTPGet, int64(expires.Seconds()))
	}
	
	// 公开文件直接返回URL
	return fmt.Sprintf("%s/%s", s.domain, path), nil
}

// Exists 检查文件是否存在
func (s *AliyunOSS) Exists(path string) (bool, error) {
	return s.bucket.IsObjectExist(path)
}

// GetInfo 获取文件信息
func (s *AliyunOSS) GetInfo(path string) (*FileInfo, error) {
	meta, err := s.bucket.GetObjectMeta(path)
	if err != nil {
		return nil, err
	}
	
	url, _ := s.GetURL(path, 0)
	
	var size int64
	if sizeStr := meta.Get("Content-Length"); sizeStr != "" {
		fmt.Sscanf(sizeStr, "%d", &size)
	}
	
	var lastModified time.Time
	if modStr := meta.Get("Last-Modified"); modStr != "" {
		lastModified, _ = time.Parse(time.RFC1123, modStr)
	}
	
	return &FileInfo{
		Path:         path,
		Size:         size,
		ContentType:  meta.Get("Content-Type"),
		LastModified: lastModified,
		ETag:         meta.Get("ETag"),
		URL:          url,
	}, nil
}
