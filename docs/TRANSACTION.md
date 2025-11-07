# 企业级事务封装使用指南

## 概述

本项目提供了一套企业级的事务管理解决方案，包含事务管理器、钩子函数、自动日志记录等功能，适用于复杂的业务场景。

## 核心特性

✅ **钩子函数** - 支持提交前/提交后/回滚后钩子  
✅ **自动日志** - 记录事务执行时间和状态  
✅ **缓存管理** - 事务提交后自动清理相关缓存  
✅ **隔离级别** - 支持自定义事务隔离级别  
✅ **流式API** - 链式调用，代码更优雅  
✅ **错误处理** - 完善的错误处理和回滚机制

---

## 快速开始

### 方式一：基础事务执行（简单场景）

```go
import (
    "gin-template/pkg/transaction"
    "context"
)

// 最简单的事务执行
err := transaction.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
    // 在事务中执行数据库操作
    if err := tx.Create(&user).Error; err != nil {
        return err // 返回错误会自动回滚
    }
    
    if err := tx.Create(&profile).Error; err != nil {
        return err
    }
    
    return nil // 返回 nil 会自动提交
})
```

### 方式二：企业级事务管理器（推荐）

```go
import (
    "gin-template/pkg/transaction"
    "context"
)

// 使用事务管理器
err := transaction.NewTransactionManager(ctx).
    WithOperationName("UpdateRole"). // 设置操作名称（用于日志）
    AfterCommit(func(ctx context.Context) error {
        // 事务提交后执行：清除缓存
        return service.GetCacheService().ClearUserPermissions(ctx, userID)
    }).
    Execute(func(tx *gorm.DB) error {
        // 执行业务逻辑
        if err := service.UpdateRoleWithTx(tx, role); err != nil {
            return err
        }
        return nil
    })
```

---

## 详细功能

### 1. 钩子函数

钩子函数允许你在事务的不同阶段执行额外的逻辑：

#### AfterCommit（提交后钩子）

**场景**：事务成功提交后执行，常用于清除缓存、发送通知等。

```go
err := transaction.NewTransactionManager(ctx).
    WithOperationName("CreateUser").
    AfterCommit(func(ctx context.Context) error {
        // 清除相关缓存
        return cacheService.ClearUserCache(ctx, userID)
    }).
    AfterCommit(func(ctx context.Context) error {
        // 发送欢迎邮件（多个钩子按顺序执行）
        return emailService.SendWelcome(ctx, user.Email)
    }).
    Execute(func(tx *gorm.DB) error {
        return tx.Create(&user).Error
    })
```

**注意**：
- AfterCommit 在事务提交后执行，**不会影响事务结果**
- 如果钩子执行失败，会记录错误日志，但不会导致事务回滚
- 适合用于非关键性的后续操作

#### BeforeCommit（提交前钩子）

**场景**：在事务提交前执行，仍在事务内，失败会导致回滚。

```go
err := transaction.NewTransactionManager(ctx).
    BeforeCommit(func(ctx context.Context) error {
        // 验证数据一致性
        return validateDataConsistency(ctx)
    }).
    Execute(func(tx *gorm.DB) error {
        return tx.Create(&order).Error
    })
```

**注意**：
- BeforeCommit 在事务内执行，**失败会导致回滚**
- 适合用于数据验证、一致性检查等关键操作

#### AfterRollback（回滚后钩子）

**场景**：事务回滚后执行，用于清理工作。

```go
err := transaction.NewTransactionManager(ctx).
    AfterRollback(func(ctx context.Context) error {
        // 清理临时文件
        return cleanupTempFiles(ctx)
    }).
    Execute(func(tx *gorm.DB) error {
        return tx.Create(&document).Error
    })
```

### 2. 自定义操作名称

为事务设置操作名称，方便日志追踪：

```go
err := transaction.NewTransactionManager(ctx).
    WithOperationName("UpdateUserProfile"). // 在日志中显示
    Execute(func(tx *gorm.DB) error {
        return tx.Updates(&user).Error
    })
```

**日志输出：**
```
[INFO] [Transaction] 开始事务: UpdateUserProfile
[INFO] [Transaction] 事务提交成功: UpdateUserProfile, 耗时: 15ms
```

### 3. 事务隔离级别

根据业务需求设置事务隔离级别：

```go
err := transaction.NewTransactionManager(ctx).
    WithIsolationLevel("REPEATABLE READ"). // 设置隔离级别
    Execute(func(tx *gorm.DB) error {
        // 业务逻辑
        return nil
    })
```

**支持的隔离级别：**
- `READ UNCOMMITTED` - 读未提交
- `READ COMMITTED` - 读已提交
- `REPEATABLE READ` - 可重复读（MySQL默认）
- `SERIALIZABLE` - 串行化

### 4. 禁用日志

在性能敏感场景下，可以禁用事务日志：

```go
err := transaction.NewTransactionManager(ctx).
    DisableLog(). // 禁用日志
    Execute(func(tx *gorm.DB) error {
        return tx.Create(&record).Error
    })
```

---

## 实战案例

### 案例 1：更新角色（完整示例）

**需求**：更新角色信息，同时更新资源绑定，并清除相关用户的权限缓存。

```go
func UpdateRole(c *gin.Context) {
    ctx := c.Request.Context()
    roleID := uint(123)
    
    // 获取要更新的数据
    role, err := service.GetRbacService().GetRoleByID(ctx, roleID)
    if err != nil {
        response.NotFound(c, "角色不存在")
        return
    }
    
    // 使用事务管理器
    err = transaction.NewTransactionManager(ctx).
        WithOperationName("UpdateRole").
        AfterCommit(func(ctx context.Context) error {
            // 事务成功后：清除所有拥有该角色的用户的权限缓存
            userIDs, err := service.GetRbacService().GetUsersWithRole(ctx, roleID)
            if err != nil {
                return err
            }
            
            if len(userIDs) > 0 {
                return service.GetCacheService().ClearMultipleUsersPermissions(ctx, userIDs)
            }
            return nil
        }).
        Execute(func(tx *gorm.DB) error {
            // 1. 更新角色基础信息
            if err := service.GetRbacService().UpdateRoleWithTx(tx, role); err != nil {
                return fmt.Errorf("更新角色基础信息失败: %w", err)
            }
            
            // 2. 更新角色资源绑定
            if err := service.GetRbacService().UpdateRoleResourcesWithTx(tx, roleID, resourceIDs); err != nil {
                return fmt.Errorf("更新角色资源失败: %w", err)
            }
            
            return nil
        })
    
    if err != nil {
        response.InternalServerError(c, err.Error())
        return
    }
    
    response.Success(c, role)
}
```

**执行流程：**
```
1. 开始事务
   ↓
2. 更新角色基础信息
   ↓
3. 更新角色资源绑定
   ↓
4. 提交事务
   ↓
5. 执行 AfterCommit 钩子：清除用户权限缓存
   ↓
6. 返回成功
```

### 案例 2：订单创建（多表操作）

**需求**：创建订单、扣减库存、创建支付记录，任何一步失败都回滚。

```go
func CreateOrder(ctx context.Context, orderReq *CreateOrderRequest) error {
    return transaction.NewTransactionManager(ctx).
        WithOperationName("CreateOrder").
        BeforeCommit(func(ctx context.Context) error {
            // 提交前验证库存
            return validateStock(ctx, orderReq.Items)
        }).
        AfterCommit(func(ctx context.Context) error {
            // 提交后发送通知
            return notifyService.SendOrderCreated(ctx, order.ID)
        }).
        AfterRollback(func(ctx context.Context) error {
            // 回滚后释放预占库存
            return releaseStock(ctx, orderReq.Items)
        }).
        Execute(func(tx *gorm.DB) error {
            // 1. 创建订单
            order := &Order{...}
            if err := tx.Create(order).Error; err != nil {
                return err
            }
            
            // 2. 扣减库存
            for _, item := range orderReq.Items {
                if err := tx.Model(&Product{}).
                    Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
                    Update("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
                    return fmt.Errorf("库存不足: %w", err)
                }
            }
            
            // 3. 创建支付记录
            payment := &Payment{OrderID: order.ID, ...}
            if err := tx.Create(payment).Error; err != nil {
                return err
            }
            
            return nil
        })
}
```

### 案例 3：批量操作

**需求**：批量导入用户，失败则全部回滚。

```go
func BatchImportUsers(ctx context.Context, users []User) error {
    return transaction.NewTransactionManager(ctx).
        WithOperationName("BatchImportUsers").
        AfterCommit(func(ctx context.Context) error {
            // 清除用户列表缓存
            return cacheService.ClearUserListCache(ctx)
        }).
        Execute(func(tx *gorm.DB) error {
            // 批量插入
            for i, user := range users {
                if err := tx.Create(&user).Error; err != nil {
                    return fmt.Errorf("导入第 %d 个用户失败: %w", i+1, err)
                }
            }
            return nil
        })
}
```

---

## Service 层改造

为了支持事务，Service 层方法需要提供两个版本：

### 模式：普通方法 + WithTx 方法

```go
// rbac_service.go

// UpdateRole 普通方法（自己管理事务）
func (s *rbacService) UpdateRole(ctx context.Context, role *Role) error {
    return core.MustNewDbWithContext(ctx).Save(role).Error
}

// UpdateRoleWithTx 事务方法（接受外部事务）
func (s *rbacService) UpdateRoleWithTx(tx *gorm.DB, role *Role) error {
    return tx.Save(role).Error
}
```

**命名规范：**
- 普通方法：`MethodName(ctx, ...)`
- 事务方法：`MethodNameWithTx(tx, ...)`

**使用场景：**
```go
// 场景1：单独调用（自动事务）
err := service.UpdateRole(ctx, role)

// 场景2：在事务中调用
transaction.Execute(func(tx *gorm.DB) error {
    return service.UpdateRoleWithTx(tx, role)
})
```

---

## 最佳实践

### 1. 事务粒度

✅ **推荐**：事务尽可能小，只包含必要的数据库操作

```go
// 好：事务只包含数据库操作
transaction.Execute(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return err
    }
    return tx.Create(&profile).Error
})
```

❌ **不推荐**：事务中包含耗时操作

```go
// 差：事务中包含HTTP调用
transaction.Execute(func(tx *gorm.DB) error {
    tx.Create(&user)
    http.Post("http://external-api.com", ...) // ❌ 耗时操作
    return nil
})
```

### 2. 钩子函数使用

✅ **AfterCommit**：用于非关键操作（缓存清理、通知发送）  
✅ **BeforeCommit**：用于关键验证（数据一致性检查）  
✅ **AfterRollback**：用于清理工作（临时文件删除）

### 3. 错误处理

```go
// 推荐：包装错误信息，方便定位问题
transaction.Execute(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return fmt.Errorf("创建用户失败: %w", err) // 包装错误
    }
    return nil
})
```

### 4. 性能优化

```go
// 批量操作使用 CreateInBatches
transaction.Execute(func(tx *gorm.DB) error {
    // 每次插入 100 条
    return tx.CreateInBatches(users, 100).Error
})
```

### 5. 缓存清理策略

```go
// 推荐：在 AfterCommit 中清理缓存
transaction.NewTransactionManager(ctx).
    AfterCommit(func(ctx context.Context) error {
        // 只清理受影响的缓存
        return cacheService.ClearUserPermissions(ctx, userID)
    }).
    Execute(...)
```

---

## 错误处理

### 事务内错误

```go
err := transaction.Execute(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        // 返回错误会自动回滚
        return err
    }
    return nil
})

if err != nil {
    // 处理错误
    log.Errorf("事务执行失败: %v", err)
}
```

### 钩子函数错误

```go
// BeforeCommit 错误会导致回滚
transaction.NewTransactionManager(ctx).
    BeforeCommit(func(ctx context.Context) error {
        if !isValid() {
            return errors.New("验证失败") // 会导致回滚
        }
        return nil
    }).
    Execute(...)

// AfterCommit 错误不会影响事务
transaction.NewTransactionManager(ctx).
    AfterCommit(func(ctx context.Context) error {
        // 错误会记录日志，但不会导致事务回滚
        return cacheService.Clear(ctx)
    }).
    Execute(...)
```

---

## 常见问题

### Q1: 什么时候使用事务管理器？

**A:** 以下场景建议使用：
- 多表操作需要保证一致性
- 需要在事务后清理缓存
- 复杂业务逻辑需要日志追踪
- 需要自定义事务隔离级别

### Q2: AfterCommit 和 BeforeCommit 的区别？

**A:** 
- **BeforeCommit**：在事务内执行，失败会回滚，用于关键验证
- **AfterCommit**：事务外执行，失败不会回滚，用于非关键操作

### Q3: 事务嵌套怎么办？

**A:** GORM 使用 `SavePoint` 机制支持事务嵌套，但建议尽量避免。

### Q4: 性能影响？

**A:** 事务管理器开销很小（<1ms），主要开销在于数据库事务本身。

### Q5: 如何调试事务？

**A:** 
1. 查看事务日志（包含执行时间和状态）
2. 使用 `WithOperationName` 标识事务
3. 在钩子函数中添加日志

---

## API 参考

### TransactionManager

| 方法 | 说明 | 返回值 |
|------|------|--------|
| `NewTransactionManager(ctx)` | 创建事务管理器 | `*TransactionManager` |
| `WithOperationName(name)` | 设置操作名称 | `*TransactionManager` |
| `WithIsolationLevel(level)` | 设置隔离级别 | `*TransactionManager` |
| `DisableLog()` | 禁用日志 | `*TransactionManager` |
| `BeforeCommit(fn)` | 注册提交前钩子 | `*TransactionManager` |
| `AfterCommit(fn)` | 注册提交后钩子 | `*TransactionManager` |
| `AfterRollback(fn)` | 注册回滚后钩子 | `*TransactionManager` |
| `Execute(fn)` | 执行事务 | `error` |
| `ExecuteWithResult(fn)` | 执行事务并返回结果 | `(interface{}, error)` |

### 快捷函数

| 函数 | 说明 |
|------|------|
| `ExecuteInTransaction(ctx, fn)` | 简单事务执行 |
| `ExecuteInTransactionWithResult[T](ctx, fn)` | 带结果的事务执行 |
| `WithTransaction(ctx, name, fn)` | 快捷创建并执行 |
| `WithTransactionAndHooks(ctx, name, fn, afterCommit)` | 带钩子的快捷执行 |

---

## 总结

本事务封装方案的核心优势：

1. **简单易用** - 链式 API，代码优雅
2. **功能完善** - 钩子函数、日志、隔离级别
3. **性能优秀** - 开销极小，支持批量操作
4. **企业级** - 完善的错误处理和缓存管理
5. **可扩展** - 易于添加新功能和钩子

开始使用时，记住：
> **简单场景用 `ExecuteInTransaction`，复杂场景用 `TransactionManager`！**

---

如有问题或建议，欢迎提 Issue！

