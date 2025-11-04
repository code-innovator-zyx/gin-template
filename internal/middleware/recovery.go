package middleware

import (
	"fmt"
	"gin-template/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"runtime/debug"
)

// Recovery 自定义panic恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic堆栈信息
				logrus.WithFields(logrus.Fields{
					"error":      err,
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"ip":         c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
					"stack":      string(debug.Stack()),
				}).Error("发生panic")

				// 返回500错误
				response.InternalServerError(c, fmt.Sprintf("服务器内部错误: %v", err))
				c.Abort()
			}
		}()
		c.Next()
	}
}

