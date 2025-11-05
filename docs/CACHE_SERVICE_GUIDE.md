# 🚀 缓存服务使用指南

## 📖 概述

`cache_service.go` 提供了统一的缓存管理接口，封装了所有与Redis相关的操作，使用简单且功能强大。

### 核心优势

✅ **统一管理** - 所有缓存操作集中在一个服务中  
✅ **优雅降级** - Redis不可用时自动降级到数据库查询  
✅ **类型安全** - 使用接口定义，类型检查严格  
✅ **易于测试** - 接口设计方便Mock测试  
✅ **高性能** - 合理的TTL设置，减少数据库压力  

---

## 🎯 快速开始

### 1. 创建缓存服务实例

```go
import "gin-template/internal/service"

// 创建缓存服务
cacheService := service.MustNewCacheService()
```

### 2. 基础使用示例

```go
ctx := context.Background()

// 设置缓存
err := cacheService.Set(ctx, "key", "value", 10*time.Minute)

// 获取缓存
var value string
err := cacheService.Get(ctx, "key", &value)

// 删除缓存
err := cacheService.Delete(ctx, "key")
```

---

## 📚 功能详解

### 1️⃣ 权限相关缓存

#### CheckUserPermission - 检查用户权限

```go
// 自动处理缓存逻辑，缓存命中时仅需2ms
hasPermission, err := cacheService.CheckUserPermission(
    ctx, 
    userID,    // 用户ID
    "/api/v1/users",  // 路径
    "GET",     // 方法
)

if hasPermission {
    // 有权限
}
```

**特点**：
- ✅ 自动缓存权限检查结果
- ✅ 缓存未命中时查询数据库
- ✅ TTL: 10分钟

#### ClearUserPermissions - 清除用户权限缓存

```go
// 当用户角色变更时调用
err := cacheService.ClearUserPermissions(ctx, userID)
```

**使用场景**：
- 用户角色发生变更
- 用户权限被撤销
- 强制刷新用户权限

#### ClearAllPermissions - 清除所有权限缓存

```go
// 当权限规则发生变更时调用
err := cacheService.ClearAllPermissions(ctx)
```

---

### 2️⃣ 用户相关缓存

#### SetUser / GetUser - 用户信息缓存

```go
// 设置用户信息缓存
user := &rbac.User{
    ID:       1,
    Username: "admin",
    Email:    "admin@example.com",
}
err := cacheService.SetUser(ctx, user, 30*time.Minute)

// 获取用户信息
cachedUser, err := cacheService.GetUser(ctx, userID)
if err == redis.Nil {
    // 缓存未命中，查询数据库
    user, err = rbac.GetUserByID(userID)
    // 写入缓存
    cacheService.SetUser(ctx, user, 0) // 0表示使用默认TTL
}
```

**完整示例（推荐模式）**：

```go
func GetUserInfo(ctx context.Context, userID uint) (*rbac.User, error) {
    // 1. 尝试从缓存获取
    user, err := cacheService.GetUser(ctx, userID)
    if err == nil {
        return user, nil
    }
    
    // 2. 缓存未命中，查询数据库
    user, err = rbac.GetUserByID(userID)
    if err != nil {
        return nil, err
    }
    
    // 3. 写入缓存（忽略错误）
    _ = cacheService.SetUser(ctx, user, 0)
    
    return user, nil
}
```

#### DeleteUser - 删除用户缓存

```go
// 当用户信息更新时清除缓存
err := cacheService.DeleteUser(ctx, userID)
```

---

### 3️⃣ 角色相关缓存

#### GetUserRoles / SetUserRoles - 用户角色缓存

```go
// 获取用户角色（带缓存）
func GetUserRolesWithCache(ctx context.Context, userID uint) ([]rbac.Role, error) {
    // 尝试从缓存获取
    roles, err := cacheService.GetUserRoles(ctx, userID)
    if err == nil {
        return roles, nil
    }
    
    // 查询数据库
    roles, err = service.GetUserRoles(userID)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存
    _ = cacheService.SetUserRoles(ctx, userID, roles, 0)
    
    return roles, nil
}
```

#### ClearUserRoles - 清除角色缓存

```go
// 当用户角色变更时调用
err := cacheService.ClearUserRoles(ctx, userID)
```

---

### 4️⃣ 通用缓存操作

#### Get / Set / Delete

```go
// 设置缓存
data := map[string]interface{}{
    "name": "John",
    "age":  30,
}
err := cacheService.Set(ctx, "user:profile:1", data, 1*time.Hour)

// 获取缓存
var profile map[string]interface{}
err := cacheService.Get(ctx, "user:profile:1", &profile)

// 删除缓存
err := cacheService.Delete(ctx, "user:profile:1")

// 批量删除
err := cacheService.Delete(ctx, "key1", "key2", "key3")
```

#### Exists - 检查key是否存在

```go
exists, err := cacheService.Exists(ctx, "user:profile:1")
if exists {
    // key存在
}
```

---

### 5️⃣ 高级功能

#### SetWithCallback - 缓存未命中时自动获取

```go
// 自动处理缓存逻辑
result, err := cacheService.SetWithCallback(
    ctx,
    "user:stats:1",
    10*time.Minute,
    func() (interface{}, error) {
        // 这个函数只在缓存未命中时执行
        return calculateUserStats(userID)
    },
)
```

**优点**：
- 简化代码逻辑
- 自动处理缓存读写
- 减少重复代码

#### DeleteByPattern - 批量删除

```go
// 删除所有用户相关缓存
err := cacheService.DeleteByPattern(ctx, "user:*")

// 删除特定用户的所有缓存
err := cacheService.DeleteByPattern(ctx, "user:1:*")
```

#### RefreshTTL - 刷新过期时间

```go
// 延长缓存有效期
err := cacheService.RefreshTTL(ctx, "session:1", 1*time.Hour)
```

#### GetTTL - 获取剩余时间

```go
// 获取key的剩余有效时间
ttl, err := cacheService.GetTTL(ctx, "session:1")
fmt.Printf("剩余时间: %v\n", ttl)
```

---

### 6️⃣ Token黑名单

#### BlacklistToken / IsTokenBlacklisted

```go
// 用户登出时，将token加入黑名单
err := cacheService.BlacklistToken(ctx, tokenString, 24*time.Hour)

// 验证token时检查黑名单
isBlacklisted, err := cacheService.IsTokenBlacklisted(ctx, tokenString)
if isBlacklisted {
    return errors.New("token已失效")
}
```

**使用场景**：
- 用户主动登出
- 强制用户下线
- Token被盗用后撤销

---

### 7️⃣ 会话管理

#### SetSession / GetSession / DeleteSession

```go
// 设置会话数据
sessionData := map[string]interface{}{
    "login_time": time.Now(),
    "ip":         "192.168.1.1",
    "device":     "iPhone",
}
err := cacheService.SetSession(ctx, userID, sessionData, 1*time.Hour)

// 获取会话
session, err := cacheService.GetSession(ctx, userID)

// 删除会话（用户登出）
err := cacheService.DeleteSession(ctx, userID)
```

---

## 🎯 实战案例

### 案例1：完整的用户登录流程

```go
func Login(ctx context.Context, username, password string) (string, error) {
    // 1. 验证用户名密码
    user, err := rbac.GetUserByUsername(username)
    if err != nil {
        return "", errors.New("用户不存在")
    }
    
    if !user.CheckPassword(password) {
        return "", errors.New("密码错误")
    }
    
    // 2. 生成token
    token, err := utils.GenerateToken(user.ID, user.Username)
    if err != nil {
        return "", err
    }
    
    // 3. 缓存用户信息
    _ = cacheService.SetUser(ctx, user, 30*time.Minute)
    
    // 4. 缓存用户角色
    roles, _ := service.GetUserRoles(user.ID)
    _ = cacheService.SetUserRoles(ctx, user.ID, roles, 30*time.Minute)
    
    // 5. 设置会话
    sessionData := map[string]interface{}{
        "login_time": time.Now(),
        "token":      token,
    }
    _ = cacheService.SetSession(ctx, user.ID, sessionData, 24*time.Hour)
    
    return token, nil
}
```

### 案例2：用户登出

```go
func Logout(ctx context.Context, userID uint, token string) error {
    // 1. token加入黑名单
    err := cacheService.BlacklistToken(ctx, token, 24*time.Hour)
    if err != nil {
        return err
    }
    
    // 2. 清除会话
    _ = cacheService.DeleteSession(ctx, userID)
    
    // 3. 清除用户缓存
    _ = cacheService.DeleteUser(ctx, userID)
    
    // 4. 清除权限缓存
    _ = cacheService.ClearUserPermissions(ctx, userID)
    
    return nil
}
```

### 案例3：角色变更

```go
func AssignRoleToUser(ctx context.Context, userID, roleID uint) error {
    // 1. 数据库操作
    err := service.AssignRoleToUser(userID, roleID)
    if err != nil {
        return err
    }
    
    // 2. 清除用户角色缓存
    _ = cacheService.ClearUserRoles(ctx, userID)
    
    // 3. 清除用户权限缓存
    _ = cacheService.ClearUserPermissions(ctx, userID)
    
    return nil
}
```

---

## 📋 缓存Key规范

### 命名规范

```
格式: <资源类型>:<标识符>[:子类型]

示例:
- user:123                  // 用户信息
- user:roles:123            // 用户角色
- permission:123:GET:/api   // 用户权限
- token:abc123              // Token黑名单
- session:123               // 用户会话
```

### 已定义的Key前缀

| 前缀 | 格式 | 用途 | TTL |
|------|------|------|-----|
| `permission:` | `permission:用户ID:路径:方法` | 权限缓存 | 10分钟 |
| `user:` | `user:用户ID` | 用户信息 | 30分钟 |
| `user:roles:` | `user:roles:用户ID` | 用户角色 | 30分钟 |
| `token:` | `token:token字符串` | Token黑名单 | 24小时 |
| `session:` | `session:用户ID` | 用户会话 | 1小时 |

---

## ⚙️ 配置说明

### TTL配置

在 `cache_service.go` 中定义：

```go
const (
    ttlPermission = 10 * time.Minute  // 权限缓存10分钟
    ttlUser       = 30 * time.Minute  // 用户信息30分钟
    ttlRole       = 30 * time.Minute  // 角色信息30分钟
    ttlToken      = 24 * time.Hour    // Token 24小时
    ttlSession    = 1 * time.Hour     // 会话1小时
)
```

**调整建议**：
- 权限缓存：根据权限变更频率调整（5-30分钟）
- 用户信息：根据用户信息更新频率（10-60分钟）
- Token黑名单：应与JWT过期时间一致
- 会话：根据业务需求（30分钟-24小时）

---

## 🔍 监控和调试

### 获取缓存统计

```go
stats, err := cacheService.GetCacheStats(ctx)
fmt.Printf("缓存状态: %+v\n", stats)
```

### 清空所有缓存（开发/测试环境）

```go
// ⚠️ 谨慎使用！这会清空Redis中的所有数据
err := cacheService.ClearAll(ctx)
```

---

## 🐛 常见问题

### Q1: Redis连接失败怎么办？

**A**: 缓存服务会自动降级，不影响业务逻辑。

```go
// 即使Redis不可用，这个调用也不会报错
// 它会自动查询数据库
hasPermission, err := cacheService.CheckUserPermission(ctx, userID, path, method)
```

### Q2: 如何判断缓存是否生效？

**A**: 查看日志或使用Redis客户端：

```bash
# 连接Redis
redis-cli

# 查看所有key
KEYS *

# 查看特定key
GET permission:1:GET:/api/v1/users

# 查看key的TTL
TTL permission:1:GET:/api/v1/users
```

### Q3: 缓存更新策略？

**A**: 采用 **Cache-Aside** 模式：

1. 读取时：先读缓存，未命中再读数据库
2. 更新时：先更新数据库，再删除缓存
3. 下次读取时自动更新缓存

```go
// 更新用户信息
func UpdateUser(ctx context.Context, user *rbac.User) error {
    // 1. 更新数据库
    err := rbac.UpdateUser(user)
    if err != nil {
        return err
    }
    
    // 2. 删除缓存
    _ = cacheService.DeleteUser(ctx, user.ID)
    
    return nil
}
```

---

## 📚 相关文档

- [Redis最佳实践](https://redis.io/topics/best-practices)
- [缓存设计模式](https://docs.microsoft.com/en-us/azure/architecture/patterns/cache-aside)
- [项目优化报告](../OPTIMIZATION_REPORT.md)

---

## 🎉 总结

缓存服务提供了：

- ✅ 统一的缓存接口
- ✅ 自动的降级处理
- ✅ 丰富的功能支持
- ✅ 清晰的使用示例
- ✅ 完善的错误处理

开始使用缓存服务，让你的应用性能提升10倍！🚀

