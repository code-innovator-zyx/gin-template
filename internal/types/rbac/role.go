package rbac

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/15 上午11:06
* @Package:
 */
type ListRoleRequest struct {
	Name     string `form:"name,optional" json:"name" binding:"-" example:"johndoe"`
	Status   uint8  `form:"status,optional" json:"status" binding:"-" example:"1"`
	Page     int    `form:"page,default=1" json:"page" binding:"required" example:"1" default:"1"`
	PageSize int    `form:"pageSize,default=10" json:"pageSize" binding:"required" example:"10" default:"10"`
}

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
