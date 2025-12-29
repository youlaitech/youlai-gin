package handler

import (
	"strconv"
	"youlai-gin/internal/system/dept/model"
	"youlai-gin/internal/system/dept/service"
	"youlai-gin/pkg/response"
	"youlai-gin/pkg/types"
	"youlai-gin/pkg/validator"

	"github.com/gin-gonic/gin"
)

func RegisterDeptRoutes(r *gin.RouterGroup) {
	// 使用复数形式符合RESTful规范
	depts := r.Group("/depts")
	{
		depts.GET("", GetDeptList)
		depts.GET("/options", GetDeptOptions)
		depts.POST("", SaveDept)
		depts.GET("/:id/form", GetDeptForm)
		depts.PUT("/:id", UpdateDept)
		depts.DELETE("/:id", DeleteDept)
	}
}

// @Summary 部门列表
// @Tags 部门管理
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/depts [get]
func GetDeptList(c *gin.Context) {
	var query model.DeptQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	list, err := service.GetDeptList(&query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, list)
}

// @Summary 部门下拉列表
// @Tags 部门管理
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/depts/options [get]
func GetDeptOptions(c *gin.Context) {
	options, err := service.GetDeptOptions()
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, options)
}

// @Summary 新增部门
// @Tags 部门管理
// @Param body body model.DeptForm true "部门信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/depts [post]
func SaveDept(c *gin.Context) {
	var form model.DeptForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	if err := service.SaveDept(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "保存成功")
}

// @Summary 获取部门表单数据
// @Tags 部门管理
// @Param id path int true "部门ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/depts/{id}/form [get]
func GetDeptForm(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的部门ID")
		return
	}

	form, err := service.GetDeptForm(id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, form)
}

// @Summary 更新部门
// @Tags 部门管理
// @Param id path int true "部门ID"
// @Param body body model.DeptForm true "部门信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/depts/{id} [put]
func UpdateDept(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的部门ID")
		return
	}

	var form model.DeptForm
	if err := validator.BindJSON(c, &form); err != nil {
		c.Error(err)
		return
	}

	form.ID = types.BigInt(id)
	if err := service.SaveDept(&form); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "更新成功")
}

// @Summary 删除部门
// @Tags 部门管理
// @Param id path int true "部门ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/depts/{id} [delete]
func DeleteDept(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的部门ID")
		return
	}

	if err := service.DeleteDept(id); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "删除成功")
}
