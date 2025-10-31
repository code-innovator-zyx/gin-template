package rbac

import (
	"github.com/code-innovator-zyx/gin-template/internal/core"
	"time"

	"gorm.io/gorm"
)

// Menu 菜单模型
// @Description 菜单信息模型
type Menu struct {
	ID        uint           `gorm:"primarykey" json:"id" example:"1" description:"菜单ID"`
	Name      string         `gorm:"size:50;not null" json:"name" example:"系统管理" description:"菜单名称"`
	Path      string         `gorm:"size:200" json:"path" example:"/system" description:"菜单路径"`
	Icon      string         `gorm:"size:50" json:"icon" example:"setting" description:"菜单图标"`
	ParentID  *uint          `gorm:"default:null" json:"parent_id" example:"0" description:"父菜单ID"`
	Sort      int            `gorm:"default:0" json:"sort" example:"1" description:"排序"`
	CreatedAt time.Time      `json:"created_at" description:"创建时间"`
	UpdatedAt time.Time      `json:"updated_at" description:"更新时间"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-" description:"删除时间"`
}

// GetMenuByID 根据ID获取菜单
func GetMenuByID(id uint) (*Menu, error) {
	var menu Menu
	err := core.MustNewDb().First(&menu, id).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

// CreateMenu 创建菜单
func CreateMenu(menu *Menu) error {
	return core.MustNewDb().Create(menu).Error
}

// UpdateMenu 更新菜单
func UpdateMenu(menu *Menu) error {
	return core.MustNewDb().Save(menu).Error
}

// DeleteMenu 删除菜单
func DeleteMenu(id uint) error {
	return core.MustNewDb().Delete(&Menu{}, id).Error
}

// GetUserMenus 获取用户菜单列表
func GetUserMenus(userID uint) ([]Menu, error) {
	var menus []Menu
	err := core.MustNewDb().Raw(`
		SELECT DISTINCT m.* FROM menus m
		JOIN role_menus rm ON m.id = rm.menu_id
		JOIN user_roles ur ON rm.role_id = ur.role_id
		WHERE ur.user_id = ?
		ORDER BY m.sort ASC
	`, userID).Find(&menus).Error

	if err != nil {
		return nil, err
	}
	return menus, nil
}
