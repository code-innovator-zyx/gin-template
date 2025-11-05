package cache

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/04
* @Package: 缓存工厂 - 根据配置创建不同的缓存实现
 */

// GlobalCache 全局缓存实例
var GlobalCache Cache

// InitCache 初始化缓存（根据配置选择实现）
func InitCache(cfg CacheConfig) error {
	var (
		cache Cache
		err   error
	)

	switch cfg.Type {
	case "redis":
		if cfg.Redis == nil {
			return fmt.Errorf("Redis配置不能为空")
		}
		cache, err = NewRedisCache(*cfg.Redis)
		if err != nil {
			return fmt.Errorf("初始化Redis缓存失败: %w", err)
		}

	case "leveldb":
		if cfg.LevelDB == nil {
			return fmt.Errorf("LevelDB配置不能为空")
		}
		cache, err = NewLevelDBCache(*cfg.LevelDB)
		if err != nil {
			return fmt.Errorf("初始化LevelDB缓存失败: %w", err)
		}

	case "memory", "":
		cache = NewMemoryCache()

	default:
		return fmt.Errorf("不支持的缓存类型: %s (支持: redis, leveldb, memory)", cfg.Type)
	}

	GlobalCache = cache
	logrus.Infof("缓存初始化成功，类型: %s", cache.Type())
	return nil
}

// MustInitCache 初始化缓存（无论如何都会成功，失败时自动降级到Memory）
// 如果没有配置或配置的缓存初始化失败，将自动使用内存缓存
func MustInitCache(cfg *CacheConfig) {
	// 如果没有配置缓存，使用默认的内存缓存
	if cfg == nil {
		GlobalCache = NewMemoryCache()
		logrus.Info("未配置缓存，使用默认内存缓存")
		return
	}
	// 尝试初始化配置的缓存
	err := InitCache(*cfg)
	if err != nil {
		// 初始化失败，降级到内存缓存
		logrus.Warnf("缓存初始化失败: %v，自动降级到内存缓存", err)
		GlobalCache = NewMemoryCache()
		logrus.Info("已切换到内存缓存")
	}
}

// GetGlobalCache 获取全局缓存实例
func GetGlobalCache() Cache {
	return GlobalCache
}

// GetClient 获取全局缓存实例（别名，为了兼容性）
func GetClient() Cache {
	return GlobalCache
}

// IsEnabled 检查缓存是否已启用
func IsEnabled() bool {
	return GlobalCache != nil
}

// IsAvailable 检查缓存是否可用（别名，为了兼容性）
func IsAvailable() bool {
	return GlobalCache != nil
}

// Close 关闭缓存连接
func Close() error {
	if GlobalCache != nil {
		return GlobalCache.Close()
	}
	return nil
}

// GetType 获取当前缓存类型
func GetType() string {
	if GlobalCache != nil {
		return GlobalCache.Type()
	}
	return "none"
}
