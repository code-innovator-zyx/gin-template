package service

import (
	"github.com/code-innovator-zyx/gin-template/core"
	"github.com/code-innovator-zyx/gin-template/internal/model/rbac"
)

// HasPermission 检查用户是否有权限访问指定路径和方法
func HasPermission(userID uint, path string, method string) bool {
	hasPermission, err := rbac.CheckPermission(userID, path, method)
	if err != nil {
		return false
	}
	return hasPermission
}

// GetUserRoles 获取用户角色列表
func GetUserRoles(userID uint) ([]rbac.Role, error) {
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

// AssignRoleToUser 为用户分配角色
func AssignRoleToUser(userID uint, roleID uint) error {
	return core.MustNewDb().Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID).Error
}

// RemoveRoleFromUser 从用户移除角色
func RemoveRoleFromUser(userID uint, roleID uint) error {
	return core.MustNewDb().Exec("DELETE FROM user_roles WHERE user_id = ? AND role_id = ?", userID, roleID).Error
}

// AssignPermissionToRole 为角色分配权限
func AssignPermissionToRole(roleID uint, permissionID uint) error {
	return core.MustNewDb().Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, permissionID).Error
}

// RemovePermissionFromRole 从角色移除权限
func RemovePermissionFromRole(roleID uint, permissionID uint) error {
	return core.MustNewDb().Exec("DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?", roleID, permissionID).Error
}

// AssignMenuToRole 为角色分配菜单
func AssignMenuToRole(roleID uint, menuID uint) error {
	return core.MustNewDb().Exec("INSERT INTO role_menus (role_id, menu_id) VALUES (?, ?)", roleID, menuID).Error
}

// RemoveMenuFromRole 从角色移除菜单
func RemoveMenuFromRole(roleID uint, menuID uint) error {
	return core.MustNewDb().Exec("DELETE FROM role_menus WHERE role_id = ? AND menu_id = ?", roleID, menuID).Error
}

// GetUserMenus 获取用户菜单列表
func GetUserMenus(userID uint) ([]rbac.Menu, error) {
	return rbac.GetUserMenus(userID)
}
