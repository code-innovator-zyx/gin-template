package rbac

import (
	"gin-admin/internal/model/rbac"
	"gin-admin/pkg/components/cache"
	_interface "gin-admin/pkg/interface"
	"gorm.io/gorm"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/12/3
* @Package: Permission Service
 */

// PermissionService 权限组服务
type PermissionService struct {
	_interface.Service[rbac.Permission]
}

func NewPermissionService(db *gorm.DB, cache cache.ICache) *PermissionService {
	return &PermissionService{
		Service: *_interface.NewService[rbac.Permission](db, cache),
	}
}
