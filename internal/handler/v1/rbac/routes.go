package rbac

import (
	"gin-admin/internal/logic/v1/rbac"
	"gin-admin/internal/middleware"
	"gin-admin/internal/routegroup"
	"gin-admin/internal/services"
)

// RegisterRBACRoutes 注册RBAC相关路由
// 使用 SetPermission 声明路由组所属的权限组，系统会自动完成资源绑定
func RegisterRBACRoutes(ctx *services.ServiceContext, api *routegroup.RouterGroup) {
	// 用户模块
	userGroup := api.Group("/users")
	{
		// 公共接口（不需要权限，也不需要登录的jwt）
		userGroup.Public().POST("/register", rbac.Register(ctx))
		userGroup.Public().POST("/login", rbac.Login(ctx))
		authGroup := userGroup.Group("")
		authGroup.Use(middleware.JWT(ctx))
		{
			// 需要登录但是不需要权限控制
			authGroup.POST("/logout", rbac.Logout(ctx))
			authGroup.GET("/options", rbac.UserOptions(ctx))
		}
		// 需要认证和权限 - 声明权限组
		authUserGroup := userGroup.WithMeta("user:manage", "用户管理")
		authUserGroup.Use(middleware.JWT(ctx), middleware.PermissionMiddleware(ctx))
		{
			authUserGroup.GET("/profile", rbac.GetProfile(ctx)).WithMeta("profile", "查询当前用户信息")
			authUserGroup.GET("", rbac.ListUser(ctx)).WithMeta("list", "查询用户列表")
			authUserGroup.POST("", rbac.CreateUser(ctx)).WithMeta("add", "创建用户")
			authUserGroup.PUT("/:id", rbac.UpdateUser(ctx)).WithMeta("update", "编辑用户")
			authUserGroup.DELETE("/:id", rbac.DeleteUser(ctx)).WithMeta("delete", "删除用户")
		}
	}

	// 角色模块 - 声明权限组
	roleGroup := api.Group("/roles").WithMeta("role:manage", "角色管理")
	roleGroup.Use(middleware.JWT(ctx), middleware.PermissionMiddleware(ctx))
	{
		roleGroup.GET("", rbac.GetRoles(ctx)).WithMeta("list", "查询角色列表")
		roleGroup.POST("", rbac.CreateRole(ctx)).WithMeta("add", "创建角色")
		roleGroup.GET("/:id", rbac.GetRole(ctx)).WithMeta("detail", "查询角色详情")
		roleGroup.PUT("/:id", rbac.UpdateRole(ctx)).WithMeta("update", "编辑角色")
		roleGroup.DELETE("/:id", rbac.DeleteRole(ctx)).WithMeta("delete", "删除角色")
		roleGroup.PUT("/:id/assign-resource", rbac.AssignRoleResources(ctx)).WithMeta("assign-perm", "绑定资源权限")
	}

	// 权限模块 - 声明权限组
	permissionGroup := api.Group("/permissions").WithMeta("permission:manage", "权限管理")
	permissionGroup.Use(middleware.JWT(ctx), middleware.PermissionMiddleware(ctx))
	{
		permissionGroup.GET("", rbac.GetPermissions(ctx)).WithMeta("list", "获取权限列表")
	}
}
