package middleware

import (
	"strings"

	"gin-template/pkg/response"
	"gin-template/pkg/utils"
	"github.com/gin-gonic/gin"
)

// JWT 认证中间件
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Unauthorized(c, "无效的Token格式")
			c.Abort()
			return
		}

		// 解析Token
		token := parts[1]
		claims, err := utils.ParseToken(token)
		if err != nil {
			if err == utils.ErrTokenExpired {
				response.Unauthorized(c, "Token已过期")
				c.Abort()
				return
			}
			response.Unauthorized(c, "无效的Token")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
