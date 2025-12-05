package rbac

import (
	"cmp"
	"context"
	"errors"
	"gin-admin/internal/model/rbac"
	_interface "gin-admin/pkg/interface"
	"gorm.io/gorm"
	"maps"
	"slices"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025 2025/12/3 上午9:31
* @Package:
 */

// UserService 用户可以自己实现一些定制化的函数
type UserService struct {
	_interface.Service[rbac.User]
}

func NewUserService(db *gorm.DB, cache _interface.ICache) *UserService {
	return &UserService{
		Service: *_interface.NewService[rbac.User](db, cache),
	}
}

func (s *UserService) CheckAccountExist(ctx context.Context, username, email string) error {
	if exist, err := s.Exists(ctx, _interface.WithConditions(map[string]interface{}{"username": username})); err != nil {
		return err
	} else if exist {
		return errors.New("用户名已存在")
	}
	if exist, err := s.Exists(ctx, _interface.WithConditions(map[string]interface{}{"email": email})); err != nil {
		return err
	} else if exist {
		return errors.New("邮箱已存在")
	}
	return nil
}

func (s *UserService) GetUserPerms(ctx context.Context, userID uint) ([]rbac.Permission, error) {
	var rows []struct {
		PermissionID   uint   `json:"permission_id"`
		PermissionName string `json:"permission_name"`
		PermissionCode string `json:"permission_code"`
		ResourceID     uint   `json:"resource_id"`
		Path           string `json:"path"`
		Code           string `json:"code"`
		Method         string `json:"method"`
		Description    string `json:"description"`
	}
	err := s.DB.WithContext(ctx).Raw(`
		SELECT DISTINCT
			p.id   AS permission_id,
			p.name AS permission_name,
			p.code AS permission_code,
			res.id AS resource_id,
			res.path,
			res.method,
			res.code,
			res.description
		FROM permissions p
		JOIN resources res ON p.id = res.permission_id
		JOIN role_resources rr ON res.id = rr.resource_id
		JOIN user_roles ur ON rr.role_id = ur.role_id
		WHERE ur.user_id = ?
		ORDER BY p.code, res.path, res.method
	`, userID).Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	permMap := make(map[uint]rbac.Permission)
	for _, row := range rows {
		p, ok := permMap[row.PermissionID]
		if !ok {
			p = rbac.Permission{
				BaseModel: rbac.BaseModel{
					ID: row.PermissionID,
				},
				Name: row.PermissionName,
				Code: row.PermissionCode,
			}
			permMap[row.PermissionID] = p
		}
		p.Resources = append(p.Resources, rbac.Resource{
			BaseModel: rbac.BaseModel{
				ID: row.PermissionID,
			},
			Path:        row.Path,
			Method:      row.Method,
			Code:        row.Code,
			Description: row.Description,
		})
		permMap[row.PermissionID] = p
	}
	return slices.SortedFunc(maps.Values(permMap), func(permission rbac.Permission, permission2 rbac.Permission) int {
		return cmp.Compare(permission.ID, permission2.ID)
	}), nil
}
