package repository

import (
	"youlai-gin/internal/database"
	"youlai-gin/internal/user/model"
)

// AutoMigrate 确保 users 表存在
func AutoMigrate() error {
	return database.DB.AutoMigrate(&model.User{})
}

// FindAll 返回所有用户
func FindAll() ([]model.User, error) {
	var users []model.User
	err := database.DB.Find(&users).Error
	return users, err
}

// Create 创建用户
func Create(user *model.User) error {
	return database.DB.Create(user).Error
}

// Update 根据 ID 更新用户
func Update(id uint64, user *model.User) error {
	user.ID = id
	return database.DB.Model(&model.User{}).Where("id = ?", id).Updates(user).Error
}

// Delete 根据 ID 删除用户
func Delete(id uint64) error {
	return database.DB.Delete(&model.User{}, id).Error
}
