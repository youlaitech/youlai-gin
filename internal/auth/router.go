package auth

import (
	"github.com/gin-gonic/gin"

	authHandler "youlai-gin/internal/auth/handler"
	authService "youlai-gin/internal/auth/service"
	pkgAuth "youlai-gin/internal/common/auth"
	"youlai-gin/internal/common/database"
	userRepo "youlai-gin/internal/system/user/repository"
)

// RegisterRoutes 装配认证模块（账号密码 / 扫码 / 微信小程序）路由。
// 所有子域均走 repository → service → handler 的依赖注入，对上层 router 保持唯一入口。
func RegisterRoutes(api *gin.RouterGroup, tokenManager pkgAuth.TokenManager) {
	db := database.DB
	userRepo := userRepo.NewRepository(db)

	// 初始化微信配置（wxma service 读取配置发起微信 API 调用）
	authService.InitWechatConfig()

	// 令牌签发组件：三种登录方式共用「角色 → 数据权限 → 签发」流程
	issuer := authService.NewTokenIssuer(tokenManager, userRepo)

	authHandler.NewAuthHandler(authService.NewAuthService(tokenManager, issuer, userRepo)).RegisterRoutes(api)

	// 扫码登录：tokenManager 供 scan/confirm/cancel 的 JWT 鉴权中间件使用
	authHandler.NewQrCodeHandler(authService.NewQrCodeService(issuer, userRepo)).RegisterRoutes(api, tokenManager)

	// 微信小程序认证（去裸 db：事务写 user_social 由 user.Repository 封装）
	authHandler.NewWxMaHandler(authService.NewWxMaService(issuer, userRepo)).RegisterRoutes(api)
}
