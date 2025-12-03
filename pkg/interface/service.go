package _interface

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"gin-admin/pkg/cache"
	"gorm.io/gorm"
	"time"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025 2025/12/2 下午1:19
* @Package: Service 层 - 在 Repo 基础上增加缓存能力
 */

// IService Service 接口
type IService[T IModel] interface {
	IRepo[T]
	// ClearCache 清空该模型的所有缓存
	ClearCache(ctx context.Context) error
}

// Service 实现，在 Repo 基础上增加缓存
type Service[T IModel] struct {
	Repo     IRepo[T]
	DB       *gorm.DB
	cache    cache.ICache
	cacheTTL time.Duration // 缓存过期时间
}

// NewService 创建 Service 实例
func NewService[T IModel](db *gorm.DB, cache cache.ICache) *Service[T] {
	return &Service[T]{
		Repo:     NewRepo[T](db),
		cache:    cache,
		DB:       db,
		cacheTTL: 5 * time.Minute,
	}
}

// ==================== 内部辅助方法 ====================

// cacheKeyPrefix 缓存键前缀
func (s *Service[T]) cacheKeyPrefix() string {
	model := new(T)
	return fmt.Sprintf("model:%s:", (*model).TableName())
}

// cacheKey 生成缓存键
func (s *Service[T]) cacheKey(suffix string) string {
	return fmt.Sprintf("%s%s", s.cacheKeyPrefix(), suffix)
}

// getFromCache 从缓存获取数据
func (s *Service[T]) getFromCache(ctx context.Context, key string, dest interface{}) bool {
	if s.cache == nil {
		return false
	}

	err := s.cache.Get(ctx, key, dest)
	if err != nil {
		// 缓存未命中或错误，都返回 false，降级到数据库
		return false
	}
	return true
}

// setToCache 设置缓存
func (s *Service[T]) setToCache(ctx context.Context, key string, value interface{}) {
	if s.cache == nil {
		return
	}
	// 缓存失败不影响业务
	_ = s.cache.Set(ctx, key, value, s.cacheTTL)
}

// invalidateIDCache 使单个ID的缓存失效
func (s *Service[T]) invalidateIDCache(ctx context.Context, id uint) {
	if s.cache == nil {
		return
	}
	// 删除该ID的所有查询缓存（不同选项可能有多个缓存键）
	_ = s.cache.DeletePrefix(ctx, s.cacheKey(fmt.Sprintf("id:%d:", id)))
}

// invalidateListCache 使列表和分页缓存失效
func (s *Service[T]) invalidateListCache(ctx context.Context) {
	if s.cache == nil {
		return
	}
	_ = s.cache.DeletePrefix(ctx, s.cacheKey("list:"))
	_ = s.cache.DeletePrefix(ctx, s.cacheKey("page:"))
	_ = s.cache.DeletePrefix(ctx, s.cacheKey("one:"))
}

// ClearCache 清空该模型的所有缓存
func (s *Service[T]) ClearCache(ctx context.Context) error {
	if s.cache == nil {
		return nil
	}
	return s.cache.DeletePrefix(ctx, s.cacheKeyPrefix())
}

// serializeOpts 序列化查询选项为字符串（用于缓存键，使用 MD5 缩短长度）
func (s *Service[T]) serializeOpts(opts ...QueryOption) string {
	if len(opts) == 0 {
		return "default"
	}

	options := ApplyQueryOptions(opts...)
	data, _ := json.Marshal(options)
	hash := md5.Sum(data)
	return fmt.Sprintf("%x", hash)
}

// ==================== 查询操作（按需缓存）====================

// FindByID 通过ID查询 - 缓存
func (s *Service[T]) FindByID(ctx context.Context, id uint, opts ...QueryOption) (*T, error) {
	optsKey := s.serializeOpts(opts...)
	cacheKey := s.cacheKey(fmt.Sprintf("id:%d:%s", id, optsKey))

	// 尝试从缓存获取
	var entity T
	if s.getFromCache(ctx, cacheKey, &entity) {
		return &entity, nil
	}

	// 缓存未命中，从数据库获取
	result, err := s.Repo.FindByID(ctx, id, opts...)
	if err != nil {
		return nil, err
	}

	// 设置缓存
	s.setToCache(ctx, cacheKey, result)
	return result, nil
}

// FindByIDs 批量查询
func (s *Service[T]) FindByIDs(ctx context.Context, ids []uint, opts ...QueryOption) ([]T, error) {
	return s.Repo.FindByIDs(ctx, ids, opts...)
}

// FindOne 条件查询单条
func (s *Service[T]) FindOne(ctx context.Context, opts ...QueryOption) (*T, error) {
	optsKey := s.serializeOpts(opts...)
	cacheKey := s.cacheKey(fmt.Sprintf("one:%s", optsKey))

	var entity T
	if s.getFromCache(ctx, cacheKey, &entity) {
		return &entity, nil
	}

	result, err := s.Repo.FindOne(ctx, opts...)
	if err != nil {
		return nil, err
	}

	s.setToCache(ctx, cacheKey, result)
	return result, nil
}

// List 列表查询
func (s *Service[T]) List(ctx context.Context, opts ...QueryOption) ([]T, error) {
	optsKey := s.serializeOpts(opts...)
	cacheKey := s.cacheKey(fmt.Sprintf("list:%s", optsKey))

	var list []T
	if s.getFromCache(ctx, cacheKey, &list) {
		return list, nil
	}

	result, err := s.Repo.List(ctx, opts...)
	if err != nil {
		return nil, err
	}

	s.setToCache(ctx, cacheKey, result)
	return result, nil
}

// FindPage 分页查询
func (s *Service[T]) FindPage(ctx context.Context, opts ...QueryOption) (*PageResult[T], error) {
	optsKey := s.serializeOpts(opts...)
	cacheKey := s.cacheKey(fmt.Sprintf("page:%s", optsKey))

	var pageResult PageResult[T]
	if s.getFromCache(ctx, cacheKey, &pageResult) {
		return &pageResult, nil
	}

	result, err := s.Repo.FindPage(ctx, opts...)
	if err != nil {
		return nil, err
	}

	s.setToCache(ctx, cacheKey, result)
	return result, nil
}

// ==================== 创建操作（清除列表缓存）====================

func (s *Service[T]) Create(ctx context.Context, entity *T) error {
	err := s.Repo.Create(ctx, entity)
	if err != nil {
		return err
	}

	// 创建后使列表缓存失效
	s.invalidateListCache(ctx)
	return nil
}

func (s *Service[T]) CreateBatch(ctx context.Context, entities []T, batchSize ...int) error {
	err := s.Repo.CreateBatch(ctx, entities, batchSize...)
	if err != nil {
		return err
	}

	s.invalidateListCache(ctx)
	return nil
}

// ==================== 更新操作（清除相关缓存）====================

func (s *Service[T]) Update(ctx context.Context, entity *T) error {
	err := s.Repo.Update(ctx, entity)
	if err != nil {
		return err
	}

	// 获取ID并使缓存失效
	id := (*entity).GetID()
	s.invalidateIDCache(ctx, id)
	s.invalidateListCache(ctx)
	return nil
}

func (s *Service[T]) UpdateByID(ctx context.Context, id uint, updates map[string]interface{}) error {
	err := s.Repo.UpdateByID(ctx, id, updates)
	if err != nil {
		return err
	}

	s.invalidateIDCache(ctx, id)
	s.invalidateListCache(ctx)
	return nil
}

func (s *Service[T]) UpdateByCondition(ctx context.Context, condition map[string]interface{}, updates map[string]interface{}) error {
	err := s.Repo.UpdateByCondition(ctx, condition, updates)
	if err != nil {
		return err
	}

	// 批量更新，不确定影响哪些记录，清空所有缓存
	_ = s.ClearCache(ctx)
	return nil
}

// ==================== 删除操作（清除相关缓存）====================

func (s *Service[T]) Delete(ctx context.Context, entity *T) error {
	err := s.Repo.Delete(ctx, entity)
	if err != nil {
		return err
	}

	// 获取ID并使缓存失效
	id := (*entity).GetID()
	s.invalidateIDCache(ctx, id)
	s.invalidateListCache(ctx)
	return nil
}

func (s *Service[T]) DeleteByID(ctx context.Context, id uint) error {
	err := s.Repo.DeleteByID(ctx, id)
	if err != nil {
		return err
	}

	s.invalidateIDCache(ctx, id)
	s.invalidateListCache(ctx)
	return nil
}

func (s *Service[T]) DeleteByIDs(ctx context.Context, ids []uint) error {
	err := s.Repo.DeleteByIDs(ctx, ids)
	if err != nil {
		return err
	}

	for _, id := range ids {
		s.invalidateIDCache(ctx, id)
	}
	s.invalidateListCache(ctx)
	return nil
}

func (s *Service[T]) DeleteByCondition(ctx context.Context, condition map[string]interface{}) error {
	err := s.Repo.DeleteByCondition(ctx, condition)
	if err != nil {
		return err
	}

	// 批量删除，清空所有缓存
	_ = s.ClearCache(ctx)
	return nil
}

// ==================== 统计操作（不缓存）====================

func (s *Service[T]) Count(ctx context.Context, condition map[string]interface{}) (int64, error) {
	return s.Repo.Count(ctx, condition)
}

func (s *Service[T]) Exists(ctx context.Context, opts ...QueryOption) (bool, error) {
	return s.Repo.Exists(ctx, opts...)
}

func (s *Service[T]) ExistsByID(ctx context.Context, id uint) (bool, error) {
	return s.Repo.ExistsByID(ctx, id)
}

// ==================== 高级操作 ====================

func (s *Service[T]) FirstOrCreate(ctx context.Context, condition map[string]interface{}, entity *T) error {
	err := s.Repo.FirstOrCreate(ctx, condition, entity)
	if err != nil {
		return err
	}

	// 可能创建了新记录，清空列表缓存
	s.invalidateListCache(ctx)
	return nil
}

// ==================== 事务支持 ====================

func (s *Service[T]) Transaction(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB, txRepo IRepo[T]) error) error {
	// 事务中直接使用 repo，不走缓存
	err := s.Repo.Transaction(ctx, fn)
	if err != nil {
		return err
	}

	// 事务成功后清空所有缓存（保守策略）
	_ = s.ClearCache(ctx)
	return nil
}
