package service

import (
	"context"
	"fmt"
	"gin-template/internal/core"
	"gin-template/internal/model/rbac"
	"sync"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
func (s *rbacService) GetAllRoles(ctx context.Context) ([]rbac.Role, error) {
	var roles []rbac.Role
	if err := core.MustNewDbWithContext(ctx).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// GetRoleByID 根据ID获取角色
func (s *rbacService) GetRoleByID(ctx context.Context, id uint) (*rbac.Role, error) {
	var role rbac.Role
	if err := core.MustNewDbWithContext(ctx).First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// CreateRole 创建角色
func (s *rbacService) CreateRole(ctx context.Context, role *rbac.Role) error {
	return core.MustNewDbWithContext(ctx).Create(role).Error
}

// UpdateRole 更新角色
func (s *rbacService) UpdateRole(ctx context.Context, role *rbac.Role) error {
	return core.MustNewDbWithContext(ctx).Save(role).Error
}

// DeleteRole 删除角色
func (s *rbacService) DeleteRole(ctx context.Context, id uint) error {
	return core.MustNewDbWithContext(ctx).Delete(&rbac.Role{}, id).Error
}

// ==================== 权限管理 ====================

// GetAllPermissions 获取所有权限
func (s *rbacService) GetAllPermissions(ctx context.Context) ([]rbac.Permission, error) {
	var permissions []rbac.Permission
	if err := core.MustNewDbWithContext(ctx).Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// CreatePermission 创建权限
func (s *rbacService) CreatePermission(ctx context.Context, permission *rbac.Permission) error {
	return core.MustNewDbWithContext(ctx).Create(permission).Error
}

// ==================== 资源管理 ====================

// GetAllResources 获取所有资源
func (s *rbacService) GetAllResources(ctx context.Context) ([]rbac.Resource, error) {
	var resources []rbac.Resource
	if err := core.MustNewDbWithContext(ctx).Find(&resources).Error; err != nil {
		return nil, err
	}
	return resources, nil
}

// CheckUserPermission 检查用户是否有权限访问指定路径和方法
func (s *rbacService) CheckUserPermission(ctx context.Context, userID uint, path string, method string) (bool, error) {
	// cache check First
	exist, err := GetCacheService().CheckUserPermission(ctx, userID, path, method, s.GetUserResources)
	if err == nil {
		return exist, nil
	}
	var count int64
	err = core.MustNewDbWithContext(ctx).Raw(`
		SELECT COUNT(*) FROM resources res
		JOIN permissions p ON res.permission_id = p.id
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ? AND res.path = ? AND res.method = ?
	`, userID, path, method).Scan(&count).Error

	return count > 0, err
}

// GetUserRoles 获取用户角色列表（完整角色信息）
func (s *rbacService) GetUserRoles(ctx context.Context, userID uint) ([]rbac.Role, error) {
	var roles []rbac.Role
	err := core.MustNewDbWithContext(ctx).Raw(`
		SELECT r.* FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ?
	`, userID).Find(&roles).Error

	if err != nil {
		return nil, err
	}
	return roles, nil
}

// GetUserPermissions 获取用户的所有权限（通过角色关联）
func (s *rbacService) GetUserPermissions(ctx context.Context, userID uint) ([]rbac.Permission, error) {
	var permissions []rbac.Permission
	err := core.MustNewDbWithContext(ctx).Raw(`
		SELECT DISTINCT p.* FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ?
	`, userID).Find(&permissions).Error

	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// GetUserResources 获取用户可访问的资源列表（通过角色和权限关联）
func (s *rbacService) GetUserResources(ctx context.Context, userID uint) ([]rbac.Resource, error) {
	var resources []rbac.Resource
	err := core.MustNewDbWithContext(ctx).Raw(`
		SELECT DISTINCT res.* FROM resources res
		JOIN permissions p ON res.permission_id = p.id
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ?
		ORDER BY res.path, res.method
	`, userID).Find(&resources).Error

	if err != nil {
		return nil, err
	}
	return resources, nil
}

// GetUserRoleRelations 获取用户角色关联记录
func (s *rbacService) GetUserRoleRelations(ctx context.Context, userID uint) ([]rbac.UserRole, error) {
	var userRoles []rbac.UserRole
	err := core.MustNewDbWithContext(ctx).Where("user_id = ?", userID).Find(&userRoles).Error
	if err != nil {
		return nil, err
	}
	return userRoles, nil
}

// CreateUserRole 创建用户角色关联
func (s *rbacService) CreateUserRole(ctx context.Context, userRole *rbac.UserRole) error {
	return core.MustNewDbWithContext(ctx).Create(userRole).Error
}

// AssignRoleToUser 为用户分配角色（通过ID）
func (s *rbacService) AssignRoleToUser(ctx context.Context, userID uint, roleID uint) error {
	return core.MustNewDbWithContext(ctx).Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID).Error
}

// RemoveRoleFromUser 从用户移除角色
func (s *rbacService) RemoveRoleFromUser(ctx context.Context, userID uint, roleID uint) error {
	return core.MustNewDbWithContext(ctx).Exec("DELETE FROM user_roles WHERE user_id = ? AND role_id = ?", userID, roleID).Error
}

// CreateRolePermission 创建角色权限关联
func (s *rbacService) CreateRolePermission(ctx context.Context, rolePermission *rbac.RolePermission) error {
	return core.MustNewDbWithContext(ctx).Create(rolePermission).Error
}

// AssignPermissionToRole 为角色分配权限（通过ID）
func (s *rbacService) AssignPermissionToRole(ctx context.Context, roleID uint, permissionID uint) error {
	return core.MustNewDbWithContext(ctx).Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, permissionID).Error
}

// RemovePermissionFromRole 从角色移除权限
func (s *rbacService) RemovePermissionFromRole(ctx context.Context, roleID uint, permissionID uint) error {
	return core.MustNewDbWithContext(ctx).Exec("DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?", roleID, permissionID).Error
}

// AssignMenuToRole 为角色分配菜单
func (s *rbacService) AssignMenuToRole(ctx context.Context, roleID uint, menuID uint) error {
	return core.MustNewDbWithContext(ctx).Exec("INSERT INTO role_menus (role_id, menu_id) VALUES (?, ?)", roleID, menuID).Error
}

// RemoveMenuFromRole 从角色移除菜单
func (s *rbacService) RemoveMenuFromRole(ctx context.Context, roleID uint, menuID uint) error {
	return core.MustNewDbWithContext(ctx).Exec("DELETE FROM role_menus WHERE role_id = ? AND menu_id = ?", roleID, menuID).Error
}

// ==================== RBAC 系统初始化 ====================

// ProtectedRoute 受保护的路由信息
type ProtectedRoute struct {
	Resource       rbac.Resource // 资源信息
	PermissionCode string        // 权限组编码
	PermissionName string        // 权限组名称
	Description    string        // 资源描述
}

// RBACInitConfig RBAC初始化配置
type RBACInitConfig struct {
	AdminUsername  string // 管理员用户名
	AdminPassword  string // 管理员密码
	AdminEmail     string // 管理员邮箱
	AdminRoleName  string // 管理员角色名称
	AdminRoleDesc  string // 管理员角色描述
	EnableAutoInit bool   // 是否启用自动初始化
}

// InitializeRBAC 自动初始化 RBAC 权限系统
// 参数 ctx: 上下文，用于控制请求生命周期
// 参数 routes: 从路由收集器获取的受保护路由列表
// 参数 config: RBAC初始化配置（如果为nil则使用默认配置）
// 执行顺序：
// 1. 从路由声明中提取权限组并自动创建
// 2. 同步路由资源到数据库
// 3. 自动绑定资源到权限组
// 4. 创建超级管理员角色
// 5. 绑定所有权限组到超级管理员
// 6. 创建默认管理员用户并分配角色
func (s *rbacService) InitializeRBAC(ctx context.Context, routes []ProtectedRoute, config *RBACInitConfig) error {
	// 如果配置为空或禁用自动初始化，则跳过
	if config != nil && !config.EnableAutoInit {
		logrus.Info("RBAC 自动初始化已禁用")
		return nil
	}
	db := core.MustNewDbWithContext(ctx)

	logrus.Info("开始初始化 RBAC 权限系统...")
	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. 从路由声明中提取并创建权限组
		if err := s.extractAndCreatePermissions(tx, routes); err != nil {
			return fmt.Errorf("创建权限组失败: %w", err)
		}

		// 2. 同步路由资源
		if err := s.syncRouteResources(tx, routes); err != nil {
			return fmt.Errorf("同步路由资源失败: %w", err)
		}

		// 3. 自动绑定资源到权限组
		if err := s.autoBindResourcesToPermissions(tx, routes); err != nil {
			return fmt.Errorf("绑定资源到权限组失败: %w", err)
		}

		// 4. 创建超级管理员角色
		adminRole, err := s.initializeAdminRole(tx, config)
		if err != nil {
			return fmt.Errorf("初始化超级管理员角色失败: %w", err)
		}

		// 5. 绑定所有权限组到超级管理员角色
		if err := s.bindAllPermissionsToRole(tx, config); err != nil {
			return fmt.Errorf("绑定权限到角色失败: %w", err)
		}

		// 6. 创建默认管理员用户
		if err := s.initializeAdminUser(tx, adminRole.ID, config); err != nil {
			return fmt.Errorf("初始化管理员用户失败: %w", err)
		}

		return nil
	})

	if err != nil {
		logrus.Errorf("RBAC 权限系统初始化失败: %v", err)
		return err
	}

	logrus.Info("✓ RBAC 权限系统初始化成功")
	logrus.Infof("✓ 默认管理员账号: %s / %s", config.AdminUsername, config.AdminPassword)
	return nil
}

// extractAndCreatePermissions 从路由声明中提取权限组并创建
func (s *rbacService) extractAndCreatePermissions(tx *gorm.DB, routes []ProtectedRoute) error {
	logrus.Info("  - 从路由声明中提取并创建权限组...")

	// 使用 map 去重，收集所有唯一的权限组
	permissionMap := make(map[string]string) // code -> name
	for _, route := range routes {
		if route.PermissionCode != "" && route.PermissionName != "" {
			permissionMap[route.PermissionCode] = route.PermissionName
		}
	}
	if len(permissionMap) == 0 {
		logrus.Warn("    ! 未发现任何权限组声明")
		return nil
	}
	// 创建权限组
	createdCount := 0
	for code, name := range permissionMap {
		var count int64
		tx.Model(&rbac.Permission{}).Where("code = ?", code).Count(&count)

		if count == 0 {
			permission := &rbac.Permission{
				Code: code,
				Name: name,
			}
			if err := tx.Create(permission).Error; err != nil {
				return fmt.Errorf("创建权限组 %s 失败: %w", code, err)
			}
			logrus.Infof("    ✓ 创建权限组: %s (%s)", name, code)
			createdCount++
		} else {
			logrus.Debugf("    - 权限组已存在: %s (%s)", name, code)
		}
	}

	logrus.Infof("    ✓ 从路由声明中发现 %d 个权限组，新建 %d 个", len(permissionMap), createdCount)
	return nil
}

// syncRouteResources 同步路由资源到数据库
func (s *rbacService) syncRouteResources(tx *gorm.DB, routes []ProtectedRoute) error {
	logrus.Info("  - 同步路由资源到数据库...")

	if len(routes) == 0 {
		logrus.Warn("    ! 未发现需要保护的路由")
		return nil
	}

	// 批量 upsert 资源
	for _, route := range routes {
		var resource rbac.Resource
		result := tx.Where("path = ? AND method = ?", route.Resource.Path, route.Resource.Method).
			First(&resource)

		if result.Error == gorm.ErrRecordNotFound {
			// 创建新资源
			newResource := rbac.Resource{
				Path:        route.Resource.Path,
				Method:      route.Resource.Method,
				Description: route.Description,
			}
			if err := tx.Create(&newResource).Error; err != nil {
				return fmt.Errorf("创建资源失败 %s %s: %w", route.Resource.Method, route.Resource.Path, err)
			}
		} else if result.Error != nil {
			return result.Error
		} else {
			// 更新描述（如果有变化）
			if resource.Description != route.Description && route.Description != "" {
				tx.Model(&resource).Update("description", route.Description)
			}
		}
	}

	logrus.Infof("    ✓ 同步了 %d 个路由资源", len(routes))
	return nil
}

// autoBindResourcesToPermissions 自动绑定资源到权限组
func (s *rbacService) autoBindResourcesToPermissions(tx *gorm.DB, routes []ProtectedRoute) error {
	logrus.Info("  - 自动绑定资源到权限组...")

	boundCount := 0

	for _, route := range routes {
		// 如果路由声明了权限组
		if route.PermissionCode != "" {
			// 查找对应的权限组
			var permission rbac.Permission
			if err := tx.Where("code = ?", route.PermissionCode).First(&permission).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					logrus.Warnf("    ! 路由 %s %s 指定的权限组 %s 不存在，跳过绑定",
						route.Resource.Method, route.Resource.Path, route.PermissionCode)
					continue
				}
				return err
			}

			// 更新资源的权限组绑定
			result := tx.Model(&rbac.Resource{}).
				Where("path = ? AND method = ?", route.Resource.Path, route.Resource.Method).
				Updates(map[string]interface{}{
					"permission_id": permission.ID,
				})

			if result.Error != nil {
				return fmt.Errorf("绑定资源到权限组失败: %w", result.Error)
			}

			if result.RowsAffected > 0 {
				boundCount++
				logrus.Debugf("    ✓ 绑定资源 %s %s 到权限组 %s",
					route.Resource.Method, route.Resource.Path, route.PermissionCode)
			}
		}
	}

	logrus.Infof("    ✓ 成功绑定 %d 个资源到权限组", boundCount)
	return nil
}

// initializeAdminRole 初始化超级管理员角色
func (s *rbacService) initializeAdminRole(tx *gorm.DB, config *RBACInitConfig) (*rbac.Role, error) {
	logrus.Info("  - 初始化超级管理员角色...")

	var role rbac.Role
	result := tx.Where("name = ?", config.AdminRoleName).First(&role)

	if result.Error == gorm.ErrRecordNotFound {
		// 创建角色
		role = rbac.Role{
			Name:        config.AdminRoleName,
			Description: config.AdminRoleDesc,
		}
		if err := tx.Create(&role).Error; err != nil {
			return nil, err
		}
		logrus.Infof("    ✓ 创建角色: %s", config.AdminRoleName)
	} else if result.Error != nil {
		return nil, result.Error
	} else {
		logrus.Debugf("    - 角色已存在: %s", config.AdminRoleName)
	}

	return &role, nil
}

// bindAllPermissionsToRole 绑定所有权限组到超级管理员角色
func (s *rbacService) bindAllPermissionsToRole(tx *gorm.DB, config *RBACInitConfig) error {
	logrus.Info("  - 绑定所有权限组到超级管理员角色...")

	// 获取超级管理员角色
	var adminRole rbac.Role
	if err := tx.Where("name = ?", config.AdminRoleName).First(&adminRole).Error; err != nil {
		return fmt.Errorf("未找到超级管理员角色: %w", err)
	}

	// 获取所有权限组
	var permissions []rbac.Permission
	if err := tx.Find(&permissions).Error; err != nil {
		return err
	}

	if len(permissions) == 0 {
		logrus.Warn("    ! 未发现任何权限组")
		return nil
	}

	boundCount := 0
	for _, perm := range permissions {
		// 检查是否已经绑定
		var count int64
		tx.Table("role_permissions").
			Where("role_id = ? AND permission_id = ?", adminRole.ID, perm.ID).
			Count(&count)

		if count == 0 {
			// 创建绑定关系
			if err := tx.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)",
				adminRole.ID, perm.ID).Error; err != nil {
				return fmt.Errorf("绑定权限 %s 到角色失败: %w", perm.Code, err)
			}
			boundCount++
			logrus.Debugf("    ✓ 绑定权限组: %s (%s)", perm.Name, perm.Code)
		}
	}
	logrus.Infof("    ✓ 成功绑定 %d 个权限组到超级管理员", boundCount)
	return nil
}

// initializeAdminUser 初始化管理员用户
func (s *rbacService) initializeAdminUser(tx *gorm.DB, roleID uint, config *RBACInitConfig) error {
	logrus.Info("  - 初始化默认管理员用户...")

	var user rbac.User
	result := tx.Where("username = ?", config.AdminUsername).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		// 创建管理员用户
		user = rbac.User{
			Username: config.AdminUsername,
			Password: config.AdminPassword,
			Email:    config.AdminEmail,
			Status:   1,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		logrus.Infof("    ✓ 创建管理员用户: %s", config.AdminUsername)

		// 分配角色
		if err := tx.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)",
			user.ID, roleID).Error; err != nil {
			return fmt.Errorf("分配角色到用户失败: %w", err)
		}
		logrus.Info("    ✓ 分配超级管理员角色")
	} else if result.Error != nil {
		return result.Error
	} else {
		logrus.Debugf("    - 管理员用户已存在: %s", config.AdminUsername)

		// 确保用户有超级管理员角色
		var count int64
		tx.Table("user_roles").
			Where("user_id = ? AND role_id = ?", user.ID, roleID).
			Count(&count)

		if count == 0 {
			if err := tx.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)",
				user.ID, roleID).Error; err != nil {
				return fmt.Errorf("分配角色到用户失败: %w", err)
			}
			logrus.Info("    ✓ 补充分配超级管理员角色")
		}
	}

	return nil
}
