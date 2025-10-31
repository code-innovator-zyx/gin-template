package service

import (
	"errors"
	"gin-template/internal/core"

	"gin-template/internal/model"
	"gin-template/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// userService 用户服务
type userService struct{}

// UserService 用户服务实例
var UserService = new(userService)

// Create 创建用户
func (s *userService) Create(user *model.User) error {
	// 检查用户名是否已存在
	var count int64
	if err := core.MustNewDb().Model(&model.User{}).Where("username = ?", user.Username).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if user.Email != "" {
		if err := core.MustNewDb().Model(&model.User{}).Where("email = ?", user.Email).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("邮箱已存在")
		}
	}

	// 创建用户
	return core.MustNewDb().Create(user).Error
}

// GetByID 根据ID获取用户
func (s *userService) GetByID(id uint) (*model.User, error) {
	var user model.User
	if err := core.MustNewDb().First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (s *userService) GetByUsername(username string) (*model.User, error) {
	var user model.User
	if err := core.MustNewDb().Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// Login 用户登录
func (s *userService) Login(username, password string) (string, error) {
	user, err := s.GetByUsername(username)
	if err != nil {
		return "", err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("密码错误")
	}

	// 生成JWT令牌
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Update 更新用户信息
func (s *userService) Update(user *model.User) error {
	return core.MustNewDb().Save(user).Error
}

// Delete 删除用户
func (s *userService) Delete(id uint) error {
	return core.MustNewDb().Delete(&model.User{}, id).Error
}
