package jwt

import (
	"context"
	"gin-admin/pkg/cache"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

/*
* @Author: zouyx
* @Email: zouyx@knowsec.com
* @Date:   2025/11/19 下午12:02
* @Package:
 */

func TestGenerateTokenPair(t *testing.T) {
	svc := NewJWTService(&Config{
		Secret:             "test-secret",
		Issuer:             "test",
		AccessTokenExpire:  time.Minute,
		RefreshTokenExpire: time.Hour,
	})

	ctx := context.Background()

	tp, err := svc.GenerateTokenPair(ctx, 1, "testuser", "test@example.com")
	t.Log(tp)
	assert.NoError(t, err)
	assert.NotEmpty(t, tp.AccessToken)
	assert.NotEmpty(t, tp.RefreshToken)
	assert.Equal(t, "Bearer", tp.TokenType)
}

func TestParseAccessToken(t *testing.T) {
	svc := NewJWTService(&Config{
		Secret:             "test-secret",
		Issuer:             "test",
		AccessTokenExpire:  time.Minute,
		RefreshTokenExpire: time.Hour,
	})

	ctx := context.Background()
	tp, _ := svc.GenerateTokenPair(ctx, 1, "user", "email")

	claims, err := svc.ParseAccessToken(ctx, tp.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, TokenTypeAccess, claims.TokenType)
}

func TestRefreshTokenRotation(t *testing.T) {
	svc := NewJWTService(&Config{
		Secret:             "test-secret",
		Issuer:             "test",
		AccessTokenExpire:  time.Minute,
		RefreshTokenExpire: time.Hour,
	})

	ctx := context.Background()

	// 生成 token pair
	tp, err := svc.GenerateTokenPair(ctx, 1, "user", "email")
	assert.NoError(t, err)

	// 刷新 token
	tp2, err := svc.RefreshToken(ctx, tp.RefreshToken)
	assert.NoError(t, err)
	assert.NotEqual(t, tp.RefreshToken, tp2.RefreshToken, "refresh 必须更新")

	// 校验 session 中的 hash 更新
	claims := &RefreshTokenClaims{}
	_, _ = jwt.ParseWithClaims(tp2.RefreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})

	session := cache.GetGlobalCache().RedisClient().Get(ctx, "jwt:session:"+claims.SessionID).Val()
	assert.NotEmpty(t, session)
}

func TestRefreshTokenStolen(t *testing.T) {
	svc := NewJWTService(&Config{
		Secret:             "test-secret",
		Issuer:             "test",
		AccessTokenExpire:  time.Minute,
		RefreshTokenExpire: time.Hour,
	})

	ctx := context.Background()

	tp, _ := svc.GenerateTokenPair(ctx, 1, "user", "email")

	// 第一次刷新（合法）
	_, err := svc.RefreshToken(ctx, tp.RefreshToken)
	assert.NoError(t, err)

	// 第二次使用“旧 Refresh Token”，应该被判定为盗用
	_, err = svc.RefreshToken(ctx, tp.RefreshToken)
	assert.Error(t, err)
	assert.Equal(t, ErrRefreshTokenStolen, err)
}

func TestRevokeSession(t *testing.T) {
	svc := NewJWTService(&Config{
		Secret:             "test-secret",
		Issuer:             "test",
		AccessTokenExpire:  time.Minute,
		RefreshTokenExpire: time.Hour,
	})

	ctx := context.Background()

	tp, _ := svc.GenerateTokenPair(ctx, 1, "user", "email")

	claims, _ := svc.ParseAccessToken(ctx, tp.AccessToken)

	err := svc.RevokeSession(ctx, claims.SessionID)
	assert.NoError(t, err)

	_, err = svc.ParseAccessToken(ctx, tp.AccessToken)
	assert.Equal(t, ErrSessionInvalid, err)
}

func TestRevokeUserAllSessions(t *testing.T) {
	svc := NewJWTService(&Config{
		Secret:             "test-secret",
		Issuer:             "test",
		AccessTokenExpire:  time.Minute,
		RefreshTokenExpire: time.Hour,
	})

	ctx := context.Background()

	// 多设备会话
	svc.GenerateTokenPair(ctx, 1, "user", "email")
	svc.GenerateTokenPair(ctx, 1, "user", "email")
	svc.GenerateTokenPair(ctx, 1, "user", "email")

	keys, _ := cache.GetGlobalCache().RedisClient().Keys(ctx, "jwt:session:*").Result()
	assert.True(t, len(keys) >= 3)

	// 删除所有 session
	err := svc.RevokeUserAllSessions(ctx, 1)
	assert.NoError(t, err)

	keys, _ = cache.GetGlobalCache().RedisClient().Keys(ctx, "jwt:session:*").Result()
	assert.Equal(t, 0, len(keys))
}
