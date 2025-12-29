package handler

import (
	"strconv"
	"github.com/gin-gonic/gin"
	"youlai-gin/internal/system/role/model"
	"youlai-gin/internal/system/role/service"
	"youlai-gin/pkg/response"
	"youlai-gin/pkg/types"
	"youlai-gin/pkg/validator"
)

func RegisterRoleRoutes(r *gin.RouterGroup) {
	roles := r.Group("/roles")
	{
		roles.GET("/page", GetRolePage)
		roles.GET("/options", GetRoleOptions)
		roles.POST("", SaveRole)
		roles.GET("/:id/form", GetRoleForm)
		roles.PUT("/:id", UpdateRole)
		roles.DELETE("/:id", DeleteRole)
		roles.GET("/:id/menu-ids", GetRoleMenuIds)
		roles.PUT("/:id/menus", UpdateRoleMenus)
	}
}

// @Summary 角色分页列表
// @Tags 角色管理
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keywords query string false "关键字"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/page [get]
func GetRolePage(c *gin.Context) {
	var query model.RolePageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	result, err := service.GetRolePage(&query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, result)
}

// @Summary 角色下拉列表
// @Tags 角色管理
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/options [get]
func GetRoleOptions(c *gin.Context) {
	options, err := service.GetRoleOptions()
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, options)
}

// @Summary 新增角色
// @Tags 角色管理
// @Param body body model.RoleForm true "角色信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles [post]
func SaveRole(c *gin.Context) {
	var form model.RoleForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := service.SaveRole(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "保存成功")
}

// @Summary 获取角色表单数据
// @Tags 角色管理
// @Param id path int true "角色ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/{id}/form [get]
func GetRoleForm(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	form, err := service.GetRoleForm(id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, form)
}

// @Summary 更新角色
// @Tags 角色管理
// @Param id path int true "角色ID"
// @Param body body model.RoleForm true "角色信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/{id} [put]
func UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	var form model.RoleForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	form.ID = types.BigInt(id)
	if err := service.SaveRole(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "更新成功")
}

// @Summary 删除角色
// @Tags 角色管理
// @Param id path int true "角色ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/{id} [delete]
func DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	if err := service.DeleteRole(id); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "删除成功")
}

// @Summary 获取角色菜单ID列表
// @Tags 角色管理
// @Param id path int true "角色ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/{id}/menu-ids [get]
func GetRoleMenuIds(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	menuIds, err := service.GetRoleMenuIds(id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, menuIds)
}

// @Summary 分配菜单权限
// @Tags 角色管理
// @Param id path int true "角色ID"
// @Param body body []int64 true "菜单ID列表"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/{id}/menus [put]
func UpdateRoleMenus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	var menuIds []int64
	if err := c.ShouldBindJSON(&menuIds); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := service.UpdateRoleMenus(id, menuIds); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "分配成功")
}
