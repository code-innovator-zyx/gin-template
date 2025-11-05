package service

import (
	"context"
	"fmt"
	"gin-template/internal/model/rbac"
	"gin-template/pkg/cache"
	"sync"
	"time"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/4 下午4:25
* @Package: 缓存服务 - 统一管理所有缓存操作（单例模式）
 */

// CacheService 缓存服务接口
type CacheService interface {
	// 权限相关缓存
	CheckUserPermission(ctx context.Context, userID uint, path, method string) (bool, error)
	ClearUserPermissions(ctx context.Context, userID uint) error
	SetUserPermissions(ctx context.Context, userID uint, resources []rbac.Resource) error

	// Token黑名单
	BlacklistToken(ctx context.Context, token string, ttl time.Duration) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)

	// JWT Token 管理
	SaveUserSession(ctx context.Context, userID uint, sessionID string, ttl time.Duration) error
	RemoveUserSession(ctx context.Context, userID uint, sessionID string) error
	GetUserSessions(ctx context.Context, userID uint) ([]string, error)
	RevokeAllUserSessions(ctx context.Context, userID uint) error
	SaveSessionTokens(ctx context.Context, sessionID string, accessToken, refreshToken string, ttl time.Duration) error
	GetSessionTokens(ctx context.Context, sessionID string) (accessToken, refreshToken string, err error)
	IncrementRefreshCount(ctx context.Context, refreshToken string, ttl time.Duration) (int, error)
	GetRefreshCount(ctx context.Context, refreshToken string) (int, error)

	// 通用缓存操作
	GetInstance(ctx context.Context, key string, dest interface{}) error
	SetInstance(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	DeleteInstance(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// cacheService 缓存服务实现
type cacheService struct {
	client cache.Cache // 使用接口而不是具体实现
}

var (
	cacheServiceOnce   sync.Once
	globalCacheService CacheService
)

// GetCacheService 获取缓存服务单例（懒加载，线程安全）
func GetCacheService() CacheService {
	cacheServiceOnce.Do(func() {
		globalCacheService = &cacheService{
			client: cache.GetGlobalCache(),
		}
	})
	return globalCacheService
}

// ================================
// 缓存Key管理
// ================================

const (
	// 缓存Key前缀
	cacheKeyPermission    = "permission:%d"     // 权限: permission:userID (使用Set存储 path_method)
	cacheKeyToken         = "token:%s"          // Token黑名单: token:tokenString
	cacheKeyJWTBlacklist  = "jwt:blacklist:%s"  // JWT黑名单: jwt:blacklist:token
	cacheKeyUserSessions  = "user:sessions:%d"  // 用户会话: user:sessions:userID -> Set[sessionID]
	cacheKeySessionTokens = "session:tokens:%s" // 会话令牌: session:tokens:sessionID -> {access, refresh}
	cacheKeyRefreshCount  = "refresh:count:%s"  // 刷新计数: refresh:count:refreshToken -> count

	// 缓存TTL
	ttlPermission = 10 * time.Minute   // 权限缓存10分钟
	ttlToken      = 24 * time.Hour     // Token 24小时
	ttlSession    = 7 * 24 * time.Hour // 会话 7天
)

// ================================
// 权限相关缓存
// ================================

// CheckUserPermission 检查用户权限（带缓存）
// 使用Redis Set存储用户的所有权限资源，格式：permission:userID -> Set["GET_/api/v1/users", "POST_/api/v1/posts"]
func (s *cacheService) CheckUserPermission(ctx context.Context, userID uint, path, method string) (bool, error) {
	if s.client == nil {
		// 缓存不可用，直接查询数据库
		return rbac.CheckPermission(userID, path, method)
	}

	cacheKey := fmt.Sprintf(cacheKeyPermission, userID)
	member := fmt.Sprintf("%s_%s", method, path)

	// 检查缓存是否存在
	exists, err := s.client.Exists(ctx, cacheKey)
	if err != nil {
		// 缓存查询失败，降级到数据库
		return rbac.CheckPermission(userID, path, method)
	}

	if !exists {
		// 缓存不存在，返回特殊错误，由调用方处理
		return false, fmt.Errorf("cache_miss")
	}

	// 检查是否是集合成员
	isMember, err := s.client.SIsMember(ctx, cacheKey, member)
	if err != nil {
		// 查询失败，降级到数据库
		return rbac.CheckPermission(userID, path, method)
	}

	// 刷新TTL（保持活跃用户的缓存）
	_ = s.client.Expire(ctx, cacheKey, ttlPermission)

	return isMember, nil
}

// SetUserPermissions 设置用户权限缓存
// 应该在用户登录或权限变更后调用
func (s *cacheService) SetUserPermissions(ctx context.Context, userID uint, resources []rbac.Resource) error {
	if s.client == nil {
		return nil
	}

	cacheKey := fmt.Sprintf(cacheKeyPermission, userID)

	// 先删除旧缓存
	_ = s.client.Delete(ctx, cacheKey)

	// 如果没有权限，就不设置缓存
	if len(resources) == 0 {
		return nil
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

	// 设置过期时间
	return s.client.Expire(ctx, cacheKey, ttlPermission)
}

// ClearUserPermissions 清除指定用户的所有权限缓存
func (s *cacheService) ClearUserPermissions(ctx context.Context, userID uint) error {
	if s.client == nil {
		return nil
	}

	key := fmt.Sprintf(cacheKeyPermission, userID)
	return s.client.Delete(ctx, key)
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

// ================================
// JWT Token 管理
// ================================

// SaveUserSession 保存用户会话
func (s *cacheService) SaveUserSession(ctx context.Context, userID uint, sessionID string, ttl time.Duration) error {
	if s.client == nil {
		return nil
	}

	cacheKey := fmt.Sprintf(cacheKeyUserSessions, userID)

	// 添加到用户会话集合
	if err := s.client.SAdd(ctx, cacheKey, sessionID); err != nil {
		return err
	}

	// 设置过期时间
	return s.client.Expire(ctx, cacheKey, ttl)
}

// RemoveUserSession 移除用户会话
func (s *cacheService) RemoveUserSession(ctx context.Context, userID uint, sessionID string) error {
	if s.client == nil {
		return nil
	}

	// 从用户会话集合中移除
	cacheKey := fmt.Sprintf(cacheKeyUserSessions, userID)
	if err := s.client.SRem(ctx, cacheKey, sessionID); err != nil {
		return err
	}

	// 删除会话令牌
	sessionTokenKey := fmt.Sprintf(cacheKeySessionTokens, sessionID)
	return s.client.Delete(ctx, sessionTokenKey)
}

// GetUserSessions 获取用户所有会话
func (s *cacheService) GetUserSessions(ctx context.Context, userID uint) ([]string, error) {
	if s.client == nil {
		return nil, nil
	}

	cacheKey := fmt.Sprintf(cacheKeyUserSessions, userID)
	members, err := s.client.SMembers(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	// 转换为字符串切片
	sessions := make([]string, 0, len(members))
	for _, member := range members {
		if str, ok := member.(string); ok {
			sessions = append(sessions, str)
		}
	}

	return sessions, nil
}

// RevokeAllUserSessions 撤销用户所有会话
func (s *cacheService) RevokeAllUserSessions(ctx context.Context, userID uint) error {
	if s.client == nil {
		return nil
	}

	// 获取所有会话
	sessions, err := s.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	// 删除所有会话令牌
	for _, sessionID := range sessions {
		sessionTokenKey := fmt.Sprintf(cacheKeySessionTokens, sessionID)
		_ = s.client.Delete(ctx, sessionTokenKey)
	}

	// 删除用户会话集合
	cacheKey := fmt.Sprintf(cacheKeyUserSessions, userID)
	return s.client.Delete(ctx, cacheKey)
}

// SaveSessionTokens 保存会话令牌
func (s *cacheService) SaveSessionTokens(ctx context.Context, sessionID string, accessToken, refreshToken string, ttl time.Duration) error {
	if s.client == nil {
		return nil
	}

	cacheKey := fmt.Sprintf(cacheKeySessionTokens, sessionID)
	tokens := map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	return s.client.Set(ctx, cacheKey, tokens, ttl)
}

// GetSessionTokens 获取会话令牌
func (s *cacheService) GetSessionTokens(ctx context.Context, sessionID string) (string, string, error) {
	if s.client == nil {
		return "", "", fmt.Errorf("cache not available")
	}

	cacheKey := fmt.Sprintf(cacheKeySessionTokens, sessionID)
	var tokens map[string]string
	if err := s.client.Get(ctx, cacheKey, &tokens); err != nil {
		return "", "", err
	}

	return tokens["access_token"], tokens["refresh_token"], nil
}

// IncrementRefreshCount 增加刷新计数
func (s *cacheService) IncrementRefreshCount(ctx context.Context, refreshToken string, ttl time.Duration) (int, error) {
	if s.client == nil {
		return 0, nil
	}

	cacheKey := fmt.Sprintf(cacheKeyRefreshCount, refreshToken)

	// 增加计数
	count, err := s.client.Incr(ctx, cacheKey)
	if err != nil {
		return 0, err
	}

	// 设置过期时间
	_ = s.client.Expire(ctx, cacheKey, ttl)

	return int(count), nil
}

// GetRefreshCount 获取刷新计数
func (s *cacheService) GetRefreshCount(ctx context.Context, refreshToken string) (int, error) {
	if s.client == nil {
		return 0, nil
	}

	cacheKey := fmt.Sprintf(cacheKeyRefreshCount, refreshToken)
	var count int
	if err := s.client.Get(ctx, cacheKey, &count); err != nil {
		// 如果key不存在，返回0
		return 0, nil
	}

	return count, nil
}
