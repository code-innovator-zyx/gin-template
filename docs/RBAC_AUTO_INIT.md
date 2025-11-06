# RBAC 权限自动初始化使用指南

## 概述

本框架提供了企业级的 RBAC（基于角色的访问控制）权限自动初始化方案，无需手动配置权限组和资源绑定，系统会在启动时自动完成所有初始化工作。

## 核心特性

✅ **声明式权限组管理** - 在路由注册时声明权限组，无需手动绑定  
✅ **自动资源同步** - 系统启动时自动同步所有受保护的路由到数据库  
✅ **智能权限绑定** - 根据代码声明自动将资源绑定到权限组  
✅ **默认管理员创建** - 自动创建超级管理员角色和默认管理员账号  
✅ **幂等性保证** - 多次执行不会产生副作用，安全可靠  
✅ **零配置文件** - 所有配置通过代码声明，便于版本控制和团队协作

## 快速开始

### 1. 注册需要权限保护的路由

使用 `SetPermission()` 方法为路由组声明所属的权限组：

```go
// 用户管理模块 - 声明权限组
userGroup := routegroup.WithAuthRouterGroup(api.Group("/users")).
    SetPermission("user:manage", "用户管理")
userGroup.Use(middleware.JWT())
{
    userGroup.GET("", handler.GetUsers)           // 获取用户列表
    userGroup.POST("", handler.CreateUser)        // 创建用户
    userGroup.PUT("/:id", handler.UpdateUser)     // 更新用户
    userGroup.DELETE("/:id", handler.DeleteUser)  // 删除用户
}
```

### 2. 系统自动初始化

系统启动时会自动执行以下操作：

1. ✓ 创建默认权限组（基于代码中的 `SetPermission` 声明）
2. ✓ 同步所有受保护的路由到 `resources` 表
3. ✓ 将资源自动绑定到对应的权限组
4. ✓ 创建"超级管理员"角色
5. ✓ 创建默认管理员账号并分配角色

### 3. 使用默认管理员账号登录

系统初始化后，可以使用以下默认账号登录：

- **用户名**: `admin`
- **密码**: `admin123`

⚠️ **生产环境请务必修改默认密码！**

## 使用示例

### 示例 1: 用户管理模块

```go
func RegisterUserRoutes(api *gin.RouterGroup) {
    // 公共路由（无需权限）
    publicGroup := api.Group("/user")
    {
        publicGroup.POST("/register", handler.Register)
        publicGroup.POST("/login", handler.Login)
    }
    
    // 需要权限的路由
    authGroup := routegroup.WithAuthRouterGroup(api.Group("/user")).
        SetPermission("user:manage", "用户管理")
    authGroup.Use(middleware.JWT())
    {
        authGroup.GET("/profile", handler.GetProfile)
        authGroup.PUT("/profile", handler.UpdateProfile)
        authGroup.GET("/:id/roles", handler.GetUserRoles)
        authGroup.POST("/:id/roles", handler.AssignRole)
    }
}
```

### 示例 2: 嵌套路由组

子路由组会自动继承父路由组的权限组设置：

```go
// 产品管理模块
productGroup := routegroup.WithAuthRouterGroup(api.Group("/products")).
    SetPermission("product:manage", "产品管理")
productGroup.Use(middleware.JWT())
{
    productGroup.GET("", handler.ListProducts)
    productGroup.POST("", handler.CreateProduct)
    
    // 子路由组自动继承 "product:manage" 权限组
    detailGroup := productGroup.Group("/:id")
    {
        detailGroup.GET("", handler.GetProduct)
        detailGroup.PUT("", handler.UpdateProduct)
        detailGroup.DELETE("", handler.DeleteProduct)
    }
}
```

### 示例 3: 覆盖子路由组的权限

如果需要，子路由组可以声明自己的权限组：

```go
// 订单模块
orderGroup := routegroup.WithAuthRouterGroup(api.Group("/orders")).
    SetPermission("order:view", "订单查看")
orderGroup.Use(middleware.JWT())
{
    orderGroup.GET("", handler.ListOrders)
    orderGroup.GET("/:id", handler.GetOrder)
    
    // 订单管理需要更高权限
    manageGroup := orderGroup.Group("/").
        SetPermission("order:manage", "订单管理")
    {
        manageGroup.POST("", handler.CreateOrder)
        manageGroup.PUT("/:id", handler.UpdateOrder)
        manageGroup.DELETE("/:id", handler.CancelOrder)
    }
}
```

## 权限组管理

**完全自动化！** 系统不再使用硬编码的权限组配置。

### 工作原理

权限组完全从路由声明中自动提取：

1. **开发者声明** - 在路由注册时使用 `SetPermission("code", "name")`
2. **自动提取** - 系统启动时扫描所有路由声明
3. **去重创建** - 自动提取唯一的权限组并创建
4. **自动绑定** - 所有权限组自动绑定给超级管理员

### 示例

如果你的代码中有：
```go
SetPermission("user:manage", "用户管理")
SetPermission("role:manage", "角色管理")
SetPermission("order:manage", "订单管理")
```

系统会自动创建这 3 个权限组，**无需任何配置文件**！

## 添加新权限组

**超级简单！** 只需在路由注册时声明即可：

```go
// 在路由文件中声明新的权限组
reportGroup := routegroup.WithAuthRouterGroup(api.Group("/reports")).
    SetPermission("report:view", "报表查看")  // ← 这就完成了！
reportGroup.Use(middleware.JWT())
{
    reportGroup.GET("", handler.ListReports)
    reportGroup.GET("/:id", handler.GetReport)
}

// 需要更高权限的操作
exportGroup := routegroup.WithAuthRouterGroup(api.Group("/reports")).
    SetPermission("report:export", "报表导出")  // ← 另一个权限组
exportGroup.Use(middleware.JWT())
{
    exportGroup.POST("/export", handler.ExportReport)
}
```

系统启动时会自动：
- ✅ 发现 `report:view` 和 `report:export` 两个权限组
- ✅ 创建到数据库
- ✅ 绑定给超级管理员
- ✅ 关联对应的路由资源

**就是这么简单，无需任何额外配置！**

## 数据库表结构

### 核心表说明

- **`permissions`** - 权限组表，存储权限组的基本信息
- **`resources`** - 资源表，存储所有受保护的路由资源
- **`roles`** - 角色表，存储系统角色
- **`role_permissions`** - 角色-权限关联表（多对多）
- **`user_roles`** - 用户-角色关联表（多对多）

### 资源与权限组的关系

```
Permission (1) ←→ (N) Resource
    ↓
    通过 permission_id 字段关联
    is_managed = true 表示资源已被权限组管理
```

## 工作原理

### 1. 路由收集阶段

```go
// 使用 AuthRouterGroup 注册路由时
authGroup := routegroup.WithAuthRouterGroup(api.Group("/users")).
    SetPermission("user:manage", "用户管理")
authGroup.GET("/:id", handler.GetUser)

// 系统自动收集路由信息到内存
ProtectedRoute {
    Resource: {Path: "/api/v1/users/:id", Method: "GET"},
    PermissionCode: "user:manage",
    PermissionName: "用户管理"
}
```

### 2. 初始化阶段

系统启动时，在 `handler.Init()` 中调用 `rbac.InitializeRBAC()`：

```go
func Init() *gin.Engine {
    // ... 路由注册
    v1.RegisterRoutes(r)
    
    // 获取收集到的路由并转换格式
    protectedRoutes := convertRoutes(routegroup.GetProtectedRoutes())
    
    // 执行自动初始化
    rbac.InitializeRBAC(protectedRoutes)
    
    return r
}
```

### 3. 初始化流程

```
1. 创建默认权限组
   ↓
2. 同步路由资源到数据库
   ↓
3. 根据 PermissionCode 自动绑定资源到权限组
   ↓
4. 创建超级管理员角色
   ↓
5. 绑定所有权限组到超级管理员角色
   ↓
6. 创建默认管理员用户并分配角色
```

## 权限检查流程

当用户访问受保护的资源时：

```go
// middleware.Permission() 中间件会：
1. 从 JWT 中提取 user_id
2. 获取请求的 path 和 method
3. 调用 rbac.CheckPermission(userID, path, method)
4. 查询用户是否拥有访问该资源的权限
5. 允许或拒绝访问
```

SQL 查询逻辑：

```sql
-- 检查用户是否有权限访问指定资源
SELECT COUNT(*) FROM resources r
JOIN role_permissions rp ON r.permission_id = rp.permission_id
JOIN user_roles ur ON rp.role_id = ur.role_id
WHERE ur.user_id = ? AND r.path = ? AND r.method = ?
```

## 常见问题

### Q: 如何添加新的权限组？

**A:** 直接在路由注册时使用 `SetPermission()` 声明即可。系统会自动：
1. 提取权限组信息
2. 创建权限组到数据库
3. 绑定资源到该权限组
4. 自动绑定给超级管理员

```go
// 就这一行！
newGroup := routegroup.WithAuthRouterGroup(api.Group("/new-module")).
    SetPermission("new:manage", "新模块管理")
```

无需任何配置文件或额外代码！

### Q: 如何让某个路由不需要权限？

**A:** 不要使用 `AuthRouterGroup`，直接使用普通的 `gin.RouterGroup`：

```go
// 公共路由，无需权限
publicGroup := api.Group("/public")
{
    publicGroup.GET("/health", handler.Health)
}
```

### Q: 新添加的权限组会自动绑定给超级管理员吗？

**A:** 是的！系统启动时会：
1. 扫描所有路由声明
2. 提取所有权限组
3. **自动绑定给超级管理员角色**

这样超级管理员始终拥有所有权限。

### Q: 如何修改默认管理员密码？

**A:** 在 `internal/model/rbac/init.go` 中修改常量：

```go
const (
    DefaultAdminUsername = "admin"
    DefaultAdminPassword = "your-secure-password"  // 修改这里
    DefaultAdminEmail    = "admin@example.com"
)
```

### Q: 能否禁用自动初始化？

**A:** 可以，但不推荐。如果确实需要，可以注释掉 `handler/router.go` 中的初始化调用：

```go
// if err := rbac.InitializeRBAC(protectedRoutes); err != nil {
//     logrus.Fatalf("RBAC 权限系统初始化失败: %v", err)
// }
```

### Q: 系统重启会重复创建数据吗？

**A:** 不会。初始化逻辑是幂等的，会先检查数据是否存在，只创建缺失的部分。

### Q: 如何给普通用户分配权限？

**A:** 通过 API 为用户分配角色：

```bash
# 1. 创建角色
POST /api/v1/roles
{"name": "普通用户", "description": "普通用户角色"}

# 2. 为角色分配权限
POST /api/v1/roles/:role_id/permissions
{"permission_id": 1}

# 3. 为用户分配角色
POST /api/v1/user/:user_id/roles
{"role_id": 1}
```

## 最佳实践

### 1. 权限组粒度

建议按**业务模块**划分权限组，而不是按单个接口：

```go
// ✅ 推荐：按模块划分
"user:manage"     // 用户管理（包含所有用户相关操作）
"order:manage"    // 订单管理
"product:manage"  // 产品管理

// ❌ 不推荐：粒度过细
"user:create"
"user:update"
"user:delete"
"user:view"
```

### 2. 权限组命名规范

建议使用 `模块:操作` 的格式：

```go
"user:manage"      // 用户管理
"order:view"       // 订单查看
"order:manage"     // 订单管理
"report:export"    // 报表导出
```

### 3. 角色设计

建议根据职责划分角色：

- **超级管理员** - 拥有所有权限
- **运营人员** - 用户管理、订单管理
- **财务人员** - 订单查看、报表导出
- **普通用户** - 基本功能访问

### 4. 安全建议

- ✅ 生产环境修改默认管理员密码
- ✅ 定期审计用户权限
- ✅ 遵循最小权限原则
- ✅ 敏感操作添加二次验证
- ✅ 记录权限变更日志

## 进阶用法

### 动态权限控制

如果需要更细粒度的权限控制，可以在 handler 中进一步判断：

```go
func UpdateUser(c *gin.Context) {
    userID := c.GetUint("user_id")  // 从 JWT 中获取
    targetID := c.Param("id")
    
    // 只能修改自己的信息，除非是管理员
    if userID != targetID && !isAdmin(userID) {
        response.Error(c, errcode.ErrForbidden)
        return
    }
    
    // ... 业务逻辑
}
```

### 跳过权限检查

某些路由需要 JWT 认证但不需要权限检查：

```go
// 只使用 JWT 中间件，不使用 Permission 中间件
authGroup := routegroup.WithAuthRouterGroup(api.Group("/profile"))
authGroup.Use(middleware.JWT())  // 只验证登录，不验证权限
// 不调用 SetPermission
{
    authGroup.GET("", handler.GetProfile)
    authGroup.PUT("", handler.UpdateProfile)
}
```

## 总结

本框架的 RBAC 自动初始化方案具有以下优势：

1. **零配置** - 无需配置文件，完全从代码声明中自动提取
2. **开发效率高** - 声明即生效，无需手动维护权限组列表
3. **维护成本低** - 单一数据源（路由声明），不会出现配置不同步
4. **扩展性强** - 添加新模块时，只需声明路由，自动完成所有配置
5. **安全可靠** - 幂等操作，所有权限自动绑定给超级管理员
6. **企业级** - 完整的 RBAC 模型，满足复杂权限需求

### 核心设计理念

传统方案需要在多个地方维护权限信息：
- ❌ 配置文件中定义权限组
- ❌ 路由注册时关联权限
- ❌ 数据库初始化脚本
- ❌ 手动绑定权限到角色

本框架的方案：
- ✅ **只在路由注册时声明一次**
- ✅ 系统自动提取、创建、绑定
- ✅ 单一数据源，避免配置漂移

开始使用时，只需记住一句话：

> **使用 `SetPermission()` 声明权限组，框架自动搞定一切！**

---

如有问题或建议，欢迎提 Issue 或 PR！

