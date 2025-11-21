<div align="center">

# 🚀 Gin Admin

**Production-Ready Go Web Application Scaffold**

A feature-complete, ready-to-use enterprise-grade Gin framework backend template for rapidly building high-performance, secure, and reliable web applications

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Gin Version](https://img.shields.io/badge/Gin-1.9-00ADD8?style=flat&logo=go)](https://github.com/gin-gonic/gin)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/the-yex/gin-admin?style=social)](https://github.com/the-yex/gin-admin/stargazers)
[![GitHub Forks](https://img.shields.io/github/forks/the-yex/gin-admin?style=social)](https://github.com/the-yex/gin-admin/network/members)

English | [简体中文](README.md)

[Quick Start](#-quick-start) • [Core Features](#-core-features) • [Documentation](#-api-documentation) • [Contributing](#-contributing)

</div>

---

## 📖 Introduction

Gin Admin is a ready-to-use Go backend development template built on the [Gin](https://github.com/gin-gonic/gin) framework, integrating core functional modules required for enterprise-grade project development. Whether you're building RESTful APIs, microservices, or complete web application backends, this template saves you significant infrastructure setup time, allowing you to focus on business logic development.

### 🎯 Why Choose Gin Admin?

- ⚡ **Ready to Use**: Clone and run, no complex configuration needed
- 🏗️ **Best Practices**: Strictly follows Go project layout and coding standards
- 🔐 **Security First**: Complete RBAC permission system and JWT authentication
- 🤖 **Route = Permission**: Revolutionary auto-registration mechanism, add route = auto-manage permissions, zero extra configuration
- 🚢 **Production Ready**: Docker containerization, graceful shutdown, health checks all included
- 📚 **Well Documented**: Auto-generated Swagger API documentation
- 🛠️ **Developer Friendly**: Powerful Makefile toolchain and hot reload support

---

## ✨ Core Features

### 🏛️ Architecture Design

- **🎨 Clear Layered Architecture**
  - Handler (Route Layer) → Logic (Business Logic Layer) → Service (Service Layer) → Model (Data Layer)
  - Strict separation of concerns for easy testing and maintenance
  - Modular design supporting rapid expansion

- **⚙️ Flexible Configuration Management**
  - Powerful configuration system based on Viper
  - Supports YAML, JSON, environment variables, and more
  - Multi-environment configuration support (dev, test, production)

### 🔒 Security & Authentication

- **🛡️ Complete RBAC + Auto Route Registration**
  - 🚀 **Revolutionary Design**: Routes auto-register to permission system, auto-sync latest resources at startup
  - 🎯 **Zero Extra Config**: No manual permission table management, no SQL, no config files needed
  - 📝 **Declarative Permissions**: One-line `WithMeta()` declaration, system handles everything
  - 🔄 **Auto Sync**: Every startup scans route changes, auto-updates database for new/deleted routes
  - 🎨 **UI-Friendly Grouping**: Permission groups for frontend display, Resources for actual authorization
  - 🔐 **Authorization Path**: User → Role → Resource (API-level precise control)
  - 🛡️ **Default Deny**: Unauthorized resources automatically rejected
  - [📖 Learn more about RBAC](RBAC_QUICKSTART.md)

- **🔑 JWT Authentication**
  - Stateless authentication based on JWT
  - Automatic token refresh mechanism
  - Secure password encryption storage (bcrypt)

### 🧩 Middleware Ecosystem

Built-in **8 production-grade middleware**, ready to use:

| Middleware | Description |
|------------|-------------|
| 🔐 JWT Auth | JWT token validation and user identification |
| 🚦 CORS | Cross-Origin Resource Sharing configuration |
| 📝 Logger | Structured request logging |
| 🔄 Recovery | Graceful panic recovery and error handling |
| 🎫 Request ID | Generate unique tracking ID for each request |
| 🔐 Permission | RBAC permission validation |
| ⏱️ Rate Limit | Token Bucket based rate limiter |
| 📊 Metrics | Request metrics and monitoring |

### 💾 Database & Cache

- **🗄️ Database Support**
  - ORM based on GORM v2
  - Support for MySQL, PostgreSQL, SQLite, and other mainstream databases
  - Auto migration and model management
  - Optimized connection pool configuration

- **⚡ Redis Cache Integration**
  - Ready-to-use Redis client
  - Cache warming and expiration strategy support
  - Distributed lock implementation

### 📊 Logging & Monitoring

- **📋 Professional Logging System**
  - Structured logging based on Logrus
  - Multi-level logging support (Debug, Info, Warn, Error)
  - Automatic log file rotation (Lumberjack)
  - JSON format output for easy log collection

### 🚀 DevOps Support

- **🐳 Docker Containerization**
  - Multi-stage Dockerfile with minimal image size
  - Docker Compose one-click start complete environment
  - Includes MySQL and Redis service orchestration

- **🛠️ Powerful Makefile**
  - `make run` - Quick run application
  - `make build` - Build binary
  - `make build-all` - Cross-platform compilation (Linux/macOS/Windows)
  - `make swagger_v1` - Generate API documentation
  - `make test` - Run test suite
  - `make dev` - Hot reload development mode (requires air)
  - `make rename` - Quick project rename
  - [View all commands](#-using-makefile)

### 📚 Documentation

- **📖 Swagger API Documentation**
  - Auto-generated interactive API documentation
  - Online interface testing support
  - Access at: `http://localhost:8080/swagger/index.html`

---

## 🚀 Quick Start

### Method 1: Docker Compose (Recommended)

**Start complete environment in 30 seconds!**

```bash
# 1. Clone the project
git clone https://github.com/the-yex/gin-admin.git
cd gin-admin

# 2. Initialize configuration
make init-config

# 3. Start all services (App + MySQL + Redis)
make up

# 4. View logs
make logs
```

🎉 Visit http://localhost:8080/swagger/index.html to view API documentation!

### Method 2: Local Run

#### Prerequisites

- Go 1.20 or higher
- MySQL 5.7+ / PostgreSQL / SQLite
- Redis (optional, for caching)

#### Installation Steps

```bash
# 1. Clone the project
git clone https://github.com/the-yex/gin-admin.git
cd gin-admin

# 2. Install dependencies
go mod tidy

# 3. Initialize configuration file
make init-config

# 4. Edit configuration file (modify database connection, etc.)
vim app.yaml
```

Edit `app.yaml` configuration:

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
  expire: 86400  # 24 hours
```

```bash
# 5. Run the application
make run

# Or use Go command directly
go run main.go
```

#### 🧪 Test API

```bash
# Health check
curl http://localhost:8080/health

# View API documentation
open http://localhost:8080/swagger/index.html
```

---

## 📁 Project Structure

```
gin-admin/
├── 📄 main.go                 # Application entry point
├── 📄 Makefile                # Make command collection
├── 📄 Dockerfile              # Docker build file
├── 📄 docker-compose.yml      # Docker Compose orchestration
├── 📄 app.yaml                # Application configuration
│
├── 📂 internal/               # Private application code
│   ├── 📂 config/            # Configuration management
│   ├── 📂 core/              # Core initialization logic
│   ├── 📂 handler/           # HTTP handlers (route layer)
│   │   └── 📂 v1/            # API v1 version
│   ├── 📂 logic/             # Business logic layer
│   │   └── 📂 v1/            # v1 business logic
│   ├── 📂 middleware/        # Middleware
│   │   ├── 📄 jwt.go         # JWT authentication
│   │   ├── 📄 permission.go  # RBAC permission
│   │   ├── 📄 rate_limit.go  # Rate limiter
│   │   ├── 📄 cors.go        # CORS handling
│   │   ├── 📄 logger.go      # Request logging
│   │   ├── 📄 recovery.go    # Panic recovery
│   │   ├── 📄 request_id.go  # Request tracking
│   │   └── 📄 metrics.go     # Monitoring metrics
│   ├── 📂 model/             # Data models
│   │   ├── 📂 rbac/          # RBAC models
│   │   └── 📄 migrate.go     # Database migration
│   ├── 📂 service/           # Business service layer
│   ├── 📂 routegroup/        # Route grouping
│   └── 📂 types/             # Type definitions
│
├── 📂 pkg/                    # Reusable public libraries
│   ├── 📂 cache/             # Cache utilities (Redis)
│   ├── 📂 logger/            # Logging utilities
│   ├── 📂 orm/               # ORM configuration
│   ├── 📂 jwt/               # JWT utilities
│   ├── 📂 response/          # Unified response format
│   ├── 📂 validator/         # Parameter validation
│   ├── 📂 errcode/           # Error code definitions
│   ├── 📂 transaction/       # Transaction management
│   └── 📂 utils/             # Common utilities
│
├── 📂 docs/                   # Swagger API documentation
│   ├── 📄 v1_docs.go
│   ├── 📄 v1_swagger.json
│   └── 📄 v1_swagger.yaml
│
├── 📂 scripts/                # Script tools
├── 📂 logs/                   # Log file directory
└── 📂 build/                  # Build output directory
```

---

## 🔧 Using Makefile

The project provides rich Makefile commands to simplify development:

### 🏃 Run & Build

```bash
make run              # Run application
make build            # Build application (current platform)
make build-linux      # Build Linux version
make build-darwin     # Build macOS version
make build-windows    # Build Windows version
make build-all        # Build all platform versions
```

### 🧪 Test & Check

```bash
make test             # Run tests
make test-coverage    # Generate test coverage report
make lint             # Code style check
make fmt              # Format code
make vet              # Static analysis
make check            # Run all checks (fmt + vet + lint)
```

### 📖 Documentation

```bash
make swagger_v1       # Generate Swagger API documentation
```

### 🐳 Docker

```bash
make docker-build     # Build Docker image
make docker-run       # Run Docker container
make docker-stop      # Stop Docker container
make up               # Start Docker Compose services
make down             # Stop Docker Compose services
make logs             # View service logs
```

### 🛠️ Tools

```bash
make init-config      # Initialize configuration file
make rename NEW_NAME=your-project  # Rename project
make dev              # Hot reload development mode (requires air)
make install          # Install dependencies
make clean            # Clean build files
make help             # View all available commands
```

---

## 🎯 RBAC Permission System + Auto Route Registration

This project implements a **revolutionary RBAC permission management system** with the highlight being the **auto route registration mechanism**:

> 💡 **Core Innovation**: Adding routes completes permission configuration, auto-syncs at startup, no manual permission table management!

### 🌟 Why Revolutionary?

#### Traditional RBAC Pain Points ❌

```sql
+-- 😓 Every new API requires a bunch of SQL
+INSERT INTO permissions (code, name) VALUES ('user:create', 'Create User');
+INSERT INTO resources (path, method, permission_id) VALUES ('/api/v1/users', 'POST', 1);
+INSERT INTO role_permissions (role_id, permission_id) VALUES (1, 1);
+-- High maintenance cost, easy to miss, error-prone
+```
+
+#### Our Framework's Solution ✅
+
+```go
+// 😎 Just one line, everything else is automatic!
+userGroup := routegroup.WrapGroup(api.Group("/users")).
+    WithMeta("user:manage", "User Management")
+{
+    userGroup.GET("", handler.GetUsers)      // Auto-registered!
+    userGroup.POST("", handler.CreateUser)   // Auto-registered!
+    userGroup.PUT("/:id", handler.UpdateUser) // Auto-registered!
+}
+```
+
+**Magic That Happens At Startup** ✨:
+1. 📡 Scans all route definitions
+2. 🔍 Identifies permission group declarations (`WithMeta()`)
+3. 📝 Auto-creates/updates permission groups in database
+4. 🔗 Auto-associates route resources to permission groups
+5. 🔐 Auto-binds resources to super admin role
+6. 🗑️ Auto-cleans deleted route resources
+
+### Architecture Design
+
+```
+User ──→ Role ──→ Resource (API)  [Actual Authorization Path]
+                      ↓
+                 Permission       [UI Grouping Only]
+```
+
+**Design Philosophy**:
+- **Actual Authorization**: Roles directly bind resources (API path + HTTP method)
+- **UI Grouping**: Permission only for frontend page logical grouping and display
+- **Auto Sync**: Route changes auto-reflect in permission system
+
+### Core Features
+
+- ✅ **Route = Permission**: Add route = auto-register resource, delete route = auto-cleanup
+- ✅ **Zero Extra Config**: No permission config files, no manual SQL
+- ✅ **Declarative API**: One-line `WithMeta()` completes permission group declaration
+- ✅ **Startup Sync**: Every startup auto-scans route changes and syncs database
+- ✅ **Precise Control**: Permission granularity to API path + HTTP method
+- ✅ **Default Security**: Routes without permission group declaration need manual `Public()` marking
+- ✅ **Developer Friendly**: Sub-routes auto-inherit parent permission groups, can override
+
+### Complete Example
+
+```go
+package v1
+
+import (
+    "gin-admin/internal/handler/v1"
+    "gin-admin/internal/middleware"
+    "gin-admin/internal/routegroup"
+    "github.com/gin-gonic/gin"
+)
+
+func RegisterRoutes(api *gin.RouterGroup) {
+    // Public routes (no permission required)
+    authGroup := routegroup.WrapGroup(api.Group("/auth")).Public()
+    {
+        authGroup.POST("/login", handler.Login)
+        authGroup.POST("/register", handler.Register)
+    }
+
+    // User management (requires user:manage permission)
+    userGroup := routegroup.WrapGroup(api.Group("/users")).
+        WithMeta("user:manage", "User Management")
+    userGroup.Use(middleware.JWT())
+    {
+        userGroup.GET("", handler.GetUsers)           // Auto-registered: GET /api/v1/users
+        userGroup.POST("", handler.CreateUser)        // Auto-registered: POST /api/v1/users
+        userGroup.GET("/:id", handler.GetUser)        // Auto-registered: GET /api/v1/users/:id
+        userGroup.PUT("/:id", handler.UpdateUser)     // Auto-registered: PUT /api/v1/users/:id
+        userGroup.DELETE("/:id", handler.DeleteUser)  // Auto-registered: DELETE /api/v1/users/:id
+    }
+
+    // Role management (requires role:manage permission)
+    roleGroup := routegroup.WrapGroup(api.Group("/roles")).
+        WithMeta("role:manage", "Role Management")
+    roleGroup.Use(middleware.JWT())
+    {
+        roleGroup.GET("", handler.GetRoles)      // Auto-registered!
+        roleGroup.POST("", handler.CreateRole)   // Auto-registered!
+        // ... all routes auto-register to permission system
+    }
+}
+```
+
+**That's it!** 🎉 No extra configuration needed, after starting the application:
+- All routes auto-register as resources
+- Permission groups auto-create and associate resources
+- Super admin auto-gets all permissions
+- Use `admin / admin123` to login
+
+### Permission Verification Flow
+
+1. User initiates API request (e.g., `GET /api/v1/users`)
+2. JWT middleware validates token and extracts user ID
+3. Permission middleware queries user's role list
+4. Queries resources bound to roles (`User → Role → Resources`)
+5. Checks if requested API (path + method) is in authorized resources
+6. Returns verification result (allow/deny)
+
+### Auto-Sync on Route Changes
+
+**Adding New Routes**:
+```go
+// Add an export feature
+userGroup.GET("/export", handler.ExportUsers)  // ← Auto-registered at startup!
+```
+
+**Deleting Routes**:
+```go
+// Comment or delete route
+// userGroup.DELETE("/:id", handler.DeleteUser)  // ← Auto-cleaned from database at startup!
+```
+
+**Modifying Permission Groups**:
+```go
+// Split user view functionality to separate permission group
+viewGroup := routegroup.WrapGroup(api.Group("/users")).
+    WithMeta("user:view", "View Users")  // ← Auto-updated at startup!
+viewGroup.Use(middleware.JWT())
+{
+    viewGroup.GET("", handler.GetUsers)
+}
+```
+
+### Advanced Usage
+
+#### 1. Sub-routes Inherit Permissions
+
+```go
+orderGroup := routegroup.WrapGroup(api.Group("/orders")).
+    WithMeta("order:view", "View Orders")
+{
+    orderGroup.GET("", handler.ListOrders)
+    
+    // Sub-routes auto-inherit parent permission group
+    detailGroup := orderGroup.Group("/:id")
+    {
+        detailGroup.GET("", handler.GetOrder)  // Also belongs to order:view
+    }
+}
+```
+
+#### 2. Sub-routes Override Permissions
+
+```go
+productGroup := routegroup.WrapGroup(api.Group("/products")).
+    WithMeta("product:view", "View Products")
+{
+    productGroup.GET("", handler.ListProducts)
+    
+    // Management features require higher permissions
+    manageGroup := routegroup.WrapGroup(productGroup.Group("/")).
+        WithMeta("product:manage", "Manage Products")
+    {
+        manageGroup.POST("", handler.CreateProduct)
+        manageGroup.DELETE("/:id", handler.DeleteProduct)
+    }
+}
+```
+
+### Comparison with Traditional Approach
+
+| Aspect | Traditional RBAC | Our Framework (Auto-Registration) |
+|--------|------------------|-----------------------------------|
+| Add New API | Write code + SQL + restart | Only write code, auto-syncs at startup |
+| Delete API | Manually clean database | Auto-cleans at startup |
+| Permission Config | Need config files or SQL | Code is configuration |
+| Maintenance Cost | High (easy to miss) | Low (automated) |
+| Learning Curve | Need to understand table structure | Just `WithMeta()` |
+| Error Risk | Error-prone | Almost risk-free |
+
+### Quick Guide
+
+For detailed RBAC usage guide, see: [📖 RBAC Quick Start](RBAC_QUICKSTART.md)

---

## 📚 API Documentation

The project integrates auto-generated interactive Swagger API documentation.

### View Documentation

1. Start application: `make run`
2. Visit Swagger UI: http://localhost:8080/swagger/index.html

### Update Documentation

```bash
# Regenerate documentation after code changes
make swagger_v1
```

### Swagger Annotation Example

```go
// @Summary      User login
// @Description  Login with username and password
// @Tags         User Auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login information"
// @Success      200 {object} response.Response{data=LoginResponse}
// @Failure      400 {object} response.Response
// @Router       /api/v1/auth/login [post]
func Login(c *gin.Context) {
    // ...
}
```

---

## 🌱 Project Extension Guide

### Adding New API Endpoints

1. **Create route handler** (`internal/handler/v1/xxx.go`)
```go
package v1

import "github.com/gin-gonic/gin"

// @Summary Example endpoint
// @Tags Example Module
// @Router /api/v1/example [get]
func ExampleHandler(c *gin.Context) {
    // Handler logic
}
```

2. **Implement business logic** (`internal/logic/v1/xxx_logic.go`)
```go
package v1

type ExampleLogic struct{}

func (l *ExampleLogic) DoSomething() error {
    // Business logic
    return nil
}
```

3. **Register route** (`internal/routegroup/v1/routes.go`)
```go
v1Group := r.Group("/api/v1")
{
    v1Group.GET("/example", handler.ExampleHandler)
}
```

4. **Generate documentation**
```bash
make swagger_v1
```

### Adding New Data Models

1. **Define model** (`internal/model/xxx.go`)
```go
package model

type Example struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"type:varchar(100);not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

2. **Register migration** (`internal/model/migrate.go`)
```go
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &Example{},
        // ... other models
    )
}
```

### Adding New Middleware

```go
// internal/middleware/custom.go
package middleware

import "github.com/gin-gonic/gin"

func CustomMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Pre-processing
        c.Next()
        // Post-processing
    }
}
```

---

## 🐳 Docker Deployment

### Quick Start (Docker Compose)

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop services
docker-compose down
```

### Build Image Separately

```bash
# Build image
docker build -t gin-admin:latest .

# Run container
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/app.yaml:/app/app.yaml \
  -v $(pwd)/logs:/app/logs \
  --name gin-admin \
  gin-admin:latest
```

---

## 🧪 Testing

```bash
# Run all tests
make test

# Generate coverage report
make test-coverage

# View coverage (opens browser)
open coverage.html
```

---

## 🔄 Hot Reload Development

Install [Air](https://github.com/cosmtrek/air) for code hot reloading:

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Start hot reload
make dev
```

---

## 📦 Project Rename

Quickly rename project to your own project name:

```bash
make rename NEW_NAME=your-awesome-project
```

This will automatically update:
- ✅ `go.mod` module name
- ✅ All import paths in Go files
- ✅ `Makefile` configuration
- ✅ `docker-compose.yml`
- ✅ Documentation files

---

## 🤝 Contributing

We welcome all forms of contribution! Whether it's new features, bug fixes, documentation improvements, or suggestions.

### How to Contribute

1. **Fork** this repository
2. **Create** your feature branch (`git checkout -b feature/AmazingFeature`)
3. **Commit** your changes (`git commit -m 'Add some AmazingFeature'`)
4. **Push** to the branch (`git push origin feature/AmazingFeature`)
5. **Open** a Pull Request

### Code Standards

- Follow [Effective Go](https://go.dev/doc/effective_go) coding standards
- Run `make check` to ensure code passes all checks
- Add unit tests for new features
- Update related documentation

---

## 📄 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) file for details

---

## 🌟 Star History

If this project helps you, please give us a ⭐️!

[![Star History Chart](https://api.star-history.com/svg?repos=the-yex/gin-admin&type=Date)](https://star-history.com/#the-yex/gin-admin&Date)

---

## 📧 Contact

- Submit Issues: [GitHub Issues](https://github.com/the-yex/gin-admin/issues)
- Project Homepage: [https://github.com/the-yex/gin-admin](https://github.com/the-yex/gin-admin)

---

<div align="center">

**If you find this useful, please ⭐️ Star to support!**

Made with ❤️ by [the-yex](https://github.com/the-yex)

</div>
