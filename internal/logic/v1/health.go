package v1

import (
	"context"
	"gin-admin/internal/core"
	"gin-admin/internal/services"
	"gin-admin/pkg/cache"
	"gin-admin/pkg/response"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

/*
* @Author: zouyx
* @Package: 健康检查 - 企业级健康检查接口
 */

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy  HealthStatus = "healthy"  // 所有组件正常
	HealthStatusDegraded HealthStatus = "degraded" // 部分组件异常但服务可用
	HealthStatusDown     HealthStatus = "down"     // 服务不可用
)

// ComponentHealth 组件健康状态
type ComponentHealth struct {
	Status       string                 `json:"status"`            // ok, error, not_configured
	ResponseTime int64                  `json:"response_time_ms"`  // 响应时间（毫秒）
	Message      string                 `json:"message,omitempty"` // 状态消息
	Details      map[string]interface{} `json:"details,omitempty"` // 详细信息
	Error        string                 `json:"error,omitempty"`   // 错误信息
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status     HealthStatus               `json:"status"`           // 总体状态
	Timestamp  int64                      `json:"timestamp"`        // 检查时间戳
	Version    string                     `json:"version"`          // 应用版本
	Uptime     int64                      `json:"uptime_seconds"`   // 运行时长（秒）
	Components map[string]ComponentHealth `json:"components"`       // 各组件状态
	System     map[string]interface{}     `json:"system,omitempty"` // 系统信息
}

var startTime = time.Now()

// HealthCheck godoc
// @Summary 健康检查
// @Description 检查系统各组件的健康状态
// @Tags 系统监控
// @Produce json
// @Success 200 {object} response.Response{data=HealthResponse} "健康状态"
// @Router /health [get]
func HealthCheck(svcCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		startCheck := time.Now()

		healthResp := HealthResponse{
			Status:     HealthStatusHealthy,
			Timestamp:  time.Now().Unix(),
			Version:    getAppVersion(),
			Uptime:     int64(time.Since(startTime).Seconds()),
			Components: make(map[string]ComponentHealth),
			System:     getSystemInfo(),
		}

		// 检查数据库
		dbHealth := checkDatabase(ctx, svcCtx)
		healthResp.Components["database"] = dbHealth
		if dbHealth.Status == "error" {
			healthResp.Status = HealthStatusDegraded
		}

		// 检查缓存
		cacheHealth := checkCache(ctx, svcCtx)
		healthResp.Components["cache"] = cacheHealth
		if cacheHealth.Status == "error" {
			healthResp.Status = HealthStatusDegraded
		}

		// 添加健康检查耗时
		healthResp.Components["health_check"] = ComponentHealth{
			Status:       "ok",
			ResponseTime: time.Since(startCheck).Milliseconds(),
			Message:      "Health check completed",
		}

		response.Success(c, healthResp)
	}
}

// checkDatabase 检查数据库健康状态
func checkDatabase(ctx context.Context, svcCtx *services.ServiceContext) ComponentHealth {
	start := time.Now()

	db := svcCtx.Db
	if db == nil {
		return ComponentHealth{
			Status:       "not_configured",
			ResponseTime: time.Since(start).Milliseconds(),
			Message:      "Database not configured",
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		return ComponentHealth{
			Status:       "error",
			ResponseTime: time.Since(start).Milliseconds(),
			Message:      "Failed to get database connection",
			Error:        err.Error(),
		}
	}

	// 执行 Ping 检查
	if err := sqlDB.PingContext(ctx); err != nil {
		return ComponentHealth{
			Status:       "error",
			ResponseTime: time.Since(start).Milliseconds(),
			Message:      "Database ping failed",
			Error:        err.Error(),
		}
	}

	// 获取连接池状态
	stats := sqlDB.Stats()

	return ComponentHealth{
		Status:       "ok",
		ResponseTime: time.Since(start).Milliseconds(),
		Message:      "Database is healthy",
		Details: map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"in_use":           stats.InUse,
			"idle":             stats.Idle,
			"max_open":         stats.MaxOpenConnections,
		},
	}
}

// checkCache 检查缓存健康状态
func checkCache(ctx context.Context, svcCtx *services.ServiceContext) ComponentHealth {
	start := time.Now()

	if !cache.IsAvailable() {
		return ComponentHealth{
			Status:       "not_configured",
			ResponseTime: time.Since(start).Milliseconds(),
			Message:      "Cache not configured",
		}
	}

	cacheClient := svcCtx.Cache

	// 执行 Ping 检查
	if err := cacheClient.Ping(ctx); err != nil {
		return ComponentHealth{
			Status:       "error",
			ResponseTime: time.Since(start).Milliseconds(),
			Message:      "Cache ping failed",
			Error:        err.Error(),
		}
	}

	return ComponentHealth{
		Status:       "ok",
		ResponseTime: time.Since(start).Milliseconds(),
		Message:      "Cache is healthy",
	}
}

// getAppVersion 获取应用版本
func getAppVersion() string {
	cfg, err := core.GetConfig()
	if err == nil && cfg != nil {
		return cfg.App.Version
	}
	return "unknown"
}

// getSystemInfo 获取系统信息
func getSystemInfo() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"go_version":      runtime.Version(),
		"goroutines":      runtime.NumGoroutine(),
		"memory_alloc_mb": m.Alloc / 1024 / 1024,
		"memory_sys_mb":   m.Sys / 1024 / 1024,
		"cpu_count":       runtime.NumCPU(),
		"os":              runtime.GOOS,
		"arch":            runtime.GOARCH,
	}
}
