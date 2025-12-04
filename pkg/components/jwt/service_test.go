package jwt

import (
	"context"
	cache2 "gin-admin/pkg/components/cache"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// getJwtSvr 创建测试用的 JWT Service
func getJwtSvr() (Service, cache2.ICache) {
	// 使用内存缓存进行测试
	memCache := cache2.NewMemoryCache()

	cfg := Config{
		Secret:             "test-secret-key-32-chars-minimum",
		Issuer:             "test",
		AccessTokenExpire:  time.Second * 2, // 2秒过期，方便测试
		RefreshTokenExpire: time.Hour,       // 1小时
	}

	svc := NewJwtService(cfg, memCache)
	return svc, memCache
}

// TestGenerateTokenPair 测试生成 Token Pair
func TestGenerateTokenPair(t *testing.T) {
	svc, _ := getJwtSvr()
	ctx := context.Background()

	tp, err := svc.GenerateTokenPair(ctx, 1, "testuser", "test@example.com")

	// 断言
	assert.NoError(t, err, "生成 Token 应该成功")
	assert.NotEmpty(t, tp.AccessToken, "Access Token 不应为空")
	assert.NotEmpty(t, tp.RefreshToken, "Refresh Token 不应为空")
	assert.Equal(t, "Bearer", tp.TokenType, "Token 类型应该是 Bearer")
	assert.Greater(t, tp.ExpiresIn, int64(0), "ExpiresIn 应该大于 0")

	t.Logf("✅ Token Pair 生成成功: AccessToken=%d chars, RefreshToken=%d chars",
		len(tp.AccessToken), len(tp.RefreshToken))
}

// TestParseAccessToken 测试解析 Access Token
func TestParseAccessToken(t *testing.T) {
	svc, _ := getJwtSvr()
	ctx := context.Background()

	// 生成 Token
	tp, err := svc.GenerateTokenPair(ctx, 1, "user", "user@example.com")
	assert.NoError(t, err)

	// 解析 Access Token
	claims, err := svc.ParseAccessToken(ctx, tp.AccessToken)
	assert.NoError(t, err, "解析 Access Token 应该成功")
	assert.Equal(t, uint(1), claims.UserID, "UserID 应该匹配")
	assert.Equal(t, "user", claims.Username, "Username 应该匹配")
	assert.Equal(t, "user@example.com", claims.Email, "Email 应该匹配")
	assert.Equal(t, TokenTypeAccess, claims.TokenType, "Token 类型应该是 Access")
	assert.NotEmpty(t, claims.SessionID, "SessionID 不应为空")

	t.Logf("✅ Access Token 解析成功: UserID=%d, Username=%s, SessionID=%s",
		claims.UserID, claims.Username, claims.SessionID)
}

// TestAccessTokenExpired 测试 Access Token 过期
func TestAccessTokenExpired(t *testing.T) {
	svc, _ := getJwtSvr()
	ctx := context.Background()

	// 生成 Token
	tp, err := svc.GenerateTokenPair(ctx, 1, "user", "email")
	assert.NoError(t, err)

	// 等待 Token 过期（配置的是 2 秒）
	time.Sleep(time.Second * 3)

	// 解析过期的 Token
	_, err = svc.ParseAccessToken(ctx, tp.AccessToken)
	assert.Error(t, err, "过期的 Token 应该解析失败")
	// 注意：ParseAccessToken 返回的是 ErrInvalidToken，不是 jwt.ErrTokenExpired

	t.Logf("✅ Access Token 过期检测成功: %v", err)
}

// TestRefreshToken 测试刷新 Token
func TestRefreshToken(t *testing.T) {
	svc, memCache := getJwtSvr()
	ctx := context.Background()

	// 生成初始 Token Pair
	tp1, err := svc.GenerateTokenPair(ctx, 1, "user", "email")
	assert.NoError(t, err)

	// 调试：检查 cache 是否工作
	t.Logf("MemCache is nil: %v", memCache == nil)

	// 调试：验证 session 是否被保存
	claims1, _ := svc.ParseAccessToken(ctx, tp1.AccessToken)
	t.Logf("Session ID: %s", claims1.SessionID)

	// 刷新 Token
	tp2, err := svc.RefreshToken(ctx, tp1.RefreshToken)
	if err != nil {
		t.Logf("RefreshToken error: %v", err)
	}
	assert.NoError(t, err, "刷新 Token 应该成功")

	if tp2 != nil {
		assert.NotEqual(t, tp1.AccessToken, tp2.AccessToken, "新的 Access Token 应该不同")
		assert.NotEqual(t, tp1.RefreshToken, tp2.RefreshToken, "新的 Refresh Token 应该不同（Token Rotation）")

		// 验证新的 Access Token 可用
		claims, err := svc.ParseAccessToken(ctx, tp2.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), claims.UserID)
	}

	t.Logf("✅ Token 刷新成功（Token Rotation）")
}

// TestRefreshTokenRotation 测试 Token Rotation 机制
func TestRefreshTokenRotation(t *testing.T) {
	svc, _ := getJwtSvr()
	ctx := context.Background()

	// 生成 token pair
	tp1, err := svc.GenerateTokenPair(ctx, 1, "user", "email")
	assert.NoError(t, err)

	// 第一次刷新
	tp2, err := svc.RefreshToken(ctx, tp1.RefreshToken)
	assert.NoError(t, err)
	assert.NotEqual(t, tp1.RefreshToken, tp2.RefreshToken, "Refresh Token 必须更新")

	// 第二次刷新
	tp3, err := svc.RefreshToken(ctx, tp2.RefreshToken)
	assert.NoError(t, err)
	assert.NotEqual(t, tp2.RefreshToken, tp3.RefreshToken, "Refresh Token 每次都要更新")

	t.Logf("✅ Token Rotation 测试通过：3 次刷新，3 个不同的 Refresh Token")
}

// TestRefreshTokenStolen 测试 Refresh Token 被盗用检测
func TestRefreshTokenStolen(t *testing.T) {
	svc, _ := getJwtSvr()
	ctx := context.Background()

	// 生成 Token
	tp, err := svc.GenerateTokenPair(ctx, 1, "user", "email")
	assert.NoError(t, err)

	// 第一次刷新（合法）
	_, err = svc.RefreshToken(ctx, tp.RefreshToken)
	assert.NoError(t, err, "第一次刷新应该成功")

	// 尝试使用旧的 Refresh Token（模拟盗用）
	_, err = svc.RefreshToken(ctx, tp.RefreshToken)
	assert.Error(t, err, "重用旧 Refresh Token 应该失败")
	assert.Equal(t, ErrRefreshTokenStolen, err, "应该检测到 Token 被盗用")

	t.Logf("✅ Refresh Token 重用检测成功：%v", err)
}

// TestConcurrentRefresh 测试并发刷新（Singleflight）
func TestConcurrentRefresh(t *testing.T) {
	svc, _ := getJwtSvr()
	ctx := context.Background()

	// 生成 Token
	tp, err := svc.GenerateTokenPair(ctx, 1, "user", "email")
	assert.NoError(t, err)

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
				assert.Equal(t, firstToken.AccessToken, result.AccessToken,
					"并发刷新应该返回相同的 Token")
			}
		case <-errors:
			// 有些请求可能失败，但至少应该有一个成功
		case <-time.After(time.Second):
			t.Fatal("超时等待并发刷新结果")
		}
	}

	assert.Greater(t, successCount, 0, "至少应该有一个请求成功")
	t.Logf("✅ 并发刷新测试通过：10 个请求，%d 个成功，Token 一致", successCount)
}

// TestRevokeSession 测试撤销单个 Session
func TestRevokeSession(t *testing.T) {
	svc, _ := getJwtSvr()
	ctx := context.Background()

	// 生成 Token
	tp, err := svc.GenerateTokenPair(ctx, 1, "user", "email")
	assert.NoError(t, err)

	// 解析获取 SessionID
	claims, err := svc.ParseAccessToken(ctx, tp.AccessToken)
	assert.NoError(t, err)

	// 撤销 Session
	err = svc.RevokeSession(ctx, claims.SessionID)
	assert.NoError(t, err, "撤销 Session 应该成功")

	// 尝试使用已撤销的 Access Token
	_, err = svc.ParseAccessToken(ctx, tp.AccessToken)
	assert.Error(t, err, "已撤销的 Session 应该无法使用")
	assert.Equal(t, ErrSessionInvalid, err, "错误类型应该是 ErrSessionInvalid")

	// 尝试刷新已撤销的 Session
	_, err = svc.RefreshToken(ctx, tp.RefreshToken)
	assert.Error(t, err, "已撤销的 Session 不能刷新")

	t.Logf("✅ Session 撤销测试通过")
}

// TestRevokeUserAllSessions 测试撤销用户所有 Session
func TestRevokeUserAllSessions(t *testing.T) {
	svc, _ := getJwtSvr()
	ctx := context.Background()

	// 模拟多设备登录（生成 3 个 Session）
	tp1, err := svc.GenerateTokenPair(ctx, 1, "user", "email", WithDeviceID("device-1"))
	assert.NoError(t, err)
	tp2, err := svc.GenerateTokenPair(ctx, 1, "user", "email", WithDeviceID("device-2"))
	assert.NoError(t, err)
	tp3, err := svc.GenerateTokenPair(ctx, 1, "user", "email", WithDeviceID("device-3"))
	assert.NoError(t, err)

	// 验证所有 Token 都可用
	_, err = svc.ParseAccessToken(ctx, tp1.AccessToken)
	assert.NoError(t, err)
	_, err = svc.ParseAccessToken(ctx, tp2.AccessToken)
	assert.NoError(t, err)
	_, err = svc.ParseAccessToken(ctx, tp3.AccessToken)
	assert.NoError(t, err)

	// 撤销用户所有 Session
	err = svc.RevokeUserAllSessions(ctx, 1)
	assert.NoError(t, err, "撤销所有 Session 应该成功")

	// 验证所有 Token 都不可用
	_, err = svc.ParseAccessToken(ctx, tp1.AccessToken)
	assert.Equal(t, ErrSessionInvalid, err)
	_, err = svc.ParseAccessToken(ctx, tp2.AccessToken)
	assert.Equal(t, ErrSessionInvalid, err)
	_, err = svc.ParseAccessToken(ctx, tp3.AccessToken)
	assert.Equal(t, ErrSessionInvalid, err)

	t.Logf("✅ 批量撤销 Session 测试通过（3 个设备全部撤销）")
}

// TestInvalidToken 测试无效 Token
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
			assert.Error(t, err, "无效 Token 应该解析失败")
			t.Logf("✅ %s: %v", tt.name, err)
		})
	}
}

// TestTokenOptions 测试 Token 选项
func TestTokenOptions(t *testing.T) {
	svc, _ := getJwtSvr()
	ctx := context.Background()

	// 使用自定义选项生成 Token
	tp, err := svc.GenerateTokenPair(
		ctx,
		1,
		"user",
		"email",
		WithDeviceID("test-device"),
	)
	assert.NoError(t, err)

	// 解析并验证
	claims, err := svc.ParseAccessToken(ctx, tp.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, "test-device", claims.DeviceID, "DeviceID 应该匹配")

	t.Logf("✅ Token 选项测试通过：DeviceID=%s", claims.DeviceID)
}

// BenchmarkGenerateTokenPair 性能测试：生成 Token Pair
func BenchmarkGenerateTokenPair(b *testing.B) {
	svc, _ := getJwtSvr()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GenerateTokenPair(ctx, uint(i), "user", "email")
	}
}

// BenchmarkParseAccessToken 性能测试：解析 Access Token
func BenchmarkParseAccessToken(b *testing.B) {
	svc, _ := getJwtSvr()
	ctx := context.Background()

	tp, _ := svc.GenerateTokenPair(ctx, 1, "user", "email")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.ParseAccessToken(ctx, tp.AccessToken)
	}
}

// BenchmarkRefreshToken 性能测试：刷新 Token
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
