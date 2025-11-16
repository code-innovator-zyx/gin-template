package rbac

import (
	"time"
)

// Permission 权限组模型（仅作为逻辑分组，用于前端UI展示，不参与实际授权）
type Permission struct {
	ID        uint       `gorm:"primarykey" json:"id" example:"1" description:"权限ID"`
	Name      string     `gorm:"size:50;not null;uniqueIndex:idx_perm_name" json:"name" example:"用户管理" description:"权限组中文名"`
	Code      string     `gorm:"size:50;not null;uniqueIndex:idx_perm_code" json:"code" example:"user:manage" description:"权限编码"`
	Resources []Resource `gorm:"foreignKey:PermissionID" json:"resources" description:"权限组下的资源列表（仅用于UI展示分组）"`
	CreatedAt time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z" description:"创建时间"`
	UpdatedAt time.Time  `json:"updated_at" example:"2023-01-01T00:00:00Z" description:"更新时间"`
}

// TableName 设置表名
func (Permission) TableName() string {
	return "permissions"
}
