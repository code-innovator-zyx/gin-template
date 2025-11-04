package handler

import (
	v1 "gin-template/internal/handler/v1"
	"gin-template/internal/middleware"
	"gin-template/internal/routegroup"

	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	r := gin.Default()

	// 跨域中间件
	r.Use(middleware.Cors())

	// 注册API路由
	v1.RegisterRoutes(r)

	// 注册所有路由到资源表
	// 设计说明：
	// 1. 通过 core.RegisterRoutes() 集中管理所有路由注册逻辑，便于统一扩展。
	// 2. 支持未来动态加载路由的扩展，例如通过配置文件或数据库。
	// 3. 提高代码清晰度，避免路由注册逻辑分散。
	routegroup.RegisterRoutes()

	return r
}
