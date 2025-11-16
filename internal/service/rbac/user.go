package rbac

import (
	"context"
	"errors"
	"fmt"
	"gin-template/internal/model/rbac"
	"gin-template/internal/service"
	types "gin-template/internal/types/rbac"
	"gin-template/pkg/consts"
	"gin-template/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"strings"
)

// userService 用户服务
type userService struct {
	ctx context.Context
	*service.BaseRepo[rbac.User]
}

// NewUserService 支持带事务传入
func NewUserService(ctx context.Context) *userService {
	return &userService{
		ctx:      ctx,
		BaseRepo: service.NewBaseRepo[rbac.User](ctx),
	}
}

// Register 注册用户
func (s *userService) Register(user *rbac.User) error {
	exist, err := s.Exists(func(db *gorm.DB) *gorm.DB {
		return db.Where("username = ?", user.Username)
	})
	if err != nil {
		return err
	}
	if exist {
		return errors.New("用户名已存在")
	}
	exist, err = s.Exists(func(db *gorm.DB) *gorm.DB {
		return db.Where("email = ?", user.Email)
	})
	if err != nil {
		return err
	}
	if exist {
		return errors.New("邮箱已存在")
	}

	// 创建用户
	return s.BaseRepo.Create(user)
}

// Login 用户登录（返回完整的 TokenPair）
func (s *userService) Login(account, password string) (*utils.TokenPair, error) {
	user, err := s.BaseRepo.FindOne(func(db *gorm.DB) *gorm.DB {
		return db.Where("username = ? OR email = ?", account, account)
	})
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

// List 获取用户列表
func (s *userService) List(request types.ListUserRequest) (*service.PageResult[rbac.User], error) {
	return s.BaseRepo.List(service.PageQuery{
		Page:     request.Page,
		PageSize: request.PageSize,
		OrderBy:  "created_at DESC",
	}, func(db *gorm.DB) *gorm.DB {
		if request.Username != "" {
			db = db.Where("username LIKE ?", request.Username+"%")
		}
		if request.Email != "" {
			db = db.Where("email = ?", request.Email)
		}
		if request.Status > 0 {
			db = db.Where("status = ?", request.Status)
		}
		if request.Gender > 0 {
			db = db.Where("gender = ?", request.Gender)
		}
		return db.Preload("Roles")
	})
}

// Update 更新用户信息
func (s *userService) Update(req types.UpsertUserRequest) error {
	// 0. 获取就得用户信息
	user, err := s.FindByID(req.Id)
	if nil != err {
		return err
	}
	// 1. 校验用户名是否存在
	if user.Username != req.Username {
		exist, err := s.Exists(func(db *gorm.DB) *gorm.DB {
			return db.Where("username = ? AND id <> ?", req.Username, req.Id)
		})
		if err != nil {
			return err
		}
		if exist {
			return fmt.Errorf("用户名 %s 已存在", req.Username)
		}
	}

	// 2. 校验用户邮箱是否冲突
	if user.Email != req.Email {
		exist, err := s.Exists(func(db *gorm.DB) *gorm.DB {
			return db.Where("email = ? AND id <> ?", req.Email, req.Id)
		})
		if err != nil {
			return err
		}
		if exist {
			return fmt.Errorf("邮箱 %s 已存在", req.Email)
		}
	}

	return s.Tx.Transaction(func(tx *gorm.DB) error {
		// 更新用户基础信息
		err = s.WithTx(tx).UpdateByID(req.Id, map[string]interface{}{
			"username": req.Username,
			"email":    req.Email,
			"gender":   req.Gender,
		})
		if err != nil {
			return err
		}
		var roles []rbac.Role
		// 获取所有角色列表
		if len(req.Roles) != 0 {
			roles, err = NewRoleService(s.ctx).FindByIDs(req.Roles)
			if err != nil {
				return err
			}
		}
		// 更新用户角色
		if err = tx.Model(&user).Association("Roles").Replace(roles); err != nil {
			return fmt.Errorf("更新用户角色失败: %w", err)
		}
		return nil
	})
}

func (s *userService) Create(req types.UpsertUserRequest) error {
	exist, err := s.Exists(func(db *gorm.DB) *gorm.DB {
		return db.Where("username = ?", req.Username)
	})
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("用户名 %s 已存在", req.Username)
	}
	exist, err = s.Exists(func(db *gorm.DB) *gorm.DB {
		return db.Where("email = ?", req.Email)
	})
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("邮箱 %s 已存在", req.Email)
	}
	roles, err := NewRoleService(s.ctx).FindAll(0)
	if err != nil {
		return err
	}
	// ==== 创建 ====
	user := rbac.User{
		Username: req.Username,
		Password: strings.Split(req.Email, "@")[0],
		Email:    req.Email,
		Gender:   req.Gender,
		Roles:    roles,
		Status:   consts.UserStatusActive,
	}
	return s.BaseRepo.Create(&user)

}

// DeleteByID 删除记录
func (s *userService) DeleteByID(id uint) error {
	user, err := s.FindByID(id)
	if err != nil {
		return err
	}
	return s.Tx.Transaction(
		func(tx *gorm.DB) error {
			err = tx.Model(&user).Association("Roles").Clear()
			if err != nil {
				return err
			}
			return tx.Delete(user).Error
		})
}
