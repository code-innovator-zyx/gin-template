package core

import (
	"gin-admin/pkg/cache"
	"gin-admin/pkg/logger"
	"log"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/10/30 下午4:19
* @Package: 应用初始化（使用懒加载单例模式）
 */

// Init 初始化应用（非必须调用，各模块支持懒加载）
// 如果希望在启动时提前初始化和检查配置，可以调用此函数
func Init() {
	// 初始化配置
	cfg := MustGetConfig()
	// 初始化日志
	logger.Init(cfg.Logger)
	// 初始化数据
	if cfg.Database != nil {
		// 执行所有已注册的迁移函数
		if err := ExecuteMigrations(); err != nil {
			log.Fatalf("数据库迁移失败: %v", err)
		}
		log.Printf("数据库初始化成功: %s", cfg.Database.DSN)
	}
	// 初始化缓存（必选，未配置时默认使用内存缓存）
	cache.MustInitCache(cfg.Cache)
	//log.Printf("缓存初始化成功: %s", cache.GetType())
}
