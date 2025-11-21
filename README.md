span

# 🚀 Gin Admin

**生产级 Go Web 应用开发脚手架**

一个功能完备、开箱即用的企业级 Gin 框架后端模板，助力快速构建高性能、安全可靠的 Web 应用

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Gin Version](https://img.shields.io/badge/Gin-1.9-00ADD8?style=flat&logo=go)](https://github.com/gin-gonic/gin)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/the-yex/gin-admin?style=social)](https://github.com/the-yex/gin-admin/stargazers)
[![GitHub Forks](https://img.shields.io/github/forks/the-yex/gin-admin?style=social)](https://github.com/the-yex/gin-admin/network/members)

[English](README_EN.md) | 简体中文

[快速开始](#-快速开始) • [核心特性](#-核心特性) • [在线文档](#-api-文档) • [贡献指南](#-贡献指南)

</div>

---

## 📖 项目简介

Gin Admin 是一个开箱即用的 Go 语言后端开发模板，基于 [Gin](https://github.com/gin-gonic/gin) 框架构建，集成了企业级项目开发所需的核心功能模块。无论你是在构建 RESTful API、微服务，还是完整的 Web 应用后端，这个模板都能帮你节省大量基础设施搭建时间，让你专注于业务逻辑开发。

> 🎨 **配套前端项目**：[gin-admin-web](https://github.com/the-yex/gin-admin-web) - 开箱即用的前端管理系统，完美适配本后端的 RBAC 权限设计！

### 🎯 为什么选择 Gin Admin？

- ⚡ **开箱即用**：克隆即可运行，无需复杂配置
- 🏗️ **最佳实践**：严格遵循 Go 项目布局和代码规范
- 🔐 **安全第一**：完善的 RBAC 权限系统和 JWT 认证
- 🤖 **路由即权限**：革命性的路由自动注册机制，添加路由 = 自动管理权限，零额外配置
- 🎨 **全栈方案**：配套前端 [gin-admin-web](https://github.com/the-yex/gin-admin-web)，前后端完美联动
- 🚢 **生产就绪**：Docker 容器化、优雅关闭、健康检查一应俱全
- 📚 **文档完善**：自动生成的 Swagger API 文档
- 🛠️ **开发友好**：强大的 Makefile 工具链和热重载支持

---

## ✨ 核心特性

### 🏛️ 架构设计

- **🎨 清晰的分层架构**

  - Handler（路由层）→ Logic（业务逻辑层）→ Service（服务层）→ Model（数据层）
  - 严格的职责分离，便于测试和维护
  - 模块化设计，支持快速扩展
- **⚙️ 灵活的配置管理**

  - 基于 Viper 的强大配置系统
  - 支持 YAML、JSON、环境变量等多种配置方式
  - 多环境配置支持（开发、测试、生产）

### 🔒 安全与认证

- **🛡️ 完善的 RBAC 权限系统 + 路由自动注册**

  - 🚀 **革命性设计**：添加路由时自动注册到权限系统，启动时自动同步最新资源
  - 🎯 **零额外配置**：不需要手动管理权限表，不需要写 SQL，不需要配置文件
  - 📝 **声明式权限**：一行 `WithMeta()` 声明权限组，系统自动完成一切
  - 🔄 **自动同步**：每次启动自动扫描路由变更，新增/删除路由自动更新数据库
  - 🎨 **UI 友好分组**：Permission（权限组）用于前端展示，Resource（资源）用于实际授权
  - 🔐 **实际授权路径**：用户 → 角色 → 资源（API 级别精确控制）
  - 🛡️ **默认拒绝策略**：未授权资源自动拒绝访问
  - [📖 详细了解 RBAC 设计](RBAC_QUICKSTART.md)
- **🔑 JWT 身份认证**

  - 基于 JWT 的无状态认证
  - Token 自动刷新机制
  - 安全的密码加密存储（bcrypt）

### 🧩 中间件生态

内置 **8 个生产级中间件**，开箱即用：


| 中间件          | 功能说明                    |
| --------------- | --------------------------- |
| 🔐 JWT Auth     | JWT 令牌验证和用户身份识别  |
| 🚦 CORS         | 跨域资源共享配置            |
| 📝 Logger       | 结构化请求日志记录          |
| 🔄 Recovery     | 优雅的 Panic 恢复和错误处理 |
| 🎫 Request ID   | 为每个请求生成唯一追踪 ID   |
| 🔐 Permission   | RBAC 权限验证               |
| ⏱️ Rate Limit | 基于 Token Bucket 的限流器  |
| 📊 Metrics      | 请求指标统计和监控          |

### 💾 数据与缓存

- **🗄️ 数据库支持**

  - 基于 GORM v2 的 ORM
  - 支持 MySQL、PostgreSQL、SQLite 等主流数据库
  - 自动迁移和模型管理
  - 连接池优化配置
- **⚡ Redis 缓存集成**

  - 开箱即用的 Redis 客户端
  - 支持缓存预热和过期策略
  - 分布式锁实现

### 📊 日志与监控

- **📋 专业日志系统**
  - 基于 Logrus 的结构化日志
  - 支持多级别日志（Debug、Info、Warn、Error）
  - 日志文件自动轮转（Lumberjack）
  - 支持 JSON 格式输出，便于日志收集

### 🚀 DevOps 支持

- **🐳 Docker 容器化**

  - 多阶段构建 Dockerfile，镜像体积小
  - Docker Compose 一键启动完整环境
  - 包含 MySQL 和 Redis 服务编排
- **🛠️ 强大的 Makefile**

  - `make run` - 快速运行应用
  - `make build` - 构建二进制文件
  - `make build-all` - 跨平台编译（Linux/macOS/Windows）
  - `make swagger_v1` - 生成 API 文档
  - `make test` - 运行测试套件
  - `make dev` - 热重载开发模式（需要 air）
  - `make rename` - 快速重命名项目
  - [查看所有命令](#-使用-makefile)

### 📚 文档

- **📖 Swagger API 文档**
  - 自动生成的交互式 API 文档
  - 支持在线测试接口
  - 访问地址：`http://localhost:8080/swagger/index.html`

---

## 🚀 快速开始

### 方式一：Docker Compose（推荐）

**最快 30 秒启动完整环境！**

```bash
# 1. 克隆项目
git clone https://github.com/the-yex/gin-admin.git
cd gin-admin

# 2. 初始化配置
make init-config

# 3. 启动所有服务（应用 + MySQL + Redis）
make up

# 4. 查看日志
make logs
```

🎉 访问 http://localhost:8080/swagger/index.html 查看 API 文档！

### 方式二：本地运行

#### 前置要求

- Go 1.20 或更高版本
- MySQL 5.7+ / PostgreSQL / SQLite
- Redis（可选，用于缓存）

#### 安装步骤

```bash
# 1. 克隆项目
git clone https://github.com/the-yex/gin-admin.git
cd gin-admin

# 2. 安装依赖
go mod tidy

# 3. 初始化配置文件
make init-config

# 4. 编辑配置文件（修改数据库连接等）
vim app.yaml
```

编辑 `app.yaml` 配置文件：

```yaml
server:
  port: 8080

database:
  driver: mysql
  host: localhost
  port: 3306
  username: root
  password: your_password
  dbname: gin_admin
  
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: your-secret-key-change-in-production
  expire: 86400  # 24小时
```

```bash
# 5. 运行应用
make run

# 或直接使用 Go 命令
go run main.go
```

#### 🧪 测试 API

```bash
# 健康检查
curl http://localhost:8080/health

# 查看 API 文档
open http://localhost:8080/swagger/index.html
```

---

## 📁 项目结构

```
gin-admin/
├── 📄 main.go                 # 应用入口
├── 📄 Makefile                # Make 命令集合
├── 📄 Dockerfile              # Docker 构建文件
├── 📄 docker-compose.yml      # Docker Compose 编排
├── 📄 app.yaml                # 应用配置文件
│
├── 📂 internal/               # 私有应用代码
│   ├── 📂 config/            # 配置管理
│   ├── 📂 core/              # 核心初始化逻辑
│   ├── 📂 handler/           # HTTP 处理器（路由层）
│   │   └── 📂 v1/            # API v1 版本
│   ├── 📂 logic/             # 业务逻辑层
│   │   └── 📂 v1/            # v1 业务逻辑
│   ├── 📂 middleware/        # 中间件
│   │   ├── 📄 jwt.go         # JWT 认证
│   │   ├── 📄 permission.go  # RBAC 权限
│   │   ├── 📄 rate_limit.go  # 限流器
│   │   ├── 📄 cors.go        # 跨域处理
│   │   ├── 📄 logger.go      # 请求日志
│   │   ├── 📄 recovery.go    # Panic 恢复
│   │   ├── 📄 request_id.go  # 请求追踪
│   │   └── 📄 metrics.go     # 监控指标
│   ├── 📂 model/             # 数据模型
│   │   ├── 📂 rbac/          # RBAC 模型
│   │   └── 📄 migrate.go     # 数据库迁移
│   ├── 📂 service/           # 业务服务层
│   ├── 📂 routegroup/        # 路由分组
│   └── 📂 types/             # 类型定义
│
├── 📂 pkg/                    # 可复用的公共库
│   ├── 📂 cache/             # 缓存工具（Redis）
│   ├── 📂 logger/            # 日志工具
│   ├── 📂 orm/               # ORM 配置
│   ├── 📂 jwt/               # JWT 工具
│   ├── 📂 response/          # 统一响应格式
│   ├── 📂 validator/         # 参数验证
│   ├── 📂 errcode/           # 错误码定义
│   ├── 📂 transaction/       # 事务管理
│   └── 📂 utils/             # 通用工具
│
├── 📂 docs/                   # Swagger API 文档
│   ├── 📄 v1_docs.go
│   ├── 📄 v1_swagger.json
│   └── 📄 v1_swagger.yaml
│
├── 📂 scripts/                # 脚本工具
├── 📂 logs/                   # 日志文件目录
└── 📂 build/                  # 编译输出目录
```

---

## 🔧 使用 Makefile

项目提供了丰富的 Makefile 命令来简化开发流程：

### 🏃 运行与构建

```bash
make run              # 运行应用
make build            # 编译应用（当前平台）
make build-linux      # 编译 Linux 版本
make build-darwin     # 编译 macOS 版本
make build-windows    # 编译 Windows 版本
make build-all        # 编译所有平台版本
```

### 🧪 测试与检查

```bash
make test             # 运行测试
make test-coverage    # 生成测试覆盖率报告
make lint             # 代码风格检查
make fmt              # 格式化代码
make vet              # 静态分析
make check            # 运行所有检查（fmt + vet + lint）
```

### 📖 文档

```bash
make swagger_v1       # 生成 Swagger API 文档
```

### 🐳 Docker

```bash
make docker-build     # 构建 Docker 镜像
make docker-run       # 运行 Docker 容器
make docker-stop      # 停止 Docker 容器
make up               # 启动 Docker Compose 服务
make down             # 停止 Docker Compose 服务
make logs             # 查看服务日志
```

### 🛠️ 工具

```bash
make init-config      # 初始化配置文件
make rename NEW_NAME=your-project  # 重命名项目
make dev              # 热重载开发模式（需要 air）
make install          # 安装依赖
make clean            # 清理构建文件
make help             # 查看所有可用命令
```

---

## 🎯 RBAC 权限系统 + 路由自动注册机制

本项目实现了一套**革命性的 RBAC 权限管理系统**，最大亮点是**路由自动注册机制**：

> 💡 **核心创新**：添加路由即完成权限配置，启动时自动同步，无需手动管理权限表！

### 🌟 为什么说是革命性的？

#### 传统 RBAC 的痛点 ❌

```sql
-- 😓 每次添加新 API 都要写一堆 SQL
INSERT INTO permissions (code, name) VALUES ('user:create', '创建用户');
INSERT INTO resources (path, method, permission_id) VALUES ('/api/v1/users', 'POST', 1);
INSERT INTO role_permissions (role_id, permission_id) VALUES (1, 1);
-- 维护成本高，容易遗漏，容易出错
```

#### 本框架的解决方案 ✅

```go
// 😎 只需一行声明，其他全自动！
userGroup := routegroup.WrapGroup(api.Group("/users")).
    WithMeta("user:manage", "用户管理")
{
    userGroup.GET("", handler.GetUsers)      // 自动注册！
    userGroup.POST("", handler.CreateUser)   // 自动注册！
    userGroup.PUT("/:id", handler.UpdateUser) // 自动注册！
}
```

**启动时自动发生的魔法** ✨：
1. 📡 扫描所有路由定义
2. 🔍 识别权限组声明（`WithMeta()`）
3. 📝 自动创建/更新权限组到数据库
4. 🔗 自动将路由资源关联到权限组
5. 🔐 自动绑定资源到超级管理员角色
6. 🗑️ 自动清理已删除的路由资源

### 架构设计

```
用户（User） ──→ 角色（Role） ──→ 资源（Resource / API） [实际授权路径]
                                     ↓
                              Permission（权限组） [仅用于 UI 分组展示]
```

**设计理念**：
- **实际授权**：角色直接绑定资源（API 路径 + HTTP 方法）
- **UI 分组**：Permission 仅用于前端页面的逻辑分组和展示
- **自动同步**：路由变更自动反映到权限系统

### 核心特性

- ✅ **路由即权限**：添加路由 = 自动注册资源，删除路由 = 自动清理
- ✅ **零额外配置**：不需要权限配置文件，不需要手动写 SQL
- ✅ **声明式 API**：一行 `WithMeta()` 完成权限组声明
- ✅ **启动时同步**：每次启动自动扫描路由变更并同步数据库
- ✅ **精确控制**：权限细粒度到 API 路径 + HTTP 方法
- ✅ **默认安全**：未声明权限组的路由需手动标记为 `Public()`
- ✅ **开发友好**：子路由自动继承父权限组，也可覆盖

### 完整示例

```go
package v1

import (
    "gin-admin/internal/handler/v1"
    "gin-admin/internal/middleware"
    "gin-admin/internal/routegroup"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
    // 公开路由（不需要权限）
    authGroup := routegroup.WrapGroup(api.Group("/auth")).Public()
    {
        authGroup.POST("/login", handler.Login)
        authGroup.POST("/register", handler.Register)
    }

    // 用户管理（需要 user:manage 权限）
    userGroup := routegroup.WrapGroup(api.Group("/users")).
        WithMeta("user:manage", "用户管理")
    userGroup.Use(middleware.JWT())
    {
        userGroup.GET("", handler.GetUsers)           // 自动注册：GET /api/v1/users
        userGroup.POST("", handler.CreateUser)        // 自动注册：POST /api/v1/users
        userGroup.GET("/:id", handler.GetUser)        // 自动注册：GET /api/v1/users/:id
        userGroup.PUT("/:id", handler.UpdateUser)     // 自动注册：PUT /api/v1/users/:id
        userGroup.DELETE("/:id", handler.DeleteUser)  // 自动注册：DELETE /api/v1/users/:id
    }

    // 角色管理（需要 role:manage 权限）
    roleGroup := routegroup.WrapGroup(api.Group("/roles")).
        WithMeta("role:manage", "角色管理")
    roleGroup.Use(middleware.JWT())
    {
        roleGroup.GET("", handler.GetRoles)      // 自动注册！
        roleGroup.POST("", handler.CreateRole)   // 自动注册！
        // ... 所有路由都会自动注册到权限系统
    }
}
```

**就这样！** 🎉 无需任何额外配置，启动应用后：
- 所有路由自动注册为资源
- 权限组自动创建并关联资源
- 超级管理员自动拥有所有权限
- 用 `admin / admin123` 即可登录使用

### 权限验证流程

1. 用户发起 API 请求（如：`GET /api/v1/users`）
2. JWT 中间件验证 Token 并提取用户 ID
3. Permission 中间件查询用户的角色列表
4. 查询角色绑定的资源列表（`User → Role → Resources`）
5. 检查请求的 API（路径 + 方法）是否在授权资源中
6. 返回验证结果（允许/拒绝）

### 路由变更自动同步

**添加新路由**：
```go
// 新增一个导出功能
userGroup.GET("/export", handler.ExportUsers)  // ← 启动时自动注册！
```

**删除路由**：
```go
// 注释或删除路由
// userGroup.DELETE("/:id", handler.DeleteUser)  // ← 启动时自动从数据库清理！
```

**修改权限组**：
```go
// 将用户查看功能拆分到单独的权限组
viewGroup := routegroup.WrapGroup(api.Group("/users")).
    WithMeta("user:view", "用户查看")  // ← 启动时自动更新！
viewGroup.Use(middleware.JWT())
{
    viewGroup.GET("", handler.GetUsers)
}
```

### 高级用法

#### 1. 子路由继承权限

```go
orderGroup := routegroup.WrapGroup(api.Group("/orders")).
    WithMeta("order:view", "订单查看")
{
    orderGroup.GET("", handler.ListOrders)
    
    // 子路由自动继承父权限组
    detailGroup := orderGroup.Group("/:id")
    {
        detailGroup.GET("", handler.GetOrder)  // 也属于 order:view
    }
}
```

#### 2. 子路由覆盖权限

```go
productGroup := routegroup.WrapGroup(api.Group("/products")).
    WithMeta("product:view", "产品查看")
{
    productGroup.GET("", handler.ListProducts)
    
    // 管理功能需要更高权限
    manageGroup := routegroup.WrapGroup(productGroup.Group("/")).
        WithMeta("product:manage", "产品管理")
    {
        manageGroup.POST("", handler.CreateProduct)
        manageGroup.DELETE("/:id", handler.DeleteProduct)
    }
}
```

### 与传统方案对比

| 对比项 | 传统 RBAC | 本框架（路由自动注册） |
|--------|-----------|------------------------|
| 添加新 API | 写代码 + 写 SQL + 重启 | 只写代码，启动自动同步 |
| 删除 API | 手动清理数据库 | 启动自动清理 |
| 权限配置 | 需要配置文件或 SQL | 代码即配置 |
| 维护成本 | 高（容易遗漏） | 低（自动化） |
| 学习成本 | 需要理解表结构 | 只需 `WithMeta()` |
| 错误风险 | 容易出错 | 几乎无风险 |

### 快速入门

详细的 RBAC 使用指南请查看：[📖 RBAC 快速开始](RBAC_QUICKSTART.md)

---

## 🎨 配套前端项目

本项目提供了完整的前端管理系统，开箱即用！

### [🌐 Gin Admin Web](https://github.com/the-yex/gin-admin-web)

**技术栈**：基于现代前端框架，完美适配后端 RBAC 权限设计

**核心特性**：
- ✅ **权限联动**：前端菜单和按钮权限自动根据后端 RBAC 权限组控制
- ✅ **开箱即用**：克隆即可运行，无需额外配置
- ✅ **完整示例**：包含用户管理、角色管理、权限管理等完整功能模块
- ✅ **响应式设计**：支持多种设备和屏幕尺寸

**快速开始**：
```bash
# 克隆前端项目
git clone https://github.com/the-yex/gin-admin-web.git
cd gin-admin-web

# 安装依赖并运行
npm install
npm run dev
```

前后端配合使用，即可获得完整的企业级后台管理系统！🚀

---

## 📚 API 文档

项目集成了 Swagger 自动生成的交互式 API 文档。

### 查看文档

1. 启动应用：`make run`
2. 访问 Swagger UI：http://localhost:8080/swagger/index.html

### 更新文档

```bash
# 修改代码后重新生成文档
make swagger_v1
```

### Swagger 注解示例

```go
// @Summary      用户登录
// @Description  使用用户名和密码登录
// @Tags         用户认证
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "登录信息"
// @Success      200 {object} response.Response{data=LoginResponse}
// @Failure      400 {object} response.Response
// @Router       /api/v1/auth/login [post]
func Login(c *gin.Context) {
    // ...
}
```

---

## 🌱 项目扩展指南

### 添加新的 API 接口

1. **创建路由处理器** （`internal/handler/v1/xxx.go`）

```go
package v1

import "github.com/gin-gonic/gin"

// @Summary 示例接口
// @Tags 示例模块
// @Router /api/v1/example [get]
func ExampleHandler(c *gin.Context) {
    // 处理逻辑
}
```

2. **实现业务逻辑** （`internal/logic/v1/xxx_logic.go`）

```go
package v1

type ExampleLogic struct{}

func (l *ExampleLogic) DoSomething() error {
    // 业务逻辑
    return nil
}
```

3. **注册路由** （`internal/routegroup/v1/routes.go`）

```go
v1Group := r.Group("/api/v1")
{
    v1Group.GET("/example", handler.ExampleHandler)
}
```

4. **生成文档**

```bash
make swagger_v1
```

### 添加新的数据模型

1. **定义模型** （`internal/model/xxx.go`）

```go
package model

type Example struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"type:varchar(100);not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

2. **注册迁移** （`internal/model/migrate.go`）

```go
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &Example{},
        // ... 其他模型
    )
}
```

### 添加新的中间件

```go
// internal/middleware/custom.go
package middleware

import "github.com/gin-gonic/gin"

func CustomMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 前置处理
        c.Next()
        // 后置处理
    }
}
```

---

## 🐳 Docker 部署

### 快速启动（Docker Compose）

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f app

# 停止服务
docker-compose down
```

### 单独构建镜像

```bash
# 构建镜像
docker build -t gin-admin:latest .

# 运行容器
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/app.yaml:/app/app.yaml \
  -v $(pwd)/logs:/app/logs \
  --name gin-admin \
  gin-admin:latest
```

---

## 🧪 测试

```bash
# 运行所有测试
make test

# 生成覆盖率报告
make test-coverage

# 查看覆盖率（会打开浏览器）
open coverage.html
```

---

## 🔄 热重载开发

安装 [Air](https://github.com/cosmtrek/air) 实现代码热重载：

```bash
# 安装 air
go install github.com/cosmtrek/air@latest

# 启动热重载
make dev
```

---

## 📦 项目重命名

快速将项目重命名为你自己的项目名：

```bash
make rename NEW_NAME=your-awesome-project
```

这会自动更新：

- ✅ `go.mod` 模块名
- ✅ 所有 Go 文件中的 import 路径
- ✅ `Makefile` 配置
- ✅ `docker-compose.yml`
- ✅ 文档文件

---

## 🤝 贡献指南

我们欢迎所有形式的贡献！无论是新功能、Bug 修复、文档改进还是建议。

### 如何贡献

1. **Fork** 本仓库
2. **创建**你的特性分支 (`git checkout -b feature/AmazingFeature`)
3. **提交**你的修改 (`git commit -m 'Add some AmazingFeature'`)
4. **推送**到分支 (`git push origin feature/AmazingFeature`)
5. **开启** Pull Request

### 代码规范

- 遵循 [Effective Go](https://go.dev/doc/effective_go) 编码规范
- 运行 `make check` 确保代码通过所有检查
- 为新功能添加单元测试
- 更新相关文档

---

## 📄 许可证

本项目采用 MIT 许可证 - 详情请查看 [LICENSE](LICENSE) 文件

---

## 🌟 Star History

如果这个项目对你有帮助，请给我们一个 ⭐️ ！

[![Star History Chart](https://api.star-history.com/svg?repos=the-yex/gin-admin&type=Date)](https://star-history.com/#the-yex/gin-admin&Date)

---

## 📧 联系方式

- 提交 Issue：[GitHub Issues](https://github.com/the-yex/gin-admin/issues)
- 项目主页：[https://github.com/the-yex/gin-admin](https://github.com/the-yex/gin-admin)

---

<div align="center">

**如果觉得有用，请点个 ⭐️ Star 支持一下！**

Made with ❤️ by [the-yex](https://github.com/the-yex)

</div>
