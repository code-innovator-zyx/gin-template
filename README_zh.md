# Gin-Admin

<div align="center">

[English](./README.md) | [简体中文](./README_zh.md)

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue.svg)](https://golang.org)
[![Gin Version](https://img.shields.io/badge/Gin-1.9%2B-green.svg)](https://gin-gonic.com)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

**企业级 Go 后端框架，支持 RBAC 权限自动化管理**

[核心特性](#-核心特性) • [快速开始](#-快速开始) • [技术文档](#-技术文档) • [技术栈](#-技术栈) • [贡献指南](#-贡献指南)

</div>

---

## ✨ 核心特性

### 🎯 功能特性

- **🔐 JWT 认证** - 双 Token 机制，支持 Token Rotation 和多设备登录
- **🚀 RBAC 自动化初始化** - 革命性的代码即配置权限系统（无需手动维护资源！）
- **💾 统一缓存层** - 支持 Redis/内存双后端，内置防穿透/防击穿/防雪崩策略
- **📦 泛型 Repository** - 类型安全的 CRUD 操作，灵活的查询选项
- **🔄 RESTful API** - 标准化 API 设计，自动生成 Swagger 文档
- **🐳 Docker 支持** - 一键部署，支持 Docker Compose

### 🛡️ 安全特性

- **Token Rotation** - 自动刷新 Token，检测重用攻击
- **权限缓存** - 高性能权限校验，集成 Singleflight
- **会话管理** - 支持多设备登录和会话撤销
- **SQL 注入防护** - GORM 参数化查询

### 🎨 开发体验

- **整洁架构** - Handler → Logic → Service → Repository 分层设计
- **自动 Swagger 文档** - API 文档自动生成
- **热重载** - Air 支持开发环境热重载
- **完善测试** - 单元测试和集成测试覆盖

---

## 📖 技术文档

### 📚 核心技术文档

- [JWT 认证系统](./docs/jwt.md) - Token Rotation、Session 管理、安全机制
- [Cache 缓存系统](./docs/cache.md) - Redis/内存适配器、防穿透策略
- [Repository 数据访问层](./docs/repository.md) - 泛型设计、查询选项、分页
- [**RBAC 自动化权限初始化**](./docs/rbac-auto-init.md) - ⭐ **项目最大亮点！自动权限管理**

### 🚀 入门指南

- [快速开始](#-快速开始)
- [配置说明](#%EF%B8%8F-配置说明)
- [部署指南](#-部署)

---

## 🚀 快速开始

### 环境要求

- **Go** 1.21+
- **MySQL** 8.0+（或兼容数据库）
- **Redis** 7.0+（可选，未配置时使用内存缓存）

### 安装步骤

```bash
# 1. 克隆项目
git clone https://github.com/the-yex/gin-admin.git
cd gin-admin

# 2. 安装依赖
go mod download

# 3. 复制配置文件
cp config/app.yaml.template config/app.yaml
# 编辑 config/app.yaml 配置数据库和 Redis

# 4. 运行数据库迁移
go run cmd/migrate/main.go

# 5. 启动服务
go run cmd/server/main.go
```

服务将在 http://localhost:8080 启动

### 访问 Swagger 文档

在浏览器中打开：
- **Swagger v1**: http://localhost:8080/swagger/v1/index.html

### 默认管理员账号

```
用户名: admin
密码: admin123
```

> ⚠️ **安全提示**：生产环境必须修改默认密码！

---

## 🛠️ 技术栈

### 核心框架

- **[Gin](https://gin-gonic.com/)** - 高性能 HTTP Web 框架
- **[GORM](https://gorm.io/)** - 功能强大的 ORM 库，支持泛型
- **[JWT-Go](https://github.com/golang-jwt/jwt)** - JWT Token 实现
- **[Viper](https://github.com/spf13/viper)** - 配置管理

### 数据库 & 缓存

- **MySQL** - 主数据库
- **Redis** - 分布式缓存（可选）
- **Memory Cache** - 内置内存缓存

### 开发工具

- **[Swagger](https://swagger.io/)** - API 文档生成
- **[Air](https://github.com/cosmtrek/air)** - 热重载
- **Docker** - 容器化部署

---

## 📁 项目结构

```
gin-admin/
├── cmd/                    # 应用程序入口
│   ├── server/            # 主服务器
│   └── migrate/           # 数据库迁移工具
├── config/                # 配置文件
│   └── app.yaml.template  # 配置模板
├── docs/                  # 技术文档
│   ├── jwt.md            # JWT 认证文档
│   ├── cache.md          # 缓存系统文档
│   ├── repository.md     # Repository 文档
│   └── rbac-auto-init.md # RBAC 自动化初始化文档
├── internal/              # 私有应用代码
│   ├── handler/          # HTTP 处理器（路由）
│   ├── logic/            # 业务逻辑层
│   ├── services/         # 服务层（外部调用）
│   ├── model/            # 数据模型
│   ├── middleware/       # HTTP 中间件
│   └── routegroup/       # 🌟 RBAC 自动化路由包装器
├── pkg/                   # 公共可复用包
│   ├── components/       # 核心组件（JWT 等）
│   ├── cache/            # 缓存抽象层
│   ├── interface/        # 泛型 Repository 接口
│   └── utils/            # 工具函数
└── docker/                # Docker 配置
```

---

## ⚙️ 配置说明

配置文件：`config/app.yaml`

```yaml
app:
  name: gin-admin
  port: 8080
  mode: dev  # dev | test | prod

database:
  driver: mysql
  host: localhost
  port: 3306
  username: root
  password: your_password
  database: gin_admin

jwt:
  secret: "your-secret-key-32-chars-minimum"
  access_token_expire: 600s   # 10 分钟
  refresh_token_expire: 168h  # 7 天

cache:
  host: localhost
  port: 6379
  password: ""
  db: 0

rbac:
  enable_auto_init: true  # 🌟 启用 RBAC 自动初始化
  admin_user:
    username: admin
    password: admin123
```

完整配置选项请参考 [config/app.yaml.template](./config/app.yaml.template)。

---

## 🌟 RBAC 自动化权限初始化

### 传统 RBAC 的痛点

❌ 需要手动编写 SQL 脚本插入资源  
❌ 代码和数据库双重维护  
❌ 容易不一致  
❌ 团队协作困难

### 我们的解决方案：代码即配置

✅ **在代码中声明权限**

```go
// 声明权限组
userGroup := api.Group("/users").WithMeta("user:manage", "用户管理")
userGroup.Use(middleware.JWT(ctx), middleware.PermissionMiddleware(ctx))
{
    // 声明资源权限
    userGroup.GET("", handler).WithMeta("list", "查询用户列表")
    userGroup.POST("", handler).WithMeta("add", "创建用户")
    userGroup.DELETE("/:id", handler).WithMeta("delete", "删除用户")
}
```

✅ **应用启动自动同步**
- 自动提取路由和元数据
- 同步资源到数据库（新增/更新/删除）
- 创建默认管理员角色和用户
- **幂等操作** - 重复运行安全

**结果**：资源始终和代码一致，零手动维护！

📖 [查看完整 RBAC 文档](./docs/rbac-auto-init.md)

---

## 🐳 部署

### Docker Compose 部署（推荐）

```bash
# 启动所有服务（应用 + MySQL + Redis）
docker-compose up -d

# 查看日志
docker-compose logs -f app

# 停止服务
docker-compose down
```

### 手动部署

```bash
# 编译二进制文件
go build -o bin/server cmd/server/main.go

# 运行
./bin/server
```

---

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 带覆盖率运行
go test -cover ./...

# 运行特定包的测试
go test ./pkg/components/jwt/...
```

---

## 🤝 贡献指南

我们欢迎任何形式的贡献！请查看 [贡献指南](CONTRIBUTING.md) 了解详情。

### 如何贡献

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m '添加某个很棒的特性'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

---

## 📄 开源协议

本项目采用 Apache License 2.0 协议 - 详见 [LICENSE](LICENSE) 文件

---

## 🌟 Star 历史

如果这个项目对你有帮助，请给个 Star! ⭐

<div align="center">

**Made with ❤️ by the-yex**

[报告 Bug](https://github.com/the-yex/gin-admin/issues) • [请求新功能](https://github.com/the-yex/gin-admin/issues) • [加入讨论](https://github.com/the-yex/gin-admin/discussions)

</div>
