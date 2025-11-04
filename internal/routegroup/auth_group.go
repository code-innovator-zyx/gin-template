package routegroup

import (
	"gin-template/internal/model/rbac"
	"github.com/sirupsen/logrus"
	"path"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	// protectedRoutes
	protectedRoutes = make([]rbac.Resource, 0)
	// mu
	mu sync.Mutex
)

// AuthRouterGroup
type AuthRouterGroup struct {
	*gin.RouterGroup
}

// WithAuthRouterGroup
func WithAuthRouterGroup(group *gin.RouterGroup) *AuthRouterGroup {
	return &AuthRouterGroup{RouterGroup: group}
}

// addProtectedRoute
func (g *AuthRouterGroup) addProtectedRoute(method, path string) {
	mu.Lock()
	defer mu.Unlock()
	fullPath := g.calculateFullPath(path)
	protectedRoutes = append(protectedRoutes, rbac.Resource{
		Path:   fullPath,
		Method: method,
	})
}

// calculateFullPath
func (g *AuthRouterGroup) calculateFullPath(relativePath string) string {
	return path.Join(g.RouterGroup.BasePath(), relativePath)
}

// GetProtectedRoutes
func GetProtectedRoutes() []rbac.Resource {
	mu.Lock()
	defer mu.Unlock()
	// 返回一个副本以防止外部修改
	routes := make([]rbac.Resource, len(protectedRoutes))
	copy(routes, protectedRoutes)
	return routes
}

func (g *AuthRouterGroup) Use(middleware ...gin.HandlerFunc) gin.IRoutes {
	g.RouterGroup.Use(middleware...)
	return g
}

func (g *AuthRouterGroup) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	g.addProtectedRoute(httpMethod, relativePath)
	g.RouterGroup.Handle(httpMethod, relativePath, handlers...)
	return g
}

func (g *AuthRouterGroup) POST(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	g.addProtectedRoute("POST", relativePath)
	g.RouterGroup.POST(relativePath, handlers...)
	return g
}

func (g *AuthRouterGroup) GET(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	g.addProtectedRoute("GET", relativePath)
	g.RouterGroup.GET(relativePath, handlers...)
	return g
}

func (g *AuthRouterGroup) DELETE(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	g.addProtectedRoute("DELETE", relativePath)
	g.RouterGroup.DELETE(relativePath, handlers...)
	return g
}

func (g *AuthRouterGroup) PATCH(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	g.addProtectedRoute("PATCH", relativePath)
	g.RouterGroup.PATCH(relativePath, handlers...)
	return g
}

func (g *AuthRouterGroup) PUT(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	g.addProtectedRoute("PUT", relativePath)
	g.RouterGroup.PUT(relativePath, handlers...)
	return g
}

func (g *AuthRouterGroup) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	g.addProtectedRoute("OPTIONS", relativePath)
	g.RouterGroup.OPTIONS(relativePath, handlers...)
	return g
}

func (g *AuthRouterGroup) HEAD(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	g.addProtectedRoute("HEAD", relativePath)
	g.RouterGroup.HEAD(relativePath, handlers...)
	return g
}

func (g *AuthRouterGroup) Any(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	methods := []string{"GET", "POST", "PUT", "PATCH", "HEAD", "OPTIONS", "DELETE", "CONNECT", "TRACE"}
	for _, method := range methods {
		g.addProtectedRoute(method, relativePath)
	}
	g.RouterGroup.Any(relativePath, handlers...)
	return g
}
func (g *AuthRouterGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *AuthRouterGroup {
	newGroup := g.RouterGroup.Group(relativePath, handlers...)
	return &AuthRouterGroup{RouterGroup: newGroup}
}

// RegisterRoutes 注册需要认证的路由到资源表
func RegisterRoutes() {
	mu.Lock()
	defer mu.Unlock()
	// 将路由添加到系统资源
	err := rbac.UpsertResource(protectedRoutes)
	if err != nil {
		// 因为这是在启动时执行的，我们不希望因为一个路由注册失败就导致整个服务无法启动
		logrus.Error("Failed to register route as resource: " + err.Error())
	}

	logrus.Info("Successfully registered routes to resource table")
}
