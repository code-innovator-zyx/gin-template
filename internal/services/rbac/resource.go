package rbac

import (
	"context"
	"gin-admin/internal/model/rbac"
	"gin-admin/pkg/cache"
	_interface "gin-admin/pkg/interface"
	"gorm.io/gorm"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/12/3
* @Package: Resource Service
 */

// ResourceService 资源服务
type ResourceService struct {
	_interface.Service[*rbac.Resource]
}

func NewResourceService(db *gorm.DB, cache cache.ICache) *ResourceService {
	return &ResourceService{
		Service: *_interface.NewService[*rbac.Resource](db, cache),
	}
}
func (s *ResourceService) CheckUserPermission(ctx context.Context, userID uint, path string, method string) (bool, error) {
	// 直接检查 role_resources（不再查询 role_permissions）
	var count int64
	err := s.DB.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM resources res
		JOIN role_resources rr ON res.id = rr.resource_id
		JOIN user_roles ur ON rr.role_id = ur.role_id
		WHERE ur.user_id = ? AND res.path = ? AND res.method = ?
	`, userID, path, method).Scan(&count).Error

	return count > 0, err

}

// GetUserResources 获取用户可访问的资源列表（直接通过 role_resources）
func (s *ResourceService) GetUserResources(ctx context.Context, userID uint) ([]rbac.Resource, error) {
	var resources []rbac.Resource
	err := s.DB.WithContext(ctx).Raw(`
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
