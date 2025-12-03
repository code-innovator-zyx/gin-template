package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/04
* @Package: Redis适配器实现
 */

// RedisConfig 缓存配置
type RedisConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db" validate:"min=0"`
	PoolSize int    `mapstructure:"pool_size" validate:"omitempty,min=1"`
}

// redisCache Redis缓存实现
type redisCache struct {
	client *redis.Client
}

// NewRedisCache 创建Redis缓存实例
func NewRedisCache(cfg RedisConfig) (ICache, error) {
	poolSize := cfg.PoolSize
	if poolSize == 0 {
		poolSize = 10
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: poolSize,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis连接失败: %w", err)
	}

	logrus.Infof("Redis缓存初始化成功: %s:%d", cfg.Host, cfg.Port)
	return &redisCache{client: client}, nil
}

// Get 获取缓存
func (r *redisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrKeyNotFound
		}
		return err
	}
	return json.Unmarshal(data, dest)
}

// Set 设置缓存
func (r *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

// Delete 删除缓存
func (r *redisCache) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}
func (r *redisCache) DeletePrefix(ctx context.Context, prefix string) error {
	if r.client == nil || prefix == "" {
		return nil
	}

	luaScript := `
local keys = redis.call('keys', ARGV[1])
for i=1,#keys,5000 do
    redis.call('del', unpack(keys, i, math.min(i+4999, #keys)))
end
return #keys
`
	pattern := prefix + "*"
	return r.client.Eval(ctx, luaScript, []string{}, pattern).Err()
}

// Exists 检查key是否存在
func (r *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

// SAdd 添加集合成员
func (r *redisCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

// SIsMember 检查是否是集合成员
func (r *redisCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.client.SIsMember(ctx, key, member).Result()
}

// SMembers 获取集合所有成员
func (r *redisCache) SMembers(ctx context.Context, key string) ([]interface{}, error) {
	result, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	members := make([]interface{}, len(result))
	for i, v := range result {
		members[i] = v
	}
	return members, nil
}

// SRem 从集合中删除成员
func (r *redisCache) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

// Incr 递增计数器
func (r *redisCache) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Decr 递减计数器
func (r *redisCache) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

// Expire 设置过期时间
func (r *redisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

// TTL 获取剩余过期时间
func (r *redisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Pipeline 创建管道
func (r *redisCache) Pipeline() Pipeline {
	return &redisPipeline{pipe: r.client.Pipeline()}
}

func (r *redisCache) RedisClient() *redis.Client {
	return r.client
}

// Ping 测试连接
func (r *redisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close 关闭连接
func (r *redisCache) Close() error {
	return r.client.Close()
}

// GetClient 获取原始Redis客户端（用于高级操作）
func (r *redisCache) GetClient() *redis.Client {
	return r.client
}

// ================================
// Pipeline 实现
// ================================

type redisPipeline struct {
	pipe redis.Pipeliner
}

func (p *redisPipeline) SAdd(ctx context.Context, key string, members ...interface{}) IntCmd {
	return &redisIntCmd{cmd: p.pipe.SAdd(ctx, key, members...)}
}

func (p *redisPipeline) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) StatusCmd {
	return &redisStatusCmd{cmd: p.pipe.Set(ctx, key, value, ttl)}
}

func (p *redisPipeline) Exists(ctx context.Context, key string) IntCmd {
	return &redisIntCmd{cmd: p.pipe.Exists(ctx, key)}
}
func (p *redisPipeline) Del(ctx context.Context, keys ...string) IntCmd {
	return &redisIntCmd{cmd: p.pipe.Del(ctx, keys...)}
}
func (p *redisPipeline) SRem(ctx context.Context, key string, members ...interface{}) IntCmd {
	return &redisIntCmd{cmd: p.pipe.SRem(ctx, key, members)}
}
func (p *redisPipeline) SIsMember(ctx context.Context, key string, member interface{}) BoolCmd {
	return &redisBoolCmd{cmd: p.pipe.SIsMember(ctx, key, member)}
}

func (p *redisPipeline) Expire(ctx context.Context, key string, ttl time.Duration) BoolCmd {
	return &redisBoolCmd{cmd: p.pipe.Expire(ctx, key, ttl)}
}

func (p *redisPipeline) Exec(ctx context.Context) error {
	_, err := p.pipe.Exec(ctx)
	return err
}

// ================================
// Command 实现
// ================================

type redisBoolCmd struct {
	cmd *redis.BoolCmd
}

func (c *redisBoolCmd) Result() (bool, error) {
	return c.cmd.Result()
}

type redisStatusCmd struct {
	cmd *redis.StatusCmd
}

func (c *redisStatusCmd) Result() (string, error) {
	return c.cmd.Result()
}

type redisIntCmd struct {
	cmd *redis.IntCmd
}

func (c *redisIntCmd) Result() (int64, error) {
	return c.cmd.Result()
}
