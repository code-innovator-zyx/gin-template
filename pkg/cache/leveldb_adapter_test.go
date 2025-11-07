package cache

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/05
* @Package: LevelDB缓存测试
 */

func createTestLevelDB(t *testing.T) (Cache, func()) {
	tmpDir := filepath.Join("./", "leveldb_test")

	cfg := LevelDBConfig{
		Path: tmpDir,
	}

	cache, err := NewLevelDBCache(cfg)
	if err != nil {
		t.Fatalf("创建LevelDB缓存失败: %v", err)
	}

	cleanup := func() {
		cache.Close()
		os.RemoveAll(tmpDir)
	}

	return cache, cleanup
}

func TestLevelDBCache_Type(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()

	if cache.Type() != "leveldb" {
		t.Errorf("期望类型为 leveldb，实际为: %s", cache.Type())
	}
}

func TestLevelDBCache_SetAndGet(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
	ctx := context.Background()

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

func TestLevelDBCache_SetWithTTL(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
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
	fmt.Println(result)
}

func TestLevelDBCache_GetNonExistent(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
	ctx := context.Background()

	var result string
	err := cache.Get(ctx, "non_existent_key", &result)
	if err == nil {
		t.Error("期望获取不存在的key失败，但成功了")
	}
}

func TestLevelDBCache_Delete(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
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

func TestLevelDBCache_DeleteMultiple(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
	ctx := context.Background()

	keys := []string{"key1", "key2", "key3"}
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

func TestLevelDBCache_Exists(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
	ctx := context.Background()

	key := "exists_key"

	exists, err := cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists 失败: %v", err)
	}
	if exists {
		t.Error("期望key不存在，但返回存在")
	}

	cache.Set(ctx, key, "value", 0)
	exists, err = cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists 失败: %v", err)
	}
	if !exists {
		t.Error("期望key存在，但返回不存在")
	}
}

func TestLevelDBCache_SAdd(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
	ctx := context.Background()

	key := "set_key"
	members := []interface{}{"member1", "member2", "member3"}

	err := cache.SAdd(ctx, key, members...)
	if err != nil {
		t.Fatalf("SAdd 失败: %v", err)
	}

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

func TestLevelDBCache_SIsMember(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
	ctx := context.Background()

	key := "set_key"
	member := "member1"

	isMember, err := cache.SIsMember(ctx, key, member)
	if err != nil {
		t.Fatalf("SIsMember 失败: %v", err)
	}
	if isMember {
		t.Error("期望成员不存在，但返回存在")
	}

	cache.SAdd(ctx, key, member)
	isMember, err = cache.SIsMember(ctx, key, member)
	if err != nil {
		t.Fatalf("SIsMember 失败: %v", err)
	}
	if !isMember {
		t.Error("期望成员存在，但返回不存在")
	}
}

func TestLevelDBCache_SMembers(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
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

	memberMap := make(map[string]bool)

	for _, expected := range expectedMembers {
		if !memberMap[expected.(string)] {
			t.Errorf("期望成员 %v 存在，但未找到", expected)
		}
	}
}

func TestLevelDBCache_SMembersEmpty(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
	ctx := context.Background()

	members, err := cache.SMembers(ctx, "non_existent_set")
	if err != nil {
		t.Fatalf("SMembers 失败: %v", err)
	}

	if len(members) != 0 {
		t.Errorf("期望空集合，实际有 %d 个成员", len(members))
	}
}

func TestLevelDBCache_Expire(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
	ctx := context.Background()

	key := "expire_key"
	value := "expire_value"

	cache.Set(ctx, key, value, 0)

	ttl := 100 * time.Millisecond
	err := cache.Expire(ctx, key, ttl)
	if err != nil {
		t.Fatalf("Expire 失败: %v", err)
	}

	var result string
	err = cache.Get(ctx, key, &result)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	err = cache.Get(ctx, key, &result)
	if err == nil {
		t.Error("期望获取过期key失败，但成功了")
	}
}

func TestLevelDBCache_ExpireNonExistent(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
	ctx := context.Background()

	err := cache.Expire(ctx, "non_existent_key", time.Second)
	if err == nil {
		t.Error("期望对不存在的key设置过期时间失败，但成功了")
	}
}

func TestLevelDBCache_TTL(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
	ctx := context.Background()

	key := "ttl_key"
	value := "ttl_value"
	ttl := 10 * time.Second

	cache.Set(ctx, key, value, ttl)

	remaining, err := cache.TTL(ctx, key)
	if err != nil {
		t.Fatalf("TTL 失败: %v", err)
	}

	if remaining < 9*time.Second || remaining > ttl {
		t.Errorf("期望TTL约为 %v，实际为: %v", ttl, remaining)
	}
}

func TestLevelDBCache_TTLNeverExpire(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
	ctx := context.Background()

	key := "never_expire_key"
	value := "never_expire_value"

	cache.Set(ctx, key, value, 0)

	remaining, err := cache.TTL(ctx, key)
	if err != nil {
		t.Fatalf("TTL 失败: %v", err)
	}

	if remaining != -1 {
		t.Errorf("期望TTL为 -1（永不过期），实际为: %v", remaining)
	}
}

func TestLevelDBCache_TTLNonExistent(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
	ctx := context.Background()

	_, err := cache.TTL(ctx, "non_existent_key")
	if err == nil {
		t.Error("期望获取不存在key的TTL失败，但成功了")
	}
}

func TestLevelDBCache_Pipeline(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
	ctx := context.Background()

	cache.Set(ctx, "key1", "value1", 0)
	cache.SAdd(ctx, "set1", "member1")

	pipe := cache.Pipeline()

	existsCmd := pipe.Exists(ctx, "key1")
	isMemberCmd := pipe.SIsMember(ctx, "set1", "member1")
	expireCmd := pipe.Expire(ctx, "key1", time.Hour)

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

func TestLevelDBCache_Ping(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
	ctx := context.Background()

	err := cache.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping 失败: %v", err)
	}
}

func TestLevelDBCache_ComplexStruct(t *testing.T) {
	cache, cleanup := createTestLevelDB(t)
	defer cleanup()
	ctx := context.Background()

	type User struct {
		ID   int
		Name string
		Age  int
	}

	key := "user:1"
	user := User{ID: 1, Name: "张三", Age: 25}

	err := cache.Set(ctx, key, user, 0)
	if err != nil {
		t.Fatalf("Set 失败: %v", err)
	}

	var result User
	err = cache.Get(ctx, key, &result)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}

	if result.ID != user.ID || result.Name != user.Name || result.Age != user.Age {
		t.Errorf("期望用户为 %+v，实际为: %+v", user, result)
	}
}

func TestLevelDBCache_Persistence(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("leveldb_persist_test_%d", time.Now().UnixNano()))
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	cfg := LevelDBConfig{Path: tmpDir}

	// 第一次创建并写入数据
	cache1, err := NewLevelDBCache(cfg)
	if err != nil {
		t.Fatalf("创建LevelDB缓存失败: %v", err)
	}

	key := "persist_key"
	value := "persist_value"
	cache1.Set(ctx, key, value, 0)
	cache1.Close()

	// 重新打开，验证数据持久化
	cache2, err := NewLevelDBCache(cfg)
	if err != nil {
		t.Fatalf("重新打开LevelDB缓存失败: %v", err)
	}
	defer cache2.Close()

	var result string
	err = cache2.Get(ctx, key, &result)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}

	if result != value {
		t.Errorf("期望值为 %s，实际为: %s", value, result)
	}
}
