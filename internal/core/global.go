package core

import (
	"context"
	"gin-admin/internal/config"
	"gin-admin/pkg/components/orm"
	"log"
	"sync"

	"gorm.io/gorm"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/10/30 下午4:37
* @Package: 全局变量和单例管理
 */

var (
	// 配置单例
	configOnce sync.Once
	appConfig  *config.AppConfig
	configErr  error

	// 数据库单例
	dbOnce sync.Once
	db     *gorm.DB
	dbErr  error

	migrateFuncs []func(*gorm.DB) error // 存储所有迁移函数
	migrateMu    sync.Mutex
)

// ================================
// 配置相关
// ================================

// GetConfig 获取全局配置（懒加载单例，线程安全）
func GetConfig() (*config.AppConfig, error) {
	configOnce.Do(func() {
		appConfig, configErr = config.Init()
	})
	return appConfig, configErr
}

// MustGetConfig 获取全局配置（失败则 panic）
func MustGetConfig() *config.AppConfig {
	cfg, err := GetConfig()
	if err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}
	return cfg
}

// ================================
// 数据库相关
// ================================

// GetDb 获取数据库连接（懒加载单例，线程安全）
func GetDb() (*gorm.DB, error) {
	dbOnce.Do(func() {
		cfg := MustGetConfig()
		if cfg.Database == nil {
			dbErr = nil // 数据库未配置，不是错误
			return
		}

		db, dbErr = orm.Init(*cfg.Database)
	})
	return db, dbErr
}

// MustGetDb 获取数据库连接（失败则 panic）
func MustGetDb() *gorm.DB {
	database, err := GetDb()
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	if database == nil {
		log.Fatal("数据库未配置")
	}
	return database
}

// MustNewDb 获取全局数据库连接（创建新 Session）
func MustNewDb() *gorm.DB {
	database := MustGetDb()
	return database.Session(&gorm.Session{})
}

// MustNewDbWithContext 获取带上下文的数据库连接
func MustNewDbWithContext(ctx context.Context) *gorm.DB {
	database := MustGetDb()
	return database.Session(&gorm.Session{}).WithContext(ctx)
}

// ================================
// 数据库迁移
// ================================

// RegisterMigrate 注册数据库迁移函数
func RegisterMigrate(fn func(*gorm.DB) error) {
	migrateMu.Lock()
	defer migrateMu.Unlock()
	migrateFuncs = append(migrateFuncs, fn)
}

// ExecuteMigrations 执行所有已注册的迁移函数
func ExecuteMigrations() error {
	database := MustGetDb()

	migrateMu.Lock()
	funcs := make([]func(*gorm.DB) error, len(migrateFuncs))
	copy(funcs, migrateFuncs)
	migrateMu.Unlock()

	for _, fn := range funcs {
		if err := fn(database); err != nil {
			return err
		}
	}
	return nil
}
