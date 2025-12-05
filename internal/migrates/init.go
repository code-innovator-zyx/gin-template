package migrates

import (
	"fmt"
	"gin-admin/internal/services"
	"github.com/sirupsen/logrus"
)

/*
 * @Author: zouyx
 * @Email: 1003941268@qq.com
 * @Date:   2025 2025/12/3 下午3:28
 * @Package: 自动化数据库迁移
 */

// Do 执行数据库迁移
// 会自动迁移所有已注册的模型
// 所有模型会通过各模块的 init() 函数自动注册到 registry
func Do(svcContext *services.ServiceContext) error {
	models := GetAllModels()

	if len(models) == 0 {
		logrus.Warn("no models registered for migration")
		return nil
	}

	logrus.Infof("migrating %d models...", len(models))

	if err := svcContext.Db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}

	logrus.Info("migration completed successfully")
	return nil
}

// DoGroup 按模块执行迁移（可选）
// 允许只迁移特定模块的表
func DoGroup(svcContext *services.ServiceContext, groups ...string) error {
	for _, group := range groups {
		models := GetGroupModels(group)
		if len(models) == 0 {
			logrus.Warnf("no models found in group: %s", group)
			continue
		}

		logrus.Infof("migrating group '%s' (%d models)...", group, len(models))

		if err := svcContext.Db.AutoMigrate(models...); err != nil {
			return fmt.Errorf("migrate group '%s' failed: %w", group, err)
		}
	}

	logrus.Info("group migration completed successfully")
	return nil
}

// ListGroups 列出所有已注册的模块（用于调试）
func ListGroups() {
	groups := GetAllGroups()
	logrus.Infof("registered groups: %v", groups)

	for _, group := range groups {
		models := GetGroupModels(group)
		logrus.Infof("  - %s: %d models", group, len(models))
	}
}
