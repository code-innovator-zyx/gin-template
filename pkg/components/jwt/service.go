package jwt

import (
	"context"
	"fmt"
	"gin-admin/pkg/cache"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const TokenPrefix string = "Bearer"

// Service JWT服务接口
type Service interface {
	// GenerateTokenPair 生成密钥对
	GenerateTokenPair(ctx context.Context, userID uint, username, email string, opts ...TokenOption) (*TokenPair, error)
	// ParseAccessToken 解析accessToken
	ParseAccessToken(ctx context.Context, tokenString string) (*CustomClaims, error)
	// RefreshToken 刷新token
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	// RevokeSession 撤销登录session
	RevokeSession(ctx context.Context, sessionId string) error
	RevokeUserAllSessions(ctx context.Context, userID uint) error
}

// =======================
// JWTService 实现
// =======================

type JWTService struct {
	config         Config
	sessionManager SessionManager
}

func NewJwtService(cfg Config, cache cache.ICache) *JWTService {
	return &JWTService{
		config:         cfg,
		sessionManager: NewCacheSessionManager(cache),
	}
}

func (s *JWTService) GenerateTokenPair(ctx context.Context, userID uint, username, email string, opts ...TokenOption) (*TokenPair, error) {
	// 初始化 token options（SessionID 必须写入 Claims）
	tokenOpts := &TokenOptions{
		DeviceID:  uuid.New().String(),
		SessionID: uuid.New().String(),
	}
	for _, opt := range opts {
		opt(tokenOpts)
	}

	// 生成 access token
	accessToken, err := s.generateAccessToken(userID, username, email, tokenOpts)
	if err != nil {
		return nil, err
	}

	// 生成 refresh token
	refreshToken, err := s.generateRefreshToken(userID, username, tokenOpts)
	if err != nil {
		return nil, err
	}

	// 保存 session 状态
	s.sessionManager.SaveSession(ctx, &SessionInfo{
		SessionID:        tokenOpts.SessionID,
		UserID:           userID,
		Username:         username,
		RefreshTokenHash: Hash(refreshToken),
		ExpiresAt:        time.Now().Add(s.config.RefreshTokenExpire),
	})

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.AccessTokenExpire.Seconds()),
		ExpiresAt:    time.Now().Add(s.config.AccessTokenExpire),
		TokenType:    "Bearer",
	}, nil
}

// =======================
// Access Token
// =======================

func (s *JWTService) generateAccessToken(userID uint, username, email string, opts *TokenOptions) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		TokenType: TokenTypeAccess,
		DeviceID:  opts.DeviceID,
		SessionID: opts.SessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.AccessTokenExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
			Subject:   fmt.Sprintf("%d", userID),
			ID:        uuid.New().String(),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.config.Secret))
}

// =======================
// Refresh Token + Rotation
// =======================

func (s *JWTService) generateRefreshToken(userID uint, username string, opts *TokenOptions) (token string, err error) {
	now := time.Now()
	claims := CustomClaims{
		UserID:    userID,
		Username:  username,
		TokenType: TokenTypeRefresh,
		DeviceID:  opts.DeviceID,
		SessionID: opts.SessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.RefreshTokenExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
			Subject:   fmt.Sprintf("%d", userID),
			ID:        uuid.New().String(),
		},
	}

	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", err
	}

	return token, nil
}

// =======================
// Parse Access Token
// =======================

func (s *JWTService) ParseAccessToken(ctx context.Context, tokenString string) (*CustomClaims, error) {
	// 解析 token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.config.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		// 验证 Token 类型
		if claims.TokenType != TokenTypeAccess {
			return nil, ErrInvalidToken
		}
		return claims, nil
	}
	return nil, nil
}

// =======================
// Refresh Token → 新 Token Pair
// =======================

func (s *JWTService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	token, err := jwt.ParseWithClaims(refreshToken, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.config.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.TokenType != TokenTypeRefresh {
		return nil, ErrInvalidToken
	}
	// 从 Redis 获取 session
	session := s.sessionManager.GetSession(ctx, claims.SessionID)
	if session == nil || session.Revoked {
		// session 不存在或者已经被注销
		return nil, ErrSessionInvalid
	}
	// 校验 refresh token hash
	if !SecureCompare(Hash(refreshToken), session.RefreshTokenHash) {
		// 说明 refresh token 不存在或者被窃取盗用
		s.sessionManager.RemoveSession(ctx, claims.SessionID)
		return nil, ErrRefreshTokenStolen
	}

	// Rotation：生成新 Refresh Token（保持 SessionID 不变）
	newRefreshToken, err := s.generateRefreshToken(claims.UserID, claims.Username, &TokenOptions{
		SessionID: claims.SessionID,
		DeviceID:  claims.DeviceID,
	})
	if err != nil {
		return nil, err
	}

	// 生成新的 Access Token
	accessToken, err := s.generateAccessToken(
		claims.UserID, claims.Username, claims.Email, &TokenOptions{
			SessionID: claims.SessionID,
			DeviceID:  claims.DeviceID,
		},
	)
	if err != nil {
		return nil, err
	}

	// 更新 session 里的 refresh hash
	err = s.sessionManager.UpdateRefreshHash(ctx, claims.SessionID, Hash(newRefreshToken))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.config.AccessTokenExpire.Seconds()),
		ExpiresAt:    time.Now().Add(s.config.AccessTokenExpire),
		TokenType:    TokenPrefix,
	}, nil
}

// =======================
// Session 撤销（退出登录）
// =======================

func (s *JWTService) RevokeSession(ctx context.Context, sessionId string) error {
	return s.sessionManager.RemoveSession(ctx, sessionId)
}

func (s *JWTService) RevokeUserAllSessions(ctx context.Context, userID uint) error {
	return s.sessionManager.RemoveUserSessions(ctx, userID)
}
