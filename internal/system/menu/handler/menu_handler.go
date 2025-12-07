package handler

import (
	"strconv"
	"github.com/gin-gonic/gin"
	"youlai-gin/internal/system/menu/model"
	"youlai-gin/internal/system/menu/service"
	pkgContext "youlai-gin/pkg/context"
	"youlai-gin/pkg/response"
	"youlai-gin/pkg/validator"
)

func RegisterMenuRoutes(r *gin.RouterGroup) {
	menus := r.Group("/menus")
	{
		menus.GET("", GetMenuList)
		menus.GET("/options", GetMenuOptions)
		menus.GET("/routes", GetCurrentUserRoutes)
		menus.POST("", SaveMenu)
		menus.GET("/:id/form", GetMenuForm)
		menus.PUT("/:id", UpdateMenu)
		menus.DELETE("/:id", DeleteMenu)
	}
	
	// 用户权限接口
	r.GET("/user/perms", GetCurrentUserPermissions)
}

// @Summary 菜单列表
// @Tags 菜单管理
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/menus [get]
func GetMenuList(c *gin.Context) {
	var query model.MenuQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	list, err := service.GetMenuList(&query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, list)
}

// @Summary 菜单下拉列表
// @Tags 菜单管理
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/menus/options [get]
func GetMenuOptions(c *gin.Context) {
	onlyParent := c.Query("onlyParent") == "true"

	options, err := service.GetMenuOptions(onlyParent)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, options)
}

// @Summary 获取当前用户路由
// @Tags 菜单管理
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/menus/routes [get]
func GetCurrentUserRoutes(c *gin.Context) {
	userId := int64(1)

	routes, err := service.GetCurrentUserRoutes(userId)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, routes)
}

// @Summary 新增菜单
// @Tags 菜单管理
// @Param body body model.MenuForm true "菜单信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/menus [post]
func SaveMenu(c *gin.Context) {
	var form model.MenuForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := service.SaveMenu(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "保存成功")
}

// @Summary 获取菜单表单数据
// @Tags 菜单管理
// @Param id path int true "菜单ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/menus/{id}/form [get]
func GetMenuForm(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的菜单ID")
		return
	}

	form, err := service.GetMenuForm(id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, form)
}

// @Summary 更新菜单
// @Tags 菜单管理
// @Param id path int true "菜单ID"
// @Param body body model.MenuForm true "菜单信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/menus/{id} [put]
func UpdateMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的菜单ID")
		return
	}

	var form model.MenuForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	form.ID = id
	if err := service.SaveMenu(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "更新成功")
}

// @Summary 删除菜单
// @Tags 菜单管理
// @Param id path int true "菜单ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/menus/{id} [delete]
func DeleteMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的菜单ID")
		return
	}

	if err := service.DeleteMenu(id); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "删除成功")
}

// @Summary 获取当前用户权限（按钮权限）
// @Tags 菜单管理
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/user/perms [get]
func GetCurrentUserPermissions(c *gin.Context) {
	userId, err := pkgContext.GetUserIDMust(c)
	if err != nil {
		response.Unauthorized(c, "未登录")
		return
	}

	perms, err := service.GetUserPermissions(userId)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, perms)
}
