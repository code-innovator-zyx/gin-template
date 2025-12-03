# Gin-Admin

<div align="center">

[English](./README.md) | [简体中文](./README_zh.md)

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue.svg)](https://golang.org)
[![Gin Version](https://img.shields.io/badge/Gin-1.9%2B-green.svg)](https://gin-gonic.com)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

**An enterprise-grade Go backend framework with automatic RBAC permission management**

[Features](#-features) • [Quick Start](#-quick-start) • [Documentation](#-documentation) • [Tech Stack](#-tech-stack) • [Contributing](#-contributing)

</div>

---

## ✨ Features

### 🎯 Core Features

- **🔐 JWT Authentication** - Dual-token mechanism with token rotation and session management
- **🚀 RBAC Auto-Initialization** - Revolutionary code-as-config permission system (no manual resource management!)
- **💾 Unified Cache Layer** - Support for Redis and in-memory backends with anti-penetration/breakdown/avalanche strategies
- **📦 Generic Repository Pattern** - Type-safe CRUD operations with flexible query options
- **🔄 RESTful API** - Standard API design with Swagger documentation
- **🐳 Docker Support** - One-command deployment with Docker Compose

### 🛡️ Security Features

- **Token Rotation** - Auto-refresh with refresh token reuse detection
- **Permission Caching** - High-performance permission checks with singleflight
- **Session Management** - Multi-device login support
- **SQL Injection Protection** - GORM parameterized queries

### 🎨 Developer Experience

- **Clean Architecture** - Handler → Logic → Service → Repository layering
- **Auto Swagger Docs** - Auto-generated API documentation
- **Hot Reload** - Air support for development
- **Comprehensive Tests** - Unit and integration tests

---

## 📖 Documentation

### 📚 Technical Documentation

- [JWT Authentication System](./docs/jwt.md) - Token rotation, session management, security mechanisms
- [Cache System](./docs/cache.md) - Redis/Memory adapters, anti-penetration strategies
- [Repository Pattern](./docs/repository.md) - Generic design, query options, pagination
- [**RBAC Auto-Initialization**](./docs/rbac-auto-init.md) - ⭐ **The killer feature! Automatic permission management**

### 🚀 Getting Started

- [Quick Start Guide](#-quick-start)
- [Configuration Guide](#%EF%B8%8F-configuration)
- [Deployment Guide](#-deployment)

---

## 🚀 Quick Start

### Prerequisites

- **Go** 1.21+
- **MySQL** 8.0+ (or compatible database)
- **Redis** 7.0+ (optional, falls back to memory cache)

### Installation

```bash
# 1. Clone the repository
git clone https://github.com/the-yex/gin-admin.git
cd gin-admin

# 2. Install dependencies
go mod download

# 3. Copy and configure
cp config/app.yaml.template config/app.yaml
# Edit config/app.yaml with your database and Redis settings

# 4. Run database migrations
go run cmd/migrate/main.go

# 5. Start the server
go run cmd/server/main.go
```

The server will start at http://localhost:8080

### Access Swagger UI

Open your browser and navigate to:
- **Swagger v1**: http://localhost:8080/swagger/v1/index.html

### Default Admin Account

```
Username: admin
Password: admin123
```

> ⚠️ **Security Warning**: Change default password in production!

---

## 🛠️ Tech Stack

### Core Framework

- **[Gin](https://gin-gonic.com/)** - High-performance HTTP web framework
- **[GORM](https://gorm.io/)** - ORM library with generic repository support
- **[JWT-Go](https://github.com/golang-jwt/jwt)** - JSON Web Token implementation
- **[Viper](https://github.com/spf13/viper)** - Configuration management

### Database & Cache

- **MySQL** - Primary database
- **Redis** - Distributed cache (optional)
- **Memory Cache** - Built-in fallback cache

### Development Tools

- **[Swagger](https://swagger.io/)** - API documentation
- **[Air](https://github.com/cosmtrek/air)** - Live reload
- **Docker** - Containerization

---

## 📁 Project Structure

```
gin-admin/
├── cmd/                    # Application entry points
│   ├── server/            # Main server
│   └── migrate/           # Database migration tool
├── config/                # Configuration files
│   └── app.yaml.template  # Configuration template
├── docs/                  # Technical documentation
│   ├── jwt.md            # JWT authentication docs
│   ├── cache.md          # Cache system docs
│   ├── repository.md     # Repository pattern docs
│   └── rbac-auto-init.md # RBAC auto-initialization docs
├── internal/              # Private application code
│   ├── handler/          # HTTP handlers (routes)
│   ├── logic/            # Business logic layer
│   ├── services/         # Service layer (external calls)
│   ├── model/            # Data models
│   ├── middleware/       # HTTP middlewares
│   └── routegroup/       # 🌟 Auto RBAC route wrapper
├── pkg/                   # Public reusable packages
│   ├── components/       # Core components (JWT, etc.)
│   ├── cache/            # Cache abstraction layer
│   ├── interface/        # Generic repository interface
│   └── utils/            # Utility functions
└── docker/                # Docker configurations
```

---

## ⚙️ Configuration

Configuration file: `config/app.yaml`

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
  access_token_expire: 600s   # 10 minutes
  refresh_token_expire: 168h  # 7 days

cache:
  host: localhost
  port: 6379
  password: ""
  db: 0

rbac:
  enable_auto_init: true  # 🌟 Enable automatic RBAC initialization
  admin_user:
    username: admin
    password: admin123
```

See [config/app.yaml.template](./config/app.yaml.template) for full configuration options.

---

## 🌟 RBAC Auto-Initialization

### The Problem with Traditional RBAC

❌ Manual SQL scripts for resources  
❌ Dual maintenance (code + database)  
❌ Risk of inconsistency  
❌ Difficult to collaborate

### Our Solution: Code-as-Config

✅ **Declare permissions in code**

```go
// Declare permission group
userGroup := api.Group("/users").WithMeta("user:manage", "User Management")
userGroup.Use(middleware.JWT(ctx), middleware.PermissionMiddleware(ctx))
{
    // Declare resource permissions
    userGroup.GET("", handler).WithMeta("list", "List Users")
    userGroup.POST("", handler).WithMeta("add", "Create User")
    userGroup.DELETE("/:id", handler).WithMeta("delete", "Delete User")
}
```

✅ **Auto-sync on startup**
- Automatically extract routes and metadata
- Sync resources to database (create/update/delete)
- Create default admin role and user
- **Idempotent** - safe to run multiple times

**Result**: Resources always match your code, zero manual maintenance!

📖 [Read Full RBAC Documentation](./docs/rbac-auto-init.md)

---

## 🐳 Deployment

### Docker Compose (Recommended)

```bash
# Start all services (app + MySQL + Redis)
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop services
docker-compose down
```

### Manual Deployment

```bash
# Build binary
go build -o bin/server cmd/server/main.go

# Run
./bin/server
```

---

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/components/jwt/...
```

---

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### How to Contribute

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

---

## 🌟 Star History

If you find this project helpful, please give it a star! ⭐

<div align="center">

**Made with ❤️ by the Gin-Admin Team**

[Report Bug](https://github.com/the-yex/gin-admin/issues) • [Request Feature](https://github.com/the-yex/gin-admin/issues) • [Join Discussion](https://github.com/the-yex/gin-admin/discussions)

</div>
