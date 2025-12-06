package context

import (
	"github.com/gin-gonic/gin"
	"youlai-gin/pkg/auth"
	"youlai-gin/pkg/errs"
)

// GetCurrentUserID 从上下文获取当前用户ID
func GetCurrentUserID(c *gin.Context) (int64, error) {
	user, exists := auth.GetCurrentUser(c)
	if !exists {
		return 0, errs.Unauthorized("未登录或登录已过期")
	}
	return user.UserID, nil
}

// GetCurrentUser 从上下文获取当前用户详情
func GetCurrentUser(c *gin.Context) (*auth.UserDetails, error) {
	user, exists := auth.GetCurrentUser(c)
	if !exists {
		return nil, errs.Unauthorized("未登录或登录已过期")
	}
	return user, nil
}

// MustGetCurrentUserID 从上下文获取当前用户ID（必须存在，否则panic）
func MustGetCurrentUserID(c *gin.Context) int64 {
	userID, err := GetCurrentUserID(c)
	if err != nil {
		panic(err)
	}
	return userID
}
