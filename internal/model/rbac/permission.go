package rbac

import (
	"gin-template/internal/core"
	"time"

	"gorm.io/gorm"
)

// Permission 权限模型
type Permission struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"size:50;not null;unique" json:"name"`
	Description string         `gorm:"size:200" json:"description"`
	Path        string         `gorm:"size:200" json:"path"`
	Method      string         `gorm:"size:10" json:"method"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 设置表名
func (Permission) TableName() string {
	return "permissions"
}

// CreatePermission 创建权限
func CreatePermission(permission *Permission) error {
	return core.MustNewDb().Create(permission).Error
}

// CheckPermission 检查用户是否有权限访问指定路径和方法
func CheckPermission(userID uint, path string, method string) (bool, error) {
	var count int64
	err := core.MustNewDb().Raw(`
		SELECT COUNT(*) FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ? AND p.path = ? AND p.method = ?
	`, userID, path, method).Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
