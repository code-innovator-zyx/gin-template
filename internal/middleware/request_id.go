package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID 添加请求ID中间件，用于追踪请求链路
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先尝试从请求头获取
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// 如果没有则生成新的UUID
			requestID = uuid.New().String()
		}

		// 设置到上下文
		c.Set("request_id", requestID)
		// 设置响应头
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}
