package _interface

import (
	"context"
	"errors"
	"gorm.io/gorm"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/12/1 下午12:56
* @Package:
 */

// QueryOptions 查询选项函数
type QueryOptions struct {
	SelectFields []string               // 查询字段
	Preloads     []string               // 预加载字段
	OrderBy      string                 // 排序
	Conditions   map[string]interface{} // 筛选条件
	Page         int                    // 分页页码
	PageSize     int                    // 分页大小
}

type QueryOption = func(*QueryOptions)

// ApplyQueryOptions 应用查询选项（供实现层使用）
func ApplyQueryOptions(opts ...QueryOption) *QueryOptions {
	options := &QueryOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}
func WithSelectFields(fields ...string) QueryOption {
	return func(option *QueryOptions) {
		option.SelectFields = fields
	}
}
func WithPreloads(preloads ...string) QueryOption {
	return func(option *QueryOptions) {
		option.Preloads = preloads
	}
}
func WithOrderBy(orderBy string) QueryOption {
	return func(option *QueryOptions) {
		option.OrderBy = orderBy
	}
}
func WithConditions(conditions map[string]interface{}) QueryOption {
	return func(option *QueryOptions) {
		option.Conditions = conditions
	}
}

func WithPagination(page, pageSize int) QueryOption {
	return func(option *QueryOptions) {
		option.Page = page
		option.PageSize = pageSize
	}
}

// PageResult 分页结果
type PageResult[T any] struct {
	List      []T   `json:"list"`      // 数据列表
	Total     int64 `json:"total"`     // 总数
	Page      int   `json:"page"`      // 当前页
	PageSize  int   `json:"pageSize"`  // 每页大小
	TotalPage int   `json:"totalPage"` // 总页数
}

// IRepo 数据库常规操作
type IRepo[T IModel] interface {
	// ==================== 查询操作 ====================

	// FindByID 通过ID查询(支持预加载关联)
	// 示例: repo.FindByID(ctx, 1)
	FindByID(ctx context.Context, id uint, opts ...QueryOption) (*T, error)

	// FindByIDs 通过id列表查询(支持预加载关联)
	FindByIDs(ctx context.Context, id []uint, opts ...QueryOption) ([]T, error)

	// FindOne 条件查询单条记录
	FindOne(ctx context.Context, opts ...QueryOption) (*T, error)

	// List 条件查询列表(支持预加载和排序，传空条件查询全部)
	// 示例: repo.List(ctx, WithConditions(map[string]interface{}{"status": 1}), WithOrderBy("created_at desc"), WithPreloads("Roles"))
	// 查询全部: repo.List(ctx)
	List(ctx context.Context, opts ...QueryOption) ([]T, error)

	// FindPage 分页查询(支持条件、排序、预加载)
	// Page / PageSize 从 opts 的 WithPagination 注入
	FindPage(ctx context.Context, opts ...QueryOption) (*PageResult[T], error)

	// ==================== 创建操作 ====================

	// Create 创建单条记录
	Create(ctx context.Context, entity *T) error

	// CreateBatch 批量创建(支持指定批次大小，默认100)
	CreateBatch(ctx context.Context, entities []T, batchSize ...int) error

	// ==================== 更新操作 ====================

	// Update 更新记录(只更新非零值字段)
	Update(ctx context.Context, entity *T) error

	// UpdateByID 根据ID更新指定字段
	// 示例: repo.UpdateByID(ctx, 1, map[string]interface{}{"status": 1, "updated_at": time.Now()})
	UpdateByID(ctx context.Context, id uint, updates map[string]interface{}) error

	// UpdateByCondition 根据条件批量更新
	// 示例: repo.UpdateByCondition(ctx, map[string]interface{}{"status": 0}, map[string]interface{}{"status": 1})
	UpdateByCondition(ctx context.Context, condition map[string]interface{}, updates map[string]interface{}) error

	// ==================== 删除操作 ====================

	// Delete 删除记录(软删除或硬删除取决于模型定义)
	Delete(ctx context.Context, entity *T) error

	// DeleteByID 根据ID删除
	DeleteByID(ctx context.Context, id uint) error

	// DeleteByIDs 根据ID列表批量删除
	DeleteByIDs(ctx context.Context, ids []uint) error

	// DeleteByCondition 根据条件删除
	DeleteByCondition(ctx context.Context, condition map[string]interface{}) error

	// ==================== 统计操作 ====================

	// Count 统计记录数
	Count(ctx context.Context, condition map[string]interface{}) (int64, error)

	// Exists 检查记录是否存在
	Exists(ctx context.Context, condition map[string]interface{}) (bool, error)

	// ExistsByID 检查ID是否存在
	ExistsByID(ctx context.Context, id uint) (bool, error)

	// ==================== 高级操作 ====================

	// FirstOrCreate 查找或创建(不存在则创建)
	FirstOrCreate(ctx context.Context, condition map[string]interface{}, entity *T) error

	// ==================== 事务支持 ====================

	// Transaction 执行事务
	// 自动 commit / rollback
	Transaction(ctx context.Context, fn func(txRepo IRepo[T]) error) error
}

type Repo[T IModel] struct {
	DB *gorm.DB
}

func NewRepo[T IModel](db *gorm.DB) *Repo[T] {
	return &Repo[T]{
		DB: db,
	}
}

func (r *Repo[T]) apply(db *gorm.DB, opts ...QueryOption) *gorm.DB {
	o := ApplyQueryOptions(opts...)

	// 条件
	if len(o.Conditions) > 0 {
		db = db.Where(o.Conditions)
	}

	// select 字段
	if len(o.SelectFields) > 0 {
		db = db.Select(o.SelectFields)
	}

	// 排序
	if o.OrderBy != "" {
		db = db.Order(o.OrderBy)
	}

	// 预加载
	for _, preload := range o.Preloads {
		db = db.Preload(preload)
	}

	return db
}

func (r *Repo[T]) FindByID(ctx context.Context, id uint, opts ...QueryOption) (*T, error) {
	var entity T
	db := r.apply(r.DB.WithContext(ctx), opts...)
	if err := db.First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *Repo[T]) FindByIDs(ctx context.Context, ids []uint, opts ...QueryOption) ([]T, error) {
	if len(ids) == 0 {
		return []T{}, nil
	}
	list := make([]T, 0, len(ids))
	db := r.apply(r.DB.WithContext(ctx), opts...)
	if err := db.Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *Repo[T]) FindOne(ctx context.Context, opts ...QueryOption) (*T, error) {
	var entity T
	db := r.apply(r.DB.WithContext(ctx), opts...)
	if err := db.First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *Repo[T]) List(ctx context.Context, opts ...QueryOption) ([]T, error) {
	var list []T
	db := r.apply(r.DB.WithContext(ctx), opts...)
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *Repo[T]) FindPage(ctx context.Context, opts ...QueryOption) (*PageResult[T], error) {
	o := ApplyQueryOptions(opts...)

	if o.Page <= 0 {
		o.Page = 1
	}
	if o.PageSize <= 0 {
		o.PageSize = 10
	}
	base := r.DB.WithContext(ctx)
	countDB := base
	if len(o.Conditions) > 0 {
		countDB = countDB.Where(o.Conditions)
	}
	var total int64
	if err := countDB.Model(new(T)).Count(&total).Error; err != nil {
		return nil, err
	}
	// 2) total = 0 或 offset 超过总数，直接返回空列表
	offset := (o.Page - 1) * o.PageSize
	if total == 0 || int64(offset) >= total {
		return &PageResult[T]{
			List:      []T{},
			Total:     total,
			Page:      o.Page,
			PageSize:  o.PageSize,
			TotalPage: int((total + int64(o.PageSize) - 1) / int64(o.PageSize)),
		}, nil
	}

	db := r.apply(base, opts...)

	list := make([]T, 0, o.PageSize)

	// 分页
	if err := db.Offset(offset).Limit(o.PageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return &PageResult[T]{
		List:      list,
		Total:     total,
		Page:      o.Page,
		PageSize:  o.PageSize,
		TotalPage: int((total + int64(o.PageSize) - 1) / int64(o.PageSize)),
	}, nil
}

// ==================== 创建 ====================

func (r *Repo[T]) Create(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Create(entity).Error
}

func (r *Repo[T]) CreateBatch(ctx context.Context, entities []T, batchSize ...int) error {
	if len(entities) == 0 {
		return nil
	}
	size := 100
	if len(batchSize) > 0 && batchSize[0] > 0 {
		size = batchSize[0]
	}
	return r.DB.WithContext(ctx).CreateInBatches(entities, size).Error
}

// ==================== 更新 ====================

func (r *Repo[T]) Update(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Save(entity).Error
}

func (r *Repo[T]) UpdateByID(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.DB.WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(updates).Error
}

func (r *Repo[T]) UpdateByCondition(ctx context.Context, condition map[string]interface{}, updates map[string]interface{}) error {
	if len(condition) == 0 {
		return errors.New("update condition cannot be empty to prevent accidental update of all records")
	}
	return r.DB.WithContext(ctx).Model(new(T)).Where(condition).Updates(updates).Error
}

// ==================== 删除 ====================

func (r *Repo[T]) Delete(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Delete(entity).Error
}

func (r *Repo[T]) DeleteByID(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Delete(new(T), id).Error
}

func (r *Repo[T]) DeleteByIDs(ctx context.Context, ids []uint) error {
	return r.DB.WithContext(ctx).Delete(new(T), ids).Error
}

func (r *Repo[T]) DeleteByCondition(ctx context.Context, condition map[string]interface{}) error {
	if len(condition) == 0 {
		return errors.New("delete condition cannot be empty to prevent accidental deletion of all records")
	}
	return r.DB.WithContext(ctx).Where(condition).Delete(new(T)).Error
}

// ==================== 统计 ====================

func (r *Repo[T]) Count(ctx context.Context, condition map[string]interface{}) (int64, error) {
	var count int64
	db := r.DB.WithContext(ctx).Model(new(T))
	if len(condition) > 0 {
		db = db.Where(condition)
	}
	err := db.Count(&count).Error
	return count, err
}

func (r *Repo[T]) Exists(ctx context.Context, condition map[string]interface{}) (bool, error) {
	var count int64
	db := r.DB.WithContext(ctx).Model(new(T))
	if len(condition) > 0 {
		db = db.Where(condition)
	}
	err := db.Limit(1).Count(&count).Error
	return count > 0, err
}

func (r *Repo[T]) ExistsByID(ctx context.Context, id uint) (bool, error) {
	return r.Exists(ctx, map[string]interface{}{"id": id})
}

// ==================== 查找或创建 ====================

func (r *Repo[T]) FirstOrCreate(ctx context.Context, condition map[string]interface{}, entity *T) error {
	return r.DB.WithContext(ctx).Where(condition).FirstOrCreate(entity).Error
}

// ==================== 事务 ====================

func (r *Repo[T]) Transaction(ctx context.Context, fn func(txRepo IRepo[T]) error) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &Repo[T]{DB: tx}
		return fn(txRepo)
	})
}
