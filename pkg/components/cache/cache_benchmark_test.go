package cache

import (
	"context"
	"fmt"
	"os"
	"testing"

	redis2 "gin-admin/pkg/components/redis"
)

/*
 * @Author: zouyx
 * @Email: 1003941268@qq.com
 * @Date:   2025/12/04
 * @Package: 缓存性能测试
 */

// ==================== MemoryCache 性能测试 ====================

// BenchmarkMemoryCache_Set 性能测试：内存缓存设置
func BenchmarkMemoryCache_Set(b *testing.B) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(ctx, key, "value", 0)
	}
}

// BenchmarkMemoryCache_Get 性能测试：内存缓存获取
func BenchmarkMemoryCache_Get(b *testing.B) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	// 预设数据
	_ = cache.Set(ctx, "test_key", "test_value", 0)

	var result string
	b.ResetTimer()
	for b.Loop() {
		_ = cache.Get(ctx, "test_key", &result)
	}
}

// BenchmarkMemoryCache_SetGet 性能测试：内存缓存混合操作
func BenchmarkMemoryCache_SetGet(b *testing.B) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i%100)
		if i%2 == 0 {
			cache.Set(ctx, key, i, 0)
		} else {
			var result int
			cache.Get(ctx, key, &result)
		}
	}
}

// BenchmarkMemoryCache_Delete 性能测试：内存缓存删除
func BenchmarkMemoryCache_Delete(b *testing.B) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	// 预设数据
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Set(ctx, key, "value", 0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Delete(ctx, key)
	}
}

// BenchmarkMemoryCache_Exists 性能测试：内存缓存存在性检查
func BenchmarkMemoryCache_Exists(b *testing.B) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	// 预设数据
	cache.Set(ctx, "test_key", "test_value", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Exists(ctx, "test_key")
	}
}

// BenchmarkMemoryCache_SAdd 性能测试：内存缓存集合添加
func BenchmarkMemoryCache_SAdd(b *testing.B) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("set_%d", i%100)
		cache.SAdd(ctx, key, fmt.Sprintf("member_%d", i))
	}
}

// ==================== RedisCache 性能测试 ====================

// skipIfNoRedisBench Redis 不可用则跳过性能测试
func skipIfNoRedisBench(b *testing.B) ICache {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	password := os.Getenv("REDIS_PASSWORD")
	if password == "" {
		password = "123456"
	}

	cfg := redis2.Config{
		Host:     host,
		Port:     6379,
		Password: password,
		DB:       0,
		PoolSize: 10,
	}

	client, err := redis2.NewClient(cfg)
	if err != nil {
		b.Skipf("Redis 连接失败，跳过测试: %v", err)
	}

	cache, err := NewRedisCache(client)
	if err != nil {
		b.Skipf("Redis 不可用，跳过测试: %v", err)
	}

	return cache
}

// BenchmarkRedisCache_Set 性能测试：Redis 缓存设置
func BenchmarkRedisCache_Set(b *testing.B) {
	cache := skipIfNoRedisBench(b)
	defer cache.Close()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		cache.Set(ctx, key, "value", 0)
	}
}

// BenchmarkRedisCache_Get 性能测试：Redis 缓存获取
func BenchmarkRedisCache_Get(b *testing.B) {
	cache := skipIfNoRedisBench(b)
	defer cache.Close()
	ctx := context.Background()

	// 预设数据
	cache.Set(ctx, "bench_test_key", "test_value", 0)
	defer cache.Delete(ctx, "bench_test_key")

	var result string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(ctx, "bench_test_key", &result)
	}
}

// BenchmarkRedisCache_SetGet 性能测试：Redis 缓存混合操作
func BenchmarkRedisCache_SetGet(b *testing.B) {
	cache := skipIfNoRedisBench(b)
	defer cache.Close()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i%100)
		if i%2 == 0 {
			cache.Set(ctx, key, i, 0)
		} else {
			var result int
			cache.Get(ctx, key, &result)
		}
	}
}

// BenchmarkRedisCache_Delete 性能测试：Redis 缓存删除
func BenchmarkRedisCache_Delete(b *testing.B) {
	cache := skipIfNoRedisBench(b)
	defer cache.Close()
	ctx := context.Background()

	// 预设数据
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		cache.Set(ctx, key, "value", 0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		cache.Delete(ctx, key)
	}
}

// BenchmarkRedisCache_Exists 性能测试：Redis 缓存存在性检查
func BenchmarkRedisCache_Exists(b *testing.B) {
	cache := skipIfNoRedisBench(b)
	defer cache.Close()
	ctx := context.Background()

	// 预设数据
	cache.Set(ctx, "bench_test_key", "test_value", 0)
	defer cache.Delete(ctx, "bench_test_key")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Exists(ctx, "bench_test_key")
	}
}

// BenchmarkRedisCache_SAdd 性能测试：Redis 缓存集合添加
func BenchmarkRedisCache_SAdd(b *testing.B) {
	cache := skipIfNoRedisBench(b)
	defer cache.Close()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_set_%d", i%100)
		cache.SAdd(ctx, key, fmt.Sprintf("member_%d", i))
	}

	// 清理
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("bench_set_%d", i)
		cache.Delete(ctx, key)
	}
}

// ==================== 对比性能测试 ====================

// BenchmarkCompare_MemoryVsRedis_Set 对比：内存 vs Redis 设置操作
func BenchmarkCompare_MemoryVsRedis_Set(b *testing.B) {
	b.Run("Memory", func(b *testing.B) {
		cache := NewMemoryCache()
		defer cache.Close()
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i)
			cache.Set(ctx, key, "value", 0)
		}
	})

	b.Run("Redis", func(b *testing.B) {
		cache := skipIfNoRedisBench(b)
		defer cache.Close()
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i)
			cache.Set(ctx, key, "value", 0)
		}
	})
}

// BenchmarkCompare_MemoryVsRedis_Get 对比：内存 vs Redis 获取操作
func BenchmarkCompare_MemoryVsRedis_Get(b *testing.B) {
	b.Run("Memory", func(b *testing.B) {
		cache := NewMemoryCache()
		defer cache.Close()
		ctx := context.Background()
		cache.Set(ctx, "test_key", "test_value", 0)

		var result string
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Get(ctx, "test_key", &result)
		}
	})

	b.Run("Redis", func(b *testing.B) {
		cache := skipIfNoRedisBench(b)
		defer cache.Close()
		ctx := context.Background()
		cache.Set(ctx, "test_key", "test_value", 0)
		defer cache.Delete(ctx, "test_key")

		var result string
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Get(ctx, "test_key", &result)
		}
	})
}
