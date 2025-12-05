package cache

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*
 * @Author: zouyx
 * @Email: 1003941268@qq.com
 * @Date:   2025/12/04
 * @Package: 分片内存缓存测试
 */

// ==============================================================================
// 基础功能测试
// ==============================================================================

func TestShardedCache_SetAndGet(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	// 测试字符串
	key := "test_key"
	value := "test_value"
	err := cache.Set(ctx, key, value, 0)
	assert.NoError(t, err)

	var result string
	err = cache.Get(ctx, key, &result)
	assert.NoError(t, err)
	assert.Equal(t, value, result)
}

func TestShardedCache_SetAndGetInt64(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	key := "int_key"
	value := int64(12345)
	err := cache.Set(ctx, key, value, 0)
	assert.NoError(t, err)

	var result int64
	err = cache.Get(ctx, key, &result)
	assert.NoError(t, err)
	assert.Equal(t, value, result)
}

func TestShardedCache_SetWithTTL(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	key := "ttl_key"
	value := "ttl_value"
	ttl := 100 * time.Millisecond

	err := cache.Set(ctx, key, value, ttl)
	assert.NoError(t, err)

	// 立即获取应该成功
	var result string
	err = cache.Get(ctx, key, &result)
	assert.NoError(t, err)

	// 等待过期
	time.Sleep(150 * time.Millisecond)
	err = cache.Get(ctx, key, &result)
	assert.Error(t, err)
	assert.Equal(t, ErrKeyNotFound, err)
}

func TestShardedCache_GetNonExistent(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	var result string
	err := cache.Get(ctx, "non_existent_key", &result)
	assert.Error(t, err)
	assert.Equal(t, ErrKeyNotFound, err)
}

func TestShardedCache_Delete(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	key := "delete_key"
	cache.Set(ctx, key, "value", 0)

	err := cache.Delete(ctx, key)
	assert.NoError(t, err)

	var result string
	err = cache.Get(ctx, key, &result)
	assert.Error(t, err)
}

func TestShardedCache_DeleteMultiple(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		cache.Set(ctx, key, "value", 0)
	}

	err := cache.Delete(ctx, keys...)
	assert.NoError(t, err)

	for _, key := range keys {
		var result string
		err = cache.Get(ctx, key, &result)
		assert.Error(t, err)
	}
}

func TestShardedCache_DeletePrefix(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	// 设置多个带前缀的 key
	cache.Set(ctx, "user:1", "data1", 0)
	cache.Set(ctx, "user:2", "data2", 0)
	cache.Set(ctx, "product:1", "data3", 0)

	// 删除 user: 前缀的所有 key
	err := cache.DeletePrefix(ctx, "user:")
	assert.NoError(t, err)

	// 验证 user: 前缀的 key 被删除
	var result string
	err = cache.Get(ctx, "user:1", &result)
	assert.Error(t, err)
	err = cache.Get(ctx, "user:2", &result)
	assert.Error(t, err)

	// product: 前缀的 key 应该还在
	err = cache.Get(ctx, "product:1", &result)
	assert.NoError(t, err)
}

func TestShardedCache_Exists(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	key := "exists_key"

	exists, err := cache.Exists(ctx, key)
	assert.NoError(t, err)
	assert.False(t, exists)

	cache.Set(ctx, key, "value", 0)
	exists, err = cache.Exists(ctx, key)
	assert.NoError(t, err)
	assert.True(t, exists)
}

// ==============================================================================
// 集合操作测试
// ==============================================================================

func TestShardedCache_SAdd(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	key := "set_key"
	members := []interface{}{"member1", "member2", "member3"}

	err := cache.SAdd(ctx, key, members...)
	assert.NoError(t, err)

	// 验证成员是否添加成功
	for _, member := range members {
		isMember, err := cache.SIsMember(ctx, key, member)
		assert.NoError(t, err)
		assert.True(t, isMember)
	}
}

func TestShardedCache_SAdd_TypeConflict(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	key := "conflict_key"
	
	// 先设置为普通值
	cache.Set(ctx, key, "value", 0)
	
	// 尝试作为Set使用应该报错
	err := cache.SAdd(ctx, key, "member1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a set")
}

func TestShardedCache_SIsMember(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	key := "set_key"
	member := "member1"

	isMember, err := cache.SIsMember(ctx, key, member)
	assert.NoError(t, err)
	assert.False(t, isMember)

	cache.SAdd(ctx, key, member)
	isMember, err = cache.SIsMember(ctx, key, member)
	assert.NoError(t, err)
	assert.True(t, isMember)
}

func TestShardedCache_SMembers(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	key := "set_key"
	expectedMembers := []interface{}{"member1", "member2", "member3"}

	cache.SAdd(ctx, key, expectedMembers...)

	members, err := cache.SMembers(ctx, key)
	assert.NoError(t, err)
	assert.Len(t, members, len(expectedMembers))

	memberMap := make(map[string]bool)
	for _, m := range members {
		memberMap[m] = true
	}

	for _, expected := range expectedMembers {
		assert.True(t, memberMap[expected.(string)])
	}
}

func TestShardedCache_SRem(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	key := "set_key"
	cache.SAdd(ctx, key, "member1", "member2", "member3")

	err := cache.SRem(ctx, key, "member2")
	assert.NoError(t, err)

	isMember, _ := cache.SIsMember(ctx, key, "member2")
	assert.False(t, isMember)

	isMember, _ = cache.SIsMember(ctx, key, "member1")
	assert.True(t, isMember)
}

func TestShardedCache_SRem_EmptySetCleanup(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	key := "set_key"
	cache.SAdd(ctx, key, "member1")
	
	// 删除最后一个成员，应该清理整个item
	err := cache.SRem(ctx, key, "member1")
	assert.NoError(t, err)
	
	// key 应该不存在了
	exists, _ := cache.Exists(ctx, key)
	assert.False(t, exists)
}

// ==============================================================================
// 计数器测试
// ==============================================================================

func TestShardedCache_Incr(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	key := "counter"

	val, err := cache.Incr(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	val, err = cache.Incr(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), val)
}

func TestShardedCache_Decr(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	key := "counter"

	val, err := cache.Decr(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(-1), val)

	val, err = cache.Decr(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(-2), val)
}

// ==============================================================================
// TTL 测试
// ==============================================================================

func TestShardedCache_Expire(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	key := "expire_key"
	cache.Set(ctx, key, "value", 0)

	ttl := 100 * time.Millisecond
	err := cache.Expire(ctx, key, ttl)
	assert.NoError(t, err)

	var result string
	err = cache.Get(ctx, key, &result)
	assert.NoError(t, err)

	time.Sleep(150 * time.Millisecond)
	err = cache.Get(ctx, key, &result)
	assert.Error(t, err)
}

func TestShardedCache_TTL(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	key := "ttl_key"
	ttl := 10 * time.Second

	cache.Set(ctx, key, "value", ttl)

	remaining, err := cache.TTL(ctx, key)
	assert.NoError(t, err)
	assert.Greater(t, remaining, 9*time.Second)
	assert.LessOrEqual(t, remaining, ttl)
}

func TestShardedCache_TTLNeverExpire(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	key := "never_expire"
	cache.Set(ctx, key, "value", 0)

	remaining, err := cache.TTL(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(-1), remaining)
}

// ==============================================================================
// 分片特性测试
// ==============================================================================

func TestShardedCache_DifferentShardCounts(t *testing.T) {
	ctx := context.Background()
	testCases := []int{8, 16, 32, 64}

	for _, shardCount := range testCases {
		t.Run(fmt.Sprintf("ShardCount_%d", shardCount), func(t *testing.T) {
			cache := NewShardedMemoryCache(shardCount)
			defer cache.Close()

			// 写入大量数据
			for i := 0; i < 1000; i++ {
				key := fmt.Sprintf("key_%d", i)
				err := cache.Set(ctx, key, i, 0)
				assert.NoError(t, err)
			}

			// 验证数据
			for i := 0; i < 1000; i++ {
				key := fmt.Sprintf("key_%d", i)
				var result int
				err := cache.Get(ctx, key, &result)
				assert.NoError(t, err)
				assert.Equal(t, i, result)
			}
		})
	}
}

func TestShardedCache_HashCollision(t *testing.T) {
	cache := NewShardedMemoryCache(8)
	defer cache.Close()
	ctx := context.Background()

	// 写入足够多的数据，确保有些会落在同一个hash bucket
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("collision_test_%d", i)
		cache.Set(ctx, key, i, 0)
	}

	// 验证所有数据都能正确读取（测试哈希冲突处理）
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("collision_test_%d", i)
		var result int
		err := cache.Get(ctx, key, &result)
		assert.NoError(t, err, "Key %s should exist", key)
		assert.Equal(t, i, result, "Value mismatch for key %s", key)
	}
}

// ==============================================================================
// 并发安全测试
// ==============================================================================

func TestShardedCache_ConcurrentReadWrite(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	const goroutines = 100
	const iterations = 100
	var wg sync.WaitGroup

	// 并发写入
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(index int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("key_%d_%d", index, j)
				err := cache.Set(ctx, key, index*iterations+j, 0)
				assert.NoError(t, err)
			}
		}(i)
	}
	wg.Wait()

	// 并发读取
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(index int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("key_%d_%d", index, j)
				var result int
				err := cache.Get(ctx, key, &result)
				assert.NoError(t, err)
				assert.Equal(t, index*iterations+j, result)
			}
		}(i)
	}
	wg.Wait()
}

func TestShardedCache_ConcurrentIncr(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	const goroutines = 100
	key := "concurrent_counter"
	var wg sync.WaitGroup

	// 并发自增
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, err := cache.Incr(ctx, key)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

	// 验证最终值
	var result int64
	cache.Get(ctx, key, &result)
	assert.Equal(t, int64(goroutines), result)
}

func TestShardedCache_ConcurrentMixedOperations(t *testing.T) {
	cache := NewShardedMemoryCache(32)
	defer cache.Close()
	ctx := context.Background()

	const goroutines = 50
	const operations = 100
	var wg sync.WaitGroup
	var setOps, getOps, delOps atomic.Int64

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(index int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := fmt.Sprintf("key_%d", j%10)
				
				switch j % 3 {
				case 0:
					cache.Set(ctx, key, j, 0)
					setOps.Add(1)
				case 1:
					var val int
					cache.Get(ctx, key, &val)
					getOps.Add(1)
				case 2:
					cache.Delete(ctx, key)
					delOps.Add(1)
				}
			}
		}(i)
	}
	wg.Wait()

	// 验证所有操作都完成
	t.Logf("Set: %d, Get: %d, Delete: %d", setOps.Load(), getOps.Load(), delOps.Load())
	assert.Greater(t, setOps.Load(), int64(0))
	assert.Greater(t, getOps.Load(), int64(0))
	assert.Greater(t, delOps.Load(), int64(0))
}

// ==============================================================================
// Pipeline 测试
// ==============================================================================

func TestShardedCache_Pipeline(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	cache.Set(ctx, "key1", "value1", 0)
	cache.SAdd(ctx, "set1", "member1")

	pipe := cache.Pipeline()
	setCmd := pipe.Set(ctx, "key2", "value2", time.Second)
	existsCmd := pipe.Exists(ctx, "key1")
	isMemberCmd := pipe.SIsMember(ctx, "set1", "member1")

	err := pipe.Exec(ctx)
	assert.NoError(t, err)

	status, err := setCmd.Result()
	assert.NoError(t, err)
	assert.Equal(t, "OK", status)

	exists, err := existsCmd.Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	isMember, err := isMemberCmd.Result()
	assert.NoError(t, err)
	assert.True(t, isMember)
}

// ==============================================================================
// 复杂类型测试
// ==============================================================================

func TestShardedCache_ComplexStruct(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	type User struct {
		ID    int
		Name  string
		Email string
		Tags  []string
	}

	key := "user:1"
	user := User{
		ID:    1,
		Name:  "Alice",
		Email: "alice@example.com",
		Tags:  []string{"admin", "premium"},
	}

	err := cache.Set(ctx, key, user, 0)
	assert.NoError(t, err)

	var result User
	err = cache.Get(ctx, key, &result)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.Name, result.Name)
	assert.Equal(t, user.Email, result.Email)
	assert.Equal(t, user.Tags, result.Tags)
}

// ==============================================================================
// 边界情况测试
// ==============================================================================

func TestShardedCache_EdgeCases(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	// 空字符串作为 key
	err := cache.Set(ctx, "", "value", 0)
	assert.NoError(t, err)

	var result string
	err = cache.Get(ctx, "", &result)
	assert.NoError(t, err)
	assert.Equal(t, "value", result)

	// 空字符串作为 value
	err = cache.Set(ctx, "key", "", 0)
	assert.NoError(t, err)

	err = cache.Get(ctx, "key", &result)
	assert.NoError(t, err)
	assert.Equal(t, "", result)

	// 删除空 key 列表
	err = cache.Delete(ctx)
	assert.NoError(t, err)

	// 空前缀删除
	err = cache.DeletePrefix(ctx, "")
	assert.NoError(t, err)
}

func TestShardedCache_Ping(t *testing.T) {
	cache := NewShardedMemoryCache(0)
	defer cache.Close()
	ctx := context.Background()

	err := cache.Ping(ctx)
	assert.NoError(t, err)
}
