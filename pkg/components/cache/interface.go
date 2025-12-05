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
* @Package: 基础缓存定义(系统依赖的基础缓存)
 */

// ErrKeyNotFound 缓存键不存在错误（包括已过期的键）
var ErrKeyNotFound = errors.New("cache: key not found")
var ErrUnreachable = errors.New("cache: cache unreachable")

// ICache 缓存接口
type ICache interface {
	// 基础操作
	Get(ctx context.Context, key string, dest interface{}) error
	// Set 存储缓存值
	// 警告：如果 value 是指针/切片/map 等引用类型，外部修改会影响缓存！
	// 传入的须是值类型，不要是指针
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)

	// 集合操作（用于权限缓存等场景）
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SRem(ctx context.Context, key string, members ...interface{}) error
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)
	SMembers(ctx context.Context, key string) ([]string, error)

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

	// 移除前缀key
	DeletePrefix(ctx context.Context, prefix string) error
}

// Pipeline 管道操作接口
type Pipeline interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) StatusCmd
	Exists(ctx context.Context, key string) IntCmd
	SAdd(ctx context.Context, key string, members ...interface{}) IntCmd
	Del(ctx context.Context, keys ...string) IntCmd
	SRem(ctx context.Context, key string, members ...interface{}) IntCmd
	SIsMember(ctx context.Context, key string, member interface{}) BoolCmd
	Expire(ctx context.Context, key string, ttl time.Duration) BoolCmd
	Exec(ctx context.Context) error
}

// BoolCmd 布尔命令结果
type BoolCmd interface {
	Result() (bool, error)
}
type StatusCmd interface {
	Result() (string, error)
}

type IntCmd interface {
	Result() (int64, error)
}
