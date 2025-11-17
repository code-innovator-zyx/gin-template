package service

import (
	"context"
	"gin-template/internal/core"
	"gorm.io/gorm"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/16 上午10:48
* @Package: 泛型仓库，提供通用 CRUD 和分页功能
 */

// BaseRepo  T: 需要操作的实体类型
type BaseRepo[T any] struct {
	Tx *gorm.DB
}

// NewBaseRepo 构造函数，可传入事务
// ctx: 上下文，可用于绑定请求生命周期
// tx: 可传入事务，如果不传则自动创建数据库连接
func NewBaseRepo[T any](ctx context.Context) *BaseRepo[T] {
	return &BaseRepo[T]{Tx: core.MustNewDbWithContext(ctx)}
}

// WithTx 返回带事务的 BaseRepo，可链式调用
// 示例：repo.WithTx(tx).UpdateByID(...)
func (r *BaseRepo[T]) WithTx(tx *gorm.DB) *BaseRepo[T] {
	return &BaseRepo[T]{Tx: tx}
}

// FindByID 根据 ID 获取单条记录，支持预加载关联字段
// preloads: 需要预加载的关联字段，如 "Roles", "Resources"
func (r *BaseRepo[T]) FindByID(id uint, preloads ...string) (*T, error) {
	db := r.Tx
	for _, p := range preloads {
		db = db.Preload(p)
	}
	var t T
	if err := db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// FindByIDs 根据 ID 列表批量查询
func (r *BaseRepo[T]) FindByIDs(ids []uint) ([]T, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []T
	if err := r.Tx.Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// FindAll 自定义查询条件获取多条记录
func (r *BaseRepo[T]) FindAll(limits int, scopes ...func(*gorm.DB) *gorm.DB) ([]T, error) {
	var list []T
	tx := r.Tx.Scopes(scopes...)
	if limits > 0 {
		tx = tx.Limit(limits)
	}
	if err := tx.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// FindOne 自定义查询条件查询一条记录
func (r *BaseRepo[T]) FindOne(scopes ...func(*gorm.DB) *gorm.DB) (*T, error) {
	var t T
	if err := r.Tx.Scopes(scopes...).Take(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// Exists 判断是否存在记录，支持传入多个查询条件（scope）
// 返回 true: 存在, false: 不存在
func (r *BaseRepo[T]) Exists(scopes ...func(*gorm.DB) *gorm.DB) (bool, error) {
	var exists bool
	err := r.Tx.Model(new(T)).
		Scopes(scopes...).
		Select("1").
		Limit(1).
		Scan(&exists).Error
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Create 插入单条记录
func (r *BaseRepo[T]) Create(t *T) error {
	return r.Tx.Create(t).Error
}

// BatchCreate 批量插入
func (r *BaseRepo[T]) BatchCreate(list []*T) error {
	if len(list) == 0 {
		return nil
	}
	return r.Tx.Create(&list).Error
}

// UpdateByID 根据 ID 更新指定字段
func (r *BaseRepo[T]) UpdateByID(id uint, data map[string]any) error {
	return r.Tx.Model(new(T)).Where("id = ?", id).Updates(data).Error
}

// UpdateSelected 安全更新选中字段，零值不会覆盖
// fields: 需要更新的字段
func (r *BaseRepo[T]) UpdateSelected(id uint, fields []string, data *T) error {
	return r.Tx.Model(new(T)).
		Select(fields).
		Where("id = ?", id).
		Updates(data).Error
}

// UpdateOmit 排除指定字段更新
// omit: 不需要更新的字段
func (r *BaseRepo[T]) UpdateOmit(id uint, omit []string, data *T) error {
	return r.Tx.Model(new(T)).
		Omit(omit...).
		Where("id = ?", id).
		Updates(data).Error
}

// DeleteByID 删除记录
func (r *BaseRepo[T]) DeleteByID(id uint) error {
	return r.Tx.Delete(new(T), id).Error
}

// PageResult 分页返回结果
type PageResult[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// PageQuery 分页请求参数
type PageQuery struct {
	Page     int    `json:"page"`      // 页码，从1开始
	PageSize int    `json:"page_size"` // 每页大小
	OrderBy  string `json:"order_by"`  // 排序字段，如 "id desc"
}

// List 分页查询
// scopes: 可传入查询条件函数
func (r *BaseRepo[T]) List(query PageQuery, scopes ...func(db *gorm.DB) *gorm.DB) (*PageResult[T], error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	db := r.Tx.Model(new(T)).Scopes(scopes...)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if query.OrderBy != "" {
		db = db.Order(query.OrderBy)
	}

	var list []T
	if err := db.Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&list).Error; err != nil {
		return nil, err
	}

	return &PageResult[T]{
		List:     list,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}
