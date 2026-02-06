package middleware

import (
	"fmt"

	"gorm.io/gorm"

	"youlai-gin/internal/system/permission/model"
	"youlai-gin/internal/system/permission/service"
	"youlai-gin/pkg/auth"
)

// 数据权限范围常量
const (
	DataScopeAll              = 1 // 全部数据权限
	DataScopeDeptAndChildren  = 2 // 本部门及以下数据权限
	DataScopeDept             = 3 // 本部门数据权限
	DataScopeSelf             = 4 // 仅本人数据权限
	DataScopeCustom           = 5 // 自定义数据权限
)

// ApplyDataScope 应用数据权限过滤
func ApplyDataScope(db *gorm.DB, user *auth.UserDetails, tableName, deptIDColumn, userIDColumn string) *gorm.DB {
	if user == nil {
		// 未登录用户，不返回任何数据
		return db.Where("1 = 0")
	}

	// 获取用户权限信息
	userPerms, err := service.GetUserPermissions(user.UserID)
	if err != nil {
		// 获取权限失败，不返回任何数据
		return db.Where("1 = 0")
	}

	// 超级管理员（ROOT），返回所有数据
	for _, role := range userPerms.Roles {
		if role == "ROOT" {
			return db
		}
	}

	// 根据数据权限范围过滤
	switch userPerms.DataScope {
	case DataScopeAll:
		// 全部数据权限，不需要过滤
		return db

	case DataScopeDeptAndChildren:
		// 本部门及以下数据权限
		if len(userPerms.DeptIds) > 0 {
			return db.Where(fmt.Sprintf("%s.%s IN (?)", tableName, deptIDColumn), userPerms.DeptIds)
		}
		// 如果没有部门信息，降级为仅本人
		return db.Where(fmt.Sprintf("%s.%s = ?", tableName, userIDColumn), user.UserID)

	case DataScopeDept:
		// 本部门数据权限
		if userPerms.DeptID > 0 {
			return db.Where(fmt.Sprintf("%s.%s = ?", tableName, deptIDColumn), userPerms.DeptID)
		}
		// 如果没有部门信息，降级为仅本人
		return db.Where(fmt.Sprintf("%s.%s = ?", tableName, userIDColumn), user.UserID)

	case DataScopeSelf:
		// 仅本人数据权限
		return db.Where(fmt.Sprintf("%s.%s = ?", tableName, userIDColumn), user.UserID)

	case DataScopeCustom:
		// 自定义数据权限（由角色配置的部门列表）
		if len(userPerms.DeptIds) > 0 {
			return db.Where(fmt.Sprintf("%s.%s IN (?)", tableName, deptIDColumn), userPerms.DeptIds)
		}
		// 如果没有自定义部门，不返回任何数据
		return db.Where("1 = 0")

	default:
		// 未知数据权限范围，不返回任何数据
		return db.Where("1 = 0")
	}
}

// DataScopeFilter 数据权限过滤器
func DataScopeFilter(user *auth.UserDetails, tableName, deptIDColumn, userIDColumn string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return ApplyDataScope(db, user, tableName, deptIDColumn, userIDColumn)
	}
}

// GetUserDataScope 获取用户数据权限信息
func GetUserDataScope(userID int64) (*model.UserPermissionsVO, error) {
	return service.GetUserPermissions(userID)
}

// HasDataScopeAll 判断用户是否有全部数据权限
func HasDataScopeAll(userID int64) bool {
	userPerms, err := service.GetUserPermissions(userID)
	if err != nil {
		return false
	}

	// 超级管理员（ROOT）
	for _, role := range userPerms.Roles {
		if role == "ROOT" {
			return true
		}
	}

	return userPerms.DataScope == DataScopeAll
}
