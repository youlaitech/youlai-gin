package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"youlai-gin/internal/database"
	"youlai-gin/pkg/redis"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string            `json:"status"`    // 状态：healthy, unhealthy
	Timestamp int64             `json:"timestamp"` // 时间戳
	Services  map[string]string `json:"services"`  // 服务状态
	Version   string            `json:"version"`   // 版本号
}

// RegisterRoutes 注册健康检查路由
func RegisterRoutes(router *gin.Engine) {
	router.GET("/api/v1/health", HealthCheck)
	router.GET("/health", HealthCheck) // 兼容路径
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Tags 系统
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /api/v1/health [get]
func HealthCheck(c *gin.Context) {
	services := make(map[string]string)
	status := "healthy"

	// 检查数据库连接
	if db, err := database.DB.DB(); err == nil {
		if err := db.Ping(); err == nil {
			services["database"] = "healthy"
		} else {
			services["database"] = "unhealthy: " + err.Error()
			status = "unhealthy"
		}
	} else {
		services["database"] = "unhealthy: " + err.Error()
		status = "unhealthy"
	}

	// 检查 Redis 连接
	if err := redis.Client.Ping(c).Err(); err == nil {
		services["redis"] = "healthy"
	} else {
		services["redis"] = "unhealthy: " + err.Error()
		status = "unhealthy"
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now().Unix(),
		Services:  services,
		Version:   "1.0.0", // 可以从配置或编译参数中获取
	}

	// 如果不健康，返回 503 状态码
	if status == "unhealthy" {
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}
