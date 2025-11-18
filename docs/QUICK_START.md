# 🚀 快速开始指南

## 📋 前置要求

- Go 1.20+
- MySQL 5.7+ (可选)
- Redis 5.0+ (可选)
- Docker & Docker Compose (可选)

## 🎯 方式一：本地开发（推荐用于开发）

### 1. 克隆项目

```bash
git clone <your-repo-url>
cd gin-admin
```

### 2. 重命名项目（可选但推荐）

如果你想将项目重命名为自己的项目名称：

```bash
# 一键重命名项目和所有依赖
make rename NEW_NAME=your-project-name

# 例如：
make rename NEW_NAME=my-awesome-api
```

这个命令会自动：
- ✅ 更新 `go.mod` 模块名
- ✅ 更新所有 Go 文件的 import 路径
- ✅ 更新 Makefile 中的应用名称
- ✅ 更新 docker-compose.yml
- ✅ 更新所有文档中的项目名称

**重命名后记得运行：**
```bash
go mod tidy
```

### 3. 安装依赖

```bash
make install
# 或
go mod tidy
```

### 4. 初始化配置

```bash
make init-config
# 或手动复制
cp app.yaml.template app.yaml
```

### 5. 编辑配置文件

编辑 `app.yaml`，配置数据库和Redis（可选）：

```yaml
# 数据库配置（可选，不配置也可以运行）
database:
  dsn: "root:password@tcp(localhost:3306)/gin_template?charset=utf8mb4&parseTime=True&loc=Local"
  max_open_conns: 100
  max_idle_conns: 10
  max_life_time: 3600
  log_level: 1

# JWT配置（如果需要认证功能）
jwt:
  secret: "your-secret-key-change-this"
  expire: 86400  # 24小时

# Redis配置（可选，不配置会降级为无缓存模式）
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
  pool_size: 10
```

### 6. 启动服务

```bash
make run
# 或
go run main.go
```

### 7. 访问服务

- 健康检查: http://localhost:8080/api/v1/health
- Swagger文档: http://localhost:8080/swagger/v1/index.html

## 🐳 方式二：Docker Compose（推荐用于快速体验）

### 1. 准备配置文件

```bash
cp app.yaml.template app.yaml
```

### 2. 启动所有服务

```bash
make up
# 或
docker-compose up -d
```

这将启动三个容器：
- `gin-admin` - 应用服务 (端口 8080)
- `gin-admin-mysql` - MySQL数据库 (端口 3306)
- `gin-admin-redis` - Redis缓存 (端口 6379)

### 3. 查看日志

```bash
make logs
# 或
docker-compose logs -f
```

### 4. 访问服务

- 健康检查: http://localhost:8080/api/v1/health
- Swagger文档: http://localhost:8080/swagger/v1/index.html

### 5. 停止服务

```bash
make down
# 或
docker-compose down
```

## 🔧 常用命令

### 项目管理

```bash
make help         # 显示所有可用命令
make rename NEW_NAME=xxx  # 重命名项目（仅第一次使用）
make init-config  # 初始化配置文件
```

### 开发相关

```bash
make run          # 运行应用
make test         # 运行测试
make test-coverage # 运行测试并生成覆盖率报告
make fmt          # 格式化代码
make lint         # 代码检查
make swagger      # 生成Swagger文档
```

### 构建相关

```bash
make build        # 编译应用
make build-linux  # 编译Linux版本
make build-darwin # 编译macOS版本
make build-windows # 编译Windows版本
make build-all    # 编译所有平台版本
```

### Docker相关

```bash
make docker-build # 构建Docker镜像
make docker-run   # 运行Docker容器
make docker-stop  # 停止Docker容器
make up           # 启动docker-compose
make down         # 停止docker-compose
make logs         # 查看日志
```

## 📝 API测试

### 1. 健康检查

```bash
curl http://localhost:8080/api/v1/health
```

响应示例：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok",
    "timestamp": 1699027200,
    "database": "ok",
    "redis": "ok"
  }
}
```

### 2. 用户注册

```bash
curl -X POST http://localhost:8080/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password123",
    "email": "admin@example.com"
  }'
```

### 3. 用户登录

```bash
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password123"
  }'
```

### 4. 获取用户信息（需要Token）

```bash
curl http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer <your-token>"
```

## 🔍 目录结构说明

```
gin-admin/
├── internal/              # 内部包（不对外暴露）
│   ├── config/           # 配置管理
│   ├── core/             # 核心组件（全局变量、初始化）
│   ├── handler/          # HTTP处理器（路由注册）
│   ├── logic/            # 业务逻辑
│   ├── middleware/       # 中间件
│   ├── model/            # 数据模型
│   ├── routegroup/       # 路由组
│   └── service/          # 业务服务
├── pkg/                  # 公共包（可被外部使用）
│   ├── cache/           # 缓存（Redis）
│   ├── logger/          # 日志
│   ├── orm/             # ORM配置
│   ├── response/        # 响应工具
│   ├── transaction/     # 事务工具
│   ├── utils/           # 工具函数
│   └── validator/       # 验证工具
├── docs/                 # Swagger文档
├── logs/                 # 日志文件
├── main.go              # 应用入口
├── app.yaml.template    # 配置模板
├── Makefile             # Make命令
├── Dockerfile           # Docker镜像
├── docker-compose.yml   # Docker Compose配置
└── README.md            # 项目说明
```

## 📚 功能特性

### ✅ 已实现功能

- [x] 完整的RBAC权限系统（用户-角色-权限-资源）
- [x] JWT身份认证
- [x] Redis缓存支持（权限检查缓存）
- [x] 请求ID追踪
- [x] 结构化日志
- [x] Panic自动恢复
- [x] 请求参数验证
- [x] 健康检查（含DB和Redis状态）
- [x] Swagger API文档
- [x] 优雅关闭
- [x] 事务支持
- [x] 单元测试示例
- [x] Docker支持
- [x] Docker Compose支持

### 🔄 数据流程

```
HTTP请求
  ↓
Recovery中间件（Panic恢复）
  ↓
RequestID中间件（生成请求ID）
  ↓
Logger中间件（记录请求日志）
  ↓
CORS中间件（跨域处理）
  ↓
JWT中间件（身份验证）- 如果需要
  ↓
Permission中间件（权限验证）- 如果需要
  ↓
Handler（业务逻辑）
  ↓
  ├─→ Redis缓存（权限/数据缓存）
  └─→ MySQL数据库（持久化存储）
  ↓
统一响应格式
  ↓
返回客户端
```

## 🐛 常见问题

### Q1: 启动时提示数据库连接失败？

**A**: 检查以下几点：
1. 确认MySQL已启动
2. 检查`app.yaml`中的数据库配置是否正确
3. 确认数据库用户有创建数据库的权限
4. 如果不需要数据库，可以注释掉database配置

### Q2: Redis连接失败？

**A**: 
1. 确认Redis已启动
2. 检查`app.yaml`中的Redis配置
3. Redis是可选的，不配置也能正常运行（会降级为无缓存模式）

### Q3: JWT Token过期？

**A**: 
1. 检查`app.yaml`中的`jwt.expire`配置（单位：秒）
2. 重新登录获取新token
3. 可以调整过期时间，建议生产环境不超过24小时

### Q4: 权限验证失败？

**A**:
1. 确认已登录并获取token
2. 确认token放在`Authorization: Bearer <token>`头中
3. 确认该API已配置权限
4. 确认用户有对应的角色和权限

### Q5: Swagger文档无法访问？

**A**:
1. 检查`app.yaml`中`app.enable_swagger`是否为true
2. 运行`make swagger`重新生成文档
3. 访问 http://localhost:8080/swagger/v1/index.html

### Q6: 如何修改端口？

**A**: 编辑`app.yaml`：
```yaml
server:
  port: 9090  # 修改为你想要的端口
```

## 📖 更多文档

- [优化报告](./OPTIMIZATION_REPORT.md) - 详细的优化内容和性能对比
- [完整README](./README.md) - 项目详细说明
- [待办事项](./todo.md) - 后续开发计划

## 🤝 贡献

欢迎提交Issue和Pull Request！

## 📄 License

MIT License

---

**开始你的Gin项目开发吧！🎉**

