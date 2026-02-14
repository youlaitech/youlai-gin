package handler

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"youlai-gin/internal/system/user/api"
	"youlai-gin/internal/system/user/service"
	"youlai-gin/pkg/errs"
	pkgContext "youlai-gin/pkg/context"
	"youlai-gin/pkg/response"
	"youlai-gin/pkg/types"
)

// RegisterUserRoutes 注册用户路由
func RegisterUserRoutes(r *gin.RouterGroup) {
	r.GET("/users", GetUserList)
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
	r.DELETE("/users/mobile", UnbindMobile)
	r.POST("/users/email/code", SendEmailCode)
	r.PUT("/users/email", BindOrChangeEmail)
	r.DELETE("/users/email", UnbindEmail)

	// Excel 导入导出
	r.GET("/users/export", ExportUsers)
	r.GET("/users/template", DownloadUserTemplate)
	r.POST("/users/import", ImportUsers)
	r.GET("/users/options", GetUserOptions)
}

// GetUserList 用户分页列表
// @Summary 用户分页列表
// @Tags 02.用户接口
// @Router /api/v1/users [get]
func GetUserList(c *gin.Context) {
	var query api.UserQueryReq
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	currentUser, err := pkgContext.GetCurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := service.GetUserPage(&query, currentUser)
	if err != nil {
		c.Error(err)
		return
	}

	response.OkPaged(c, result)
}

// SaveUser 保存用户（新增或更新）
// @Summary 保存用户
// @Tags 02.用户接口
// @Router /api/v1/users [post]
func SaveUser(c *gin.Context) {
	var form api.UserSaveReq
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
// @Summary 获取用户表单
// @Tags 02.用户接口
// @Param userId path int true "用户ID"
// @Router /api/v1/users/{userId}/form [get]
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
// @Summary 更新用户
// @Tags 02.用户接口
// @Param userId path int true "用户ID"
// @Router /api/v1/users/{userId} [put]
func UpdateUser(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, _ := strconv.ParseInt(userIdStr, 10, 64)

	var form api.UserSaveReq
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	form.ID = types.BigInt(userId)
	if err := service.SaveUser(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "修改成功")
}

// DeleteUsers 删除用户
// @Summary 删除用户
// @Tags 02.用户接口
// @Param ids path string true "用户ID列表"
// @Router /api/v1/users/{ids} [delete]
func DeleteUsers(c *gin.Context) {
	ids := c.Param("ids")

	if err := service.DeleteUsers(ids); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "删除成功")
}

// UpdateUserStatus 修改用户状态
// @Summary 修改用户状态
// @Tags 02.用户接口
// @Param userId path int true "用户ID"
// @Router /api/v1/users/{userId}/status [patch]
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
// @Summary 当前登录用户
// @Tags 02.用户接口
// @Router /api/v1/users/me [get]
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
// @Summary 个人中心信息
// @Tags 02.用户接口
// @Router /api/v1/users/profile [get]
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
// @Summary 更新个人中心信息
// @Tags 02.用户接口
// @Router /api/v1/users/profile [put]
func UpdateUserProfile(c *gin.Context) {
	userId, err := pkgContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var form api.UserProfileUpdateReq
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	if err := service.UpdateUserProfile(userId, &form); err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, true)
}

// ResetUserPassword 重置指定用户密码
// @Summary 重置用户密码
// @Tags 02.用户接口
// @Param userId path int true "用户ID"
// @Router /api/v1/users/{userId}/password/reset [put]
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
// @Summary 修改当前用户密码
// @Tags 02.用户接口
// @Router /api/v1/users/password [put]
func ChangeCurrentUserPassword(c *gin.Context) {
	userId, err := pkgContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var form api.PasswordUpdateReq
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	if err := service.ChangeUserPassword(userId, &form); err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, true)
}

// SendMobileCode 发送短信验证码
// @Summary 发送手机号验证码
// @Tags 02.用户接口
// @Router /api/v1/users/mobile/code [post]
func SendMobileCode(c *gin.Context) {
	mobile := c.Query("mobile")

	if err := service.SendMobileCode(mobile); err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, true)
}

// BindOrChangeMobile 绑定或更换手机号
// @Summary 绑定或更换手机号
// @Tags 02.用户接口
// @Router /api/v1/users/mobile [put]
func BindOrChangeMobile(c *gin.Context) {
	userId, err := pkgContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var form api.MobileUpdateReq
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	if err := service.BindOrChangeMobile(userId, &form); err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, true)
}

// UnbindMobile 解绑手机号
// @Summary 解绑手机号
// @Tags 02.用户接口
// @Router /api/v1/users/mobile [delete]
func UnbindMobile(c *gin.Context) {
	userId, err := pkgContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var form api.PasswordVerifyReq
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	if err := service.UnbindMobile(userId, &form); err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, true)
}

// SendEmailCode 发送邮箱验证码
// @Summary 发送邮箱验证码
// @Tags 02.用户接口
// @Router /api/v1/users/email/code [post]
func SendEmailCode(c *gin.Context) {
	email := c.Query("email")

	if err := service.SendEmailCode(email); err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, true)
}

// BindOrChangeEmail 绑定或更换邮箱
// @Summary 绑定或更换邮箱
// @Tags 02.用户接口
// @Router /api/v1/users/email [put]
func BindOrChangeEmail(c *gin.Context) {
	userId, err := pkgContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var form api.EmailUpdateReq
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	if err := service.BindOrChangeEmail(userId, &form); err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, true)
}

// UnbindEmail 解绑邮箱
// @Summary 解绑邮箱
// @Tags 02.用户接口
// @Router /api/v1/users/email [delete]
func UnbindEmail(c *gin.Context) {
	userId, err := pkgContext.GetCurrentUserID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var form api.PasswordVerifyReq
	if err := c.ShouldBindJSON(&form); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	if err := service.UnbindEmail(userId, &form); err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, true)
}

// GetUserOptions 获取用户下拉选项
// @Summary 用户下拉选项
// @Tags 02.用户接口
// @Router /api/v1/users/options [get]
func GetUserOptions(c *gin.Context) {
	options, err := service.GetUserOptions()
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, options)
}

// ExportUsers 导出用户列表
// @Summary 导出用户
// @Tags 02.用户接口
// @Router /api/v1/users/export [get]
func ExportUsers(c *gin.Context) {
	// 绑定查询参数（支持过滤条件）
	var query api.UserQueryReq
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	currentUser, err := pkgContext.GetCurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	// 导出用户数据
	exporter, err := service.ExportUsersToExcel(&query, currentUser)
	if err != nil {
		c.Error(err)
		return
	}
	defer exporter.Close()

	// 设置响应头
	filename := fmt.Sprintf("用户列表_%s.xlsx", time.Now().Format("20060102150405"))
	encodedFilename := url.QueryEscape(filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", filename, encodedFilename))
	c.Header("Content-Transfer-Encoding", "binary")

	// 输出文件
	if err := exporter.Write(c.Writer); err != nil {
		c.Error(errs.SystemError("导出文件失败"))
		return
	}
}

// DownloadUserTemplate 下载用户导入模板
// @Summary 下载用户导入模板
// @Tags 02.用户接口
// @Router /api/v1/users/template [get]
func DownloadUserTemplate(c *gin.Context) {
	exporter, err := service.GenerateUserTemplate()
	if err != nil {
		c.Error(err)
		return
	}
	defer exporter.Close()

	// 设置响应头
	filename := "用户导入模板.xlsx"
	encodedFilename := url.QueryEscape(filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", filename, encodedFilename))
	c.Header("Content-Transfer-Encoding", "binary")

	// 输出文件
	if err := exporter.Write(c.Writer); err != nil {
		c.Error(errs.SystemError("生成模板失败"))
		return
	}
}

// ImportUsers 导入用户数据
// @Summary 导入用户
// @Tags 02.用户接口
// @Router /api/v1/users/import [post]
func ImportUsers(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, "请选择要导入的文件")
		return
	}

	// 打开文件
	f, err := file.Open()
	if err != nil {
		response.Fail(c, "文件打开失败")
		return
	}
	defer f.Close()

	// 导入用户数据
	result, err := service.ImportUsersFromExcel(f)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}
