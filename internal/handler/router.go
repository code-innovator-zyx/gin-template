package handler

import (
	"context"
	"gin-template/internal/core"
	v1 "gin-template/internal/handler/v1"
	"gin-template/internal/middleware"
	"gin-template/internal/model/rbac"
	"gin-template/internal/routegroup"
	"gin-template/internal/service"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	r := gin.New()

	// 添加中间件（按顺序）
	r.Use(middleware.Recovery())  // panic恢复
	r.Use(middleware.RequestID()) // 请求ID追踪
	r.Use(middleware.Logger())    // 日志记录
	r.Use(middleware.Cors())      // 跨域处理

	// 注册API路由
	v1.RegisterRoutes(r)

	// 自动初始化 RBAC 权限系统
	// 企业级自动化设计：
	// 1. 自动创建默认权限组（基于代码中的声明）
	// 2. 自动同步路由资源到数据库
	// 3. 自动将资源绑定到对应的权限组
	// 4. 自动创建超级管理员角色和默认管理员账号
	// 5. 幂等操作：重复执行不会产生副作用
	protectedRoutes := convertRoutes(routegroup.GetProtectedRoutes())
	rbacConfig := buildRBACConfig()
	// 使用 Service 层来处理 RBAC 初始化业务逻辑（应用启动时使用 Background context）
	if err := service.GetRbacService().InitializeRBAC(context.Background(), protectedRoutes, rbacConfig); err != nil {
		logrus.Fatalf("RBAC 权限系统初始化失败: %v", err)
	}

	return r
}

// convertRoutes 转换路由格式（从 routegroup 到 service 层）
func convertRoutes(routes []routegroup.ProtectedRoute) []service.ProtectedRoute {
	result := make([]service.ProtectedRoute, len(routes))
	for i, route := range routes {
		result[i] = service.ProtectedRoute{
			Resource: rbac.Resource{
				Path:        route.Resource.Path,
				Method:      route.Resource.Method,
				Description: route.Description,
			},
			PermissionCode: route.PermissionCode,
			PermissionName: route.PermissionName,
			Description:    route.Description,
		}
	}
	return result
}

// buildRBACConfig 从配置文件构建 RBAC 初始化配置
func buildRBACConfig() *service.RBACInitConfig {
	cfg := core.MustGetConfig()

	// 如果没有配置 RBAC 或未启用自动初始化，返回 nil 使用默认配置
	if cfg.RBAC == nil {
		return nil
	}

	return &service.RBACInitConfig{
		AdminUsername:  cfg.RBAC.AdminUser.Username,
		AdminPassword:  cfg.RBAC.AdminUser.Password,
		AdminEmail:     cfg.RBAC.AdminUser.Email,
		AdminRoleName:  cfg.RBAC.AdminRole.Name,
		AdminRoleDesc:  cfg.RBAC.AdminRole.Description,
		EnableAutoInit: cfg.RBAC.EnableAutoInit,
	}
}
