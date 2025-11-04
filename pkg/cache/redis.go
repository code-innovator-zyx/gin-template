package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gin-template/internal/config"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/10/31 下午6:26
* @Package:
 */

var RedisClient *redis.Client

// Init 初始化Redis连接
func Init(cfg config.RedisConfig) error {
	poolSize := cfg.PoolSize
	if poolSize == 0 {
		poolSize = 10 // 默认连接池大小
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: poolSize,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis连接失败: %w", err)
	}

	logrus.Infof("Redis连接成功: %s:%d", cfg.Host, cfg.Port)
	return nil
}

// Set 设置缓存（支持任意类型，自动JSON序列化）
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if RedisClient == nil {
		return fmt.Errorf("Redis客户端未初始化")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	return RedisClient.Set(ctx, key, data, expiration).Err()
}

// Get 获取缓存（自动JSON反序列化到目标对象）
func Get(ctx context.Context, key string, dest interface{}) error {
	if RedisClient == nil {
		return fmt.Errorf("Redis客户端未初始化")
	}

	data, err := RedisClient.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// Delete 删除缓存
func Delete(ctx context.Context, keys ...string) error {
	if RedisClient == nil {
		return fmt.Errorf("Redis客户端未初始化")
	}

	return RedisClient.Del(ctx, keys...).Err()
}

// Exists 检查key是否存在
func Exists(ctx context.Context, key string) (bool, error) {
	if RedisClient == nil {
		return false, fmt.Errorf("Redis客户端未初始化")
	}

	count, err := RedisClient.Exists(ctx, key).Result()
	return count > 0, err
}

// Close 关闭Redis连接
func Close() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}
