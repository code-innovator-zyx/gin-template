package core

import (
	"gin-template/internal/config"
	"gin-template/pkg/cache"
	"gin-template/pkg/logger"
	"gin-template/pkg/orm"
	"log"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/10/30 下午4:19
* @Package:
 */

func Init() {
	// 初始化配置
	var err error
	Config, err = config.Init()
	if err != nil {
		log.Fatal(err)
	}
	// 初始化日志
	logger.Init(Config.Logger)
	// 初始化数据库
	if Config.Database != nil {
		db, err = orm.Init(*Config.Database)
		if err != nil {
			log.Fatal(err)
		}
		// 执行所有已注册的迁移函数
		for _, fn := range migrateFuncs {
			if err = fn(db); err != nil {
				log.Fatalf("数据库迁移失败: %v", err)
			}
		}
	}
	// 初始化Redis（可选）
	if Config.Redis != nil {
		if err = cache.Init(*Config.Redis); err != nil {
			log.Printf("Redis初始化失败（非致命错误）: %v", err)
		}
	}
}
