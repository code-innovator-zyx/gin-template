package services

import (
	"context"
	"fmt"
	"gin-admin/internal/model/rbac"
	"gin-admin/pkg/consts"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// rbacService RBAC权限服务
type rbacService struct {
}

func NewRbacService() *rbacService {
	return &rbacService{}
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
func (s *rbacService) InitializeRBAC(routes []ProtectedRoute, config *RBACInitConfig) error {
	// 如果配置为空或禁用自动初始化，则跳过
	if config != nil && !config.EnableAutoInit {
		logrus.Info("RBAC 自动初始化已禁用")
		return nil
	}
	db := SvcContext.Db

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
		_ = SvcContext.CacheService.ClearAllPermissions(context.TODO())
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

	// -------- 1) 收集唯一权限 code 和唯一资源 key --------
	permCodesSet := make(map[string]struct{})
	routeKeySet := make(map[string]struct{})
	pathSet := make(map[string]struct{})

	for _, rt := range routes {
		if rt.PermissionCode != "" {
			permCodesSet[rt.PermissionCode] = struct{}{}
		}
		key := rt.Resource.Path + "|" + rt.Resource.Method
		routeKeySet[key] = struct{}{}
		pathSet[rt.Resource.Path] = struct{}{}
	}

	if len(permCodesSet) == 0 || len(routeKeySet) == 0 {
		logrus.Warn("    ! 未发现需要绑定的权限分组或资源")
		return nil
	}

	// -------- 2) 查询权限分组 --------
	permCodes := make([]string, 0, len(permCodesSet))
	for code := range permCodesSet {
		permCodes = append(permCodes, code)
	}

	var permissions []rbac.Permission
	if err := tx.Where("code IN ?", permCodes).Find(&permissions).Error; err != nil {
		return fmt.Errorf("查询权限分组失败: %w", err)
	}

	permCodeToID := make(map[string]uint, len(permissions))
	for _, p := range permissions {
		permCodeToID[p.Code] = p.ID
	}

	// -------- 3) 查询相关资源（按 path 过滤）--------
	paths := make([]string, 0, len(pathSet))
	for p := range pathSet {
		paths = append(paths, p)
	}

	var dbResources []rbac.Resource
	if err := tx.Where("path IN ?", paths).Find(&dbResources).Error; err != nil {
		return fmt.Errorf("查询资源失败: %w", err)
	}

	resKeyToID := make(map[string]uint, len(dbResources))
	for _, r := range dbResources {
		key := r.Path + "|" + r.Method
		if _, ok := routeKeySet[key]; ok {
			resKeyToID[key] = r.ID
		}
	}

	// -------- 4) 构建 updates 映射表 --------
	updates := make(map[uint]uint)
	for _, rt := range routes {
		if rt.PermissionCode == "" {
			continue
		}
		pid, ok := permCodeToID[rt.PermissionCode]
		if !ok {
			logrus.Warnf("    ! 路由 %s %s 指定的权限分组 %s 不存在，跳过绑定",
				rt.Resource.Method, rt.Resource.Path, rt.PermissionCode)
			continue
		}
		key := rt.Resource.Path + "|" + rt.Resource.Method
		rid, ok := resKeyToID[key]
		if !ok {
			logrus.Warnf("    ! 资源 %s %s 不存在，跳过绑定", rt.Resource.Method, rt.Resource.Path)
			continue
		}
		updates[rid] = pid
	}

	if len(updates) == 0 {
		logrus.Info("    - 无需要绑定的资源")
		return nil
	}

	// -------- 5) 构建 CASE WHEN 批量 UPDATE --------
	ids := make([]uint, 0, len(updates))
	caseSQL := "CASE resources.id "
	args := make([]interface{}, 0, len(updates)*2+1)

	for id, pid := range updates {
		caseSQL += "WHEN ? THEN ? "
		args = append(args, id, pid)
		ids = append(ids, id)
	}

	caseSQL += "END"

	// 注意：这里必须使用 IN ? 而不是 IN (?)！
	args = append(args, ids)

	sql := "UPDATE resources SET permission_id = " + caseSQL + " WHERE id IN ?"

	if err := tx.Exec(sql, args...).Error; err != nil {
		return fmt.Errorf("批量绑定资源到权限分组失败: %w", err)
	}

	logrus.Infof("    ✓ 成功绑定 %d 个资源到权限分组", len(updates))
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

	var list = make([]rbac.RoleResource, 0, len(resources))
	for _, res := range resources {
		list = append(list, rbac.RoleResource{
			RoleId:     roleID,
			ResourceId: res.ID,
		})
	}
	if err := tx.Table("role_resources").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "role_id"}, {Name: "resource_id"}},
		DoNothing: true,
	}).Create(&list).Error; err != nil {
		return err
	}
	logrus.Infof("    ✓ 成功绑定 %d 个资源到超级管理员", len(list))
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
			BuiltIn:  true,
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
