package rbac

import (
	"gin-admin/internal/logic/v1/rbac"
	"gin-admin/internal/middleware"
	"gin-admin/internal/routegroup"
)

// RegisterRBACRoutes 注册RBAC相关路由
// 使用 SetPermission 声明路由组所属的权限组，系统会自动完成资源绑定
func RegisterRBACRoutes(api *routegroup.RouterGroup) {
	// 用户模块
	userGroup := api.Group("/users")
	{
		// 公共接口（不需要权限，也不需要登录的jwt）
		userGroup.Public().POST("/register", rbac.Register)
		userGroup.Public().POST("/login", rbac.Login)
		userGroup.Public().POST("/refresh", rbac.RefreshToken) // 刷新令牌
		authGroup := userGroup.Group("")
		authGroup.Use(middleware.JWT())
		{
			// 需要登录但是不需要权限控制
			authGroup.POST("/logout", rbac.Logout)
			authGroup.GET("/options", rbac.UserOptions)
		}
		// 需要认证和权限 - 声明权限组
		authUserGroup := userGroup.WithMeta("user:manage", "用户管理")
		authUserGroup.Use(middleware.JWT(), middleware.PermissionMiddleware())
		{
			authUserGroup.GET("/profile", rbac.GetProfile).WithMeta("profile", "查询当前用户信息")
			authUserGroup.GET("", rbac.ListUser).WithMeta("list", "查询用户列表")
			authUserGroup.POST("", rbac.CreateUser).WithMeta("create", "创建用户")
			authUserGroup.PUT("/:id", rbac.UpdateUser).WithMeta("update", "编辑用户")
			authUserGroup.DELETE("/:id", rbac.DeleteUser).WithMeta("delete", "删除用户")
		}
	}

	// 角色模块 - 声明权限组
	roleGroup := api.Group("/roles").WithMeta("role:manage", "角色管理")
	roleGroup.Use(middleware.JWT(), middleware.PermissionMiddleware())
	{
		roleGroup.GET("", rbac.GetRoles).WithMeta("list", "查询角色列表")
		roleGroup.POST("", rbac.CreateRole).WithMeta("create", "创建角色")
		roleGroup.GET("/:id", rbac.GetRole).WithMeta("detail", "查询角色详情")
		roleGroup.PUT("/:id", rbac.UpdateRole).WithMeta("update", "编辑角色")
		roleGroup.DELETE("/:id", rbac.DeleteRole).WithMeta("delete", "删除角色")
		roleGroup.PUT("/:id/assign-resource", rbac.AssignRoleResources).WithMeta("assign-perm", "绑定资源权限")
	}

	// 权限模块 - 声明权限组
	permissionGroup := api.Group("/permissions").WithMeta("permission:manage", "权限管理")
	permissionGroup.Use(middleware.JWT(), middleware.PermissionMiddleware())
	{
		permissionGroup.GET("", rbac.GetPermissions).WithMeta("list", "获取权限列表")
	}
}
