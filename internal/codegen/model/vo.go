package model

// TableInfoVO 数据表信息（视图对象）
type TableInfoVO struct {
	TableName      string `json:"tableName"`
	TableComment   string `json:"tableComment"`
	Engine         string `json:"engine"`
	TableCollation string `json:"tableCollation"`
	CreateTime     string `json:"createTime"`
	IsConfigured   int    `json:"isConfigured"`
}

// CodegenPreviewVO 预览文件（视图对象）
type CodegenPreviewVO struct {
	Path     string `json:"path"`
	FileName string `json:"fileName"`
	Content  string `json:"content"`
	Scope    string `json:"scope"`
	Language string `json:"language"`
}
