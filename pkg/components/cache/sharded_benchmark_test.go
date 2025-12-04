package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

/*
 * @Author: zouyx
 * @Email: 1003941268@qq.com
 * @Date:   2025/12/04
 * @Package: 分片内存缓存 vs 普通内存缓存性能对比测试
 */

// ==================== 单锁 vs 分片锁性能对比 ====================

// BenchmarkCompare_Single_vs_Sharded_Set 对比：单锁 vs 分片锁 - Set操作
func BenchmarkCompare_Single_vs_Sharded_Set(b *testing.B) {
	b.Run("SingleLock", func(b *testing.B) {
		cache := NewMemoryCache()
		defer cache.Close()
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i)
			cache.Set(ctx, key, "value", 0)
		}
	})

	b.Run("Sharded_8", func(b *testing.B) {
		cache := NewShardedMemoryCache(8)
		defer cache.Close()
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i)
			cache.Set(ctx, key, "value", 0)
		}
	})

	b.Run("Sharded_16", func(b *testing.B) {
		cache := NewShardedMemoryCache(16)
		defer cache.Close()
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i)
			cache.Set(ctx, key, "value", 0)
		}
	})

	b.Run("Sharded_32", func(b *testing.B) {
		cache := NewShardedMemoryCache(32)
		defer cache.Close()
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i)
			cache.Set(ctx, key, "value", 0)
		}
	})

	b.Run("Sharded_64", func(b *testing.B) {
		cache := NewShardedMemoryCache(64)
		defer cache.Close()
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i)
			cache.Set(ctx, key, "value", 0)
		}
	})
}

// BenchmarkCompare_Single_vs_Sharded_Get 对比：单锁 vs 分片锁 - Get操作
func BenchmarkCompare_Single_vs_Sharded_Get(b *testing.B) {
	b.Run("SingleLock", func(b *testing.B) {
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

	b.Run("Sharded_32", func(b *testing.B) {
		cache := NewShardedMemoryCache(32)
		defer cache.Close()
		ctx := context.Background()
		cache.Set(ctx, "test_key", "test_value", 0)

		var result string
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Get(ctx, "test_key", &result)
		}
	})
}

// ==================== 并发读写性能测试 ====================

// BenchmarkConcurrent_Single_vs_Sharded_Read 对比：并发读性能
func BenchmarkConcurrent_Single_vs_Sharded_Read(b *testing.B) {
	const goroutines = 100

	b.Run("SingleLock", func(b *testing.B) {
		cache := NewMemoryCache()
		defer cache.Close()
		ctx := context.Background()

		// 预设数据
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key_%d", i)
			cache.Set(ctx, key, fmt.Sprintf("value_%d", i), 0)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			var result string
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key_%d", i%100)
				cache.Get(ctx, key, &result)
				i++
			}
		})
	})

	b.Run("Sharded_32", func(b *testing.B) {
		cache := NewShardedMemoryCache(32)
		defer cache.Close()
		ctx := context.Background()

		// 预设数据
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key_%d", i)
			cache.Set(ctx, key, fmt.Sprintf("value_%d", i), 0)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			var result string
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key_%d", i%100)
				cache.Get(ctx, key, &result)
				i++
			}
		})
	})
}

// BenchmarkConcurrent_Single_vs_Sharded_Write 对比：并发写性能
func BenchmarkConcurrent_Single_vs_Sharded_Write(b *testing.B) {
	b.Run("SingleLock", func(b *testing.B) {
		cache := NewMemoryCache()
		defer cache.Close()
		ctx := context.Background()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key_%d", i)
				cache.Set(ctx, key, i, 0)
				i++
			}
		})
	})

	b.Run("Sharded_32", func(b *testing.B) {
		cache := NewShardedMemoryCache(32)
		defer cache.Close()
		ctx := context.Background()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key_%d", i)
				cache.Set(ctx, key, i, 0)
				i++
			}
		})
	})
}

// BenchmarkConcurrent_Single_vs_Sharded_Mixed 对比：并发混合读写
func BenchmarkConcurrent_Single_vs_Sharded_Mixed(b *testing.B) {
	b.Run("SingleLock", func(b *testing.B) {
		cache := NewMemoryCache()
		defer cache.Close()
		ctx := context.Background()

		// 预设数据
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key_%d", i)
			cache.Set(ctx, key, fmt.Sprintf("value_%d", i), 0)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			var result string
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key_%d", i%100)
				if i%3 == 0 {
					cache.Set(ctx, key, i, 0) // 33% 写操作
				} else {
					cache.Get(ctx, key, &result) // 67% 读操作
				}
				i++
			}
		})
	})

	b.Run("Sharded_32", func(b *testing.B) {
		cache := NewShardedMemoryCache(32)
		defer cache.Close()
		ctx := context.Background()

		// 预设数据
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key_%d", i)
			cache.Set(ctx, key, fmt.Sprintf("value_%d", i), 0)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			var result string
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key_%d", i%100)
				if i%3 == 0 {
					cache.Set(ctx, key, i, 0) // 33% 写操作
				} else {
					cache.Get(ctx, key, &result) // 67% 读操作
				}
				i++
			}
		})
	})
}

// ==================== 锁竞争压力测试 ====================

// BenchmarkLockContention_Single_vs_Sharded 对比：高并发锁竞争场景
func BenchmarkLockContention_Single_vs_Sharded(b *testing.B) {
	// 模拟热点key访问（所有goroutine访问相同的key集合）
	b.Run("SingleLock_HotKeys", func(b *testing.B) {
		cache := NewMemoryCache()
		defer cache.Close()
		ctx := context.Background()

		// 预设10个热点key
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("hot_key_%d", i)
			cache.Set(ctx, key, i, 0)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			var result int
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("hot_key_%d", i%10)
				cache.Get(ctx, key, &result)
				i++
			}
		})
	})

	b.Run("Sharded_32_HotKeys", func(b *testing.B) {
		cache := NewShardedMemoryCache(32)
		defer cache.Close()
		ctx := context.Background()

		// 预设10个热点key
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("hot_key_%d", i)
			cache.Set(ctx, key, i, 0)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			var result int
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("hot_key_%d", i%10)
				cache.Get(ctx, key, &result)
				i++
			}
		})
	})
}

// ==================== 实际并发度测试 ====================

// TestActualConcurrency 测试实际并发度
func TestActualConcurrency(t *testing.T) {
	const goroutines = 100
	const opsPerGoroutine = 10000

	t.Run("SingleLock", func(t *testing.T) {
		cache := NewMemoryCache()
		defer cache.Close()
		ctx := context.Background()

		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < opsPerGoroutine; j++ {
					key := fmt.Sprintf("key_%d_%d", id, j)
					cache.Set(ctx, key, j, 0)
				}
			}(i)
		}

		wg.Wait()
		t.Logf("SingleLock: %d goroutines × %d ops = %d total ops completed",
			goroutines, opsPerGoroutine, goroutines*opsPerGoroutine)
	})

	t.Run("Sharded_32", func(t *testing.T) {
		cache := NewShardedMemoryCache(32)
		defer cache.Close()
		ctx := context.Background()

		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < opsPerGoroutine; j++ {
					key := fmt.Sprintf("key_%d_%d", id, j)
					cache.Set(ctx, key, j, 0)
				}
			}(i)
		}

		wg.Wait()
		t.Logf("Sharded_32: %d goroutines × %d ops = %d total ops completed",
			goroutines, opsPerGoroutine, goroutines*opsPerGoroutine)
	})
}
