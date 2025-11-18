package model

import (
	"gin-admin/internal/core"
	"gin-admin/internal/model/rbac"
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
		&rbac.Resource{},
		//&rbac.UserRole{},
		//&rbac.RoleResource{},
	); err != nil {
		return err
	}
	logrus.Info("success migration")
	return nil
}
