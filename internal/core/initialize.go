package core

import (
	"gin-template/pkg/cache"
	"gin-template/pkg/logger"
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
	// 初始化配置（懒加载，自动触发）
	cfg := MustGetConfig()
	
	// 初始化日志
	logger.Init(cfg.Logger)
	
	// 初始化数据库（懒加载，仅在有配置时初始化）
	if cfg.Database != nil {
		database := MustGetDb() // 触发数据库初始化
		
		// 执行所有已注册的迁移函数
		if err := ExecuteMigrations(); err != nil {
			log.Fatalf("数据库迁移失败: %v", err)
		}
		
		log.Printf("数据库初始化成功: %s", cfg.Database.DSN)
		_ = database // 避免未使用警告
	}
	
	// 初始化缓存（必选，未配置时默认使用内存缓存）
	cache.MustInitCache(cfg.Cache)
	log.Printf("缓存初始化成功: %s", cache.GetType())
}

// QuickInit 快速初始化（推荐）
// 仅初始化必要的组件，其他组件使用懒加载
func QuickInit() {
	// 仅初始化配置和日志
	cfg := MustGetConfig()
	logger.Init(cfg.Logger)
	
	log.Println("应用快速初始化完成（使用懒加载模式）")
}
