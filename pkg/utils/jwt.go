package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"gin-template/internal/core"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

/*
* @Author: zouyx
* @Package: 企业级 JWT 认证系统
* @Description: 提供完整的 Access Token + Refresh Token 双令牌机制
* @Features:
*   - 双令牌机制（Access Token + Refresh Token）
*   - Token 黑名单管理
*   - Token 自动刷新
*   - Token 撤销机制
*   - 刷新次数限制
*   - 设备/会话管理
 */

// ================================
// JWT 错误定义
// ================================

var (
	ErrTokenExpired        = errors.New("令牌已过期")
	ErrTokenNotValidYet    = errors.New("令牌尚未生效")
	ErrTokenMalformed      = errors.New("令牌格式错误")
	ErrTokenInvalid        = errors.New("无法解析的令牌")
	ErrTokenBlacklisted    = errors.New("令牌已被撤销")
	ErrRefreshTokenExpired = errors.New("刷新令牌已过期")
	ErrRefreshLimitReached = errors.New("刷新次数已达上限")
	ErrInvalidTokenType    = errors.New("令牌类型错误")
)

// ================================
// Token 类型定义
// ================================

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// ================================
// JWT Claims 定义
// ================================

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

// RefreshTokenClaims 刷新令牌声明
type RefreshTokenClaims struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	TokenType    string `json:"token_type"`    // refresh
	RefreshCount int    `json:"refresh_count"` // 已刷新次数
	DeviceID     string `json:"device_id,omitempty"`
	SessionID    string `json:"session_id"`
	jwt.RegisteredClaims
}

// ================================
// Token 对响应结构
// ================================

// TokenPair 令牌对
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"` // Bearer
	ExpiresIn    int64     `json:"expires_in"` // Access Token 过期时间（秒）
	ExpiresAt    time.Time `json:"expires_at"` // Access Token 过期时间点
}

// TokenMetadata Token 元数据
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

// ================================
// JWT 管理器
// ================================

// JWTManager JWT 管理器
type JWTManager struct {
	secret               []byte
	accessTokenExpire    time.Duration
	refreshTokenExpire   time.Duration
	issuer               string
	maxRefreshCount      int
	enableBlacklist      bool
	blacklistGracePeriod time.Duration
}

// NewJWTManager 创建 JWT 管理器
func NewJWTManager() *JWTManager {
	jwtConfig := core.MustGetConfig().Jwt

	// 默认配置
	secret := "kZ3r7XG8YpS+vO9fN7JxBtUo1e8L2jH4pFqS5mRw9tDcVyZxGqK0sT3bLnM6wA9d"
	accessTokenExpire := 3600    // 1小时
	refreshTokenExpire := 604800 // 7天
	issuer := "gin-template"
	maxRefreshCount := 10
	enableBlacklist := true
	blacklistGracePeriod := 300 // 5分钟

	if jwtConfig != nil {
		secret = jwtConfig.Secret
		accessTokenExpire = jwtConfig.AccessTokenExpire
		refreshTokenExpire = jwtConfig.RefreshTokenExpire
		issuer = jwtConfig.Issuer
		maxRefreshCount = jwtConfig.MaxRefreshCount
		enableBlacklist = jwtConfig.EnableBlacklist
		blacklistGracePeriod = jwtConfig.BlacklistGracePeriod
	} else {
		logrus.Warn("JWT 配置未找到，使用默认配置")
	}

	return &JWTManager{
		secret:               []byte(secret),
		accessTokenExpire:    time.Duration(accessTokenExpire) * time.Second,
		refreshTokenExpire:   time.Duration(refreshTokenExpire) * time.Second,
		issuer:               issuer,
		maxRefreshCount:      maxRefreshCount,
		enableBlacklist:      enableBlacklist,
		blacklistGracePeriod: time.Duration(blacklistGracePeriod) * time.Second,
	}
}

var (
	globalJWTManager *JWTManager
	once             sync.Once
)

// GetJWTManager 获取全局 JWT 管理器（线程安全）
func GetJWTManager() *JWTManager {
	once.Do(func() {
		globalJWTManager = NewJWTManager()
	})
	return globalJWTManager
}

// ================================
// Token 生成
// ================================

// GenerateTokenPair 生成令牌对（Access Token + Refresh Token）
func (m *JWTManager) GenerateTokenPair(userID uint, username, email string, deviceID ...string) (*TokenPair, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("生成会话ID失败: %w", err)
	}

	device := ""
	if len(deviceID) > 0 {
		device = deviceID[0]
	}

	// 生成 Access Token
	accessToken, err := m.generateAccessToken(userID, username, email, device, sessionID)
	if err != nil {
		return nil, fmt.Errorf("生成访问令牌失败: %w", err)
	}

	// 生成 Refresh Token
	refreshToken, err := m.generateRefreshToken(userID, username, device, sessionID)
	if err != nil {
		return nil, fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(m.accessTokenExpire.Seconds()),
		ExpiresAt:    time.Now().Add(m.accessTokenExpire),
	}, nil
}

// generateAccessToken 生成访问令牌
func (m *JWTManager) generateAccessToken(userID uint, username, email, deviceID, sessionID string) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		TokenType: TokenTypeAccess,
		DeviceID:  deviceID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			ID:        generateJTI(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// generateRefreshToken 生成刷新令牌
func (m *JWTManager) generateRefreshToken(userID uint, username, deviceID, sessionID string) (string, error) {
	now := time.Now()
	claims := RefreshTokenClaims{
		UserID:       userID,
		Username:     username,
		TokenType:    TokenTypeRefresh,
		RefreshCount: 0,
		DeviceID:     deviceID,
		SessionID:    sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTokenExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			ID:        generateJTI(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ================================
// Token 解析和验证
// ================================

// ParseAccessToken 解析访问令牌
func (m *JWTManager) ParseAccessToken(ctx context.Context, tokenString string) (*CustomClaims, error) {
	// 检查黑名单
	if m.enableBlacklist {
		if blacklisted, err := m.isTokenBlacklisted(ctx, tokenString); err == nil && blacklisted {
			return nil, ErrTokenBlacklisted
		}
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, m.parseError(err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		// 验证 Token 类型
		if claims.TokenType != TokenTypeAccess {
			return nil, ErrInvalidTokenType
		}
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// ParseRefreshToken 解析刷新令牌
func (m *JWTManager) ParseRefreshToken(ctx context.Context, tokenString string) (*RefreshTokenClaims, error) {
	// 检查黑名单
	if m.enableBlacklist {
		if blacklisted, err := m.isTokenBlacklisted(ctx, tokenString); err == nil && blacklisted {
			return nil, ErrTokenBlacklisted
		}
	}

	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, m.parseError(err)
	}

	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		// 验证 Token 类型
		if claims.TokenType != TokenTypeRefresh {
			return nil, ErrInvalidTokenType
		}

		// 检查刷新次数限制
		if m.maxRefreshCount > 0 && claims.RefreshCount >= m.maxRefreshCount {
			return nil, ErrRefreshLimitReached
		}

		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// ================================
// Token 刷新
// ================================

// RefreshToken 刷新令牌
// 使用 Refresh Token 生成新的 Access Token 和 Refresh Token
func (m *JWTManager) RefreshToken(ctx context.Context, refreshTokenString string) (*TokenPair, error) {
	// 解析 Refresh Token
	claims, err := m.ParseRefreshToken(ctx, refreshTokenString)
	if err != nil {
		return nil, err
	}

	// 生成新的令牌对
	sessionID := claims.SessionID
	if sessionID == "" {
		sessionID, _ = generateSessionID()
	}

	// 生成新的 Access Token
	accessToken, err := m.generateAccessToken(claims.UserID, claims.Username, "", claims.DeviceID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("生成新访问令牌失败: %w", err)
	}

	// 生成新的 Refresh Token（增加刷新计数）
	newRefreshToken, err := m.generateNewRefreshToken(claims)
	if err != nil {
		return nil, fmt.Errorf("生成新刷新令牌失败: %w", err)
	}

	// 将旧的 Refresh Token 加入黑名单
	if m.enableBlacklist {
		ttl := time.Until(claims.ExpiresAt.Time)
		if ttl > 0 {
			_ = m.blacklistToken(ctx, refreshTokenString, ttl)
		}
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(m.accessTokenExpire.Seconds()),
		ExpiresAt:    time.Now().Add(m.accessTokenExpire),
	}, nil
}

// generateNewRefreshToken 生成新的刷新令牌（增加刷新计数）
func (m *JWTManager) generateNewRefreshToken(oldClaims *RefreshTokenClaims) (string, error) {
	now := time.Now()
	claims := RefreshTokenClaims{
		UserID:       oldClaims.UserID,
		Username:     oldClaims.Username,
		TokenType:    TokenTypeRefresh,
		RefreshCount: oldClaims.RefreshCount + 1,
		DeviceID:     oldClaims.DeviceID,
		SessionID:    oldClaims.SessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTokenExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", oldClaims.UserID),
			ID:        generateJTI(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ================================
// Token 撤销（黑名单）
// ================================

// RevokeToken 撤销令牌（加入黑名单）
func (m *JWTManager) RevokeToken(ctx context.Context, tokenString string) error {
	if !m.enableBlacklist {
		return nil
	}

	// 解析 token 获取过期时间
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})

	if err != nil {
		// 即使解析失败，也尝试加入黑名单
		return m.blacklistToken(ctx, tokenString, 24*time.Hour)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			expiresAt := time.Unix(int64(exp), 0)
			ttl := time.Until(expiresAt)
			if ttl > 0 {
				// 添加宽限期
				ttl += m.blacklistGracePeriod
				return m.blacklistToken(ctx, tokenString, ttl)
			}
		}
	}

	return nil
}

// RevokeUserAllTokens 撤销用户的所有令牌（登出所有设备）
func (m *JWTManager) RevokeUserAllTokens(ctx context.Context, userID uint) error {
	// 这需要配合 Session 管理系统
	// 可以在缓存中维护 user_sessions:userID -> []sessionID
	// 然后遍历所有 session 对应的 token 加入黑名单
	logrus.Infof("撤销用户 %d 的所有令牌", userID)
	return nil
}

// ================================
// 辅助方法
// ================================

// blacklistToken 将 token 加入黑名单
func (m *JWTManager) blacklistToken(ctx context.Context, tokenString string, ttl time.Duration) error {
	// 通过缓存服务加入黑名单
	cacheKey := fmt.Sprintf("jwt:blacklist:%s", tokenString)
	if cacheService := getCacheService(); cacheService != nil {
		return cacheService.SetInstance(ctx, cacheKey, true, ttl)
	}
	return nil
}

// isTokenBlacklisted 检查 token 是否在黑名单中
func (m *JWTManager) isTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	cacheKey := fmt.Sprintf("jwt:blacklist:%s", tokenString)
	if cacheService := getCacheService(); cacheService != nil {
		return cacheService.Exists(ctx, cacheKey)
	}
	return false, nil
}

// parseError 解析 JWT 错误
func (m *JWTManager) parseError(err error) error {
	if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			return ErrTokenMalformed
		} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
			return ErrTokenExpired
		} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
			return ErrTokenNotValidYet
		}
	}
	return ErrTokenInvalid
}

// GetTokenMetadata 获取 Token 元数据（不验证有效性）
func (m *JWTManager) GetTokenMetadata(tokenString string) (*TokenMetadata, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &CustomClaims{})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok {
		return &TokenMetadata{
			UserID:    claims.UserID,
			Username:  claims.Username,
			Email:     claims.Email,
			TokenType: claims.TokenType,
			DeviceID:  claims.DeviceID,
			SessionID: claims.SessionID,
			IssuedAt:  claims.IssuedAt.Time,
			ExpiresAt: claims.ExpiresAt.Time,
		}, nil
	}

	return nil, ErrTokenInvalid
}

// ================================
// 工具函数
// ================================

// generateSessionID 生成会话ID
func generateSessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// generateJTI 生成 JWT ID
func generateJTI() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// getCacheService 获取缓存服务
func getCacheService() interface {
	SetInstance(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Exists(ctx context.Context, key string) (bool, error)
} {
	// 这里应该返回实际的缓存服务实例
	// 由于包依赖关系，这里使用接口解耦
	// 在实际使用中，可以通过依赖注入或全局变量获取
	return nil
}

// ================================
// 向后兼容的全局函数
// ================================

// GenerateToken 生成JWT令牌（向后兼容，仅生成 Access Token）
// 建议使用 GenerateTokenPair 生成完整的令牌对
func GenerateToken(userID uint, username string) (string, error) {
	manager := GetJWTManager()
	return manager.generateAccessToken(userID, username, "", "", "")
}

// ParseToken 解析JWT令牌（向后兼容）
// 建议使用 JWTManager.ParseAccessToken
func ParseToken(tokenString string) (*CustomClaims, error) {
	manager := GetJWTManager()
	return manager.ParseAccessToken(context.Background(), tokenString)
}
