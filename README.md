# Gin Admin

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

## RBAC权限控制设计

本项目采用基于角色的访问控制（Role-Based Access Control，RBAC）模型，实现了灵活而高效的权限管理系统。

### 核心模型设计

#### 1. 用户（User）
- 系统的操作主体，可以被分配一个或多个角色
- 用户通过角色间接获取权限，而不是直接分配权限，简化了权限管理

#### 2. 角色（Role）
- 权限的集合，作为用户和权限之间的桥梁
- 支持多角色分配，满足复杂的组织结构需求
- 角色可以根据业务需求灵活创建和调整

#### 3. 权限（Permission）
- 代表系统中的操作权限，是权限控制的最小粒度
- 每个权限通过唯一的编码（Code）标识，便于程序逻辑判断
- 权限与资源是一对多关系，一个权限可以管理多个API资源

#### 4. 资源（Resource）
- 代表系统中的API接口资源
- 每个资源由路径（Path）和HTTP方法（Method）唯一标识
- 资源可以被权限组管理，也可以处于未管理状态
- **安全优先原则**：未被权限组管理的资源（游离状态）默认无法被任何用户访问，确保API安全

### 关系模型

1. **用户-角色**：多对多关系，通过`user_roles`关联表实现
2. **角色-权限**：多对多关系，通过`role_permissions`关联表实现
3. **权限-资源**：一对多关系，资源通过`permission_id`外键关联到权限

### 设计优势

1. **简化的权限分配**
   - 通过角色批量分配权限，减少权限管理的复杂度
   - 用户权限变更只需调整角色分配，无需逐一修改权限

2. **精细的权限控制**
   - 基于API路径和HTTP方法的精确权限控制
   - 支持动态识别和管理新增的API接口

3. **高性能设计**
   - 合理的数据库索引设计，提高查询效率
   - 优化的权限检查SQL，减少表连接次数

4. **灵活的资源管理**
   - 资源可以动态绑定到权限组，也可以解绑成未管理状态
   - 支持系统启动时自动扫描和更新API资源

5. **安全优先策略**
   - 采用"默认拒绝"的安全原则，未明确授权的资源不可访问
   - 资源必须被权限组管理且分配权限后才能被访问
   - 游离状态的资源（未被管理或未分配权限组）对所有用户不可见

6. **事务安全**
   - 关键操作（如删除角色）采用事务处理，确保数据一致性
   - 级联删除相关数据，避免出现孤立数据

7. **可扩展性**
   - 清晰的模型边界，便于扩展新的权限控制需求
   - 支持未来添加更复杂的权限规则和条件

### 权限验证流程

1. 用户请求API接口
2. 中间件拦截请求，提取用户ID、请求路径和HTTP方法
3. **首先检查资源状态**：
   - 如果资源未被管理（`is_managed=false`）或未分配权限组（`permission_id=0`），直接拒绝访问
   - 只有被明确管理且分配了权限组的资源才会进入下一步验证
4. 查询用户→角色→权限→资源的关联关系
5. 判断用户是否有权限访问请求的资源
6. 根据验证结果允许或拒绝访问

### 实现细节

- 使用GORM进行ORM映射，简化数据库操作
- 为关键字段添加索引，优化查询性能
- 采用预加载技术减少N+1查询问题
- 权限检查使用优化的原生SQL，减少表连接次数
- 两阶段权限验证，确保API安全性

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
│   └── service/        # 业务服务
│       ├── rbac_service.go # RBAC服务
│       └── user_service.go # 用户服务
├── pkg/                # 公共包
│   ├── cache/          # 缓存 配置
│   │   └── redis.go    # redis 配置
│   ├── logger/         # 日志工具
│   │   └── logger.go   # 日志配置
│   ├── response/       # 响应工具
│   │   └── response.go # 统一响应格式
│   ├── orm/            # ORM 配置
│   │   └── gorm.go     # GORM 配置

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
git clone https://github.com/yourusername/gin-admin.git
cd gin-admin
```

2. 安装依赖

```bash
go mod tidy
```

3. 配置数据库
```shell
mv app.yaml.template app.yaml
```
修改 `app.yaml`  文件中的数据库配置：

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

## 贡献指南

欢迎提交 Issue 和 Pull Request 来完善这个模板。

## 许可证

[MIT License](LICENSE)