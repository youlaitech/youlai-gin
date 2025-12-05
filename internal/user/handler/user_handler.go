package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"youlai-gin/internal/user/model"
	"youlai-gin/internal/user/service"
	"youlai-gin/pkg/response"
	"youlai-gin/pkg/validator"
)

// RegisterUserRoutes 注册用户相关 HTTP 路由
func RegisterUserRoutes(r *gin.RouterGroup) {
	// 将路由与具名处理函数绑定，方便 Swag 扫描注释
	r.GET("/users/page", ListUsers)
	r.POST("/users", CreateUser)
	r.PUT("/users/:id", UpdateUser)
	r.DELETE("/users/:id", DeleteUser)
}

// ListUsers 用户列表
// @Summary 用户列表
// @Description 简单返回所有用户（示例，不带分页参数）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "code/msg/data 格式，data.list 为用户列表"
// @Router /api/v1/users/page [get]
func ListUsers(c *gin.Context) {
	users, err := service.ListUsers()
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, gin.H{
		"list":  users,
		"total": len(users),
	})
}

// CreateUser 新增用户
// @Summary 新增用户
// @Description 创建一个新的用户记录
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param body body model.User true "用户信息"
// @Success 200 {object} map[string]interface{} "code/msg"
// @Router /api/v1/users [post]
func CreateUser(c *gin.Context) {
	var req model.User
	if err := validator.BindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}

	if err := service.CreateUser(&req); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "创建成功")
}

// UpdateUser 修改用户
// @Summary 修改用户
// @Description 根据 ID 修改用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path uint64 true "用户ID"
// @Param body body model.User true "用户信息"
// @Success 200 {object} map[string]interface{} "code/msg"
// @Router /api/v1/users/{id} [put]
func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)

	var req model.User
	if err := validator.BindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}

	if err := service.UpdateUser(id, &req); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "修改成功")
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 根据 ID 删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path uint64 true "用户ID"
// @Success 200 {object} map[string]interface{} "code/msg"
// @Router /api/v1/users/{id} [delete]
func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)

	if err := service.DeleteUser(id); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "删除成功")
}
