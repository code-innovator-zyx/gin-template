package rbac

import (
	"fmt"
	"gin-template/internal/core"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// 默认配置常量
const (
	DefaultAdminUsername = "admin"
	DefaultAdminPassword = "admin123"
	DefaultAdminEmail    = "admin@template.com"
	DefaultRoleName      = "超级管理员"
	DefaultRoleDesc      = "系统超级管理员，拥有所有权限"
)

// ProtectedRoute 受保护的路由信息（避免循环导入）
type ProtectedRoute struct {
	Resource       Resource // 资源信息
	PermissionCode string   // 权限组编码
	PermissionName string   // 权限组名称
	Description    string   // 资源描述
}

// InitializeRBAC 自动初始化 RBAC 权限系统
// 参数 routes: 从路由收集器获取的受保护路由列表
// 执行顺序：
// 1. 从路由声明中提取权限组并自动创建
// 2. 同步路由资源到数据库
// 3. 自动绑定资源到权限组
// 4. 创建超级管理员角色
// 5. 绑定所有权限组到超级管理员
// 6. 创建默认管理员用户并分配角色
func InitializeRBAC(routes []ProtectedRoute) error {
	db := core.MustNewDb()

	logrus.Info("开始初始化 RBAC 权限系统...")
	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. 从路由声明中提取并创建权限组
		if err := extractAndCreatePermissions(tx, routes); err != nil {
			return fmt.Errorf("创建权限组失败: %w", err)
		}

		// 2. 同步路由资源
		if err := syncRouteResources(tx, routes); err != nil {
			return fmt.Errorf("同步路由资源失败: %w", err)
		}

		// 3. 自动绑定资源到权限组
		if err := autoBindResourcesToPermissions(tx, routes); err != nil {
			return fmt.Errorf("绑定资源到权限组失败: %w", err)
		}

		// 4. 创建超级管理员角色
		adminRole, err := initializeAdminRole(tx)
		if err != nil {
			return fmt.Errorf("初始化超级管理员角色失败: %w", err)
		}

		// 5. 绑定所有权限组到超级管理员角色
		if err := bindAllPermissionsToRole(tx); err != nil {
			return fmt.Errorf("绑定权限到角色失败: %w", err)
		}

		// 6. 创建默认管理员用户
		if err := initializeAdminUser(tx, adminRole.ID); err != nil {
			return fmt.Errorf("初始化管理员用户失败: %w", err)
		}

		return nil
	})

	if err != nil {
		logrus.Errorf("RBAC 权限系统初始化失败: %v", err)
		return err
	}

	logrus.Info("✓ RBAC 权限系统初始化成功")
	logrus.Infof("✓ 默认管理员账号: %s / %s", DefaultAdminUsername, DefaultAdminPassword)
	return nil
}

// extractAndCreatePermissions 从路由声明中提取权限组并创建
func extractAndCreatePermissions(tx *gorm.DB, routes []ProtectedRoute) error {
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
		tx.Model(&Permission{}).Where("code = ?", code).Count(&count)

		if count == 0 {
			permission := &Permission{
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
func syncRouteResources(tx *gorm.DB, routes []ProtectedRoute) error {
	logrus.Info("  - 同步路由资源到数据库...")

	if len(routes) == 0 {
		logrus.Warn("    ! 未发现需要保护的路由")
		return nil
	}

	// 批量 upsert 资源
	for _, route := range routes {
		var resource Resource
		result := tx.Where("path = ? AND method = ?", route.Resource.Path, route.Resource.Method).
			First(&resource)

		if result.Error == gorm.ErrRecordNotFound {
			// 创建新资源
			newResource := Resource{
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
func autoBindResourcesToPermissions(tx *gorm.DB, routes []ProtectedRoute) error {
	logrus.Info("  - 自动绑定资源到权限组...")

	boundCount := 0

	for _, route := range routes {
		// 如果路由声明了权限组
		if route.PermissionCode != "" {
			// 查找对应的权限组
			var permission Permission
			if err := tx.Where("code = ?", route.PermissionCode).First(&permission).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					logrus.Warnf("    ! 路由 %s %s 指定的权限组 %s 不存在，跳过绑定",
						route.Resource.Method, route.Resource.Path, route.PermissionCode)
					continue
				}
				return err
			}

			// 更新资源的权限组绑定
			result := tx.Model(&Resource{}).
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
func initializeAdminRole(tx *gorm.DB) (*Role, error) {
	logrus.Info("  - 初始化超级管理员角色...")

	var role Role
	result := tx.Where("name = ?", DefaultRoleName).First(&role)

	if result.Error == gorm.ErrRecordNotFound {
		// 创建角色
		role = Role{
			Name:        DefaultRoleName,
			Description: DefaultRoleDesc,
		}
		if err := tx.Create(&role).Error; err != nil {
			return nil, err
		}
		logrus.Infof("    ✓ 创建角色: %s", DefaultRoleName)
	} else if result.Error != nil {
		return nil, result.Error
	} else {
		logrus.Debugf("    - 角色已存在: %s", DefaultRoleName)
	}

	return &role, nil
}

// bindAllPermissionsToRole 绑定所有权限组到超级管理员角色
func bindAllPermissionsToRole(tx *gorm.DB) error {
	logrus.Info("  - 绑定所有权限组到超级管理员角色...")

	// 获取超级管理员角色
	var adminRole Role
	if err := tx.Where("name = ?", DefaultRoleName).First(&adminRole).Error; err != nil {
		return fmt.Errorf("未找到超级管理员角色: %w", err)
	}

	// 获取所有权限组
	var permissions []Permission
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
func initializeAdminUser(tx *gorm.DB, roleID uint) error {
	logrus.Info("  - 初始化默认管理员用户...")

	var user User
	result := tx.Where("username = ?", DefaultAdminUsername).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		// 创建管理员用户
		user = User{
			Username: DefaultAdminUsername,
			Password: DefaultAdminPassword,
			Email:    DefaultAdminEmail,
			Status:   1,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		logrus.Infof("    ✓ 创建管理员用户: %s", DefaultAdminUsername)

		// 分配角色
		if err := tx.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)",
			user.ID, roleID).Error; err != nil {
			return fmt.Errorf("分配角色到用户失败: %w", err)
		}
		logrus.Info("    ✓ 分配超级管理员角色")
	} else if result.Error != nil {
		return result.Error
	} else {
		logrus.Debugf("    - 管理员用户已存在: %s", DefaultAdminUsername)

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
