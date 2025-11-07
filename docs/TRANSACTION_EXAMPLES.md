# 事务管理器使用示例

本文档提供企业级事务管理器的实战代码示例。

## 目录

- [基础示例](#基础示例)
- [中级示例](#中级示例)
- [高级示例](#高级示例)
- [实战案例](#实战案例)

---

## 基础示例

### 示例 1：最简单的事务

```go
import (
    "gin-template/pkg/transaction"
    "context"
)

func CreateUser(ctx context.Context, user *User) error {
    // 最简单的事务执行
    return transaction.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
        return tx.Create(user).Error
    })
}
```

### 示例 2：多个操作

```go
func CreateUserWithProfile(ctx context.Context, user *User, profile *Profile) error {
    return transaction.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
        // 创建用户
        if err := tx.Create(user).Error; err != nil {
            return fmt.Errorf("创建用户失败: %w", err)
        }
        
        // 设置 profile 的 user_id
        profile.UserID = user.ID
        
        // 创建用户资料
        if err := tx.Create(profile).Error; err != nil {
            return fmt.Errorf("创建用户资料失败: %w", err)
        }
        
        return nil
    })
}
```

### 示例 3：带返回值的事务

```go
func CreateUserAndGetID(ctx context.Context, user *User) (uint, error) {
    return transaction.ExecuteInTransactionWithResult[uint](ctx, func(tx *gorm.DB) (uint, error) {
        if err := tx.Create(user).Error; err != nil {
            return 0, err
        }
        return user.ID, nil
    })
}
```

---

## 中级示例

### 示例 4：事务 + 缓存清理

```go
func UpdateUserStatus(ctx context.Context, userID uint, status int) error {
    return transaction.NewTransactionManager(ctx).
        WithOperationName("UpdateUserStatus").
        AfterCommit(func(ctx context.Context) error {
            // 事务成功后清除用户缓存
            return cacheService.ClearUserCache(ctx, userID)
        }).
        Execute(func(tx *gorm.DB) error {
            return tx.Model(&User{}).
                Where("id = ?", userID).
                Update("status", status).Error
        })
}
```

### 示例 5：事务 + 验证

```go
func TransferMoney(ctx context.Context, fromUserID, toUserID uint, amount decimal.Decimal) error {
    return transaction.NewTransactionManager(ctx).
        WithOperationName("TransferMoney").
        BeforeCommit(func(ctx context.Context) error {
            // 提交前验证余额
            var balance decimal.Decimal
            if err := db.Model(&Account{}).
                Where("user_id = ?", fromUserID).
                Pluck("balance", &balance).Error; err != nil {
                return err
            }
            
            if balance.LessThan(amount) {
                return errors.New("余额不足")
            }
            return nil
        }).
        Execute(func(tx *gorm.DB) error {
            // 扣款
            if err := tx.Model(&Account{}).
                Where("user_id = ?", fromUserID).
                Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
                return err
            }
            
            // 入账
            if err := tx.Model(&Account{}).
                Where("user_id = ?", toUserID).
                Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
                return err
            }
            
            return nil
        })
}
```

### 示例 6：批量操作

```go
func BatchCreateProducts(ctx context.Context, products []Product) error {
    return transaction.NewTransactionManager(ctx).
        WithOperationName("BatchCreateProducts").
        AfterCommit(func(ctx context.Context) error {
            // 清除产品列表缓存
            return cacheService.ClearProductListCache(ctx)
        }).
        Execute(func(tx *gorm.DB) error {
            // 每次插入 100 条
            return tx.CreateInBatches(products, 100).Error
        })
}
```

---

## 高级示例

### 示例 7：完整的更新角色（生产环境）

```go
func UpdateRole(ctx context.Context, roleID uint, updateReq UpdateRoleRequest) (*Role, error) {
    // 事务前检查
    role, err := service.GetRoleByID(ctx, roleID)
    if err != nil {
        return nil, fmt.Errorf("角色不存在: %w", err)
    }
    
    // 检查名称重复
    if updateReq.Name != "" && updateReq.Name != role.Name {
        exists, err := service.CheckRoleNameExists(ctx, updateReq.Name, roleID)
        if err != nil {
            return nil, err
        }
        if exists {
            return nil, errors.New("角色名称已存在")
        }
        role.Name = updateReq.Name
    }
    
    if updateReq.Description != "" {
        role.Description = updateReq.Description
    }
    
    // 使用事务管理器
    err = transaction.NewTransactionManager(ctx).
        WithOperationName("UpdateRole").
        AfterCommit(func(ctx context.Context) error {
            // 获取拥有该角色的所有用户
            userIDs, err := service.GetUsersWithRole(ctx, roleID)
            if err != nil {
                logrus.Errorf("获取角色用户失败: %v", err)
                return err
            }
            
            // 批量清除权限缓存
            if len(userIDs) > 0 {
                if err := cacheService.ClearMultipleUsersPermissions(ctx, userIDs); err != nil {
                    logrus.Errorf("清除权限缓存失败: %v", err)
                    return err
                }
                logrus.Infof("已清除 %d 个用户的权限缓存", len(userIDs))
            }
            
            return nil
        }).
        Execute(func(tx *gorm.DB) error {
            // 1. 更新角色基础信息
            if err := service.UpdateRoleWithTx(tx, role); err != nil {
                return fmt.Errorf("更新角色信息失败: %w", err)
            }
            
            // 2. 更新资源绑定
            if len(updateReq.ResourceIDs) > 0 {
                if err := service.UpdateRoleResourcesWithTx(tx, roleID, updateReq.ResourceIDs); err != nil {
                    return fmt.Errorf("更新角色资源失败: %w", err)
                }
            }
            
            return nil
        })
    
    if err != nil {
        return nil, err
    }
    
    // 返回更新后的角色
    return service.GetRoleByID(ctx, roleID)
}
```

### 示例 8：订单创建（复杂业务）

```go
func CreateOrder(ctx context.Context, orderReq CreateOrderRequest) (*Order, error) {
    var order *Order
    
    err := transaction.NewTransactionManager(ctx).
        WithOperationName("CreateOrder").
        WithIsolationLevel("REPEATABLE READ"). // 使用可重复读隔离级别
        BeforeCommit(func(ctx context.Context) error {
            // 提交前再次验证库存（防止超卖）
            for _, item := range orderReq.Items {
                var stock int
                if err := db.Model(&Product{}).
                    Where("id = ?", item.ProductID).
                    Pluck("stock", &stock).Error; err != nil {
                    return err
                }
                
                if stock < item.Quantity {
                    return fmt.Errorf("商品 %d 库存不足", item.ProductID)
                }
            }
            return nil
        }).
        AfterCommit(func(ctx context.Context) error {
            // 发送订单创建通知
            if err := notifyService.SendOrderCreated(ctx, order.ID); err != nil {
                logrus.Errorf("发送订单通知失败: %v", err)
                // 不返回错误，避免影响事务
            }
            
            // 清除用户订单缓存
            return cacheService.ClearUserOrderCache(ctx, order.UserID)
        }).
        AfterRollback(func(ctx context.Context) error {
            // 如果创建失败，释放预占库存（如果有）
            logrus.Warnf("订单创建失败，执行清理工作")
            return nil
        }).
        Execute(func(tx *gorm.DB) error {
            // 1. 创建订单
            order = &Order{
                UserID:      orderReq.UserID,
                TotalAmount: orderReq.TotalAmount,
                Status:      OrderStatusPending,
            }
            if err := tx.Create(order).Error; err != nil {
                return fmt.Errorf("创建订单失败: %w", err)
            }
            
            // 2. 创建订单明细
            for _, item := range orderReq.Items {
                orderItem := &OrderItem{
                    OrderID:   order.ID,
                    ProductID: item.ProductID,
                    Quantity:  item.Quantity,
                    Price:     item.Price,
                }
                if err := tx.Create(orderItem).Error; err != nil {
                    return fmt.Errorf("创建订单明细失败: %w", err)
                }
            }
            
            // 3. 扣减库存
            for _, item := range orderReq.Items {
                result := tx.Model(&Product{}).
                    Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
                    Update("stock", gorm.Expr("stock - ?", item.Quantity))
                
                if result.Error != nil {
                    return fmt.Errorf("扣减库存失败: %w", result.Error)
                }
                
                if result.RowsAffected == 0 {
                    return fmt.Errorf("商品 %d 库存不足", item.ProductID)
                }
            }
            
            // 4. 创建支付记录
            payment := &Payment{
                OrderID: order.ID,
                Amount:  orderReq.TotalAmount,
                Status:  PaymentStatusPending,
            }
            if err := tx.Create(payment).Error; err != nil {
                return fmt.Errorf("创建支付记录失败: %w", err)
            }
            
            return nil
        })
    
    if err != nil {
        return nil, err
    }
    
    return order, nil
}
```

### 示例 9：数据迁移（大批量）

```go
func MigrateUserData(ctx context.Context) error {
    const batchSize = 1000
    var offset int
    
    for {
        // 分批处理，每批一个事务
        err := transaction.NewTransactionManager(ctx).
            WithOperationName(fmt.Sprintf("MigrateUserData_Offset_%d", offset)).
            Execute(func(tx *gorm.DB) error {
                var users []OldUser
                
                // 查询一批数据
                if err := tx.Offset(offset).Limit(batchSize).Find(&users).Error; err != nil {
                    return err
                }
                
                // 没有更多数据
                if len(users) == 0 {
                    return nil
                }
                
                // 转换并插入新表
                newUsers := make([]NewUser, 0, len(users))
                for _, oldUser := range users {
                    newUsers = append(newUsers, ConvertUser(oldUser))
                }
                
                if err := tx.CreateInBatches(newUsers, 100).Error; err != nil {
                    return err
                }
                
                logrus.Infof("迁移了 %d 条用户数据", len(users))
                return nil
            })
        
        if err != nil {
            return err
        }
        
        offset += batchSize
        
        // 检查是否完成
        var count int64
        if err := db.Model(&OldUser{}).Offset(offset).Limit(1).Count(&count).Error; err != nil {
            return err
        }
        if count == 0 {
            break
        }
    }
    
    logrus.Info("用户数据迁移完成")
    return nil
}
```

---

## 实战案例

### 案例 1：用户注册

```go
type RegisterRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Email    string `json:"email"`
}

func RegisterUser(ctx context.Context, req RegisterRequest) (*User, error) {
    var user *User
    
    err := transaction.NewTransactionManager(ctx).
        WithOperationName("RegisterUser").
        AfterCommit(func(ctx context.Context) error {
            // 发送欢迎邮件
            go emailService.SendWelcome(user.Email) // 异步发送
            return nil
        }).
        Execute(func(tx *gorm.DB) error {
            // 1. 创建用户
            user = &User{
                Username: req.Username,
                Password: hashPassword(req.Password),
                Email:    req.Email,
                Status:   UserStatusActive,
            }
            if err := tx.Create(user).Error; err != nil {
                return err
            }
            
            // 2. 创建用户资料
            profile := &UserProfile{
                UserID: user.ID,
            }
            if err := tx.Create(profile).Error; err != nil {
                return err
            }
            
            // 3. 分配默认角色
            userRole := &UserRole{
                UserID: user.ID,
                RoleID: DefaultRoleID,
            }
            if err := tx.Create(userRole).Error; err != nil {
                return err
            }
            
            return nil
        })
    
    if err != nil {
        return nil, err
    }
    
    return user, nil
}
```

### 案例 2：库存管理

```go
func AdjustStock(ctx context.Context, productID uint, quantity int, reason string) error {
    return transaction.NewTransactionManager(ctx).
        WithOperationName("AdjustStock").
        BeforeCommit(func(ctx context.Context) error {
            // 验证调整后库存不为负
            var currentStock int
            if err := db.Model(&Product{}).
                Where("id = ?", productID).
                Pluck("stock", &currentStock).Error; err != nil {
                return err
            }
            
            if currentStock+quantity < 0 {
                return errors.New("调整后库存不能为负数")
            }
            return nil
        }).
        AfterCommit(func(ctx context.Context) error {
            // 清除产品缓存
            return cacheService.ClearProductCache(ctx, productID)
        }).
        Execute(func(tx *gorm.DB) error {
            // 1. 更新库存
            if err := tx.Model(&Product{}).
                Where("id = ?", productID).
                Update("stock", gorm.Expr("stock + ?", quantity)).Error; err != nil {
                return err
            }
            
            // 2. 记录库存变动日志
            log := &StockLog{
                ProductID: productID,
                Quantity:  quantity,
                Reason:    reason,
            }
            if err := tx.Create(log).Error; err != nil {
                return err
            }
            
            return nil
        })
}
```

### 案例 3：删除用户（级联删除）

```go
func DeleteUser(ctx context.Context, userID uint) error {
    return transaction.NewTransactionManager(ctx).
        WithOperationName("DeleteUser").
        AfterCommit(func(ctx context.Context) error {
            // 清除所有相关缓存
            cacheService.ClearUserCache(ctx, userID)
            cacheService.ClearUserPermissions(ctx, userID)
            return nil
        }).
        Execute(func(tx *gorm.DB) error {
            // 1. 删除用户资料
            if err := tx.Where("user_id = ?", userID).Delete(&UserProfile{}).Error; err != nil {
                return err
            }
            
            // 2. 删除用户角色关联
            if err := tx.Where("user_id = ?", userID).Delete(&UserRole{}).Error; err != nil {
                return err
            }
            
            // 3. 删除用户订单（软删除）
            if err := tx.Where("user_id = ?", userID).Delete(&Order{}).Error; err != nil {
                return err
            }
            
            // 4. 删除用户
            if err := tx.Delete(&User{}, userID).Error; err != nil {
                return err
            }
            
            return nil
        })
}
```

---

## 性能优化技巧

### 技巧 1：批量操作

```go
// ❌ 不推荐：逐条插入
for _, user := range users {
    tx.Create(&user)
}

// ✅ 推荐：批量插入
tx.CreateInBatches(users, 100) // 每次插入 100 条
```

### 技巧 2：预加载关联

```go
transaction.Execute(func(tx *gorm.DB) error {
    // 预加载关联数据
    return tx.Preload("Profile").Preload("Roles").Find(&users).Error
})
```

### 技巧 3：减少事务范围

```go
// ❌ 不推荐：事务范围太大
transaction.Execute(func(tx *gorm.DB) error {
    // 耗时的计算
    result := expensiveCalculation()
    
    // 数据库操作
    return tx.Create(&record).Error
})

// ✅ 推荐：事务只包含数据库操作
result := expensiveCalculation() // 事务外执行
transaction.Execute(func(tx *gorm.DB) error {
    return tx.Create(&record).Error
})
```

---

## 总结

本文档提供了从基础到高级的完整示例，涵盖：

- ✅ 基础事务执行
- ✅ 钩子函数使用
- ✅ 缓存管理
- ✅ 复杂业务场景
- ✅ 性能优化

**最佳实践**：
1. 简单场景用 `ExecuteInTransaction`
2. 复杂场景用 `TransactionManager`
3. 缓存清理放在 `AfterCommit`
4. 关键验证放在 `BeforeCommit`
5. 事务尽可能小，只包含数据库操作

开始使用企业级事务管理，让你的代码更健壮！🚀

