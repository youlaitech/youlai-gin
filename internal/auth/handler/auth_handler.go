package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"youlai-gin/internal/auth/model"
	"youlai-gin/internal/auth/service"
	pkgAuth "youlai-gin/pkg/auth"
	"youlai-gin/pkg/response"
	"youlai-gin/pkg/validator"
)

// RegisterAuthRoutes 注册认证相关 HTTP 路由
func RegisterAuthRoutes(r *gin.RouterGroup) {
	r.POST("/auth/login", Login)
	r.DELETE("/auth/logout", Logout)
	r.POST("/auth/refresh-token", RefreshToken)
}

// Login 账号密码登录
// @Summary 账号密码登录
// @Description 用户名密码登录，返回访问令牌和刷新令牌
// @Tags 认证中心
// @Accept json
// @Produce json
// @Param body body model.LoginRequest true "登录信息"
// @Success 200 {object} map[string]interface{} "code/msg/data，data 为 AuthenticationToken"
// @Router /api/v1/auth/login [post]
func Login(c *gin.Context) {
	var req model.LoginRequest
	if err := validator.BindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}

	token, err := service.Login(&req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, token)
}

// Logout 退出登录
// @Summary 退出登录
// @Description 使当前访问令牌失效
// @Tags 认证中心
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "code/msg"
// @Router /api/v1/auth/logout [delete]
func Logout(c *gin.Context) {
	// 从 Header 中获取 Token
	authHeader := c.GetHeader(pkgAuth.AuthorizationHeader)
	token := strings.TrimPrefix(authHeader, pkgAuth.BearerPrefix)

	if err := service.Logout(token); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "退出成功")
}

// RefreshToken 刷新令牌
// @Summary 刷新令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证中心
// @Accept json
// @Produce json
// @Param refreshToken query string true "刷新令牌"
// @Success 200 {object} map[string]interface{} "code/msg/data，data 为 AuthenticationToken"
// @Router /api/v1/auth/refresh-token [post]
func RefreshToken(c *gin.Context) {
	refreshToken := c.Query("refreshToken")
	if refreshToken == "" {
		response.BadRequest(c, "刷新令牌不能为空")
		return
	}

	token, err := service.RefreshToken(refreshToken)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, token)
}
