package service

import (
	"context"
	"errors"
	"fmt"
	"gin-template/internal/core"
	"gin-template/internal/model/rbac"
	types "gin-template/internal/types/rbac"
	"gin-template/pkg/consts"
	"gin-template/pkg/orm"
	"gin-template/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"strings"
	"sync"
)

// userService 用户服务
type userService struct{}

var (
	userServiceOnce   sync.Once
	globalUserService *userService
)

// GetUserService 获取用户服务单例（懒加载，线程安全）
func GetUserService() *userService {
	userServiceOnce.Do(func() {
		globalUserService = &userService{}
	})
	return globalUserService
}

// Register 注册用户
func (s *userService) Register(ctx context.Context, user *rbac.User) error {
	// 检查用户名是否已存在
	var count int64
	if err := core.MustNewDbWithContext(ctx).Model(&rbac.User{}).Where("username = ?", user.Username).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if user.Email != "" {
		if err := core.MustNewDbWithContext(ctx).Model(&rbac.User{}).Where("email = ?", user.Email).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("邮箱已存在")
		}
	}

	// 创建用户
	return core.MustNewDbWithContext(ctx).Create(user).Error
}

// GetByID 根据ID获取用户
func (s *userService) GetByID(ctx context.Context, id uint) (*rbac.User, error) {
	var user rbac.User
	if err := core.MustNewDbWithContext(ctx).Preload("Roles").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (s *userService) GetByUsername(ctx context.Context, username string) (*rbac.User, error) {
	var user rbac.User
	if err := core.MustNewDbWithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// GetByAccount 根据用户名或邮箱获取用户
func (s *userService) GetByAccount(ctx context.Context, account string) (*rbac.User, error) {
	var user rbac.User
	db := core.MustNewDbWithContext(ctx)
	if err := db.Where("username = ? OR email = ?", account, account).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	return &user, nil
}

// Login 用户登录（返回完整的 TokenPair）
func (s *userService) Login(ctx context.Context, account, password string) (*utils.TokenPair, error) {
	user, err := s.GetByAccount(ctx, account)
	if err != nil {
		return nil, err
	}
	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("密码错误")
	}
	// 生成JWT令牌对
	jwtManager := utils.GetJWTManager()
	tokenPair, err := jwtManager.GenerateTokenPair(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}
	return tokenPair, nil
}

// Update 更新用户信息
func (s *userService) Update(ctx context.Context, user *rbac.User) error {
	return core.MustNewDbWithContext(ctx).Save(user).Error
}

// Delete 删除用户
func (s *userService) Delete(ctx context.Context, id uint) error {
	return core.MustNewDbWithContext(ctx).Delete(&rbac.User{}, id).Error
}

// List 获取用户列表
func (s *userService) List(ctx context.Context, request types.ListUserRequest) (*orm.PageResult[rbac.User], error) {
	tx := core.MustNewDbWithContext(ctx)
	if request.Username != "" {
		tx = tx.Where("username LIKE ?", request.Username+"%")
	}
	if request.Email != "" {
		tx = tx.Where("email = ?", request.Email)
	}
	if request.Status > 0 {
		tx = tx.Where("status = ?", request.Status)
	}
	if request.Gender > 0 {
		tx = tx.Where("gender = ?", request.Gender)
	}
	tx = tx.Preload("Roles")

	return orm.Paginate[rbac.User](ctx, tx, orm.PageQuery{
		Page:     request.Page,
		PageSize: request.PageSize,
		OrderBy:  "-created_at",
	})
}
func (s *userService) UpsertUser(ctx context.Context, req types.UpsertUserRequest) error {
	db := core.MustNewDbWithContext(ctx)

	return db.Transaction(func(tx *gorm.DB) error {
		var count int64
		tx.Model(&rbac.User{}).
			Where("username = ? AND id <> ?", req.Username, req.Id).
			Count(&count)
		if count > 0 {
			return fmt.Errorf("用户名 %s 已存在", req.Username)
		}

		tx.Model(&rbac.User{}).
			Where("email = ? AND id <> ?", req.Email, req.Id).
			Count(&count)
		if count > 0 {
			return fmt.Errorf("邮箱 %s 已存在", req.Email)
		}

		// ==== 加载角色 ====
		var roles []rbac.Role
		if len(req.Roles) > 0 {
			rs, err := GetRbacService().GetAllRoles(ctx, req.Roles...)
			if err != nil {
				return err
			}
			roles = rs
		}

		var user rbac.User
		// ==== 更新 ====
		if req.Id != 0 {
			if err := tx.First(&user, req.Id).Error; err != nil {
				return fmt.Errorf("用户不存在: %w", err)
			}
			updates := map[string]interface{}{
				"username": req.Username,
				"email":    req.Email,
				"gender":   req.Gender,
			}
			if err := tx.Model(&user).Updates(updates).Error; err != nil {
				return fmt.Errorf("更新用户失败: %w", err)
			}
		} else {
			// ==== 创建 ====
			user = rbac.User{
				Username: req.Username,
				Password: strings.Split(req.Email, "@")[0],
				Email:    req.Email,
				Gender:   req.Gender,
				Status:   consts.UserStatusActive,
			}
			if err := tx.Create(&user).Error; err != nil {
				return fmt.Errorf("创建用户失败: %w", err)
			}
		}
		// ==== 处理用户角色 ====
		if err := tx.Model(&user).Association("Roles").Replace(roles); err != nil {
			return fmt.Errorf("更新用户角色失败: %w", err)
		}

		return nil
	})
}
