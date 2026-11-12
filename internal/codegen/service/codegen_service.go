package service

import (
	"archive/zip"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/viant/velty"

	"youlai-gin/internal/codegen/model"
	"youlai-gin/internal/codegen/repository"
	"youlai-gin/internal/common/ai"
	"youlai-gin/internal/common/logger"
	"youlai-gin/internal/common/utils"
	menuService "youlai-gin/internal/system/menu/service"
	"youlai-gin/pkg/errs"
	commonModel "youlai-gin/pkg/model"
)

type templateName string

type templateConfig struct {
	templatePath   string
	subpackageName string
	extension      string
}

type templateFieldConfig struct {
	ColumnName     string `velty:"name=columnName"`
	ColumnType     string `velty:"name=columnType"`
	FieldName      string `velty:"name=fieldName"`
	GoFieldName    string `velty:"name=goFieldName"`
	FieldType      string `velty:"name=fieldType"`
	FieldComment   string `velty:"name=fieldComment"`
	IsShowInList   int    `velty:"name=isShowInList"`
	IsShowInForm   int    `velty:"name=isShowInForm"`
	IsShowInQuery  int    `velty:"name=isShowInQuery"`
	IsRequired     int    `velty:"name=isRequired"`
	FormType       string `velty:"name=formType"`
	QueryType      string `velty:"name=queryType"`
	MaxLength      *int   `velty:"name=maxLength"`
	FieldSort      *int   `velty:"name=fieldSort"`
	DictType       string `velty:"name=dictType"`
	TsType         string `velty:"name=tsType"`
	GoType         string `velty:"name=goType"`
	QueryGoType    string `velty:"name=queryGoType"`
	TsQueryType    string `velty:"name=tsQueryType"`
	GormTag        string `velty:"name=gormTag"`
	JSONTag        string `velty:"name=jsonTag"`
	QueryIsSetExpr string `velty:"name=queryIsSetExpr"`
	QueryArgExpr   string `velty:"name=queryArgExpr"`

	// 模板计算字段：Velty 对字符串的 and/or/裸选择器语义不可靠，统一在 Go 侧算好布尔标志。
	HasDictType     bool `velty:"name=hasDictType"`
	IsDate          bool `velty:"name=isDate"`
	IsDateTime      bool `velty:"name=isDateTime"`
	IsNumeric       bool `velty:"name=isNumeric"`
	IsDecimal       bool `velty:"name=isDecimal"`
	VisibleInForm   bool `velty:"name=visibleInForm"`
	RequiredInForm  bool `velty:"name=requiredInForm"`
	BindingRequired bool `velty:"name=bindingRequired"`
	ShowDictInList  bool `velty:"name=showDictInList"`
	ShowDictInQuery bool `velty:"name=showDictInQuery"`
	ShowDictInForm  bool `velty:"name=showDictInForm"`
	IsTimeColumn    bool `velty:"name=isTimeColumn"`
}

// nonFormColumns 服务端维护的列，不进入生成表单，避免客户端篡改主键、审计字段与逻辑删除标记。
var nonFormColumns = map[string]struct{}{
	"id":          {},
	"create_time": {},
	"update_time": {},
	"create_by":   {},
	"update_by":   {},
	"is_deleted":  {},
}

func isEditableFormColumn(columnName string) bool {
	_, reserved := nonFormColumns[strings.ToLower(strings.TrimSpace(columnName))]
	return !reserved
}

// identifierPatterns 生成产物输出名约束：实体为 Go 导出类型（MixedCaps 大写驼峰），
// 模块为 Go 包目录名（全小写无下划线），字段为驼峰字母数字
var (
	fieldNamePattern  = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`)
	entityNamePattern = regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`)
	moduleNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*$`)
)

func usesTypes(fields []templateFieldConfig) bool {
	for i := range fields {
		if strings.HasPrefix(fields[i].GoType, "types.") {
			return true
		}
	}
	return false
}

func hasColumn(fields []templateFieldConfig, columnName string) bool {
	for i := range fields {
		if strings.EqualFold(strings.TrimSpace(fields[i].ColumnName), columnName) {
			return true
		}
	}
	return false
}

const (
	tplAPI         templateName = "API"
	tplAPITypes    templateName = "API_TYPES"
	tplView        templateName = "VIEW"
	tplHandler     templateName = "Handler"
	tplService     templateName = "Service"
	tplRepository  templateName = "Repository"
	tplModelEntity templateName = "ModelEntity"
	tplModelForm   templateName = "ModelForm"
	tplModelQuery  templateName = "ModelQuery"
	tplModelVo     templateName = "ModelVo"
	tplRouter      templateName = "Router"
)

// codegenConfig 全局代码生成参数
var codegenConfig = struct {
	downloadFileName         string
	backendAppName           string
	frontendAppName          string
	defaultAuthor            string
	defaultModuleName        string
	defaultPackageName       string
	defaultRemoveTablePrefix string
}{
	downloadFileName:         "youlai-admin-code.zip",
	backendAppName:           "youlai-gin",
	frontendAppName:          "vue3-element-admin",
	defaultAuthor:            "youlaitech",
	defaultModuleName:        "system",
	defaultPackageName:       "internal",
	defaultRemoveTablePrefix: "sys_",
}

// permissionNamespaces 模块名 → 权限前缀，未配置时用模块名
var permissionNamespaces = map[string]string{"system": "sys"}

// ResolvePermPrefix 权限标识前缀（如 sys:user）：后端守卫、前端 v-hasPerm、菜单按钮权限共用同一个值
func ResolvePermPrefix(moduleName, entityName string) string {
	namespace, ok := permissionNamespaces[moduleName]
	if !ok {
		namespace = moduleName
	}
	return namespace + ":" + utils.KebabCase(entityName)
}

// trimTablePrefix 去掉表前缀，避免 sys_user 推导出 SysUser 这种把模块名重复进实体名的结果
func trimTablePrefix(tableName, removePrefix string) string {
	if removePrefix != "" && strings.HasPrefix(tableName, removePrefix) {
		return strings.TrimPrefix(tableName, removePrefix)
	}
	return tableName
}

// templateOrder 预览与下载的固定输出顺序：模型层在前，前端模板在后
var templateOrder = []templateName{
	tplModelEntity, tplModelForm, tplModelQuery, tplModelVo,
	tplRepository, tplService, tplHandler, tplRouter,
	tplAPI, tplAPITypes, tplView,
}

// templateConfigs 各层代码模板映射：backend 为 Go 后端，frontend 为 Vue 前端
var templateConfigs = map[templateName]templateConfig{
	tplAPI:         {templatePath: "frontend/ts/api.ts.velty", subpackageName: "api", extension: ".ts"},
	tplAPITypes:    {templatePath: "frontend/ts/api-types.ts.velty", subpackageName: "types", extension: ".ts"},
	tplView:        {templatePath: "frontend/ts/index.vue.velty", subpackageName: "views", extension: ".vue"},
	tplHandler:     {templatePath: "backend/handler.go.velty", subpackageName: "handler", extension: ".go"},
	tplService:     {templatePath: "backend/service.go.velty", subpackageName: "service", extension: ".go"},
	tplRepository:  {templatePath: "backend/repository.go.velty", subpackageName: "repository", extension: ".go"},
	tplModelEntity: {templatePath: "backend/model-entity.go.velty", subpackageName: "model", extension: ".go"},
	tplModelForm:   {templatePath: "backend/model-form.go.velty", subpackageName: "model", extension: ".go"},
	tplModelQuery:  {templatePath: "backend/model-query.go.velty", subpackageName: "model", extension: ".go"},
	tplModelVo:     {templatePath: "backend/model-vo.go.velty", subpackageName: "model", extension: ".go"},
	tplRouter:      {templatePath: "backend/router.go.velty", subpackageName: "", extension: ".go"},
}

// CodegenService 代码生成业务逻辑层
// 所有数据表/配置查询与生成走 Repository，模板渲染与类型映射留 service
type CodegenService struct {
	repo *repository.Repository
}

// NewCodegenService 创建 CodegenService 实例
func NewCodegenService(repo *repository.Repository) *CodegenService {
	return &CodegenService{repo: repo}
}

// GetTablePage 分页查询可代码生成的数据表列表
func (s *CodegenService) GetTablePage(ctx context.Context, query *model.TableQuery) (*commonModel.PagedData, error) {
	return s.repo.TableListPage(ctx, query.Keywords, query.GetOffset(), query.GetLimit())
}

// GetGenConfig 读取指定表的代码生成配置，未配置时按表结构生成默认配置
func (s *CodegenService) GetGenConfig(ctx context.Context, tableName string) (*model.GenConfigForm, error) {
	cfg, found, err := s.repo.GetGenTable(ctx, tableName)
	if err != nil {
		return nil, err
	}

	if found {
		fields, err := s.repo.GetGenColumns(ctx, cfg.ID)
		if err != nil {
			return nil, err
		}

		removePrefix := defaultStr(cfg.RemoveTablePrefix, codegenConfig.defaultRemoveTablePrefix)

		resp := &model.GenConfigForm{
			ID:                cfg.ID,
			TableName:         cfg.TableName,
			ModuleName:        defaultStr(cfg.ModuleName, codegenConfig.defaultModuleName),
			PackageName:       codegenConfig.defaultPackageName,
			BusinessName:      cfg.BusinessName,
			EntityName:        defaultStr(cfg.EntityName, toPascalCase(trimTablePrefix(cfg.TableName, removePrefix))),
			Author:            cfg.Author,
			ParentMenuId:      cfg.ParentMenuID,
			BackendAppName:    codegenConfig.backendAppName,
			FrontendAppName:   codegenConfig.frontendAppName,
			PageType:          defaultStr(cfg.PageType, "crud"),
			RemoveTablePrefix: removePrefix,
			FieldConfigs:      make([]model.FieldConfigForm, 0, len(fields)),
		}

		for _, f := range fields {
			resp.FieldConfigs = append(resp.FieldConfigs, model.FieldConfigForm{
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
	tableComment, err := s.repo.TableComment(ctx, tableName)
	if err != nil {
		tableComment = ""
	}

	businessName := strings.TrimSpace(strings.ReplaceAll(tableComment, "表", ""))
	if businessName == "" {
		businessName = tableName
	}

	removePrefix := codegenConfig.defaultRemoveTablePrefix
	entityName := toPascalCase(trimTablePrefix(tableName, removePrefix))

	cols, err := s.repo.TableColumns(ctx, tableName)
	if err != nil {
		return nil, err
	}

	fieldConfigs := make([]model.FieldConfigForm, 0, len(cols))
	for i, col := range cols {
		fieldType := getFieldTypeByColumnType(col.ColumnType)
		isRequired := 1
		if strings.ToUpper(col.IsNullable) == "YES" {
			isRequired = 0
		}
		sort := i + 1
		fieldConfigs = append(fieldConfigs, model.FieldConfigForm{
			ColumnName:    col.ColumnName,
			ColumnType:    col.ColumnType,
			FieldName:     toCamelCase(col.ColumnName),
			FieldType:     fieldType,
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

	return &model.GenConfigForm{
		TableName:         tableName,
		BusinessName:      businessName,
		ModuleName:        codegenConfig.defaultModuleName,
		PackageName:       codegenConfig.defaultPackageName,
		EntityName:        entityName,
		Author:            codegenConfig.defaultAuthor,
		BackendAppName:    codegenConfig.backendAppName,
		FrontendAppName:   codegenConfig.frontendAppName,
		PageType:          "crud",
		RemoveTablePrefix: removePrefix,
		FieldConfigs:      fieldConfigs,
	}, nil
}

// SaveGenConfig 新增或更新代码生成配置及字段，并联动生成菜单
func (s *CodegenService) SaveGenConfig(ctx context.Context, tableName string, body *model.GenConfigForm) error {
	if body == nil {
		return errs.BadRequest("参数错误")
	}
	if len(body.FieldConfigs) == 0 {
		return errs.BadRequest("字段配置不能为空")
	}

	// 列名必须真实存在于表结构，防止伪造列生成无效代码
	cols, err := s.repo.TableColumns(ctx, tableName)
	if err != nil {
		return err
	}
	columnSet := make(map[string]struct{}, len(cols))
	for _, col := range cols {
		columnSet[col.ColumnName] = struct{}{}
	}
	for _, f := range body.FieldConfigs {
		if _, ok := columnSet[f.ColumnName]; !ok {
			return errs.BadRequest("列 " + f.ColumnName + " 不存在于表 " + tableName)
		}
	}

	now := time.Now()
	cfg, found, err := s.repo.GetGenTable(ctx, tableName)
	if err != nil {
		return err
	}

	moduleName := defaultStr(body.ModuleName, codegenConfig.defaultModuleName)
	packageName := codegenConfig.defaultPackageName
	businessName := defaultStr(body.BusinessName, tableName)
	removePrefix := defaultStr(body.RemoveTablePrefix, codegenConfig.defaultRemoveTablePrefix)
	entityName := defaultStr(body.EntityName, toPascalCase(trimTablePrefix(tableName, removePrefix)))
	author := defaultStr(body.Author, codegenConfig.defaultAuthor)
	pageType := defaultStr(body.PageType, "crud")

	// 输出名进入 Go 类型、目录与导入路径，必须是合法标识符
	if !entityNamePattern.MatchString(entityName) {
		return errs.BadRequest("实体名须为 PascalCase 字母数字组合")
	}
	if !moduleNamePattern.MatchString(moduleName) {
		return errs.BadRequest("模块名须为全小写字母数字")
	}

	updates := map[string]interface{}{
		"module_name":         moduleName,
		"package_name":        packageName,
		"business_name":       businessName,
		"entity_name":         entityName,
		"author":              author,
		"parent_menu_id":      body.ParentMenuId,
		"remove_table_prefix": removePrefix,
		"page_type":           pageType,
		"update_time":         now,
		"is_deleted":          0,
	}

	var tableID int64

	columns := make([]repository.GenColumn, 0, len(body.FieldConfigs))
	for i := range body.FieldConfigs {
		f := body.FieldConfigs[i]
		sort := i + 1
		if f.FieldSort != nil {
			sort = *f.FieldSort
		}

		fieldName := defaultStr(f.FieldName, toCamelCase(f.ColumnName))
		if !fieldNamePattern.MatchString(fieldName) {
			return errs.BadRequest("字段名 " + fieldName + " 须为驼峰字母数字组合")
		}
		columns = append(columns, repository.GenColumn{
			ColumnName:    f.ColumnName,
			ColumnType:    f.ColumnType,
			FieldName:     fieldName,
			FieldType:     defaultStr(f.FieldType, getFieldTypeByColumnType(f.ColumnType)),
			FieldSort:     &sort,
			FieldComment:  f.FieldComment,
			MaxLength:     f.MaxLength,
			IsRequired:    defaultInt(f.IsRequired, 0),
			IsShowInList:  defaultInt(f.IsShowInList, 0),
			IsShowInForm:  defaultInt(f.IsShowInForm, 0),
			IsShowInQuery: defaultInt(f.IsShowInQuery, 0),
			QueryType:     defaultInt(f.QueryType, 1),
			FormType:      defaultInt(f.FormType, 1),
			DictType:      f.DictType,
		})
	}

	// 表配置与字段同一事务写入，中途失败整体回滚，避免留下无字段的孤儿配置
	if err := s.repo.Transaction(ctx, func(txRepo *repository.Repository) error {
		if found && cfg != nil && cfg.ID > 0 {
			tableID = cfg.ID
			if err := txRepo.UpdateGenTable(ctx, tableID, updates); err != nil {
				return err
			}
		} else {
			inserts := updates
			inserts["table_name"] = tableName
			inserts["create_time"] = now
			newID, err := txRepo.CreateGenTable(ctx, inserts)
			if err != nil {
				return err
			}
			tableID = newID
		}
		for i := range columns {
			columns[i].TableID = tableID
		}
		return txRepo.ReplaceGenColumns(ctx, tableID, columns)
	}); err != nil {
		return err
	}

	// 菜单与按钮权限和配置一起保存，失败直接返回，避免留下“有配置无权限”的半成品；重试按路由名幂等
	if body.ParentMenuId != nil && *body.ParentMenuId > 0 {
		return menuService.AddMenuForCodegen(*body.ParentMenuId, moduleName, businessName, entityName, ResolvePermPrefix(moduleName, entityName))
	}

	return nil
}

// DeleteGenConfig 逻辑删除代码生成配置（置 is_deleted=1）
func (s *CodegenService) DeleteGenConfig(ctx context.Context, tableName string) error {
	return s.repo.DeleteGenConfig(ctx, tableName)
}

//go:embed prompts/codegen/system.md
var codegenSystemPrompt string

// AiFillConfig 用 AI 推断代码生成配置，未推断出的字段沿用原有默认值
func (s *CodegenService) AiFillConfig(ctx context.Context, tableName string, requirement string) (*model.GenConfigForm, error) {
	config, err := s.GetGenConfig(ctx, tableName)
	if err != nil {
		return nil, err
	}

	var aiConfig model.GenConfigAiVO
	if err := ai.ChatJSON(ctx, codegenSystemPrompt, buildCodegenUserPrompt(config, requirement), &aiConfig); err != nil {
		return nil, err
	}

	if strings.TrimSpace(aiConfig.BusinessName) != "" {
		config.BusinessName = strings.TrimSpace(aiConfig.BusinessName)
	}
	applyAiFieldConfigs(config, aiConfig.FieldConfigs)
	return config, nil
}

// buildCodegenUserPrompt 构造用户提示词，附带表结构与字段清单
func buildCodegenUserPrompt(config *model.GenConfigForm, requirement string) string {
	columns := make([]map[string]any, 0, len(config.FieldConfigs))
	for _, field := range config.FieldConfigs {
		columns = append(columns, map[string]any{
			"columnName":   field.ColumnName,
			"columnType":   field.ColumnType,
			"fieldComment": field.FieldComment,
			"isRequired":   field.IsRequired,
		})
	}
	payload, _ := json.Marshal(columns)
	return fmt.Sprintf("表名：%s\n当前业务名：%s\n补充说明：%s\n字段列表：\n%s\n",
		config.TableName, config.BusinessName, defaultStr(requirement, "无"), string(payload))
}

// applyAiFieldConfigs 按列名匹配回填字段配置
func applyAiFieldConfigs(config *model.GenConfigForm, aiFields []model.GenConfigAiField) {
	if len(aiFields) == 0 || len(config.FieldConfigs) == 0 {
		return
	}

	aiFieldMap := make(map[string]model.GenConfigAiField, len(aiFields))
	for _, field := range aiFields {
		if strings.TrimSpace(field.ColumnName) != "" {
			aiFieldMap[field.ColumnName] = field
		}
	}

	for i := range config.FieldConfigs {
		field := &config.FieldConfigs[i]
		aiField, ok := aiFieldMap[field.ColumnName]
		if !ok {
			continue
		}
		if strings.TrimSpace(aiField.FieldComment) != "" {
			field.FieldComment = strings.TrimSpace(aiField.FieldComment)
		}
		if strings.TrimSpace(aiField.DictType) != "" {
			field.DictType = strings.TrimSpace(aiField.DictType)
		}
		if value, ok := formTypeValue(aiField.FormType); ok {
			field.FormType = value
		}
		if value, ok := queryTypeValue(aiField.QueryType); ok {
			field.QueryType = value
		}
		field.IsRequired = pickAiInt(aiField.IsRequired, field.IsRequired)
		field.IsShowInList = pickAiInt(aiField.IsShowInList, field.IsShowInList)
		field.IsShowInForm = pickAiInt(aiField.IsShowInForm, field.IsShowInForm)
		field.IsShowInQuery = pickAiInt(aiField.IsShowInQuery, field.IsShowInQuery)
	}
}

// pickAiInt AI 未给出时沿用原值
func pickAiInt(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

// formTypeValue 表单类型枚举名转为存储值
func formTypeValue(name string) (int, bool) {
	values := map[string]int{
		"INPUT": 1, "SELECT": 2, "RADIO": 3, "CHECK_BOX": 4, "INPUT_NUMBER": 5,
		"SWITCH": 6, "TEXT_AREA": 7, "DATE": 8, "DATE_TIME": 9, "HIDDEN": 10,
	}
	value, ok := values[strings.ToUpper(strings.TrimSpace(name))]
	return value, ok
}

// queryTypeValue 查询类型枚举名转为存储值
func queryTypeValue(name string) (int, bool) {
	values := map[string]int{
		"EQ": 1, "LIKE": 2, "IN": 3, "BETWEEN": 4, "GT": 5,
		"GE": 6, "LT": 7, "LE": 8, "NE": 9, "LIKE_LEFT": 10,
	}
	value, ok := values[strings.ToUpper(strings.TrimSpace(name))]
	return value, ok
}

// GetPreview 按模板渲染指定表的各层代码预览
func (s *CodegenService) GetPreview(ctx context.Context, tableName string, pageType string, typeParam string) ([]model.CodegenPreviewVO, error) {
	cfg, err := s.GetGenConfig(ctx, tableName)
	if err != nil {
		return nil, err
	}

	frontendType := "ts"
	if strings.ToLower(strings.TrimSpace(typeParam)) == "js" {
		frontendType = "js"
	}

	previews := make([]model.CodegenPreviewVO, 0)
	for _, name := range templateOrder {
		tc := templateConfigs[name]
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
		content = formatGeneratedSource(content, fileName)

		previews = append(previews, model.CodegenPreviewVO{
			Path:     filepath.ToSlash(filePath),
			FileName: fileName,
			Content:  content,
			Scope:    resolveScope(name),
			Language: resolveLanguage(fileName),
		})
	}

	return previews, nil
}

// DownloadZip 生成多表代码压缩包，返回文件名与压缩内容
func (s *CodegenService) DownloadZip(ctx context.Context, tableNames []string, pageType string, typeParam string) (string, []byte, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	for _, t := range tableNames {
		list, err := s.GetPreview(ctx, t, pageType, typeParam)
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

func resolveFrontendTemplatePath(name templateName, tc templateConfig, frontendType string) string {
	if frontendType == "js" {
		switch name {
		case tplAPI:
			return "frontend/js/api.js.velty"
		case tplView:
			return "frontend/js/index.vue.velty"
		default:
			return tc.templatePath
		}
	}

	switch name {
	case tplAPI:
		return "frontend/ts/api.ts.velty"
	case tplAPITypes:
		return "frontend/ts/api-types.ts.velty"
	case tplView:
		return "frontend/ts/index.vue.velty"
	default:
		return tc.templatePath
	}
}

func resolveFrontendExtension(name templateName, tc templateConfig, frontendType string) string {
	if frontendType != "js" {
		return tc.extension
	}
	if name == tplAPI {
		return ".js"
	}
	return tc.extension
}

func resolveScope(name templateName) string {
	if name == tplAPI || name == tplAPITypes || name == tplView {
		return "frontend"
	}
	return "backend"
}

func resolveLanguage(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	return strings.TrimPrefix(ext, ".")
}

// formatGeneratedSource 对生成的代码做基础整理：
//   - 行尾统一为 LF（模板文件可能是 CRLF）；
//   - Go 代码执行 gofmt，保证模板控制指令不影响产物格式；
//   - 前端代码折叠连续空行，避免控制指令留下成片空白。
func formatGeneratedSource(content, fileName string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	if strings.EqualFold(filepath.Ext(fileName), ".go") {
		formatted, err := format.Source([]byte(content))
		if err != nil {
			logger.Log.Sugar().Warnf("生成代码格式化失败: %s, 错误: %s", fileName, err.Error())
			return content
		}
		return string(formatted)
	}
	return compactBlankLines(content)
}

func compactBlankLines(content string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))
	blank := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			blank = 0
			result = append(result, line)
			continue
		}
		blank++
		if blank == 1 {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

func renderTemplate(
	name templateName,
	templatePath string,
	subpackageName string,
	cfg *model.GenConfigForm,
	pageType string,
) (string, error) {
	effectivePath := templatePath
	if name == tplView && pageType == "crud" {
		if strings.HasSuffix(effectivePath, "index.vue.velty") {
			effectivePath = strings.Replace(effectivePath, "index.vue.velty", "index.crud.vue.velty", 1)
		}
	}

	absPath := resolveBootTemplatePath(effectivePath)
	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", errs.SystemError("读取模板失败: " + absPath)
	}

	logger.Log.Sugar().Debugf("编译模板: %s (%d bytes)", absPath, len(content))
	planner := velty.New()
	_ = planner.RegisterFunction("trim", strings.TrimSpace)
	_ = planner.RegisterFunction("contains", strings.Contains)

	_ = planner.DefineVariable("packageName", "")
	_ = planner.DefineVariable("moduleName", "")
	_ = planner.DefineVariable("subpackageName", "")
	_ = planner.DefineVariable("date", "")
	_ = planner.DefineVariable("entityName", "")
	_ = planner.DefineVariable("tableName", "")
	_ = planner.DefineVariable("author", "")
	_ = planner.DefineVariable("entityLowerCamel", "")
	_ = planner.DefineVariable("entityKebab", "")
	_ = planner.DefineVariable("entityKebabPlural", "")
	_ = planner.DefineVariable("entityPackage", "")
	_ = planner.DefineVariable("entityUpperSnake", "")
	_ = planner.DefineVariable("entitySnake", "")
	_ = planner.DefineVariable("permissionPrefix", "")
	_ = planner.DefineVariable("businessName", "")
	_ = planner.DefineVariable("entityComment", "")
	_ = planner.DefineVariable("fieldConfigs", reflect.TypeOf([]templateFieldConfig{}))
	_ = planner.DefineVariable("fieldConfigsInList", reflect.TypeOf([]templateFieldConfig{}))
	_ = planner.DefineVariable("fieldConfigsInForm", reflect.TypeOf([]templateFieldConfig{}))
	_ = planner.DefineVariable("fieldConfigsInQuery", reflect.TypeOf([]templateFieldConfig{}))
	_ = planner.DefineVariable("useTypesInEntity", false)
	_ = planner.DefineVariable("useTypesInList", false)
	_ = planner.DefineVariable("useTypesInForm", false)
	_ = planner.DefineVariable("useTypesInQuery", false)
	_ = planner.DefineVariable("hasIsDeleted", false)
	_ = planner.DefineVariable("hasCreateTime", false)

	exec, newState, err := planner.Compile(content)
	if err != nil {
		logger.Log.Sugar().Errorf("编译失败: %s, 错误: %s", absPath, err.Error())
		return "", errs.SystemError("编译模板失败[" + effectivePath + "]: " + err.Error())
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
	_ = state.SetValue("entityKebab", utils.KebabCase(cfg.EntityName))
	_ = state.SetValue("entityKebabPlural", utils.PluralKebab(cfg.EntityName))
	_ = state.SetValue("entityPackage", toGoPackageName(cfg.EntityName))
	_ = state.SetValue("entityUpperSnake", toSnakeUpper(cfg.EntityName))
	_ = state.SetValue("entitySnake", toSnakeLower(cfg.EntityName))
	_ = state.SetValue("permissionPrefix", ResolvePermPrefix(cfg.ModuleName, cfg.EntityName))
	_ = state.SetValue("businessName", cfg.BusinessName)
	_ = state.SetValue("entityComment", cfg.BusinessName)

	fields := make([]templateFieldConfig, 0, len(cfg.FieldConfigs))
	for i := range cfg.FieldConfigs {
		f := cfg.FieldConfigs[i]
		fieldType := defaultStr(f.FieldType, getFieldTypeByColumnType(f.ColumnType))
		goType := getGoTypeByColumnType(f.ColumnType)
		tsType := getTsTypeByFieldType(fieldType)
		formTypeName := getFormTypeName(f.FormType)
		queryTypeName := getQueryTypeName(f.QueryType)
		if goType != "string" && strings.HasPrefix(queryTypeName, "LIKE") {
			queryTypeName = "EQ"
		}
		queryBaseType := goType
		if queryBaseType == "types.LocalTime" {
			// 查询参数沿用字符串，前端按 YYYY-MM-DD[ HH:mm:ss] 传值，避免自定义时间类型缺少表单反序列化入口。
			queryBaseType = "string"
		}
		queryGoType := queryBaseType
		if queryTypeName == "BETWEEN" || queryTypeName == "IN" {
			queryGoType = "[]" + queryBaseType
		} else if queryBaseType != "string" {
			queryGoType = "*" + queryBaseType
		}
		tsQueryType := tsType
		if queryTypeName == "BETWEEN" {
			tsQueryType = "[" + tsType + ", " + tsType + "]"
		} else if queryTypeName == "IN" {
			tsQueryType = tsType + "[]"
		}
		hasDict := strings.TrimSpace(f.DictType) != ""
		visibleInForm := f.IsShowInForm == 1 && formTypeName != "HIDDEN"
		gormTag := "column:" + f.ColumnName
		jsonTag := f.FieldName
		switch strings.ToLower(strings.TrimSpace(f.ColumnName)) {
		case "id":
			gormTag += ";primaryKey;autoIncrement"
		case "create_time":
			gormTag += ";autoCreateTime"
		case "update_time":
			gormTag += ";autoUpdateTime"
		case "is_deleted":
			gormTag += ";default:0"
			jsonTag = "-"
		}
		queryIsSetExpr := "query." + toGoFieldName(f.FieldName) + ` != ""`
		queryArgExpr := "query." + toGoFieldName(f.FieldName)
		if queryBaseType != "string" {
			queryIsSetExpr = "query." + toGoFieldName(f.FieldName) + " != nil"
			queryArgExpr = "*" + queryArgExpr
		}
		fields = append(fields, templateFieldConfig{
			ColumnName:     f.ColumnName,
			ColumnType:     f.ColumnType,
			FieldName:      f.FieldName,
			GoFieldName:    toGoFieldName(f.FieldName),
			FieldType:      fieldType,
			FieldComment:   f.FieldComment,
			IsShowInList:   f.IsShowInList,
			IsShowInForm:   f.IsShowInForm,
			IsShowInQuery:  f.IsShowInQuery,
			IsRequired:     f.IsRequired,
			FormType:       formTypeName,
			QueryType:      queryTypeName,
			MaxLength:      f.MaxLength,
			FieldSort:      f.FieldSort,
			DictType:       f.DictType,
			TsType:         tsType,
			GoType:         goType,
			QueryGoType:    queryGoType,
			TsQueryType:    tsQueryType,
			GormTag:        gormTag,
			JSONTag:        jsonTag,
			QueryIsSetExpr: queryIsSetExpr,
			QueryArgExpr:   queryArgExpr,

			HasDictType:     hasDict,
			IsDate:          formTypeName == "DATE" || formTypeName == "DATE_TIME",
			IsDateTime:      formTypeName == "DATE_TIME",
			IsNumeric:       fieldType == "Integer" || fieldType == "Long" || fieldType == "BigDecimal",
			IsDecimal:       fieldType == "BigDecimal" || fieldType == "Double",
			VisibleInForm:   visibleInForm,
			RequiredInForm:  visibleInForm && f.IsRequired == 1,
			BindingRequired: f.IsRequired == 1 && goType == "string",
			ShowDictInList:  f.IsShowInList == 1 && hasDict,
			ShowDictInQuery: f.IsShowInQuery == 1 && hasDict,
			ShowDictInForm:  visibleInForm && hasDict,
			IsTimeColumn:    goType == "types.LocalTime",
		})
	}

	fieldsInList := make([]templateFieldConfig, 0, len(fields))
	fieldsInForm := make([]templateFieldConfig, 0, len(fields))
	fieldsInQuery := make([]templateFieldConfig, 0, len(fields))
	for i := range fields {
		f := fields[i]
		if f.IsShowInList == 1 && !strings.EqualFold(strings.TrimSpace(f.ColumnName), "is_deleted") {
			fieldsInList = append(fieldsInList, f)
		}
		if f.VisibleInForm && isEditableFormColumn(f.ColumnName) {
			fieldsInForm = append(fieldsInForm, f)
		}
		if f.IsShowInQuery == 1 {
			fieldsInQuery = append(fieldsInQuery, f)
		}
	}

	_ = state.SetValue("fieldConfigs", fields)
	_ = state.SetValue("fieldConfigsInList", fieldsInList)
	_ = state.SetValue("fieldConfigsInForm", fieldsInForm)
	_ = state.SetValue("fieldConfigsInQuery", fieldsInQuery)
	_ = state.SetValue("useTypesInEntity", usesTypes(fields))
	_ = state.SetValue("useTypesInList", usesTypes(fieldsInList))
	_ = state.SetValue("useTypesInForm", usesTypes(fieldsInForm))
	_ = state.SetValue("useTypesInQuery", usesTypes(fieldsInQuery))
	_ = state.SetValue("hasIsDeleted", hasColumn(fields, "is_deleted"))
	_ = state.SetValue("hasCreateTime", hasColumn(fields, "create_time"))

	exec.Exec(state)
	if !state.IsValid() {
		logger.Log.Sugar().Errorf("渲染失败: %s, state=%v", absPath, state.IsValid())
		return "", errs.SystemError("渲染模板失败: " + absPath)
	}
	result := state.Buffer.String()
	logger.Log.Sugar().Debugf("渲染成功: %s, 输出=%d bytes", absPath, len(result))
	return result, nil
}

func resolveBootTemplatePath(templatePath string) string {
	return filepath.Join("internal", "codegen", "templates", filepath.FromSlash(templatePath))
}

func getFileName(entityName string, name templateName, extension string) string {
	if name == tplAPI {
		return "index" + extension
	}
	if name == tplAPITypes {
		return "types" + extension
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
		return filepath.Join(frontend, "src", subpackageName, moduleName, utils.KebabCase(entityName))
	}
	if name == tplAPITypes {
		return filepath.Join(frontend, "src", "api", moduleName, utils.KebabCase(entityName))
	}
	if name == tplView {
		return filepath.Join(frontend, "src", subpackageName, moduleName, utils.KebabCase(entityName))
	}

	base := filepath.Join(backend, "internal", moduleName, utils.KebabCase(entityName))
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

// toGoPackageName 生成合法 Go 包标识符：目录可用 kebab-case，但 package 声明不能含连字符。
func toGoPackageName(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "module"
	}
	return b.String()
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
	case "datetime", "timestamp":
		return "types.LocalTime"
	case "date":
		return "string"
	case "boolean", "bool", "bit":
		return "bool"
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

// getFieldTypeByColumnType 将数据库列类型映射为通用字段类型（前端表单项与 TS 类型的推导基础）。
// 该类型表示"字段在业务域的抽象类型"，而非底层语言的类型——生成器据此派生 tsType，
// 后端 Go 字段类型由 getGoTypeByColumnType 单独映射，避免 Go/TS 类型耦合在同一处。
func getFieldTypeByColumnType(columnType string) string {
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

func getTsTypeByFieldType(fieldType string) string {
	switch fieldType {
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
