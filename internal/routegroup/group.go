package routegroup

import (
	"gin-admin/internal/model/rbac"

	"path"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	protectedRoutesMu sync.Mutex
	protectedRoutes   []*ProtectedRoute
)

// ProtectedRoute represents a route that requires permission
type ProtectedRoute struct {
	Resource       rbac.Resource
	PermissionCode string // 权限组 code
	PermissionName string // 权限组 name
	Description    string
}

// RouterGroup wraps gin.RouterGroup and adds RBAC metadata
type RouterGroup struct {
	*gin.RouterGroup
	isPublic       bool
	permissionCode string
	permissionName string
}

// -------------------- RouterGroup --------------------

func WrapGroup(group *gin.RouterGroup) *RouterGroup {
	return &RouterGroup{RouterGroup: group}
}

/*
WithMeta 设置permission 权限组的code 和 name
code 控制前端菜单页面是否显示的权限
*/
func (g *RouterGroup) WithMeta(code, name string) *RouterGroup {
	g.permissionCode = code
	g.permissionName = name
	g.isPublic = false
	return g
}

// Public marks the RouterGroup as public
func (g *RouterGroup) Public() *RouterGroup {
	g.isPublic = true
	return g
}

// Group creates a sub-group, inheriting parent metadata
func (g *RouterGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	newGroup := g.RouterGroup.Group(relativePath, handlers...)
	return &RouterGroup{
		RouterGroup:    newGroup,
		permissionCode: g.permissionCode,
		permissionName: g.permissionName,
		isPublic:       g.isPublic,
	}
}

// Use wraps gin.RouterGroup.Use
func (g *RouterGroup) Use(middlewares ...gin.HandlerFunc) *RouterGroup {
	g.RouterGroup.Use(middlewares...)
	return g
}

// Handle adds a generic route with permission registration
func (g *RouterGroup) handle(method, relativePath string, handlers ...gin.HandlerFunc) *Route {
	fullPath := g.calculateFullPath(relativePath)

	var prot *ProtectedRoute
	if !g.isPublic {
		protectedRoutesMu.Lock()
		prot = &ProtectedRoute{
			Resource: rbac.Resource{
				Path:   fullPath,
				Method: method,
			},
			PermissionCode: g.permissionCode,
			PermissionName: g.permissionName,
		}
		protectedRoutes = append(protectedRoutes, prot)
		protectedRoutesMu.Unlock()
	}

	g.RouterGroup.Handle(method, relativePath, handlers...)
	return &Route{
		group:   g,
		Path:    fullPath,
		Methods: []string{method},
		prots:   []*ProtectedRoute{prot},
	}
}

// GET /POST/PUT/DELETE/PATCH/OPTIONS/HEAD helpers
func (g *RouterGroup) GET(path string, handlers ...gin.HandlerFunc) *Route {
	return g.handle("GET", path, handlers...)
}
func (g *RouterGroup) POST(path string, handlers ...gin.HandlerFunc) *Route {
	return g.handle("POST", path, handlers...)
}
func (g *RouterGroup) PUT(path string, handlers ...gin.HandlerFunc) *Route {
	return g.handle("PUT", path, handlers...)
}
func (g *RouterGroup) DELETE(path string, handlers ...gin.HandlerFunc) *Route {
	return g.handle("DELETE", path, handlers...)
}
func (g *RouterGroup) PATCH(path string, handlers ...gin.HandlerFunc) *Route {
	return g.handle("PATCH", path, handlers...)
}
func (g *RouterGroup) OPTIONS(path string, handlers ...gin.HandlerFunc) *Route {
	return g.handle("OPTIONS", path, handlers...)
}
func (g *RouterGroup) HEAD(path string, handlers ...gin.HandlerFunc) *Route {
	return g.handle("HEAD", path, handlers...)
}

// Any adds a route for all standard HTTP methods
func (g *RouterGroup) Any(relativePath string, handlers ...gin.HandlerFunc) *Route {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}
	prots := make([]*ProtectedRoute, 0, len(methods))
	for _, method := range methods {
		r := g.handle(method, relativePath, handlers...)
		prots = append(prots, r.prots...)
	}
	return &Route{
		group:   g,
		Path:    g.calculateFullPath(relativePath),
		Methods: methods,
		prots:   prots,
	}
}
func (g *RouterGroup) calculateFullPath(relativePath string) string {
	return path.Join(g.RouterGroup.BasePath(), relativePath)
}

// GetProtectedRoutes returns a copy of all registered protected routes
func GetProtectedRoutes() []*ProtectedRoute {
	protectedRoutesMu.Lock()
	defer protectedRoutesMu.Unlock()
	prots := make([]*ProtectedRoute, len(protectedRoutes))
	copy(prots, protectedRoutes)
	return prots
}
