package user

import (
	userApi "github.com/code-innovator-zyx/gin-template/internal/logic/v1/user"
	"github.com/code-innovator-zyx/gin-template/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes 注册用户相关路由
func RegisterUserRoutes(r *gin.RouterGroup) {
	userGroup := r.Group("/user")
	{
		// 公开路由
		userGroup.POST("/register", userApi.Register)
		userGroup.POST("/login", userApi.Login)

		// 需要认证的路由
		authGroup := userGroup.Use(middleware.JWT())
		{
			authGroup.GET("/profile", userApi.GetProfile)
		}
	}
}
