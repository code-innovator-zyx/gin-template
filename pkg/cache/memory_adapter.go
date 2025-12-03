package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
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
func NewMemoryCache() ICache {
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
func (m *memoryCache) DeletePrefix(ctx context.Context, prefix string) error {
	if prefix == "" {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for key := range m.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(m.data, key)
		}
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

func (m *memoryCache) RedisClient() *redis.Client {
	panic("redis client not reachable")
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
	keys    []string
	value   interface{}   // 用于 set 或 get
	member  interface{}   // 用于集合操作
	ttl     time.Duration // 用于 set 或 expire
	members []interface{} // 用于 sadd 或 srem
}

type memoryPipeline struct {
	cache   *memoryCache
	cmds    []memoryPipelineCmd
	results []interface{}
	mu      sync.Mutex
}

func (p *memoryPipeline) Get(ctx context.Context, key string, dest interface{}) StatusCmd {
	p.mu.Lock()
	defer p.mu.Unlock()

	idx := len(p.cmds)
	p.cmds = append(p.cmds, memoryPipelineCmd{
		cmdType: "get",
		key:     key,
		value:   dest,
	})
	p.results = append(p.results, nil)

	return &memoryStatusCmd{pipeline: p, index: idx}
}

func (p *memoryPipeline) SAdd(ctx context.Context, key string, members ...interface{}) IntCmd {
	p.mu.Lock()
	defer p.mu.Unlock()

	idx := len(p.cmds)
	// 修复：使用 members 字段传递成员
	p.cmds = append(p.cmds, memoryPipelineCmd{
		cmdType: "sadd",
		key:     key,
		members: members,
	})
	p.results = append(p.results, nil)

	return &memoryIntCmd{pipeline: p, index: idx}
}
func (p *memoryPipeline) Del(ctx context.Context, keys ...string) IntCmd {
	p.mu.Lock()
	defer p.mu.Unlock()

	idx := len(p.cmds)
	p.cmds = append(p.cmds, memoryPipelineCmd{
		cmdType: "del",
		keys:    keys,
	})
	p.results = append(p.results, nil)

	return &memoryIntCmd{pipeline: p, index: idx}
}
func (p *memoryPipeline) SRem(ctx context.Context, key string, members ...interface{}) IntCmd {
	p.mu.Lock()
	defer p.mu.Unlock()

	idx := len(p.cmds)
	p.cmds = append(p.cmds, memoryPipelineCmd{
		cmdType: "srem",
		key:     key,
		members: members,
	})
	p.results = append(p.results, nil)

	return &memoryIntCmd{pipeline: p, index: idx}
}
func (p *memoryPipeline) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) StatusCmd {
	p.mu.Lock()
	defer p.mu.Unlock()

	idx := len(p.cmds)
	p.cmds = append(p.cmds, memoryPipelineCmd{
		cmdType: "set",
		key:     key,
		value:   value,
		ttl:     ttl,
	})
	p.results = append(p.results, nil)

	return &memoryStatusCmd{pipeline: p, index: idx}
}

func (p *memoryPipeline) Exists(ctx context.Context, key string) IntCmd {
	p.mu.Lock()
	defer p.mu.Unlock()

	idx := len(p.cmds)
	p.cmds = append(p.cmds, memoryPipelineCmd{
		cmdType: "exists",
		key:     key,
	})
	p.results = append(p.results, nil)

	return &memoryIntCmd{pipeline: p, index: idx}
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
			if exists {
				p.results[i] = int64(1)
			} else {
				p.results[i] = int64(0)
			}
		case "ismember":
			isMember, _ := p.cache.SIsMember(ctx, cmd.key, cmd.member)
			p.results[i] = isMember

		case "expire":
			err := p.cache.Expire(ctx, cmd.key, cmd.ttl)
			p.results[i] = err == nil

		case "set":
			err := p.cache.Set(ctx, cmd.key, cmd.value, cmd.ttl)
			if err != nil {
				p.results[i] = err.Error()
			} else {
				p.results[i] = "OK"
			}

		case "sadd":
			err := p.cache.SAdd(ctx, cmd.key, cmd.members...)
			if err != nil {
				p.results[i] = err
			} else {
				// 返回添加成员数
				p.results[i] = int64(len(cmd.members))
			}

		case "get":
			// cmd.value is dest interface{}
			err := p.cache.Get(ctx, cmd.key, cmd.value)
			if err != nil {
				p.results[i] = err
			} else {
				p.results[i] = "OK"
			}
		case "del":
			err := p.cache.Delete(ctx, cmd.keys...)
			if err != nil {
				p.results[i] = err
			} else {
				// 返回删除的数量
				p.results[i] = int64(len(cmd.keys))
			}
		case "srem":
			err := p.cache.SRem(ctx, cmd.key, cmd.members...)
			if err != nil {
				p.results[i] = err
			} else {
				// 返回删除的数量
				p.results[i] = int64(len(cmd.members))
			}
		}
	}

	return nil
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
			switch val := c.pipeline.results[c.index].(type) {
			case bool:
				return val, nil
			case error:
				return false, val
			}
		}
	}
	return c.result, nil
}

type memoryStatusCmd struct {
	pipeline *memoryPipeline
	index    int
	result   string
}

func (c *memoryStatusCmd) Result() (string, error) {
	if c.pipeline != nil {
		c.pipeline.mu.Lock()
		defer c.pipeline.mu.Unlock()

		if c.index < len(c.pipeline.results) && c.pipeline.results[c.index] != nil {
			switch val := c.pipeline.results[c.index].(type) {
			case string:
				return val, nil
			case error:
				return "", val
			}
		}
	}
	return c.result, nil
}

type memoryIntCmd struct {
	pipeline *memoryPipeline
	index    int
	result   int64
}

func (c *memoryIntCmd) Result() (int64, error) {
	if c.pipeline != nil {
		c.pipeline.mu.Lock()
		defer c.pipeline.mu.Unlock()

		if c.index < len(c.pipeline.results) && c.pipeline.results[c.index] != nil {
			switch val := c.pipeline.results[c.index].(type) {
			case int64:
				return val, nil
			case error:
				return 0, val
			}
		}
	}
	return c.result, nil
}
