package service

import (
	"context"
	"errors"
	"gin-template/internal/core"
	"gin-template/internal/model/rbac"
	"gin-template/pkg/utils"
	"sync"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

// Create 创建用户
func (s *userService) Create(ctx context.Context, user *rbac.User) error {
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
	if err := core.MustNewDbWithContext(ctx).First(&user, id).Error; err != nil {
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

// Login 用户登录（返回完整的 TokenPair）
func (s *userService) Login(ctx context.Context, username, password string) (*utils.TokenPair, error) {
	user, err := s.GetByUsername(ctx, username)
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
