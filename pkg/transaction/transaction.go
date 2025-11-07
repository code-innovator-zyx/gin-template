package transaction

import (
	"context"
	"fmt"
	"gin-template/internal/core"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ================ 基础事务执行器 ================

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

// ================ 事务管理器 ================

// HookFunc 钩子函数类型
type HookFunc func(ctx context.Context) error

// TransactionManager 事务管理器
type TransactionManager struct {
	ctx            context.Context
	db             *gorm.DB
	beforeCommit   []HookFunc // 提交前钩子（在事务内执行）
	afterCommit    []HookFunc // 提交后钩子（事务成功后执行）
	afterRollback  []HookFunc // 回滚后钩子
	enableLog      bool       // 是否启用日志
	startTime      time.Time
	operationName  string // 操作名称（用于日志）
	isolationLevel string // 事务隔离级别
}

// NewTransactionManager 创建事务管理器
func NewTransactionManager(ctx context.Context) *TransactionManager {
	return &TransactionManager{
		ctx:            ctx,
		db:             core.MustNewDbWithContext(ctx),
		beforeCommit:   make([]HookFunc, 0),
		afterCommit:    make([]HookFunc, 0),
		afterRollback:  make([]HookFunc, 0),
		enableLog:      true,
		operationName:  "Transaction",
		isolationLevel: "", // 默认不设置，使用数据库默认隔离级别
	}
}

// WithOperationName 设置操作名称（用于日志追踪）
func (tm *TransactionManager) WithOperationName(name string) *TransactionManager {
	tm.operationName = name
	return tm
}

// WithIsolationLevel 设置事务隔离级别
// 可选值: "READ UNCOMMITTED", "READ COMMITTED", "REPEATABLE READ", "SERIALIZABLE"
func (tm *TransactionManager) WithIsolationLevel(level string) *TransactionManager {
	tm.isolationLevel = level
	return tm
}

// DisableLog 禁用日志
func (tm *TransactionManager) DisableLog() *TransactionManager {
	tm.enableLog = false
	return tm
}

// BeforeCommit 注册提交前钩子（在事务内执行，失败会导致回滚）
func (tm *TransactionManager) BeforeCommit(fn HookFunc) *TransactionManager {
	tm.beforeCommit = append(tm.beforeCommit, fn)
	return tm
}

// AfterCommit 注册提交后钩子（事务成功后执行）
func (tm *TransactionManager) AfterCommit(fn HookFunc) *TransactionManager {
	tm.afterCommit = append(tm.afterCommit, fn)
	return tm
}

// AfterRollback 注册回滚后钩子（事务回滚后执行，用于清理工作）
func (tm *TransactionManager) AfterRollback(fn HookFunc) *TransactionManager {
	tm.afterRollback = append(tm.afterRollback, fn)
	return tm
}

// Execute 执行事务
func (tm *TransactionManager) Execute(fn func(tx *gorm.DB) error) error {
	tm.startTime = time.Now()

	if tm.enableLog {
		logrus.Infof("[Transaction] 开始事务: %s", tm.operationName)
	}

	var txErr error
	err := tm.db.Transaction(func(tx *gorm.DB) error {
		// 设置事务隔离级别
		if tm.isolationLevel != "" {
			if err := tx.Exec(fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", tm.isolationLevel)).Error; err != nil {
				return fmt.Errorf("设置事务隔离级别失败: %w", err)
			}
		}

		// 执行主业务逻辑
		if err := fn(tx); err != nil {
			txErr = err
			return err
		}

		// 执行提交前钩子
		for i, hook := range tm.beforeCommit {
			if err := hook(tm.ctx); err != nil {
				txErr = fmt.Errorf("提交前钩子 #%d 执行失败: %w", i+1, err)
				return txErr
			}
		}

		return nil
	})

	duration := time.Since(tm.startTime)

	// 事务执行失败
	if err != nil {
		if tm.enableLog {
			logrus.Errorf("[Transaction] 事务回滚: %s, 耗时: %v, 错误: %v",
				tm.operationName, duration, err)
		}

		// 执行回滚后钩子
		for i, hook := range tm.afterRollback {
			if hookErr := hook(tm.ctx); hookErr != nil {
				logrus.Errorf("[Transaction] 回滚后钩子 #%d 执行失败: %v", i+1, hookErr)
			}
		}

		return err
	}

	// 事务执行成功
	if tm.enableLog {
		logrus.Infof("[Transaction] 事务提交成功: %s, 耗时: %v", tm.operationName, duration)
	}

	// 执行提交后钩子（异步执行，不影响事务结果）
	for i, hook := range tm.afterCommit {
		go func(i int, hook HookFunc) {
			if err := hook(tm.ctx); err != nil {
				logrus.Errorf("[Transaction] 提交后钩子 #%d 执行失败: %v", i+1, err)
			}
		}(i, hook)
	}

	return nil
}

// ExecuteWithResult 执行事务并返回结果
func (tm *TransactionManager) ExecuteWithResult(fn func(tx *gorm.DB) (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := tm.Execute(func(tx *gorm.DB) error {
		var err error
		result, err = fn(tx)
		return err
	})
	return result, err
}

// ================ 快捷方法 ================

// WithTransaction 快捷创建事务管理器并执行
func WithTransaction(ctx context.Context, operationName string, fn func(tx *gorm.DB) error) error {
	return NewTransactionManager(ctx).
		WithOperationName(operationName).
		Execute(fn)
}

// WithTransactionAndHooks 带钩子的事务执行
func WithTransactionAndHooks(ctx context.Context, operationName string,
	fn func(tx *gorm.DB) error,
	afterCommit HookFunc) error {
	tm := NewTransactionManager(ctx).
		WithOperationName(operationName)

	if afterCommit != nil {
		tm.AfterCommit(afterCommit)
	}

	return tm.Execute(fn)
}
