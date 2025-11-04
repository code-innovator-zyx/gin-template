package handler

import (
	v1 "gin-template/internal/handler/v1"
	"gin-template/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	r := gin.Default()

	// 跨域中间件
	r.Use(middleware.Cors())

	// 注册API路由
	v1.RegisterRoutes(r)

	// 注册所有路由到资源表  todo 开放的接口不用注册到资源表
	middleware.RegisterRoutes()

	return r
}
