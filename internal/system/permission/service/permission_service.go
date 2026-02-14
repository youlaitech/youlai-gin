package service

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	deptModel "youlai-gin/internal/system/dept/model"
	permModel "youlai-gin/internal/system/permission/model"
	"youlai-gin/pkg/auth"
	"youlai-gin/pkg/database"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/redis"
	"youlai-gin/pkg/types"
)

func getUserRoleCodes(userID int64) ([]string, error) {
	var roleCodes []string
	err := database.DB.Table("sys_user_role ur").
		Select("r.code").
		Joins("INNER JOIN sys_role r ON ur.role_id = r.id").
		Where("ur.user_id = ? AND r.is_deleted = 0 AND r.status = 1", userID).
		Pluck("r.code", &roleCodes).Error
	return roleCodes, err
}

func getUserDeptID(userID int64) (int64, error) {
	var deptID int64
	err := database.DB.Table("sys_user").
		Select("COALESCE(dept_id, 0)").
		Where("id = ? AND is_deleted = 0", userID).
		Scan(&deptID).Error
	return deptID, err
}

const rolePermsKey = "system:role:perms"

// 数据权限范围常量
const (
	DataScopeAll             = 1 // 全部数据
	DataScopeDeptAndChildren = 2 // 部门及子部门
	DataScopeDept            = 3 // 本部门
	DataScopeSelf            = 4 // 仅本人
	DataScopeCustom          = 5 // 自定义
)

// GetUserPermissions 获取用户权限信息
func GetUserPermissions(userID int64) (*permModel.UserPermissionsVO, error) {
	if userID <= 0 {
		return nil, errs.BadRequest("用户ID不能为空")
	}

	roles, err := getUserRoleCodes(userID)
	if err != nil {
		return nil, errs.SystemError("查询用户角色失败")
	}

	deptID, err := getUserDeptID(userID)
	if err != nil {
		return nil, errs.SystemError("查询用户部门失败")
	}

	perms, err := getUserPermsByRoles(roles)
	if err != nil {
		return nil, err
	}

	dataScopes, err := GetUserDataScopes(userID, roles, deptID)
	if err != nil {
		return nil, err
	}

	return &permModel.UserPermissionsVO{
		UserID:     types.BigInt(userID),
		Roles:      roles,
		Perms:      perms,
		DataScopes: dataScopes,
		DeptID:     types.BigInt(deptID),
	}, nil
}

type rolePermsRow struct {
	RoleCode string
	Perm     string
}

func getRolePermsByCodes(roleCodes []string) ([]rolePermsRow, error) {
	if len(roleCodes) == 0 {
		return nil, nil
	}
	var rows []rolePermsRow
	err := database.DB.Table("sys_role r").
		Select("r.code as role_code, p.perm").
		Joins("INNER JOIN sys_role_menu rm ON rm.role_id = r.id").
		Joins("INNER JOIN sys_menu p ON p.id = rm.menu_id").
		Where("r.code IN ? AND r.status = 1 AND r.is_deleted = 0 AND p.is_deleted = 0 AND p.status = 1", roleCodes).
		Scan(&rows).Error
	return rows, err
}

// GetUserDataScopes 获取用户所有角色的数据权限列表（多角色并集策略）
func GetUserDataScopes(userID int64, roleCodes []string, deptID int64) ([]auth.RoleDataScope, error) {
	// ROOT 角色直接返回全部权限
	for _, r := range roleCodes {
		if r == "ROOT" {
			return []auth.RoleDataScope{auth.NewRoleDataScopeAll("ROOT")}, nil
		}
	}

	// 查询用户角色的数据权限
	type roleDataScopeRow struct {
		Code      string
		DataScope int
	}
	var rows []roleDataScopeRow
	err := database.DB.Table("sys_role r").
		Select("r.code, r.data_scope").
		Joins("INNER JOIN sys_user_role ur ON ur.role_id = r.id").
		Where("ur.user_id = ? AND r.status = 1 AND r.is_deleted = 0", userID).
		Scan(&rows).Error
	if err != nil {
		return nil, errs.SystemError("查询角色数据权限失败")
	}

	if len(rows) == 0 {
		return []auth.RoleDataScope{auth.NewRoleDataScopeSelf("DEFAULT")}, nil
	}

	// 收集需要查询自定义部门的角色
	customRoleCodes := make([]string, 0)
	for _, row := range rows {
		if row.DataScope == DataScopeCustom {
			customRoleCodes = append(customRoleCodes, row.Code)
		}
	}

	// 批量查询自定义部门
	customDeptMap := make(map[string][]int64)
	if len(customRoleCodes) > 0 {
		type roleDeptRow struct {
			RoleCode string
			DeptID   int64
		}
		var deptRows []roleDeptRow
		err := database.DB.Table("sys_role_dept rd").
			Select("r.code as role_code, rd.dept_id").
			Joins("INNER JOIN sys_role r ON r.id = rd.role_id").
			Where("r.code IN ? AND r.status = 1 AND r.is_deleted = 0", customRoleCodes).
			Scan(&deptRows).Error
		if err == nil {
			for _, dr := range deptRows {
				customDeptMap[dr.RoleCode] = append(customDeptMap[dr.RoleCode], dr.DeptID)
			}
		}
	}

	// 构建 RoleDataScope 列表
	result := make([]auth.RoleDataScope, 0, len(rows))
	for _, row := range rows {
		switch row.DataScope {
		case DataScopeAll:
			result = append(result, auth.NewRoleDataScopeAll(row.Code))
		case DataScopeDeptAndChildren:
			deptIDs := getDeptAndChildrenIDs(deptID)
			if len(deptIDs) > 0 {
				result = append(result, auth.RoleDataScope{
					RoleCode:      row.Code,
					DataScope:     DataScopeDeptAndChildren,
					CustomDeptIDs: deptIDs,
				})
			}
		case DataScopeDept:
			if deptID > 0 {
				result = append(result, auth.RoleDataScope{
					RoleCode:      row.Code,
					DataScope:     DataScopeDept,
					CustomDeptIDs: []int64{deptID},
				})
			}
		case DataScopeSelf:
			result = append(result, auth.NewRoleDataScopeSelf(row.Code))
		case DataScopeCustom:
			if deptIDs, ok := customDeptMap[row.Code]; ok && len(deptIDs) > 0 {
				result = append(result, auth.NewRoleDataScopeCustom(row.Code, deptIDs))
			}
		}
	}

	if len(result) == 0 {
		return []auth.RoleDataScope{auth.NewRoleDataScopeSelf("DEFAULT")}, nil
	}

	return result, nil
}

// getDeptAndChildrenIDs 获取部门及子部门ID列表
func getDeptAndChildrenIDs(deptID int64) []int64 {
	if deptID <= 0 {
		return nil
	}

	// 获取部门的 tree_path
	var treePath string
	err := database.DB.Table(deptModel.Dept{}.TableName()).
		Select("tree_path").
		Where("id = ? AND is_deleted = 0", deptID).
		Scan(&treePath).Error
	if err != nil {
		return []int64{deptID}
	}

	// 查询部门及子部门
	pattern := treePath + "," + strconv.FormatInt(deptID, 10) + "%"
	var deptIDs []int64
	err = database.DB.Table(deptModel.Dept{}.TableName()).
		Select("id").
		Where("is_deleted = 0 AND (id = ? OR tree_path LIKE ?)", deptID, pattern).
		Pluck("id", &deptIDs).Error
	if err != nil {
		return []int64{deptID}
	}

	return deptIDs
}

// HasAllDataScope 判断是否有全部数据权限
func HasAllDataScope(dataScopes []auth.RoleDataScope) bool {
	for _, ds := range dataScopes {
		if ds.DataScope == DataScopeAll {
			return true
		}
	}
	return false
}

// CheckPermission 检查用户是否拥有指定权限
func CheckPermission(userID int64, perm string) (bool, error) {
	if strings.TrimSpace(perm) == "" {
		return false, nil
	}
	perms, err := GetUserPermissions(userID)
	if err != nil {
		return false, err
	}
	for _, p := range perms.Perms {
		if p == perm {
			return true, nil
		}
	}
	return false, nil
}

// CheckAnyPermission 检查是否拥有任意一个权限
func CheckAnyPermission(userID int64, perms []string) (bool, error) {
	if len(perms) == 0 {
		return true, nil
	}
	userPerms, err := GetUserPermissions(userID)
	if err != nil {
		return false, err
	}
	set := make(map[string]struct{}, len(userPerms.Perms))
	for _, p := range userPerms.Perms {
		set[p] = struct{}{}
	}
	for _, p := range perms {
		if _, ok := set[p]; ok {
			return true, nil
		}
	}
	return false, nil
}

// CheckAllPermissions 检查是否拥有所有权限
func CheckAllPermissions(userID int64, perms []string) (bool, error) {
	if len(perms) == 0 {
		return true, nil
	}
	userPerms, err := GetUserPermissions(userID)
	if err != nil {
		return false, err
	}
	set := make(map[string]struct{}, len(userPerms.Perms))
	for _, p := range userPerms.Perms {
		set[p] = struct{}{}
	}
	for _, p := range perms {
		if _, ok := set[p]; !ok {
			return false, nil
		}
	}
	return true, nil
}

// CheckRole 检查是否拥有指定角色
func CheckRole(userID int64, roleCode string) (bool, error) {
	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return false, nil
	}
	userPerms, err := GetUserPermissions(userID)
	if err != nil {
		return false, err
	}
	for _, r := range userPerms.Roles {
		if r == roleCode {
			return true, nil
		}
	}
	return false, nil
}

// CheckAnyRole 检查是否拥有任意一个角色
func CheckAnyRole(userID int64, roleCodes []string) (bool, error) {
	if len(roleCodes) == 0 {
		return true, nil
	}
	userPerms, err := GetUserPermissions(userID)
	if err != nil {
		return false, err
	}
	set := make(map[string]struct{}, len(userPerms.Roles))
	for _, r := range userPerms.Roles {
		set[r] = struct{}{}
	}
	for _, r := range roleCodes {
		if _, ok := set[r]; ok {
			return true, nil
		}
	}
	return false, nil
}

func getUserPermsByRoles(roleCodes []string) ([]string, error) {
	if len(roleCodes) == 0 {
		return []string{}, nil
	}

	ctx := context.Background()
	permsSet := make(map[string]struct{})
	missingRoles := make([]string, 0)

	for _, roleCode := range roleCodes {
		val, err := redis.Client.HGet(ctx, rolePermsKey, roleCode).Result()
		if err != nil {
			missingRoles = append(missingRoles, roleCode)
			continue
		}
		if val == "" {
			continue
		}

		var rolePerms []string
		if err := json.Unmarshal([]byte(val), &rolePerms); err != nil {
			missingRoles = append(missingRoles, roleCode)
			continue
		}
		for _, p := range rolePerms {
			p = strings.TrimSpace(p)
			if p != "" {
				permsSet[p] = struct{}{}
			}
		}
	}

	// 降级：从DB查询缺失角色的权限
	if len(missingRoles) > 0 {
		rolePermsList, err := getRolePermsByCodes(missingRoles)
		if err == nil {
			for _, rp := range rolePermsList {
				p := strings.TrimSpace(rp.Perm)
				if p != "" {
					permsSet[p] = struct{}{}
				}
			}
		}
	}

	perms := make([]string, 0, len(permsSet))
	for p := range permsSet {
		perms = append(perms, p)
	}
	return perms, nil
}
