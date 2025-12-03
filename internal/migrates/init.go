package migrates

import (
	"gin-admin/internal/model/rbac"
	"gin-admin/internal/services"
	"github.com/sirupsen/logrus"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025 2025/12/3 下午3:28
* @Package:
 */

func Do(svcContext *services.ServiceContext) error {
	if err := svcContext.Db.AutoMigrate(
		&rbac.User{},
		&rbac.Role{},
		&rbac.Permission{},
		&rbac.Resource{},
	); err != nil {
		return err
	}
	logrus.Info("success migration")
	return nil
}
