package rbac

import (
	"github.com/code-innovator-zyx/gin-template/core"
	"time"

	"gorm.io/gorm"
)

// Role 角色模型
// @Description 角色信息模型
type Role struct {
	ID          uint           `gorm:"primarykey" json:"id" example:"1" description:"角色ID"`
	Name        string         `gorm:"size:50;not null;unique" json:"name" example:"admin" description:"角色名称"`
	Description string         `gorm:"size:200" json:"description" example:"系统管理员" description:"角色描述"`
	CreatedAt   time.Time      `json:"created_at" description:"创建时间"`
	UpdatedAt   time.Time      `json:"updated_at" description:"更新时间"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-" description:"删除时间"`
}

// TableName 设置表名
func (Role) TableName() string {
	return "roles"
}

// GetRoleByID 根据ID获取角色
func GetRoleByID(id uint) (*Role, error) {
	var role Role
	err := core.MustNewDb().First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleByName 根据名称获取角色
func GetRoleByName(name string) (*Role, error) {
	var role Role
	err := core.MustNewDb().Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// CreateRole 创建角色
func CreateRole(role *Role) error {
	return core.MustNewDb().Create(role).Error
}

// UpdateRole 更新角色
func UpdateRole(role *Role) error {
	return core.MustNewDb().Save(role).Error
}

// DeleteRole 删除角色
func DeleteRole(id uint) error {
	return core.MustNewDb().Delete(&Role{}, id).Error
}