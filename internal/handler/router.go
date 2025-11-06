package handler

import (
	v1 "gin-template/internal/handler/v1"
	"gin-template/internal/middleware"
	"gin-template/internal/model/rbac"
	"gin-template/internal/routegroup"

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
	if err := rbac.InitializeRBAC(protectedRoutes); err != nil {
		logrus.Fatalf("RBAC 权限系统初始化失败: %v", err)
	}

	return r
}

// convertRoutes 转换路由格式
func convertRoutes(routes []routegroup.ProtectedRoute) []rbac.ProtectedRoute {
	result := make([]rbac.ProtectedRoute, len(routes))
	for i, route := range routes {
		result[i] = rbac.ProtectedRoute{
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
