package rbac

import (
	"gorm.io/gorm"
	"time"
)

// Resource 资源模型 服务启动的时候会自动更新该表
type Resource struct {
	BaseModel
	Path         string `gorm:"size:200;not null;uniqueIndex:idx_path_method" json:"path" example:"/api/users" description:"资源路径"`
	Method       string `gorm:"size:10;not null;uniqueIndex:idx_path_method" json:"method" example:"GET" description:"HTTP方法"`
	Code         string `gorm:"size:50;not null;index:idx_code" json:"code" description:"唯一Code，前端权限控制使用的"`
	Description  string `gorm:"size:200" json:"description" example:"获取用户列表" description:"接口中文描述"`
	PermissionID *uint  `gorm:"index:idx_resource_permission" json:"permission_id" example:"1" description:"所属权限分组ID（仅用于UI展示分组）"`
	Roles        []Role `gorm:"many2many:role_resources;" json:"roles" description:"拥有该资源的角色（实际授权）"`
}

func (r *Resource) BeforeCreate(tx *gorm.DB) error {
	r.CreatedAt = time.Now()
	return nil
}

func (r *Resource) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now()
	return nil
}

// TableName 设置表名
func (Resource) TableName() string {
	return "resources"
}
