package service

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"youlai-gin/internal/database"
	deptModel "youlai-gin/internal/system/dept/model"
	permModel "youlai-gin/internal/system/permission/model"
	userRepo "youlai-gin/internal/system/user/repository"
	"youlai-gin/pkg/errs"
	"youlai-gin/pkg/redis"
)

const rolePermsKey = "system:role:perms"

// GetUserPermissions 获取用户权限信息（角色、按钮权限、数据权限范围等）
func GetUserPermissions(userID int64) (*permModel.UserPermissionsVO, error) {
	if userID <= 0 {
		return nil, errs.BadRequest("用户ID不能为空")
	}

	roles, err := userRepo.GetUserRoles(userID)
	if err != nil {
		return nil, errs.SystemError("查询用户角色失败")
	}

	user, err := userRepo.GetUserByID(userID)
	if err != nil {
		return nil, errs.SystemError("查询用户信息失败")
	}

	perms, err := getUserPermsByRoles(roles)
	if err != nil {
		return nil, err
	}

	dataScope, deptIds, err := resolveUserDataScope(userID, roles, user.DeptID)
	if err != nil {
		return nil, err
	}

	return &permModel.UserPermissionsVO{
		UserID:    userID,
		Roles:     roles,
		Perms:     perms,
		DataScope: dataScope,
		DeptID:    user.DeptID,
		DeptIds:   deptIds,
	}, nil
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
			// 缓存异常也视为缺失，走降级
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
		rolePermsList, err := userRepo.GetRolePermsByCodes(missingRoles)
		if err == nil {
			for _, rp := range rolePermsList {
				for _, p := range rp.Perms {
					p = strings.TrimSpace(p)
					if p != "" {
						permsSet[p] = struct{}{}
					}
				}
			}
		}
		// DB失败时不阻断，只返回已从缓存拿到的
	}

	perms := make([]string, 0, len(permsSet))
	for p := range permsSet {
		perms = append(perms, p)
	}
	return perms, nil
}

func resolveUserDataScope(userID int64, roleCodes []string, deptID int64) (int, []int64, error) {
	// 1) ROOT 超级管理员：全部数据
	for _, r := range roleCodes {
		if r == "ROOT" {
			return 1, nil, nil
		}
	}

	// 2) 取用户角色对应的 data_scope（最宽的数据权限）
	// 约定：1 全部 > 2 本部门及以下 > 3 本部门 > 4 仅本人；5 自定义（若没有配置表则降级为本部门）
	var scopes []int
	err := database.DB.Table("sys_role r").
		Select("DISTINCT r.data_scope").
		Joins("INNER JOIN sys_user_role ur ON ur.role_id = r.id").
		Where("ur.user_id = ? AND r.status = 1 AND r.is_deleted = 0", userID).
		Pluck("r.data_scope", &scopes).Error
	if err != nil {
		return 0, nil, errs.SystemError("查询数据权限范围失败")
	}

	best := 4 // 默认仅本人
	custom := false
	for _, s := range scopes {
		switch s {
		case 1:
			best = 1
		case 2:
			if best != 1 {
				best = 2
			}
		case 3:
			if best != 1 && best != 2 {
				best = 3
			}
		case 5:
			custom = true
		default:
			// ignore
		}
	}

	if custom {
		// 项目SQL里没发现 sys_role_dept；自定义数据权限无法解析时，降级为本部门
		best = 3
	}

	if best == 1 {
		return 1, nil, nil
	}

	// 本部门及以下：取 tree_path 前缀匹配
	if best == 2 {
		if deptID <= 0 {
			return 4, nil, nil
		}
		var treePath string
		err := database.DB.Table(deptModel.Dept{}.TableName()).
			Select("tree_path").
			Where("id = ? AND is_deleted = 0", deptID).
			Scan(&treePath).Error
		if err != nil {
			return 4, nil, nil
		}
		pattern := treePath + "," + strconv.FormatInt(deptID, 10) + "%"
		var deptIds []int64
		err = database.DB.Table(deptModel.Dept{}.TableName()).
			Select("id").
			Where("is_deleted = 0 AND (id = ? OR tree_path LIKE ?)", deptID, pattern).
			Pluck("id", &deptIds).Error
		if err != nil {
			return 4, nil, nil
		}
		return 2, deptIds, nil
	}

	if best == 3 {
		return 3, nil, nil
	}

	return 4, nil, nil
}
