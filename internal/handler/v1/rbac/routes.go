package rbac

import (
	"gin-template/internal/logic/v1/rbac"
	"gin-template/internal/middleware"
	"gin-template/internal/routegroup"
	"github.com/gin-gonic/gin"
)

// RegisterRBACRoutes 注册RBAC相关路由
func RegisterRBACRoutes(api *gin.RouterGroup) {
	// 用户模块
	userGroup := api.Group("/user")
	{
		// 公共接口
		userGroup.POST("/register", rbac.Register)
		userGroup.POST("/login", rbac.Login)

		// 需要认证
		authUserGroup := routegroup.WithAuthRouterGroup(userGroup.Group("/"))
		authUserGroup.Use(middleware.JWT())
		{
			authUserGroup.GET("/profile", rbac.GetProfile)
			authUserGroup.GET("/:id/roles", rbac.GetUserRoles)
			authUserGroup.POST("/:id/roles", rbac.AssignRoleToUser)
			authUserGroup.DELETE("/:id/roles/:role_id", rbac.RemoveRoleFromUser)
		}
	}
	// 角色模块
	roleGroup := routegroup.WithAuthRouterGroup(api.Group("/roles"))
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

	// 权限模块
	permissionGroup := routegroup.WithAuthRouterGroup(api.Group("/permissions"))
	permissionGroup.Use(middleware.JWT())
	{
		permissionGroup.GET("", rbac.GetPermissions)
		permissionGroup.POST("", rbac.CreatePermission)
	}

	// 资源模块
	resourceGroup := routegroup.WithAuthRouterGroup(api.Group("/resources"))
	resourceGroup.Use(middleware.JWT())
	{
		resourceGroup.GET("", rbac.GetResources)
	}
}
