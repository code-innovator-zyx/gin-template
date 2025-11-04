package routegroup

import (
	"gin-template/internal/model/rbac"
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

// NewAuthRouterGroup
func NewAuthRouterGroup(group *gin.RouterGroup) *AuthRouterGroup {
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
