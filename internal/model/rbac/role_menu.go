package rbac

import (
	"time"

	"gorm.io/gorm"
)

// RoleMenu 角色菜单关联模型
type RoleMenu struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	RoleID    uint           `gorm:"not null;index" json:"role_id"`
	MenuID    uint           `gorm:"not null;index" json:"menu_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 设置表名
func (RoleMenu) TableName() string {
	return "role_menus"
}