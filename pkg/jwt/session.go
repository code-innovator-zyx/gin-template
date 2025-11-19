package jwt

import (
	"context"
	"encoding/json"
	"fmt"
	"gin-admin/pkg/cache"
	"time"
)

// SessionInfo 会话信息
type SessionInfo struct {
	SessionID        string    `json:"session_id"`
	UserID           uint      `json:"user_id"`
	Username         string    `json:"username"`
	DeviceID         string    `json:"device_id"`
	RefreshTokenHash string    `json:"refresh_hash"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	Revoked          bool      `json:"revoked"`
}

// SessionManager 会话管理器接口
type SessionManager interface {
	SaveSession(ctx context.Context, s *SessionInfo) error
	GetSession(ctx context.Context, sessionID interface{}) *SessionInfo
	RemoveSession(ctx context.Context, sessionID string) error
	UpdateRefreshHash(ctx context.Context, sessionID, hash string) error
	GetUserSessions(ctx context.Context, userID uint) ([]*SessionInfo, error)
	RemoveUserSessions(ctx context.Context, userID uint) error
}
type CacheSessionManager struct {
	cache cache.Cache
}

func NewRedisSessionManager() SessionManager {
	return &CacheSessionManager{cache: cache.GetGlobalCache()}
}

func (m *CacheSessionManager) sessionKey(id interface{}) string {
	return fmt.Sprintf("jwt:session:%+v", id)
}

func (m *CacheSessionManager) userSessionsKey(uid uint) string {
	return fmt.Sprintf("jwt:user:%+v:sessions", uid)
}

// SaveSession 保存 session
func (m *CacheSessionManager) SaveSession(ctx context.Context, s *SessionInfo) error {
	data, _ := json.Marshal(s)

	ttl := time.Until(s.ExpiresAt)
	key := m.sessionKey(s.SessionID)

	pipe := m.cache.Pipeline()
	pipe.Set(ctx, key, data, ttl)
	pipe.SAdd(ctx, m.userSessionsKey(s.UserID), s.SessionID)
	pipe.Expire(ctx, m.userSessionsKey(s.UserID), ttl)

	err := pipe.Exec(ctx)
	return err
}

func (m *CacheSessionManager) GetSession(ctx context.Context, sessionID interface{}) *SessionInfo {
	var s SessionInfo
	err := m.cache.Get(ctx, m.sessionKey(sessionID), &s)
	if err != nil {
		return nil
	}
	return &s
}

func (m *CacheSessionManager) RemoveSession(ctx context.Context, sessionID string) error {
	s := m.GetSession(ctx, sessionID)
	if s == nil {
		return nil
	}
	pipe := m.cache.Pipeline()
	pipe.Del(ctx, m.sessionKey(sessionID))
	pipe.SRem(ctx, m.userSessionsKey(s.UserID), sessionID)
	return pipe.Exec(ctx)
}

// UpdateRefreshHash 刷新token 更新 sessionId的新token 防重入
func (m *CacheSessionManager) UpdateRefreshHash(ctx context.Context, sessionID, hash string) error {
	s := m.GetSession(ctx, sessionID)
	if s == nil {
		return nil
	}
	s.RefreshTokenHash = hash
	data, _ := json.Marshal(s)
	ttl := time.Until(s.ExpiresAt)
	return m.cache.Set(ctx, m.sessionKey(sessionID), data, ttl)
}

// GetUserSessions 获取用户所有会话
func (m *CacheSessionManager) GetUserSessions(ctx context.Context, userID uint) ([]*SessionInfo, error) {
	keys, err := m.cache.SMembers(ctx, m.userSessionsKey(userID))
	if err != nil {
		return nil, err
	}
	var list []*SessionInfo
	for _, sid := range keys {
		s := m.GetSession(ctx, sid)
		if s != nil {
			list = append(list, s)
		}
	}
	return list, nil
}

// RemoveUserSessions 删除用户所有 session（退出所有设备）
func (m *CacheSessionManager) RemoveUserSessions(ctx context.Context, userID uint) error {
	sessions, err := m.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}
	for _, s := range sessions {
		_ = m.RemoveSession(ctx, s.SessionID)
	}
	return nil
}
