package rbac

import (
	"gin-template/internal/logic/v1/rbac"
	"gin-template/internal/middleware"
	"gin-template/internal/routegroup"
)

// RegisterRBACRoutes 注册RBAC相关路由
// 使用 SetPermission 声明路由组所属的权限组，系统会自动完成资源绑定
func RegisterRBACRoutes(api *routegroup.RouterGroup) {
	// 用户模块
	userGroup := api.Group("/user")
	{
		// 公共接口（不需要权限）
		userGroup.Public().POST("/register", rbac.Register)
		userGroup.Public().POST("/login", rbac.Login)
		userGroup.Public().POST("/refresh", rbac.RefreshToken) // 刷新令牌
		authGroup := userGroup.Group("")
		authGroup.Use(middleware.JWT())
		{
			authGroup.POST("/logout", rbac.Logout)
		}
		// 需要认证和权限 - 声明权限组
		authUserGroup := userGroup.WithMeta("user:manage", "用户管理")
		authUserGroup.Use(middleware.JWT(), middleware.PermissionMiddleware())
		{
			authUserGroup.GET("", rbac.ListUser).WithMeta("list", "查询用户列表")
			authUserGroup.GET("/profile", rbac.GetProfile).WithMeta("profile", "获取当前用户信息")
			authUserGroup.GET("/:id").WithMeta("detail", "查询用户详情")
			authUserGroup.POST("").WithMeta("create", "创建用户")
			authUserGroup.PUT("/:id").WithMeta("update", "修改用户")
			authUserGroup.GET("/:id/roles", rbac.GetUserRoles).WithMeta("roles", "获取用户角色")
			authUserGroup.POST("/:id/roles", rbac.AssignRoleToUser).WithMeta("assign-role", "分配角色给用户")
			authUserGroup.DELETE("/:id/roles/:role_id", rbac.RemoveRoleFromUser).WithMeta("remove-role", "移除用户角色")
		}
	}

	// 角色模块 - 声明权限组
	roleGroup := api.Group("/roles").WithMeta("role:manage", "角色管理")
	roleGroup.Use(middleware.JWT(), middleware.PermissionMiddleware())
	{
		roleGroup.GET("", rbac.GetRoles).WithMeta("list", "获取角色列表")
		roleGroup.POST("", rbac.CreateRole).WithMeta("create", "创建角色")
		roleGroup.GET("/:id", rbac.GetRole).WithMeta("detail", "获取角色详情")
		roleGroup.PUT("/:id", rbac.UpdateRole).WithMeta("update", "更新角色")
		roleGroup.DELETE("/:id", rbac.DeleteRole).WithMeta("delete", "删除角色")
		// 角色-资源绑定（实际授权）
		roleGroup.POST("/:id/resources/:resource_id", rbac.AssignResourceToRole).WithMeta("assign-resource", "分配资源给角色")
		roleGroup.DELETE("/:id/resources/:resource_id", rbac.RemoveResourceFromRole).WithMeta("remove-resource", "移除角色资源")
	}

	// 权限模块 - 声明权限组
	permissionGroup := api.Group("/permissions").WithMeta("permission:manage", "权限管理")
	permissionGroup.Use(middleware.JWT(), middleware.PermissionMiddleware())
	{
		permissionGroup.GET("", rbac.GetPermissions).WithMeta("list", "获取权限列表")
	}
}
