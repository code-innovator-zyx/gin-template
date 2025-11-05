package cache

import (
	"context"
	"errors"
	"time"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/04
* @Package: 缓存接口定义
 */

// ErrKeyNotFound 缓存键不存在错误（包括已过期的键）
var ErrKeyNotFound = errors.New("cache: key not found")

// Cache 缓存接口 - 支持多种实现（Redis、LevelDB、Memory等）
type Cache interface {
	// 基础操作
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)

	// 集合操作（用于权限缓存等场景）
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SRem(ctx context.Context, key string, members ...interface{}) error
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)
	SMembers(ctx context.Context, key string) ([]interface{}, error)
	
	// 计数器操作
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)

	// TTL管理
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// 批量操作
	Pipeline() Pipeline

	// 连接管理
	Ping(ctx context.Context) error
	Close() error

	// 类型标识
	Type() string
}

// Pipeline 管道操作接口
type Pipeline interface {
	Exists(ctx context.Context, key string) ExistsCmd
	SIsMember(ctx context.Context, key string, member interface{}) BoolCmd
	Expire(ctx context.Context, key string, ttl time.Duration) BoolCmd
	Exec(ctx context.Context) error
}

// ExistsCmd Exists命令结果
type ExistsCmd interface {
	Result() (int64, error)
}

// BoolCmd 布尔命令结果
type BoolCmd interface {
	Result() (bool, error)
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Type    string         `mapstructure:"type" validate:"required,oneof=redis leveldb memory"`
	Redis   *RedisConfig   `mapstructure:"redis"`
	LevelDB *LevelDBConfig `mapstructure:"leveldb"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db" validate:"min=0"`
	PoolSize int    `mapstructure:"pool_size" validate:"omitempty,min=1"`
}

// LevelDBConfig LevelDB配置
type LevelDBConfig struct {
	Path string `mapstructure:"path" validate:"required"`
}
