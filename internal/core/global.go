package core

import (
	"context"
	"gin-template/internal/config"
	"log"

	"gorm.io/gorm"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/10/30 下午4:37
* @Package:
 */
var (
	Config       *config.AppConfig
	db           *gorm.DB
	migrateFuncs []func(*gorm.DB) error // 存储所有迁移函数
)

// MustNewDb 获取全局数据库连接
func MustNewDb() *gorm.DB {
	if db == nil {
		log.Fatal("数据库未初始化")
	}
	return db.Session(&gorm.Session{})
}

func MustNewDbWithContext(ctx context.Context) *gorm.DB {
	if db == nil {
		log.Fatal("数据库未初始化")
	}
	return db.Session(&gorm.Session{}).WithContext(ctx)
}

// RegisterMigrate 注册数据库迁移函数
func RegisterMigrate(fn func(*gorm.DB) error) {
	migrateFuncs = append(migrateFuncs, fn)
}
