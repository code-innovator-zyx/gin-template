package cache

import (
	"sync"

	"github.com/sirupsen/logrus"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/04
* @Package: 缓存工厂 - 根据配置创建不同的缓存实现（懒加载单例）
 */

var (
	// 缓存单例
	cacheOnce    sync.Once
	globalCache  ICache
	cacheInitErr error
)

// initCacheWithConfig 根据配置初始化缓存（内部使用）
func initCacheWithConfig(cfg *RedisConfig) (ICache, error) {
	// 如果没有配置，使用内存缓存
	if cfg == nil {
		logrus.Info("未配置缓存，使用默认内存缓存")
		return NewMemoryCache(), nil
	}
	cache, err := NewRedisCache(*cfg)
	if err != nil {
		logrus.Errorf("初始化redis 缓存失败，降级为内存缓存:%v", err)
		return NewMemoryCache(), nil
	}
	return cache, nil
}

// GetGlobalCache 获取全局缓存实例
func GetGlobalCache() ICache {
	if globalCache == nil {
		MustInitCache(nil)
	}
	return globalCache
}
func NewCache(cfg *RedisConfig) (ICache, error) {
	return initCacheWithConfig(cfg)
}

// MustInitCache 初始化缓存（无论如何都会成功，失败时自动降级到Memory）
// 推荐在应用启动时调用
func MustInitCache(cfg *RedisConfig) {
	cacheOnce.Do(func() {
		cache, err := initCacheWithConfig(cfg)
		if err != nil {
			logrus.Warnf("缓存初始化失败: %v，自动降级到内存缓存", err)
			globalCache = NewMemoryCache()
			logrus.Info("已切换到内存缓存")
		} else {
			globalCache = cache
		}
	})
}

// IsAvailable 检查缓存是否可用（别名，为了兼容性）
func IsAvailable() bool {
	return GetGlobalCache() != nil
}

// Close 关闭缓存连接
func Close() error {
	cache := GetGlobalCache()
	if cache != nil {
		return cache.Close()
	}
	return nil
}
