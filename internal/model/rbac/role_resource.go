package rbac

// RoleResource 角色资源关联模型（细粒度权限控制）
// 角色可以：
// 1. 通过 role_permissions 绑定整个权限组（粗粒度）
// 2. 通过 role_resources 绑定特定资源（细粒度）
type RoleResource struct {
	ID         uint `gorm:"primarykey" json:"id"`
	RoleID     uint `gorm:"not null;index" json:"role_id"`
	ResourceID uint `gorm:"not null;index" json:"resource_id"`
}

// TableName 设置表名
func (RoleResource) TableName() string {
	return "role_resources"
}

