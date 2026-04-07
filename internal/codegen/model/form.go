package model

// GenConfigForm 代码生成配置（请求体）
type GenConfigForm struct {
	ID                int64             `json:"id"`
	TableName         string            `json:"tableName"`
	BusinessName      string            `json:"businessName"`
	ModuleName        string            `json:"moduleName"`
	PackageName       string            `json:"packageName"`
	EntityName        string            `json:"entityName"`
	Author            string            `json:"author"`
	ParentMenuId      *int64            `json:"parentMenuId"`
	BackendAppName    string            `json:"backendAppName"`
	FrontendAppName   string            `json:"frontendAppName"`
	PageType          string            `json:"pageType"`
	RemoveTablePrefix string            `json:"removeTablePrefix"`
	FieldConfigs      []FieldConfigForm `json:"fieldConfigs"`
}

// FieldConfigForm 字段配置（表单子结构）
type FieldConfigForm struct {
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
