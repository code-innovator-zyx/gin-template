package service

import (
	"gin-template/internal/core"
	"gin-template/internal/model/rbac"
	"sync"
)

// rbacService RBAC权限服务
type rbacService struct{}

var (
	rbacServiceOnce   sync.Once
	globalRbacService *rbacService
)

// GetRbacService 获取RBAC服务单例（懒加载，线程安全）
func GetRbacService() *rbacService {
	rbacServiceOnce.Do(func() {
		globalRbacService = &rbacService{}
	})
	return globalRbacService
}

// ==================== 角色管理 ====================

// GetAllRoles 获取所有角色
func (s *rbacService) GetAllRoles() ([]rbac.Role, error) {
	var roles []rbac.Role
	if err := core.MustNewDb().Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// GetRoleByID 根据ID获取角色
func (s *rbacService) GetRoleByID(id uint) (*rbac.Role, error) {
	return rbac.GetRoleByID(id)
}

// CreateRole 创建角色
func (s *rbacService) CreateRole(role *rbac.Role) error {
	return rbac.CreateRole(role)
}

// UpdateRole 更新角色
func (s *rbacService) UpdateRole(role *rbac.Role) error {
	return rbac.UpdateRole(role)
}

// DeleteRole 删除角色
func (s *rbacService) DeleteRole(id uint) error {
	return rbac.DeleteRole(id)
}

// ==================== 权限管理 ====================

// GetAllPermissions 获取所有权限
func (s *rbacService) GetAllPermissions() ([]rbac.Permission, error) {
	var permissions []rbac.Permission
	if err := core.MustNewDb().Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// CreatePermission 创建权限
func (s *rbacService) CreatePermission(permission *rbac.Permission) error {
	return rbac.CreatePermission(permission)
}

// ==================== 资源管理 ====================

// GetAllResources 获取所有资源
func (s *rbacService) GetAllResources() ([]rbac.Resource, error) {
	var resources []rbac.Resource
	if err := core.MustNewDb().Find(&resources).Error; err != nil {
		return nil, err
	}
	return resources, nil
}

// HasPermission 检查用户是否有权限访问指定路径和方法
func (s *rbacService) HasPermission(userID uint, path string, method string) bool {
	hasPermission, err := rbac.CheckPermission(userID, path, method)
	if err != nil {
		return false
	}
	return hasPermission
}

// GetUserRoles 获取用户角色列表（完整角色信息）
func (s *rbacService) GetUserRoles(userID uint) ([]rbac.Role, error) {
	var roles []rbac.Role
	err := core.MustNewDb().Raw(`
		SELECT r.* FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ?
	`, userID).Find(&roles).Error

	if err != nil {
		return nil, err
	}
	return roles, nil
}

// GetUserRoleRelations 获取用户角色关联记录
func (s *rbacService) GetUserRoleRelations(userID uint) ([]rbac.UserRole, error) {
	var userRoles []rbac.UserRole
	err := core.MustNewDb().Where("user_id = ?", userID).Find(&userRoles).Error
	if err != nil {
		return nil, err
	}
	return userRoles, nil
}

// CreateUserRole 创建用户角色关联
func (s *rbacService) CreateUserRole(userRole *rbac.UserRole) error {
	return core.MustNewDb().Create(userRole).Error
}

// AssignRoleToUser 为用户分配角色（通过ID）
func (s *rbacService) AssignRoleToUser(userID uint, roleID uint) error {
	return core.MustNewDb().Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID).Error
}

// RemoveRoleFromUser 从用户移除角色
func (s *rbacService) RemoveRoleFromUser(userID uint, roleID uint) error {
	return core.MustNewDb().Exec("DELETE FROM user_roles WHERE user_id = ? AND role_id = ?", userID, roleID).Error
}

// CreateRolePermission 创建角色权限关联
func (s *rbacService) CreateRolePermission(rolePermission *rbac.RolePermission) error {
	return core.MustNewDb().Create(rolePermission).Error
}

// AssignPermissionToRole 为角色分配权限（通过ID）
func (s *rbacService) AssignPermissionToRole(roleID uint, permissionID uint) error {
	return core.MustNewDb().Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, permissionID).Error
}

// RemovePermissionFromRole 从角色移除权限
func (s *rbacService) RemovePermissionFromRole(roleID uint, permissionID uint) error {
	return core.MustNewDb().Exec("DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?", roleID, permissionID).Error
}

// AssignMenuToRole 为角色分配菜单
func (s *rbacService) AssignMenuToRole(roleID uint, menuID uint) error {
	return core.MustNewDb().Exec("INSERT INTO role_menus (role_id, menu_id) VALUES (?, ?)", roleID, menuID).Error
}

// RemoveMenuFromRole 从角色移除菜单
func (s *rbacService) RemoveMenuFromRole(roleID uint, menuID uint) error {
	return core.MustNewDb().Exec("DELETE FROM role_menus WHERE role_id = ? AND menu_id = ?", roleID, menuID).Error
}
