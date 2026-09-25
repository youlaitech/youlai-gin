package service

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"youlai-gin/internal/common/config"
	cRedis "youlai-gin/internal/common/redis"
	"youlai-gin/internal/system/form/model"
	"youlai-gin/internal/system/form/repository"
	menuModel "youlai-gin/internal/system/menu/model"
	"youlai-gin/pkg/errs"
	baseModel "youlai-gin/pkg/model"
	"youlai-gin/pkg/types"
)

//go:embed prompts/form/system.md
var formSystemPrompt string

// menuEntity 表单访问菜单实体（复用系统菜单模型）
type menuEntity = menuModel.Menu

const (
	statusDraft     = 0
	statusPublished = 1
	statusDisabled  = -1

	defaultCatalogName = "表单中心"
	formAdminName      = "动态表单"
	publicSubmitLimit  = 10
	renderCacheTTL     = 30 * time.Minute
)

// Service 动态表单业务
type Service struct {
	repo *repository.Repository
}

// NewService 创建动态表单服务
func NewService(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

// ── 规则校验 ──────────────────────────────────────────────────

func checkNodes(nodes []any) error {
	for _, raw := range nodes {
		node, ok := raw.(map[string]any)
		if !ok {
			return errs.BadRequest("表单规则格式非法：节点必须为对象")
		}
		if _, ok := node["type"].(string); !ok {
			return errs.BadRequest("表单规则格式非法：节点缺少 type")
		}
		children, _ := node["children"].([]any)
		if len(children) > 0 {
			if err := checkNodes(children); err != nil {
				return err
			}
			continue
		}
		if _, ok := node["field"].(string); !ok {
			return errs.BadRequest("表单规则格式非法：字段节点缺少 field")
		}
	}
	return nil
}

// ValidateRule 校验表单规则结构
func ValidateRule(rules []any) error {
	if len(rules) == 0 {
		return errs.BadRequest("表单规则格式非法：应为非空数组")
	}
	return checkNodes(rules)
}

func isEmptyValue(value any) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []any:
		return len(v) == 0
	}
	return false
}

func matchType(kind string, value any) bool {
	switch kind {
	case "number", "integer", "float":
		switch value.(type) {
		case float64, int, int64:
			return true
		}
		return false
	case "array":
		_, ok := value.([]any)
		return ok
	}
	return true
}

func walkNodes(nodes []any, data, filtered map[string]any, errList *[]string) {
	for _, raw := range nodes {
		node, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if field, ok := node["field"].(string); ok && field != "" {
			// 字段白名单：规则未声明的 field 直接丢弃
			if value, exists := data[field]; exists {
				filtered[field] = value
			}
			title, _ := node["title"].(string)
			if title == "" {
				title = field
			}
			value := data[field]
			validates, _ := node["validate"].([]any)
			for _, vr := range validates {
				rule, ok := vr.(map[string]any)
				if !ok {
					continue
				}
				if required, _ := rule["required"].(bool); required && isEmptyValue(value) {
					if msg, _ := rule["message"].(string); msg != "" {
						*errList = append(*errList, msg)
					} else {
						*errList = append(*errList, fmt.Sprintf("「%s」不能为空", title))
					}
				}
				if kind, _ := rule["type"].(string); kind != "" && !isEmptyValue(value) && !matchType(kind, value) {
					*errList = append(*errList, fmt.Sprintf("「%s」数据类型非法", title))
				}
			}
		}
		if children, ok := node["children"].([]any); ok && len(children) > 0 {
			walkNodes(children, data, filtered, errList)
		}
	}
}

// ValidateAndFilter 按规则白名单过滤提交数据并校验必填与类型
func ValidateAndFilter(rules []any, data map[string]any) (map[string]any, error) {
	filtered := make(map[string]any)
	var list []string
	walkNodes(rules, data, filtered, &list)
	if len(list) > 0 {
		return nil, errs.BadRequest(strings.Join(list, "；"))
	}
	return filtered, nil
}

// ── 表单定义 ──────────────────────────────────────────────────

// DefinitionPage 表单定义分页列表
func (s *Service) DefinitionPage(ctx context.Context, query *model.FormDefinitionQuery) (*baseModel.PagedData, error) {
	list, total, err := s.repo.DefinitionPage(ctx, query)
	if err != nil {
		return nil, errs.SystemError("查询表单列表失败")
	}

	voList := make([]model.FormDefinitionVO, len(list))
	for i, item := range list {
		voList[i] = model.FormDefinitionVO{
			ID:          item.ID,
			FormKey:     item.FormKey,
			FormName:    item.FormName,
			Description: item.Description,
			Status:      item.Status,
			IsPublic:    item.IsPublic,
			Category:    item.Category,
			Version:     item.Version,
			CreateTime:  item.CreateTime,
		}
	}
	return &baseModel.PagedData{List: voList, Total: total}, nil
}

// GetDefinitionForm 表单设计数据回显
func (s *Service) GetDefinitionForm(ctx context.Context, id int64) (*model.FormDefinition, error) {
	entity, err := s.repo.GetDefinition(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("表单不存在")
		}
		return nil, errs.SystemError("查询表单失败")
	}
	return entity, nil
}

// CreateDefinition 新增表单定义
func (s *Service) CreateDefinition(ctx context.Context, form *model.FormDefinitionForm) error {
	formKey := strings.TrimSpace(form.FormKey)
	exists, err := s.repo.DefinitionKeyExists(ctx, formKey, 0)
	if err != nil {
		return errs.SystemError("校验表单标识失败")
	}
	if exists {
		return errs.BadRequest("表单标识已存在")
	}

	entity := &model.FormDefinition{
		FormKey:     formKey,
		FormName:    strings.TrimSpace(form.FormName),
		Description: form.Description,
		FormJson:    form.FormJson,
		OptionsJson: form.OptionsJson,
		Category:    form.Category,
		Status:      statusDraft,
		Version:     1,
	}
	if entity.Category == "" {
		entity.Category = "normal"
	}
	if form.IsPublic != nil {
		entity.IsPublic = *form.IsPublic
	}
	if err := s.repo.CreateDefinition(ctx, entity); err != nil {
		return errs.SystemError("创建表单失败")
	}
	form.ID = entity.ID
	return nil
}

// UpdateDefinition 修改表单定义
func (s *Service) UpdateDefinition(ctx context.Context, id int64, form *model.FormDefinitionForm) error {
	entity, err := s.repo.GetDefinition(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("表单不存在")
		}
		return errs.SystemError("查询表单失败")
	}

	if formKey := strings.TrimSpace(form.FormKey); formKey != "" && formKey != entity.FormKey {
		return errs.BadRequest("表单标识创建后不可修改")
	}

	updates := map[string]any{
		"form_name":   strings.TrimSpace(form.FormName),
		"description": form.Description,
		"update_time": time.Now(),
	}
	if form.FormJson != nil {
		updates["form_json"] = form.FormJson
	}
	if form.OptionsJson != nil {
		updates["options_json"] = form.OptionsJson
	}
	if form.IsPublic != nil {
		updates["is_public"] = *form.IsPublic
	}
	if entity.Status == statusDraft && form.Category != "" {
		updates["category"] = form.Category
	}

	if err := s.repo.UpdateDefinition(ctx, id, updates); err != nil {
		return errs.SystemError("修改表单失败")
	}
	if entity.Status == statusPublished {
		s.evictRenderCache(ctx, entity.FormKey)
	}
	return nil
}

// DeleteDefinitions 删除表单（级联逻辑删除数据与快照，并清理生成的菜单）
func (s *Service) DeleteDefinitions(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return errs.BadRequest("删除的表单数据为空")
	}

	for _, id := range ids {
		entity, err := s.repo.GetDefinition(ctx, id)
		if err != nil {
			continue
		}
		if entity.MenuID != nil && *entity.MenuID > 0 {
			if err := s.repo.DeleteMenuWithRoles(ctx, int64(*entity.MenuID)); err != nil {
				return errs.SystemError("删除表单菜单失败")
			}
		}
		s.evictRenderCache(ctx, entity.FormKey)
	}

	if err := s.repo.SoftDeleteDefinitions(ctx, ids); err != nil {
		return errs.SystemError("删除表单失败")
	}
	if err := s.repo.SoftDeleteDataByFormIDs(ctx, ids); err != nil {
		return errs.SystemError("删除表单数据失败")
	}
	if err := s.repo.SoftDeleteSnapshotsByFormIDs(ctx, ids); err != nil {
		return errs.SystemError("删除表单快照失败")
	}
	return nil
}

// Publish 发布表单（版本递增并固化快照）
func (s *Service) Publish(ctx context.Context, id int64) error {
	entity, err := s.repo.GetDefinition(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("表单不存在")
		}
		return errs.SystemError("查询表单失败")
	}

	version := entity.Version + 1
	if err := s.repo.UpdateDefinition(ctx, id, map[string]any{
		"status":      statusPublished,
		"version":     version,
		"update_time": time.Now(),
	}); err != nil {
		return errs.SystemError("发布表单失败")
	}

	// 固化版本快照：数据回显按提交时版本加载，防止规则变更导致历史数据漂移
	if err := s.repo.CreateSnapshot(ctx, &model.FormSnapshot{
		FormID:      entity.ID,
		Version:     version,
		FormJson:    entity.FormJson,
		OptionsJson: entity.OptionsJson,
	}); err != nil {
		return errs.SystemError("保存表单快照失败")
	}

	s.evictRenderCache(ctx, entity.FormKey)
	return nil
}

// Disable 停用表单
func (s *Service) Disable(ctx context.Context, id int64) error {
	if err := s.repo.UpdateDefinition(ctx, id, map[string]any{
		"status":      statusDisabled,
		"update_time": time.Now(),
	}); err != nil {
		return errs.SystemError("停用表单失败")
	}
	return nil
}

// ── 渲染 ──────────────────────────────────────────────────────

func renderCacheKey(formKey string) string {
	return "form:render:" + formKey
}

func (s *Service) evictRenderCache(ctx context.Context, formKey string) {
	if cRedis.Client == nil {
		return
	}
	cRedis.Client.Del(ctx, renderCacheKey(formKey))
}

func renderPayload(entity *model.FormDefinition) *model.FormRenderVO {
	return &model.FormRenderVO{
		FormKey:     entity.FormKey,
		FormName:    entity.FormName,
		Version:     entity.Version,
		FormJson:    entity.FormJson,
		OptionsJson: entity.OptionsJson,
	}
}

// Render 已发布表单渲染规则（Redis 缓存兜底 30 分钟）
func (s *Service) Render(ctx context.Context, formKey string) (*model.FormRenderVO, error) {
	if cRedis.Client != nil {
		if cached, err := cRedis.Client.Get(ctx, renderCacheKey(formKey)).Result(); err == nil && cached != "" {
			var vo model.FormRenderVO
			if json.Unmarshal([]byte(cached), &vo) == nil {
				return &vo, nil
			}
		}
	}

	entity, err := s.repo.GetPublishedByKey(ctx, formKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("表单不存在或未发布")
		}
		return nil, errs.SystemError("查询表单失败")
	}

	vo := renderPayload(entity)
	if cRedis.Client != nil {
		if data, err := json.Marshal(vo); err == nil {
			cRedis.Client.Set(ctx, renderCacheKey(formKey), data, renderCacheTTL)
		}
	}
	return vo, nil
}

// PublicRender 公开表单渲染规则（直查库校验 is_public，不走缓存）
func (s *Service) PublicRender(ctx context.Context, formKey string) (*model.FormRenderVO, error) {
	entity, err := s.repo.GetPublishedByKey(ctx, formKey)
	if err != nil || entity.IsPublic != 1 {
		return nil, errs.NotFound("表单不存在或未开放公开访问")
	}
	return renderPayload(entity), nil
}

// WorkflowOptions 审批表单下拉选项
func (s *Service) WorkflowOptions(ctx context.Context) ([]baseModel.Option[string], error) {
	list, err := s.repo.WorkflowFormList(ctx)
	if err != nil {
		return nil, errs.SystemError("查询表单失败")
	}

	options := make([]baseModel.Option[string], 0, len(list))
	for _, item := range list {
		options = append(options, baseModel.Option[string]{Value: item.FormKey, Label: item.FormName})
	}
	return options, nil
}

// ── 访问菜单 ──────────────────────────────────────────────────

func toUpperCamel(value string) string {
	parts := strings.Split(value, "_")
	var b strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		b.WriteString(strings.ToUpper(part[:1]))
		b.WriteString(part[1:])
	}
	return b.String()
}

func catalogTreePath(parent *menuEntity) string {
	if parent.TreePath == "" {
		return fmt.Sprintf("%d", parent.ID)
	}
	return fmt.Sprintf("%s,%d", parent.TreePath, parent.ID)
}

// GetMenu 表单访问菜单回显
func (s *Service) GetMenu(ctx context.Context, formID int64) (*model.FormMenuVO, error) {
	entity, err := s.repo.GetDefinition(ctx, formID)
	if err != nil {
		return nil, errs.NotFound("表单不存在")
	}
	if entity.MenuID == nil || *entity.MenuID <= 0 {
		return nil, nil
	}

	menu, err := s.repo.GetMenu(ctx, int64(*entity.MenuID))
	if err != nil {
		return nil, nil
	}

	roleIDs, err := s.repo.MenuRoleIDs(ctx, int64(menu.ID))
	if err != nil {
		roleIDs = []int64{}
	}
	return &model.FormMenuVO{
		MenuID:   menu.ID,
		MenuName: menu.Name,
		ParentID: menu.ParentID,
		RoleIDs:  roleIDs,
	}, nil
}

func (s *Service) getOrCreateCatalog(ctx context.Context) (*menuEntity, error) {
	catalog, err := s.repo.FindMenuByNameAndType(ctx, defaultCatalogName, "C")
	if err == nil {
		return catalog, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	parentID := int64(0)
	treePath := "0"
	routePath := "/form-center"

	if adminCatalog, err := s.repo.FindMenuByNameAndType(ctx, formAdminName, "C"); err == nil {
		parentID = int64(adminCatalog.ID)
		treePath = catalogTreePath(adminCatalog)
		routePath = "center"
	}

	created := &menuEntity{
		ParentID:  types.BigInt(parentID),
		TreePath:  treePath,
		Name:      defaultCatalogName,
		Type:      "C",
		RoutePath: routePath,
		Component: "Layout",
		Visible:   1,
		Sort:      5,
		Icon:      "el-icon-EditPen",
	}
	if err := s.repo.CreateMenu(ctx, created); err != nil {
		return nil, err
	}
	return created, nil
}

// SaveMenu 生成或更新表单访问菜单，并按需授权角色（幂等）
func (s *Service) SaveMenu(ctx context.Context, formID int64, form *model.FormMenuForm) error {
	entity, err := s.repo.GetDefinition(ctx, formID)
	if err != nil {
		return errs.NotFound("表单不存在")
	}

	var parent *menuEntity
	if form.ParentID != nil && *form.ParentID > 0 {
		parent, err = s.repo.GetMenu(ctx, *form.ParentID)
		if err != nil {
			return errs.NotFound("上级菜单不存在")
		}
		if parent.Type != "C" {
			return errs.BadRequest("上级菜单必须为目录类型")
		}
	} else {
		parent, err = s.getOrCreateCatalog(ctx)
		if err != nil {
			return errs.SystemError("创建默认目录失败")
		}
	}

	routeName := "FormRender" + toUpperCamel(entity.FormKey)
	treePath := catalogTreePath(parent)
	params := map[string]any{"formKey": entity.FormKey}

	var menuID int64
	if entity.MenuID != nil && *entity.MenuID > 0 {
		if existing, getErr := s.repo.GetMenu(ctx, int64(*entity.MenuID)); getErr == nil {
			existing.ParentID = parent.ID
			existing.Name = strings.TrimSpace(form.MenuName)
			existing.RouteName = routeName
			existing.RoutePath = entity.FormKey
			existing.Component = "dynamic-form/render"
			existing.Params = params
			existing.TreePath = treePath
			if err := s.repo.UpdateMenu(ctx, existing); err != nil {
				return errs.SystemError("更新表单菜单失败")
			}
			menuID = int64(existing.ID)
		}
	}

	if menuID == 0 {
		created := &menuEntity{
			ParentID:  parent.ID,
			TreePath:  treePath,
			Name:      strings.TrimSpace(form.MenuName),
			Type:      "M",
			RouteName: routeName,
			RoutePath: entity.FormKey,
			Component: "dynamic-form/render",
			Params:    params,
			Visible:   1,
			Sort:      1,
		}
		if err := s.repo.CreateMenu(ctx, created); err != nil {
			return errs.SystemError("创建表单菜单失败")
		}
		menuID = int64(created.ID)
	}

	// 回写 menu_id：下次发布走更新语义，防重复生成菜单
	if err := s.repo.UpdateDefinition(ctx, formID, map[string]any{
		"menu_id":     menuID,
		"update_time": time.Now(),
	}); err != nil {
		return errs.SystemError("更新表单失败")
	}

	if len(form.RoleIDs) > 0 {
		if err := s.repo.AssignMenuRoles(ctx, menuID, treePath, form.RoleIDs); err != nil {
			return errs.SystemError("授权表单菜单失败")
		}
	}
	return nil
}

// ── 表单数据 ──────────────────────────────────────────────────

// Submit 提交表单数据（需已发布）
func (s *Service) Submit(ctx context.Context, formKey string, data map[string]any) error {
	entity, err := s.repo.GetDefinitionByKey(ctx, formKey)
	if err != nil {
		return errs.NotFound("表单不存在")
	}
	if entity.Status != statusPublished {
		return errs.BadRequest("表单未发布或已停用")
	}
	return s.persist(ctx, entity, data)
}

// SubmitPublic 匿名提交公开表单
func (s *Service) SubmitPublic(ctx context.Context, formKey string, data map[string]any, ip string) error {
	entity, err := s.repo.GetDefinitionByKey(ctx, formKey)
	if err != nil || entity.IsPublic != 1 || entity.Status != statusPublished {
		return errs.NotFound("表单不存在或未开放公开访问")
	}
	if !s.checkPublicLimit(ctx, formKey, ip) {
		return errs.BadRequest("提交过于频繁，请稍后再试")
	}
	return s.persist(ctx, entity, data)
}

func (s *Service) persist(ctx context.Context, entity *model.FormDefinition, data map[string]any) error {
	filtered, err := ValidateAndFilter(entity.FormJson, data)
	if err != nil {
		return err
	}
	if err := s.repo.CreateData(ctx, &model.FormData{
		FormID:      entity.ID,
		FormVersion: entity.Version,
		DataJson:    filtered,
	}); err != nil {
		return errs.SystemError("保存表单数据失败")
	}
	return nil
}

// checkPublicLimit 公开表单按 formKey + IP 限流（固定窗口计数，Redis 异常时放行）
func (s *Service) checkPublicLimit(ctx context.Context, formKey, ip string) bool {
	if cRedis.Client == nil {
		return true
	}
	key := fmt.Sprintf("form:public:submit:%s:%s", formKey, ip)
	count, err := cRedis.Client.Incr(ctx, key).Result()
	if err != nil {
		return true
	}
	if count == 1 {
		cRedis.Client.Expire(ctx, key, time.Minute)
	}
	return count <= publicSubmitLimit
}

// DataPage 表单数据分页
func (s *Service) DataPage(ctx context.Context, formKey string, query *model.FormDataQuery) (*baseModel.PagedData, error) {
	entity, err := s.repo.GetDefinitionByKey(ctx, formKey)
	if err != nil {
		return nil, errs.NotFound("表单不存在")
	}

	list, total, err := s.repo.DataPage(ctx, int64(entity.ID), query)
	if err != nil {
		return nil, errs.SystemError("查询表单数据失败")
	}

	// 提交人昵称按批查询补全，避免逐行查库
	userIDs := make([]int64, 0, len(list))
	for _, row := range list {
		if row.CreateBy != nil {
			userIDs = append(userIDs, int64(*row.CreateBy))
		}
	}
	nicknameMap, _ := s.repo.NicknamesByUserIDs(ctx, userIDs)

	voList := make([]model.FormDataVO, len(list))
	for i, row := range list {
		vo := model.FormDataVO{
			ID:          row.ID,
			FormVersion: row.FormVersion,
			DataJson:    row.DataJson,
			CreateBy:    row.CreateBy,
		}
		if row.CreateBy != nil {
			vo.CreateByName = nicknameMap[int64(*row.CreateBy)]
		}
		if !time.Time(row.CreateTime).IsZero() {
			vo.CreateTime = time.Time(row.CreateTime).Format("2006-01-02 15:04:05")
		}
		voList[i] = vo
	}
	return &baseModel.PagedData{List: voList, Total: total}, nil
}

// DataDetail 表单数据详情（按提交时版本快照回显规则）
func (s *Service) DataDetail(ctx context.Context, id int64) (*model.FormDataDetailVO, error) {
	row, err := s.repo.GetData(ctx, id)
	if err != nil {
		return nil, errs.NotFound("表单数据不存在")
	}

	var formJson []any
	var optionsJson map[string]any
	if snapshot, snapErr := s.repo.GetSnapshot(ctx, int64(row.FormID), row.FormVersion); snapErr == nil {
		formJson, optionsJson = snapshot.FormJson, snapshot.OptionsJson
	} else if definition, defErr := s.repo.GetDefinition(ctx, int64(row.FormID)); defErr == nil {
		formJson, optionsJson = definition.FormJson, definition.OptionsJson
	}

	vo := &model.FormDataDetailVO{
		FormDataVO: model.FormDataVO{
			ID:          row.ID,
			FormVersion: row.FormVersion,
			DataJson:    row.DataJson,
			CreateBy:    row.CreateBy,
		},
		FormJson:    formJson,
		OptionsJson: optionsJson,
	}
	if !time.Time(row.CreateTime).IsZero() {
		vo.CreateTime = time.Time(row.CreateTime).Format("2006-01-02 15:04:05")
	}
	return vo, nil
}

// DeleteData 逻辑删除表单数据
func (s *Service) DeleteData(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return errs.BadRequest("删除的表单数据为空")
	}
	if err := s.repo.SoftDeleteData(ctx, ids); err != nil {
		return errs.SystemError("删除表单数据失败")
	}
	return nil
}

// ── AI ────────────────────────────────────────────────────────

// AiGenerateRule 用 AI 把需求描述转成 form-create 规则，产出经结构校验后才返回
func (s *Service) AiGenerateRule(ctx context.Context, description string) ([]any, error) {
	content, err := chat(ctx, formSystemPrompt, "需求描述：\n"+description)
	if err != nil {
		return nil, err
	}

	var rules []any
	if err := json.Unmarshal([]byte(content), &rules); err != nil {
		return nil, errs.BadRequest("AI 返回内容非合法 JSON，请重试")
	}
	if err := ValidateRule(rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	ai := config.Cfg.AI
	if !ai.Enabled || ai.ApiKey == "" {
		return "", errs.BadRequest("AI 功能未开启，请配置 ai 段与 apiKey")
	}

	body, err := json.Marshal(map[string]any{
		"model": ai.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"response_format": map[string]string{"type": "json_object"},
	})
	if err != nil {
		return "", errs.SystemError("组装 AI 请求失败")
	}

	url := strings.TrimRight(ai.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", errs.SystemError("构建 AI 请求失败")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ai.ApiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", errs.ThirdPartyError("AI 服务调用失败")
	}
	defer resp.Body.Close()

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil || len(parsed.Choices) == 0 {
		return "", errs.ThirdPartyError("AI 返回内容解析失败")
	}
	return stripCodeFence(parsed.Choices[0].Message.Content), nil
}

func stripCodeFence(content string) string {
	text := strings.TrimSpace(content)
	if !strings.HasPrefix(text, "```") {
		return text
	}
	start := strings.Index(text, "\n")
	end := strings.LastIndex(text, "```")
	if start == -1 || end <= start {
		return text
	}
	return strings.TrimSpace(text[start+1 : end])
}
