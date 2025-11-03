package rbac

import (
	"gin-template/internal/core"
	"time"

	"gorm.io/gorm"
)

// Permission 权限组模型 (用户自己管理，哪些接口需要权限，自己去管理，这样更方便，可以随时取消)
type Permission struct {
	ID        uint           `gorm:"primarykey" json:"id" example:"1" description:"权限ID"`
	Name      string         `gorm:"size:50;not null;uniqueIndex:idx_perm_name" json:"name" example:"用户管理" description:"权限资源组中文名"`
	Code      string         `gorm:"size:50;not null;uniqueIndex:idx_perm_code" json:"code" example:"user:manage" description:"权限编码"`
	Resources []Resource     `gorm:"foreignKey:PermissionID" json:"resources" description:"权限组绑定的资源(一对多)"`
	Roles     []Role         `gorm:"many2many:role_permissions;" json:"roles" description:"拥有该权限的角色"`
	CreatedAt time.Time      `json:"created_at" example:"2023-01-01T00:00:00Z" description:"创建时间"`
	UpdatedAt time.Time      `json:"updated_at" example:"2023-01-01T00:00:00Z" description:"更新时间"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-" description:"删除时间"`
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
// 如果资源未被任何权限组管理（游离状态），则任何人都无法访问（安全优先）
func CheckPermission(userID uint, path string, method string) (bool, error) {
	// 首先检查资源是否被管理
	var isManaged bool
	var permissionID uint
	err := core.MustNewDb().Raw(`
		SELECT is_managed, permission_id FROM resources 
		WHERE path = ? AND method = ? LIMIT 1
	`, path, method).Row().Scan(&isManaged, &permissionID)

	if err != nil {
		return false, err
	}

	// 如果资源未被管理或未分配权限组，出于安全考虑，任何人都无法访问
	if !isManaged || permissionID == 0 {
		return false, nil
	}

	// 检查用户是否有权限访问该资源
	var count int64
	err = core.MustNewDb().Raw(`
		SELECT COUNT(*) FROM resources r
		JOIN role_permissions rp ON r.permission_id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ? AND r.path = ? AND r.method = ?
	`, userID, path, method).Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
