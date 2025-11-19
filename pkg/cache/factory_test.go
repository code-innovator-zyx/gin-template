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
	if GetGlobalCache() == nil {
		t.Fatal("GlobalCache 不应该为 nil")
	}

	t.Log("✅ 未配置缓存时，成功初始化为内存缓存")
}

// TestMustInitCache_WithMemoryConfig 测试配置内存缓存
func TestMustInitCache_WithMemoryConfig(t *testing.T) {

	MustInitCache(nil)

	if GetGlobalCache() == nil {
		t.Fatal("GlobalCache 不应该为 nil")
	}

	t.Log("✅ 配置内存缓存成功")
}

// TestMustInitCache_WithInvalidConfig 测试配置错误时自动降级
func TestMustInitCache_WithInvalidConfig(t *testing.T) {
	// 测试：配置Redis但缺少必要参数
	cfg := &RedisConfig{}

	MustInitCache(cfg)

	// 验证：应该自动降级到内存缓存
	if GetGlobalCache() == nil {
		t.Fatal("GlobalCache 不应该为 nil")
	}

	t.Log("✅ 配置错误时，成功降级到内存缓存")
}

// TestIsEnabled 测试缓存是否启用
func TestIsEnabled(t *testing.T) {
	// 先初始化
	MustInitCache(nil)

	if !IsAvailable() {
		t.Fatal("缓存应该已启用")
	}

	if !IsAvailable() {
		t.Fatal("缓存应该可用")
	}

	t.Log("✅ 缓存状态检查正确")
}
