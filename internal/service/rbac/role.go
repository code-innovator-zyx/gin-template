package rbac

import (
	"context"
	"fmt"
	"gin-template/internal/core"
	"gin-template/internal/model/rbac"
	"gin-template/internal/service"
	types "gin-template/internal/types/rbac"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

/*
* @Author: zouyx
* @Email: zouyx@knowsec.com
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
		BaseRepo: service.NewBaseRepo[rbac.Role](ctx, core.MustNewDbWithContext(ctx)),
	}
}

func (s *roleService) List(request types.ListRoleRequest) (*service.PageResult[rbac.Role], error) {
	return s.BaseRepo.List(service.PageQuery{
		Page:     request.Page,
		PageSize: request.PageSize,
		OrderBy:  "-created_at",
	}, func(db *gorm.DB) *gorm.DB {
		if request.Name != "" {
			db = db.Where("name LIKE ?", request.Name+"%")
		}
		if request.Status > 0 {
			db = db.Where("status = ?", request.Status)
		}
		return db.Preload("Roles")
	})
}

func (s *roleService) Create(request types.CreateRoleRequest) error {
	exist, err := s.Exists(func(db *gorm.DB) *gorm.DB {
		return db.Where("name = ?", request.Name)
	})
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("角色名 %s 已存在", request.Name)
	}
	resources, err := NewResourceService(s.ctx).FindByIDs(request.Resources)
	if err != nil {
		return err
	}
	role := rbac.Role{
		Name:        request.Name,
		Description: request.Description,
		Resources:   resources,
	}
	return s.BaseRepo.Create(&role)
}

// Update 更新角色
func (s *roleService) Update(request types.UpdateRoleRequest) error {
	role, err := s.FindByID(request.Id)
	if err != nil {
		return err
	}
	if request.Name != "" && request.Name != role.Name {
		exist, err := s.Exists(func(db *gorm.DB) *gorm.DB {
			return db.Where("name = ? AND id <> ?", request.Name, request.Id)
		})
		if err != nil {
			return err
		}
		if exist {
			return fmt.Errorf("角色名 %s 已存在", request.Name)
		}
	}
	return s.Tx.Transaction(func(tx *gorm.DB) error {
		err = s.WithTx(tx).UpdateByID(request.Id, map[string]interface{}{
			"name":        request.Name,
			"description": request.Description,
		})
		if err != nil {
			return err
		}
		var resources []rbac.Resource
		// 获取所有角色列表
		if len(request.Resources) != 0 {
			resources, err = NewResourceService(s.ctx).FindByIDs(request.Resources)
			if err != nil {
				return err
			}
		}
		// 更新用户角色
		if err = tx.Model(&role).Association("Resources").Replace(resources); err != nil {
			return fmt.Errorf("更新用户角色失败: %w", err)
		}
		// 移除所有相关用户的权限
		// 事务提交后 清除所有拥有该角色的用户的权限缓存
		userIDs, err := s.getRoleUserIds(request.Id)
		if err != nil {
			logrus.Errorf("获取角色用户列表失败: %v", err)
			return err
		}
		if len(userIDs) > 0 {
			if err := service.GetCacheService().ClearMultipleUsersPermissions(s.ctx, userIDs); err != nil {
				logrus.Errorf("批量清除用户权限缓存失败: %v", err)
				return err
			}
			logrus.Infof("已清除 %d 个用户的权限缓存", len(userIDs))
		}
		return nil
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
		return err
	}
	return s.Tx.Transaction(
		func(tx *gorm.DB) error {
			err = tx.Model(&role).Association("Resources").Clear()
			if err != nil {
				return err
			}
			return tx.Delete(role).Error
		})
}
