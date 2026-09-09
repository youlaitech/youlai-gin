package repository

import (
	"context"

	"gorm.io/gorm"

	"youlai-gin/internal/common/database"
	"youlai-gin/internal/system/notice/model"
	"youlai-gin/pkg/types"
)

// Repository 通知公告数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建 Repository 实例
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// Page 通知分页查询
func (r *Repository) Page(ctx context.Context, query *model.NoticeQuery) ([]model.Notice, int64, error) {
	var notices []model.Notice
	var total int64

	db := r.db.WithContext(ctx).Table("sys_notice n").
		Select("n.*, u.nickname AS publisher_name").
		Joins("LEFT JOIN sys_user u ON n.publisher_id = u.id").
		Where("n.is_deleted = 0")

	if query.Title != "" {
		db = db.Where("title LIKE ?", "%"+query.Title+"%")
	}

	if query.Type != nil {
		db = db.Where("type = ?", *query.Type)
	}

	if query.Status != nil {
		db = db.Where("n.publish_status = ?", *query.Status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Scopes(database.PaginateFromQuery(query)).
		Order("n.create_time DESC").
		Find(&notices).Error

	return notices, total, err
}

// Get 根据 ID 获取通知
func (r *Repository) Get(ctx context.Context, id int64) (*model.Notice, error) {
	var notice model.Notice
	err := r.db.WithContext(ctx).Table("sys_notice n").
		Select("n.*, u.nickname AS publisher_name").
		Joins("LEFT JOIN sys_user u ON n.publisher_id = u.id").
		Where("n.id = ? AND n.is_deleted = 0", id).
		First(&notice).Error
	return &notice, err
}

// Create 创建通知（ctx 携带操作人，由审计钩子填充 create_by）
func (r *Repository) Create(ctx context.Context, notice *model.Notice) error {
	return r.db.WithContext(ctx).Create(notice).Error
}

// Update 部分更新通知
// patch 为「列名→值」映射（BuildPatchMap 生成，可补充 publish_time、target_user_ids 等转换字段），
// 指针字段 nil 跳过、非 nil（含 0）写入，从根上解决 GORM Updates(struct) 默认跳过零值的问题。
func (r *Repository) Update(ctx context.Context, id int64, patch map[string]any) error {
	if id <= 0 || len(patch) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.Notice{}).
		Where("id = ?", id).
		Updates(patch).Error
}

// Delete 软删除通知
func (r *Repository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.Notice{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// DeleteUserNoticesByNoticeID 删除通知关联的用户通知记录（重新发布/撤回场景）
func (r *Repository) DeleteUserNoticesByNoticeID(ctx context.Context, noticeID int64) error {
	return r.db.WithContext(ctx).Exec("DELETE FROM sys_user_notice WHERE notice_id = ?", noticeID).Error
}

// UserNoticePage 用户通知分页查询
func (r *Repository) UserNoticePage(ctx context.Context, userID int64, query *model.UserNoticeQuery) ([]model.Notice, int64, error) {
	var notices []model.Notice
	var total int64

	baseWhere := func(db *gorm.DB) *gorm.DB {
		db = db.Where("n.is_deleted = 0 AND n.publish_status = 1")
		db = db.Where("(n.target_type = 1 OR (n.target_type = 2 AND FIND_IN_SET(?, n.target_user_ids)))", userID)

		if query.Type != nil {
			db = db.Where("n.type = ?", *query.Type)
		}

		if query.IsRead != nil {
			if *query.IsRead == 1 {
				db = db.Joins("INNER JOIN sys_user_notice un ON n.id = un.notice_id AND un.user_id = ? AND un.is_read = 1 AND un.is_deleted = 0", userID)
			} else {
				db = db.Where("NOT EXISTS (SELECT 1 FROM sys_user_notice WHERE notice_id = n.id AND user_id = ? AND is_read = 1 AND is_deleted = 0)", userID)
			}
		}
		return db
	}

	countDB := baseWhere(r.db.WithContext(ctx).Table("sys_notice n"))
	if err := countDB.Distinct("n.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	dataDB := baseWhere(r.db.WithContext(ctx).Table("sys_notice n").
		Select("DISTINCT n.*, u.nickname AS publisher_name").
		Joins("LEFT JOIN sys_user u ON n.publisher_id = u.id"))
	err := dataDB.Scopes(database.PaginateFromQuery(query)).
		Order("n.create_time DESC").
		Find(&notices).Error

	return notices, total, err
}

// MarkRead 标记通知为已读（记录不存在时创建）
func (r *Repository) MarkRead(ctx context.Context, noticeID, userID int64) error {
	db := r.db.WithContext(ctx)

	var userNotice model.UserNotice
	err := db.Where("notice_id = ? AND user_id = ? AND is_deleted = 0", noticeID, userID).First(&userNotice).Error

	if err == nil {
		if userNotice.IsRead == 0 {
			return db.Model(&userNotice).Updates(map[string]interface{}{
				"is_read":   1,
				"read_time": db.NowFunc(),
			}).Error
		}
		return nil
	}

	userNotice = model.UserNotice{
		NoticeID: types.BigInt(noticeID),
		UserID:   types.BigInt(userID),
		IsRead:   1,
	}
	return db.Create(&userNotice).Error
}

// UnreadCount 用户未读通知数量
func (r *Repository) UnreadCount(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("sys_notice n").
		Where("n.is_deleted = 0 AND n.publish_status = 1").
		Where("(n.target_type = 1 OR (n.target_type = 2 AND FIND_IN_SET(?, n.target_user_ids)))", userID).
		Where("NOT EXISTS (SELECT 1 FROM sys_user_notice WHERE notice_id = n.id AND user_id = ? AND is_read = 1 AND is_deleted = 0)", userID).
		Count(&count).Error
	return count, err
}
