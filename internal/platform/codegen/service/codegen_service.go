package service

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/viant/velty"

	"youlai-gin/pkg/database"
	"youlai-gin/internal/platform/codegen/model"
	"youlai-gin/pkg/common"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/types"
)

type templateName string

type templateConfig struct {
	templatePath   string
	subpackageName string
	extension      string
}

type templateFieldConfig struct {
	ColumnName    string `velty:"name=columnName"`
	ColumnType    string `velty:"name=columnType"`
	FieldName     string `velty:"name=fieldName"`
	GoFieldName   string `velty:"name=goFieldName"`
	FieldType     string `velty:"name=fieldType"`
	FieldComment  string `velty:"name=fieldComment"`
	IsShowInList  int    `velty:"name=isShowInList"`
	IsShowInForm  int    `velty:"name=isShowInForm"`
	IsShowInQuery int    `velty:"name=isShowInQuery"`
	IsRequired    int    `velty:"name=isRequired"`
	FormType      string `velty:"name=formType"`
	QueryType     string `velty:"name=queryType"`
	MaxLength     *int   `velty:"name=maxLength"`
	FieldSort     *int   `velty:"name=fieldSort"`
	DictType      string `velty:"name=dictType"`
	JavaType      string `velty:"name=javaType"`
	TsType        string `velty:"name=tsType"`
	GoType        string `velty:"name=goType"`
}

const (
	tplAPI        templateName = "API"
	tplAPITypes   templateName = "API_TYPES"
	tplView       templateName = "VIEW"
	tplHandler    templateName = "Handler"
	tplService    templateName = "Service"
	tplRepository templateName = "Repository"
	tplModelEntity templateName = "ModelEntity"
	tplModelForm   templateName = "ModelForm"
	tplModelQuery  templateName = "ModelQuery"
	tplModelVo     templateName = "ModelVo"
	tplRouter     templateName = "Router"
)

var codegenConfig = struct {
	downloadFileName       string
	backendAppName         string
	frontendAppName        string
	defaultAuthor          string
	defaultModuleName      string
	defaultPackageName     string
	defaultRemoveTablePrefix string
}{
	downloadFileName:       "youlai-admin-code.zip",
	backendAppName:         "youlai-gin",
	frontendAppName:        "vue3-element-admin",
	defaultAuthor:          "youlaitech",
	defaultModuleName:      "system",
	defaultPackageName:     "com.youlai.boot",
	defaultRemoveTablePrefix: "sys_",
}

var templateConfigs = map[templateName]templateConfig{
	tplAPI:        {templatePath: "api.ts.vm", subpackageName: "api", extension: ".ts"},
	tplAPITypes:   {templatePath: "api-types.ts.vm", subpackageName: "types", extension: ".ts"},
	tplView:       {templatePath: "index.vue.vm", subpackageName: "views", extension: ".vue"},
	tplHandler:    {templatePath: "handler.go.vm", subpackageName: "handler", extension: ".go"},
	tplService:    {templatePath: "service.go.vm", subpackageName: "service", extension: ".go"},
	tplRepository: {templatePath: "repository.go.vm", subpackageName: "repository", extension: ".go"},
	tplModelEntity:{templatePath: "model-entity.go.vm", subpackageName: "model", extension: ".go"},
	tplModelForm:  {templatePath: "model-form.go.vm", subpackageName: "model", extension: ".go"},
	tplModelQuery: {templatePath: "model-query.go.vm", subpackageName: "model", extension: ".go"},
	tplModelVo:    {templatePath: "model-vo.go.vm", subpackageName: "model", extension: ".go"},
	tplRouter:     {templatePath: "router.go.vm", subpackageName: "", extension: ".go"},
}

// resolveFrontendTemplatePath 解析前端模板路径
func resolveFrontendTemplatePath(name templateName, tc templateConfig, frontendType string) string {
	if frontendType != "js" {
		return tc.templatePath
	}
	if name == tplAPI {
		return "api.js.vm"
	}
	if name == tplView {
		return "index.js.vue.vm"
	}
	return tc.templatePath
}

// resolveFrontendExtension 解析前端文件后缀
func resolveFrontendExtension(name templateName, tc templateConfig, frontendType string) string {
	if frontendType != "js" {
		return tc.extension
	}
	if name == tplAPI {
		return ".js"
	}
	return tc.extension
}

type genTableRow struct {
	ID              int64  `gorm:"column:id"`
	TableName        string `gorm:"column:table_name"`
	ModuleName       string `gorm:"column:module_name"`
	PackageName      string `gorm:"column:package_name"`
	BusinessName     string `gorm:"column:business_name"`
	EntityName       string `gorm:"column:entity_name"`
	Author           string `gorm:"column:author"`
	ParentMenuID     *int64 `gorm:"column:parent_menu_id"`
	RemoveTablePrefix string `gorm:"column:remove_table_prefix"`
	PageType         string `gorm:"column:page_type"`
	IsDeleted        int    `gorm:"column:is_deleted"`
}

type genTableColumnRow struct {
	ID           int64  `gorm:"column:id"`
	TableID      int64  `gorm:"column:table_id"`
	ColumnName   string `gorm:"column:column_name"`
	ColumnType   string `gorm:"column:column_type"`
	FieldName    string `gorm:"column:field_name"`
	FieldType    string `gorm:"column:field_type"`
	FieldSort    *int   `gorm:"column:field_sort"`
	FieldComment string `gorm:"column:field_comment"`
	MaxLength    *int   `gorm:"column:max_length"`
	IsRequired   int    `gorm:"column:is_required"`
	IsShowInList int    `gorm:"column:is_show_in_list"`
	IsShowInForm int    `gorm:"column:is_show_in_form"`
	IsShowInQuery int   `gorm:"column:is_show_in_query"`
	QueryType    int    `gorm:"column:query_type"`
	FormType     int    `gorm:"column:form_type"`
	DictType     string `gorm:"column:dict_type"`
}

func GetTablePage(query *model.TableQuery) (*common.PagedData, error) {
	offset := query.GetOffset()
	limit := query.GetLimit()

	params := make([]interface{}, 0)
	where := "t.TABLE_SCHEMA = DATABASE() AND t.TABLE_NAME NOT IN ('gen_table','gen_table_column')"
	if query.Keywords != "" {
		where += " AND t.TABLE_NAME LIKE ?"
		params = append(params, "%"+query.Keywords+"%")
	}

	listSQL := fmt.Sprintf(`
SELECT
  t.TABLE_NAME AS tableName,
  t.TABLE_COMMENT AS tableComment,
  t.TABLE_COLLATION AS tableCollation,
  t.ENGINE AS engine,
  DATE_FORMAT(t.CREATE_TIME, '%%Y-%%m-%%d %%H:%%i:%%s') AS createTime,
  IF(c.id IS NULL, 0, 1) AS isConfigured
FROM information_schema.TABLES t
LEFT JOIN gen_table c
  ON c.table_name = t.TABLE_NAME AND c.is_deleted = 0
WHERE %s
ORDER BY t.CREATE_TIME DESC
LIMIT ? OFFSET ?`, where)

	listParams := append(params, limit, offset)
	list := make([]model.TableInfoVo, 0)
	if err := database.DB.Raw(listSQL, listParams...).Scan(&list).Error; err != nil {
		return nil, errs.SystemError("查询数据表失败")
	}

	totalSQL := fmt.Sprintf(`SELECT COUNT(1) AS total FROM information_schema.TABLES t WHERE %s`, where)
	var total int64
	if err := database.DB.Raw(totalSQL, params...).Scan(&total).Error; err != nil {
		return nil, errs.SystemError("查询数据表失败")
	}

	pageMeta := common.NewPageMeta(query.PageNum, query.PageSize, total)
	return &common.PagedData{Data: list, Page: pageMeta}, nil
}

func GetGenConfig(tableName string) (*model.GenConfigFormDto, error) {
	var cfg genTableRow
	err := database.DB.Table("gen_table").Where("table_name = ? AND is_deleted = 0", tableName).First(&cfg).Error
	if err == nil {
		fields := make([]genTableColumnRow, 0)
		if err := database.DB.Table("gen_table_column").Where("table_id = ?", cfg.ID).Order("field_sort ASC").Find(&fields).Error; err != nil {
			return nil, errs.SystemError("查询字段配置失败")
		}

		resp := &model.GenConfigFormDto{
			ID:               cfg.ID,
			TableName:        cfg.TableName,
			ModuleName:       cfg.ModuleName,
			PackageName:      cfg.PackageName,
			BusinessName:     cfg.BusinessName,
			EntityName:       cfg.EntityName,
			Author:           cfg.Author,
			ParentMenuId:     cfg.ParentMenuID,
			BackendAppName:   codegenConfig.backendAppName,
			FrontendAppName:  codegenConfig.frontendAppName,
			PageType:         defaultStr(cfg.PageType, "classic"),
			RemoveTablePrefix: defaultStr(cfg.RemoveTablePrefix, codegenConfig.defaultRemoveTablePrefix),
			FieldConfigs:     make([]model.FieldConfigDto, 0, len(fields)),
		}

		for _, f := range fields {
			resp.FieldConfigs = append(resp.FieldConfigs, model.FieldConfigDto{
				ID:            f.ID,
				ColumnName:    f.ColumnName,
				ColumnType:    f.ColumnType,
				FieldName:     f.FieldName,
				FieldType:     f.FieldType,
				FieldComment:  f.FieldComment,
				IsShowInList:  f.IsShowInList,
				IsShowInForm:  f.IsShowInForm,
				IsShowInQuery: f.IsShowInQuery,
				IsRequired:    f.IsRequired,
				FormType:      f.FormType,
				QueryType:     f.QueryType,
				MaxLength:     f.MaxLength,
				FieldSort:     f.FieldSort,
				DictType:      f.DictType,
			})
		}
		return resp, nil
	}

	// 未配置：从 information_schema 生成默认配置
	tableComment := ""
	_ = database.DB.Raw(`SELECT TABLE_COMMENT AS tableComment FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? LIMIT 1`, tableName).Scan(&tableComment).Error

	businessName := strings.TrimSpace(strings.ReplaceAll(tableComment, "表", ""))
	if businessName == "" {
		businessName = tableName
	}

	removePrefix := codegenConfig.defaultRemoveTablePrefix
	processed := tableName
	if removePrefix != "" && strings.HasPrefix(tableName, removePrefix) {
		processed = strings.TrimPrefix(tableName, removePrefix)
	}
	entityName := toPascalCase(processed)

	type columnRow struct {
		ColumnName  string `gorm:"column:columnName"`
		ColumnType  string `gorm:"column:columnType"`
		ColumnComment string `gorm:"column:columnComment"`
		IsNullable  string `gorm:"column:isNullable"`
		MaxLength   *int   `gorm:"column:maxLength"`
		OrdinalPosition int `gorm:"column:ordinalPosition"`
	}

	cols := make([]columnRow, 0)
	if err := database.DB.Raw(`
SELECT
  COLUMN_NAME AS columnName,
  DATA_TYPE AS columnType,
  COLUMN_COMMENT AS columnComment,
  IS_NULLABLE AS isNullable,
  CHARACTER_MAXIMUM_LENGTH AS maxLength,
  ORDINAL_POSITION AS ordinalPosition
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
ORDER BY ORDINAL_POSITION ASC`, tableName).Scan(&cols).Error; err != nil {
		return nil, errs.SystemError("查询表字段失败")
	}

	fieldConfigs := make([]model.FieldConfigDto, 0, len(cols))
	for i, col := range cols {
		javaType := getJavaTypeByColumnType(col.ColumnType)
		isRequired := 1
		if strings.ToUpper(col.IsNullable) == "YES" {
			isRequired = 0
		}
		sort := i + 1
		fieldConfigs = append(fieldConfigs, model.FieldConfigDto{
			ColumnName:    col.ColumnName,
			ColumnType:    col.ColumnType,
			FieldName:     toCamelCase(col.ColumnName),
			FieldType:     javaType,
			FieldComment:  col.ColumnComment,
			IsRequired:    isRequired,
			FormType:      getDefaultFormTypeByColumnType(col.ColumnType),
			QueryType:     1,
			MaxLength:     col.MaxLength,
			FieldSort:     &sort,
			IsShowInList:  1,
			IsShowInForm:  1,
			IsShowInQuery: 0,
		})
	}

	return &model.GenConfigFormDto{
		TableName:        tableName,
		BusinessName:     businessName,
		ModuleName:       codegenConfig.defaultModuleName,
		PackageName:      codegenConfig.defaultPackageName,
		EntityName:       entityName,
		Author:           codegenConfig.defaultAuthor,
		BackendAppName:   codegenConfig.backendAppName,
		FrontendAppName:  codegenConfig.frontendAppName,
		PageType:         "classic",
		RemoveTablePrefix: removePrefix,
		FieldConfigs:     fieldConfigs,
	}, nil
}

func SaveGenConfig(tableName string, body *model.GenConfigFormDto) error {
	if body == nil {
		return errs.BadRequest("参数错误")
	}

	now := time.Now()

	var existing genTableRow
	err := database.DB.Table("gen_table").Where("table_name = ?", tableName).First(&existing).Error

	moduleName := defaultStr(body.ModuleName, codegenConfig.defaultModuleName)
	packageName := defaultStr(body.PackageName, codegenConfig.defaultPackageName)
	businessName := defaultStr(body.BusinessName, tableName)
	entityName := defaultStr(body.EntityName, toPascalCase(tableName))
	author := defaultStr(body.Author, codegenConfig.defaultAuthor)
	pageType := defaultStr(body.PageType, "classic")
	removePrefix := defaultStr(body.RemoveTablePrefix, codegenConfig.defaultRemoveTablePrefix)

	if err == nil && existing.ID > 0 {
		updates := map[string]interface{}{
			"module_name":        moduleName,
			"package_name":       packageName,
			"business_name":      businessName,
			"entity_name":        entityName,
			"author":             author,
			"parent_menu_id":     body.ParentMenuId,
			"remove_table_prefix": removePrefix,
			"page_type":          pageType,
			"update_time":        now,
			"is_deleted":         0,
		}
		if err := database.DB.Table("gen_table").Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
			return errs.SystemError("保存配置失败")
		}

		if err := database.DB.Table("gen_table_column").Where("table_id = ?", existing.ID).Delete(&genTableColumnRow{}).Error; err != nil {
			return errs.SystemError("保存字段配置失败")
		}

		for i := range body.FieldConfigs {
			f := body.FieldConfigs[i]
			sort := i + 1
			if f.FieldSort != nil {
				sort = *f.FieldSort
			}

			row := genTableColumnRow{
				TableID:      existing.ID,
				ColumnName:   f.ColumnName,
				ColumnType:   f.ColumnType,
				FieldName:    defaultStr(f.FieldName, toCamelCase(f.ColumnName)),
				FieldType:    defaultStr(f.FieldType, getJavaTypeByColumnType(f.ColumnType)),
				FieldSort:    &sort,
				FieldComment: f.FieldComment,
				MaxLength:    f.MaxLength,
				IsRequired:   defaultInt(f.IsRequired, 0),
				IsShowInList: defaultInt(f.IsShowInList, 0),
				IsShowInForm: defaultInt(f.IsShowInForm, 0),
				IsShowInQuery: defaultInt(f.IsShowInQuery, 0),
				QueryType:    defaultInt(f.QueryType, 1),
				FormType:     defaultInt(f.FormType, 1),
				DictType:     f.DictType,
			}

			if err := database.DB.Table("gen_table_column").Create(&row).Error; err != nil {
				return errs.SystemError("保存字段配置失败")
			}
		}

		return nil
	}

	insert := map[string]interface{}{
		"table_name":         tableName,
		"module_name":        moduleName,
		"package_name":       packageName,
		"business_name":      businessName,
		"entity_name":        entityName,
		"author":             author,
		"parent_menu_id":     body.ParentMenuId,
		"remove_table_prefix": removePrefix,
		"page_type":          pageType,
		"create_time":        now,
		"update_time":        now,
		"is_deleted":         0,
	}

	if err := database.DB.Table("gen_table").Create(insert).Error; err != nil {
		return errs.SystemError("保存配置失败")
	}

	var created genTableRow
	if err := database.DB.Table("gen_table").Where("table_name = ?", tableName).First(&created).Error; err != nil {
		return errs.SystemError("保存配置失败")
	}

	for i := range body.FieldConfigs {
		f := body.FieldConfigs[i]
		sort := i + 1
		if f.FieldSort != nil {
			sort = *f.FieldSort
		}

		row := genTableColumnRow{
			TableID:      created.ID,
			ColumnName:   f.ColumnName,
			ColumnType:   f.ColumnType,
			FieldName:    defaultStr(f.FieldName, toCamelCase(f.ColumnName)),
			FieldType:    defaultStr(f.FieldType, getJavaTypeByColumnType(f.ColumnType)),
			FieldSort:    &sort,
			FieldComment: f.FieldComment,
			MaxLength:    f.MaxLength,
			IsRequired:   defaultInt(f.IsRequired, 0),
			IsShowInList: defaultInt(f.IsShowInList, 0),
			IsShowInForm: defaultInt(f.IsShowInForm, 0),
			IsShowInQuery: defaultInt(f.IsShowInQuery, 0),
			QueryType:    defaultInt(f.QueryType, 1),
			FormType:     defaultInt(f.FormType, 1),
			DictType:     f.DictType,
		}

		if err := database.DB.Table("gen_table_column").Create(&row).Error; err != nil {
			return errs.SystemError("保存字段配置失败")
		}
	}

	return nil
}

func DeleteGenConfig(tableName string) error {
	var cfg genTableRow
	if err := database.DB.Table("gen_table").Where("table_name = ? AND is_deleted = 0", tableName).First(&cfg).Error; err != nil {
		return nil
	}

	if err := database.DB.Table("gen_table_column").Where("table_id = ?", cfg.ID).Delete(&genTableColumnRow{}).Error; err != nil {
		return errs.SystemError("删除失败")
	}

	if err := database.DB.Table("gen_table").Where("id = ?", cfg.ID).Updates(map[string]interface{}{"is_deleted": 1, "update_time": time.Now()}).Error; err != nil {
		return errs.SystemError("删除失败")
	}

	return nil
}

func GetPreview(tableName string, pageType string, typeParam string) ([]model.CodegenPreviewVo, error) {
	cfg, err := GetGenConfig(tableName)
	if err != nil {
		return nil, err
	}
	frontendType := "ts"
	if strings.ToLower(strings.TrimSpace(typeParam)) == "js" {
		frontendType = "js"
	}

	previews := make([]model.CodegenPreviewVo, 0)
	for name, tc := range templateConfigs {
		if frontendType == "js" && name == tplAPITypes {
			continue
		}

		templatePath := resolveFrontendTemplatePath(name, tc, frontendType)
		extension := resolveFrontendExtension(name, tc, frontendType)
		fileName := getFileName(cfg.EntityName, name, extension)
		filePath := getFilePath(name, cfg.ModuleName, cfg.PackageName, tc.subpackageName, cfg.EntityName)

		content, err := renderTemplate(name, templatePath, tc.subpackageName, cfg, pageType)
		if err != nil {
			return nil, err
		}

		previews = append(previews, model.CodegenPreviewVo{Path: filepath.ToSlash(filePath), FileName: fileName, Content: content})
	}

	return previews, nil
}

func DownloadZip(tableNames []string, pageType string, typeParam string) (string, []byte, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	for _, t := range tableNames {
		list, err := GetPreview(t, pageType, typeParam)
		if err != nil {
			_ = zw.Close()
			return "", nil, err
		}

		for _, item := range list {
			zipPath := filepath.ToSlash(filepath.Join(item.Path, item.FileName))
			w, err := zw.Create(zipPath)
			if err != nil {
				_ = zw.Close()
				return "", nil, errs.SystemError("生成压缩包失败")
			}
			_, _ = w.Write([]byte(item.Content))
		}
	}

	if err := zw.Close(); err != nil {
		return "", nil, errs.SystemError("生成压缩包失败")
	}

	return codegenConfig.downloadFileName, buf.Bytes(), nil
}

func renderTemplate(
	name templateName,
	templatePath string,
	subpackageName string,
	cfg *model.GenConfigFormDto,
	pageType string,
) (string, error) {
	effectivePath := templatePath
	if name == tplView && pageType == "curd" {
		if strings.HasSuffix(effectivePath, "index.js.vue.vm") {
			effectivePath = strings.Replace(effectivePath, "index.js.vue.vm", "index.curd.js.vue.vm", 1)
		} else if strings.HasSuffix(effectivePath, "index.vue.vm") {
			effectivePath = strings.Replace(effectivePath, "index.vue.vm", "index.curd.vue.vm", 1)
		}
	}

	absPath := resolveBootTemplatePath(effectivePath)
	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", errs.SystemError("读取模板失败")
	}

	// velty 对 Velocity 的 silent reference（$!）支持不完整，这里做一次兼容转换
	// $!{foo} -> ${foo}
	// $!bar   -> $bar
	content = []byte(strings.ReplaceAll(string(content), "$!{", "${"))
	content = []byte(strings.ReplaceAll(string(content), "$!", "$"))

	planner := velty.New()
	_ = planner.RegisterFunction("trim", strings.TrimSpace)

	_ = planner.DefineVariable("packageName", "")
	_ = planner.DefineVariable("moduleName", "")
	_ = planner.DefineVariable("subpackageName", "")
	_ = planner.DefineVariable("date", "")
	_ = planner.DefineVariable("entityName", "")
	_ = planner.DefineVariable("tableName", "")
	_ = planner.DefineVariable("author", "")
	_ = planner.DefineVariable("entityLowerCamel", "")
	_ = planner.DefineVariable("entityKebab", "")
	_ = planner.DefineVariable("entityUpperSnake", "")
	_ = planner.DefineVariable("entitySnake", "")
	_ = planner.DefineVariable("businessName", "")
	_ = planner.DefineVariable("fieldConfigs", reflect.TypeOf([]templateFieldConfig{}))

	exec, newState, err := planner.Compile(content)
	if err != nil {
		return "", errs.SystemError("编译模板失败")
	}

	state := newState()
	_ = state.SetValue("packageName", cfg.PackageName)
	_ = state.SetValue("moduleName", cfg.ModuleName)
	_ = state.SetValue("subpackageName", subpackageName)
	_ = state.SetValue("date", formatDateTime(time.Now()))
	_ = state.SetValue("entityName", cfg.EntityName)
	_ = state.SetValue("tableName", cfg.TableName)
	_ = state.SetValue("author", cfg.Author)
	_ = state.SetValue("entityLowerCamel", lowerFirst(cfg.EntityName))
	_ = state.SetValue("entityKebab", toKebabCase(cfg.EntityName))
	_ = state.SetValue("entityUpperSnake", toSnakeUpper(cfg.EntityName))
	_ = state.SetValue("entitySnake", toSnakeLower(cfg.EntityName))
	_ = state.SetValue("businessName", cfg.BusinessName)

	fields := make([]templateFieldConfig, 0, len(cfg.FieldConfigs))
	for i := range cfg.FieldConfigs {
		f := cfg.FieldConfigs[i]
		javaType := defaultStr(f.FieldType, getJavaTypeByColumnType(f.ColumnType))
		goType := getGoTypeByColumnType(f.ColumnType)
		fields = append(fields, templateFieldConfig{
			ColumnName:    f.ColumnName,
			ColumnType:    f.ColumnType,
			FieldName:     f.FieldName,
			GoFieldName:   toGoFieldName(f.FieldName),
			FieldType:     javaType,
			FieldComment:  f.FieldComment,
			IsShowInList:  f.IsShowInList,
			IsShowInForm:  f.IsShowInForm,
			IsShowInQuery: f.IsShowInQuery,
			IsRequired:    f.IsRequired,
			FormType:      getFormTypeName(f.FormType),
			QueryType:     getQueryTypeName(f.QueryType),
			MaxLength:     f.MaxLength,
			FieldSort:     f.FieldSort,
			DictType:      f.DictType,
			JavaType:      javaType,
			TsType:        getTsTypeByJavaType(javaType),
			GoType:        goType,
		})
	}
	_ = state.SetValue("fieldConfigs", fields)

	exec.Exec(state)
	if !state.IsValid() {
		return "", errs.SystemError("渲染模板失败")
	}
	return state.Buffer.String(), nil
}

func resolveBootTemplatePath(templatePath string) string {
	return filepath.Join("internal", "platform", "codegen", "templates", filepath.FromSlash(templatePath))
}

func getFileName(entityName string, name templateName, extension string) string {
	if name == tplAPI {
		return toKebabCase(entityName) + extension
	}
	if name == tplAPITypes {
		return toKebabCase(entityName) + extension
	}
	if name == tplView {
		return "index.vue"
	}
	if name == tplHandler {
		return toSnakeLower(entityName) + "_handler" + extension
	}
	if name == tplService {
		return toSnakeLower(entityName) + "_service" + extension
	}
	if name == tplRepository {
		return toSnakeLower(entityName) + "_repo" + extension
	}
	if name == tplModelEntity {
		return "entity" + extension
	}
	if name == tplModelForm {
		return "form" + extension
	}
	if name == tplModelQuery {
		return "query" + extension
	}
	if name == tplModelVo {
		return "vo" + extension
	}
	if name == tplRouter {
		return "router" + extension
	}
	return entityName + string(name) + extension
}

func getFilePath(name templateName, moduleName string, packageName string, subpackageName string, entityName string) string {
	backend := codegenConfig.backendAppName
	frontend := codegenConfig.frontendAppName

	if name == tplAPI {
		return filepath.Join(frontend, "src", subpackageName, moduleName)
	}
	if name == tplAPITypes {
		return filepath.Join(frontend, "src", "types", "api")
	}
	if name == tplView {
		return filepath.Join(frontend, "src", subpackageName, moduleName, toKebabCase(entityName))
	}

	base := filepath.Join(backend, "internal", moduleName, toKebabCase(entityName))
	if subpackageName == "" {
		return base
	}
	return filepath.Join(base, subpackageName)
}

func defaultStr(v string, dv string) string {
	if strings.TrimSpace(v) == "" {
		return dv
	}
	return v
}

func defaultInt(v int, dv int) int {
	if v == 0 {
		return dv
	}
	return v
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func toGoFieldName(name string) string {
	if name == "" {
		return name
	}
	if strings.Contains(name, "_") || strings.Contains(name, "-") {
		name = toPascalCase(name)
	} else {
		name = strings.ToUpper(name[:1]) + name[1:]
	}
	name = strings.ReplaceAll(name, "Ids", "IDs")
	name = strings.ReplaceAll(name, "Id", "ID")
	return name
}

func toCamelCase(s string) string {
	s = strings.ToLower(s)
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' })
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		p := parts[i]
		if p == "" {
			continue
		}
		out += strings.ToUpper(p[:1]) + p[1:]
	}
	return out
}

func toPascalCase(s string) string {
	c := toCamelCase(s)
	if c == "" {
		return c
	}
	return strings.ToUpper(c[:1]) + c[1:]
}

func toKebabCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('-')
		}
		if r == '_' {
			b.WriteByte('-')
			continue
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}

func getGoTypeByColumnType(columnType string) string {
	t := normalizeColumnType(columnType)
	switch t {
	case "varchar", "char", "text", "longtext", "mediumtext", "json":
		return "string"
	case "int", "tinyint", "smallint", "mediumint":
		return "int"
	case "bigint":
		return "types.BigInt"
	case "float", "double", "decimal":
		return "float64"
	case "date", "datetime", "timestamp":
		return "string"
	case "boolean", "bool", "bit":
		return "int"
	default:
		return "string"
	}
}

func toSnakeUpper(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		if r == '-' {
			b.WriteByte('_')
			continue
		}
		b.WriteRune(r)
	}
	return strings.ToUpper(b.String())
}

func toSnakeLower(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		if r == '-' {
			b.WriteByte('_')
			continue
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}

func formatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

func normalizeColumnType(columnType string) string {
	t := strings.ToLower(strings.TrimSpace(columnType))
	t = strings.ReplaceAll(t, "unsigned", "")
	t = strings.ReplaceAll(t, "zerofill", "")
	t = strings.TrimSpace(t)
	if idx := strings.Index(t, "("); idx >= 0 {
		t = t[:idx]
	}
	t = strings.TrimSpace(t)
	return t
}

func getJavaTypeByColumnType(columnType string) string {
	t := normalizeColumnType(columnType)
	switch t {
	case "varchar", "char", "text", "json":
		return "String"
	case "blob":
		return "byte[]"
	case "int", "tinyint", "smallint", "mediumint":
		return "Integer"
	case "bigint":
		return "Long"
	case "float":
		return "Float"
	case "double":
		return "Double"
	case "decimal":
		return "BigDecimal"
	case "date":
		return "LocalDate"
	case "datetime", "timestamp":
		return "LocalDateTime"
	case "boolean", "bit":
		return "Boolean"
	default:
		return "String"
	}
}

func getTsTypeByJavaType(javaType string) string {
	switch javaType {
	case "String":
		return "string"
	case "Integer", "Long", "Float", "Double", "BigDecimal":
		return "number"
	case "Boolean":
		return "boolean"
	case "byte[]":
		return "Uint8Array"
	case "LocalDate", "LocalDateTime":
		return "string"
	default:
		return "any"
	}
}

func getDefaultFormTypeByColumnType(columnType string) int {
	t := normalizeColumnType(columnType)
	if t == "date" {
		return 8
	}
	if t == "datetime" || t == "timestamp" {
		return 9
	}
	return 1
}

func getFormTypeName(value int) string {
	switch value {
	case 1:
		return "INPUT"
	case 2:
		return "SELECT"
	case 3:
		return "RADIO"
	case 4:
		return "CHECK_BOX"
	case 5:
		return "INPUT_NUMBER"
	case 6:
		return "SWITCH"
	case 7:
		return "TEXT_AREA"
	case 8:
		return "DATE"
	case 9:
		return "DATE_TIME"
	case 10:
		return "HIDDEN"
	default:
		return "INPUT"
	}
}

func getQueryTypeName(value int) string {
	switch value {
	case 1:
		return "EQ"
	case 2:
		return "LIKE"
	case 3:
		return "IN"
	case 4:
		return "BETWEEN"
	case 5:
		return "GT"
	case 6:
		return "GE"
	case 7:
		return "LT"
	case 8:
		return "LE"
	case 9:
		return "NE"
	case 10:
		return "LIKE_LEFT"
	case 11:
		return "LIKE_RIGHT"
	default:
		return "EQ"
	}
}
