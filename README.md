# Gin Web 应用框架模板

一个基于 Gin 框架的企业级 Web 应用模板，集成了常用功能和最佳实践，为快速开发高质量 Go 后端应用提供坚实基础。

## 核心特性

- **模块化架构**：清晰的目录结构，严格的关注点分离，便于扩展和维护
- **配置管理**：基于 Viper 的配置系统，支持多环境配置和热加载
- **日志系统**：集成 Logrus 日志库，支持多种输出格式、日志级别和文件轮转
- **中间件集成**：内置常用中间件（请求日志、异常恢复、CORS、JWT 认证等）
- **数据库支持**：基于 GORM 的数据库操作，支持多种数据库和迁移管理
- **API 文档**：集成 Swagger 自动生成 API 文档
- **用户认证**：完整的 JWT 认证机制实现
- **标准化响应**：统一的 API 响应格式和错误处理
- **健康检查**：内置健康检查接口，便于监控和部署

## 详细目录结构

```
.
├── docs/               # API 文档
│   ├── v1_docs.go      # Swagger 文档定义
│   ├── v1_swagger.json # Swagger JSON 文件
│   └── v1_swagger.yaml # Swagger YAML 文件
├── internal/           # 内部包
│   ├── config/         # 配置文件和配置管理
│   │   ├── app.yaml    # 应用配置文件
│   │   └── config.go   # 配置加载和管理
│   ├── core/           # 核心组件
│   │   ├── global.go   # 全局变量
│   │   └── initialize.go # 应用初始化
│   ├── handler/        # 请求处理器
│   │   ├── router.go   # 路由配置
│   │   └── v1/         # 版本1 API处理器
│   ├── logic/          # 业务逻辑
│   │   └── v1/         # 版本1 业务逻辑
│   ├── middleware/     # 中间件
│   │   ├── cors.go     # 跨域中间件
│   │   ├── jwt.go      # JWT 认证中间件
│   │   └── permission.go # 权限中间件
│   ├── model/          # 数据模型
│   │   ├── rbac/       # RBAC模型
│   │   └── user.go     # 用户模型
│   ├── orm/            # ORM 配置
│   │   └── gorm.go     # GORM 配置
│   └── service/        # 业务服务
│       ├── rbac_service.go # RBAC服务
│       └── user_service.go # 用户服务
├── pkg/                # 公共包
│   ├── logger/         # 日志工具
│   │   └── logger.go   # 日志配置
│   ├── response/       # 响应工具
│   │   └── response.go # 统一响应格式
│   └── utils/          # 通用工具
│       └── jwt.go      # JWT 工具
├── router/             # 路由配置
│   ├── router.go       # 主路由
│   └── v1/             # 版本 1 路由
│       ├── routes.go   # 路由注册
│       └── user/       # 用户路由
├── scripts/            # 脚本文件
├── Makefile            # 项目管理命令
├── main.go             # 应用入口
├── go.mod              # Go 模块定义
└── README.md           # 项目说明
```

## 快速开始

### 前置条件

- Go 1.20+
- MySQL 5.7+ 或其他支持的数据库

### 安装

1. 克隆项目

```bash
git clone https://github.com/yourusername/gin-template.git
cd gin-template
```

2. 安装依赖

```bash
go mod tidy
```

3. 配置数据库
```shell
mv config/app.yaml.template config/app.yaml
```
修改 `config/app.yaml`  文件中的数据库配置：

```yaml
database:
  driver: mysql
  host: localhost
  port: 3306
  username: root
  password: your_password
  dbname: gin_template
  max_idle_conns: 10
  max_open_conns: 100
```

4. 运行应用

```bash
# 直接运行
go run main.go

# 或使用 Makefile
make run
```

应用将在 `http://localhost:8080` 启动。

### 使用 Makefile

项目提供了 Makefile 来简化常用操作：

```bash
# 运行应用
make run

# 构建应用
make build

# 生成 API 文档
make swagger

# 运行测试
make test

# 清理构建文件
make clean
```

## API 文档

项目集成了 Swagger 文档，启动应用后可以通过以下地址访问：

```
http://localhost:8080/swagger/index.html
```

## 项目扩展指南

### 添加新 API 接口

1. 在 `api/v1/` 目录下创建新的控制器
2. 在 `internal/service/` 中实现业务逻辑
3. 在 `router/v1/routes.go` 中注册路由

### 添加新模型

1. 在 `internal/model/` 目录下创建新的模型文件
2. 在模型中定义表结构和关联关系
3. 在应用初始化时注册模型

## 工程化优化建议

以下是进一步完善项目的建议：

1. **单元测试覆盖**：为核心业务逻辑和工具函数添加单元测试
2. **CI/CD 集成**：添加 GitHub Actions 或 GitLab CI 配置
3. **容器化部署**：添加 Dockerfile 和 docker-compose.yml
4. **数据库迁移**：集成数据库迁移工具，如 golang-migrate
5. **缓存层**：添加 Redis 缓存支持
6. **限流和熔断**：实现 API 限流和服务熔断机制
7. **监控集成**：添加 Prometheus 指标收集
8. **分布式追踪**：集成 OpenTelemetry 或 Jaeger
9. **国际化支持**：添加多语言支持
10. **更完善的错误处理**：实现统一的错误码系统

## 贡献指南

欢迎提交 Issue 和 Pull Request 来完善这个模板。

## 许可证

[MIT License](LICENSE)