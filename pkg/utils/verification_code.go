package utils

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/redis"
)

const (
	// CodeLength 验证码长度
	CodeLength = 6
	// CodeExpiration 验证码有效期（分钟）
	CodeExpiration = 5
	// CodeSendInterval 验证码发送间隔（秒）
	CodeSendInterval = 60
)

// GenerateVerificationCode 生成验证码
func GenerateVerificationCode() string {
	rand.Seed(time.Now().UnixNano())
	code := ""
	for i := 0; i < CodeLength; i++ {
		code += fmt.Sprintf("%d", rand.Intn(10))
	}
	return code
}

// StoreVerificationCode 存储验证码到 Redis
func StoreVerificationCode(ctx context.Context, key string, code string) error {
	// 存储验证码，设置过期时间
	err := redis.Client.Set(ctx, key, code, time.Duration(CodeExpiration)*time.Minute).Err()
	if err != nil {
		return errs.SystemError("验证码存储失败")
	}
	return nil
}

// VerifyCode 验证验证码
func VerifyCode(ctx context.Context, key string, inputCode string) error {
	// 从 Redis 获取验证码
	storedCode, err := redis.Client.Get(ctx, key).Result()
	if err != nil {
		return errs.BadRequest("验证码已过期或不存在")
	}

	// 验证码比对
	if storedCode != inputCode {
		return errs.BadRequest("验证码错误")
	}

	// 验证成功后删除验证码
	redis.Client.Del(ctx, key)

	return nil
}

// CheckSendInterval 检查发送间隔
func CheckSendInterval(ctx context.Context, intervalKey string) error {
	// 检查是否在发送间隔内
	exists, err := redis.Client.Exists(ctx, intervalKey).Result()
	if err != nil {
		return errs.SystemError("验证码发送检查失败")
	}

	if exists > 0 {
		ttl, _ := redis.Client.TTL(ctx, intervalKey).Result()
		return errs.BadRequest(fmt.Sprintf("验证码发送过于频繁，请 %d 秒后再试", int(ttl.Seconds())))
	}

	// 设置发送间隔标记
	redis.Client.Set(ctx, intervalKey, "1", time.Duration(CodeSendInterval)*time.Second)

	return nil
}

// GetMobileCodeKey 获取手机验证码 Redis Key
func GetMobileCodeKey(mobile string) string {
	return fmt.Sprintf("sms:code:%s", mobile)
}

// GetMobileIntervalKey 获取手机验证码发送间隔 Redis Key
func GetMobileIntervalKey(mobile string) string {
	return fmt.Sprintf("sms:interval:%s", mobile)
}

// GetEmailCodeKey 获取邮箱验证码 Redis Key
func GetEmailCodeKey(email string) string {
	return fmt.Sprintf("email:code:%s", email)
}

// GetEmailIntervalKey 获取邮箱验证码发送间隔 Redis Key
func GetEmailIntervalKey(email string) string {
	return fmt.Sprintf("email:interval:%s", email)
}
