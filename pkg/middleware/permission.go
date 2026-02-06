package middleware

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/system/permission/service"
	pkgContext "youlai-gin/pkg/context"
	"youlai-gin/pkg/response"
)

// PermissionMiddleware 权限验证中间件（全局）
func PermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID
		userID, err := pkgContext.GetCurrentUserID(c)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		// 获取用户权限信息（会使用缓存）
		perms, err := service.GetUserPermissions(userID)
		if err != nil {
			response.InternalServerError(c, "获取用户权限失败")
			c.Abort()
			return
		}

		// 将权限信息存入上下文，供后续使用
		c.Set("userPermissions", perms)

		c.Next()
	}
}

// RequirePermission 需要指定权限（按钮级）
func RequirePermission(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := pkgContext.GetCurrentUserID(c)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		// 检查权限
		hasPermission, err := service.CheckPermission(userID, perm)
		if err != nil {
			response.InternalServerError(c, "权限检查失败")
			c.Abort()
			return
		}

		if !hasPermission {
			response.Unauthorized(c, "无权限访问")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission 需要任意一个权限
func RequireAnyPermission(perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := pkgContext.GetCurrentUserID(c)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		// 检查是否有任意一个权限
		hasPermission, err := service.CheckAnyPermission(userID, perms)
		if err != nil {
			response.InternalServerError(c, "权限检查失败")
			c.Abort()
			return
		}

		if !hasPermission {
			response.Unauthorized(c, "无权限访问")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAllPermissions 需要所有权限
func RequireAllPermissions(perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := pkgContext.GetCurrentUserID(c)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		// 检查是否拥有所有权限
		hasAllPermissions, err := service.CheckAllPermissions(userID, perms)
		if err != nil {
			response.InternalServerError(c, "权限检查失败")
			c.Abort()
			return
		}

		if !hasAllPermissions {
			response.Unauthorized(c, "无权限访问")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole 需要指定角色
func RequireRole(roleCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := pkgContext.GetCurrentUserID(c)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		// 检查角色
		hasRole, err := service.CheckRole(userID, roleCode)
		if err != nil {
			response.InternalServerError(c, "角色检查失败")
			c.Abort()
			return
		}

		if !hasRole {
			response.Unauthorized(c, "无权限访问")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole 需要任意一个角色
func RequireAnyRole(roleCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := pkgContext.GetCurrentUserID(c)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}

		// 检查是否有任意一个角色
		hasAnyRole, err := service.CheckAnyRole(userID, roleCodes)
		if err != nil {
			response.InternalServerError(c, "角色检查失败")
			c.Abort()
			return
		}

		if !hasAnyRole {
			response.Unauthorized(c, "无权限访问")
			c.Abort()
			return
		}

		c.Next()
	}
}
