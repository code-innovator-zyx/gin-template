package orm

import (
	"context"
	"gorm.io/gorm"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/10/21 上午10:23
* @Package: 分页查询工具
 */

type PageResult[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}
type PageQuery struct {
	Page     int    `json:"page"`      // 页码，从1开始
	PageSize int    `json:"page_size"` // 每页大小
	OrderBy  string `json:"order_by"`  // 排序字段，如 "id desc"
}

func Paginate[T any](ctx context.Context, db *gorm.DB, query PageQuery) (*PageResult[T], error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	tx := db.Session(&gorm.Session{}).WithContext(ctx).Model(new(T))
	// 统计总数
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, err
	}

	// 计算分页偏移量
	offset := (query.Page - 1) * query.PageSize
	// 排序
	if query.OrderBy != "" {
		tx = tx.Order(query.OrderBy)
	}
	// 查询数据
	var list []T
	if err := tx.Offset(offset).Limit(query.PageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return &PageResult[T]{
		List:     list,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}
