<div align="center">

# 🚀 Gin Enterprise Template

### Enterprise-Grade Go Web Development Template

*Modern, high-performance, production-ready web application template based on Gin framework*

[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)](https://go.dev)
[![Gin Version](https://img.shields.io/badge/Gin-1.9+-00ADD8?style=flat&logo=go)](https://github.com/gin-gonic/gin)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/code-innovator-zyx/gin-template/pulls)

English | [简体中文](./README.md)

</div>

---

## ✨ Key Features

<table>
<tr>
<td width="50%">

### 🏗️ Enterprise Architecture
- ✅ **Clean Layered Architecture** - Handler/Logic/Service/Model
- ✅ **Complete RBAC** - User-Role-Permission-Resource
- ✅ **Redis Cache Optimization** - 96% permission check speedup
- ✅ **Transaction Support** - Complete transaction handling

</td>
<td width="50%">

### ⚡ Development Efficiency
- ✅ **One-Click Rename** - Quick project setup
- ✅ **Swagger Docs** - Auto-generated API documentation
- ✅ **Hot Reload** - Air hot reload support
- ✅ **Docker Support** - One-click environment setup

</td>
</tr>
<tr>
<td width="50%">

### 🛡️ Production Ready
- ✅ **Health Check** - DB/Redis status monitoring
- ✅ **Graceful Shutdown** - Zero-downtime deployment
- ✅ **Request Tracing** - Complete request tracking
- ✅ **Panic Recovery** - Automatic fault recovery

</td>
<td width="50%">

### 🔧 Developer Tools
- ✅ **Complete Makefile** - 20+ useful commands
- ✅ **Parameter Validation** - Auto validation with i18n
- ✅ **Unit Tests** - Test examples and coverage
- ✅ **Environment Variables** - Environment variable support

</td>
</tr>
</table>

---

## 📚 Documentation

<table>
<tr>
<td align="center" width="25%">
<a href="./docs/QUICK_START.md">
<img src="https://img.icons8.com/color/96/000000/rocket.png" width="48" height="48" alt="Quick Start"/>
<br />
<b>Quick Start</b>
</a>
<br />
<sub>5-minute guide</sub>
</td>
<td align="center" width="25%">
<a href="./docs/RENAME_GUIDE.md">
<img src="https://img.icons8.com/color/96/000000/edit.png" width="48" height="48" alt="Rename"/>
<br />
<b>Rename Guide</b>
</a>
<br />
<sub>One-click rename</sub>
</td>
<td align="center" width="25%">
<a href="./docs/OPTIMIZATION_REPORT.md">
<img src="https://img.icons8.com/color/96/000000/document.png" width="48" height="48" alt="Report"/>
<br />
<b>Optimization Report</b>
</a>
<br />
<sub>Technical details</sub>
</td>
<td align="center" width="25%">
<a href="./docs/CHANGELOG.md">
<img src="https://img.icons8.com/color/96/000000/clock.png" width="48" height="48" alt="Changelog"/>
<br />
<b>Changelog</b>
</a>
<br />
<sub>Version history</sub>
</td>
</tr>
</table>

---

## 🚀 Quick Start

### Option 1: Local Development

```bash
# 1. Clone the project
git clone https://github.com/code-innovator-zyx/gin-template.git
cd gin-template

# 2. Rename project (recommended)
make rename NEW_NAME=my-awesome-api

# 3. Install dependencies
go mod tidy

# 4. Initialize config
make init-config

# 5. Run
make run
```

### Option 2: Docker Compose (Recommended)

```bash
# One-click start complete environment (MySQL + Redis included)
docker-compose up -d

# View logs
docker-compose logs -f

# Access
open http://localhost:8080/api/v1/health
```

**🎉 That's it!**

---

## 💻 Core Features

### 1. Complete RBAC Permission System (New Architecture)

```
User → Role → Resource  [Actual Authorization Path]
 ↓      ↓         ↓
Alice  Admin  GET /api/v1/users
 Bob   Editor  POST /api/v1/posts
                  ↓
           Permission [UI Grouping Only]
```

**New Architecture Features:**
- 🎯 **Direct Authorization** - Roles bind resources directly, faster verification
- 🎨 **UI Friendly** - Permission groups for frontend display
- 🔐 **Security First** - Default deny, explicit grant
- ⚡ **High Performance** - Multiple cache options, 2ms permission check
- 🎯 **Fine-grained** - Precise to API path + HTTP method
- 🔄 **Dynamic** - Runtime permission adjustment

### 2. One-Click Rename

```bash
make rename NEW_NAME=blog-api
```

Automatically updates:
- ✅ go.mod module name
- ✅ All import paths
- ✅ Makefile config
- ✅ Docker Compose config
- ✅ All documentation

> 📖 Details: [Rename Guide](./docs/RENAME_GUIDE.md)

### 3. Middleware Ecosystem

| Middleware | Function | Description |
|-----------|----------|-------------|
| Recovery | Panic Recovery | Auto capture and log panics |
| RequestID | Request Tracing | Generate unique ID for each request |
| Logger | Logging | Structured logs with timing and status |
| JWT | Authentication | JWT-based user authentication |
| Permission | Authorization | RBAC permission check (cached) |
| CORS | Cross-Origin | Configurable CORS policy |

---

## 🛠️ Tech Stack

<table>
<tr>
<td width="50%">

**Core**
- [Gin](https://github.com/gin-gonic/gin) - Web Framework
- [GORM](https://gorm.io) - ORM
- [Viper](https://github.com/spf13/viper) - Configuration
- [Logrus](https://github.com/sirupsen/logrus) - Logging

</td>
<td width="50%">

**Middleware/Tools**
- [JWT](https://github.com/golang-jwt/jwt) - Authentication
- [Redis](https://github.com/go-redis/redis) - Cache
- [Swagger](https://github.com/swaggo/swag) - Documentation
- [Validator](https://github.com/go-playground/validator) - Validation

</td>
</tr>
</table>

---

## 📊 Performance

| Metric | Without Cache | With Redis Cache | Improvement |
|--------|---------------|------------------|-------------|
| Permission Check | ~50ms | ~2ms | **96%** ⬆️ |
| Concurrent Requests | 1000 req/s | 5000+ req/s | **400%** ⬆️ |
| Database Queries | 3 per request | 0 (cache hit) | **100%** ⬇️ |

---

## 📁 Project Structure

```
gin-template/
├── 📂 internal/          # Internal packages (not exported)
│   ├── config/          # Configuration
│   ├── core/            # Core components
│   ├── handler/         # HTTP handlers
│   ├── logic/           # Business logic
│   ├── middleware/      # Middlewares
│   ├── model/           # Data models
│   ├── routegroup/      # Route groups
│   └── service/         # Business services
│
├── 📂 pkg/              # Public packages (exportable)
│   ├── cache/          # Cache (Redis)
│   ├── logger/         # Logger
│   ├── orm/            # ORM config
│   ├── response/       # Response format
│   ├── transaction/    # Transaction utils
│   ├── utils/          # Utilities
│   └── validator/      # Validation
│
├── 📂 docs/             # Swagger docs
├── 📄 main.go           # Application entry
├── 📄 Makefile          # Make commands (20+)
├── 📄 Dockerfile        # Docker image
├── 📄 docker-compose.yml # Docker Compose
└── 📄 app.yaml.template  # Config template
```

---

## 🎯 Use Cases

<table>
<tr>
<td width="33%">

### 🏢 Enterprise Apps
- Admin dashboards
- Management platforms
- Permission systems

</td>
<td width="33%">

### 🌐 API Services
- RESTful APIs
- Microservices
- Data interfaces

</td>
<td width="33%">

### 🚀 Rapid Development
- Project scaffold
- Tech validation
- Quick prototyping

</td>
</tr>
</table>

---

## 🤝 Contributing

We welcome all forms of contributions!

### How to Contribute

1. **Fork** this repository
2. **Create** feature branch (`git checkout -b feature/AmazingFeature`)
3. **Commit** your changes (`git commit -m 'Add some AmazingFeature'`)
4. **Push** to the branch (`git push origin feature/AmazingFeature`)
5. **Open** a Pull Request

### Types of Contributions

- 🐛 **Bug Fixes** - Found a bug? Submit an issue or PR
- ✨ **New Features** - Have an idea? We'd love to hear it
- 📝 **Documentation** - Docs unclear? Help us improve
- 🌍 **Translations** - Help support more languages
- 💡 **Suggestions** - All feedback welcome

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).

You are free to:
- ✅ Commercial use
- ✅ Modification
- ✅ Distribution
- ✅ Private use

The only requirement is to keep the copyright notice.

---

## 🙏 Acknowledgments

Thanks to all contributors!

<a href="https://github.com/code-innovator-zyx/gin-template/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=code-innovator-zyx/gin-template" />
</a>

### Inspiration

- [gin-gonic/gin](https://github.com/gin-gonic/gin) - Excellent web framework
- [gin-admin](https://github.com/LyricTian/gin-admin) - RBAC design reference
- [go-clean-arch](https://github.com/bxcodec/go-clean-arch) - Architecture inspiration

---

<div align="center">

## 🎉 Get Started

**Don't just star, try it now!**

[Quick Start](./docs/QUICK_START.md) · [Documentation](./docs/OPTIMIZATION_REPORT.md) · [Submit Issue](https://github.com/code-innovator-zyx/gin-template/issues)

### If this project helps you, please give it a ⭐️

**Made with ❤️ by [mortal](https://github.com/code-innovator-zyx)**

</div>

---

<div align="center">

**[⬆ Back to Top](#-gin-enterprise-template)**

</div>

