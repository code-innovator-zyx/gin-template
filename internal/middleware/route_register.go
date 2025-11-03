package middleware

import (
	"gin-template/internal/model/rbac"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterRoutes 注册所有路由到资源表
func RegisterRoutes(engine *gin.Engine) {
	var resources = make([]rbac.Resource, 0, len(engine.Routes()))
	for _, route := range engine.Routes() {
		// 跳过一些特殊路由，如 swagger 文档等
		if shouldSkipRoute(route.Path) {
			continue
		}
		resources = append(resources, rbac.Resource{
			Path:        route.Path,
			Method:      route.Method,
			Description: "",
			IsManaged:   false,
		})
	}
	// 将路由添加到系统资源
	err := rbac.UpsertResource(resources)
	if err != nil {
		// 因为这是在启动时执行的，我们不希望因为一个路由注册失败就导致整个服务无法启动
		logrus.Error("Failed to register route as resource: " + err.Error())
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
