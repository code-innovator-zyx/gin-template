package rbac

import (
	"gin-admin/internal/model/rbac"
	"gin-admin/pkg/cache"
	_interface "gin-admin/pkg/interface"
	"gorm.io/gorm"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/12/3
* @Package: Role Service
 */

// RoleService 角色服务
type RoleService struct {
	_interface.Service[rbac.Role]
}

func NewRoleService(db *gorm.DB, cache cache.ICache) *RoleService {
	return &RoleService{
		Service: *_interface.NewService[rbac.Role](db, cache),
	}
}

func (rs *RoleService) ListRoleUsers(roleId uint) (uids []uint, err error) {
	err = rs.DB.
		Table("user_roles").
		Select("user_id").
		Where("role_id = ?", roleId).
		Find(&uids).Error
	if err != nil {
		return nil, err
	}
	return uids, nil

}
