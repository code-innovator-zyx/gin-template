package rbac

import (
	"gin-template/pkg/consts"
	"golang.org/x/crypto/bcrypt"
	"time"

	"gorm.io/gorm"
)

// User 用户模型
// @Description 用户信息模型
type User struct {
	ID        uint              `gorm:"primarykey" json:"id" example:"1" description:"用户ID"`
	Username  string            `gorm:"size:50;not null;uniqueIndex" json:"username" example:"johndoe" description:"用户名"`
	Password  string            `gorm:"size:100;not null" json:"-" description:"密码"`
	Email     string            `gorm:"size:100;uniqueIndex" json:"email" example:"john@example.com" description:"邮箱"`
	Avatar    string            `gorm:"size:255" json:"avatar" example:"https://example.com/avatar.jpg" description:"头像URL"`
	BuiltIn   bool              `gorm:"default:false" json:"built_in" description:"保护内置用户不被外部删除"`
	Gender    consts.Gender     `gorm:"type:tinyint;default:0;not null" json:"gender" example:"1"`
	Status    consts.UserStatus `gorm:"type:tinyint;default:1;not null" json:"status" example:"1" description:"用户状态"`
	Roles     []Role            `gorm:"many2many:user_roles;" json:"roles" description:"用户角色"`
	CreatedAt time.Time         `json:"created_at" example:"2023-01-01T00:00:00Z" description:"创建时间"`
	UpdatedAt time.Time         `json:"updated_at" example:"2023-01-01T00:00:00Z" description:"更新时间"`
}

// BeforeSave 保存前的钩子函数
func (u *User) BeforeSave(tx *gorm.DB) error {
	// 如果密码已经是哈希值，则不再加密
	if len(u.Password) == 60 && u.Password[0:4] == "$2a$" {
		return nil
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

// CheckPassword 检查密码是否正确
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
