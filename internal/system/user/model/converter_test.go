package model

import (
	"testing"

	"youlai-gin/pkg/types"
)

// TestToUserVO 测试 User 转 UserVO
func TestToUserVO(t *testing.T) {
	// 测试正常转换
	t.Run("正常转换", func(t *testing.T) {
		user := &User{
			ID:       types.BigInt(1),
			Username: "testuser",
			Nickname: "测试用户",
			Avatar:   "avatar.jpg",
			Mobile:   "13800138000",
			Gender:   1,
			Email:    "test@example.com",
			Status:   1,
		}

		vo := ToUserVO(user)

		if vo == nil {
			t.Fatal("ToUserVO() 返回 nil")
		}

		if vo.ID != user.ID {
			t.Errorf("ID 不匹配，期望 %v，实际 %v", user.ID, vo.ID)
		}

		if vo.Username != user.Username {
			t.Errorf("Username 不匹配，期望 %v，实际 %v", user.Username, vo.Username)
		}

		if vo.Nickname != user.Nickname {
			t.Errorf("Nickname 不匹配，期望 %v，实际 %v", user.Nickname, vo.Nickname)
		}
	})

	// 测试 nil 输入
	t.Run("nil 输入", func(t *testing.T) {
		vo := ToUserVO(nil)
		if vo != nil {
			t.Errorf("ToUserVO(nil) 应该返回 nil，实际返回 %v", vo)
		}
	})
}

// TestToUser 测试 UserForm 转 User
func TestToUser(t *testing.T) {
	t.Run("正常转换", func(t *testing.T) {
		form := &UserForm{
			ID:       types.BigInt(1),
			Username: "testuser",
			Nickname: "测试用户",
			Mobile:   "13800138000",
			Gender:   1,
			Email:    "test@example.com",
			Status:   1,
			DeptID:   types.BigInt(10),
		}

		user := ToUser(form)

		if user == nil {
			t.Fatal("ToUser() 返回 nil")
		}

		if user.ID != form.ID {
			t.Errorf("ID 不匹配")
		}

		if user.Username != form.Username {
			t.Errorf("Username 不匹配")
		}

		if user.DeptID != form.DeptID {
			t.Errorf("DeptID 不匹配")
		}
	})

	t.Run("nil 输入", func(t *testing.T) {
		user := ToUser(nil)
		if user != nil {
			t.Errorf("ToUser(nil) 应该返回 nil")
		}
	})
}

// TestToUserVOList 测试批量转换
func TestToUserVOList(t *testing.T) {
	t.Run("正常批量转换", func(t *testing.T) {
		users := []*User{
			{ID: types.BigInt(1), Username: "user1", Nickname: "用户1"},
			{ID: types.BigInt(2), Username: "user2", Nickname: "用户2"},
			{ID: types.BigInt(3), Username: "user3", Nickname: "用户3"},
		}

		vos := ToUserVOList(users)

		if len(vos) != len(users) {
			t.Errorf("转换后数量不匹配，期望 %d，实际 %d", len(users), len(vos))
		}

		for i, vo := range vos {
			if vo.ID != users[i].ID {
				t.Errorf("第 %d 个元素 ID 不匹配", i)
			}
			if vo.Username != users[i].Username {
				t.Errorf("第 %d 个元素 Username 不匹配", i)
			}
		}
	})

	t.Run("空列表", func(t *testing.T) {
		vos := ToUserVOList([]*User{})
		if len(vos) != 0 {
			t.Errorf("空列表转换应该返回空列表")
		}
	})

	t.Run("nil 输入", func(t *testing.T) {
		vos := ToUserVOList(nil)
		if vos != nil {
			t.Errorf("ToUserVOList(nil) 应该返回 nil")
		}
	})
}

// 运行测试：
// go test -v ./internal/system/user/model/
