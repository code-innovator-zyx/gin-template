package services

import (
	"gin-admin/internal/config"
	rbac2 "gin-admin/internal/services/rbac"
	"gin-admin/pkg/cache"
	"gin-admin/pkg/components/orm"
	"gin-admin/pkg/jwt"
	"gorm.io/gorm"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025 2025/12/2 下午4:49
* @Package: Service Context - 统一管理所有 Service 实例
 */

var SvcContext *ServiceContext

// ServiceContext 服务上下文，包含所有业务 相关依赖(TODO 直接注入到logic)
type ServiceContext struct {
	// components
	Db    *gorm.DB
	Cache cache.ICache
	Jwt   jwt.Service
	// 服务缓存
	CacheService ICacheService
	// RBAC Services
	UserService       *rbac2.UserService
	RoleService       *rbac2.RoleService
	PermissionService *rbac2.PermissionService
	ResourceService   *rbac2.ResourceService
}

func MustInitServiceContext(c *config.AppConfig) *ServiceContext {
	db, err := orm.Init(*c.Database)
	if err != nil {
		panic(err)
	}

	// 初始化缓存
	cacheInstance, err := cache.NewCache(c.Cache)
	if err != nil {
		panic(err)
	}
	SvcContext = &ServiceContext{
		Db:                db,
		Cache:             cacheInstance,
		CacheService:      NewCacheService(cacheInstance),
		Jwt:               jwt.NewJwtService(*c.Jwt, cacheInstance),
		UserService:       rbac2.NewUserService(db, cacheInstance),
		RoleService:       rbac2.NewRoleService(db, cacheInstance),
		PermissionService: rbac2.NewPermissionService(db, cacheInstance),
		ResourceService:   rbac2.NewResourceService(db, cacheInstance),
	}
	return SvcContext
}
