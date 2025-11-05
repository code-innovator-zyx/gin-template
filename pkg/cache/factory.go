package cache

import (
	"fmt"
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
	globalCache  Cache
	cacheInitErr error
)

// initCacheWithConfig 根据配置初始化缓存（内部使用）
func initCacheWithConfig(cfg *CacheConfig) (Cache, error) {
	// 如果没有配置，使用内存缓存
	if cfg == nil {
		logrus.Info("未配置缓存，使用默认内存缓存")
		return NewMemoryCache(), nil
	}

	var cache Cache
	var err error

	switch cfg.Type {
	case "redis":
		if cfg.Redis == nil {
			return nil, fmt.Errorf("Redis配置不能为空")
		}
		cache, err = NewRedisCache(*cfg.Redis)
		if err != nil {
			return nil, fmt.Errorf("初始化Redis缓存失败: %w", err)
		}

	case "leveldb":
		if cfg.LevelDB == nil {
			return nil, fmt.Errorf("LevelDB配置不能为空")
		}
		cache, err = NewLevelDBCache(*cfg.LevelDB)
		if err != nil {
			return nil, fmt.Errorf("初始化LevelDB缓存失败: %w", err)
		}

	case "memory", "":
		cache = NewMemoryCache()

	default:
		return nil, fmt.Errorf("不支持的缓存类型: %s (支持: redis, leveldb, memory)", cfg.Type)
	}

	logrus.Infof("缓存初始化成功，类型: %s", cache.Type())
	return cache, nil
}

// GetGlobalCache 获取全局缓存实例
func GetGlobalCache() Cache {
	cacheOnce.Do(func() {
		// 这里需要从配置中获取缓存配置
		globalCache = NewMemoryCache()
		logrus.Info("缓存自动初始化为内存缓存（懒加载）")
	})
	return globalCache
}

// MustInitCache 初始化缓存（无论如何都会成功，失败时自动降级到Memory）
// 推荐在应用启动时调用
func MustInitCache(cfg *CacheConfig) {
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

// GetClient 获取全局缓存实例（别名，为了兼容性）
func GetClient() Cache {
	return GetGlobalCache()
}

// IsEnabled 检查缓存是否已启用
func IsEnabled() bool {
	return GetGlobalCache() != nil
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

// GetType 获取当前缓存类型
func GetType() string {
	cache := GetGlobalCache()
	if cache != nil {
		return cache.Type()
	}
	return "none"
}

// GlobalCache 全局缓存实例（向后兼容，已废弃）
// Deprecated: 直接使用 GetGlobalCache()
var GlobalCache Cache
