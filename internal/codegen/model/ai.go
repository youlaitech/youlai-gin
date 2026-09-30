package model

// AiFillConfigForm AI 填充生成配置入参
type AiFillConfigForm struct {
	Requirement string `json:"requirement"`
}

// GenConfigAiVO AI 推断的代码生成配置
type GenConfigAiVO struct {
	BusinessName string             `json:"businessName"`
	FieldConfigs []GenConfigAiField `json:"fieldConfigs"`
}

// GenConfigAiField AI 推断的字段配置
type GenConfigAiField struct {
	ColumnName    string `json:"columnName"`
	FieldComment  string `json:"fieldComment"`
	FormType      string `json:"formType"`
	QueryType     string `json:"queryType"`
	IsRequired    *int   `json:"isRequired"`
	IsShowInList  *int   `json:"isShowInList"`
	IsShowInForm  *int   `json:"isShowInForm"`
	IsShowInQuery *int   `json:"isShowInQuery"`
	DictType      string `json:"dictType"`
}
