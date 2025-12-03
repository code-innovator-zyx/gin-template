package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// 自定义错误类型

var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrExpiredToken         = errors.New("expired token")
	ErrSessionInvalid       = errors.New("session invalid or revoked")
	ErrRefreshTokenStolen   = errors.New("refresh token stolen or reused")
	ErrRefreshNotAllowed    = errors.New("refresh not allowed")
	ErrUnsupportedTokenType = errors.New("unsupported token type")
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
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
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email,omitempty"`
	TokenType TokenType `json:"token_type"` // access or refresh
	DeviceID  string    `json:"device_id,omitempty"`
	SessionID string    `json:"session_id,omitempty"`
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

// ServiceOption 服务选项函数类型
type ServiceOption func(*JWTService)

// WithSessionManager 设置会话管理器
func WithSessionManager(manager SessionManager) ServiceOption {
	return func(s *JWTService) {
		s.sessionManager = manager
	}
}

// SessionInfo 会话信息
type SessionInfo struct {
	SessionID        string    `json:"session_id"`
	UserID           uint      `json:"user_id"`
	Username         string    `json:"username"`
	RefreshTokenHash string    `json:"refresh_hash"`
	ExpiresAt        time.Time `json:"expires_at"`
	Revoked          bool      `json:"revoked"`
}
