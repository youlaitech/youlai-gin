package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	response "youlai-gin/internal/common"
	"youlai-gin/internal/common/auth"
	appContext "youlai-gin/internal/common/context"
	"youlai-gin/internal/middleware"
	"youlai-gin/internal/system/form/model"
	"youlai-gin/internal/system/form/service"
	"youlai-gin/pkg/enums"
	"youlai-gin/pkg/errs"
)

// Handler 动态表单接口
type Handler struct {
	svc *service.Service
}

// NewHandler 创建 Handler 实例
func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

// RegisterRoutes 注册动态表单模块路由。
// 说明：gin 同一路径位置不允许出现不同参数名，故 id 段与 formKey 段统一使用 :id，
// 由 handler 内 resolveFormKey 判定「纯数字按主键、其余按 formKey」。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	forms := r.Group("/forms")
	{
		forms.GET("", auth.RequirePermission("form:definition:list"), h.Page)
		forms.POST("", auth.RequirePermission("form:definition:create"),
			middleware.OperationLog(enums.LogModuleForm, enums.ActionTypeInsert), h.Create)
		forms.GET("/options", h.WorkflowOptions)
		forms.POST("/ai-generate", auth.RequirePermission("form:definition:create"), h.AiGenerate)

		// 公开接口：匿名访问，不加鉴权中间件
		forms.GET("/public/:id/render", h.PublicRender)
		forms.POST("/public/:id/data", h.PublicSubmit)

		forms.GET("/:id/form", auth.RequirePermission("form:definition:list"), h.GetForm)
		forms.PUT("/:id", auth.RequirePermission("form:definition:update"),
			middleware.OperationLog(enums.LogModuleForm, enums.ActionTypeUpdate), h.Update)
		forms.DELETE("/:id", auth.RequirePermission("form:definition:delete"),
			middleware.OperationLog(enums.LogModuleForm, enums.ActionTypeDelete), h.Delete)
		forms.PUT("/:id/publish", auth.RequirePermission("form:definition:update"),
			middleware.OperationLog(enums.LogModuleForm, enums.ActionTypeUpdate), h.Publish)
		forms.PUT("/:id/disable", auth.RequirePermission("form:definition:update"),
			middleware.OperationLog(enums.LogModuleForm, enums.ActionTypeUpdate), h.Disable)
		forms.GET("/:id/menu", auth.RequirePermission("form:definition:list"), h.GetMenu)
		forms.POST("/:id/menu", auth.RequirePermission("form:definition:update"),
			middleware.OperationLog(enums.LogModuleForm, enums.ActionTypeInsert), h.SaveMenu)

		forms.GET("/:id/render", auth.RequirePermission("form:definition:list"), h.Render)
		forms.GET("/:id/data", auth.RequirePermission("form:data:list"), h.DataPage)
		forms.POST("/:id/data", h.SubmitData)
		forms.GET("/:id/data/:dataId", auth.RequirePermission("form:data:list"), h.DataDetail)
		forms.DELETE("/:id/data/:dataId", auth.RequirePermission("form:data:delete"),
			middleware.OperationLog(enums.LogModuleForm, enums.ActionTypeDelete), h.DeleteData)
	}
}

// resolveFormKey 解析 id 段：纯数字按主键查 formKey，其余按 formKey 直接返回
func (h *Handler) resolveFormKey(c *gin.Context) (string, error) {
	keyOrID := c.Param("id")
	if id, err := strconv.ParseInt(keyOrID, 10, 64); err == nil {
		definition, err := h.svc.GetDefinitionForm(c.Request.Context(), id)
		if err != nil {
			return "", err
		}
		return definition.FormKey, nil
	}
	return keyOrID, nil
}

func parseIDs(raw string) ([]int64, error) {
	parts := strings.Split(raw, ",")
	ids := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return nil, errs.BadRequest("ID格式不正确")
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// @Summary 表单分页列表
// @Tags 12.动态表单
// @Router /api/v1/forms [get]
func (h *Handler) Page(c *gin.Context) {
	var query model.FormDefinitionQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(errs.BadRequest("查询参数不正确"))
		return
	}

	result, err := h.svc.DefinitionPage(c.Request.Context(), &query)
	if err != nil {
		c.Error(err)
		return
	}
	response.OkPaged(c, result)
}

// @Summary 新增表单
// @Tags 12.动态表单
// @Router /api/v1/forms [post]
func (h *Handler) Create(c *gin.Context) {
	var form model.FormDefinitionForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.Error(errs.BadRequest("请求参数不正确"))
		return
	}

	if err := h.svc.CreateDefinition(appContext.OperatorCtx(c), &form); err != nil {
		c.Error(err)
		return
	}
	response.OkMsg(c, "保存成功")
}

// @Summary 审批表单下拉选项
// @Tags 12.动态表单
// @Router /api/v1/forms/options [get]
func (h *Handler) WorkflowOptions(c *gin.Context) {
	options, err := h.svc.WorkflowOptions(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, options)
}

// @Summary AI 生成表单规则
// @Tags 12.动态表单
// @Router /api/v1/forms/ai-generate [post]
func (h *Handler) AiGenerate(c *gin.Context) {
	var form model.FormAiGenerateForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.Error(errs.BadRequest("需求描述不能为空"))
		return
	}

	rules, err := h.svc.AiGenerateRule(c.Request.Context(), form.Description)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, rules)
}

// @Summary 公开表单渲染规则
// @Tags 12.动态表单
// @Router /api/v1/forms/public/{formKey}/render [get]
func (h *Handler) PublicRender(c *gin.Context) {
	vo, err := h.svc.PublicRender(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, vo)
}

// @Summary 公开表单提交
// @Tags 12.动态表单
// @Router /api/v1/forms/public/{formKey}/data [post]
func (h *Handler) PublicSubmit(c *gin.Context) {
	var data map[string]any
	if err := c.ShouldBindJSON(&data); err != nil {
		c.Error(errs.BadRequest("请求参数不正确"))
		return
	}

	if err := h.svc.SubmitPublic(c.Request.Context(), c.Param("id"), data, c.ClientIP()); err != nil {
		c.Error(err)
		return
	}
	response.OkMsg(c, "提交成功")
}

// @Summary 表单设计数据
// @Tags 12.动态表单
// @Router /api/v1/forms/{id}/form [get]
func (h *Handler) GetForm(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errs.BadRequest("表单ID格式不正确"))
		return
	}

	entity, err := h.svc.GetDefinitionForm(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, entity)
}

// @Summary 修改表单
// @Tags 12.动态表单
// @Router /api/v1/forms/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errs.BadRequest("表单ID格式不正确"))
		return
	}

	var form model.FormDefinitionForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.Error(errs.BadRequest("请求参数不正确"))
		return
	}

	if err := h.svc.UpdateDefinition(appContext.OperatorCtx(c), id, &form); err != nil {
		c.Error(err)
		return
	}
	response.OkMsg(c, "修改成功")
}

// @Summary 删除表单
// @Tags 12.动态表单
// @Router /api/v1/forms/{ids} [delete]
func (h *Handler) Delete(c *gin.Context) {
	ids, err := parseIDs(c.Param("id"))
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.DeleteDefinitions(c.Request.Context(), ids); err != nil {
		c.Error(err)
		return
	}
	response.OkMsg(c, "删除成功")
}

// @Summary 发布表单
// @Tags 12.动态表单
// @Router /api/v1/forms/{id}/publish [put]
func (h *Handler) Publish(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errs.BadRequest("表单ID格式不正确"))
		return
	}

	if err := h.svc.Publish(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	response.OkMsg(c, "发布成功")
}

// @Summary 停用表单
// @Tags 12.动态表单
// @Router /api/v1/forms/{id}/disable [put]
func (h *Handler) Disable(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errs.BadRequest("表单ID格式不正确"))
		return
	}

	if err := h.svc.Disable(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	response.OkMsg(c, "停用成功")
}

// @Summary 表单访问菜单回显
// @Tags 12.动态表单
// @Router /api/v1/forms/{id}/menu [get]
func (h *Handler) GetMenu(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errs.BadRequest("表单ID格式不正确"))
		return
	}

	vo, err := h.svc.GetMenu(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, vo)
}

// @Summary 生成表单访问菜单
// @Tags 12.动态表单
// @Router /api/v1/forms/{id}/menu [post]
func (h *Handler) SaveMenu(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errs.BadRequest("表单ID格式不正确"))
		return
	}

	var form model.FormMenuForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.Error(errs.BadRequest("菜单名称不能为空"))
		return
	}

	if err := h.svc.SaveMenu(c.Request.Context(), id, &form); err != nil {
		c.Error(err)
		return
	}
	response.OkMsg(c, "菜单生成成功")
}

// @Summary 表单渲染规则
// @Tags 12.动态表单
// @Router /api/v1/forms/{formKey}/render [get]
func (h *Handler) Render(c *gin.Context) {
	formKey, err := h.resolveFormKey(c)
	if err != nil {
		c.Error(err)
		return
	}

	vo, err := h.svc.Render(c.Request.Context(), formKey)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, vo)
}

// @Summary 表单数据分页
// @Tags 12.动态表单
// @Router /api/v1/forms/{formKey}/data [get]
func (h *Handler) DataPage(c *gin.Context) {
	formKey, err := h.resolveFormKey(c)
	if err != nil {
		c.Error(err)
		return
	}

	var query model.FormDataQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(errs.BadRequest("查询参数不正确"))
		return
	}

	result, err := h.svc.DataPage(c.Request.Context(), formKey, &query)
	if err != nil {
		c.Error(err)
		return
	}
	response.OkPaged(c, result)
}

// @Summary 提交表单数据
// @Tags 12.动态表单
// @Router /api/v1/forms/{formKey}/data [post]
func (h *Handler) SubmitData(c *gin.Context) {
	formKey, err := h.resolveFormKey(c)
	if err != nil {
		c.Error(err)
		return
	}

	var data map[string]any
	if err := c.ShouldBindJSON(&data); err != nil {
		c.Error(errs.BadRequest("请求参数不正确"))
		return
	}

	if err := h.svc.Submit(appContext.OperatorCtx(c), formKey, data); err != nil {
		c.Error(err)
		return
	}
	response.OkMsg(c, "提交成功")
}

// @Summary 表单数据详情
// @Tags 12.动态表单
// @Router /api/v1/forms/{formKey}/data/{dataId} [get]
func (h *Handler) DataDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("dataId"), 10, 64)
	if err != nil {
		c.Error(errs.BadRequest("数据ID格式不正确"))
		return
	}

	vo, err := h.svc.DataDetail(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	response.Ok(c, vo)
}

// @Summary 删除表单数据
// @Tags 12.动态表单
// @Router /api/v1/forms/{formKey}/data/{ids} [delete]
func (h *Handler) DeleteData(c *gin.Context) {
	ids, err := parseIDs(c.Param("dataId"))
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.svc.DeleteData(c.Request.Context(), ids); err != nil {
		c.Error(err)
		return
	}
	response.OkMsg(c, "删除成功")
}
