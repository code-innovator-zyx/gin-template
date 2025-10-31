package rbac

import (
	"gin-template/internal/logic/v1/rbac"
	"gin-template/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRBACRoutes 注册RBAC相关路由
func RegisterRBACRoutes(api *gin.RouterGroup) {
	// RBAC相关路由
	rbacGroup := api.Group("/rbac")
	{
		// 需要JWT认证的路由
		authGroup := rbacGroup.Group("/")
		authGroup.Use(middleware.JWT())
		{
			// 角色管理
			authGroup.GET("/roles", rbac.GetRoles)
			authGroup.POST("/roles", rbac.CreateRole)
			authGroup.GET("/roles/:id", rbac.GetRole)
			authGroup.PUT("/roles/:id", rbac.UpdateRole)
			authGroup.DELETE("/roles/:id", rbac.DeleteRole)

			// 权限管理
			authGroup.GET("/permissions", rbac.GetPermissions)
			authGroup.POST("/permissions", rbac.CreatePermission)

			// 用户角色管理
			authGroup.GET("/users/:id/roles", rbac.GetUserRoles)
			authGroup.POST("/users/:id/roles", rbac.AssignRoleToUser)
			authGroup.DELETE("/users/:id/roles/:role_id", rbac.RemoveRoleFromUser)

			// 角色权限管理
			authGroup.POST("/roles/:id/permissions", rbac.AssignPermissionToRole)
			authGroup.DELETE("/roles/:id/permissions/:permission_id", rbac.RemovePermissionFromRole)

			// 角色菜单管理
			authGroup.POST("/roles/:id/menus", rbac.AssignMenuToRole)
			authGroup.DELETE("/roles/:id/menus/:menu_id", rbac.RemoveMenuFromRole)
		}
	}
}
