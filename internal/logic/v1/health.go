package v1

import (
	"context"
	"gin-template/internal/core"
	"gin-template/pkg/cache"
	"gin-template/pkg/response"
	"github.com/gin-gonic/gin"
	"time"
)

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	health := gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	}

	// 检查数据库连接
	if db := core.GetDb(); db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			if err := sqlDB.PingContext(ctx); err == nil {
				health["database"] = "ok"
			} else {
				health["database"] = "error"
				health["status"] = "degraded"
			}
		} else {
			health["database"] = "error"
			health["status"] = "degraded"
		}
	} else {
		health["database"] = "not_configured"
	}

	// 检查Redis连接
	if cache.RedisClient != nil {
		if _, err := cache.RedisClient.Ping(ctx).Result(); err == nil {
			health["redis"] = "ok"
		} else {
			health["redis"] = "error"
			health["status"] = "degraded"
		}
	} else {
		health["redis"] = "not_configured"
	}

	response.Success(c, health)
}
