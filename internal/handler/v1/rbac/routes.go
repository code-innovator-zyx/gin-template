package rbac

import (
	"gin-template/internal/logic/v1/rbac"
	"gin-template/internal/middleware"
	"gin-template/internal/routegroup"
	"github.com/gin-gonic/gin"
)

// RegisterRBACRoutes 注册RBAC相关路由
// 使用 SetPermission 声明路由组所属的权限组，系统会自动完成资源绑定
func RegisterRBACRoutes(api *gin.RouterGroup) {
	// 认证模块（不需要登录）
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/refresh", rbac.RefreshToken) // 刷新令牌
	}

	// 认证模块（需要登录）- 声明权限组
	authAuthGroup := routegroup.WithAuthRouterGroup(api.Group("/auth")).
		SetPermission("auth:manage", "认证管理")
	authAuthGroup.Use(middleware.JWT())
	{
		authAuthGroup.POST("/logout", rbac.Logout) // 登出
	}

	// 用户模块
	userGroup := api.Group("/user")
	{
		// 公共接口（不需要权限）
		userGroup.POST("/register", rbac.Register)
		userGroup.POST("/login", rbac.Login)

		// 需要认证和权限 - 声明权限组
		authUserGroup := routegroup.WithAuthRouterGroup(userGroup.Group("/")).
			SetPermission("user:manage", "用户管理")
		authUserGroup.Use(middleware.JWT())
		{
			authUserGroup.GET("/profile", rbac.GetProfile)
			authUserGroup.GET("/:id/roles", rbac.GetUserRoles)
			authUserGroup.POST("/:id/roles", rbac.AssignRoleToUser)
			authUserGroup.DELETE("/:id/roles/:role_id", rbac.RemoveRoleFromUser)
		}
	}

	// 角色模块 - 声明权限组
	roleGroup := routegroup.WithAuthRouterGroup(api.Group("/roles")).
		SetPermission("role:manage", "角色管理")
	roleGroup.Use(middleware.JWT())
	{
		roleGroup.GET("", rbac.GetRoles)
		roleGroup.POST("", rbac.CreateRole)
		roleGroup.GET("/:id", rbac.GetRole)
		roleGroup.PUT("/:id", rbac.UpdateRole)
		roleGroup.DELETE("/:id", rbac.DeleteRole)

		// 角色-权限绑定
		roleGroup.POST("/:id/permissions", rbac.AssignPermissionToRole)
		roleGroup.DELETE("/:id/permissions/:permission_id", rbac.RemovePermissionFromRole)
	}

	// 权限模块 - 声明权限组
	permissionGroup := routegroup.WithAuthRouterGroup(api.Group("/permissions")).
		SetPermission("permission:manage", "权限管理")
	permissionGroup.Use(middleware.JWT())
	{
		permissionGroup.GET("", rbac.GetPermissions)
		permissionGroup.POST("", rbac.CreatePermission)
	}

	// 资源模块 - 声明权限组
	resourceGroup := routegroup.WithAuthRouterGroup(api.Group("/resources")).
		SetPermission("resource:view", "资源查看")
	resourceGroup.Use(middleware.JWT())
	{
		resourceGroup.GET("", rbac.GetResources)
	}
}
