package rbac

import (
	"gin-template/internal/core"
	"time"
)

// Role 角色模型
// @Description 角色信息模型
type Role struct {
	ID          uint       `gorm:"primarykey" json:"id" example:"1" description:"角色ID"`
	Name        string     `gorm:"size:50;not null;uniqueIndex:idx_role_name" json:"name" example:"admin" description:"角色名称"`
	Description string     `gorm:"size:200;index:idx_role_desc" json:"description" example:"系统管理员" description:"角色描述"`
	Resources   []Resource `gorm:"many2many:role_resources;" json:"resources" description:"角色可访问的资源（实际授权）"`
	CreatedAt   time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z" description:"创建时间"`
	UpdatedAt   time.Time  `json:"updated_at" example:"2023-01-01T00:00:00Z" description:"更新时间"`
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
	tx := core.MustNewDb().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除角色
	if err := tx.Delete(&Role{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除角色关联
	if err := tx.Exec("DELETE FROM user_roles WHERE role_id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Exec("DELETE FROM role_resources WHERE role_id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetRoles 获取角色列表
func GetRoles(page, pageSize int) ([]Role, int64, error) {
	var roles []Role
	var total int64

	db := core.MustNewDb().Model(&Role{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}
