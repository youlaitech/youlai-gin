package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	pkgRedis "youlai-gin/internal/common/redis"
	"youlai-gin/pkg/constant"
)

var rolePermsKey = constant.RedisKeyRolePerms

// RefreshPermsCacheByCode 刷新单个角色的权限缓存
func (s *Service) RefreshPermsCacheByCode(roleCode string) error {
	ctx := context.Background()

	rolePerms, err := s.repo.PermsByCode(ctx, roleCode)
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
		_ = pkgRedis.Client.Expire(ctx, rolePermsKey, 10*time.Minute).Err()
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
	_ = pkgRedis.Client.Expire(ctx, rolePermsKey, 10*time.Minute).Err()

	log.Printf("刷新角色[%s]权限缓存: %v", roleCode, rolePerms.Perms)
	return nil
}

// RefreshPermsCacheByCodes 批量刷新多个角色的权限缓存
func (s *Service) RefreshPermsCacheByCodes(roleCodes []string) error {
	if len(roleCodes) == 0 {
		return nil
	}

	ctx := context.Background()

	rolePermsList, err := s.repo.PermsByCodes(ctx, roleCodes)
	if err != nil {
		log.Printf("查询角色权限失败: %v", err)
		return err
	}

	_ = pkgRedis.Client.HDel(ctx, rolePermsKey, roleCodes...).Err()

	successCount := 0
	for _, rolePerms := range rolePermsList {
		var permsJSON []byte

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
		_ = pkgRedis.Client.Expire(ctx, rolePermsKey, 10*time.Minute).Err()

		successCount++
	}

	log.Printf("批量刷新角色权限缓存完成: %d/%d 个角色", successCount, len(roleCodes))
	return nil
}

// RefreshPermsCacheByMenus 批量刷新受菜单变更影响的角色权限缓存（供菜单模块调用）
func (s *Service) RefreshPermsCacheByMenus(menuIds []int64) error {
	if len(menuIds) == 0 {
		return nil
	}

	roleCodes, err := s.repo.CodesByMenuIds(context.Background(), menuIds)
	if err != nil {
		return fmt.Errorf("查询受影响的角色失败: %w", err)
	}
	if len(roleCodes) == 0 {
		return nil
	}

	if err := s.RefreshPermsCacheByCodes(roleCodes); err != nil {
		return fmt.Errorf("批量刷新角色权限缓存失败: %w", err)
	}

	return nil
}
