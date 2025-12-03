package cache

import (
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/04
* @Package: 缓存工厂 - 根据配置创建不同的缓存实现
 */

func NewCache(client *redis.Client) ICache {
	if client == nil {
		logrus.Info("未配置缓存，使用默认内存缓存")
		return NewMemoryCache()
	}
	cache, err := NewRedisCache(client)
	if err != nil {
		logrus.Errorf("初始化redis 缓存失败，降级为内存缓存:%v", err)
		return NewMemoryCache()
	}
	return cache
}
