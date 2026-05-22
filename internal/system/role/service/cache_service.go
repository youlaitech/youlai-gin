package service

import (
	"context"
	"encoding/json"
	"log"

	"youlai-gin/internal/system/role/repository"
	pkgRedis "youlai-gin/internal/common/redis"
	"youlai-gin/pkg/constant"
)

var rolePermsKey = constant.RedisKeyRolePerms

// RefreshRolePermsCacheByCode 刷新单个角色的权限缓存
func RefreshRolePermsCacheByCode(roleCode string) error {
	ctx := context.Background()

	rolePerms, err := repository.GetRolePermsByCode(roleCode)
	if err != nil {
		log.Printf("查询角色[%s]权限失败: %v", roleCode, err)
		return err
	}

	_ = pkgRedis.Client.HDel(ctx, rolePermsKey, roleCode).Err()

	if len(rolePerms.Perms) == 0 {
		if err := pkgRedis.Client.HSet(ctx, rolePermsKey, roleCode, "[]").Err(); err != nil {
			log.Printf("缓存角色[%s]空权限失败: %v", roleCode, err)
			return err
		}
		log.Printf("刷新角色[%s]权限缓存: []", roleCode)
		return nil
	}

	permsJSON, err := json.Marshal(rolePerms.Perms)
	if err != nil {
		log.Printf("序列化角色[%s]权限失败: %v", roleCode, err)
		return err
	}

	if err := pkgRedis.Client.HSet(ctx, rolePermsKey, roleCode, string(permsJSON)).Err(); err != nil {
		log.Printf("缓存角色[%s]权限失败: %v", roleCode, err)
		return err
	}

	log.Printf("刷新角色[%s]权限缓存: %v", roleCode, rolePerms.Perms)
	return nil
}

// RefreshRolePermsCacheByCodes 批量刷新多个角色的权限缓存
func RefreshRolePermsCacheByCodes(roleCodes []string) error {
	if len(roleCodes) == 0 {
		return nil
	}

	ctx := context.Background()

	rolePermsList, err := repository.GetRolePermsByCodes(roleCodes)
	if err != nil {
		log.Printf("查询角色权限失败: %v", err)
		return err
	}

	_ = pkgRedis.Client.HDel(ctx, rolePermsKey, roleCodes...).Err()

	successCount := 0
	for _, rolePerms := range rolePermsList {
		var permsJSON []byte
		var err error

		if len(rolePerms.Perms) == 0 {
			permsJSON = []byte("[]")
		} else {
			permsJSON, err = json.Marshal(rolePerms.Perms)
			if err != nil {
				log.Printf("序列化角色[%s]权限失败: %v", rolePerms.RoleCode, err)
				continue
			}
		}

		if err := pkgRedis.Client.HSet(ctx, rolePermsKey, rolePerms.RoleCode, string(permsJSON)).Err(); err != nil {
			log.Printf("缓存角色[%s]权限失败: %v", rolePerms.RoleCode, err)
			continue
		}

		successCount++
	}

	log.Printf("批量刷新角色权限缓存完成: %d/%d 个角色", successCount, len(roleCodes))
	return nil
}
