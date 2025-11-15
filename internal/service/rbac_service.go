package service

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"gin-template/internal/core"
	"gin-template/internal/model/rbac"
	types "gin-template/internal/types/rbac"
	"gin-template/pkg/consts"
	"gin-template/pkg/orm"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"maps"
	"slices"
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

// ListRoles 获取所有角色
func (s *rbacService) ListRoles(ctx context.Context, request types.ListRoleRequest) (*orm.PageResult[rbac.Role], error) {
	tx := core.MustNewDbWithContext(ctx)
	if request.Name != "" {
		tx.Where("name LIKE ?", "%"+request.Name+"%")
	}
	if request.Status > 0 {
		tx.Where("status = ?", request.Status)
	}
	return orm.Paginate[rbac.Role](ctx, tx, orm.PageQuery{
		Page:     request.Page,
		PageSize: request.PageSize,
		OrderBy:  "-created_at",
	})
}

// GetAllRoles 获取所有角色
func (s *rbacService) GetAllRoles(ctx context.Context) ([]rbac.Role, error) {
	var roles []rbac.Role
	if err := core.MustNewDbWithContext(ctx).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// GetRoleByID 根据ID获取角色
func (s *rbacService) GetRoleByID(ctx context.Context, id uint, preloads ...string) (*rbac.Role, error) {
	var role rbac.Role
	db := core.MustNewDbWithContext(ctx)

	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	if err := db.First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// CreateRole 创建角色
func (s *rbacService) CreateRole(ctx context.Context, role *rbac.Role) error {
	return core.MustNewDbWithContext(ctx).Create(role).Error
}

// UpdateRole 更新角色（不使用事务）
func (s *rbacService) UpdateRole(ctx context.Context, role *rbac.Role) error {
	return core.MustNewDbWithContext(ctx).Save(role).Error
}

// UpdateRoleWithTx 更新角色（使用事务）
func (s *rbacService) UpdateRoleWithTx(tx *gorm.DB, role *rbac.Role) error {
	return tx.Save(role).Error
}

// CheckRoleNameExists 检查角色名称是否存在
func (s *rbacService) CheckRoleNameExists(ctx context.Context, name string, excludeID uint) (bool, error) {
	var count int64
	query := core.MustNewDbWithContext(ctx).Model(&rbac.Role{}).Where("name = ?", name)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateRoleResourcesWithTx 更新角色的资源绑定（使用事务）
func (s *rbacService) UpdateRoleResourcesWithTx(tx *gorm.DB, roleID uint, resourceIDs []uint) error {
	// 1. 删除角色的所有旧资源绑定
	if err := tx.Exec("DELETE FROM role_resources WHERE role_id = ?", roleID).Error; err != nil {
		return err
	}

	// 2. 批量插入新的资源绑定
	if len(resourceIDs) > 0 {
		roleResources := make([]rbac.RoleResource, 0, len(resourceIDs))
		for _, resID := range resourceIDs {
			roleResources = append(roleResources, rbac.RoleResource{
				RoleID:     roleID,
				ResourceID: resID,
			})
		}
		if err := tx.CreateInBatches(&roleResources, 100).Error; err != nil {
			return err
		}
	}
	return nil

}

// DeleteRole 删除角色
func (s *rbacService) DeleteRole(ctx context.Context, id uint) error {
	return core.MustNewDbWithContext(ctx).Delete(&rbac.Role{}, id).Error
}

// ==================== 权限分组管理（仅用于UI展示） ====================

// GetAllPermissions 获取所有权限分组（带资源列表）
func (s *rbacService) GetAllPermissions(ctx context.Context) ([]rbac.Permission, error) {
	var permissions []rbac.Permission
	if err := core.MustNewDbWithContext(ctx).Preload("Resources").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// CreatePermission 创建权限分组
func (s *rbacService) CreatePermission(ctx context.Context, permission *rbac.Permission) error {
	return core.MustNewDbWithContext(ctx).Create(permission).Error
}

// CheckUserPermission 检查用户是否有权限访问指定路径和方法
// 新架构：Permission 仅作为逻辑分组，实际授权通过 role_resources
func (s *rbacService) CheckUserPermission(ctx context.Context, userID uint, path string, method string) (bool, error) {
	// cache check First
	exist, err := GetCacheService().CheckUserPermission(ctx, userID, path, method, s.GetUserResources)
	if err == nil {
		return exist, nil
	}

	// 直接检查 role_resources（不再查询 role_permissions）
	var count int64
	err = core.MustNewDbWithContext(ctx).Raw(`
		SELECT COUNT(*) FROM resources res
		JOIN role_resources rr ON res.id = rr.resource_id
		JOIN user_roles ur ON rr.role_id = ur.role_id
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

// GetUserPermissionGroups 获取用户的权限分组（基于用户拥有的资源）
// 返回用户有资源的 Permission
func (s *rbacService) GetUserPermissionGroups(ctx context.Context, userID uint) ([]rbac.Permission, error) {
	var rows []struct {
		PermissionID   uint   `json:"permission_id"`
		PermissionName string `json:"permission_name"`
		PermissionCode string `json:"permission_code"`
		ResourceID     uint   `json:"resource_id"`
		Path           string `json:"path"`
		Code           string `json:"code"`
		Method         string `json:"method"`
		Description    string `json:"description"`
	}
	err := core.MustNewDbWithContext(ctx).Raw(`
		SELECT DISTINCT
			p.id   AS permission_id,
			p.name AS permission_name,
			p.code AS permission_code,
			res.id AS resource_id,
			res.path,
			res.method,
			res.code,
			res.description
		FROM permissions p
		JOIN resources res ON p.id = res.permission_id
		JOIN role_resources rr ON res.id = rr.resource_id
		JOIN user_roles ur ON rr.role_id = ur.role_id
		WHERE ur.user_id = ?
		ORDER BY p.code, res.path, res.method
	`, userID).Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	permMap := make(map[uint]rbac.Permission)
	for _, row := range rows {
		p, ok := permMap[row.PermissionID]
		if !ok {
			p = rbac.Permission{
				ID:   row.PermissionID,
				Name: row.PermissionName,
				Code: row.PermissionCode,
			}
			permMap[row.PermissionID] = p
		}
		p.Resources = append(p.Resources, rbac.Resource{
			ID:          row.ResourceID,
			Path:        row.Path,
			Method:      row.Method,
			Code:        row.Code,
			Description: row.Description,
		})
		permMap[row.PermissionID] = p
	}
	return slices.SortedFunc(maps.Values(permMap), func(permission rbac.Permission, permission2 rbac.Permission) int {
		return cmp.Compare(permission.ID, permission2.ID)
	}), nil
}

// GetUserResources 获取用户可访问的资源列表（直接通过 role_resources）
func (s *rbacService) GetUserResources(ctx context.Context, userID uint) ([]rbac.Resource, error) {
	var resources []rbac.Resource
	err := core.MustNewDbWithContext(ctx).Raw(`
		SELECT DISTINCT res.* FROM resources res
		JOIN role_resources rr ON res.id = rr.resource_id
		JOIN user_roles ur ON rr.role_id = ur.role_id
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

// AssignResourceToRole 为角色分配资源
func (s *rbacService) AssignResourceToRole(ctx context.Context, roleID uint, resourceID uint) error {
	return core.MustNewDbWithContext(ctx).Exec("INSERT INTO role_resources (role_id, resource_id) VALUES (?, ?)", roleID, resourceID).Error
}

// RemoveResourceFromRole 从角色移除资源
func (s *rbacService) RemoveResourceFromRole(ctx context.Context, roleID uint, resourceID uint) error {
	return core.MustNewDbWithContext(ctx).Exec("DELETE FROM role_resources WHERE role_id = ? AND resource_id = ?", roleID, resourceID).Error
}

// GetRoleResources 获取角色的所有资源
func (s *rbacService) GetRoleResources(ctx context.Context, roleID uint) ([]rbac.Resource, error) {
	var resources []rbac.Resource
	err := core.MustNewDbWithContext(ctx).Raw(`
		SELECT res.* FROM resources res
		JOIN role_resources rr ON res.id = rr.resource_id
		WHERE rr.role_id = ?
		ORDER BY res.permission_id, res.path, res.method
	`, roleID).Find(&resources).Error

	if err != nil {
		return nil, err
	}
	return resources, nil
}

// GetUsersWithRole 获取拥有指定角色的所有用户ID
func (s *rbacService) GetUsersWithRole(ctx context.Context, roleID uint) ([]uint, error) {
	var userIDs []uint
	err := core.MustNewDbWithContext(ctx).
		Table("user_roles").
		Select("user_id").
		Where("role_id = ?", roleID).
		Find(&userIDs).Error

	if err != nil {
		return nil, err
	}
	return userIDs, nil
}

// GetResources 获取资源列表
func (s *rbacService) GetResources(ctx context.Context, ids []uint) ([]rbac.Resource, error) {
	var resources []rbac.Resource
	err := core.MustNewDbWithContext(ctx).Raw(`
		SELECT * FROM resources where id IN ?
	`, ids).Find(&resources).Error

	if err != nil {
		return nil, err
	}
	return resources, nil
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
// 新架构执行顺序：
// 1. 从路由声明中提取并创建权限分组（Permission - 仅用于UI展示）
// 2. 同步路由资源到数据库
// 3. 自动绑定资源到权限分组（仅用于展示分组）
// 4. 创建超级管理员角色
// 5. 绑定所有资源到超级管理员角色（role_resources - 实际授权）
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
		// 1. 从路由声明中提取并创建权限分组（用于UI展示）
		if err := s.extractAndCreatePermissions(tx, routes); err != nil {
			return fmt.Errorf("创建权限分组失败: %w", err)
		}

		// 2. 同步路由资源
		if err := s.syncRouteResources(tx, routes); err != nil {
			return fmt.Errorf("同步路由资源失败: %w", err)
		}

		// 3. 自动绑定资源到权限分组（仅用于UI展示分组）
		if err := s.autoBindResourcesToPermissions(tx, routes); err != nil {
			return fmt.Errorf("绑定资源到权限分组失败: %w", err)
		}

		// 4. 创建超级管理员角色
		adminRole, err := s.initializeAdminRole(tx, config)
		if err != nil {
			return fmt.Errorf("初始化超级管理员角色失败: %w", err)
		}

		// 5. 绑定所有资源到超级管理员角色（实际授权）
		if err := s.bindAllResourcesToRole(tx, adminRole.ID); err != nil {
			return fmt.Errorf("绑定资源到角色失败: %w", err)
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

// extractAndCreatePermissions 从路由声明中提取权限分组并创建（仅用于UI展示）
func (s *rbacService) extractAndCreatePermissions(tx *gorm.DB, routes []ProtectedRoute) error {
	logrus.Info("  - 从路由声明中提取并创建权限分组（用于UI展示）...")
	permissionMap := make(map[string]string) // code -> name
	for _, route := range routes {
		if route.PermissionCode != "" && route.PermissionName != "" {
			permissionMap[route.PermissionCode] = route.PermissionName
		}
	}
	if len(permissionMap) == 0 {
		logrus.Warn("    ! 未发现任何权限分组声明")
		return nil
	}
	var dbPerms []rbac.Permission
	if err := tx.Find(&dbPerms).Error; err != nil {
		return err
	}
	dbPermMap := make(map[string]rbac.Permission) // code -> Permission
	for _, p := range dbPerms {
		dbPermMap[p.Code] = p
	}
	var toInsert []rbac.Permission
	var toDelete []uint

	for code, name := range permissionMap {
		if _, exists := dbPermMap[code]; !exists {
			toInsert = append(toInsert, rbac.Permission{
				Code: code,
				Name: name,
			})
		}
	}
	for code, perm := range dbPermMap {
		if _, exists := permissionMap[code]; !exists {
			toDelete = append(toDelete, perm.ID)
		}
	}
	// 批量插入新增的
	if len(toInsert) > 0 {
		if err := tx.Create(&toInsert).Error; err != nil {
			return fmt.Errorf("批量创建权限分组失败: %w", err)
		}
		logrus.Infof("    ✓ 新增 %d 个权限分组", len(toInsert))
	}

	// 批量删除无效的
	if len(toDelete) > 0 {
		if err := tx.Where("id IN ?", toDelete).Delete(&rbac.Permission{}).Error; err != nil {
			return fmt.Errorf("批量删除无效权限分组失败: %w", err)
		}
		logrus.Infof("    ✓ 删除 %d 个已失效权限分组", len(toDelete))
	}

	logrus.Infof("    ✓ 从路由声明中发现 %d 个权限分组，新建 %d 个 ,删除 %d 个", len(permissionMap), len(toInsert), len(toDelete))
	return nil
}

// syncRouteResources 同步路由资源到数据库（保留 permission_id 用于UI分组）
func (s *rbacService) syncRouteResources(tx *gorm.DB, routes []ProtectedRoute) error {
	logrus.Info("  - 同步路由资源到数据库...")

	if len(routes) == 0 {
		logrus.Warn("    ! 未发现需要保护的路由")
		return nil
	}

	routeMap := make(map[string]rbac.Resource)
	for _, rt := range routes {
		key := rt.Resource.Path + "|" + rt.Resource.Method
		routeMap[key] = rt.Resource
	}

	var dbResources []rbac.Resource
	if err := tx.Find(&dbResources).Error; err != nil {
		return err
	}

	dbMap := make(map[string]rbac.Resource)
	for _, r := range dbResources {
		key := r.Path + "|" + r.Method
		dbMap[key] = r
	}

	var toInsert []rbac.Resource
	var toUpdate []rbac.Resource
	var toDelete []uint

	for key, res := range routeMap {
		if dbRes, exists := dbMap[key]; !exists {
			toInsert = append(toInsert, res)
		} else {
			res.ID = dbRes.ID
			res.PermissionID = dbRes.PermissionID
			toUpdate = append(toUpdate, res)
		}
	}

	for key, res := range dbMap {
		if _, exists := routeMap[key]; !exists {
			toDelete = append(toDelete, res.ID)
		}
	}

	if len(toInsert) > 0 {
		if err := tx.Create(&toInsert).Error; err != nil {
			return fmt.Errorf("批量创建资源失败: %w", err)
		}
		logrus.Infof("    ✓ 新增 %d 个资源", len(toInsert))
	}

	// 5) 批量更新
	for _, res := range toUpdate {
		if err := tx.Model(&rbac.Resource{}).Where("id = ?", res.ID).
			Omit("permission_id").
			Updates(res).Error; err != nil {
			return fmt.Errorf("更新资源失败: %w", err)
		}
	}

	if len(toDelete) > 0 {
		if err := tx.Where("id IN ?", toDelete).Delete(&rbac.Resource{}).Error; err != nil {
			return fmt.Errorf("批量删除旧资源失败: %w", err)
		}
		logrus.Infof("    ✓ 删除 %d 个已失效资源", len(toDelete))
	}

	logrus.Infof("    ✓ 资源同步完成: 新增 %d ，更新 %d ，删除 %d", len(toInsert), len(toUpdate), len(toDelete))
	return nil
}

// autoBindResourcesToPermissions 自动绑定资源到权限分组（仅用于UI展示分组，不影响实际授权）
func (s *rbacService) autoBindResourcesToPermissions(tx *gorm.DB, routes []ProtectedRoute) error {
	logrus.Info("  - 自动绑定资源到权限分组（仅用于UI展示）...")

	boundCount := 0
	for _, route := range routes {
		// 如果路由声明了权限分组
		if route.PermissionCode != "" {
			// 查找对应的权限分组
			var permission rbac.Permission
			if err := tx.Where("code = ?", route.PermissionCode).First(&permission).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					logrus.Warnf("    ! 路由 %s %s 指定的权限分组 %s 不存在，跳过绑定",
						route.Resource.Method, route.Resource.Path, route.PermissionCode)
					continue
				}
				return err
			}
			// 更新资源的权限分组绑定（仅用于UI展示）
			result := tx.Model(&rbac.Resource{}).
				Where("path = ? AND method = ?", route.Resource.Path, route.Resource.Method).
				Updates(map[string]interface{}{
					"permission_id": permission.ID,
				})
			if result.Error != nil {
				return fmt.Errorf("绑定资源到权限分组失败: %w", result.Error)
			}
			if result.RowsAffected > 0 {
				boundCount++
				logrus.Debugf("    ✓ 绑定资源 %s %s 到权限分组 %s（用于UI展示）",
					route.Resource.Method, route.Resource.Path, route.PermissionCode)
			}
		}
	}
	logrus.Infof("    ✓ 成功绑定 %d 个资源到权限分组", boundCount)
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
			Status:      consts.ROLESTATUS_ACTIVE,
			BuiltIn:     true, // 系统创建的资源，不允许任何人修改
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

// bindAllResourcesToRole 绑定所有资源到超级管理员角色（实际授权）
func (s *rbacService) bindAllResourcesToRole(tx *gorm.DB, roleID uint) error {
	logrus.Info("  - 绑定所有资源到超级管理员角色...")

	// 获取所有资源
	var resources []rbac.Resource
	if err := tx.Find(&resources).Error; err != nil {
		return err
	}

	if len(resources) == 0 {
		logrus.Warn("    ! 未发现任何资源")
		return nil
	}

	boundCount := 0
	for _, res := range resources {
		// 检查是否已经绑定
		var count int64
		tx.Table("role_resources").
			Where("role_id = ? AND resource_id = ?", roleID, res.ID).
			Count(&count)

		if count == 0 {
			// 创建绑定关系
			if err := tx.Exec("INSERT INTO role_resources (role_id, resource_id) VALUES (?, ?)",
				roleID, res.ID).Error; err != nil {
				return fmt.Errorf("绑定资源 %s %s 到角色失败: %w", res.Method, res.Path, err)
			}
			boundCount++
			logrus.Debugf("    ✓ 绑定资源: %s %s", res.Method, res.Path)
		}
	}
	logrus.Infof("    ✓ 成功绑定 %d 个资源到超级管理员", boundCount)
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
			Status:   consts.UserStatusActive,
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
