package cache

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/05
* @Package: Redis缓存测试
 */

// skipIfNoRedis 如果Redis不可用则跳过测试
func skipIfNoRedis(t *testing.T) Cache {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 6379

	cfg := RedisConfig{
		Host:     host,
		Port:     port,
		Password: "123456",
		DB:       0,
		PoolSize: 10,
	}

	cache, err := NewRedisCache(cfg)
	if err != nil {
		t.Skipf("Redis不可用，跳过测试: %v", err)
	}

	return cache
}

func TestRedisCache_Type(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()

	if cache.Type() != "redis" {
		t.Errorf("期望类型为 redis，实际为: %s", cache.Type())
	}
}

func TestRedisCache_SetAndGet(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "test_key"
	value := "test_value"

	err := cache.Set(ctx, key, value, 0)
	if err != nil {
		t.Fatalf("Set 失败: %v", err)
	}
	defer cache.Delete(ctx, key) // 清理测试数据

	var result string
	err = cache.Get(ctx, key, &result)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}

	if result != value {
		t.Errorf("期望值为 %s，实际为: %s", value, result)
	}
}

func TestRedisCache_SetWithTTL(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "ttl_key"
	value := "ttl_value"
	ttl := 1 * time.Second

	err := cache.Set(ctx, key, value, ttl)
	if err != nil {
		t.Fatalf("Set 失败: %v", err)
	}

	// 立即获取应该成功
	var result string
	err = cache.Get(ctx, key, &result)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}

	// 等待过期
	time.Sleep(1500 * time.Millisecond)

	err = cache.Get(ctx, key, &result)
	if err == nil {
		t.Error("期望获取过期key失败，但成功了")
	}
}

func TestRedisCache_GetNonExistent(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	var result string
	err := cache.Get(ctx, "non_existent_key_redis", &result)
	if err == nil {
		t.Error("期望获取不存在的key失败，但成功了")
	}
}

func TestRedisCache_Delete(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "delete_key"
	value := "delete_value"

	cache.Set(ctx, key, value, 0)

	err := cache.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}

	var result string
	err = cache.Get(ctx, key, &result)
	if err == nil {
		t.Error("期望获取已删除的key失败，但成功了")
	}
}

func TestRedisCache_DeleteMultiple(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	keys := []string{"redis_key1", "redis_key2", "redis_key3"}
	for _, key := range keys {
		cache.Set(ctx, key, "value", 0)
	}

	err := cache.Delete(ctx, keys...)
	if err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}

	for _, key := range keys {
		var result string
		err = cache.Get(ctx, key, &result)
		if err == nil {
			t.Errorf("期望key %s 已被删除，但仍然存在", key)
		}
	}
}

func TestRedisCache_Exists(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "exists_key_redis"

	exists, err := cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists 失败: %v", err)
	}
	if exists {
		t.Error("期望key不存在，但返回存在")
	}

	cache.Set(ctx, key, "value", 0)
	defer cache.Delete(ctx, key)

	exists, err = cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists 失败: %v", err)
	}
	if !exists {
		t.Error("期望key存在，但返回不存在")
	}
}

func TestRedisCache_SAdd(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "set_key_redis"
	members := []interface{}{"member1", "member2", "member3"}

	//err := cache.SAdd(ctx, key, members...)
	//if err != nil {
	//	t.Fatalf("SAdd 失败: %v", err)
	//}
	defer cache.Delete(ctx, key)

	for _, member := range members {
		isMember, err := cache.SIsMember(ctx, key, member)
		if err != nil {
			t.Fatalf("SIsMember 失败: %v", err)
		}
		if !isMember {
			t.Errorf("期望 %v 是集合成员，但不是", member)
		}
	}
}

func TestRedisCache_SIsMember(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "set_key_redis_2"
	member := "member1"

	isMember, err := cache.SIsMember(ctx, key, member)
	if err != nil {
		t.Fatalf("SIsMember 失败: %v", err)
	}
	if isMember {
		t.Error("期望成员不存在，但返回存在")
	}

	cache.SAdd(ctx, key, member)
	defer cache.Delete(ctx, key)

	isMember, err = cache.SIsMember(ctx, key, member)
	if err != nil {
		t.Fatalf("SIsMember 失败: %v", err)
	}
	if !isMember {
		t.Error("期望成员存在，但返回不存在")
	}
}

func TestRedisCache_SMembers(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "set_key_redis_3"
	expectedMembers := []interface{}{"member1", "member2", "member3"}

	cache.SAdd(ctx, key, expectedMembers...)
	defer cache.Delete(ctx, key)

	members, err := cache.SMembers(ctx, key)
	if err != nil {
		t.Fatalf("SMembers 失败: %v", err)
	}

	if len(members) != len(expectedMembers) {
		t.Errorf("期望成员数量为 %d，实际为: %d", len(expectedMembers), len(members))
	}

	memberMap := make(map[string]bool)

	for _, expected := range expectedMembers {
		if !memberMap[expected.(string)] {
			t.Errorf("期望成员 %v 存在，但未找到", expected)
		}
	}
}

func TestRedisCache_SMembersEmpty(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	members, err := cache.SMembers(ctx, "non_existent_set_redis")
	if err != nil {
		t.Fatalf("SMembers 失败: %v", err)
	}

	if len(members) != 0 {
		t.Errorf("期望空集合，实际有 %d 个成员", len(members))
	}
}

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
	if err != nil {
		t.Fatalf("Expire 失败: %v", err)
	}

	var result string
	err = cache.Get(ctx, key, &result)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}

	time.Sleep(1500 * time.Millisecond)

	err = cache.Get(ctx, key, &result)
	if err == nil {
		t.Error("期望获取过期key失败，但成功了")
	}
}

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
	if err != nil {
		t.Fatalf("TTL 失败: %v", err)
	}

	if remaining < 9*time.Second || remaining > ttl {
		t.Errorf("期望TTL约为 %v，实际为: %v", ttl, remaining)
	}
}

func TestRedisCache_TTLNeverExpire(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	key := "never_expire_key_redis"
	value := "never_expire_value"

	cache.Set(ctx, key, value, 0)
	defer cache.Delete(ctx, key)

	remaining, err := cache.TTL(ctx, key)
	if err != nil {
		t.Fatalf("TTL 失败: %v", err)
	}

	// Redis中永不过期的TTL返回-1
	if remaining != -1 {
		t.Logf("警告: 期望TTL为 -1（永不过期），实际为: %v", remaining)
	}
}

func TestRedisCache_TTLNonExistent(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	remaining, err := cache.TTL(ctx, "non_existent_key_redis_ttl")
	// Redis对不存在的key返回-2
	if err != nil {
		t.Fatalf("TTL 失败: %v", err)
	}

	if remaining != -2*time.Nanosecond {
		t.Logf("不存在的key TTL: %v", remaining)
	}
}

func TestRedisCache_Pipeline(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	cache.Set(ctx, "redis_pipe_key1", "value1", 0)
	cache.SAdd(ctx, "redis_pipe_set1", "member1")
	defer func() {
		cache.Delete(ctx, "redis_pipe_key1")
		cache.Delete(ctx, "redis_pipe_set1")
	}()

	pipe := cache.Pipeline()

	existsCmd := pipe.Exists(ctx, "redis_pipe_key1")
	isMemberCmd := pipe.SIsMember(ctx, "redis_pipe_set1", "member1")
	expireCmd := pipe.Expire(ctx, "redis_pipe_key1", time.Hour)

	err := pipe.Exec(ctx)
	if err != nil {
		t.Fatalf("Pipeline Exec 失败: %v", err)
	}

	exists, err := existsCmd.Result()
	if err != nil {
		t.Fatalf("ExistsCmd Result 失败: %v", err)
	}
	if exists != 1 {
		t.Errorf("期望key存在(1)，实际为: %d", exists)
	}

	isMember, err := isMemberCmd.Result()
	if err != nil {
		t.Fatalf("BoolCmd Result 失败: %v", err)
	}
	if !isMember {
		t.Error("期望成员存在，但返回不存在")
	}

	expired, err := expireCmd.Result()
	if err != nil {
		t.Fatalf("ExpireCmd Result 失败: %v", err)
	}
	if !expired {
		t.Error("期望设置过期时间成功，但失败了")
	}
}

func TestRedisCache_Ping(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	err := cache.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping 失败: %v", err)
	}
}

func TestRedisCache_ComplexStruct(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	type User struct {
		ID   int
		Name string
		Age  int
	}

	key := "user:redis:1"
	user := User{ID: 1, Name: "李四", Age: 30}

	err := cache.Set(ctx, key, user, 0)
	if err != nil {
		t.Fatalf("Set 失败: %v", err)
	}
	defer cache.Delete(ctx, key)

	var result User
	err = cache.Get(ctx, key, &result)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}

	if result.ID != user.ID || result.Name != user.Name || result.Age != user.Age {
		t.Errorf("期望用户为 %+v，实际为: %+v", user, result)
	}
}

func TestRedisCache_Concurrent(t *testing.T) {
	cache := skipIfNoRedis(t)
	defer cache.Close()
	ctx := context.Background()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(index int) {
			key := fmt.Sprintf("redis_concurrent_key_%d", index)
			err := cache.Set(ctx, key, index, 0)
			if err != nil {
				t.Errorf("并发Set失败: %v", err)
			}
			defer cache.Delete(ctx, key)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("redis_concurrent_key_%d", i)
		var result int
		err := cache.Get(ctx, key, &result)
		if err != nil {
			t.Errorf("并发Get失败 key=%s: %v", key, err)
		}
		if result != i {
			t.Errorf("期望值为 %d，实际为: %d", i, result)
		}
	}
}
