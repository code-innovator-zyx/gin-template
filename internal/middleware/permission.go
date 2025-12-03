package middleware

import (
	"gin-admin/internal/services"
	"gin-admin/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// PermissionMiddleware 权限验证中间件
func PermissionMiddleware(svrCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户ID
		userID, exists := c.Get("uid")
		if !exists {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		has, err := svrCtx.CacheService.CheckUserPermission(c.Request.Context(), userID.(uint), c.FullPath(), c.Request.Method, svrCtx.ResourceService.GetUserResources)
		if err != nil {
			has, err = svrCtx.ResourceService.CheckUserPermission(c.Request.Context(), userID.(uint), c.FullPath(), c.Request.Method)
			if nil != err {
				logrus.Error("failed check user permission: ", err)
				response.InternalServerError(c, "权限检查失败")
				c.Abort()
				return
			}
		}
		if !has {
			response.Forbidden(c, "没有权限")
			c.Abort()
			return
		}

		c.Next()
	}
}
