package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioStorage MinIO 对象存储（S3 协议兼容）
type MinioStorage struct {
	client           *minio.Client
	bucket           string
	endpoint         string // 原始地址（含 scheme）
	domain           string // 自定义域名（优先于 endpoint 拼接访问 URL）
	pathNoBucketName bool   // 返回url中不包含存储桶名称（兼容Cloudflare R2）
}

// NewMinioStorage 创建 MinIO 客户端，并校验/创建存储桶（含公共读策略）
func NewMinioStorage(config *Config) (*MinioStorage, error) {
	u, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("解析 MinIO 地址失败: %w", err)
	}
	useSSL := strings.EqualFold(u.Scheme, "https")

	client, err := minio.New(u.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: useSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 MinIO 客户端失败: %w", err)
	}

	s := &MinioStorage{
		client:           client,
		bucket:           config.Bucket,
		endpoint:         config.Endpoint,
		domain:           config.Domain,
		pathNoBucketName: config.PathNoBucketName,
	}
	if err := s.ensureBucket(); err != nil {
		return nil, err
	}
	return s, nil
}

// ensureBucket 存储桶不存在则创建，并赋予公共读策略（使直链可匿名访问）
func (s *MinioStorage) ensureBucket() error {
	ctx := context.Background()
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("检查存储桶失败: %w", err)
	}
	if exists {
		return nil
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("创建存储桶失败: %w", err)
	}
	policy := fmt.Sprintf(
		`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`,
		s.bucket,
	)
	if err := s.client.SetBucketPolicy(ctx, s.bucket, policy); err != nil {
		return fmt.Errorf("设置存储桶公共读策略失败: %w", err)
	}
	return nil
}

// Upload 上传文件
func (s *MinioStorage) Upload(path string, file io.Reader, contentType string) (string, error) {
	return s.UploadWithOptions(path, file, &UploadOptions{ContentType: contentType})
}

// UploadWithOptions 带选项的上传
func (s *MinioStorage) UploadWithOptions(path string, file io.Reader, opts *UploadOptions) (string, error) {
	putOpts := minio.PutObjectOptions{}
	if opts != nil {
		if opts.ContentType != "" {
			putOpts.ContentType = opts.ContentType
		}
		if opts.CacheControl != "" {
			putOpts.CacheControl = opts.CacheControl
		}
	}
	if _, err := s.client.PutObject(context.Background(), s.bucket, path, file, -1, putOpts); err != nil {
		return "", fmt.Errorf("上传文件失败: %w", err)
	}
	return s.GetURL(path, 0)
}

// Delete 删除文件
func (s *MinioStorage) Delete(path string) error {
	if err := s.client.RemoveObject(context.Background(), s.bucket, path, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

// GetURL 获取文件访问 URL（domain 优先，否则用 endpoint）
func (s *MinioStorage) GetURL(path string, expires time.Duration) (string, error) {
	base := s.endpoint
	if s.domain != "" {
		base = s.domain
	}
	base = strings.TrimRight(base, "/")
	if s.pathNoBucketName {
		return fmt.Sprintf("%s/%s", base, path), nil
	}
	return fmt.Sprintf("%s/%s/%s", base, s.bucket, path), nil
}

// Exists 检查文件是否存在
func (s *MinioStorage) Exists(path string) (bool, error) {
	_, err := s.client.StatObject(context.Background(), s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetInfo 获取文件信息
func (s *MinioStorage) GetInfo(path string) (*FileInfo, error) {
	info, err := s.client.StatObject(context.Background(), s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	url, _ := s.GetURL(path, 0)
	return &FileInfo{
		Path:         path,
		Size:         info.Size,
		ContentType:  info.ContentType,
		LastModified: info.LastModified,
		ETag:         info.ETag,
		URL:          url,
	}, nil
}
