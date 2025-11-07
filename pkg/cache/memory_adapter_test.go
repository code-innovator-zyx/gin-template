package cache

import (
	"context"
	"fmt"
	"testing"
	"time"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/05
* @Package: 内存缓存测试
 */

func TestMemoryCache_Type(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	if cache.Type() != "memory" {
		t.Errorf("期望类型为 memory，实际为: %s", cache.Type())
	}
}

func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	// 测试设置和获取字符串
	key := "test_key"
	value := "test_value"
	err := cache.Set(ctx, key, value, 0)
	if err != nil {
		t.Fatalf("Set 失败: %v", err)
	}

	var result string
	err = cache.Get(ctx, key, &result)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}

	if result != value {
		t.Errorf("期望值为 %s，实际为: %s", value, result)
	}
}

func TestMemoryCache_SetWithTTL(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "ttl_key"
	value := "ttl_value"
	ttl := 100 * time.Millisecond

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
	time.Sleep(150 * time.Millisecond)

	err = cache.Get(ctx, key, &result)
	if err == nil {
		t.Error("期望获取过期key失败，但成功了")
	}
}

func TestMemoryCache_GetNonExistent(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	var result string
	err := cache.Get(ctx, "non_existent_key", &result)
	if err == nil {
		t.Error("期望获取不存在的key失败，但成功了")
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "delete_key"
	value := "delete_value"

	// 设置key
	cache.Set(ctx, key, value, 0)

	// 删除key
	err := cache.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}

	// 验证key已被删除
	var result string
	err = cache.Get(ctx, key, &result)
	if err == nil {
		t.Error("期望获取已删除的key失败，但成功了")
	}
}

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
	if err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}

	// 验证所有key都已删除
	for _, key := range keys {
		var result string
		err = cache.Get(ctx, key, &result)
		if err == nil {
			t.Errorf("期望key %s 已被删除，但仍然存在", key)
		}
	}
}

func TestMemoryCache_Exists(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "exists_key"

	// key不存在时
	exists, err := cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists 失败: %v", err)
	}
	if exists {
		t.Error("期望key不存在，但返回存在")
	}

	// 设置key后
	cache.Set(ctx, key, "value", 0)
	exists, err = cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists 失败: %v", err)
	}
	if !exists {
		t.Error("期望key存在，但返回不存在")
	}
}

func TestMemoryCache_SAdd(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "set_key"
	members := []interface{}{"member1", "member2", "member3"}

	err := cache.SAdd(ctx, key, members...)
	if err != nil {
		t.Fatalf("SAdd 失败: %v", err)
	}

	// 验证成员是否添加成功
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

func TestMemoryCache_SIsMember(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "set_key"
	member := "member1"

	// 成员不存在时
	isMember, err := cache.SIsMember(ctx, key, member)
	if err != nil {
		t.Fatalf("SIsMember 失败: %v", err)
	}
	if isMember {
		t.Error("期望成员不存在，但返回存在")
	}

	// 添加成员后
	cache.SAdd(ctx, key, member)
	isMember, err = cache.SIsMember(ctx, key, member)
	if err != nil {
		t.Fatalf("SIsMember 失败: %v", err)
	}
	if !isMember {
		t.Error("期望成员存在，但返回不存在")
	}
}

func TestMemoryCache_SMembers(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "set_key"
	expectedMembers := []interface{}{"member1", "member2", "member3"}

	cache.SAdd(ctx, key, expectedMembers...)

	members, err := cache.SMembers(ctx, key)
	if err != nil {
		t.Fatalf("SMembers 失败: %v", err)
	}

	if len(members) != len(expectedMembers) {
		t.Errorf("期望成员数量为 %d，实际为: %d", len(expectedMembers), len(members))
	}

	// 验证所有成员都存在
	memberMap := make(map[string]bool)

	for _, expected := range expectedMembers {
		if !memberMap[expected.(string)] {
			t.Errorf("期望成员 %v 存在，但未找到", expected)
		}
	}
}

func TestMemoryCache_SMembersEmpty(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	members, err := cache.SMembers(ctx, "non_existent_set")
	if err != nil {
		t.Fatalf("SMembers 失败: %v", err)
	}

	if len(members) != 0 {
		t.Errorf("期望空集合，实际有 %d 个成员", len(members))
	}
}

func TestMemoryCache_Expire(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "expire_key"
	value := "expire_value"

	// 设置永不过期的key
	cache.Set(ctx, key, value, 0)

	// 设置过期时间
	ttl := 100 * time.Millisecond
	err := cache.Expire(ctx, key, ttl)
	if err != nil {
		t.Fatalf("Expire 失败: %v", err)
	}

	// 立即获取应该成功
	var result string
	err = cache.Get(ctx, key, &result)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	err = cache.Get(ctx, key, &result)
	if err == nil {
		t.Error("期望获取过期key失败，但成功了")
	}
}

func TestMemoryCache_ExpireNonExistent(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	err := cache.Expire(ctx, "non_existent_key", time.Second)
	if err == nil {
		t.Error("期望对不存在的key设置过期时间失败，但成功了")
	}
}

func TestMemoryCache_TTL(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "ttl_key"
	value := "ttl_value"
	ttl := 10 * time.Second

	// 设置带TTL的key
	cache.Set(ctx, key, value, ttl)

	// 获取TTL
	remaining, err := cache.TTL(ctx, key)
	if err != nil {
		t.Fatalf("TTL 失败: %v", err)
	}

	// TTL应该接近设置的值（允许一些误差）
	if remaining < 9*time.Second || remaining > ttl {
		t.Errorf("期望TTL约为 %v，实际为: %v", ttl, remaining)
	}
}

func TestMemoryCache_TTLNeverExpire(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	key := "never_expire_key"
	value := "never_expire_value"

	// 设置永不过期的key
	cache.Set(ctx, key, value, 0)

	// 获取TTL
	remaining, err := cache.TTL(ctx, key)
	if err != nil {
		t.Fatalf("TTL 失败: %v", err)
	}

	// 永不过期应该返回-1
	if remaining != -1 {
		t.Errorf("期望TTL为 -1（永不过期），实际为: %v", remaining)
	}
}

func TestMemoryCache_TTLNonExistent(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	_, err := cache.TTL(ctx, "non_existent_key")
	if err == nil {
		t.Error("期望获取不存在key的TTL失败，但成功了")
	}
}

func TestMemoryCache_Pipeline(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	// 设置测试数据
	cache.Set(ctx, "key1", "value1", 0)
	cache.SAdd(ctx, "set1", "member1")

	// 创建管道
	pipe := cache.Pipeline()

	// 执行管道操作
	existsCmd := pipe.Exists(ctx, "key1")
	isMemberCmd := pipe.SIsMember(ctx, "set1", "member1")
	expireCmd := pipe.Expire(ctx, "key1", time.Hour)

	err := pipe.Exec(ctx)
	if err != nil {
		t.Fatalf("Pipeline Exec 失败: %v", err)
	}

	// 验证结果
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

func TestMemoryCache_Ping(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	err := cache.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping 失败: %v", err)
	}
}

func TestMemoryCache_ComplexStruct(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	type User struct {
		ID   int
		Name string
		Age  int
	}

	key := "user:1"
	user := User{ID: 1, Name: "张三", Age: 25}

	// 设置复杂结构
	err := cache.Set(ctx, key, user, 0)
	if err != nil {
		t.Fatalf("Set 失败: %v", err)
	}

	// 获取复杂结构
	var result User
	err = cache.Get(ctx, key, &result)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}

	if result.ID != user.ID || result.Name != user.Name || result.Age != user.Age {
		t.Errorf("期望用户为 %+v，实际为: %+v", user, result)
	}
}

func TestMemoryCache_Concurrent(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	ctx := context.Background()

	// 并发写入
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(index int) {
			key := fmt.Sprintf("key_%d", index)
			err := cache.Set(ctx, key, index, 0)
			if err != nil {
				t.Errorf("并发Set失败: %v", err)
			}
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证所有key都已设置
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key_%d", i)
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
