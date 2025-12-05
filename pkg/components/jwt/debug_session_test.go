package jwt

import (
	"context"
	cache2 "gin-admin/pkg/components/cache"
	"testing"
	"time"
)

// TestDebugSessionSave 详细调试Session保存过程
func TestDebugSessionSave(t *testing.T) {
	cache := cache2.NewShardedMemoryCache(0)
	ctx := context.Background()

	cfg := Config{
		Secret:             "test-secret-key-32-chars-minimum",
		Issuer:             "test",
		AccessTokenExpire:  time.Second * 10,
		RefreshTokenExpire: time.Hour,
	}
	svc := NewJwtService(cfg, cache)

	// 步骤1: 生成Token
	t.Log("步骤1: 生成Token Pair...")
	tp, err := svc.GenerateTokenPair(ctx, 1, "testuser", "test@example.com")
	if err != nil {
		t.Fatalf("生成Token失败: %v", err)
	}
	t.Logf("✓ Token生成成功: AccessToken长度=%d", len(tp.AccessToken))

	// 步骤2: 立即解析Token获取SessionID
	t.Log("步骤2: 解析Access Token获取SessionID...")
	claims, err := svc.ParseAccessToken(ctx, tp.AccessToken)
	if err != nil {
		t.Fatalf("❌ 解析Token失败: %v", err)
	}
	t.Logf("✓ Token解析成功: SessionID=%s, UserID=%d", claims.SessionID, claims.UserID)

	// 步骤3: 直接检查缓存中的Session
	t.Log("步骤3: 检查缓存中的Session...")
	sessionKey := "jwt:session:" + claims.SessionID

	var sessionInfo SessionInfo
	err = cache.Get(ctx, sessionKey, &sessionInfo)
	if err != nil {
		t.Errorf("❌ 从缓存读取Session失败: %v", err)
		t.Logf("Session Key: %s", sessionKey)

		// 尝试检查key是否存在
		exists, _ := cache.Exists(ctx, sessionKey)
		t.Logf("Session Key exists: %v", exists)
	} else {
		t.Logf("✓ Session读取成功: UserID=%d, ExpiresAt=%v",
			sessionInfo.UserID, sessionInfo.ExpiresAt)
	}

	// 步骤4: 检查用户所有sessions set
	t.Log("步骤4: 检查用户Sessions集合...")
	userSessionsKey := "jwt:user_sessions:1"
	members, err := cache.SMembers(ctx, userSessionsKey)
	if err != nil {
		t.Errorf("❌ 读取用户Sessions集合失败: %v", err)
	} else {
		t.Logf("✓ 用户Sessions集合成员数: %d", len(members))
		for _, member := range members {
			t.Logf("  - Session: %s", member)
		}
	}

	// 步骤5: 再次解析Token验证
	t.Log("步骤5: 再次解析Token验证...")
	claims2, err := svc.ParseAccessToken(ctx, tp.AccessToken)
	if err != nil {
		t.Fatalf("❌ 第二次解析Token失败: %v", err)
	}
	t.Logf("✓ 第二次解析成功: SessionID=%s", claims2.SessionID)
}

// TestDebugPipelineSet 测试Pipeline Set是否正常工作
func TestDebugPipelineSet(t *testing.T) {
	cache := cache2.NewShardedMemoryCache(0)
	ctx := context.Background()

	// 测试直接Set结构体
	t.Log("测试1: 直接Set结构体...")
	testData := SessionInfo{
		SessionID: "test-session-123",
		UserID:    999,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	err := cache.Set(ctx, "direct_set_key", testData, time.Hour)
	if err != nil {
		t.Errorf("❌ 直接Set失败: %v", err)
	} else {
		var retrieved SessionInfo
		err = cache.Get(ctx, "direct_set_key", &retrieved)
		if err != nil {
			t.Errorf("❌ 直接Get失败: %v", err)
		} else {
			t.Logf("✓ 直接Set/Get成功: SessionID=%s, UserID=%d",
				retrieved.SessionID, retrieved.UserID)
		}
	}

	// 测试Pipeline Set结构体
	t.Log("测试2: Pipeline Set结构体...")
	pipe := cache.Pipeline()
	pipe.Set(ctx, "pipeline_set_key", testData, time.Hour)
	pipe.SAdd(ctx, "test", testData.SessionID)
	pipe.Expire(ctx, "test", time.Hour)
	err = pipe.Exec(ctx)
	if err != nil {
		t.Errorf("❌ Pipeline Exec失败: %v", err)
	} else {
		t.Log("✓ Pipeline Exec成功")

		var retrieved SessionInfo
		err = cache.Get(ctx, "pipeline_set_key", &retrieved)
		if err != nil {
			t.Errorf("❌ Pipeline后Get失败: %v", err)
		} else {
			t.Logf("✓ Pipeline Set/Get成功: SessionID=%s, UserID=%d",
				retrieved.SessionID, retrieved.UserID)
		}
	}
}

// TestDebugRefreshTokenFlow 模拟 RefreshToken 流程调试
func TestDebugRefreshTokenFlow(t *testing.T) {
	cache := cache2.NewShardedMemoryCache(0)
	ctx := context.Background()

	cfg := Config{
		Secret:             "test-secret-key-32-chars-minimum",
		Issuer:             "test",
		AccessTokenExpire:  time.Second * 10,
		RefreshTokenExpire: time.Hour,
	}
	svc := NewJwtService(cfg, cache)

	// 1. 生成 Token
	t.Log("1. 生成 Token Pair...")
	tp, err := svc.GenerateTokenPair(ctx, 1, "user", "email")
	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}

	// 2. 解析 Refresh Token 获取 SessionID
	t.Log("2. 解析 Refresh Token...")
	claims, err := svc.ParseAccessToken(ctx, tp.AccessToken) // 注意：这里用 AccessToken 解析是为了获取 SessionID 验证
	if err != nil {
		t.Fatalf("解析 AccessToken 失败: %v", err)
	}
	sessionID := claims.SessionID
	t.Logf("SessionID: %s", sessionID)

	// 3. 直接检查缓存
	t.Log("3. 检查缓存中的 Session...")
	var s SessionInfo
	err = cache.Get(ctx, "jwt:session:"+sessionID, &s)
	if err != nil {
		t.Errorf("❌ 缓存中找不到 Session: %v", err)
	} else {
		t.Logf("✓ 缓存中存在 Session: UserID=%d", s.UserID)
	}

	// 4. 执行 RefreshToken
	t.Log("4. 执行 RefreshToken...")
	tp2, err := svc.RefreshToken(ctx, tp.RefreshToken)
	if err != nil {
		t.Errorf("❌ RefreshToken 失败: %v", err)
	} else {
		t.Log("✓ RefreshToken 成功")
		t.Logf("New AccessToken: %s...", tp2.AccessToken[:10])
	}
}
