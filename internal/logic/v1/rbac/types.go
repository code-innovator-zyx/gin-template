package rbac

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/10/31 下午4:40
* @Package:
 */

type CreateRoleRequest struct {
	Name        string `json:"name,omitempty" example:"管理员"`
	Description string `json:"description,omitempty" example:"系统管理员"`
	Resources   []uint `json:"resources,omitempty" description:"角色绑定的资源ID列表"`
}

// UpdateRoleRequest 更新角色
type UpdateRoleRequest struct {
	Name        string `json:"name,omitempty" example:"管理员"`
	Description string `json:"description,omitempty" example:"系统管理员"`
	Resources   []uint `json:"resources,omitempty" description:"角色绑定的资源ID列表"`
}
