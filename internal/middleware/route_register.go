package middleware

import (
	"gin-template/internal/model/rbac"
	"gin-template/internal/routegroup"
	"github.com/sirupsen/logrus"
	"strings"
)

// RegisterRoutes 注册需要认证的路由到资源表
func RegisterRoutes() {
	// 获取所有通过 AuthRouterGroup 注册的受保护路由
	protectedRoutes := routegroup.GetProtectedRoutes()

	// 将路由添加到系统资源
	err := rbac.UpsertResource(protectedRoutes)
	if err != nil {
		// 因为这是在启动时执行的，我们不希望因为一个路由注册失败就导致整个服务无法启动
		logrus.Error("Failed to register route as resource: " + err.Error())
	}

	logrus.Info("Successfully registered routes to resource table")
}

// shouldSkipRoute 判断是否应该跳过某些路由的注册
func shouldSkipRoute(path string) bool {
	// 跳过 swagger 相关路由
	if strings.HasPrefix(path, "/swagger") {
		return true
	}

	// 跳过健康检查路由
	if path == "/health" || path == "/ping" {
		return true
	}

	return false
}
