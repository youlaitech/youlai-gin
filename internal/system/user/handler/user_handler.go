package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"youlai-gin/internal/system/user/model"
	"youlai-gin/internal/system/user/service"
	pkgContext "youlai-gin/pkg/context"
	"youlai-gin/pkg/response"
)

// RegisterUserRoutes 注册用户路由
func RegisterUserRoutes(r *gin.RouterGroup) {
	r.GET("/users/page", GetUserPage)
	r.POST("/users", SaveUser)
	r.GET("/users/:userId/form", GetUserForm)
	r.PUT("/users/:userId", UpdateUser)
	r.DELETE("/users/:ids", DeleteUsers)
	r.PATCH("/users/:userId/status", UpdateUserStatus)
	r.GET("/users/me", GetCurrentUser)
	r.GET("/users/profile", GetUserProfile)
	r.PUT("/users/profile", UpdateUserProfile)
	r.PUT("/users/:userId/password/reset", ResetUserPassword)
	r.PUT("/users/password", ChangeCurrentUserPassword)
	r.POST("/users/mobile/code", SendMobileCode)
	r.PUT("/users/mobile", BindOrChangeMobile)
	r.POST("/users/email/code", SendEmailCode)
	r.PUT("/users/email", BindOrChangeEmail)
	
	// Excel 导入导出
	r.GET("/users/export", ExportUsers)
	r.GET("/users/template", DownloadUserTemplate)
	r.POST("/users/import", ImportUsers)
	r.GET("/users/options", GetUserOptions)
}

// GetUserPage 用户分页列表
func GetUserPage(c *gin.Context) {
	var query model.UserPageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	result, err := service.GetUserPage(&query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}

// SaveUser 保存用户（新增或更新）
func SaveUser(c *gin.Context) {
	var form model.UserForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	if err := service.SaveUser(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "保存成功")
}

// GetUserForm 获取用户表单数据
func GetUserForm(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, _ := strconv.ParseInt(userIdStr, 10, 64)

	formData, err := service.GetUserForm(userId)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, formData)
}

// UpdateUser 修改用户
func UpdateUser(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, _ := strconv.ParseInt(userIdStr, 10, 64)

	var form model.UserForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	form.ID = userId
	if err := service.SaveUser(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "修改成功")
}

// DeleteUsers 删除用户
func DeleteUsers(c *gin.Context) {
	ids := c.Param("ids")

	if err := service.DeleteUsers(ids); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "删除成功")
}

// UpdateUserStatus 修改用户状态
func UpdateUserStatus(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, _ := strconv.ParseInt(userIdStr, 10, 64)

	statusStr := c.Query("status")
	status, _ := strconv.Atoi(statusStr)

	if err := service.UpdateUserStatus(userId, status); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "修改成功")
}

// GetCurrentUser 获取当前登录用户信息
func GetCurrentUser(c *gin.Context) {
	// 从token中获取用户详情（包含角色信息）
	userDetails, err := pkgContext.GetCurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	// 使用token中的角色信息获取用户详情和权限
	currentUser, err := service.GetCurrentUserInfoWithRoles(userDetails.UserID, userDetails.Roles)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, currentUser)
}

// GetUserProfile 获取个人中心用户信息
func GetUserProfile(c *gin.Context) {
	userId, err := pkgContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	profile, err := service.GetUserProfile(userId)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, profile)
}

// UpdateUserProfile 个人中心修改用户信息
func UpdateUserProfile(c *gin.Context) {
	userId, err := pkgContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var form model.UserProfileForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	if err := service.UpdateUserProfile(userId, &form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "修改成功")
}

// ResetUserPassword 重置指定用户密码
func ResetUserPassword(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, _ := strconv.ParseInt(userIdStr, 10, 64)

	password := c.Query("password")

	if err := service.ResetUserPassword(userId, password); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "重置成功")
}

// ChangeCurrentUserPassword 当前用户修改密码
func ChangeCurrentUserPassword(c *gin.Context) {
	userId, err := pkgContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var form model.PasswordUpdateForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	if err := service.ChangeUserPassword(userId, &form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "修改成功")
}

// SendMobileCode 发送短信验证码
func SendMobileCode(c *gin.Context) {
	mobile := c.Query("mobile")

	if err := service.SendMobileCode(mobile); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "发送成功")
}

// BindOrChangeMobile 绑定或更换手机号
func BindOrChangeMobile(c *gin.Context) {
	userId, err := pkgContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var form model.MobileUpdateForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	if err := service.BindOrChangeMobile(userId, &form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "绑定成功")
}

// SendEmailCode 发送邮箱验证码
func SendEmailCode(c *gin.Context) {
	email := c.Query("email")

	if err := service.SendEmailCode(email); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "发送成功")
}

// BindOrChangeEmail 绑定或更换邮箱
func BindOrChangeEmail(c *gin.Context) {
	userId, err := pkgContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var form model.EmailUpdateForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	if err := service.BindOrChangeEmail(userId, &form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "绑定成功")
}

// GetUserOptions 获取用户下拉选项
func GetUserOptions(c *gin.Context) {
	options, err := service.GetUserOptions()
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, options)
}

// ExportUsers 导出用户列表
func ExportUsers(c *gin.Context) {
	// TODO: 实现 Excel 导出功能
	response.OkMsg(c, "导出功能待实现")
}

// DownloadUserTemplate 下载用户导入模板
func DownloadUserTemplate(c *gin.Context) {
	// TODO: 实现模板下载功能
	response.OkMsg(c, "模板下载功能待实现")
}

// ImportUsers 导入用户数据
func ImportUsers(c *gin.Context) {
	// TODO: 实现 Excel 导入功能
	response.OkMsg(c, "导入功能待实现")
}
