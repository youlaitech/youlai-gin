package middleware

import (
	"github.com/gin-gonic/gin"
	
	"youlai-gin/internal/system/permission/service"
	"youlai-gin/pkg/context"
	"youlai-gin/pkg/response"
)

// PermissionMiddleware 权限验证中间件（全局应用）
// 注意：此中间件只做通用权限检查，具体权限通过 RequirePermission 注解控制
func PermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID
		userID, exists := context.GetUserID(c)
		if !exists {
			response.Unauthorized(c, "未获取到用户信息")
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

// RequirePermission 需要指定权限（按钮级权限）
// 用法：router.POST("/users", middleware.RequirePermission("sys:user:add"), handler.CreateUser)
func RequirePermission(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := context.GetUserID(c)
		if !exists {
			response.Unauthorized(c, "未获取到用户信息")
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
			response.Forbidden(c, "无权限访问")
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequireAnyPermission 需要任意一个权限
// 用法：router.POST("/users", middleware.RequireAnyPermission("sys:user:add", "sys:user:edit"), handler.SaveUser)
func RequireAnyPermission(perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := context.GetUserID(c)
		if !exists {
			response.Unauthorized(c, "未获取到用户信息")
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
			response.Forbidden(c, "无权限访问")
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequireAllPermissions 需要所有权限
// 用法：router.POST("/users/batch", middleware.RequireAllPermissions("sys:user:add", "sys:user:delete"), handler.BatchUser)
func RequireAllPermissions(perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := context.GetUserID(c)
		if !exists {
			response.Unauthorized(c, "未获取到用户信息")
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
			response.Forbidden(c, "无权限访问")
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequireRole 需要指定角色
// 用法：router.GET("/admin", middleware.RequireRole("ADMIN"), handler.AdminPanel)
func RequireRole(roleCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := context.GetUserID(c)
		if !exists {
			response.Unauthorized(c, "未获取到用户信息")
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
			response.Forbidden(c, "无权限访问")
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequireAnyRole 需要任意一个角色
// 用法：router.GET("/manage", middleware.RequireAnyRole("ADMIN", "MANAGER"), handler.Manage)
func RequireAnyRole(roleCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := context.GetUserID(c)
		if !exists {
			response.Unauthorized(c, "未获取到用户信息")
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
			response.Forbidden(c, "无权限访问")
			c.Abort()
			return
		}
		
		c.Next()
	}
}
