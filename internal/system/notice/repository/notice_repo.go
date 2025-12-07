package repository

import (
	"gorm.io/gorm"
	
	"youlai-gin/internal/database"
	"youlai-gin/internal/system/notice/model"
	pkgDatabase "youlai-gin/pkg/database"
)

// GetNoticePage 通知分页查询
func GetNoticePage(query *model.NoticePageQuery) ([]model.Notice, int64, error) {
	var notices []model.Notice
	var total int64
	
	db := database.DB.Model(&model.Notice{}).Where("is_deleted = 0")
	
	if query.Title != "" {
		db = db.Where("title LIKE ?", "%"+query.Title+"%")
	}
	
	if query.Type != nil {
		db = db.Where("type = ?", *query.Type)
	}
	
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	err := db.Scopes(pkgDatabase.PaginateFromQuery(query)).
		Order("create_time DESC").
		Find(&notices).Error
	
	return notices, total, err
}

// GetNoticeByID 根据ID获取通知
func GetNoticeByID(id int64) (*model.Notice, error) {
	var notice model.Notice
	err := database.DB.Where("id = ? AND is_deleted = 0", id).First(&notice).Error
	return &notice, err
}

// CreateNotice 创建通知
func CreateNotice(notice *model.Notice) error {
	return database.DB.Create(notice).Error
}

// UpdateNotice 更新通知
func UpdateNotice(notice *model.Notice) error {
	return database.DB.Model(notice).Updates(notice).Error
}

// DeleteNotice 删除通知
func DeleteNotice(id int64) error {
	return database.DB.Model(&model.Notice{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// GetUserNoticePage 获取用户通知列表（分页）
func GetUserNoticePage(userID int64, query *model.UserNoticeQuery) ([]model.Notice, int64, error) {
	var notices []model.Notice
	var total int64
	
	// 构建基础查询条件
	baseWhere := func(db *gorm.DB) *gorm.DB {
		db = db.Where("n.is_deleted = 0 AND n.publish_status = 1")
		db = db.Where("(n.target_type = 1 OR (n.target_type = 2 AND FIND_IN_SET(?, n.target_user_ids)))", userID)
		
		if query.Type != nil {
			db = db.Where("n.type = ?", *query.Type)
		}
		
		// 已读/未读过滤
		if query.IsRead != nil {
			if *query.IsRead == 1 {
				// 已读
				db = db.Joins("INNER JOIN sys_user_notice un ON n.id = un.notice_id AND un.user_id = ? AND un.is_read = 1 AND un.is_deleted = 0", userID)
			} else {
				// 未读
				db = db.Where("NOT EXISTS (SELECT 1 FROM sys_user_notice WHERE notice_id = n.id AND user_id = ? AND is_read = 1 AND is_deleted = 0)", userID)
			}
		}
		return db
	}
	
	// 统计总数 - 使用 COUNT(DISTINCT n.id) 避免 JOIN 导致的重复
	countDB := database.DB.Table("sys_notice n")
	countDB = baseWhere(countDB)
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 查询数据
	dataDB := database.DB.Table("sys_notice n").Select("n.*")
	dataDB = baseWhere(dataDB)
	err := dataDB.Scopes(pkgDatabase.PaginateFromQuery(query)).
		Order("n.create_time DESC").
		Find(&notices).Error
	
	return notices, total, err
}

// MarkNoticeAsRead 标记通知为已读
func MarkNoticeAsRead(noticeID, userID int64) error {
	// 检查记录是否存在
	var userNotice model.UserNotice
	err := database.DB.Where("notice_id = ? AND user_id = ? AND is_deleted = 0", noticeID, userID).First(&userNotice).Error
	
	if err == nil {
		// 记录存在，更新为已读
		if userNotice.IsRead == 0 {
			return database.DB.Model(&userNotice).Updates(map[string]interface{}{
				"is_read":   1,
				"read_time": database.DB.NowFunc(),
			}).Error
		}
		return nil // 已读，无需更新
	}
	
	// 记录不存在，创建新记录
	userNotice = model.UserNotice{
		NoticeID: noticeID,
		UserID:   userID,
		IsRead:   1,
	}
	return database.DB.Create(&userNotice).Error
}

// GetUnreadCount 获取用户未读通知数量
func GetUnreadCount(userID int64) (int64, error) {
	var count int64
	err := database.DB.Table("sys_notice n").
		Where("n.is_deleted = 0 AND n.publish_status = 1").
		Where("(n.target_type = 1 OR (n.target_type = 2 AND FIND_IN_SET(?, n.target_user_ids)))", userID).
		Where("NOT EXISTS (SELECT 1 FROM sys_user_notice WHERE notice_id = n.id AND user_id = ? AND is_read = 1 AND is_deleted = 0)", userID).
		Count(&count).Error
	return count, err
}
