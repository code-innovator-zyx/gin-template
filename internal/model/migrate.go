package model

import (
	"gin-template/internal/core"
	"gin-template/internal/model/rbac"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func init() {
	core.RegisterMigrate(autoMigrate)
}

// autoMigrate 自动迁移数据库表结构
func autoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&rbac.User{},
		&rbac.Role{},
		&rbac.Permission{},
		&rbac.UserRole{},
		&rbac.RolePermission{},
		&rbac.Resource{},
	); err != nil {
		return err
	}
	logrus.Info("success migration")
	return nil
}
