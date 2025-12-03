package rbac

import (
	"gin-admin/pkg/consts"
	"gorm.io/gorm"
	"time"
)

type RoleResource struct {
	RoleId     uint `json:"role_id"`
	ResourceId uint `json:"resource_id"`
}

// Role 角色模型
// @Description 角色信息模型
type Role struct {
	BaseModel
	Name        string            `gorm:"size:50;not null;uniqueIndex:idx_role_name" json:"name" example:"admin" description:"角色名称"`
	Status      consts.RoleStatus `gorm:"type:tinyint;default:1;not null" json:"status" example:"1" description:"角色状态（1:启用 2:禁用）"`
	BuiltIn     bool              `gorm:"default:false" json:"built_in" description:"保护内置角色不被外部删除"`
	Description string            `gorm:"size:200;index:idx_role_desc" json:"description" example:"系统管理员" description:"角色描述"`
	Resources   []Resource        `gorm:"many2many:role_resources;" json:"resources" description:"角色可访问的资源（实际授权）"`
}

func (Role) TableName() string {
	return "roles"
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	r.CreatedAt = time.Now()
	return nil
}

func (r *Role) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now()
	return nil
}
