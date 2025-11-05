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
* @Package: LevelDB性能基准测试
 */

// BenchmarkLevelDBCache_Set 基准测试Set操作
func BenchmarkLevelDBCache_Set(b *testing.B) {
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("leveldb_bench_%d", time.Now().UnixNano()))
	defer os.RemoveAll(tmpDir)

	cfg := LevelDBConfig{Path: tmpDir}
	cache, err := NewLevelDBCache(cfg)
	if err != nil {
		b.Fatalf("创建LevelDB缓存失败: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	testData := map[string]interface{}{
		"id":   12345,
		"name": "测试用户",
		"age":  25,
		"tags": []string{"tag1", "tag2", "tag3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		cache.Set(ctx, key, testData, time.Hour)
	}
}

// BenchmarkLevelDBCache_Get 基准测试Get操作
func BenchmarkLevelDBCache_Get(b *testing.B) {
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("leveldb_bench_%d", time.Now().UnixNano()))
	defer os.RemoveAll(tmpDir)

	cfg := LevelDBConfig{Path: tmpDir}
	cache, err := NewLevelDBCache(cfg)
	if err != nil {
		b.Fatalf("创建LevelDB缓存失败: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	testData := map[string]interface{}{
		"id":   12345,
		"name": "测试用户",
		"age":  25,
		"tags": []string{"tag1", "tag2", "tag3"},
	}

	// 预先写入测试数据
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		cache.Set(ctx, key, testData, 0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i%1000)
		var result map[string]interface{}
		cache.Get(ctx, key, &result)
	}
}

// BenchmarkLevelDBCache_SetGet 基准测试Set+Get组合操作
func BenchmarkLevelDBCache_SetGet(b *testing.B) {
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("leveldb_bench_%d", time.Now().UnixNano()))
	defer os.RemoveAll(tmpDir)

	cfg := LevelDBConfig{Path: tmpDir}
	cache, err := NewLevelDBCache(cfg)
	if err != nil {
		b.Fatalf("创建LevelDB缓存失败: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	testData := map[string]interface{}{
		"id":   12345,
		"name": "测试用户",
		"age":  25,
		"tags": []string{"tag1", "tag2", "tag3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		cache.Set(ctx, key, testData, time.Hour)
		
		var result map[string]interface{}
		cache.Get(ctx, key, &result)
	}
}

// BenchmarkLevelDBCache_ComplexStruct 基准测试复杂结构
func BenchmarkLevelDBCache_ComplexStruct(b *testing.B) {
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("leveldb_bench_%d", time.Now().UnixNano()))
	defer os.RemoveAll(tmpDir)

	cfg := LevelDBConfig{Path: tmpDir}
	cache, err := NewLevelDBCache(cfg)
	if err != nil {
		b.Fatalf("创建LevelDB缓存失败: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	
	type ComplexUser struct {
		ID       int                    `json:"id"`
		Name     string                 `json:"name"`
		Email    string                 `json:"email"`
		Age      int                    `json:"age"`
		Tags     []string               `json:"tags"`
		Metadata map[string]interface{} `json:"metadata"`
		Created  time.Time              `json:"created"`
	}

	testData := ComplexUser{
		ID:      12345,
		Name:    "测试用户",
		Email:   "test@example.com",
		Age:     25,
		Tags:    []string{"tag1", "tag2", "tag3", "tag4", "tag5"},
		Metadata: map[string]interface{}{
			"level":  10,
			"points": 1000,
			"status": "active",
		},
		Created: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_user_%d", i)
		cache.Set(ctx, key, testData, 0)
		
		var result ComplexUser
		cache.Get(ctx, key, &result)
	}
}

// BenchmarkMemoryCache_SetGet 基准测试内存缓存（对比参考）
func BenchmarkMemoryCache_SetGet(b *testing.B) {
	cache := NewMemoryCache()
	defer cache.Close()

	ctx := context.Background()
	testData := map[string]interface{}{
		"id":   12345,
		"name": "测试用户",
		"age":  25,
		"tags": []string{"tag1", "tag2", "tag3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		cache.Set(ctx, key, testData, time.Hour)
		
		var result map[string]interface{}
		cache.Get(ctx, key, &result)
	}
}

