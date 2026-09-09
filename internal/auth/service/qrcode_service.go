package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	authModel "youlai-gin/internal/auth/model"
	"youlai-gin/internal/common/auth"
	"youlai-gin/internal/common/redis"
	"youlai-gin/pkg/errs"
)

// 扫码登录票据默认有效期与最小补足 TTL（Redis Key 统一定义在 common/redis/keys.go）
const (
	qrCodeExpireSeconds    = 300
	qrCodeMinRemainSeconds = 30
)

// QrCodeService 扫码登录业务逻辑层（票据状态机 + 令牌签发）
type QrCodeService struct {
	issuer   *TokenIssuer
	userRepo UserRepository
}

// NewQrCodeService 创建 QrCodeService 实例
func NewQrCodeService(issuer *TokenIssuer, userRepo UserRepository) *QrCodeService {
	return &QrCodeService{issuer: issuer, userRepo: userRepo}
}

// Generate 生成扫码登录票据，状态为 WAITING
func (s *QrCodeService) Generate(ctx context.Context, clientIP string) (*authModel.QrCodeGenerateVO, error) {
	ticket := newQrCodeTicket()
	qctx := &authModel.QrCodeLoginContext{
		Ticket:    ticket,
		Status:    authModel.QrCodeStatusWaiting,
		CreatedAt: time.Now().UnixMilli(),
		ClientIP:  clientIP,
	}
	if err := saveQrCodeContext(ctx, qctx, qrCodeExpireSeconds); err != nil {
		return nil, err
	}
	return &authModel.QrCodeGenerateVO{
		Ticket:        ticket,
		ExpireSeconds: qrCodeExpireSeconds,
	}, nil
}

// Status 查询票据当前状态
func (s *QrCodeService) Status(ctx context.Context, ticket string) (*authModel.QrCodeStatusVO, error) {
	qctx, err := loadQrCodeContext(ctx, ticket)
	if err != nil {
		return nil, err
	}
	return toQrCodeStatusVO(qctx, remainQrCodeSeconds(ctx, ticket)), nil
}

// Scan APP 标记已扫码：须处于 WAITING，写入扫码用户昵称与头像
func (s *QrCodeService) Scan(ctx context.Context, ticket string, userID int64) (*authModel.QrCodeStatusVO, error) {
	qctx, err := loadQrCodeContext(ctx, ticket)
	if err != nil {
		return nil, err
	}
	if err := requireQrCodeStatus(qctx, authModel.QrCodeStatusWaiting); err != nil {
		return nil, err
	}
	if err := s.fillUserInfo(ctx, qctx, userID); err != nil {
		return nil, err
	}
	qctx.Status = authModel.QrCodeStatusScanned
	qctx.ScannedAt = time.Now().UnixMilli()
	if err := saveQrCodeContext(ctx, qctx, refreshQrCodeTtl(ctx, ticket)); err != nil {
		return nil, err
	}
	return toQrCodeStatusVO(qctx, remainQrCodeSeconds(ctx, ticket)), nil
}

// Confirm APP 确认登录：须处于 SCANNED，操作者须为扫码本人
func (s *QrCodeService) Confirm(ctx context.Context, ticket string, userID int64) (*authModel.QrCodeStatusVO, error) {
	qctx, err := loadQrCodeContext(ctx, ticket)
	if err != nil {
		return nil, err
	}
	if err := requireQrCodeStatus(qctx, authModel.QrCodeStatusScanned); err != nil {
		return nil, err
	}
	if err := requireQrCodeSameUser(qctx, userID); err != nil {
		return nil, err
	}
	qctx.Status = authModel.QrCodeStatusConfirmed
	qctx.ConfirmedAt = time.Now().UnixMilli()
	if err := saveQrCodeContext(ctx, qctx, refreshQrCodeTtl(ctx, ticket)); err != nil {
		return nil, err
	}
	return toQrCodeStatusVO(qctx, remainQrCodeSeconds(ctx, ticket)), nil
}

// Cancel APP 取消登录：WAITING/SCANNED/CONFIRMED 允许，已扫码后仅扫码本人可取消
func (s *QrCodeService) Cancel(ctx context.Context, ticket string, userID int64) (*authModel.QrCodeStatusVO, error) {
	qctx, err := loadQrCodeContext(ctx, ticket)
	if err != nil {
		return nil, err
	}
	if qctx.Status != authModel.QrCodeStatusWaiting &&
		qctx.Status != authModel.QrCodeStatusScanned &&
		qctx.Status != authModel.QrCodeStatusConfirmed {
		return nil, errs.QrCodeStatusIllegal()
	}
	if qctx.Status != authModel.QrCodeStatusWaiting && qctx.UserID != 0 {
		if err := requireQrCodeSameUser(qctx, userID); err != nil {
			return nil, err
		}
	}
	qctx.Status = authModel.QrCodeStatusCanceled
	if err := saveQrCodeContext(ctx, qctx, refreshQrCodeTtl(ctx, ticket)); err != nil {
		return nil, err
	}
	return toQrCodeStatusVO(qctx, remainQrCodeSeconds(ctx, ticket)), nil
}

// Login PC 端用票据换取会话令牌：须处于 CONFIRMED，成功后票据置 LOGGED_IN 杜绝重放
func (s *QrCodeService) Login(ctx context.Context, ticket string) (*auth.AuthenticationToken, error) {
	qctx, err := loadQrCodeContext(ctx, ticket)
	if err != nil {
		return nil, err
	}
	if err := requireQrCodeStatus(qctx, authModel.QrCodeStatusConfirmed); err != nil {
		return nil, err
	}
	user, err := s.userRepo.Get(ctx, qctx.UserID)
	if err != nil {
		return nil, errs.UserNotFound()
	}
	if user.Status != 1 {
		return nil, errs.BadRequest("用户已被禁用")
	}

	token, err := s.issuer.Issue(ctx, user)
	if err != nil {
		return nil, err
	}

	qctx.Status = authModel.QrCodeStatusLoggedIn
	remain := remainQrCodeSeconds(ctx, ticket)
	if remain < qrCodeMinRemainSeconds {
		remain = qrCodeMinRemainSeconds
	}
	if err := saveQrCodeContext(ctx, qctx, remain); err != nil {
		return nil, err
	}
	return token, nil
}

// fillUserInfo 扫码时把当前 APP 用户的昵称、头像写入上下文，供 PC 端 status 展示
func (s *QrCodeService) fillUserInfo(ctx context.Context, qctx *authModel.QrCodeLoginContext, userID int64) error {
	user, err := s.userRepo.Get(ctx, userID)
	if err != nil {
		return errs.UserNotFound()
	}
	qctx.UserID = userID
	qctx.Nickname = user.Nickname
	qctx.Avatar = user.Avatar
	return nil
}

// ======================== 内部工具 ========================

func newQrCodeTicket() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// qrCodeRedisKey 由票据拼接 Redis Key
func qrCodeRedisKey(ticket string) string {
	return redis.QrCodeTicketPrefix + ticket
}

// saveQrCodeContext 序列化扫码上下文并写入 Redis
func saveQrCodeContext(ctx context.Context, qctx *authModel.QrCodeLoginContext, ttl int) error {
	data, err := json.Marshal(qctx)
	if err != nil {
		return errs.SystemError("序列化扫码上下文失败")
	}
	return redis.Client.Set(ctx, qrCodeRedisKey(qctx.Ticket), string(data), time.Duration(ttl)*time.Second).Err()
}

// loadQrCodeContext 从 Redis 读取并反序列化扫码上下文
func loadQrCodeContext(ctx context.Context, ticket string) (*authModel.QrCodeLoginContext, error) {
	if ticket == "" {
		return nil, errs.QrCodeNotFound()
	}
	val, err := redis.Client.Get(ctx, qrCodeRedisKey(ticket)).Result()
	if err != nil {
		return nil, errs.QrCodeNotFound()
	}
	var qctx authModel.QrCodeLoginContext
	if err := json.Unmarshal([]byte(val), &qctx); err != nil {
		return nil, errs.QrCodeNotFound()
	}
	return &qctx, nil
}

// remainQrCodeSeconds 返回票据剩余有效秒数
func remainQrCodeSeconds(ctx context.Context, ticket string) int {
	d, err := redis.Client.TTL(ctx, qrCodeRedisKey(ticket)).Result()
	if err != nil || d < 0 {
		return 0
	}
	return int(d.Seconds())
}

// refreshQrCodeTtl 状态流转后写回的 TTL：维持剩余时间，不足最小补足值则补足
func refreshQrCodeTtl(ctx context.Context, ticket string) int {
	remain := remainQrCodeSeconds(ctx, ticket)
	if remain < qrCodeMinRemainSeconds {
		return qrCodeMinRemainSeconds
	}
	return remain
}

// requireQrCodeStatus 校验上下文处于期望状态
func requireQrCodeStatus(qctx *authModel.QrCodeLoginContext, expected string) error {
	if qctx.Status != expected {
		return errs.QrCodeStatusIllegal()
	}
	return nil
}

// requireQrCodeSameUser 操作者必须是当初扫码的用户，防止 A 扫码 B 确认
func requireQrCodeSameUser(qctx *authModel.QrCodeLoginContext, userID int64) error {
	if qctx.UserID == 0 || qctx.UserID != userID {
		return errs.QrCodeUserMismatch()
	}
	return nil
}

// toQrCodeStatusVO 上下文转前端 VO；仅 SCANNED/CONFIRMED 阶段回传脱敏昵称与头像，WAITING 返回 null
func toQrCodeStatusVO(qctx *authModel.QrCodeLoginContext, expireSeconds int) *authModel.QrCodeStatusVO {
	vo := &authModel.QrCodeStatusVO{
		Ticket:        qctx.Ticket,
		Status:        qctx.Status,
		ExpireSeconds: expireSeconds,
	}
	if qctx.Status == authModel.QrCodeStatusScanned || qctx.Status == authModel.QrCodeStatusConfirmed {
		nick := maskQrCodeNickname(qctx.Nickname)
		vo.Nickname = &nick
		av := qctx.Avatar
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
