package middleware

import (
	"gin-template/internal/model/rbac"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由到资源表
func RegisterRoutes(engine *gin.Engine) {
	for _, route := range engine.Routes() {
		// 跳过一些特殊路由，如 swagger 文档等
		if shouldSkipRoute(route.Path) {
			continue
		}
		// 将路由添加到系统资源
		err := rbac.UpsertResource(route.Path, route.Method)
		if err != nil {
			// 这里可以选择记录日志而不是直接返回错误
			// 因为这是在启动时执行的，我们不希望因为一个路由注册失败就导致整个服务无法启动
			gin.DefaultWriter.Write([]byte("Failed to register route as resource: " + err.Error() + "\n"))
		}
	}
}

// shouldSkipRoute 判断是否应该跳过某些路由的注册
func shouldSkipRoute(path string) bool {
	// 跳过 swagger 相关路由
	if len(path) >= 8 && path[:8] == "/swagger" {
		return true
	}

	// 跳过健康检查路由
	if path == "/health" || path == "/ping" {
		return true
	}

	return false
}
