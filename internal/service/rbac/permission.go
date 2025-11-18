package rbac

import (
	"cmp"
	"context"
	"gin-admin/internal/model/rbac"
	"gin-admin/internal/service"
	"maps"
	"slices"
)

/*
* @Author: zouyx
* @Email: zouyx@knowsec.com
* @Date:   2025/11/16 下午2:06
* @Package: 资源相关操作
 */

type permissionService struct {
	ctx context.Context
	*service.BaseRepo[rbac.Permission]
}

func NewPermissionService(ctx context.Context) *permissionService {
	return &permissionService{
		ctx:      ctx,
		BaseRepo: service.NewBaseRepo[rbac.Permission](ctx),
	}
}

func (s *permissionService) GetUserPerms(userID uint) ([]rbac.Permission, error) {
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
	err := s.Tx.Raw(`
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
				ID:   row.PermissionID,
				Name: row.PermissionName,
				Code: row.PermissionCode,
			}
			permMap[row.PermissionID] = p
		}
		p.Resources = append(p.Resources, rbac.Resource{
			ID:          row.ResourceID,
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
