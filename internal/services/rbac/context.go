package rbac

import (
	_interface "gin-admin/pkg/interface"
	"gorm.io/gorm"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025 2025/12/5 下午6:19
* @Package:
 */

type Context struct {
	PermissionService *PermissionService
	RoleService       *RoleService
	ResourceService   *ResourceService
	UserService       *UserService
}

func NewContext(db *gorm.DB, cache _interface.ICache) *Context {
	return &Context{
		PermissionService: NewPermissionService(db, cache),
		RoleService:       NewRoleService(db, cache),
		ResourceService:   NewResourceService(db, cache),
		UserService:       NewUserService(db, cache),
	}
}
