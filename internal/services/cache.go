package services

import (
	"context"
	"fmt"
	"gin-admin/internal/model/rbac"
	"gin-admin/pkg/cache"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"

	"golang.org/x/sync/singleflight"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/4 下午4:25
* @Package: 缓存服务 - 统一管理所有缓存操作
 */

// ICacheService 缓存服务接口
type ICacheService interface {
	// 权限相关缓存
	CheckUserPermission(ctx context.Context, userID uint, path, method string, fn func(ctx context.Context, uid uint) ([]rbac.Resource, error)) (bool, error)
	ClearUserPermissions(ctx context.Context, userID uint, ttl time.Duration, updateFn func() error) error
	ClearMultipleUsersPermissions(ctx context.Context, userIDs []uint, ttl time.Duration, updateFn func() error) error
	SetUserPermissions(ctx context.Context, userID uint, resources []rbac.Resource) error
	ClearAllPermissions(ctx context.Context) error
	// Token黑名单
	BlacklistToken(ctx context.Context, token string, ttl time.Duration) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)

	// 通用缓存操作
	GetInstance(ctx context.Context, key string, dest interface{}) error
	SetInstance(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	DeleteInstance(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// cacheService 缓存服务实现
type cacheService struct {
	client cache.ICache
	sg     singleflight.Group // 防止缓存击穿（多个请求同时查询同一个不存在的key）
}

func NewCacheService(cache cache.ICache) ICacheService {
	return &cacheService{
		client: cache,
	}
}

// ================================
// 缓存Key管理
// ================================

const (
	// 缓存Key前缀
	cacheKeyPermissionPrefix = "permission:"       // 权限: permission:userID (使用Set存储 path_method)
	cacheKeyToken            = "token:%s"          // Token黑名单: token:tokenString
	cacheKeyJWTBlacklist     = "jwt:blacklist:%s"  // JWT黑名单: jwt:blacklist:token
	cacheKeyUserSessions     = "user:sessions:%d"  // 用户会话: user:sessions:userID -> Set[sessionID]
	cacheKeySessionTokens    = "session:tokens:%s" // 会话令牌: session:tokens:sessionID -> {access, refresh}
	cacheKeyRefreshCount     = "refresh:count:%s"  // 刷新计数: refresh:count:refreshToken -> count
	cacheKeyEmptyMarker      = "empty:%s"          // 空值标记（防止缓存穿透）

	// 缓存TTL
	ttlPermission       = 10 * time.Minute   // 权限缓存10分钟
	ttlPermissionOffset = 2 * time.Minute    // 权限缓存随机偏移（防止雪崩）
	ttlToken            = 24 * time.Hour     // Token 24小时
	ttlSession          = 7 * 24 * time.Hour // 会话 7天
	ttlEmpty            = 5 * time.Minute    // 空值缓存5分钟（防止穿透）
)

// ================================
// 权限相关缓存
// ================================

// CheckUserPermission 检查用户权限（带缓存 + 防穿透 + 防击穿）
// 使用Redis Set存储用户的所有权限资源，格式：permission:userID -> Set["GET_/api/v1/users", "POST_/api/v1/posts"]
// 优化措施：
// 1. 防穿透：缓存空权限（用户没有任何权限时也缓存）
// 2. 防击穿：使用 singleflight 确保同一个 key 只有一个请求去查询数据库
// 3. 防雪崩：TTL 添加随机偏移
func (s *cacheService) CheckUserPermission(ctx context.Context, userID uint, path, method string, fn func(ctx context.Context, uid uint) ([]rbac.Resource, error)) (bool, error) {
	if s.client == nil {
		// 缓存不可用
		return false, cache.ErrUnreachable
	}

	cacheKey := fmt.Sprintf("%s%d", cacheKeyPermissionPrefix, userID)
	member := fmt.Sprintf("%s_%s", method, path)

	// 查询缓存key是否存在
	exists, err := s.client.Exists(ctx, cacheKey)
	if err != nil {
		// 缓存查询失败
		return false, err
	}
	// 缓存命中
	if exists {
		logrus.Info("success check user permission from cache")
		return s.client.SIsMember(ctx, cacheKey, member)
	}

	// 缓存未命中：使用 singleflight 防止缓存击穿
	// 同一时刻只有一个请求去查询数据库并设置缓存
	sfKey := fmt.Sprintf("load_permission:%d", userID)
	_, err, _ = s.sg.Do(sfKey, func() (interface{}, error) {
		// Double Check
		exists, err = s.client.Exists(ctx, cacheKey)
		if err == nil && exists {
			return nil, nil // 缓存已存在，直接返回
		}
		// 从数据库加载用户所有权限
		resources, err := fn(ctx, userID)
		if err != nil {
			return nil, err
		}
		logrus.Info("success find resources from db")

		// 设置缓存
		return nil, s.SetUserPermissions(ctx, userID, resources)
	})
	if err != nil {
		// 加载失败
		return false, err
	}
	// check
	return s.client.SIsMember(ctx, cacheKey, member)
}

// SetUserPermissions 设置用户权限缓存
// 应该在用户登录或权限变更后调用
func (s *cacheService) SetUserPermissions(ctx context.Context, userID uint, resources []rbac.Resource) error {
	if s.client == nil {
		return nil
	}
	cacheKey := fmt.Sprintf("%s%d", cacheKeyPermissionPrefix, userID)

	// 先删除旧缓存
	_ = s.client.Delete(ctx, cacheKey)
	// 如果没有权限，设置一个空标记
	if len(resources) == 0 {
		// 使用一个特殊值标记"空权限"
		emptyMarker := "_EMPTY_"
		if err := s.client.SAdd(ctx, cacheKey, emptyMarker); err != nil {
			return err
		}
		// todo: 理论上empty权限标记超时时间较短，这儿后面来加吧
		return s.client.Expire(ctx, cacheKey, s.getPermissionTTL())
	}
	// 添加所有权限到Set
	members := make([]interface{}, 0, len(resources))
	for _, resource := range resources {
		member := fmt.Sprintf("%s_%s", resource.Method, resource.Path)
		members = append(members, member)
	}
	if err := s.client.SAdd(ctx, cacheKey, members...); err != nil {
		return err
	}
	return s.client.Expire(ctx, cacheKey, s.getPermissionTTL())
}

// getPermissionTTL 获取权限缓存TTL（添加随机偏移防止缓存雪崩）
func (s *cacheService) getPermissionTTL() time.Duration {
	// 基础TTL + 随机偏移（0-2分钟）
	offset := time.Duration(rand.Int63n(int64(ttlPermissionOffset)))
	return ttlPermission + offset
}

// ClearUserPermissions 清除指定用户的所有权限缓存
func (s *cacheService) ClearUserPermissions(ctx context.Context, userID uint, ttl time.Duration, updateFn func() error) error {
	if s.client == nil {
		return nil
	}

	key := fmt.Sprintf("%s%d", cacheKeyPermissionPrefix, userID)
	if err := s.client.Delete(ctx, key); err != nil {
		logrus.Printf("[DelayDoubleDelete] first delete failed: key=%s, err=%v", key, err)
		return err
	}
	// 更新数据库
	if err := updateFn(); err != nil {
		logrus.Printf("[DelayDoubleDelete] updateFn failed: key=%s, err=%v", key, err)
		return err
	}
	// 异步延迟删除
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Printf("[DelayDoubleDelete] panic: %v", r)
			}
		}()
		time.Sleep(ttl)
		if err := s.client.Delete(context.Background(), key); err != nil {
			logrus.Printf("[DelayDoubleDelete] second delete failed: key=%s, err=%v", key, err)
		}
	}()
	return nil
}

// ClearMultipleUsersPermissions 批量清除多个用户的权限缓存
func (s *cacheService) ClearMultipleUsersPermissions(ctx context.Context, userIDs []uint, ttl time.Duration, updateFn func() error) error {
	if s.client == nil || len(userIDs) == 0 {
		return nil
	}

	keys := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		keys = append(keys, fmt.Sprintf("%s%d", cacheKeyPermissionPrefix, userID))
	}
	err := s.client.Delete(ctx, keys...)
	if err != nil {
		return err
	}
	if err = updateFn(); err != nil {
		return err
	}
	// 异步延迟删除
	go func(rKeys []string) {
		defer func() {
			if r := recover(); r != nil {
				logrus.Printf("[DelayDoubleDelete] panic: %v", r)
			}
		}()
		time.Sleep(ttl)
		if err := s.client.Delete(context.Background(), keys...); err != nil {
			logrus.Printf("[DelayDoubleDelete] second delete failed:  err=%v", err)
		}
	}(keys)
	return nil
}
func (s *cacheService) ClearAllPermissions(ctx context.Context) error {
	if s.client == nil {
		return nil
	}
	return s.client.DeletePrefix(ctx, cacheKeyPermissionPrefix)
}

// ================================
// Token 相关缓存（JWT黑名单）
// ================================

// BlacklistToken 将token加入黑名单
func (s *cacheService) BlacklistToken(ctx context.Context, token string, ttl time.Duration) error {
	if s.client == nil {
		return nil
	}

	cacheKey := fmt.Sprintf(cacheKeyToken, token)
	return s.client.Set(ctx, cacheKey, true, ttl)
}

// IsTokenBlacklisted 检查token是否在黑名单
func (s *cacheService) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	if s.client == nil {
		return false, nil
	}

	cacheKey := fmt.Sprintf(cacheKeyToken, token)
	return s.client.Exists(ctx, cacheKey)
}

// ================================
// 通用缓存操作（代理到底层Cache）
// ================================

// GetInstance 获取缓存
func (s *cacheService) GetInstance(ctx context.Context, key string, dest interface{}) error {
	if s.client == nil {
		return fmt.Errorf("cache not available")
	}
	return s.client.Get(ctx, key, dest)
}

// SetInstance 设置缓存
func (s *cacheService) SetInstance(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if s.client == nil {
		return nil
	}
	return s.client.Set(ctx, key, value, ttl)
}

// DeleteInstance 删除缓存
func (s *cacheService) DeleteInstance(ctx context.Context, keys ...string) error {
	if s.client == nil {
		return nil
	}
	return s.client.Delete(ctx, keys...)
}

// Exists 检查key是否存在
func (s *cacheService) Exists(ctx context.Context, key string) (bool, error) {
	if s.client == nil {
		return false, nil
	}
	return s.client.Exists(ctx, key)
}
