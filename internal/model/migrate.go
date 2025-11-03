package model

import (
	"gin-template/internal/core"
	"gin-template/internal/model/rbac"
	"gorm.io/gorm"
)

func init() {
	core.RegisterMigrate(autoMigrate)
}

// autoMigrate 自动迁移数据库表结构
func autoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&User{},
		&rbac.Role{},
		&rbac.Permission{},
		&rbac.Menu{},
		&rbac.UserRole{},
		&rbac.RolePermission{},
		&rbac.RoleMenu{},
	); err != nil {
		return err
	}
	return nil
}
