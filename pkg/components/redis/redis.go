package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025 2025/12/3 下午3:45
* @Package:
 */

// Config 缓存配置
type Config struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db" validate:"min=0"`
	PoolSize int    `mapstructure:"pool_size" validate:"omitempty,min=1"`
}

func NewClient(cfg Config) (*redis.Client, error) {
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
		return nil, err
	}
	return client, nil

}
