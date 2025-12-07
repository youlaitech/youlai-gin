package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"youlai-gin/internal/system/role/repository"
	pkgRedis "youlai-gin/pkg/redis"
)

const (
	rolePermsKey     = "system:role:perms"
	rolePermsCacheTTL = 24 * time.Hour // 缓存24小时过期
)

// InitRolePermsCache 初始化角色权限缓存
func InitRolePermsCache() error {
	log.Println("初始化角色权限缓存...")
	return RefreshRolePermsCache()
}

// RefreshRolePermsCache 刷新所有角色权限缓存
func RefreshRolePermsCache() error {
	ctx := context.Background()
	
	// 1. 从数据库查询所有角色权限
	rolePermsList, err := repository.GetAllRolePerms()
	if err != nil {
		log.Printf("❌ 查询角色权限失败: %v", err)
		return err
	}
	
	if len(rolePermsList) == 0 {
		log.Println("⚠️  没有找到角色权限数据")
		return nil
	}
	
	// 2. 清理旧缓存（在查询成功后再删除，避免删除后查询失败导致缓存丢失）
	if err := pkgRedis.Client.Del(ctx, rolePermsKey).Err(); err != nil {
		log.Printf("⚠️  清理角色权限缓存失败: %v", err)
		// 继续执行，不返回错误
	}
	
	// 3. 批量写入Redis缓存
	successCount := 0
	for _, rolePerms := range rolePermsList {
		if len(rolePerms.Perms) == 0 {
			// 空权限也要缓存，避免穿透
			if err := pkgRedis.Client.HSet(ctx, rolePermsKey, rolePerms.RoleCode, "[]").Err(); err != nil {
				log.Printf("⚠️  缓存角色[%s]空权限失败: %v", rolePerms.RoleCode, err)
				continue
			}
			successCount++
			continue
		}
		
		// 将权限列表序列化为JSON
		permsJSON, err := json.Marshal(rolePerms.Perms)
		if err != nil {
			log.Printf("❌ 序列化角色[%s]权限失败: %v", rolePerms.RoleCode, err)
			continue
		}
		
		// 存储到Redis Hash
		if err := pkgRedis.Client.HSet(ctx, rolePermsKey, rolePerms.RoleCode, string(permsJSON)).Err(); err != nil {
			log.Printf("❌ 缓存角色[%s]权限失败: %v", rolePerms.RoleCode, err)
			continue
		}
		
		log.Printf("✅ 缓存角色[%s]权限: %v", rolePerms.RoleCode, rolePerms.Perms)
		successCount++
	}
	
	// 4. 设置缓存过期时间（防止永久缓存导致一致性问题）
	if err := pkgRedis.Client.Expire(ctx, rolePermsKey, rolePermsCacheTTL).Err(); err != nil {
		log.Printf("⚠️  设置角色权限缓存过期时间失败: %v", err)
	}
	
	log.Printf("✅ 角色权限缓存刷新完成，共缓存 %d/%d 个角色", successCount, len(rolePermsList))
	return nil
}

// RefreshRolePermsCacheByCode 刷新指定角色的权限缓存
func RefreshRolePermsCacheByCode(roleCode string) error {
	ctx := context.Background()
	
	// 1. 查询该角色的权限
	rolePerms, err := repository.GetRolePermsByCode(roleCode)
	if err != nil {
		log.Printf("❌ 查询角色[%s]权限失败: %v", roleCode, err)
		return err
	}
	
	// 2. 删除旧缓存
	if err := pkgRedis.Client.HDel(ctx, rolePermsKey, roleCode).Err(); err != nil {
		log.Printf("⚠️  删除角色[%s]权限缓存失败: %v", roleCode, err)
	}
	
	// 3. 更新缓存
	if len(rolePerms.Perms) == 0 {
		// 空权限也要缓存，避免穿透
		if err := pkgRedis.Client.HSet(ctx, rolePermsKey, roleCode, "[]").Err(); err != nil {
			log.Printf("❌ 缓存角色[%s]空权限失败: %v", roleCode, err)
			return err
		}
		log.Printf("✅ 刷新角色[%s]权限缓存: []", roleCode)
		return nil
	}
	
	// 将权限列表序列化为JSON
	permsJSON, err := json.Marshal(rolePerms.Perms)
	if err != nil {
		log.Printf("❌ 序列化角色[%s]权限失败: %v", roleCode, err)
		return err
	}
	
	// 存储到Redis Hash
	if err := pkgRedis.Client.HSet(ctx, rolePermsKey, roleCode, string(permsJSON)).Err(); err != nil {
		log.Printf("❌ 缓存角色[%s]权限失败: %v", roleCode, err)
		return err
	}
	
	log.Printf("✅ 刷新角色[%s]权限缓存: %v", roleCode, rolePerms.Perms)
	return nil
}

// RefreshRolePermsCacheByCodes 批量刷新多个角色的权限缓存
func RefreshRolePermsCacheByCodes(roleCodes []string) error {
	if len(roleCodes) == 0 {
		return nil
	}
	
	ctx := context.Background()
	
	// 1. 查询这些角色的权限
	rolePermsList, err := repository.GetRolePermsByCodes(roleCodes)
	if err != nil {
		log.Printf("❌ 查询角色权限失败: %v", err)
		return err
	}
	
	// 2. 批量删除旧缓存
	if len(roleCodes) > 0 {
		if err := pkgRedis.Client.HDel(ctx, rolePermsKey, roleCodes...).Err(); err != nil {
			log.Printf("⚠️  批量删除角色权限缓存失败: %v", err)
		}
	}
	
	// 3. 批量更新缓存
	successCount := 0
	for _, rolePerms := range rolePermsList {
		var permsJSON []byte
		var err error
		
		if len(rolePerms.Perms) == 0 {
			permsJSON = []byte("[]")
		} else {
			permsJSON, err = json.Marshal(rolePerms.Perms)
			if err != nil {
				log.Printf("❌ 序列化角色[%s]权限失败: %v", rolePerms.RoleCode, err)
				continue
			}
		}
		
		// 存储到Redis Hash
		if err := pkgRedis.Client.HSet(ctx, rolePermsKey, rolePerms.RoleCode, string(permsJSON)).Err(); err != nil {
			log.Printf("❌ 缓存角色[%s]权限失败: %v", rolePerms.RoleCode, err)
			continue
		}
		
		successCount++
	}
	
	log.Printf("✅ 批量刷新角色权限缓存完成: %d/%d 个角色", successCount, len(roleCodes))
	return nil
}
