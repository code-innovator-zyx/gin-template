package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// Service JWT服务接口
type Service interface {
	GenerateTokenPair(ctx context.Context, userID uint, username, email string, opts ...TokenOption) (*TokenPair, error)
	ParseAccessToken(ctx context.Context, tokenString string) (*CustomClaims, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeUserAllSessions(ctx context.Context, userID uint) error
}

// =======================
// JWTService 实现
// =======================

type JWTService struct {
	config         *Config
	sessionManager SessionManager // 关键：所有状态集中在这里
}

// NewJWTService 创建实例
func NewJWTService(config *Config) Service {
	svc := &JWTService{
		config: config,
	}
	svc.sessionManager = NewRedisSessionManager()

	return svc
}

// =======================
// Token Pair 生成
// =======================

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
	refreshToken, refreshHash, err := s.generateRefreshToken(userID, username, tokenOpts)
	if err != nil {
		return nil, err
	}

	// 保存 session 状态
	s.sessionManager.SaveSession(ctx, &SessionInfo{
		SessionID:        tokenOpts.SessionID,
		UserID:           userID,
		Username:         username,
		RefreshTokenHash: refreshHash,
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

func (s *JWTService) generateRefreshToken(userID uint, username string, opts *TokenOptions) (token string, hash string, err error) {
	now := time.Now()
	claims := RefreshTokenClaims{
		UserID:       userID,
		Username:     username,
		TokenType:    TokenTypeRefresh,
		DeviceID:     opts.DeviceID,
		SessionID:    opts.SessionID,
		RefreshCount: 0,
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
		return "", "", err
	}

	return token, Hash(token), nil
}

// =======================
// Parse Access Token
// =======================

func (s *JWTService) ParseAccessToken(ctx context.Context, tokenString string) (*CustomClaims, error) {
	claims := &CustomClaims{}

	// 解析 token
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.config.Secret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	// 必须是 access token
	if claims.TokenType != TokenTypeAccess {
		return nil, ErrInvalidToken
	}

	// session 校验（access token 虽然无状态，但需校验 session 是否被撤销）
	session := s.sessionManager.GetSession(ctx, claims.SessionID)
	if session == nil || session.Revoked {
		return nil, ErrSessionInvalid
	}

	return claims, nil
}

// =======================
// Refresh Token → 新 Token Pair
// =======================

func (s *JWTService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims := &RefreshTokenClaims{}
	_, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.config.Secret), nil
	})
	if err != nil || claims.TokenType != TokenTypeRefresh {
		return nil, ErrInvalidToken
	}

	// 从 Redis 获取 session
	session := s.sessionManager.GetSession(ctx, claims.SessionID)
	if session == nil || session.Revoked {
		return nil, ErrSessionInvalid
	}

	// 校验 refresh token hash
	if session.RefreshTokenHash != Hash(refreshToken) {
		// 说明 refresh token 被窃取（旧 token）
		s.sessionManager.RemoveSession(ctx, claims.SessionID)
		return nil, ErrRefreshTokenStolen
	}

	// Rotation：生成新 Refresh Token（保持 SessionID 不变）
	newRefreshToken, newHash, err := s.generateRefreshToken(claims.UserID, claims.Username, &TokenOptions{
		SessionID: claims.SessionID,
		DeviceID:  claims.DeviceID,
	})
	if err != nil {
		return nil, err
	}

	// 生成新的 Access Token
	accessToken, err := s.generateAccessToken(
		claims.UserID, claims.Username, "", &TokenOptions{
			SessionID: claims.SessionID,
			DeviceID:  claims.DeviceID,
		},
	)
	if err != nil {
		return nil, err
	}

	// 更新 session 里的 refresh hash
	s.sessionManager.UpdateRefreshHash(ctx, claims.SessionID, newHash)

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.config.AccessTokenExpire.Seconds()),
		ExpiresAt:    time.Now().Add(s.config.AccessTokenExpire),
		TokenType:    "Bearer",
	}, nil
}

// =======================
// Session 撤销（退出登录）
// =======================

func (s *JWTService) RevokeSession(ctx context.Context, sessionID string) error {
	return s.sessionManager.RemoveSession(ctx, sessionID)
}

func (s *JWTService) RevokeUserAllSessions(ctx context.Context, userID uint) error {
	return s.sessionManager.RemoveUserSessions(ctx, userID)
}
