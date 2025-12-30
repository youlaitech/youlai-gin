package model

import (
	"youlai-gin/pkg/types"
)

// Converter 转换器
// 作用：分离实体转换逻辑，避免代码重复
// 参考：Java 的 MapStruct、BeanUtils

// ToUserVO 将 User 实体转换为 UserVO
func ToUserVO(user *User) *UserVO {
	if user == nil {
		return nil
	}

	return &UserVO{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Mobile:   user.Mobile,
		Gender:   user.Gender,
		Email:    user.Email,
		Status:   user.Status,
	}
}

// ToUserForm 将 User 实体转换为 UserForm
func ToUserForm(user *User, roleIDs []types.BigInt) *UserForm {
	if user == nil {
		return nil
	}

	return &UserForm{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Mobile:   user.Mobile,
		Gender:   user.Gender,
		Avatar:   user.Avatar,
		Email:    user.Email,
		Status:   user.Status,
		DeptID:   user.DeptID,
		RoleIDs:  roleIDs,
	}
}

// ToUser 将 UserForm 转换为 User 实体
func ToUser(form *UserForm) *User {
	if form == nil {
		return nil
	}

	return &User{
		ID:       form.ID,
		Username: form.Username,
		Nickname: form.Nickname,
		Mobile:   form.Mobile,
		Gender:   form.Gender,
		Avatar:   form.Avatar,
		Email:    form.Email,
		Status:   form.Status,
		DeptID:   form.DeptID,
	}
}

// ToUserProfileVO 将 User 实体转换为 UserProfileVO
func ToUserProfileVO(user *User, deptName string, roleNames string) *UserProfileVO {
	if user == nil {
		return nil
	}

	return &UserProfileVO{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Gender:    user.Gender,
		Mobile:    user.Mobile,
		Email:     user.Email,
		DeptName:  deptName,
		RoleNames: roleNames,
	}
}

// ToUserPageVO 将 User 实体转换为 UserPageVO
func ToUserPageVO(user *User, deptName string, roleNames string) *UserPageVO {
	if user == nil {
		return nil
	}

	return &UserPageVO{
		ID:         user.ID,
		Username:   user.Username,
		Nickname:   user.Nickname,
		Mobile:     user.Mobile,
		Gender:     user.Gender,
		Avatar:     user.Avatar,
		Email:      user.Email,
		Status:     user.Status,
		DeptName:   deptName,
		RoleNames:  roleNames,
		CreateTime: user.CreateTime,
	}
}

// ToUserList 批量转换 User 列表为 UserVO 列表
func ToUserVOList(users []*User) []*UserVO {
	if users == nil {
		return nil
	}

	vos := make([]*UserVO, 0, len(users))
	for _, user := range users {
		vos = append(vos, ToUserVO(user))
	}
	return vos
}
