package middleware

import (
	"github.com/code-innovator-zyx/gin-template/internal/model/rbac"
	"github.com/code-innovator-zyx/gin-template/pkg/response"
	"github.com/gin-gonic/gin"
)

// PermissionMiddleware 权限验证中间件
func PermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户ID
		userID, exists := c.Get("userID")
		if !exists {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		// 获取请求路径和方法
		path := c.Request.URL.Path
		method := c.Request.Method

		// 检查用户是否有权限访问
		hasPermission, err := rbac.CheckPermission(userID.(uint), path, method)
		if err != nil {
			response.InternalServerError(c, "权限检查失败")
			c.Abort()
			return
		}

		if !hasPermission {
			response.Forbidden(c, "没有权限")
			c.Abort()
			return
		}

		c.Next()
	}
}
