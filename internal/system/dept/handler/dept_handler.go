package handler

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/internal/system/dept/model"
	"youlai-gin/internal/system/dept/service"
	pkgContext "youlai-gin/internal/common/context"
	"youlai-gin/pkg/enums"
	"youlai-gin/internal/middleware"
	response "youlai-gin/internal/common"
	"youlai-gin/pkg/types"
	"youlai-gin/internal/common/validator"
)

func RegisterDeptRoutes(r *gin.RouterGroup) {
	// 使用复数形式
	depts := r.Group("/depts")
	{
		depts.GET("", GetDeptList)
		depts.GET("/options", GetDeptOptions)
		depts.POST("", middleware.OperationLog(enums.LogModuleDept, enums.ActionTypeInsert), SaveDept)
		depts.GET("/:id/form", GetDeptForm)
		depts.PUT("/:id", middleware.OperationLog(enums.LogModuleDept, enums.ActionTypeUpdate), UpdateDept)
		depts.DELETE("/:id", middleware.OperationLog(enums.LogModuleDept, enums.ActionTypeDelete), DeleteDept)
	}
}

// @Summary 部门列表
// @Tags 05.部门接口
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/depts [get]
func GetDeptList(c *gin.Context) {
	var query model.DeptQuery
	if err := validator.BindQuery(c, &query); err != nil {
		c.Error(err)
		return
	}

	currentUser, err := pkgContext.GetCurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	list, err := service.GetDeptList(&query, currentUser)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, list)
}

// @Summary 部门下拉列表
// @Tags 05.部门接口
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/depts/options [get]
func GetDeptOptions(c *gin.Context) {
	currentUser, err := pkgContext.GetCurrentUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	options, err := service.GetDeptOptions(currentUser)
	if err != nil {
		c.Error(err)
		return
	}

	response.Ok(c, options)
}

// @Summary 新增部门
// @Tags 05.部门接口
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
// @Tags 05.部门接口
// @Param id path int true "部门ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/depts/{id}/form [get]
func GetDeptForm(c *gin.Context) {
	id, err := pkgContext.ParsePathParam(c, "id", "部门")
	if err != nil {
		c.Error(err)
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
// @Tags 05.部门接口
// @Param id path int true "部门ID"
// @Param body body model.DeptForm true "部门信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/depts/{id} [put]
func UpdateDept(c *gin.Context) {
	id, err := pkgContext.ParsePathParam(c, "id", "部门")
	if err != nil {
		c.Error(err)
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
// @Tags 05.部门接口
// @Param id path int true "部门ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/depts/{id} [delete]
func DeleteDept(c *gin.Context) {
	id, err := pkgContext.ParsePathParam(c, "id", "部门")
	if err != nil {
		c.Error(err)
		return
	}

	if err := service.DeleteDept(id); err != nil {
		c.Error(err)
		return
	}

	response.OkMsg(c, "删除成功")
}
