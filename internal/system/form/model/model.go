package model

import (
	baseModel "youlai-gin/pkg/model"
	"youlai-gin/pkg/types"
)

// FormDefinition 表单定义表
type FormDefinition struct {
	ID          types.BigInt    `gorm:"primaryKey;autoIncrement" json:"id"`
	FormKey     string          `gorm:"column:form_key;uniqueIndex" json:"formKey"`             // 表单唯一标识
	FormName    string          `gorm:"column:form_name" json:"formName"`                       // 表单名称
	Description string          `gorm:"column:description" json:"description"`                  // 表单描述
	FormJson    []any           `gorm:"column:form_json;serializer:json" json:"formJson"`       // 表单规则
	OptionsJson map[string]any  `gorm:"column:options_json;serializer:json" json:"optionsJson"` // 表单全局配置
	Status      int             `gorm:"column:status;default:0" json:"status"`                  // 0草稿 1已发布 -1已停用
	IsPublic    int             `gorm:"column:is_public;default:0" json:"isPublic"`             // 是否允许匿名公开访问
	Category    string          `gorm:"column:category;default:normal" json:"category"`         // normal通用 workflow审批表单
	MenuID      *types.BigInt   `gorm:"column:menu_id" json:"menuId"`                           // 生成的访问菜单ID
	Version     int             `gorm:"column:version;default:1" json:"version"`                // 版本号
	CreateBy    *types.BigInt   `gorm:"column:create_by" json:"createBy,omitempty"`
	CreateTime  types.LocalTime `gorm:"column:create_time;autoCreateTime" json:"createTime,omitempty"`
	UpdateBy    *types.BigInt   `gorm:"column:update_by" json:"updateBy,omitempty"`
	UpdateTime  types.LocalTime `gorm:"column:update_time;autoUpdateTime" json:"updateTime,omitempty"`
	IsDeleted   int             `gorm:"column:is_deleted;default:0" json:"isDeleted"`
}

func (FormDefinition) TableName() string {
	return "form_definition"
}

// FormData 表单数据表
type FormData struct {
	ID          types.BigInt    `gorm:"primaryKey;autoIncrement" json:"id"`
	FormID      types.BigInt    `gorm:"column:form_id" json:"formId"`                     // 表单定义ID
	FormVersion int             `gorm:"column:form_version" json:"formVersion"`           // 提交时表单版本
	DataJson    map[string]any  `gorm:"column:data_json;serializer:json" json:"dataJson"` // 表单数据
	CreateBy    *types.BigInt   `gorm:"column:create_by" json:"createBy,omitempty"`       // 提交人ID，匿名为空
	CreateTime  types.LocalTime `gorm:"column:create_time;autoCreateTime" json:"createTime,omitempty"`
	UpdateBy    *types.BigInt   `gorm:"column:update_by" json:"updateBy,omitempty"`
	UpdateTime  types.LocalTime `gorm:"column:update_time;autoUpdateTime" json:"updateTime,omitempty"`
	IsDeleted   int             `gorm:"column:is_deleted;default:0" json:"isDeleted"`
}

func (FormData) TableName() string {
	return "form_data"
}

// FormSnapshot 表单版本快照表（发布时固化规则，只增不改）
type FormSnapshot struct {
	ID          types.BigInt    `gorm:"primaryKey;autoIncrement" json:"id"`
	FormID      types.BigInt    `gorm:"column:form_id" json:"formId"`
	Version     int             `gorm:"column:version" json:"version"`
	FormJson    []any           `gorm:"column:form_json;serializer:json" json:"formJson"`
	OptionsJson map[string]any  `gorm:"column:options_json;serializer:json" json:"optionsJson"`
	CreateTime  types.LocalTime `gorm:"column:create_time;autoCreateTime" json:"createTime,omitempty"`
	UpdateTime  types.LocalTime `gorm:"column:update_time;autoUpdateTime" json:"updateTime,omitempty"`
	IsDeleted   int             `gorm:"column:is_deleted;default:0" json:"isDeleted"`
}

func (FormSnapshot) TableName() string {
	return "form_snapshot"
}

// FormDefinitionQuery 表单定义分页查询
type FormDefinitionQuery struct {
	baseModel.BaseQuery
	Keywords string `form:"keywords"`
	Status   *int   `form:"status"`
	Category string `form:"category"`
}

// FormDataQuery 表单数据分页查询
type FormDataQuery struct {
	baseModel.BaseQuery
}

// FormDefinitionForm 表单定义入参
type FormDefinitionForm struct {
	ID          types.BigInt   `json:"id"`
	FormKey     string         `json:"formKey" binding:"required"`
	FormName    string         `json:"formName" binding:"required"`
	Description string         `json:"description"`
	FormJson    []any          `json:"formJson"`
	OptionsJson map[string]any `json:"optionsJson"`
	IsPublic    *int           `json:"isPublic"`
	Category    string         `json:"category"`
}

// FormMenuForm 生成访问菜单入参
type FormMenuForm struct {
	MenuName string  `json:"menuName" binding:"required"`
	ParentID *int64  `json:"parentId"`
	RoleIDs  []int64 `json:"roleIds"`
}

// FormAiGenerateForm AI 生成表单规则入参
type FormAiGenerateForm struct {
	Description string `json:"description" binding:"required"`
}

// FormDefinitionVO 表单定义分页行
type FormDefinitionVO struct {
	ID          types.BigInt    `json:"id"`
	FormKey     string          `json:"formKey"`
	FormName    string          `json:"formName"`
	Description string          `json:"description"`
	Status      int             `json:"status"`
	IsPublic    int             `json:"isPublic"`
	Category    string          `json:"category"`
	Version     int             `json:"version"`
	CreateTime  types.LocalTime `json:"createTime"`
}

// FormRenderVO 表单渲染规则
type FormRenderVO struct {
	FormKey     string         `json:"formKey"`
	FormName    string         `json:"formName"`
	Version     int            `json:"version"`
	FormJson    []any          `json:"formJson"`
	OptionsJson map[string]any `json:"optionsJson"`
}

// FormMenuVO 访问菜单回显
type FormMenuVO struct {
	MenuID   types.BigInt `json:"menuId"`
	MenuName string       `json:"menuName"`
	ParentID types.BigInt `json:"parentId"`
	RoleIDs  []int64      `json:"roleIds"`
}

// FormDataVO 表单数据分页行
type FormDataVO struct {
	ID           types.BigInt   `json:"id"`
	FormVersion  int            `json:"formVersion"`
	DataJson     map[string]any `json:"dataJson"`
	CreateBy     *types.BigInt  `json:"createBy"`
	CreateByName string         `json:"createByName"`
	CreateTime   string         `json:"createTime"`
}

// FormDataDetailVO 表单数据详情（附提交时版本规则）
type FormDataDetailVO struct {
	FormDataVO
	FormJson    []any          `json:"formJson"`
	OptionsJson map[string]any `json:"optionsJson"`
}
