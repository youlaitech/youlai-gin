package service

import (
	"youlai-gin/internal/user/model"
	"youlai-gin/internal/user/repository"
	"youlai-gin/pkg/apperror"
)

// AutoMigrate 对外暴露，供启动时调用
func AutoMigrate() error {
	return repository.AutoMigrate()
}

// ListUsers 简单返回全部用户
func ListUsers() ([]model.User, error) {
	return repository.FindAll()
}

// CreateUser 创建用户
func CreateUser(user *model.User) error {
	// 示例：业务校验（用户名不能为空）。真实项目可根据实际业务补充更多规则。
	if user.Username == "" {
		return apperror.ErrBadRequest("用户名不能为空")
	}

	return repository.Create(user)
}

// UpdateUser 更新用户
func UpdateUser(id uint64, user *model.User) error {
	return repository.Update(id, user)
}

// DeleteUser 删除用户
func DeleteUser(id uint64) error {
	return repository.Delete(id)
}
