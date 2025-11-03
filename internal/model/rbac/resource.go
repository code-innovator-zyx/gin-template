package rbac

import (
	"gin-template/internal/core"
	"time"

	"gorm.io/gorm/clause"
)

// Resource 资源模型 服务启动的时候会自动更新该表
type Resource struct {
	ID           uint      `gorm:"primarykey" json:"id" example:"1" description:"资源ID"`
	Path         string    `gorm:"size:200;not null;index:idx_resource_path" json:"path" example:"/api/users" description:"资源路径"`
	Method       string    `gorm:"size:10;not null;index:idx_resource_method" json:"method" example:"GET" description:"HTTP方法"`
	Description  string    `gorm:"size:200" json:"description" example:"获取用户列表" description:"接口中文描述"`
	IsManaged    bool      `gorm:"default:false;index:idx_resource_managed" json:"is_managed" example:"false" description:"是否被权限组管理"`
	PermissionID *uint     `gorm:"index:idx_resource_permission" json:"permission_id" example:"0" description:"所属权限ID"`
	CreatedAt    time.Time `json:"created_at" example:"2023-01-01T00:00:00Z" description:"创建时间"`
	UpdatedAt    time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z" description:"更新时间"`
}

// TableName 设置表名
func (Resource) TableName() string {
	return "resources"
}

// CreateResource 创建资源
func CreateResource(resource *Resource) error {
	return core.MustNewDb().Create(resource).Error
}

// GetUnmanagedResources 获取所有未被管理的资源
func GetUnmanagedResources() ([]Resource, error) {
	var resources []Resource
	err := core.MustNewDb().Where("is_managed = ?", false).Find(&resources).Error
	return resources, err
}

// UpdateResourceStatus 更新资源的管理状态
func UpdateResourceStatus(resourceID uint, isManaged bool) error {
	return core.MustNewDb().Model(&Resource{}).Where("id = ?", resourceID).Update("is_managed", isManaged).Error
}

// UpsertResource 更新或创建资源
func UpsertResource(resources []Resource) error {
	db := core.MustNewDb()

	// 使用 clause.OnConflict 实现 upsert 操作
	// 当 path 和 method 冲突时，更新 description 字段
	err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "path"},
			{Name: "method"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"description"}),
	}).Create(&resources).Error

	return err
}
