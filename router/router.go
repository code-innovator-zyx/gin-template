package router

import (
	"github.com/code-innovator-zyx/gin-template/internal/middleware"
	v1 "github.com/code-innovator-zyx/gin-template/router/v1"
	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	r := gin.Default()

	// 跨域中间件
	r.Use(middleware.Cors())

	// 注册API路由
	v1.RegisterRoutes(r)

	return r
}
