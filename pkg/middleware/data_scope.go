package middleware

import (
	"fmt"

	"gorm.io/gorm"

	"youlai-gin/internal/system/permission/service"
	"youlai-gin/pkg/auth"
)

// 数据权限范围常量（与 youlai-boot 对齐）
const (
	DataScopeAll             = 1 // 全部数据
	DataScopeDeptAndChildren = 2 // 部门及子部门
	DataScopeDept            = 3 // 本部门
	DataScopeSelf            = 4 // 仅本人
	DataScopeCustom          = 5 // 自定义
)

// DataPermissionConfig 数据权限配置
type DataPermissionConfig struct {
	DeptAlias      string // 部门表别名
	DeptIDColumn   string // 部门ID字段名，默认 "dept_id"
	UserAlias      string // 用户表别名
	UserIDColumn   string // 用户ID字段名，默认 "create_by"
}

// ApplyDataScope 应用数据权限过滤（多角色并集策略）
func ApplyDataScope(db *gorm.DB, user *auth.UserDetails, config DataPermissionConfig) *gorm.DB {
	if user == nil {
		return db.Where("1 = 0")
	}

	// ROOT 角色跳过
	for _, role := range user.Roles {
		if role == "ROOT" {
			return db
		}
	}

	// 获取用户数据权限
	dataScopes, err := service.GetUserDataScopes(user.UserID, user.Roles, int64(user.DeptID))
	if err != nil {
		return db.Where("1 = 0")
	}

	// 任一角色有全部数据权限
	if service.HasAllDataScope(dataScopes) {
		return db
	}

	if len(dataScopes) == 0 {
		return db.Where("1 = 0")
	}

	// 设置默认值
	deptColumn := config.DeptIDColumn
	if deptColumn == "" {
		deptColumn = "dept_id"
	}
	userColumn := config.UserIDColumn
	if userColumn == "" {
		userColumn = "create_by"
	}

	// 添加表别名前缀
	if config.DeptAlias != "" {
		deptColumn = config.DeptAlias + "." + deptColumn
	}
	if config.UserAlias != "" {
		userColumn = config.UserAlias + "." + userColumn
	}

	// 构建多角色并集条件
	return buildUnionCondition(db, dataScopes, deptColumn, userColumn, user.UserID)
}

// buildUnionCondition 构建多角色并集查询条件
func buildUnionCondition(db *gorm.DB, dataScopes []auth.RoleDataScope, deptColumn, userColumn string, userID int64) *gorm.DB {
	// 收集各角色的条件
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

	// 使用 OR 连接所有条件
	if len(orConditions) == 1 {
		return db.Where(orConditions[0], args...)
	}

	// 构建并集条件
	unionCond := "(" + orConditions[0] + ")"
	for i := 1; i < len(orConditions); i++ {
		unionCond += " OR (" + orConditions[i] + ")"
	}

	return db.Where(unionCond, args...)
}

// buildRoleCondition 构建单角色数据权限条件
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

// joinPlaceholders 连接占位符
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

// HasDataScopeAll 判断用户是否有全部数据权限
func HasDataScopeAll(userID int64) bool {
	dataScopes, err := service.GetUserDataScopes(userID, nil, 0)
	if err != nil {
		return false
	}
	return service.HasAllDataScope(dataScopes)
}

// GetViewableUserIDs 获取用户可查看的用户ID范围
// 返回 "all" 或用户ID列表
func GetViewableUserIDs(user *auth.UserDetails) (string, []int64) {
	if user == nil {
		return "", nil
	}

	// ROOT 角色返回全部
	for _, role := range user.Roles {
		if role == "ROOT" {
			return "all", nil
		}
	}

	dataScopes, err := service.GetUserDataScopes(user.UserID, user.Roles, int64(user.DeptID))
	if err != nil {
		return "", []int64{user.UserID}
	}

	if service.HasAllDataScope(dataScopes) {
		return "all", nil
	}

	if len(dataScopes) == 0 {
		return "", []int64{user.UserID}
	}

	// 收集所有部门ID
	deptIDSet := make(map[int64]struct{})
	hasSelf := false

	for _, ds := range dataScopes {
		switch ds.DataScope {
		case DataScopeAll:
			return "all", nil
		case DataScopeDeptAndChildren, DataScopeDept, DataScopeCustom:
			for _, id := range ds.CustomDeptIDs {
				deptIDSet[id] = struct{}{}
			}
		case DataScopeSelf:
			hasSelf = true
		}
	}

	// 如果有部门权限，查询这些部门的用户
	if len(deptIDSet) > 0 {
		deptIDs := make([]int64, 0, len(deptIDSet))
		for id := range deptIDSet {
			deptIDs = append(deptIDs, id)
		}
		return "", deptIDs
	}

	// 只有本人权限
	if hasSelf {
		return "", []int64{user.UserID}
	}

	return "", []int64{user.UserID}
}

// HasPermissionToUser 判断当前用户是否有权限操作目标用户
func HasPermissionToUser(currentUser *auth.UserDetails, targetUserID int64, db *gorm.DB) bool {
	if currentUser == nil {
		return false
	}

	// ROOT 角色有全部权限
	for _, role := range currentUser.Roles {
		if role == "ROOT" {
			return true
		}
	}

	// 操作自己
	if currentUser.UserID == targetUserID {
		return true
	}

	dataScopes, err := service.GetUserDataScopes(currentUser.UserID, currentUser.Roles, int64(currentUser.DeptID))
	if err != nil {
		return false
	}

	if service.HasAllDataScope(dataScopes) {
		return true
	}

	// 获取目标用户的部门
	var targetDeptID int64
	err = db.Table("sys_user").
		Select("dept_id").
		Where("id = ? AND is_deleted = 0", targetUserID).
		Scan(&targetDeptID).Error
	if err != nil {
		return false
	}

	// 检查目标用户是否在权限范围内
	for _, ds := range dataScopes {
		switch ds.DataScope {
		case DataScopeAll:
			return true
		case DataScopeDeptAndChildren, DataScopeDept, DataScopeCustom:
			for _, id := range ds.CustomDeptIDs {
				if id == targetDeptID {
					return true
				}
			}
		}
	}

	return false
}
