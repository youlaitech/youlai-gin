package datascope

import (
	"fmt"
	"strings"
	"unicode"

	"gorm.io/gorm"

	permService "youlai-gin/internal/system/permission/service"
	"youlai-gin/pkg/auth"
)

// 数据权限范围常量
const (
	DataScopeAll             = 1 // 全部数据
	DataScopeDeptAndChildren = 2 // 部门及子部门
	DataScopeDept            = 3 // 本部门
	DataScopeSelf            = 4 // 仅本人
	DataScopeCustom          = 5 // 自定义
)

// DataPermissionConfig 数据权限配置
type DataPermissionConfig struct {
	DeptAlias    string // 部门表别名
	DeptIDColumn string // 部门ID字段名，默认 "dept_id"
	UserAlias    string // 用户表别名
	UserIDColumn string // 用户ID字段名，默认 "create_by"
}

func isSafeIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return false
	}
	return true
}

func toSnakeCase(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// ApplyDataScope 应用数据权限过滤（多角色并集策略）
func ApplyDataScope(db *gorm.DB, user *auth.UserDetails, config DataPermissionConfig) *gorm.DB {
	if user == nil {
		return db.Where("1 = 0")
	}

	for _, role := range user.Roles {
		if role == "ROOT" {
			return db
		}
	}

	dataScopes, err := permService.GetUserDataScopes(user.UserID, user.Roles, int64(user.DeptID))
	if err != nil {
		return db.Where("1 = 0")
	}

	if permService.HasAllDataScope(dataScopes) {
		return db
	}

	if len(dataScopes) == 0 {
		return db.Where("1 = 0")
	}

	deptColumn := config.DeptIDColumn
	if deptColumn == "" {
		deptColumn = "dept_id"
	}
	userColumn := config.UserIDColumn
	if userColumn == "" {
		userColumn = "create_by"
	}

	deptColumn = toSnakeCase(deptColumn)
	userColumn = toSnakeCase(userColumn)
	if !isSafeIdentifier(deptColumn) {
		deptColumn = "dept_id"
	}
	if !isSafeIdentifier(userColumn) {
		userColumn = "create_by"
	}

	if config.DeptAlias != "" && isSafeIdentifier(config.DeptAlias) {
		deptColumn = config.DeptAlias + "." + deptColumn
	}
	if config.UserAlias != "" && isSafeIdentifier(config.UserAlias) {
		userColumn = config.UserAlias + "." + userColumn
	}

	return buildUnionCondition(db, dataScopes, deptColumn, userColumn, user.UserID)
}

func buildUnionCondition(db *gorm.DB, dataScopes []auth.RoleDataScope, deptColumn, userColumn string, userID int64) *gorm.DB {
	var orConditions []string
	var args []interface{}

	for _, ds := range dataScopes {
		cond, condArgs := buildRoleCondition(ds, deptColumn, userColumn, userID)
		if cond != "" {
			orConditions = append(orConditions, cond)
			args = append(args, condArgs...)
		}
	}

	if len(orConditions) == 0 {
		return db.Where("1 = 0")
	}

	if len(orConditions) == 1 {
		return db.Where(orConditions[0], args...)
	}

	unionCond := "(" + orConditions[0] + ")"
	for i := 1; i < len(orConditions); i++ {
		unionCond += " OR (" + orConditions[i] + ")"
	}

	return db.Where(unionCond, args...)
}

func buildRoleCondition(ds auth.RoleDataScope, deptColumn, userColumn string, userID int64) (string, []interface{}) {
	switch ds.DataScope {
	case DataScopeAll:
		return "", nil
	case DataScopeDeptAndChildren, DataScopeDept, DataScopeCustom:
		if len(ds.CustomDeptIDs) > 0 {
			placeholders := make([]string, len(ds.CustomDeptIDs))
			args := make([]interface{}, len(ds.CustomDeptIDs))
			for i, id := range ds.CustomDeptIDs {
				placeholders[i] = "?"
				args[i] = id
			}
			return fmt.Sprintf("%s IN (%s)", deptColumn, joinPlaceholders(placeholders)), args
		}
		return "", nil
	case DataScopeSelf:
		return fmt.Sprintf("%s = ?", userColumn), []interface{}{userID}
	default:
		return "", nil
	}
}

func joinPlaceholders(placeholders []string) string {
	result := ""
	for i, p := range placeholders {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}

// DataScopeFilter 数据权限过滤器（GORM Scope 函数）
func DataScopeFilter(user *auth.UserDetails, config DataPermissionConfig) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return ApplyDataScope(db, user, config)
	}
}
