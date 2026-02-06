package utils

import (
	"testing"
)

// TestGenerateVerificationCode 验证码生成
func TestGenerateVerificationCode(t *testing.T) {
	// 生成验证码
	code := GenerateVerificationCode()

	// 验证长度
	if len(code) != CodeLength {
		t.Errorf("验证码长度错误，期望 %d，实际 %d", CodeLength, len(code))
	}

	// 验证是否全是数字
	for _, c := range code {
		if c < '0' || c > '9' {
			t.Errorf("验证码包含非数字字符: %c", c)
		}
	}

	// 测试多次生成，确保不重复（概率性测试）
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code := GenerateVerificationCode()
		codes[code] = true
	}

	// 100 次生成应该有多个不同的验证码
	if len(codes) < 10 {
		t.Errorf("验证码重复率过高，100 次生成只有 %d 个不同的验证码", len(codes))
	}
}

// TestGetMobileCodeKey 手机验证码 Key
func TestGetMobileCodeKey(t *testing.T) {
	tests := []struct {
		name     string
		mobile   string
		expected string
	}{
		{
			name:     "正常手机号",
			mobile:   "13800138000",
			expected: "sms:code:13800138000",
		},
		{
			name:     "空手机号",
			mobile:   "",
			expected: "sms:code:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMobileCodeKey(tt.mobile)
			if result != tt.expected {
				t.Errorf("GetMobileCodeKey() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetEmailCodeKey 邮箱验证码 Key
func TestGetEmailCodeKey(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "正常邮箱",
			email:    "test@example.com",
			expected: "email:code:test@example.com",
		},
		{
			name:     "空邮箱",
			email:    "",
			expected: "email:code:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEmailCodeKey(tt.email)
			if result != tt.expected {
				t.Errorf("GetEmailCodeKey() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// BenchmarkGenerateVerificationCode 性能测试
func BenchmarkGenerateVerificationCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateVerificationCode()
	}
}

// 运行测试：go test -v ./pkg/utils/
// 性能测试：go test -bench=. ./pkg/utils/
// 覆盖率：go test -cover ./pkg/utils/
