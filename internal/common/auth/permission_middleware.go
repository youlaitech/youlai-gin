package auth

import (
	"github.com/gin-gonic/gin"
	permService "youlai-gin/internal/common/permission/service"
	"youlai-gin/pkg/errs"

	"net/http"
)

// getCurrentUserID 从当前请求上下文获取已认证用户的ID
func getCurrentUserID(c *gin.Context) (int64, error) {
	user, exists := GetCurrentUser(c)
	if !exists {
		return 0, errs.Unauthorized("未登录或登录已过期")
	}
	return user.UserID, nil
}

// RequirePermission 权限校验中间件
// 用法: users.GET("", RequirePermission("sys:user:list"), handler.ListUser)
func RequirePermission(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getCurrentUserID(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errs.Unauthorized("未登录"))
			return
		}

		hasPerm, err := permService.CheckPermission(userID, perm)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.SystemError("权限校验失败"))
			return
		}
		if !hasPerm {
			c.AbortWithStatusJSON(http.StatusForbidden, errs.Forbidden("无操作权限"))
			return
		}

		c.Next()
	}
}

// RequireAnyPermission 需要拥有任意一个权限
func RequireAnyPermission(perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getCurrentUserID(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errs.Unauthorized("未登录"))
			return
		}

		hasAny, err := permService.CheckAnyPermission(userID, perms)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.SystemError("权限校验失败"))
			return
		}
		if !hasAny {
			c.AbortWithStatusJSON(http.StatusForbidden, errs.Forbidden("无操作权限"))
			return
		}

		c.Next()
	}
}

// RequireRole 角色校验中间件
func RequireRole(roleCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getCurrentUserID(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errs.Unauthorized("未登录"))
			return
		}

		hasRole, err := permService.CheckRole(userID, roleCode)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errs.SystemError("角色校验失败"))
			return
		}
		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, errs.Forbidden("无操作权限"))
			return
		}

		c.Next()
	}
}
