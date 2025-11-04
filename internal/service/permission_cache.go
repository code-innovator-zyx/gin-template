package service

import (
	"context"
	"fmt"
	"gin-template/internal/model/rbac"
	"gin-template/pkg/cache"
	"github.com/go-redis/redis/v8"
	"time"
)

const (
	permissionCachePrefix = "permission:"
	permissionCacheTTL    = 10 * time.Minute
)

// HasPermissionWithCache 带缓存的权限检查
func HasPermissionWithCache(ctx context.Context, userID uint, path string, method string) (bool, error) {
	// 生成缓存key
	cacheKey := fmt.Sprintf("%s%d:%s:%s", permissionCachePrefix, userID, path, method)

	// 先尝试从缓存获取
	if cache.RedisClient != nil {
		var hasPermission bool
		err := cache.Get(ctx, cacheKey, &hasPermission)
		if err == nil {
			return hasPermission, nil
		}
		// 缓存未命中或出错，继续查询数据库
		if err != redis.Nil {
			// 记录错误但不影响主流程
			// logrus.Warnf("从缓存获取权限失败: %v", err)
		}
	}

	// 查询数据库
	hasPermission, err := rbac.CheckPermission(userID, path, method)
	if err != nil {
		return false, err
	}

	// 写入缓存
	if cache.RedisClient != nil {
		_ = cache.Set(ctx, cacheKey, hasPermission, permissionCacheTTL)
	}

	return hasPermission, nil
}

// ClearUserPermissionCache 清除用户的权限缓存（当角色变更时调用）
func ClearUserPermissionCache(ctx context.Context, userID uint) error {
	if cache.RedisClient == nil {
		return nil
	}

	// 使用通配符删除该用户的所有权限缓存
	pattern := fmt.Sprintf("%s%d:*", permissionCachePrefix, userID)
	iter := cache.RedisClient.Scan(ctx, 0, pattern, 0).Iterator()

	keys := make([]string, 0)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return cache.Delete(ctx, keys...)
	}

	return nil
}

