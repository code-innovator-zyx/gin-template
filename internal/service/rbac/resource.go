package rbac

import (
	"context"
	"gin-admin/internal/model/rbac"
	"gin-admin/internal/service"
)

/*
* @Author: zouyx
* @Email: zouyx@knowsec.com
* @Date:   2025/11/16 下午2:06
* @Package: 资源相关操作
 */

type resourceService struct {
	ctx context.Context
	*service.BaseRepo[rbac.Resource]
}

func NewResourceService(ctx context.Context) *resourceService {
	return &resourceService{
		ctx:      ctx,
		BaseRepo: service.NewBaseRepo[rbac.Resource](ctx),
	}
}

func (s *resourceService) CheckUserPermission(userID uint, path string, method string) (bool, error) {
	// cache check First
	exist, err := service.GetCacheService().CheckUserPermission(s.ctx, userID, path, method, s.GetUserResources)
	if err == nil {
		return exist, nil
	}
	// 直接检查 role_resources（不再查询 role_permissions）
	var count int64
	err = s.Tx.Raw(`
		SELECT COUNT(*) FROM resources res
		JOIN role_resources rr ON res.id = rr.resource_id
		JOIN user_roles ur ON rr.role_id = ur.role_id
		WHERE ur.user_id = ? AND res.path = ? AND res.method = ?
	`, userID, path, method).Scan(&count).Error

	return count > 0, err

}

// GetUserResources 获取用户可访问的资源列表（直接通过 role_resources）
func (s *resourceService) GetUserResources(userID uint) ([]rbac.Resource, error) {
	var resources []rbac.Resource
	err := s.Tx.Raw(`
		SELECT DISTINCT res.* FROM resources res
		JOIN role_resources rr ON res.id = rr.resource_id
		JOIN user_roles ur ON rr.role_id = ur.role_id
		WHERE ur.user_id = ?
		ORDER BY res.path, res.method
	`, userID).Find(&resources).Error

	if err != nil {
		return nil, err
	}
	return resources, nil
}
