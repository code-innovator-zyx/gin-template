package v1

import (
	_ "gin-admin/docs"
	"gin-admin/internal/core"
	"gin-admin/internal/handler/v1/rbac"
	v1 "gin-admin/internal/logic/v1"
	"gin-admin/internal/routegroup"
	"gin-admin/internal/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Gin Template API
// @version         1.0.0
// @description     企业级 Gin 框架模板项目 - RESTful API 文档
// @description     提供完整的 RBAC 权限管理、用户认证、资源管理等企业级功能
// @description
// @description     ## 认证说明
// @description     大部分 API 需要 JWT Token 认证，请先调用登录接口获取 token
// @description     然后点击右上角 Authorize 按钮，输入格式：`Bearer {token}`
// @description
// @description     ## 错误码说明
// @description     - 200: 成功
// @description     - 400: 请求参数错误
// @description     - 401: 未授权或 Token 过期
// @description     - 403: 权限不足
// @description     - 404: 资源不存在
// @description     - 500: 服务器内部错误

// @contact.name    技术支持团队
// @contact.url     https://github.com/your-org/gin-template
// @contact.email   support@yourcompany.com

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8080
// @BasePath        /api/v1

// @schemes         http https

// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        Authorization
// @description                 JWT Token 认证，格式：Bearer {token}

// @tag.name            RBAC-用户管理
// @tag.description     用户注册、登录、个人信息管理

// @tag.name            RBAC-角色管理
// @tag.description     角色的增删改查操作

// @tag.name            RBAC-权限管理
// @tag.description     权限列表查询与管理

// @tag.name            RBAC-资源管理
// @tag.description     系统资源的查询与配置

// @tag.name            RBAC-用户角色管理
// @tag.description     用户与角色的关联管理

// @tag.name            RBAC-角色权限管理
// @tag.description     角色与权限的关联管理

// @query.collection.format    multi

// @externalDocs.description    项目文档

// @externalDocs.url            https://github.com/your-org/gin-admin/docs
func RegisterRoutes(ctx *services.ServiceContext, r *gin.Engine) {
	// API版本v1 使用authGroup 自动维护权限管理
	apiV1 := routegroup.WrapGroup(r.Group("/api/v1"))

	// 注册各个模块的路由
	registerHealthRoutes(apiV1)
	// 用户管理已整合到RBAC系统中
	rbac.RegisterRBACRoutes(ctx, apiV1)
	if core.MustGetConfig().App.EnableSwagger {
		r.GET("/swagger/v1/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.InstanceName("v1")))
	}
}

// registerHealthRoutes 注册健康检查相关路由
func registerHealthRoutes(api *routegroup.RouterGroup) {
	// 健康检查
	api.Public().GET("/health", v1.HealthCheck)
}
