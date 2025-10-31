package core

import (
	"context"
	"github.com/code-innovator-zyx/gin-template/config"
	"gorm.io/gorm"
	"log"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/10/30 下午4:37
* @Package:
 */
var (
	Config *config.AppConfig
	db     *gorm.DB
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
