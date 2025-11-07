package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

// Logger 自定义日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 处理请求
		c.Next()

		// 计算请求时长
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()

		// 获取请求ID
		requestID, _ := c.Get("request_id")

		// 记录日志
		entry := logrus.WithFields(logrus.Fields{
			"request_id": requestID,
			"method":     method,
			"path":       path,
			"status":     statusCode,
			"latency":    latency.String(),
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		})

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			entry.Error("请求处理失败", statusCode)
		} else if statusCode >= 400 {
			entry.Warn("客户端请求错误", statusCode)
		} else {
			entry.Info("请求处理成功")
		}
	}
}
