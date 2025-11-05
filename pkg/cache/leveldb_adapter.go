package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
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

// cacheItem 缓存项（包含数据和过期时间）
type cacheItem struct {
	Data     json.RawMessage `json:"data"`
	ExpireAt int64           `json:"expire_at"` // Unix时间戳，0表示永不过期
}

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
			return fmt.Errorf("key not found")
		}
		return err
	}

	var item cacheItem
	if err := json.Unmarshal(data, &item); err != nil {
		return err
	}

	// 检查是否过期
	if item.ExpireAt > 0 && time.Now().Unix() > item.ExpireAt {
		l.db.Delete([]byte(key), nil)
		return fmt.Errorf("key expired")
	}

	return json.Unmarshal(item.Data, dest)
}

// Set 设置缓存
func (l *levelDBCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	var expireAt int64
	if ttl > 0 {
		expireAt = time.Now().Add(ttl).Unix()
		l.ttlMu.Lock()
		l.ttlMap[key] = time.Now().Add(ttl)
		l.ttlMu.Unlock()
	}

	item := cacheItem{
		Data:     data,
		ExpireAt: expireAt,
	}

	itemData, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return l.db.Put([]byte(key), itemData, nil)
}

// Delete 删除缓存
func (l *levelDBCache) Delete(ctx context.Context, keys ...string) error {
	batch := new(leveldb.Batch)
	for _, key := range keys {
		batch.Delete([]byte(key))
		l.ttlMu.Lock()
		delete(l.ttlMap, key)
		l.ttlMu.Unlock()
	}
	return l.db.Write(batch, nil)
}

// Exists 检查key是否存在
func (l *levelDBCache) Exists(ctx context.Context, key string) (bool, error) {
	_, err := l.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// SAdd 添加集合成员（LevelDB模拟实现）
func (l *levelDBCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	// 使用前缀存储集合成员
	for _, member := range members {
		memberKey := fmt.Sprintf("%s:member:%v", key, member)
		if err := l.Set(ctx, memberKey, true, 0); err != nil {
			return err
		}
	}
	return nil
}

// SIsMember 检查是否是集合成员
func (l *levelDBCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	memberKey := fmt.Sprintf("%s:member:%v", key, member)
	return l.Exists(ctx, memberKey)
}

// SMembers 获取集合所有成员
func (l *levelDBCache) SMembers(ctx context.Context, key string) ([]string, error) {
	prefix := []byte(fmt.Sprintf("%s:member:", key))
	iter := l.db.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()

	members := make([]string, 0)
	for iter.Next() {
		keyStr := string(iter.Key())
		// 提取member部分
		parts := strings.Split(keyStr, ":member:")
		if len(parts) == 2 {
			members = append(members, parts[1])
		}
	}

	return members, iter.Error()
}

// Expire 设置过期时间
func (l *levelDBCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	// 获取现有数据
	data, err := l.db.Get([]byte(key), nil)
	if err != nil {
		return err
	}

	var item cacheItem
	if err := json.Unmarshal(data, &item); err != nil {
		return err
	}

	// 更新过期时间
	item.ExpireAt = time.Now().Add(ttl).Unix()
	l.ttlMu.Lock()
	l.ttlMap[key] = time.Now().Add(ttl)
	l.ttlMu.Unlock()

	itemData, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return l.db.Put([]byte(key), itemData, nil)
}

// TTL 获取剩余过期时间
func (l *levelDBCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	data, err := l.db.Get([]byte(key), nil)
	if err != nil {
		return 0, err
	}

	var item cacheItem
	if err := json.Unmarshal(data, &item); err != nil {
		return 0, err
	}

	if item.ExpireAt == 0 {
		return -1, nil // 永不过期
	}

	remaining := time.Unix(item.ExpireAt, 0).Sub(time.Now())
	if remaining < 0 {
		return 0, nil // 已过期
	}

	return remaining, nil
}

// Pipeline 创建管道（LevelDB使用batch模拟）
func (l *levelDBCache) Pipeline() Pipeline {
	return &levelDBPipeline{
		db:      l.db,
		cache:   l,
		results: make(map[string]interface{}),
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

// cleanupExpired 清理过期key
func (l *levelDBCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.ttlMu.Lock()
			now := time.Now()
			for key, expireTime := range l.ttlMap {
				if now.After(expireTime) {
					l.db.Delete([]byte(key), nil)
					delete(l.ttlMap, key)
				}
			}
			l.ttlMu.Unlock()
		case <-l.stopChan:
			return
		}
	}
}

// ================================
// Pipeline 实现
// ================================

type levelDBPipeline struct {
	db      *leveldb.DB
	cache   *levelDBCache
	batch   leveldb.Batch
	results map[string]interface{}
	mu      sync.Mutex
}

func (p *levelDBPipeline) Exists(ctx context.Context, key string) ExistsCmd {
	p.mu.Lock()
	defer p.mu.Unlock()

	exists, _ := p.cache.Exists(ctx, key)
	var result int64
	if exists {
		result = 1
	}
	p.results[key+"_exists"] = result
	return &levelDBExistsCmd{result: result}
}

func (p *levelDBPipeline) SIsMember(ctx context.Context, key string, member interface{}) BoolCmd {
	p.mu.Lock()
	defer p.mu.Unlock()

	isMember, _ := p.cache.SIsMember(ctx, key, member)
	p.results[key+"_ismember"] = isMember
	return &levelDBBoolCmd{result: isMember}
}

func (p *levelDBPipeline) Expire(ctx context.Context, key string, ttl time.Duration) BoolCmd {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.cache.Expire(ctx, key, ttl)
	result := err == nil
	p.results[key+"_expire"] = result
	return &levelDBBoolCmd{result: result}
}

func (p *levelDBPipeline) Exec(ctx context.Context) error {
	return p.db.Write(&p.batch, nil)
}

// ================================
// Command 实现
// ================================

type levelDBExistsCmd struct {
	result int64
}

func (c *levelDBExistsCmd) Result() (int64, error) {
	return c.result, nil
}

type levelDBBoolCmd struct {
	result bool
}

func (c *levelDBBoolCmd) Result() (bool, error) {
	return c.result, nil
}
