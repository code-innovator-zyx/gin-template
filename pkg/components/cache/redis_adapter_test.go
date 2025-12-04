package cache

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	redis2 "gin-admin/pkg/components/redis"
	"github.com/stretchr/testify/assert"
)

/*
 * @Author: zouyx
 * @Email: 1003941268@qq.com
 * @Date:   2025/11/05
 * @Package: Redis 缓存测试
 */

// skipIfNoRedis 如果 Redis 不可用则跳过测试
func skipIfNoRedis(t *testing.T) ICache {
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
		t.Skipf("Redis 连接失败，跳过测试: %v", err)
	}

	cache, err := NewRedisCache(client)
	if err != nil {
		t.Skipf("Redis 不可用，跳过测试: %v", err)
	}

	return cache
}

// TestRedisCache_SetAndGet 测试基本读写
func TestRedisCache_SetAndGet(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "test_key"
	value := "test_value"

	err := cache.Set(ctx, key, value, 0)
	assert.NoError(t, err, "Set 应该成功")
	defer cache.Delete(ctx, key)

	var result string
	err = cache.Get(ctx, key, &result)
	assert.NoError(t, err, "Get 应该成功")
	assert.Equal(t, value, result, "获取的值应该与设置的值相同")
}

// TestRedisCache_SetWithTTL 测试带过期时间的缓存
func TestRedisCache_SetWithTTL(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "ttl_key"
	value := "ttl_value"
	ttl := 1 * time.Second

	err := cache.Set(ctx, key, value, ttl)
	assert.NoError(t, err)

	// 立即获取应该成功
	var result string
	err = cache.Get(ctx, key, &result)
	assert.NoError(t, err, "TTL 未过期时应该能获取到值")

	// 等待过期
	time.Sleep(1500 * time.Millisecond)

	err = cache.Get(ctx, key, &result)
	assert.Error(t, err, "TTL 过期后应该获取失败")
}

// TestRedisCache_GetNonExistent 测试获取不存在的 key
func TestRedisCache_GetNonExistent(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	var result string
	err := cache.Get(ctx, "non_existent_key_redis", &result)
	assert.Error(t, err, "获取不存在的 key 应该失败")
}

// TestRedisCache_Delete 测试删除功能
func TestRedisCache_Delete(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "delete_key"
	value := "delete_value"

	cache.Set(ctx, key, value, 0)

	err := cache.Delete(ctx, key)
	assert.NoError(t, err, "Delete 应该成功")

	var result string
	err = cache.Get(ctx, key, &result)
	assert.Error(t, err, "删除后应该获取失败")
}

// TestRedisCache_DeleteMultiple 测试批量删除
func TestRedisCache_DeleteMultiple(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	keys := []string{"redis_key1", "redis_key2", "redis_key3"}
	for _, key := range keys {
		cache.Set(ctx, key, "value", 0)
	}

	err := cache.Delete(ctx, keys...)
	assert.NoError(t, err)

	for _, key := range keys {
		var result string
		err = cache.Get(ctx, key, &result)
		assert.Error(t, err, "key %s 应该已被删除", key)
	}
}

// TestRedisCache_DeletePrefix 测试前缀删除
func TestRedisCache_DeletePrefix(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	// 设置多个带前缀的 key
	cache.Set(ctx, "prefix:key1", "data1", 0)
	cache.Set(ctx, "prefix:key2", "data2", 0)
	cache.Set(ctx, "other:key1", "data3", 0)

	// 删除 prefix: 前缀的所有 key
	err := cache.DeletePrefix(ctx, "prefix:")
	assert.NoError(t, err)

	// 验证 prefix: 前缀的 key 被删除
	var result string
	err = cache.Get(ctx, "prefix:key1", &result)
	assert.Error(t, err)

	// other: 前缀的 key 应该还在
	err = cache.Get(ctx, "other:key1", &result)
	assert.NoError(t, err)

	// 清理
	cache.Delete(ctx, "other:key1")
}

// TestRedisCache_Exists 测试 key 是否存在
func TestRedisCache_Exists(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "exists_key_redis"

	exists, err := cache.Exists(ctx, key)
	assert.NoError(t, err)
	assert.False(t, exists, "key 应该不存在")

	cache.Set(ctx, key, "value", 0)
	defer cache.Delete(ctx, key)

	exists, err = cache.Exists(ctx, key)
	assert.NoError(t, err)
	assert.True(t, exists, "key 应该存在")
}

// TestRedisCache_SAdd 测试集合添加
func TestRedisCache_SAdd(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "set_key_redis"
	members := []interface{}{"member1", "member2", "member3"}

	err := cache.SAdd(ctx, key, members...)
	assert.NoError(t, err)
	defer cache.Delete(ctx, key)

	// 验证成员是否添加成功
	for _, member := range members {
		isMember, err := cache.SIsMember(ctx, key, member)
		assert.NoError(t, err)
		assert.True(t, isMember, "member %v 应该在集合中", member)
	}
}

// TestRedisCache_SIsMember 测试集合成员检查
func TestRedisCache_SIsMember(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "set_key_redis_2"
	member := "member1"

	isMember, err := cache.SIsMember(ctx, key, member)
	assert.NoError(t, err)
	assert.False(t, isMember, "成员应该不存在")

	cache.SAdd(ctx, key, member)
	defer cache.Delete(ctx, key)

	isMember, err = cache.SIsMember(ctx, key, member)
	assert.NoError(t, err)
	assert.True(t, isMember, "成员应该存在")
}

// TestRedisCache_SMembers 测试获取集合所有成员
func TestRedisCache_SMembers(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "set_key_redis_3"
	expectedMembers := []interface{}{"member1", "member2", "member3"}

	cache.SAdd(ctx, key, expectedMembers...)
	defer cache.Delete(ctx, key)

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

// TestRedisCache_SMembersEmpty 测试空集合
func TestRedisCache_SMembersEmpty(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	members, err := cache.SMembers(ctx, "non_existent_set_redis")
	assert.NoError(t, err)
	assert.Empty(t, members, "空集合应该返回空数组")
}

// TestRedisCache_SRem 测试集合删除成员
func TestRedisCache_SRem(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "set_key_redis_rem"
	cache.SAdd(ctx, key, "member1", "member2", "member3")
	defer cache.Delete(ctx, key)

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

// TestRedisCache_Expire 测试设置过期时间
func TestRedisCache_Expire(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "expire_key_redis"
	value := "expire_value"

	cache.Set(ctx, key, value, 0)
	defer cache.Delete(ctx, key)

	ttl := 1 * time.Second
	err := cache.Expire(ctx, key, ttl)
	assert.NoError(t, err)

	var result string
	err = cache.Get(ctx, key, &result)
	assert.NoError(t, err)

	time.Sleep(1500 * time.Millisecond)

	err = cache.Get(ctx, key, &result)
	assert.Error(t, err, "过期后应该获取失败")
}

// TestRedisCache_TTL 测试获取剩余过期时间
func TestRedisCache_TTL(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "ttl_key_redis"
	value := "ttl_value"
	ttl := 10 * time.Second

	cache.Set(ctx, key, value, ttl)
	defer cache.Delete(ctx, key)

	remaining, err := cache.TTL(ctx, key)
	assert.NoError(t, err)
	assert.Greater(t, remaining, 9*time.Second, "TTL 应该大于 9 秒")
	assert.LessOrEqual(t, remaining, ttl, "TTL 应该小于等于设置的值")
}

// TestRedisCache_TTLNeverExpire 测试永不过期的 key
func TestRedisCache_TTLNeverExpire(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "never_expire_key_redis"
	value := "never_expire_value"

	cache.Set(ctx, key, value, 0)
	defer cache.Delete(ctx, key)

	remaining, err := cache.TTL(ctx, key)
	assert.NoError(t, err)
	// Redis 中永不过期的 TTL 返回 -1
	assert.Equal(t, time.Duration(-1), remaining, "永不过期应该返回 -1")
}

// TestRedisCache_TTLNonExistent 测试获取不存在 key 的 TTL
func TestRedisCache_TTLNonExistent(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	remaining, err := cache.TTL(ctx, "non_existent_key_redis_ttl")
	assert.NoError(t, err)
	// Redis 对不存在的 key 返回 -2
	assert.Equal(t, -2*time.Nanosecond, remaining, "不存在的 key TTL 应该返回 -2")
}

// TestRedisCache_Incr 测试递增计数器
func TestRedisCache_Incr(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "redis_counter"
	defer cache.Delete(ctx, key)

	// 第一次递增（从 0 开始）
	val, err := cache.Incr(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	// 第二次递增
	val, err = cache.Incr(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), val)
}

// TestRedisCache_Decr 测试递减计数器
func TestRedisCache_Decr(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "redis_counter_decr"
	defer cache.Delete(ctx, key)

	// 第一次递减（从 0 开始）
	val, err := cache.Decr(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(-1), val)

	// 第二次递减
	val, err = cache.Decr(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(-2), val)
}

// TestRedisCache_Pipeline 测试管道操作
func TestRedisCache_Pipeline(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	cache.Set(ctx, "redis_pipe_key1", "value1", 0)
	cache.SAdd(ctx, "redis_pipe_set1", "member1")
	defer func() {
		cache.Delete(ctx, "redis_pipe_key1", "redis_pipe_key2")
		cache.Delete(ctx, "redis_pipe_set1")
	}()

	pipe := cache.Pipeline()

	setCmd := pipe.Set(ctx, "redis_pipe_key2", "value2", time.Second)
	existsCmd := pipe.Exists(ctx, "redis_pipe_key1")
	isMemberCmd := pipe.SIsMember(ctx, "redis_pipe_set1", "member1")
	expireCmd := pipe.Expire(ctx, "redis_pipe_key1", time.Hour)

	err := pipe.Exec(ctx)
	assert.NoError(t, err, "Pipeline 执行应该成功")

	// 验证结果
	status, err := setCmd.Result()
	assert.NoError(t, err)
	assert.Equal(t, "OK", status)

	exists, err := existsCmd.Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), exists, "key1 应该存在")

	isMember, err := isMemberCmd.Result()
	assert.NoError(t, err)
	assert.True(t, isMember, "member1 应该存在")

	expired, err := expireCmd.Result()
	assert.NoError(t, err)
	assert.True(t, expired, "设置过期时间应该成功")
}

// TestRedisCache_Ping 测试连接
func TestRedisCache_Ping(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	err := cache.Ping(ctx)
	assert.NoError(t, err, "Ping 应该成功")
}

// TestRedisCache_ComplexStruct 测试复杂结构体
func TestRedisCache_ComplexStruct(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	type User struct {
		ID    int
		Name  string
		Age   int
		Email string
		Tags  []string
	}

	key := "user:redis:1"
	user := User{
		ID:    1,
		Name:  "李四",
		Age:   30,
		Email: "lisi@example.com",
		Tags:  []string{"vip", "active"},
	}

	err := cache.Set(ctx, key, user, 0)
	assert.NoError(t, err)
	defer cache.Delete(ctx, key)

	var result User
	err = cache.Get(ctx, key, &result)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.Name, result.Name)
	assert.Equal(t, user.Age, result.Age)
	assert.Equal(t, user.Email, result.Email)
	assert.Equal(t, user.Tags, result.Tags)
}

// TestRedisCache_Concurrent 测试并发安全
func TestRedisCache_Concurrent(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	const goroutines = 100
	var wg sync.WaitGroup

	// 并发写入
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(index int) {
			defer wg.Done()
			key := fmt.Sprintf("redis_concurrent_key_%d", index)
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
			key := fmt.Sprintf("redis_concurrent_key_%d", index)
			var result int
			err := cache.Get(ctx, key, &result)
			assert.NoError(t, err)
			assert.Equal(t, index, result)
		}(i)
	}
	wg.Wait()

	// 清理
	keys := make([]string, goroutines)
	for i := 0; i < goroutines; i++ {
		keys[i] = fmt.Sprintf("redis_concurrent_key_%d", i)
	}
	cache.Delete(ctx, keys...)
}

// TestRedisCache_EdgeCases 测试边界情况
func TestRedisCache_EdgeCases(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	// 空字符串作为 value
	key := "edge_case_key"
	err := cache.Set(ctx, key, "", 0)
	assert.NoError(t, err)
	defer cache.Delete(ctx, key)

	var result string
	err = cache.Get(ctx, key, &result)
	assert.NoError(t, err)
	assert.Equal(t, "", result)
}
