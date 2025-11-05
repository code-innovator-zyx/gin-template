package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/04
* @Package: 内存缓存适配器
 */

// memoryCache 内存缓存实现
type memoryCache struct {
	data     map[string]*memoryCacheItem
	mu       sync.RWMutex
	stopChan chan struct{}
}

type memoryCacheItem struct {
	value    []byte
	expireAt time.Time
	isSet    bool                // 标识是否是集合
	setData  map[string]struct{} // 集合数据
}

// NewMemoryCache 创建内存缓存实例
func NewMemoryCache() Cache {
	cache := &memoryCache{
		data:     make(map[string]*memoryCacheItem),
		stopChan: make(chan struct{}),
	}

	// 启动过期清理
	go cache.cleanupExpired()

	logrus.Info("内存缓存初始化成功")
	return cache
}

// Get 获取缓存
func (m *memoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	m.mu.RLock()
	item, exists := m.data[key]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("key not found")
	}

	if !item.expireAt.IsZero() && time.Now().After(item.expireAt) {
		m.Delete(ctx, key)
		return fmt.Errorf("key expired")
	}

	return json.Unmarshal(item.value, dest)
}

// Set 设置缓存
func (m *memoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	item := &memoryCacheItem{
		value: data,
	}

	if ttl > 0 {
		item.expireAt = time.Now().Add(ttl)
	}

	m.data[key] = item
	return nil
}

// Delete 删除缓存
func (m *memoryCache) Delete(ctx context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}

// Exists 检查key是否存在
func (m *memoryCache) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	item, exists := m.data[key]
	m.mu.RUnlock()

	if !exists {
		return false, nil
	}

	if !item.expireAt.IsZero() && time.Now().After(item.expireAt) {
		m.Delete(ctx, key)
		return false, nil
	}

	return true, nil
}

// SAdd 添加集合成员
func (m *memoryCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.data[key]
	if !exists {
		item = &memoryCacheItem{
			isSet:   true,
			setData: make(map[string]struct{}),
		}
		m.data[key] = item
	}

	for _, member := range members {
		memberStr := fmt.Sprintf("%v", member)
		item.setData[memberStr] = struct{}{}
	}

	return nil
}

// SIsMember 检查是否是集合成员
func (m *memoryCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	m.mu.RLock()
	item, exists := m.data[key]
	m.mu.RUnlock()

	if !exists || !item.isSet {
		return false, nil
	}

	memberStr := fmt.Sprintf("%v", member)
	_, exists = item.setData[memberStr]
	return exists, nil
}

// SMembers 获取集合所有成员
func (m *memoryCache) SMembers(ctx context.Context, key string) ([]string, error) {
	m.mu.RLock()
	item, exists := m.data[key]
	m.mu.RUnlock()

	if !exists || !item.isSet {
		return []string{}, nil
	}

	members := make([]string, 0, len(item.setData))
	for member := range item.setData {
		members = append(members, member)
	}

	return members, nil
}

// Expire 设置过期时间
func (m *memoryCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.data[key]
	if !exists {
		return fmt.Errorf("key not found")
	}

	item.expireAt = time.Now().Add(ttl)
	return nil
}

// TTL 获取剩余过期时间
func (m *memoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	m.mu.RLock()
	item, exists := m.data[key]
	m.mu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("key not found")
	}

	if item.expireAt.IsZero() {
		return -1, nil // 永不过期
	}

	remaining := time.Until(item.expireAt)
	if remaining < 0 {
		return 0, nil
	}

	return remaining, nil
}

// Pipeline 创建管道
func (m *memoryCache) Pipeline() Pipeline {
	return &memoryPipeline{cache: m}
}

// Ping 测试连接
func (m *memoryCache) Ping(ctx context.Context) error {
	return nil
}

// Close 关闭连接
func (m *memoryCache) Close() error {
	close(m.stopChan)
	return nil
}

// Type 返回缓存类型
func (m *memoryCache) Type() string {
	return "memory"
}

// cleanupExpired 清理过期key
func (m *memoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for key, item := range m.data {
				if !item.expireAt.IsZero() && now.After(item.expireAt) {
					delete(m.data, key)
				}
			}
			m.mu.Unlock()
		case <-m.stopChan:
			return
		}
	}
}

// ================================
// Pipeline 实现
// ================================

type memoryPipeline struct {
	cache   *memoryCache
	results struct {
		exists   int64
		isMember bool
		expire   bool
	}
}

func (p *memoryPipeline) Exists(ctx context.Context, key string) ExistsCmd {
	exists, _ := p.cache.Exists(ctx, key)
	if exists {
		p.results.exists = 1
	}
	return &memoryExistsCmd{result: p.results.exists}
}

func (p *memoryPipeline) SIsMember(ctx context.Context, key string, member interface{}) BoolCmd {
	isMember, _ := p.cache.SIsMember(ctx, key, member)
	p.results.isMember = isMember
	return &memoryBoolCmd{result: isMember}
}

func (p *memoryPipeline) Expire(ctx context.Context, key string, ttl time.Duration) BoolCmd {
	err := p.cache.Expire(ctx, key, ttl)
	p.results.expire = err == nil
	return &memoryBoolCmd{result: p.results.expire}
}

func (p *memoryPipeline) Exec(ctx context.Context) error {
	return nil
}

type memoryExistsCmd struct {
	result int64
}

func (c *memoryExistsCmd) Result() (int64, error) {
	return c.result, nil
}

type memoryBoolCmd struct {
	result bool
}

func (c *memoryBoolCmd) Result() (bool, error) {
	return c.result, nil
}
