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

	// 用户模块
	userGroup := api.Group("/user")
	{
		// 公共接口（不需要权限）
		userGroup.POST("/register", rbac.Register)
		userGroup.POST("/login", rbac.Login)
		userGroup.POST("/refresh", rbac.RefreshToken) // 刷新令牌
		authGroup := userGroup.Group("")
		authGroup.Use(middleware.JWT())
		{
			authGroup.POST("/logout", rbac.Logout)
		}
		// 需要认证和权限 - 声明权限组
		authUserGroup := routegroup.WithAuthRouterGroup(userGroup).
			SetPermission("user:manage", "用户管理")
		authUserGroup.Use(middleware.JWT(), middleware.PermissionMiddleware())
		{
			authUserGroup.GETDesc("", "查询用户列表")
			authUserGroup.GETDesc("/profile", "获取当前用户信息", rbac.GetProfile)
			authUserGroup.GETDesc("/:id", "查询用户详情")
			authUserGroup.POSTDesc("", "创建用户")
			authUserGroup.PUTDesc("/:id", "修改用户")
			authUserGroup.GETDesc("/:id/roles", "获取用户角色", rbac.GetUserRoles)
			authUserGroup.POSTDesc("/:id/roles", "分配角色给用户", rbac.AssignRoleToUser)
			authUserGroup.DELETEDesc("/:id/roles/:role_id", "移除用户角色", rbac.RemoveRoleFromUser)
		}
	}

	// 角色模块 - 声明权限组
	roleGroup := routegroup.WithAuthRouterGroup(api.Group("/roles")).
		SetPermission("role:manage", "角色管理")
	roleGroup.Use(middleware.JWT(), middleware.PermissionMiddleware())
	{
		roleGroup.GETDesc("", "获取角色列表", rbac.GetRoles)
		roleGroup.POSTDesc("", "创建角色", rbac.CreateRole)
		roleGroup.GETDesc("/:id", "获取角色详情", rbac.GetRole)
		roleGroup.PUTDesc("/:id", "更新角色", rbac.UpdateRole)
		roleGroup.DELETEDesc("/:id", "删除角色", rbac.DeleteRole)

		// 角色-资源绑定（实际授权）
		roleGroup.POSTDesc("/:id/resources/:resource_id", "分配资源给角色", rbac.AssignResourceToRole)
		roleGroup.DELETEDesc("/:id/resources/:resource_id", "移除角色资源", rbac.RemoveResourceFromRole)
	}

	// 权限模块 - 声明权限组
	permissionGroup := routegroup.WithAuthRouterGroup(api.Group("/permissions")).
		SetPermission("permission:manage", "权限管理")
	permissionGroup.Use(middleware.JWT(), middleware.PermissionMiddleware())
	{
		permissionGroup.GETDesc("", "获取权限列表", rbac.GetPermissions)
	}
}
