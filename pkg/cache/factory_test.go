package cache

import (
	"testing"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/04
* @Package: 缓存工厂测试
 */

// TestMustInitCache_WithNilConfig 测试未配置缓存时使用默认内存缓存
func TestMustInitCache_WithNilConfig(t *testing.T) {
	// 测试：没有配置缓存
	MustInitCache(nil)

	// 验证：应该使用内存缓存
	if GlobalCache == nil {
		t.Fatal("GlobalCache 不应该为 nil")
	}

	if GlobalCache.Type() != "memory" {
		t.Fatalf("期望缓存类型为 memory，实际为: %s", GlobalCache.Type())
	}

	t.Log("✅ 未配置缓存时，成功初始化为内存缓存")
}

// TestMustInitCache_WithMemoryConfig 测试配置内存缓存
func TestMustInitCache_WithMemoryConfig(t *testing.T) {
	cfg := &CacheConfig{
		Type: "memory",
	}

	MustInitCache(cfg)

	if GlobalCache == nil {
		t.Fatal("GlobalCache 不应该为 nil")
	}

	if GlobalCache.Type() != "memory" {
		t.Fatalf("期望缓存类型为 memory，实际为: %s", GlobalCache.Type())
	}

	t.Log("✅ 配置内存缓存成功")
}

// TestMustInitCache_WithInvalidConfig 测试配置错误时自动降级
func TestMustInitCache_WithInvalidConfig(t *testing.T) {
	// 测试：配置Redis但缺少必要参数
	cfg := &CacheConfig{
		Type:  "redis",
		Redis: nil, // 缺少Redis配置
	}

	MustInitCache(cfg)

	// 验证：应该自动降级到内存缓存
	if GlobalCache == nil {
		t.Fatal("GlobalCache 不应该为 nil")
	}

	if GlobalCache.Type() != "memory" {
		t.Fatalf("期望降级为 memory，实际为: %s", GlobalCache.Type())
	}

	t.Log("✅ 配置错误时，成功降级到内存缓存")
}

// TestMustInitCache_WithEmptyType 测试空类型时使用内存缓存
func TestMustInitCache_WithEmptyType(t *testing.T) {
	cfg := &CacheConfig{
		Type: "", // 空类型
	}

	MustInitCache(cfg)

	if GlobalCache == nil {
		t.Fatal("GlobalCache 不应该为 nil")
	}

	if GlobalCache.Type() != "memory" {
		t.Fatalf("期望缓存类型为 memory，实际为: %s", GlobalCache.Type())
	}

	t.Log("✅ 空类型时，成功初始化为内存缓存")
}

// TestGetType 测试获取缓存类型
func TestGetType(t *testing.T) {
	// 先初始化为内存缓存
	MustInitCache(nil)

	cacheType := GetType()
	if cacheType != "memory" {
		t.Fatalf("期望缓存类型为 memory，实际为: %s", cacheType)
	}

	t.Log("✅ GetType 返回正确")
}

// TestIsEnabled 测试缓存是否启用
func TestIsEnabled(t *testing.T) {
	// 先初始化
	MustInitCache(nil)

	if !IsEnabled() {
		t.Fatal("缓存应该已启用")
	}

	if !IsAvailable() {
		t.Fatal("缓存应该可用")
	}

	t.Log("✅ 缓存状态检查正确")
}

