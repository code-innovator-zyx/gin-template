package v1

import (
	"gin-template/pkg/response"
	"github.com/gin-gonic/gin"
)

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "ok",
	})
}
