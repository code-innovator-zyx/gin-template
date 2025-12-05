package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*
 * @Author: zouyx
 * @Email: 1003941268@qq.com
 * @Date:   2025/11/05
 * @Package: 内存缓存测试
 */

// TestMemoryCache_SetAndGet 测试基本的设置和获取功能
func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	// 测试字符串
	key := "test_key"
	value := "test_value"
	err := cache.Set(ctx, key, value, 0)
	assert.NoError(t, err, "Set 应该成功")

	var result string
	err = cache.Get(ctx, key, &result)
	assert.NoError(t, err, "Get 应该成功")
	assert.Equal(t, value, result, "获取的值应该与设置的值相同")
}

// TestMemoryCache_SetWithTTL 测试带过期时间的缓存
func TestMemoryCache_SetWithTTL(t *testing.T) {
	cache := NewMemoryCache()
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
	assert.NoError(t, err, "TTL 未过期时应该能获取到值")

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	err = cache.Get(ctx, key, &result)
	assert.Error(t, err, "TTL 过期后应该获取失败")
	assert.Equal(t, ErrKeyNotFound, err, "错误类型应该是 ErrKeyNotFound")
}

// TestMemoryCache_GetNonExistent 测试获取不存在的 key
func TestMemoryCache_GetNonExistent(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	var result string
	err := cache.Get(ctx, "non_existent_key", &result)
	assert.Error(t, err, "获取不存在的 key 应该失败")
	assert.Equal(t, ErrKeyNotFound, err)
}

// TestMemoryCache_Delete 测试删除功能
func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "delete_key"
	value := "delete_value"

	// 设置 key
	cache.Set(ctx, key, value, 0)

	// 删除 key
	err := cache.Delete(ctx, key)
	assert.NoError(t, err, "Delete 应该成功")

	// 验证 key 已被删除
	var result string
	err = cache.Get(ctx, key, &result)
	assert.Error(t, err, "删除后应该获取失败")
}

// TestMemoryCache_DeleteMultiple 测试批量删除
func TestMemoryCache_DeleteMultiple(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		cache.Set(ctx, key, "value", 0)
	}

	// 批量删除
	err := cache.Delete(ctx, keys...)
	assert.NoError(t, err)

	// 验证所有 key 都已删除
	for _, key := range keys {
		var result string
		err = cache.Get(ctx, key, &result)
		assert.Error(t, err, "key %s 应该已被删除", key)
	}
}

// TestMemoryCache_DeletePrefix 测试前缀删除
func TestMemoryCache_DeletePrefix(t *testing.T) {
	cache := NewMemoryCache()
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

// TestMemoryCache_Exists 测试 key 是否存在
func TestMemoryCache_Exists(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "exists_key"

	// key 不存在时
	exists, err := cache.Exists(ctx, key)
	assert.NoError(t, err)
	assert.False(t, exists, "key 应该不存在")

	// 设置 key 后
	cache.Set(ctx, key, "value", 0)
	exists, err = cache.Exists(ctx, key)
	assert.NoError(t, err)
	assert.True(t, exists, "key 应该存在")
}

// TestMemoryCache_SAdd 测试集合添加
func TestMemoryCache_SAdd(t *testing.T) {
	cache := NewMemoryCache()
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
		assert.True(t, isMember, "member %v 应该在集合中", member)
	}
}

// TestMemoryCache_SIsMember 测试集合成员检查
func TestMemoryCache_SIsMember(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "set_key"
	member := "member1"

	// 成员不存在时
	isMember, err := cache.SIsMember(ctx, key, member)
	assert.NoError(t, err)
	assert.False(t, isMember, "成员应该不存在")

	// 添加成员后
	cache.SAdd(ctx, key, member)
	isMember, err = cache.SIsMember(ctx, key, member)
	assert.NoError(t, err)
	assert.True(t, isMember, "成员应该存在")
}

// TestMemoryCache_SMembers 测试获取集合所有成员
func TestMemoryCache_SMembers(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "set_key"
	expectedMembers := []interface{}{"member1", "member2", "member3"}

	cache.SAdd(ctx, key, expectedMembers...)

	members, err := cache.SMembers(ctx, key)
	assert.NoError(t, err)
	assert.Len(t, members, len(expectedMembers), "成员数量应该相同")

	// 验证所有成员都存在
	memberMap := make(map[string]bool)
	for _, m := range members {
		memberMap[m] = true
	}

	for _, expected := range expectedMembers {
		assert.True(t, memberMap[expected.(string)], "member %v 应该存在", expected)
	}
}

// TestMemoryCache_SMembersEmpty 测试空集合
func TestMemoryCache_SMembersEmpty(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	members, err := cache.SMembers(ctx, "non_existent_set")
	assert.NoError(t, err)
	assert.Empty(t, members, "空集合应该返回空数组")
}

// TestMemoryCache_SRem 测试集合删除成员
func TestMemoryCache_SRem(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "set_key"
	cache.SAdd(ctx, key, "member1", "member2", "member3")

	// 删除一个成员
	err := cache.SRem(ctx, key, "member2")
	assert.NoError(t, err)

	// 验证成员已删除
	isMember, _ := cache.SIsMember(ctx, key, "member2")
	assert.False(t, isMember, "member2 应该已被删除")

	// 其他成员应该还在
	isMember, _ = cache.SIsMember(ctx, key, "member1")
	assert.True(t, isMember, "member1 应该还在")
}

// TestMemoryCache_Expire 测试设置过期时间
func TestMemoryCache_Expire(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "expire_key"
	value := "expire_value"

	// 设置永不过期的 key
	cache.Set(ctx, key, value, 0)

	// 设置过期时间
	ttl := 100 * time.Millisecond
	err := cache.Expire(ctx, key, ttl)
	assert.NoError(t, err)

	// 立即获取应该成功
	var result string
	err = cache.Get(ctx, key, &result)
	assert.NoError(t, err)

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	err = cache.Get(ctx, key, &result)
	assert.Error(t, err, "过期后应该获取失败")
}

// TestMemoryCache_ExpireNonExistent 测试对不存在的 key 设置过期时间
func TestMemoryCache_ExpireNonExistent(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	err := cache.Expire(ctx, "non_existent_key", time.Second)
	assert.Error(t, err, "对不存在的 key 设置过期时间应该失败")
	assert.Equal(t, ErrKeyNotFound, err)
}

// TestMemoryCache_TTL 测试获取剩余过期时间
func TestMemoryCache_TTL(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "ttl_key"
	value := "ttl_value"
	ttl := 10 * time.Second

	// 设置带 TTL 的 key
	cache.Set(ctx, key, value, ttl)

	// 获取 TTL
	remaining, err := cache.TTL(ctx, key)
	assert.NoError(t, err)
	assert.Greater(t, remaining, 9*time.Second, "TTL 应该大于 9 秒")
	assert.LessOrEqual(t, remaining, ttl, "TTL 应该小于等于设置的值")
}

// TestMemoryCache_TTLNeverExpire 测试永不过期的 key
func TestMemoryCache_TTLNeverExpire(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "never_expire_key"
	value := "never_expire_value"

	// 设置永不过期的 key
	cache.Set(ctx, key, value, 0)

	// 获取 TTL
	remaining, err := cache.TTL(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(-1), remaining, "永不过期应该返回 -1")
}

// TestMemoryCache_TTLNonExistent 测试获取不存在 key 的 TTL
func TestMemoryCache_TTLNonExistent(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	_, err := cache.TTL(ctx, "non_existent_key")
	assert.Error(t, err, "获取不存在 key 的 TTL 应该失败")
	assert.Equal(t, ErrKeyNotFound, err)
}

// TestMemoryCache_Incr 测试递增计数器
func TestMemoryCache_Incr(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "counter"

	// 第一次递增（从 0 开始）
	val, err := cache.Incr(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	// 第二次递增
	val, err = cache.Incr(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), val)
}

// TestMemoryCache_Decr 测试递减计数器
func TestMemoryCache_Decr(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "counter"

	// 第一次递减（从 0 开始）
	val, err := cache.Decr(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(-1), val)

	// 第二次递减
	val, err = cache.Decr(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(-2), val)
}

// TestMemoryCache_Ping 测试连接
func TestMemoryCache_Ping(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	err := cache.Ping(ctx)
	assert.NoError(t, err, "Ping 应该成功")
}

// TestMemoryCache_ComplexStruct 测试复杂结构体
func TestMemoryCache_ComplexStruct(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	type User struct {
		ID    int
		Name  string
		Age   int
		Email string
		Tags  []string
	}

	key := "user:1"
	user := User{
		ID:    1,
		Name:  "张三",
		Age:   25,
		Email: "zhangsan@example.com",
		Tags:  []string{"admin", "developer"},
	}

	// 设置复杂结构
	err := cache.Set(ctx, key, user, 0)
	assert.NoError(t, err)

	// 获取复杂结构
	var result User
	err = cache.Get(ctx, key, &result)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.Name, result.Name)
	assert.Equal(t, user.Age, result.Age)
	assert.Equal(t, user.Email, result.Email)
	assert.Equal(t, user.Tags, result.Tags)
}

// TestMemoryCache_Concurrent 测试并发安全
func TestMemoryCache_Concurrent(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	const goroutines = 100
	var wg sync.WaitGroup

	// 并发写入
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(index int) {
			defer wg.Done()
			key := fmt.Sprintf("key_%d", index)
			err := cache.Set(ctx, key, index, 0)
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()

	// 并发读取
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(index int) {
			defer wg.Done()
			key := fmt.Sprintf("key_%d", index)
			var result int
			err := cache.Get(ctx, key, &result)
			assert.NoError(t, err)
			assert.Equal(t, index, result)
		}(i)
	}
	wg.Wait()

	// 并发删除
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(index int) {
			defer wg.Done()
			key := fmt.Sprintf("key_%d", index)
			err := cache.Delete(ctx, key)
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()
}

// TestMemoryCache_EdgeCases 测试边界情况
func TestMemoryCache_EdgeCases(t *testing.T) {
	cache := NewMemoryCache()
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

	// nil slice
	err = cache.Delete(ctx)
	assert.NoError(t, err, "删除空 key 列表应该成功")

	// 空前缀
	err = cache.DeletePrefix(ctx, "")
	assert.NoError(t, err, "删除空前缀应该成功")
}
