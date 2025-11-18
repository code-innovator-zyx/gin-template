package middleware

import (
	"errors"
	"strings"

	"gin-admin/pkg/response"
	"gin-admin/pkg/utils"

	"github.com/gin-gonic/gin"
)

// JWT 认证中间件
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Unauthorized(c, "无效的Token格式")
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := utils.ParseToken(token)
		if err != nil {
			if errors.Is(err, utils.ErrTokenExpired) {
				response.Unauthorized(c, "Token已过期")
				c.Abort()
				return
			}
			response.Unauthorized(c, "无效的Token")
			c.Abort()
			return
		}
		c.Set("uid", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
