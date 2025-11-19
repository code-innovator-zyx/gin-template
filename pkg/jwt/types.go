package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Token类型
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// TokenPair 令牌对响应结构
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"` // Bearer
	ExpiresIn    int64     `json:"expires_in"` // Access Token 过期时间（秒）
	ExpiresAt    time.Time `json:"expires_at"` // Access Token 过期时间点
}

// CustomClaims 自定义JWT声明（Access Token）
type CustomClaims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email,omitempty"`
	TokenType string `json:"token_type"` // access or refresh
	DeviceID  string `json:"device_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	jwt.RegisteredClaims
}

// RefreshTokenClaims 刷新令牌
type RefreshTokenClaims struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	DeviceID     string `json:"device_id,omitempty"`
	SessionID    string `json:"session_id"`
	TokenType    string `json:"token_type"`    // refresh
	RefreshCount int    `json:"refresh_count"` // 已刷新次数
	jwt.RegisteredClaims
}

// TokenOptions Token生成选项
type TokenOptions struct {
	DeviceID  string // 设备ID
	SessionID string // 会话ID
}

// TokenOption Token选项函数类型
type TokenOption func(*TokenOptions)

// WithDeviceID 设置设备ID
func WithDeviceID(deviceID string) TokenOption {
	return func(o *TokenOptions) {
		o.DeviceID = deviceID
	}
}

// WithSessionID 设置会话ID
func WithSessionID(sessionID string) TokenOption {
	return func(o *TokenOptions) {
		o.SessionID = sessionID
	}
}

// 自定义错误类型
var (
	ErrInvalidToken       = errors.New("token 无效")
	ErrTokenExpired       = errors.New("token 已过期")
	ErrSessionInvalid     = errors.New("会话已失效")
	ErrRefreshTokenStolen = errors.New("refresh token 已被盗用（旧 token）")
)

// ================================ Token结构定义 ================================

// TokenMetadata Token元数据
type TokenMetadata struct {
	UserID    uint
	Username  string
	Email     string
	TokenType string
	DeviceID  string
	SessionID string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// ================================ 选项结构 ================================

// ServiceOption 服务选项函数类型
type ServiceOption func(*JWTService)

// WithSessionManager 设置会话管理器
func WithSessionManager(manager SessionManager) ServiceOption {
	return func(s *JWTService) {
		s.sessionManager = manager
	}
}
