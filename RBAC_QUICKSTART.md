# RBAC 权限系统快速开始 🚀

## 🎯 核心理念

**一行代码声明权限组，框架自动完成所有配置！**

## ✨ 使用方法

### 1️⃣ 注册需要权限的路由

```go
// 使用 SetPermission() 声明权限组
userGroup := routegroup.WithAuthRouterGroup(api.Group("/users")).
    SetPermission("user:manage", "用户管理")  // ← 就是这里！
userGroup.Use(middleware.JWT())
{
    userGroup.GET("", handler.GetUsers)
    userGroup.POST("", handler.CreateUser)
    userGroup.PUT("/:id", handler.UpdateUser)
    userGroup.DELETE("/:id", handler.DeleteUser)
}
```

### 2️⃣ 启动服务

```bash
go run main.go
```

系统会自动：
- ✅ 创建权限组 `user:manage`
- ✅ 同步所有路由到 `resources` 表
- ✅ 绑定资源到权限组
- ✅ 创建超级管理员角色
- ✅ 创建默认管理员账号

### 3️⃣ 使用默认管理员登录

```bash
POST /api/v1/user/login
{
  "username": "admin",
  "password": "admin123"
}
```

## 📋 完整示例

```go
package rbac

import (
    "gin-template/internal/logic/v1/rbac"
    "gin-template/internal/middleware"
    "gin-template/internal/routegroup"
    "github.com/gin-gonic/gin"
)

func RegisterRBACRoutes(api *gin.RouterGroup) {
    // 公共路由（无需权限）
    authGroup := api.Group("/auth")
    {
        authGroup.POST("/login", rbac.Login)
        authGroup.POST("/register", rbac.Register)
    }

    // 用户管理（需要 user:manage 权限）
    userGroup := routegroup.WithAuthRouterGroup(api.Group("/user")).
        SetPermission("user:manage", "用户管理")
    userGroup.Use(middleware.JWT())
    {
        userGroup.GET("/profile", rbac.GetProfile)
        userGroup.GET("/:id/roles", rbac.GetUserRoles)
        userGroup.POST("/:id/roles", rbac.AssignRoleToUser)
    }

    // 角色管理（需要 role:manage 权限）
    roleGroup := routegroup.WithAuthRouterGroup(api.Group("/roles")).
        SetPermission("role:manage", "角色管理")
    roleGroup.Use(middleware.JWT())
    {
        roleGroup.GET("", rbac.GetRoles)
        roleGroup.POST("", rbac.CreateRole)
        roleGroup.PUT("/:id", rbac.UpdateRole)
        roleGroup.DELETE("/:id", rbac.DeleteRole)
    }

    // 权限管理（需要 permission:manage 权限）
    permissionGroup := routegroup.WithAuthRouterGroup(api.Group("/permissions")).
        SetPermission("permission:manage", "权限管理")
    permissionGroup.Use(middleware.JWT())
    {
        permissionGroup.GET("", rbac.GetPermissions)
        permissionGroup.POST("", rbac.CreatePermission)
    }
}
```

## 🎨 权限组命名规范

推荐使用 `模块:操作` 格式：

```go
"user:manage"      // 用户管理
"user:view"        // 用户查看
"order:manage"     // 订单管理
"order:view"       // 订单查看
"product:manage"   // 产品管理
"report:export"    // 报表导出
```

## 🔐 权限组管理

**完全自动化！** 权限组会从路由声明中自动提取和创建，无需手动配置。

系统启动时会：
1. 扫描所有 `SetPermission()` 声明
2. 自动提取唯一的权限组
3. 创建权限组到数据库
4. 自动绑定给超级管理员角色

例如，你的路由声明了：
```go
SetPermission("user:manage", "用户管理")
SetPermission("role:manage", "角色管理")
SetPermission("order:manage", "订单管理")
```

系统会自动创建这 3 个权限组，无需其他配置！

## 🔄 工作流程

```
开发者声明权限组
    ↓
系统启动时自动初始化
    ↓
创建权限组 → 同步资源 → 绑定关系
    ↓
创建超管角色 → 创建管理员账号
    ↓
完成！可以使用 admin/admin123 登录
```

## 🆚 对比传统方式

### ❌ 传统方式

```sql
-- 需要手动执行大量 SQL
INSERT INTO permissions (code, name) VALUES ('user:manage', '用户管理');
INSERT INTO resources (path, method, permission_id) VALUES ('/api/users', 'GET', 1);
INSERT INTO resources (path, method, permission_id) VALUES ('/api/users', 'POST', 1);
-- ... 几十条 SQL
INSERT INTO roles (name) VALUES ('超级管理员');
INSERT INTO role_permissions (role_id, permission_id) VALUES (1, 1);
-- ... 更多 SQL
```

### ✅ 本框架

```go
// 一行代码搞定！
userGroup := routegroup.WithAuthRouterGroup(api.Group("/users")).
    SetPermission("user:manage", "用户管理")
```

## 💡 常见场景

### 场景 1: 添加新模块

```go
// 只需声明权限组，系统自动处理其他
productGroup := routegroup.WithAuthRouterGroup(api.Group("/products")).
    SetPermission("product:manage", "产品管理")
productGroup.Use(middleware.JWT())
{
    productGroup.GET("", handler.ListProducts)
    productGroup.POST("", handler.CreateProduct)
}
```

### 场景 2: 子路由继承权限

```go
// 父路由组
orderGroup := routegroup.WithAuthRouterGroup(api.Group("/orders")).
    SetPermission("order:view", "订单查看")
orderGroup.Use(middleware.JWT())
{
    orderGroup.GET("", handler.ListOrders)
    
    // 子路由组自动继承父权限
    detailGroup := orderGroup.Group("/:id")
    {
        detailGroup.GET("", handler.GetOrder)
    }
}
```

### 场景 3: 子路由覆盖权限

```go
// 查看订单需要较低权限
viewGroup := routegroup.WithAuthRouterGroup(api.Group("/orders")).
    SetPermission("order:view", "订单查看")
viewGroup.Use(middleware.JWT())
{
    viewGroup.GET("", handler.ListOrders)
    
    // 管理订单需要更高权限
    manageGroup := viewGroup.Group("/").
        SetPermission("order:manage", "订单管理")
    {
        manageGroup.POST("", handler.CreateOrder)
        manageGroup.DELETE("/:id", handler.DeleteOrder)
    }
}
```

## ⚙️ 自定义配置

### 修改默认管理员信息

编辑 `internal/model/rbac/init.go`：

```go
const (
    DefaultAdminUsername = "admin"              // 修改用户名
    DefaultAdminPassword = "your-password"      // 修改密码
    DefaultAdminEmail    = "admin@example.com"  // 修改邮箱
    DefaultRoleName      = "超级管理员"
)
```

### 添加新权限组

**无需配置文件！** 直接在路由注册时声明即可：

```go
// 在你的路由文件中直接声明
customGroup := routegroup.WithAuthRouterGroup(api.Group("/custom")).
    SetPermission("custom:manage", "自定义管理")
customGroup.Use(middleware.JWT())
{
    customGroup.GET("", handler.ListCustom)
    customGroup.POST("", handler.CreateCustom)
}
```

系统启动时会自动：
- ✅ 发现 `custom:manage` 权限组
- ✅ 创建到数据库
- ✅ 绑定给超级管理员
- ✅ 关联对应的路由资源

## 🔍 查看初始化结果

启动服务后，查看日志：

```
[INFO] 开始初始化 RBAC 权限系统...
[INFO]   - 从路由声明中提取并创建权限组...
[INFO]     ✓ 创建权限组: 用户管理 (user:manage)
[INFO]     ✓ 创建权限组: 角色管理 (role:manage)
[INFO]     ✓ 创建权限组: 权限管理 (permission:manage)
[INFO]     ✓ 创建权限组: 资源查看 (resource:view)
[INFO]     ✓ 创建权限组: 认证管理 (auth:manage)
[INFO]     ✓ 从路由声明中发现 5 个权限组，新建 5 个
[INFO]   - 同步路由资源到数据库...
[INFO]     ✓ 同步了 15 个路由资源
[INFO]   - 自动绑定资源到权限组...
[INFO]     ✓ 成功绑定 15 个资源到权限组
[INFO]   - 初始化超级管理员角色...
[INFO]     ✓ 创建角色: 超级管理员
[INFO]   - 绑定所有权限组到超级管理员角色...
[INFO]     ✓ 成功绑定 5 个权限组到超级管理员
[INFO]   - 初始化默认管理员用户...
[INFO]     ✓ 创建管理员用户: admin
[INFO]     ✓ 分配超级管理员角色
[INFO] ✓ RBAC 权限系统初始化成功
[INFO] ✓ 默认管理员账号: admin / admin123
```

## 📚 更多文档

详细文档请参考：[docs/RBAC_AUTO_INIT.md](docs/RBAC_AUTO_INIT.md)

## 🎉 总结

使用本框架的 RBAC 系统，你只需要：

1. **声明权限组** - 使用 `SetPermission()`
2. **启动服务** - 系统自动初始化
3. **开始使用** - 用默认账号登录

就是这么简单！🚀

---

**一句话总结：声明即生效，框架帮你搞定一切！**

