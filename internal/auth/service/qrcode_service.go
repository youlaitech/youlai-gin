package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	authModel "youlai-gin/internal/auth/model"
	"youlai-gin/internal/common/auth"
	"youlai-gin/internal/common/redis"
	permService "youlai-gin/internal/common/permission/service"
	userRepo "youlai-gin/internal/system/user/repository"
	"youlai-gin/pkg/errs"
)

// 扫码登录票据在 Redis 中的 Key 模板、默认有效期与最小补足 TTL
const (
	qrCodeRedisKeyFmt        = "auth:qr_code:%s"
	qrCodeExpireSeconds      = 300
	qrCodeMinRemainSeconds   = 30
)

// QrCodeGenerate 生成扫码登录票据，状态为 WAITING
func QrCodeGenerate(clientIP string) (*authModel.QrCodeGenerateVO, error) {
	ticket := newQrCodeTicket()
	ctx := &authModel.QrCodeLoginContext{
		Ticket:    ticket,
		Status:    authModel.QrCodeStatusWaiting,
		CreatedAt: time.Now().UnixMilli(),
		ClientIP:  clientIP,
	}
	if err := saveQrCodeContext(ctx, qrCodeExpireSeconds); err != nil {
		return nil, err
	}
	return &authModel.QrCodeGenerateVO{
		Ticket:        ticket,
		ExpireSeconds: qrCodeExpireSeconds,
	}, nil
}

// QrCodeStatus 查询票据当前状态
func QrCodeStatus(ticket string) (*authModel.QrCodeStatusVO, error) {
	ctx, err := loadQrCodeContext(ticket)
	if err != nil {
		return nil, err
	}
	return toQrCodeStatusVO(ctx, remainQrCodeSeconds(ticket)), nil
}

// QrCodeScan APP 标记已扫码：须处于 WAITING，写入扫码用户昵称与头像
func QrCodeScan(ticket string, userID int64) (*authModel.QrCodeStatusVO, error) {
	ctx, err := loadQrCodeContext(ticket)
	if err != nil {
		return nil, err
	}
	if err := requireQrCodeStatus(ctx, authModel.QrCodeStatusWaiting); err != nil {
		return nil, err
	}
	if err := fillQrCodeUserInfo(ctx, userID); err != nil {
		return nil, err
	}
	ctx.Status = authModel.QrCodeStatusScanned
	ctx.ScannedAt = time.Now().UnixMilli()
	if err := saveQrCodeContext(ctx, refreshQrCodeTtl(ticket)); err != nil {
		return nil, err
	}
	return toQrCodeStatusVO(ctx, remainQrCodeSeconds(ticket)), nil
}

// QrCodeConfirm APP 确认登录：须处于 SCANNED，操作者须为扫码本人
func QrCodeConfirm(ticket string, userID int64) (*authModel.QrCodeStatusVO, error) {
	ctx, err := loadQrCodeContext(ticket)
	if err != nil {
		return nil, err
	}
	if err := requireQrCodeStatus(ctx, authModel.QrCodeStatusScanned); err != nil {
		return nil, err
	}
	if err := requireQrCodeSameUser(ctx, userID); err != nil {
		return nil, err
	}
	ctx.Status = authModel.QrCodeStatusConfirmed
	ctx.ConfirmedAt = time.Now().UnixMilli()
	if err := saveQrCodeContext(ctx, refreshQrCodeTtl(ticket)); err != nil {
		return nil, err
	}
	return toQrCodeStatusVO(ctx, remainQrCodeSeconds(ticket)), nil
}

// QrCodeCancel APP 取消登录：WAITING/SCANNED/CONFIRMED 允许，已扫码后仅扫码本人可取消
func QrCodeCancel(ticket string, userID int64) (*authModel.QrCodeStatusVO, error) {
	ctx, err := loadQrCodeContext(ticket)
	if err != nil {
		return nil, err
	}
	if ctx.Status != authModel.QrCodeStatusWaiting &&
		ctx.Status != authModel.QrCodeStatusScanned &&
		ctx.Status != authModel.QrCodeStatusConfirmed {
		return nil, errs.QrCodeStatusIllegal()
	}
	if ctx.Status != authModel.QrCodeStatusWaiting && ctx.UserID != 0 {
		if err := requireQrCodeSameUser(ctx, userID); err != nil {
			return nil, err
		}
	}
	ctx.Status = authModel.QrCodeStatusCanceled
	if err := saveQrCodeContext(ctx, refreshQrCodeTtl(ticket)); err != nil {
		return nil, err
	}
	return toQrCodeStatusVO(ctx, remainQrCodeSeconds(ticket)), nil
}

// QrCodeLogin PC 端用票据换取会话令牌：须处于 CONFIRMED，成功后票据置 LOGGED_IN
func QrCodeLogin(ticket string) (*auth.AuthenticationToken, error) {
	ctx, err := loadQrCodeContext(ticket)
	if err != nil {
		return nil, err
	}
	if err := requireQrCodeStatus(ctx, authModel.QrCodeStatusConfirmed); err != nil {
		return nil, err
	}
	user, err := userRepo.GetUserByID(ctx.UserID)
	if err != nil {
		return nil, errs.UserNotFound()
	}
	if user.Status != 1 {
		return nil, errs.BadRequest("用户已被禁用")
	}
	roles, err := userRepo.GetUserRoles(ctx.UserID)
	if err != nil {
		return nil, errs.SystemError("查询用户角色失败")
	}
	dataScopes, err := permService.GetUserDataScopes(ctx.UserID, roles, int64(user.DeptID))
	if err != nil {
		return nil, err
	}
	userDetails := &auth.UserDetails{
		UserID:     ctx.UserID,
		Username:   user.Username,
		DeptID:     user.DeptID,
		DataScopes: dataScopes,
		Roles:      roles,
		Avatar:     user.Avatar,
	}
	token, err := tokenManager.GenerateToken(userDetails)
	if err != nil {
		return nil, errs.SystemError("生成令牌失败")
	}
	// 置为已使用，再次 login 会在 requireQrCodeStatus(CONFIRMED) 处被拒，杜绝重放
	ctx.Status = authModel.QrCodeStatusLoggedIn
	remain := remainQrCodeSeconds(ticket)
	if remain < qrCodeMinRemainSeconds {
		remain = qrCodeMinRemainSeconds
	}
	if err := saveQrCodeContext(ctx, remain); err != nil {
		return nil, err
	}
	return token, nil
}

// ======================== 内部工具 ========================

func newQrCodeTicket() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func qrCodeRedisKey(ticket string) string {
	return fmt.Sprintf(qrCodeRedisKeyFmt, ticket)
}

func saveQrCodeContext(ctx *authModel.QrCodeLoginContext, ttl int) error {
	data, err := json.Marshal(ctx)
	if err != nil {
		return errs.SystemError("序列化扫码上下文失败")
	}
	c := context.Background()
	return redis.Client.Set(c, qrCodeRedisKey(ctx.Ticket), string(data), time.Duration(ttl)*time.Second).Err()
}

func loadQrCodeContext(ticket string) (*authModel.QrCodeLoginContext, error) {
	if ticket == "" {
		return nil, errs.QrCodeNotFound()
	}
	c := context.Background()
	val, err := redis.Client.Get(c, qrCodeRedisKey(ticket)).Result()
	if err != nil {
		return nil, errs.QrCodeNotFound()
	}
	var ctx authModel.QrCodeLoginContext
	if err := json.Unmarshal([]byte(val), &ctx); err != nil {
		return nil, errs.QrCodeNotFound()
	}
	return &ctx, nil
}

func remainQrCodeSeconds(ticket string) int {
	c := context.Background()
	d, err := redis.Client.TTL(c, qrCodeRedisKey(ticket)).Result()
	if err != nil || d < 0 {
		return 0
	}
	return int(d.Seconds())
}

// 状态流转后写回的 TTL：维持剩余时间，不足最小补足值则补足
func refreshQrCodeTtl(ticket string) int {
	remain := remainQrCodeSeconds(ticket)
	if remain < qrCodeMinRemainSeconds {
		return qrCodeMinRemainSeconds
	}
	return remain
}

func requireQrCodeStatus(ctx *authModel.QrCodeLoginContext, expected string) error {
	if ctx.Status != expected {
		return errs.QrCodeStatusIllegal()
	}
	return nil
}

// 操作者必须是当初扫码的用户，防止 A 扫码 B 确认
func requireQrCodeSameUser(ctx *authModel.QrCodeLoginContext, userID int64) error {
	if ctx.UserID == 0 || ctx.UserID != userID {
		return errs.QrCodeUserMismatch()
	}
	return nil
}

// 扫码时把当前 APP 用户的昵称、头像写入上下文，供 PC 端 status 展示
func fillQrCodeUserInfo(ctx *authModel.QrCodeLoginContext, userID int64) error {
	user, err := userRepo.GetUserByID(userID)
	if err != nil {
		return errs.UserNotFound()
	}
	ctx.UserID = userID
	ctx.Nickname = user.Nickname
	ctx.Avatar = user.Avatar
	return nil
}

// 上下文转前端 VO；仅 SCANNED/CONFIRMED 阶段回传脱敏昵称与头像，WAITING 返回 null
func toQrCodeStatusVO(ctx *authModel.QrCodeLoginContext, expireSeconds int) *authModel.QrCodeStatusVO {
	vo := &authModel.QrCodeStatusVO{
		Ticket:        ctx.Ticket,
		Status:        ctx.Status,
		ExpireSeconds: expireSeconds,
	}
	if ctx.Status == authModel.QrCodeStatusScanned || ctx.Status == authModel.QrCodeStatusConfirmed {
		nick := maskQrCodeNickname(ctx.Nickname)
		vo.Nickname = &nick
		av := ctx.Avatar
		vo.Avatar = &av
	}
	return vo
}

// maskQrCodeNickname 昵称脱敏：保留首尾字符，中间以 * 填充
func maskQrCodeNickname(nickname string) string {
	runes := []rune(nickname)
	n := len(runes)
	if n <= 1 {
		return nickname
	}
	if n == 2 {
		return string(runes[0]) + "*"
	}
	masked := string(runes[0])
	for i := 0; i < n-2; i++ {
		masked += "*"
	}
	masked += string(runes[n-1])
	return masked
}
