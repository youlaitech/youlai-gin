package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage S3 对象存储（通过 RustFS 等 S3 兼容服务实现）
type S3Storage struct {
	client           *s3.Client
	bucket           string
	endpoint         string // 原始地址（含 scheme）
	domain           string // 自定义域名（优先于 endpoint 拼接访问 URL）
	pathNoBucketName bool   // 返回url中不包含存储桶名称（兼容Cloudflare R2）
}

// NewS3Storage 创建 S3 客户端，并校验/创建存储桶（含公共读策略）
func NewS3Storage(config *Config) (*S3Storage, error) {
	ctx := context.Background()
	region := config.Region
	if region == "" {
		region = "us-east-1"
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(config.AccessKey, config.SecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 S3 客户端失败: %w", err)
	}
	// 指向自托管 S3 兼容服务（RustFS），使用路径风格寻址
	awsCfg.BaseEndpoint = aws.String(config.Endpoint)
	usePathStyle := true
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = usePathStyle
	})

	s := &S3Storage{
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
func (s *S3Storage) ensureBucket() error {
	ctx := context.Background()
	exists, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(s.bucket)})
	if err != nil && exists == nil && !strings.Contains(err.Error(), "NotFound") {
		// HeadBucket 对不存在的桶可能返回错误码，下面用 ListObjectsV2 兜底判断
	}
	_, err = s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(s.bucket)})
	if err == nil {
		return nil
	}
	if _, err := s.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(s.bucket)}); err != nil {
		return fmt.Errorf("创建存储桶失败: %w", err)
	}
	policy := fmt.Sprintf(
		`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`,
		s.bucket,
	)
	if _, err := s.client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(s.bucket),
		Policy: aws.String(policy),
	}); err != nil {
		return fmt.Errorf("设置存储桶公共读策略失败: %w", err)
	}
	return nil
}

// Upload 上传文件
func (s *S3Storage) Upload(path string, file io.Reader, contentType string) (string, error) {
	return s.UploadWithOptions(path, file, &UploadOptions{ContentType: contentType})
}

// UploadWithOptions 带选项的上传
func (s *S3Storage) UploadWithOptions(path string, file io.Reader, opts *UploadOptions) (string, error) {
	putOpts := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
		Body:   file,
	}
	if opts != nil {
		if opts.ContentType != "" {
			putOpts.ContentType = aws.String(opts.ContentType)
		}
		if opts.CacheControl != "" {
			putOpts.CacheControl = aws.String(opts.CacheControl)
		}
	}
	if _, err := s.client.PutObject(context.Background(), putOpts); err != nil {
		return "", fmt.Errorf("上传文件失败: %w", err)
	}
	return s.GetURL(path, 0)
}

// Delete 删除文件
func (s *S3Storage) Delete(path string) error {
	if _, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

// GetURL 获取文件访问 URL（domain 优先，否则用 endpoint）
func (s *S3Storage) GetURL(path string, expires time.Duration) (string, error) {
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
func (s *S3Storage) Exists(path string) (bool, error) {
	_, err := s.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetInfo 获取文件信息
func (s *S3Storage) GetInfo(path string) (*FileInfo, error) {
	out, err := s.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, err
	}
	url, _ := s.GetURL(path, 0)
	info := &FileInfo{
		Path:         path,
		ContentType:  aws.ToString(out.ContentType),
		LastModified: aws.ToTime(out.LastModified),
		ETag:         aws.ToString(out.ETag),
		URL:          url,
	}
	if out.ContentLength != nil {
		info.Size = *out.ContentLength
	}
	return info, nil
}
