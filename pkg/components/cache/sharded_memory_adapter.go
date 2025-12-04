package cache

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	_ "unsafe"
)

/*
 * @Author: zouyx
 * @Email: 1003941268@qq.com
 * @Date:   2025/12/04
 * @Package: 分片内存缓存适配器
 */

var DefaultShardCount = runtime.NumCPU() * 4

//go:linkname stringHash runtime.stringHash
func stringHash(s string, seed uintptr) uintptr

func fastHash(key string) uint64 {
	return uint64(stringHash(key, 0))
}

type shardCacheItem struct {
	key      string
	value    atomic.Value // 不序列化，直接存 interface{}
	expireAt int64        //unixNano 存储过期时间
	isSet    bool
	setData  map[string]struct{}
}

// memoryShard 内存缓存分片
type memoryShard struct {
	data map[uint64]*shardCacheItem
	mu   sync.RWMutex
}

// shardedMemoryCache 分片内存缓存实现
type shardedMemoryCache struct {
	shards    []*memoryShard
	shardMask uint64
	stopChan  chan struct{}
}

// NewShardedMemoryCache 创建分片内存缓存实例
func NewShardedMemoryCache(shardCount int) ICache {
	if shardCount <= 0 {
		shardCount = DefaultShardCount
	}

	// 确保分片数是 2 的幂次方
	shardCount = nextPowerOfTwo(shardCount)

	shards := make([]*memoryShard, shardCount)
	for i := 0; i < shardCount; i++ {
		shards[i] = &memoryShard{
			data: make(map[uint64]*shardCacheItem, 20),
		}
	}

	cache := &shardedMemoryCache{
		shards:    shards,
		shardMask: uint64(shardCount - 1),
		stopChan:  make(chan struct{}),
	}

	go cache.cleanupExpired()
	return cache
}

// getShard 根据 key 获取对应的分片
func (c *shardedMemoryCache) getShard(key string) *memoryShard {
	hash := fastHash(key)
	return c.shards[hash&c.shardMask]
}

// nextPowerOfTwo 计算 >= n 的最小 2 的幂
func nextPowerOfTwo(n int) int {
	if n <= 0 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	return n + 1
}

// Get 获取缓存：dest 必须是指针
func (c *shardedMemoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	shard := c.getShard(key)
	shard.mu.RLock()
	item, exists := shard.data[fastHash(key)]
	shard.mu.RUnlock()
	if !exists || item.key != key {
		return ErrKeyNotFound
	}

	// 检查过期
	if ttl := atomic.LoadInt64(&item.expireAt); ttl > 0 && time.Now().UnixNano() > ttl {
		_ = c.Delete(ctx, key)
		return ErrKeyNotFound
	}
	v := item.value.Load()
	// fast-path
	switch d := dest.(type) {
	case *string:
		if val, ok := v.(string); ok {
			*d = val
			return nil
		}
	case *int:
		if val, ok := v.(int); ok {
			*d = val
			return nil
		}
	case *int64:
		if val, ok := v.(int64); ok {
			*d = val
			return nil
		}
	}
	// dest 必须是指针，且可设置
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}
	rv = rv.Elem()
	iv := reflect.ValueOf(v)
	if iv.Type().AssignableTo(rv.Type()) {
		rv.Set(iv)
		return nil
	}
	// 类型不兼容
	return fmt.Errorf("type mismatch: cannot assign %T to %T", v, dest)
}

// Set 直接存储 interface{}
func (c *shardedMemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	item := &shardCacheItem{
		key: key,
	}
	item.value.Store(value)
	if ttl > 0 {
		item.expireAt = time.Now().Add(ttl).UnixNano()
	}
	h := fastHash(key)
	item.key = key
	shard.data[h] = item
	return nil
}

// Delete 删除缓存
func (c *shardedMemoryCache) Delete(ctx context.Context, keys ...string) error {
	shardKeys := make(map[*memoryShard][]string)
	for _, key := range keys {
		shard := c.getShard(key)
		shardKeys[shard] = append(shardKeys[shard], key)
	}

	for shard, keys := range shardKeys {
		shard.mu.Lock()
		for _, key := range keys {
			h := fastHash(key)
			if item, exists := shard.data[h]; exists && item.key == key {
				delete(shard.data, h)
			}
		}
		shard.mu.Unlock()
	}

	return nil
}

// DeletePrefix 删除指定前缀的 key
func (c *shardedMemoryCache) DeletePrefix(ctx context.Context, prefix string) error {
	if prefix == "" {
		return nil
	}

	for _, shard := range c.shards {
		shard.mu.Lock()
		for key, item := range shard.data {
			if strings.HasPrefix(item.key, prefix) {
				delete(shard.data, key)
			}
		}
		shard.mu.Unlock()
	}

	return nil
}

// Exists 检查 key 是否存在
func (c *shardedMemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	shard := c.getShard(key)
	shard.mu.RLock()
	item, exists := shard.data[fastHash(key)]
	shard.mu.RUnlock()

	if !exists || item.key != key {
		return false, nil
	}
	if ttl := atomic.LoadInt64(&item.expireAt); ttl > 0 && time.Now().UnixNano() > ttl {
		_ = c.Delete(ctx, key)
		return false, nil
	}
	return true, nil
}

// SAdd 添加集合成员
func (c *shardedMemoryCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	h := fastHash(key)
	item, exists := shard.data[h]
	if !exists || item.key != key {
		item = &shardCacheItem{
			key:     key,
			isSet:   true,
			setData: make(map[string]struct{}),
		}
		shard.data[h] = item
	} else if !item.isSet {
		return fmt.Errorf("key %s exists but is not a set", key)
	}

	for _, member := range members {
		item.setData[fmt.Sprintf("%v", member)] = struct{}{}
	}

	return nil
}

// SIsMember 是否是集合成员
func (c *shardedMemoryCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	shard := c.getShard(key)
	shard.mu.RLock()
	item, exists := shard.data[fastHash(key)]
	shard.mu.RUnlock()

	if !exists || item.key != key || !item.isSet {
		return false, nil
	}

	_, ok := item.setData[fmt.Sprintf("%v", member)]
	return ok, nil
}

// SMembers 获取所有集合成员
func (c *shardedMemoryCache) SMembers(ctx context.Context, key string) ([]string, error) {
	shard := c.getShard(key)
	shard.mu.RLock()
	item, exists := shard.data[fastHash(key)]
	if !exists || !item.isSet || item.key != key {
		shard.mu.RUnlock()
		return nil, nil
	}

	members := make([]string, 0, len(item.setData))
	for m := range item.setData {
		members = append(members, m)
	}
	shard.mu.RUnlock()
	return members, nil
}

// SRem 移除集合成员
func (c *shardedMemoryCache) SRem(ctx context.Context, key string, members ...interface{}) error {
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	h := fastHash(key)
	item, exists := shard.data[h]
	if !exists || !item.isSet || item.key != key {
		return nil
	}

	for _, member := range members {
		delete(item.setData, fmt.Sprintf("%v", member))
	}
	// 如果 set 为空，删除整个 item
	if len(item.setData) == 0 {
		delete(shard.data, h)
	}
	return nil
}
func (c *shardedMemoryCache) incrBy(ctx context.Context, key string, delta int64) (int64, error) {
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	h := fastHash(key)
	item, exists := shard.data[h]

	var count int64
	if exists && item.key == key && !item.isSet {
		v := item.value.Load()
		if v != nil {
			val, ok := v.(int64)
			if !ok {
				return 0, fmt.Errorf("value is not int64")
			}
			count = val
		}
	}

	count += delta

	if exists && item.key == key {
		item.value.Store(count)
	} else {
		newItem := &shardCacheItem{key: key}
		newItem.value.Store(count)
		shard.data[h] = newItem
	}

	return count, nil
}

// Incr 自增
func (c *shardedMemoryCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.incrBy(ctx, key, 1)
}

// Decr 自减
func (c *shardedMemoryCache) Decr(ctx context.Context, key string) (int64, error) {
	return c.incrBy(ctx, key, -1)
}

// Expire 设置过期
func (c *shardedMemoryCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	shard := c.getShard(key)
	shard.mu.Lock()
	item, exists := shard.data[fastHash(key)]
	shard.mu.Unlock()
	if !exists || item.key != key {
		return ErrKeyNotFound
	}

	atomic.StoreInt64(&item.expireAt, time.Now().Add(ttl).UnixNano())
	return nil
}

// TTL 获取剩余过期时间
func (c *shardedMemoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	shard := c.getShard(key)
	shard.mu.RLock()
	item, exists := shard.data[fastHash(key)]
	shard.mu.RUnlock()

	if !exists || item.key != key {
		return 0, ErrKeyNotFound
	}

	expireAtUnixNano := atomic.LoadInt64(&item.expireAt)
	if expireAtUnixNano == 0 {
		return -1, nil
	}

	expireAt := time.Unix(0, expireAtUnixNano)
	remaining := time.Until(expireAt)
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}

// Pipeline 创建管道
func (c *shardedMemoryCache) Pipeline() Pipeline {
	return &memoryPipeline{
		sharded: c,
		cmds:    make([]memoryPipelineCmd, 0),
		results: make([]interface{}, 0),
	}
}

func (c *shardedMemoryCache) Ping(ctx context.Context) error { return nil }

func (c *shardedMemoryCache) Close() error {
	close(c.stopChan)
	return nil
}

// cleanupExpired 定期清理过期 key
func (c *shardedMemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	const maxCleanupPerShard = 100

	for {
		select {
		case <-ticker.C:
			var totalCleaned atomic.Int64
			var wg sync.WaitGroup

			// 并发清理每个分片
			for _, shard := range c.shards {
				wg.Add(1)
				go func(s *memoryShard) {
					defer wg.Done()
					cleaned := c.cleanupShard(s, maxCleanupPerShard)
					totalCleaned.Add(int64(cleaned))
				}(shard)
			}

			wg.Wait()

			if total := totalCleaned.Load(); total > 0 {
				logrus.Debugf("分片内存缓存清理了 %d 个过期 key", total)
			}

		case <-c.stopChan:
			return
		}
	}
}

func (c *shardedMemoryCache) cleanupShard(shard *memoryShard, maxCleanup int) int {
	// 1. 读锁收集
	shard.mu.RLock()
	now := time.Now().UnixNano()
	toDelete := make([][2]interface{}, 0, maxCleanup)

	for hash, item := range shard.data {
		if len(toDelete) >= maxCleanup {
			break
		}
		expireAtUnixNano := atomic.LoadInt64(&item.expireAt)
		if expireAtUnixNano > 0 && now > expireAtUnixNano {
			toDelete = append(toDelete, [2]interface{}{hash, item.key})
		}
	}
	shard.mu.RUnlock()

	if len(toDelete) == 0 {
		return 0
	}

	// 2. 写锁删除 + double-check
	shard.mu.Lock()
	defer shard.mu.Unlock()

	nowAgain := time.Now().UnixNano()
	cleaned := 0

	for _, pair := range toDelete {
		hash := pair[0].(uint64)
		key := pair[1].(string)

		if item, exists := shard.data[hash]; exists && item.key == key {
			expireAt := atomic.LoadInt64(&item.expireAt)
			if expireAt > 0 && nowAgain > expireAt {
				delete(shard.data, hash)
				cleaned++
			}
		}
	}
	return cleaned
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
	sharded *shardedMemoryCache // 支持分片缓存
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
			exists, _ := p.sharded.Exists(ctx, cmd.key)
			if exists {
				p.results[i] = int64(1)
			} else {
				p.results[i] = int64(0)
			}
		case "ismember":
			isMember, _ := p.sharded.SIsMember(ctx, cmd.key, cmd.member)
			p.results[i] = isMember

		case "expire":
			err := p.sharded.Expire(ctx, cmd.key, cmd.ttl)
			p.results[i] = err == nil

		case "set":
			err := p.sharded.Set(ctx, cmd.key, cmd.value, cmd.ttl)
			if err != nil {
				p.results[i] = err.Error()
			} else {
				p.results[i] = "OK"
			}

		case "sadd":
			err := p.sharded.SAdd(ctx, cmd.key, cmd.members...)
			if err != nil {
				p.results[i] = err
			} else {
				// 返回添加成员数
				p.results[i] = int64(len(cmd.members))
			}

		case "get":
			// cmd.value is dest interface{}
			err := p.sharded.Get(ctx, cmd.key, cmd.value)
			if err != nil {
				p.results[i] = err
			} else {
				p.results[i] = "OK"
			}
		case "del":
			err := p.sharded.Delete(ctx, cmd.keys...)
			if err != nil {
				p.results[i] = err
			} else {
				// 返回删除的数量
				p.results[i] = int64(len(cmd.keys))
			}
		case "srem":
			err := p.sharded.SRem(ctx, cmd.key, cmd.members...)
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
