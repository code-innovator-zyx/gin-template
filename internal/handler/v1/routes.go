package v1

import (
	_ "gin-template/docs"
	"gin-template/internal/core"
	"gin-template/internal/handler/v1/rbac"
	"gin-template/internal/handler/v1/user"
	v1 "gin-template/internal/logic/v1"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /api/v1

func RegisterRoutes(r *gin.Engine) {
	// API版本v1
	apiV1 := r.Group("/api/v1")

	// 注册各个模块的路由
	registerHealthRoutes(apiV1)
	user.RegisterUserRoutes(apiV1)
	rbac.RegisterRBACRoutes(apiV1)
	if core.Config.App.EnableSwagger {
		r.GET("/swagger/v1/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.InstanceName("v1")))
	}
}

// registerHealthRoutes 注册健康检查相关路由
func registerHealthRoutes(api *gin.RouterGroup) {
	// 健康检查
	api.GET("/health", v1.HealthCheck)
}
