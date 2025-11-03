package rbac

import (
	"time"
)

// RolePermission 角色权限关联模型
type RolePermission struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	RoleID       uint      `gorm:"not null;index" json:"role_id"`
	PermissionID uint      `gorm:"not null;index" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 设置表名
func (RolePermission) TableName() string {
	return "role_permissions"
}
