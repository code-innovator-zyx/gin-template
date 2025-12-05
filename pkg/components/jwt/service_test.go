package jwt

import (
	"context"
	cache2 "gin-admin/pkg/components/cache"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getJwtSvr 创建测试用的 JWT Service
func getJwtSvr() (Service, cache2.ICache) {
	// 使用分片内存缓存进行测试
	memCache := cache2.NewShardedMemoryCache(0)

	cfg := Config{
		Secret:             "test-secret-key-32-chars-minimum",
		Issuer:             "test",
		AccessTokenExpire:  time.Second * 2, // 2秒过期，方便测试
		RefreshTokenExpire: time.Hour * 1,   // 1小时
	}

	svc := NewJwtService(cfg, memCache)
	return svc, memCache
}

// ==============================================================================
// 基础Token生成和解析测试
// ==============================================================================

func TestGenerateTokenPair(t *testing.T) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	tp, err := svc.GenerateTokenPair(ctx, 1, "testuser", "test@example.com")

	require.NoError(t, err)
	assert.NotEmpty(t, tp.AccessToken)
	assert.NotEmpty(t, tp.RefreshToken)
	assert.Equal(t, "Bearer", tp.TokenType)
	assert.Greater(t, tp.ExpiresIn, int64(0))
}

func TestParseAccessToken(t *testing.T) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	tp, err := svc.GenerateTokenPair(ctx, 1, "user", "user@example.com")
	require.NoError(t, err)

	claims, err := svc.ParseAccessToken(ctx, tp.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "user", claims.Username)
	assert.Equal(t, "user@example.com", claims.Email)
	assert.Equal(t, TokenTypeAccess, claims.TokenType)
	assert.NotEmpty(t, claims.SessionID)
}

func TestAccessTokenExpired(t *testing.T) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	tp, err := svc.GenerateTokenPair(ctx, 1, "user", "email")
	require.NoError(t, err)

	// 等待 Token 过期（配置的是 2 秒）
	time.Sleep(time.Second * 3)
	_, err = svc.ParseAccessToken(ctx, tp.AccessToken)
	assert.Error(t, err)
}

// ==============================================================================
// Token 刷新测试
// ==============================================================================

func TestRefreshToken(t *testing.T) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	// 生成初始 Token Pair
	tp1, err := svc.GenerateTokenPair(ctx, 1, "user", "user@example.com")
	require.NoError(t, err)
	// 刷新 Token
	tp2, err := svc.RefreshToken(ctx, tp1.RefreshToken)
	require.NoError(t, err)

	assert.NotEqual(t, tp1.AccessToken, tp2.AccessToken)
	assert.NotEqual(t, tp1.RefreshToken, tp2.RefreshToken)
	// 验证新的 Access Token 可用
	claims2, err2 := svc.ParseAccessToken(ctx, tp2.AccessToken)
	require.NoError(t, err2)
	assert.Equal(t, uint(1), claims2.UserID)
}

func TestRefreshTokenRotation(t *testing.T) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	tp1, err := svc.GenerateTokenPair(ctx, 1, "user", "email")
	require.NoError(t, err)

	// 第一次刷新
	tp2, err := svc.RefreshToken(ctx, tp1.RefreshToken)
	require.NoError(t, err)
	assert.NotEqual(t, tp1.RefreshToken, tp2.RefreshToken)

	// 第二次刷新
	tp3, err := svc.RefreshToken(ctx, tp2.RefreshToken)
	require.NoError(t, err)
	assert.NotEqual(t, tp2.RefreshToken, tp3.RefreshToken)
}

func TestRefreshTokenStolen(t *testing.T) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	tp, err := svc.GenerateTokenPair(ctx, 1, "user", "email")
	require.NoError(t, err)

	// 第一次刷新（合法）
	_, err = svc.RefreshToken(ctx, tp.RefreshToken)
	require.NoError(t, err)

	// 尝试使用旧的 Refresh Token（模拟盗用）
	_, err = svc.RefreshToken(ctx, tp.RefreshToken)
	assert.Error(t, err)
	assert.Equal(t, ErrRefreshTokenStolen, err)
}

// ==============================================================================
// 并发测试
// ==============================================================================

func TestConcurrentRefresh(t *testing.T) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	tp, err := svc.GenerateTokenPair(ctx, 1, "user", "email")
	require.NoError(t, err)

	// 模拟 10 个并发请求同时刷新同一个 Refresh Token
	results := make(chan *TokenPair, 10)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			newTP, err := svc.RefreshToken(ctx, tp.RefreshToken)
			if err != nil {
				errors <- err
			} else {
				results <- newTP
			}
		}()
	}

	// 收集结果
	var successCount int
	var firstToken *TokenPair
	for i := 0; i < 10; i++ {
		select {
		case result := <-results:
			successCount++
			if firstToken == nil {
				firstToken = result
			} else {
				// 所有成功的请求应该得到相同的 Token（Singleflight 合并）
				assert.Equal(t, firstToken.AccessToken, result.AccessToken)
			}
		case <-errors:
			// 有些请求可能失败，但至少应该有一个成功
		case <-time.After(time.Second * 2):
			t.Fatal("超时等待并发刷新结果")
		}
	}

	assert.Greater(t, successCount, 0)
}

func TestConcurrentGenerate(t *testing.T) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	const goroutines = 50
	var wg sync.WaitGroup
	errors := make(chan error, goroutines)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			_, err := svc.GenerateTokenPair(ctx, uint(id), "user", "email")
			if err != nil {
				errors <- err
			}
		}(i)
	}
	wg.Wait()
	close(errors)

	assert.Empty(t, errors)
}

// ==============================================================================
// Session 撤销测试
// ==============================================================================

func TestRevokeSession(t *testing.T) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	tp, err := svc.GenerateTokenPair(ctx, 1, "user", "email")
	require.NoError(t, err)

	claims, err := svc.ParseAccessToken(ctx, tp.AccessToken)
	require.NoError(t, err)

	// 撤销 Session
	err = svc.RevokeSession(ctx, claims.SessionID)
	require.NoError(t, err)

	// 尝试使用已撤销的 Access Token
	_, err = svc.ParseAccessToken(ctx, tp.AccessToken)
	assert.Error(t, err)
	assert.Equal(t, ErrSessionInvalid, err)

	// 尝试刷新已撤销的 Session
	_, err = svc.RefreshToken(ctx, tp.RefreshToken)
	assert.Error(t, err)
}

func TestRevokeUserAllSessions(t *testing.T) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	// 模拟多设备登录（生成 3 个 Session）
	tp1, err := svc.GenerateTokenPair(ctx, 1, "user", "email", WithDeviceID("device-1"))
	require.NoError(t, err)
	tp2, err := svc.GenerateTokenPair(ctx, 1, "user", "email", WithDeviceID("device-2"))
	require.NoError(t, err)
	tp3, err := svc.GenerateTokenPair(ctx, 1, "user", "email", WithDeviceID("device-3"))
	require.NoError(t, err)

	// 验证所有 Token 都可用
	_, err = svc.ParseAccessToken(ctx, tp1.AccessToken)
	require.NoError(t, err)
	_, err = svc.ParseAccessToken(ctx, tp2.AccessToken)
	require.NoError(t, err)
	_, err = svc.ParseAccessToken(ctx, tp3.AccessToken)
	require.NoError(t, err)

	// 撤销用户所有 Session
	err = svc.RevokeUserAllSessions(ctx, 1)
	require.NoError(t, err)

	// 验证所有 Token 都不可用
	_, err = svc.ParseAccessToken(ctx, tp1.AccessToken)
	assert.Equal(t, ErrSessionInvalid, err)
	_, err = svc.ParseAccessToken(ctx, tp2.AccessToken)
	assert.Equal(t, ErrSessionInvalid, err)
	_, err = svc.ParseAccessToken(ctx, tp3.AccessToken)
	assert.Equal(t, ErrSessionInvalid, err)
}

// ==============================================================================
// 无效Token测试
// ==============================================================================

func TestInvalidToken(t *testing.T) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	tests := []struct {
		name  string
		token string
	}{
		{"空 Token", ""},
		{"格式错误", "invalid.token.format"},
		{"签名错误", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ParseAccessToken(ctx, tt.token)
			assert.Error(t, err)
		})
	}
}

func TestInvalidRefreshToken(t *testing.T) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	tests := []struct {
		name  string
		token string
	}{
		{"空 Refresh Token", ""},
		{"格式错误", "invalid.refresh.token"},
		{"已过期的Token", "expired.token.xxxx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.RefreshToken(ctx, tt.token)
			assert.Error(t, err)
		})
	}
}

// ==============================================================================
// Token 选项测试
// ==============================================================================

func TestTokenOptions(t *testing.T) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	tp, err := svc.GenerateTokenPair(
		ctx,
		1,
		"user",
		"email",
		WithDeviceID("test-device"),
	)
	require.NoError(t, err)

	claims, err := svc.ParseAccessToken(ctx, tp.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "test-device", claims.DeviceID)
}

func TestMultipleDevices(t *testing.T) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	// 同一用户，多个设备
	devices := []string{"mobile", "tablet", "desktop"}
	tokens := make([]*TokenPair, len(devices))

	for i, device := range devices {
		tp, err := svc.GenerateTokenPair(ctx, 1, "user", "email", WithDeviceID(device))
		require.NoError(t, err)
		tokens[i] = tp
	}

	// 验证所有设备的Token都有效
	for _, tp := range tokens {
		_, err := svc.ParseAccessToken(ctx, tp.AccessToken)
		assert.NoError(t, err)
	}
}

// ==============================================================================
// 边界情况测试
// ==============================================================================

func TestEdgeCases(t *testing.T) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	t.Run("空用户名", func(t *testing.T) {
		tp, err := svc.GenerateTokenPair(ctx, 1, "", "email")
		assert.NoError(t, err)
		assert.NotEmpty(t, tp.AccessToken)
	})

	t.Run("空邮箱", func(t *testing.T) {
		tp, err := svc.GenerateTokenPair(ctx, 1, "user", "")
		assert.NoError(t, err)
		assert.NotEmpty(t, tp.AccessToken)
	})

	t.Run("UserID为0", func(t *testing.T) {
		tp, err := svc.GenerateTokenPair(ctx, 0, "user", "email")
		assert.NoError(t, err)

		claims, err := svc.ParseAccessToken(ctx, tp.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, uint(0), claims.UserID)
	})
}

func TestRevokeNonExistentSession(t *testing.T) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	// 撤销不存在的 Session应该不报错
	err := svc.RevokeSession(ctx, "non-existent-session-id")
	assert.NoError(t, err)
}

func TestRefreshExpiredRefreshToken(t *testing.T) {
	// 创建一个极短过期时间的JWT Service
	memCache := cache2.NewShardedMemoryCache(0)
	defer memCache.Close()

	cfg := Config{
		Secret:             "test-secret-key-32-chars-minimum",
		Issuer:             "test",
		AccessTokenExpire:  time.Millisecond * 100,
		RefreshTokenExpire: time.Millisecond * 200, // 200ms过期
	}
	svc := NewJwtService(cfg, memCache)
	ctx := context.Background()

	tp, err := svc.GenerateTokenPair(ctx, 1, "user", "email")
	require.NoError(t, err)

	// 等待 Refresh Token过期
	time.Sleep(time.Millisecond * 300)

	_, err = svc.RefreshToken(ctx, tp.RefreshToken)
	assert.Error(t, err)
}

// ==============================================================================
// 性能基准测试
// ==============================================================================

func BenchmarkGenerateTokenPair(b *testing.B) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GenerateTokenPair(ctx, uint(i), "user", "email")
	}
}

func BenchmarkParseAccessToken(b *testing.B) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	tp, _ := svc.GenerateTokenPair(ctx, 1, "user", "email")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.ParseAccessToken(ctx, tp.AccessToken)
	}
}

func BenchmarkRefreshToken(b *testing.B) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tp, _ := svc.GenerateTokenPair(ctx, 1, "user", "email")
		b.StartTimer()

		_, _ = svc.RefreshToken(ctx, tp.RefreshToken)
	}
}

func BenchmarkConcurrentParse(b *testing.B) {
	svc, _ := getJwtSvr()

	ctx := context.Background()

	tp, _ := svc.GenerateTokenPair(ctx, 1, "user", "email")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = svc.ParseAccessToken(ctx, tp.AccessToken)
		}
	})
}
