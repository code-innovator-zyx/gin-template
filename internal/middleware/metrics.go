package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics 指标收集结构
type Metrics struct {
	RequestCount   map[string]int64
	ResponseTime   map[string][]time.Duration
	StatusCodeDist map[int]int64
}

var globalMetrics = &Metrics{
	RequestCount:   make(map[string]int64),
	ResponseTime:   make(map[string][]time.Duration),
	StatusCodeDist: make(map[int]int64),
}

// MetricsMiddleware 性能指标收集中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 处理请求
		c.Next()

		// 收集指标
		duration := time.Since(start)
		status := c.Writer.Status()
		
		// 记录请求计数
		key := method + " " + path
		globalMetrics.RequestCount[key]++
		
		// 记录响应时间
		globalMetrics.ResponseTime[key] = append(globalMetrics.ResponseTime[key], duration)
		
		// 记录状态码分布
		globalMetrics.StatusCodeDist[status]++

		// 设置响应头
		c.Header("X-Response-Time", duration.String())
	}
}

// PrometheusMetrics Prometheus格式的指标中间件
// 需要集成 prometheus client_golang
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath() // 使用路由模板而不是实际路径
		if path == "" {
			path = c.Request.URL.Path
		}

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		// 这里可以集成 Prometheus 指标
		// httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		// httpRequestDuration.WithLabelValues(method, path).Observe(duration)
		
		_ = duration
		_ = status
		_ = method
	}
}

// GetMetrics 获取当前指标
func GetMetrics() *Metrics {
	return globalMetrics
}

