package middleware

import (
	"gin-template/internal/service"
	"gin-template/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

		// 使用缓存服务检查权限
		hasPermission, err := service.GetRbacService().CheckUserPermission(c.Request.Context(), userID.(uint), path, method)
		if err != nil {
			logrus.Error("failed check user permission: ", err)
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
