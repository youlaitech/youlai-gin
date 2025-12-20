package model

import "youlai-gin/pkg/common"

// TablePageQuery 数据表分页查询参数
type TablePageQuery struct {
	common.PageQuery
	Keywords string `form:"keywords"`
}

// TableInfoVo 数据表信息
type TableInfoVo struct {
	TableName    string `json:"tableName" gorm:"column:tableName"`
	TableComment string `json:"tableComment" gorm:"column:tableComment"`
	Engine       string `json:"engine" gorm:"column:engine"`
	TableCollation string `json:"tableCollation" gorm:"column:tableCollation"`
	CreateTime   string `json:"createTime" gorm:"column:createTime"`
	IsConfigured int    `json:"isConfigured" gorm:"column:isConfigured"`
}

// GenConfigFormDto 代码生成配置
type GenConfigFormDto struct {
	ID              int64           `json:"id"`
	TableName        string          `json:"tableName"`
	BusinessName     string          `json:"businessName"`
	ModuleName       string          `json:"moduleName"`
	PackageName      string          `json:"packageName"`
	EntityName       string          `json:"entityName"`
	Author           string          `json:"author"`
	ParentMenuId     *int64          `json:"parentMenuId"`
	BackendAppName   string          `json:"backendAppName"`
	FrontendAppName  string          `json:"frontendAppName"`
	PageType         string          `json:"pageType"`
	RemoveTablePrefix string         `json:"removeTablePrefix"`
	FieldConfigs     []FieldConfigDto `json:"fieldConfigs"`
}

// FieldConfigDto 字段配置
type FieldConfigDto struct {
	ID           int64  `json:"id"`
	ColumnName   string `json:"columnName"`
	ColumnType   string `json:"columnType"`
	FieldName    string `json:"fieldName"`
	FieldType    string `json:"fieldType"`
	FieldComment string `json:"fieldComment"`
	IsShowInList int    `json:"isShowInList"`
	IsShowInForm int    `json:"isShowInForm"`
	IsShowInQuery int   `json:"isShowInQuery"`
	IsRequired   int    `json:"isRequired"`
	FormType     int    `json:"formType"`
	QueryType    int    `json:"queryType"`
	MaxLength    *int   `json:"maxLength"`
	FieldSort    *int   `json:"fieldSort"`
	DictType     string `json:"dictType"`
}

// CodegenPreviewVo 预览文件
type CodegenPreviewVo struct {
	Path     string `json:"path"`
	FileName string `json:"fileName"`
	Content  string `json:"content"`
}
