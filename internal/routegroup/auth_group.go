package routegroup

import (
	"gin-template/internal/model/rbac"
	"path"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

var (
	// protectedRoutes 受保护的路由列表
	protectedRoutes = make([]ProtectedRoute, 0)
	// mu 保护并发访问
	mu sync.Mutex
)

// ProtectedRoute 受保护的路由信息
type ProtectedRoute struct {
	Resource       rbac.Resource // 资源信息
	PermissionCode string        // 权限组编码
	PermissionName string        // 权限组名称
	Description    string        // 资源描述
}

// AuthRouterGroup 带权限管理的路由组
type AuthRouterGroup struct {
	*gin.RouterGroup
	permissionCode string // 当前路由组所属的权限组编码
	permissionName string // 当前路由组所属的权限组名称
}

// WithAuthRouterGroup 创建带权限管理的路由组
func WithAuthRouterGroup(group *gin.RouterGroup) *AuthRouterGroup {
	return &AuthRouterGroup{
		RouterGroup: group,
	}
}

// SetPermission 设置路由组的权限组信息
func (g *AuthRouterGroup) SetPermission(code, name string) *AuthRouterGroup {
	g.permissionCode = code
	g.permissionName = name
	return g
}

// addProtectedRoute 添加受保护的路由
func (g *AuthRouterGroup) addProtectedRoute(method, relativePath, description string) {
	mu.Lock()
	defer mu.Unlock()
	fullPath := g.calculateFullPath(relativePath)

	protectedRoutes = append(protectedRoutes, ProtectedRoute{
		Resource: rbac.Resource{
			Path:        fullPath,
			Method:      method,
			Description: description,
		},
		PermissionCode: g.permissionCode,
		PermissionName: g.permissionName,
		Description:    description,
	})
}

// calculateFullPath
func (g *AuthRouterGroup) calculateFullPath(relativePath string) string {
	return path.Join(g.RouterGroup.BasePath(), relativePath)
}

// GetProtectedRoutes 获取所有受保护的路由信息
func GetProtectedRoutes() []ProtectedRoute {
	mu.Lock()
	defer mu.Unlock()
	// 返回一个副本以防止外部修改
	routes := make([]ProtectedRoute, len(protectedRoutes))
	copy(routes, protectedRoutes)
	return routes
}

func (g *AuthRouterGroup) Use(middleware ...gin.HandlerFunc) gin.IRoutes {
	g.RouterGroup.Use(middleware...)
	return g
}

func (g *AuthRouterGroup) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	g.addProtectedRoute(httpMethod, relativePath, "")
	g.RouterGroup.Handle(httpMethod, relativePath, handlers...)
	return g
}

func (g *AuthRouterGroup) POST(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	g.addProtectedRoute("POST", relativePath, "")
	g.RouterGroup.POST(relativePath, handlers...)
	return g
}

func (g *AuthRouterGroup) GET(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	g.addProtectedRoute("GET", relativePath, "")
	g.RouterGroup.GET(relativePath, handlers...)
	return g
}

func (g *AuthRouterGroup) DELETE(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	g.addProtectedRoute("DELETE", relativePath, "")
	g.RouterGroup.DELETE(relativePath, handlers...)
	return g
}

func (g *AuthRouterGroup) PATCH(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	g.addProtectedRoute("PATCH", relativePath, "")
	g.RouterGroup.PATCH(relativePath, handlers...)
	return g
}

func (g *AuthRouterGroup) PUT(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	g.addProtectedRoute("PUT", relativePath, "")
	g.RouterGroup.PUT(relativePath, handlers...)
	return g
}

func (g *AuthRouterGroup) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	g.addProtectedRoute("OPTIONS", relativePath, "")
	g.RouterGroup.OPTIONS(relativePath, handlers...)
	return g
}

func (g *AuthRouterGroup) HEAD(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	g.addProtectedRoute("HEAD", relativePath, "")
	g.RouterGroup.HEAD(relativePath, handlers...)
	return g
}

func (g *AuthRouterGroup) Any(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	methods := []string{"GET", "POST", "PUT", "PATCH", "HEAD", "OPTIONS", "DELETE", "CONNECT", "TRACE"}
	for _, method := range methods {
		g.addProtectedRoute(method, relativePath, "")
	}
	g.RouterGroup.Any(relativePath, handlers...)
	return g
}

// Group 创建子路由组，继承父组的权限组信息
func (g *AuthRouterGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *AuthRouterGroup {
	newGroup := g.RouterGroup.Group(relativePath, handlers...)
	return &AuthRouterGroup{
		RouterGroup:    newGroup,
		permissionCode: g.permissionCode, // 继承父组的权限组编码
		permissionName: g.permissionName, // 继承父组的权限组名称
	}
}

// RegisterRoutes 注册需要认证的路由到资源表（已弃用，请使用 InitializeRBAC）
// Deprecated: 请在系统初始化时调用 rbac.InitializeRBAC() 代替
func RegisterRoutes() {
	logrus.Warn("RegisterRoutes is deprecated, please use rbac.InitializeRBAC() instead")
}
