package utils

import (
	"mime/multipart"
	"testing"
)

// TestValidateFilename 测试文件名验证
func TestValidateFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "正常文件名",
			filename: "test.jpg",
			wantErr:  false,
		},
		{
			name:     "中文文件名",
			filename: "测试文件.pdf",
			wantErr:  false,
		},
		{
			name:     "包含路径遍历",
			filename: "../etc/passwd",
			wantErr:  true,
		},
		{
			name:     "包含特殊字符",
			filename: "test<>.jpg",
			wantErr:  true,
		},
		{
			name:     "文件名过长",
			filename: string(make([]byte, 300)),
			wantErr:  true,
		},
		{
			name:     "空文件名",
			filename: "",
			wantErr:  true,
		},
		{
			name:     "只有空格",
			filename: "   ",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFilename(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFilename() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFormatFileSize 测试文件大小格式化
func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{
			name:     "字节",
			size:     100,
			expected: "100 B",
		},
		{
			name:     "KB",
			size:     1024,
			expected: "1.0 KB",
		},
		{
			name:     "MB",
			size:     1024 * 1024,
			expected: "1.0 MB",
		},
		{
			name:     "GB",
			size:     1024 * 1024 * 1024,
			expected: "1.0 GB",
		},
		{
			name:     "5MB",
			size:     5 * 1024 * 1024,
			expected: "5.0 MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFileSize(tt.size)
			if result != tt.expected {
				t.Errorf("formatFileSize() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestImageValidator 测试图片验证器
func TestImageValidator(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		size     int64
		wantErr  bool
	}{
		{
			name:     "正常 JPG",
			filename: "test.jpg",
			size:     1024 * 1024, // 1MB
			wantErr:  false,
		},
		{
			name:     "正常 PNG",
			filename: "test.png",
			size:     2 * 1024 * 1024, // 2MB
			wantErr:  false,
		},
		{
			name:     "文件过大",
			filename: "test.jpg",
			size:     10 * 1024 * 1024, // 10MB > 5MB 限制
			wantErr:  true,
		},
		{
			name:     "不支持的格式",
			filename: "test.exe",
			size:     1024,
			wantErr:  true,
		},
		{
			name:     "禁止的格式",
			filename: "test.php",
			size:     1024,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟的文件头
			file := &multipart.FileHeader{
				Filename: tt.filename,
				Size:     tt.size,
			}

			err := ImageValidator.Validate(file)
			if (err != nil) != tt.wantErr {
				t.Errorf("ImageValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// 运行测试：
// go test -v ./pkg/utils/ -run TestImageValidator
//
// 运行所有测试：
// go test -v ./pkg/utils/
//
// 查看覆盖率：
// go test -cover ./pkg/utils/
//
// 生成覆盖率报告：
// go test -coverprofile=coverage.out ./pkg/utils/
// go tool cover -html=coverage.out
