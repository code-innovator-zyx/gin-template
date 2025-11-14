package routegroup

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/14 上午10:53
* @Package:
 */

// Route wraps a single route and allows chainable metadata configuration
type Route struct {
	group   *RouterGroup
	Path    string
	Methods []string
	prots   []*ProtectedRoute
}

/*
WithMeta 设置接口权限的code 和中文描述
code 作为前端接口级别的按钮权限控制
*/
func (r *Route) WithMeta(code, description string) gin.IRoutes {
	for i := range r.prots {
		r.prots[i].Resource.Code = fmt.Sprintf("%s:%s", r.group.permissionCode, code)
		r.prots[i].Resource.Description = description
	}
	return r.group.RouterGroup
}

// WithDescription sets only description
func (r *Route) WithDescription(desc string) *Route {
	for i := range r.prots {
		r.prots[i].Resource.Description = desc
	}
	return r
}
