package cache

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/04
* @Package: LevelDB适配器实现
 */

// levelDBCache LevelDB缓存实现
type levelDBCache struct {
	db       *leveldb.DB
	ttlMap   map[string]time.Time // key过期时间映射
	ttlMu    sync.RWMutex
	stopChan chan struct{}
}

// 存储格式：[8字节过期时间戳][JSON数据]
// 前8字节存储过期时间戳（int64微秒），0表示永不过期
const expireAtSize = 8

// NewLevelDBCache 创建LevelDB缓存实例
func NewLevelDBCache(cfg LevelDBConfig) (Cache, error) {
	db, err := leveldb.OpenFile(cfg.Path, nil)
	if err != nil {
		return nil, fmt.Errorf("LevelDB打开失败: %w", err)
	}

	cache := &levelDBCache{
		db:       db,
		ttlMap:   make(map[string]time.Time),
		stopChan: make(chan struct{}),
	}

	// 启动过期清理协程
	go cache.cleanupExpired()

	logrus.Infof("LevelDB缓存初始化成功: %s", cfg.Path)
	return cache, nil
}

// Get 获取缓存
func (l *levelDBCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := l.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return ErrKeyNotFound
		}
		return err
	}

	// 数据格式：[8字节过期时间戳][JSON数据]
	if len(data) < expireAtSize {
		return fmt.Errorf("invalid data format")
	}

	expireAt := int64(binary.BigEndian.Uint64(data[:expireAtSize]))

	if expireAt > 0 && time.Now().UnixMicro() > expireAt {
		l.db.Delete([]byte(key), nil)
		l.ttlMu.Lock()
		delete(l.ttlMap, key)
		l.ttlMu.Unlock()
		return ErrKeyNotFound
	}
	return json.Unmarshal(data[expireAtSize:], dest)
}

// Set 设置缓存
func (l *levelDBCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	// 计算过期时间
	var expireAt int64
	if ttl > 0 {
		expireAt = time.Now().Add(ttl).UnixMicro()
		l.ttlMu.Lock()
		l.ttlMap[key] = time.Now().Add(ttl)
		l.ttlMu.Unlock()
	}

	// 构建存储数据：[8字节过期时间戳][JSON数据]
	data := make([]byte, expireAtSize+len(jsonData))
	binary.BigEndian.PutUint64(data[:expireAtSize], uint64(expireAt))
	copy(data[expireAtSize:], jsonData)

	return l.db.Put([]byte(key), data, nil)
}

// Delete 删除缓存（优化：减少锁操作）
func (l *levelDBCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	batch := new(leveldb.Batch)
	for _, key := range keys {
		batch.Delete([]byte(key))
	}

	err := l.db.Write(batch, nil)
	if err != nil {
		return err
	}

	// 批量删除ttlMap（减少锁操作次数）
	l.ttlMu.Lock()
	for _, key := range keys {
		delete(l.ttlMap, key)
	}
	l.ttlMu.Unlock()

	return nil
}

// Exists 检查key是否存在
func (l *levelDBCache) Exists(ctx context.Context, key string) (bool, error) {
	data, err := l.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	// 检查是否过期
	if len(data) >= expireAtSize {
		expireAt := int64(binary.BigEndian.Uint64(data[:expireAtSize]))
		if expireAt > 0 && time.Now().UnixMicro() > expireAt {
			l.db.Delete([]byte(key), nil)
			l.ttlMu.Lock()
			delete(l.ttlMap, key)
			l.ttlMu.Unlock()
			return false, nil
		}
	}

	return true, nil
}

// setType 集合类型标识
type setType struct {
	Members map[string]bool `json:"members"`
}

// SAdd 添加集合成员（优化：整个集合存为一个 key）
func (l *levelDBCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	// 获取现有集合
	var set setType
	err := l.Get(ctx, key, &set)
	if err != nil && err != ErrKeyNotFound {
		return err
	}

	// 初始化集合
	if set.Members == nil {
		set.Members = make(map[string]bool)
	}

	// 添加成员
	for _, member := range members {
		memberStr := fmt.Sprintf("%v", member)
		set.Members[memberStr] = true
	}

	// 保存集合（不设置TTL，由外部通过 Expire 管理）
	return l.Set(ctx, key, set, 0)
}

// SIsMember 检查是否是集合成员
func (l *levelDBCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	var set setType
	err := l.Get(ctx, key, &set)
	if err != nil {
		if err == ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}

	memberStr := fmt.Sprintf("%v", member)
	return set.Members[memberStr], nil
}

// SMembers 获取集合所有成员
func (l *levelDBCache) SMembers(ctx context.Context, key string) ([]interface{}, error) {
	var set setType
	err := l.Get(ctx, key, &set)
	if err != nil {
		if err == ErrKeyNotFound {
			return []interface{}{}, nil
		}
		return nil, err
	}

	members := make([]interface{}, 0, len(set.Members))
	for member := range set.Members {
		members = append(members, member)
	}

	return members, nil
}

// SRem 从集合中删除成员
func (l *levelDBCache) SRem(ctx context.Context, key string, members ...interface{}) error {
	var set setType
	err := l.Get(ctx, key, &set)
	if err != nil {
		if err == ErrKeyNotFound {
			return nil // 集合不存在，不需要删除
		}
		return err
	}

	// 删除成员
	for _, member := range members {
		memberStr := fmt.Sprintf("%v", member)
		delete(set.Members, memberStr)
	}

	// 如果集合为空，删除整个 key
	if len(set.Members) == 0 {
		return l.Delete(ctx, key)
	}

	// 保存更新后的集合
	return l.Set(ctx, key, set, 0)
}

// counterMu 计数器锁（确保原子性）
var counterMu sync.Mutex

// Incr 递增计数器（加锁保证原子性）
func (l *levelDBCache) Incr(ctx context.Context, key string) (int64, error) {
	counterMu.Lock()
	defer counterMu.Unlock()

	var count int64
	err := l.Get(ctx, key, &count)
	if err != nil && err != ErrKeyNotFound {
		return 0, err
	}

	count++
	if err := l.Set(ctx, key, count, 0); err != nil {
		return 0, err
	}

	return count, nil
}

// Decr 递减计数器（加锁保证原子性）
func (l *levelDBCache) Decr(ctx context.Context, key string) (int64, error) {
	counterMu.Lock()
	defer counterMu.Unlock()

	var count int64
	err := l.Get(ctx, key, &count)
	if err != nil && err != ErrKeyNotFound {
		return 0, err
	}

	count--
	if err := l.Set(ctx, key, count, 0); err != nil {
		return 0, err
	}

	return count, nil
}

// Expire 设置过期时间
func (l *levelDBCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	// 获取现有数据
	data, err := l.db.Get([]byte(key), nil)
	if err != nil {
		return err
	}

	// 数据格式：[8字节过期时间戳][JSON数据]
	if len(data) < expireAtSize {
		return fmt.Errorf("invalid data format")
	}

	// 更新过期时间（只修改前8字节）
	expireAt := time.Now().Add(ttl).UnixMicro()
	binary.BigEndian.PutUint64(data[:expireAtSize], uint64(expireAt))

	l.ttlMu.Lock()
	l.ttlMap[key] = time.Now().Add(ttl)
	l.ttlMu.Unlock()

	return l.db.Put([]byte(key), data, nil)
}

// TTL 获取剩余过期时间
func (l *levelDBCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	data, err := l.db.Get([]byte(key), nil)
	if err != nil {
		return 0, err
	}

	// 数据格式：[8字节过期时间戳][JSON数据]
	if len(data) < expireAtSize {
		return 0, fmt.Errorf("invalid data format")
	}

	// 读取过期时间戳
	expireAt := int64(binary.BigEndian.Uint64(data[:expireAtSize]))

	if expireAt == 0 {
		return -1, nil // 永不过期
	}

	remaining := time.Until(time.UnixMicro(expireAt))
	if remaining < 0 {
		return 0, nil // 已过期
	}

	return remaining, nil
}

// Pipeline 创建管道
func (l *levelDBCache) Pipeline() Pipeline {
	return &levelDBPipeline{
		cache:   l,
		cmds:    make([]pipelineCmd, 0),
		results: make([]interface{}, 0),
	}
}

// Ping 测试连接
func (l *levelDBCache) Ping(ctx context.Context) error {
	// LevelDB没有ping概念，尝试读取一个key
	_, err := l.db.Get([]byte("__ping__"), nil)
	if err != nil && err != leveldb.ErrNotFound {
		return err
	}
	return nil
}

// Close 关闭连接
func (l *levelDBCache) Close() error {
	close(l.stopChan)
	return l.db.Close()
}

// Type 返回缓存类型
func (l *levelDBCache) Type() string {
	return "leveldb"
}

// cleanupExpired 清理过期key（优化：批量删除 + 限制单次处理数量）
func (l *levelDBCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	const maxCleanupPerBatch = 1000 // 每批最多清理1000个key

	for {
		select {
		case <-ticker.C:
			l.ttlMu.RLock()
			now := time.Now()
			expiredKeys := make([]string, 0, maxCleanupPerBatch)

			// 收集过期的key（限制数量避免锁持有时间过长）
			for key, expireTime := range l.ttlMap {
				if now.After(expireTime) {
					expiredKeys = append(expiredKeys, key)
					if len(expiredKeys) >= maxCleanupPerBatch {
						break
					}
				}
			}
			l.ttlMu.RUnlock()

			// 批量删除过期key
			if len(expiredKeys) > 0 {
				batch := new(leveldb.Batch)
				for _, key := range expiredKeys {
					batch.Delete([]byte(key))
				}

				if err := l.db.Write(batch, nil); err != nil {
					logrus.Errorf("LevelDB清理过期key失败: %v", err)
					continue
				}

				// 从ttlMap中删除
				l.ttlMu.Lock()
				for _, key := range expiredKeys {
					delete(l.ttlMap, key)
				}
				l.ttlMu.Unlock()

				logrus.Debugf("LevelDB清理了%d个过期key", len(expiredKeys))
			}

		case <-l.stopChan:
			return
		}
	}
}

// ================================
// Pipeline 实现
// ================================

type pipelineCmd struct {
	cmdType string // "exists", "ismember", "expire"
	key     string
	member  interface{}
	ttl     time.Duration
}

type levelDBPipeline struct {
	cache   *levelDBCache
	cmds    []pipelineCmd
	results []interface{}
	mu      sync.Mutex
}

func (p *levelDBPipeline) Exists(ctx context.Context, key string) ExistsCmd {
	p.mu.Lock()
	defer p.mu.Unlock()

	idx := len(p.cmds)
	p.cmds = append(p.cmds, pipelineCmd{
		cmdType: "exists",
		key:     key,
	})
	p.results = append(p.results, nil)

	return &levelDBExistsCmd{pipeline: p, index: idx}
}

func (p *levelDBPipeline) SIsMember(ctx context.Context, key string, member interface{}) BoolCmd {
	p.mu.Lock()
	defer p.mu.Unlock()

	idx := len(p.cmds)
	p.cmds = append(p.cmds, pipelineCmd{
		cmdType: "ismember",
		key:     key,
		member:  member,
	})
	p.results = append(p.results, nil)

	return &levelDBBoolCmd{pipeline: p, index: idx}
}

func (p *levelDBPipeline) Expire(ctx context.Context, key string, ttl time.Duration) BoolCmd {
	p.mu.Lock()
	defer p.mu.Unlock()

	idx := len(p.cmds)
	p.cmds = append(p.cmds, pipelineCmd{
		cmdType: "expire",
		key:     key,
		ttl:     ttl,
	})
	p.results = append(p.results, nil)

	return &levelDBBoolCmd{pipeline: p, index: idx}
}

func (p *levelDBPipeline) Exec(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 执行所有命令
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
			p.results[i] = (err == nil)
		}
	}

	return nil
}

// ================================
// Command 实现
// ================================

type levelDBExistsCmd struct {
	pipeline *levelDBPipeline
	index    int
	result   int64
}

func (c *levelDBExistsCmd) Result() (int64, error) {
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

type levelDBBoolCmd struct {
	pipeline *levelDBPipeline
	index    int
	result   bool
}

func (c *levelDBBoolCmd) Result() (bool, error) {
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
