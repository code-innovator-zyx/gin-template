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
		return ErrKeyNotFound
	}

	if !item.expireAt.IsZero() && time.Now().After(item.expireAt) {
		m.Delete(ctx, key)
		return ErrKeyNotFound
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
func (m *memoryCache) SMembers(ctx context.Context, key string) ([]interface{}, error) {
	m.mu.RLock()
	item, exists := m.data[key]
	m.mu.RUnlock()

	if !exists || !item.isSet {
		return []interface{}{}, nil
	}

	members := make([]interface{}, 0, len(item.setData))
	for member := range item.setData {
		members = append(members, member)
	}

	return members, nil
}

// SRem 从集合中删除成员
func (m *memoryCache) SRem(ctx context.Context, key string, members ...interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.data[key]
	if !exists || !item.isSet {
		return nil
	}

	for _, member := range members {
		memberStr := fmt.Sprintf("%v", member)
		delete(item.setData, memberStr)
	}

	return nil
}

// Incr 递增计数器
func (m *memoryCache) Incr(ctx context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.data[key]
	var count int64

	if exists && !item.isSet {
		// 反序列化现有值
		if err := json.Unmarshal(item.value, &count); err != nil {
			return 0, fmt.Errorf("value is not a number")
		}
	}

	count++

	// 序列化新值
	data, err := json.Marshal(count)
	if err != nil {
		return 0, err
	}

	if exists {
		item.value = data
	} else {
		m.data[key] = &memoryCacheItem{
			value:    data,
			expireAt: time.Time{},
		}
	}

	return count, nil
}

// Decr 递减计数器
func (m *memoryCache) Decr(ctx context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.data[key]
	var count int64

	if exists && !item.isSet {
		// 反序列化现有值
		if err := json.Unmarshal(item.value, &count); err != nil {
			return 0, fmt.Errorf("value is not a number")
		}
	}

	count--

	// 序列化新值
	data, err := json.Marshal(count)
	if err != nil {
		return 0, err
	}

	if exists {
		item.value = data
	} else {
		m.data[key] = &memoryCacheItem{
			value:    data,
			expireAt: time.Time{},
		}
	}

	return count, nil
}

// Expire 设置过期时间
func (m *memoryCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.data[key]
	if !exists {
		return ErrKeyNotFound
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
		return 0, ErrKeyNotFound
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
	return &memoryPipeline{
		cache:   m,
		cmds:    make([]memoryPipelineCmd, 0),
		results: make([]interface{}, 0),
	}
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

// cleanupExpired 清理过期key（优化：读写锁分离）
func (m *memoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	const maxCleanupPerBatch = 1000

	for {
		select {
		case <-ticker.C:
			// 使用读锁收集过期的key
			m.mu.RLock()
			now := time.Now()
			expiredKeys := make([]string, 0, maxCleanupPerBatch)

			for key, item := range m.data {
				if !item.expireAt.IsZero() && now.After(item.expireAt) {
					expiredKeys = append(expiredKeys, key)
					if len(expiredKeys) >= maxCleanupPerBatch {
						break
					}
				}
			}
			m.mu.RUnlock()

			// 使用写锁批量删除
			if len(expiredKeys) > 0 {
				m.mu.Lock()
				for _, key := range expiredKeys {
					delete(m.data, key)
				}
				m.mu.Unlock()

				logrus.Debugf("内存缓存清理了%d个过期key", len(expiredKeys))
			}

		case <-m.stopChan:
			return
		}
	}
}

// ================================
// Pipeline 实现
// ================================

type memoryPipelineCmd struct {
	cmdType string
	key     string
	member  interface{}
	ttl     time.Duration
}

type memoryPipeline struct {
	cache   *memoryCache
	cmds    []memoryPipelineCmd
	results []interface{}
	mu      sync.Mutex
}

func (p *memoryPipeline) Exists(ctx context.Context, key string) ExistsCmd {
	p.mu.Lock()
	defer p.mu.Unlock()

	idx := len(p.cmds)
	p.cmds = append(p.cmds, memoryPipelineCmd{
		cmdType: "exists",
		key:     key,
	})
	p.results = append(p.results, nil)

	return &memoryExistsCmd{pipeline: p, index: idx}
}

func (p *memoryPipeline) SIsMember(ctx context.Context, key string, member interface{}) BoolCmd {
	p.mu.Lock()
	defer p.mu.Unlock()

	idx := len(p.cmds)
	p.cmds = append(p.cmds, memoryPipelineCmd{
		cmdType: "ismember",
		key:     key,
		member:  member,
	})
	p.results = append(p.results, nil)

	return &memoryBoolCmd{pipeline: p, index: idx}
}

func (p *memoryPipeline) Expire(ctx context.Context, key string, ttl time.Duration) BoolCmd {
	p.mu.Lock()
	defer p.mu.Unlock()

	idx := len(p.cmds)
	p.cmds = append(p.cmds, memoryPipelineCmd{
		cmdType: "expire",
		key:     key,
		ttl:     ttl,
	})
	p.results = append(p.results, nil)

	return &memoryBoolCmd{pipeline: p, index: idx}
}

func (p *memoryPipeline) Exec(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, cmd := range p.cmds {
		switch cmd.cmdType {
		case "exists":
			exists, _ := p.cache.Exists(ctx, cmd.key)
			var result int64
			if exists {
				result = 1
			}
			p.results[i] = result

		case "ismember":
			isMember, _ := p.cache.SIsMember(ctx, cmd.key, cmd.member)
			p.results[i] = isMember

		case "expire":
			err := p.cache.Expire(ctx, cmd.key, cmd.ttl)
			p.results[i] = err == nil
		}
	}

	return nil
}

type memoryExistsCmd struct {
	pipeline *memoryPipeline
	index    int
	result   int64
}

func (c *memoryExistsCmd) Result() (int64, error) {
	if c.pipeline != nil {
		c.pipeline.mu.Lock()
		defer c.pipeline.mu.Unlock()

		if c.index < len(c.pipeline.results) && c.pipeline.results[c.index] != nil {
			if val, ok := c.pipeline.results[c.index].(int64); ok {
				return val, nil
			}
		}
	}
	return c.result, nil
}

type memoryBoolCmd struct {
	pipeline *memoryPipeline
	index    int
	result   bool
}

func (c *memoryBoolCmd) Result() (bool, error) {
	if c.pipeline != nil {
		c.pipeline.mu.Lock()
		defer c.pipeline.mu.Unlock()

		if c.index < len(c.pipeline.results) && c.pipeline.results[c.index] != nil {
			if val, ok := c.pipeline.results[c.index].(bool); ok {
				return val, nil
			}
		}
	}
	return c.result, nil
}
