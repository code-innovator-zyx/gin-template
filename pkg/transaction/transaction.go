package transaction

import (
	"context"
	"gin-template/internal/core"
	"gorm.io/gorm"
)

// ExecuteInTransaction 在事务中执行函数
func ExecuteInTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	db := core.MustNewDbWithContext(ctx)
	return db.Transaction(fn)
}

// ExecuteInTransactionWithResult 在事务中执行函数并返回结果
func ExecuteInTransactionWithResult[T any](ctx context.Context, fn func(tx *gorm.DB) (T, error)) (T, error) {
	var result T
	db := core.MustNewDbWithContext(ctx)

	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		result, err = fn(tx)
		return err
	})

	return result, err
}

