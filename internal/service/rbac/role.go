package rbac

import (
	"context"
	"fmt"
	"gin-admin/internal/model/rbac"
	"gin-admin/internal/service"
	types "gin-admin/internal/types/rbac"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/16 下午2:06
* @Package: 角色相关操作
 */

type roleService struct {
	ctx context.Context
	*service.BaseRepo[rbac.Role]
}

func NewRoleService(ctx context.Context) *roleService {
	return &roleService{
		ctx:      ctx,
		BaseRepo: service.NewBaseRepo[rbac.Role](ctx),
	}
}

func (s *roleService) List(request types.ListRoleRequest) (*service.PageResult[rbac.Role], error) {
	return s.BaseRepo.List(service.PageQuery{
		Page:     request.Page,
		PageSize: request.PageSize,
		OrderBy:  "created_at DESC",
	}, func(db *gorm.DB) *gorm.DB {
		if request.Name != "" {
			db = db.Where("name LIKE ?", request.Name+"%")
		}
		if request.Status > 0 {
			db = db.Where("status = ?", request.Status)
		}
		return db
	})
}

func (s *roleService) Create(request types.UpsertRoleRequest) error {
	exist, err := s.Exists(func(db *gorm.DB) *gorm.DB {
		return db.Where("name = ?", request.Name)
	})
	if err != nil {
		logrus.Errorf("failed to check role exist %s", err)
		return err
	}
	if exist {
		return fmt.Errorf("角色名 %s 已存在", request.Name)
	}
	role := rbac.Role{
		Name:        request.Name,
		Description: request.Description,
		Status:      request.Status,
	}
	return s.BaseRepo.Create(&role)
}

// Update 更新角色
func (s *roleService) Update(request types.UpsertRoleRequest) error {
	role, err := s.FindByID(request.Id)
	if err != nil {
		logrus.Errorf("failed to find role: %v", err)
		return err
	}
	if request.Name != "" && request.Name != role.Name {
		exist, err := s.Exists(func(db *gorm.DB) *gorm.DB {
			return db.Where("name = ? AND id <> ?", request.Name, request.Id)
		})
		if err != nil {
			logrus.Errorf("failed to check role exist %s", err)
			return err
		}
		if exist {
			return fmt.Errorf("角色名 %s 已存在", request.Name)
		}
	}
	// TODO 如果禁用角色，需要移除所有绑定了角色的用户权限缓存，并且限制所有用户权限
	return s.UpdateByID(request.Id, map[string]interface{}{
		"name":        request.Name,
		"description": request.Description,
		"status":      request.Status,
	})

}

// getRoleUserIds 获取绑定了该角色的所有用户ID
func (s *roleService) getRoleUserIds(roleId uint) ([]uint, error) {
	var userIDs []uint
	err := s.Tx.
		Table("user_roles").
		Select("user_id").
		Where("role_id = ?", roleId).
		Find(&userIDs).Error
	if err != nil {
		return nil, err
	}
	return userIDs, nil
}

// DeleteByID 删除记录
func (s *roleService) DeleteByID(id uint) error {
	role, err := s.FindByID(id)
	if err != nil {
		logrus.Errorf("failed to find role: %v", err)
		return err
	}
	err = s.Tx.Transaction(
		func(tx *gorm.DB) error {
			err = tx.Model(&role).Association("Resources").Clear()
			if err != nil {
				return err
			}
			return tx.Delete(role).Error
		})
	if nil != err {
		logrus.Errorf("failed to delete role: %v", err)
		return fmt.Errorf("删除角色失败")
	}
	return nil
}

func (s *roleService) AssignResource(request types.AssignResource) error {
	role, err := s.FindByID(request.Id)
	if err != nil {
		logrus.Errorf("failed to find role: %v", err)
		return err
	}
	resources, err := NewResourceService(s.ctx).FindByIDs(request.ResourceIds)
	if err != nil {
		logrus.Errorf("failed to find resources: %v", err)
		return err
	}
	// 给角色重新分配了资源，就要移除角色缓存
	roleIds, err := NewRoleService(s.ctx).getRoleUserIds(request.Id)
	if err != nil {
		logrus.Errorf("failed to find role user_ids: %v", err)
		return err
	}
	return service.GetCacheService().ClearMultipleUsersPermissions(s.ctx, roleIds, time.Millisecond*50, func() error {
		return s.Tx.Model(&role).Association("Resources").Replace(resources)
	})
}
