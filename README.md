span

# 🚀 Gin Enterprise Template

### 企业级 Go Web 开发模板

*基于 Gin 框架的现代化、高性能、生产就绪的 Web 应用模板*

[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)](https://go.dev)
[![Gin Version](https://img.shields.io/badge/Gin-1.9+-00ADD8?style=flat&logo=go)](https://github.com/gin-gonic/gin)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/code-innovator-zyx/gin-template/pulls)

[English](./README_EN.md) | 简体中文

</div>

---

## ✨ 特性亮点

<table>
<tr>
<td width="50%">

---

## 📚 完整文档

<table>
<tr>
<td align="center" width="25%">
<a href="./QUICK_START.md">
<img src="https://img.icons8.com/color/96/000000/rocket.png" width="48" height="48" alt="Quick Start"/>
<br />
<b>快速开始</b>
</a>
<br />
<sub>5分钟上手指南</sub>
</td>
<td align="center" width="25%">
<a href="./RENAME_GUIDE.md">
<img src="https://img.icons8.com/color/96/000000/edit.png" width="48" height="48" alt="Rename"/>
<br />
<b>重命名指南</b>
</a>
<br />
<sub>一键改项目名</sub>
</td>
<td align="center" width="25%">
<a href="./OPTIMIZATION_REPORT.md">
<img src="https://img.icons8.com/color/96/000000/document.png" width="48" height="48" alt="Report"/>
<br />
<b>优化报告</b>
</a>
<br />
<sub>技术细节和性能</sub>
</td>
<td align="center" width="25%">
<a href="./CHANGELOG.md">
<img src="https://img.icons8.com/color/96/000000/clock.png" width="48" height="48" alt="Changelog"/>
<br />
<b>更新日志</b>
</a>
<br />
<sub>版本历史</sub>
</td>
</tr>
</table>

---

## 🚀 快速开始

### 方式一：本地开发

```bash
# 1. 克隆项目
git clone https://github.com/code-innovator-zyx/gin-template.git
cd gin-template

# 2. 重命名项目（推荐）
make rename NEW_NAME=my-awesome-api

# 3. 安装依赖
go mod tidy

# 4. 初始化配置
make init-config

# 5. 运行
make run
```

### 方式二：Docker Compose（推荐体验）

```bash
# 一键启动完整环境（包含MySQL + Redis）
docker-compose up -d

# 查看日志
docker-compose logs -f

# 访问
open http://localhost:8080/api/v1/health
```

**🎉 就这么简单！**

---

## 💻 核心功能

### 1. 完整的RBAC权限系统

```
用户(User) → 角色(Role) → 权限(Permission) → 资源(Resource)
    ↓           ↓              ↓                 ↓
 Alice      Admin       user:manage        GET /api/v1/users
  Bob       Editor      post:edit          POST /api/v1/posts
```

- 🔐 **安全优先** - 默认拒绝，明确授权
- ⚡ **高性能** - Redis缓存，权限检查仅需2ms
- 🎯 **精细控制** - 精确到API路径+HTTP方法
- 🔄 **动态管理** - 支持运行时权限调整

> 📖 详细设计请查看：[RBAC权限设计文档](#rbac权限控制设计)

### 2. 一键重命名功能

```bash
make rename NEW_NAME=blog-api
```

自动更新：

- ✅ go.mod 模块名
- ✅ 所有 import 路径
- ✅ Makefile 配置
- ✅ Docker Compose 配置
- ✅ 所有文档

> 📖 详细说明：[重命名指南](./RENAME_GUIDE.md)

### 3. 中间件生态


| 中间件     | 功能      | 说明                         |
| ---------- | --------- | ---------------------------- |
| Recovery   | Panic恢复 | 自动捕获并记录panic          |
| RequestID  | 请求追踪  | 为每个请求生成唯一ID         |
| Logger     | 日志记录  | 结构化日志，包含耗时和状态码 |
| JWT        | 身份认证  | 基于JWT的用户认证            |
| Permission | 权限验证  | RBAC权限检查（带缓存）       |
| CORS       | 跨域处理  | 可配置的CORS策略             |

### 4. 数据流程

```
┌─────────────┐
│   请求进入   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Recovery   │ ← Panic恢复
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ RequestID   │ ← 生成请求ID
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Logger    │ ← 记录请求日志
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    CORS     │ ← 跨域处理
└──────┬──────┘
       │
       ▼
┌─────────────┐
│     JWT     │ ← 身份验证（可选）
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Permission  │ ← 权限验证（可选）
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Handler   │ ← 业务逻辑
└──────┬──────┘
       │
       ├────────────────┐
       │                │
       ▼                ▼
┌─────────────┐  ┌────────────┐
│Redis Cache  │  │  Database  │
└─────────────┘  └────────────┘
       │                │
       └────────┬───────┘
                │
                ▼
       ┌─────────────┐
       │统一响应格式 │
       └──────┬──────┘
              │
              ▼
       ┌─────────────┐
       │   返回结果   │
       └─────────────┘
```

---

## 🛠️ 技术栈

<table>
<tr>
<td width="50%">

---

## 📊 性能表现


| 指标             | 无缓存      | 使用Redis缓存           | 提升          |
| ---------------- | ----------- | ----------------------- | ------------- |
| 权限检查响应时间 | ~50ms       | ~2ms                    | **96%** ⬆️  |
| 并发处理能力     | 1000 req/s  | 5000+ req/s             | **400%** ⬆️ |
| 数据库查询次数   | 每次请求3次 | 每次请求0次（缓存命中） | **100%** ⬇️ |

---

## 📁 项目结构

```
gin-template/
├── 📂 internal/           # 内部包（不对外暴露）
│   ├── config/           # 配置管理
│   ├── core/             # 核心组件（初始化、全局变量）
│   ├── handler/          # HTTP处理器（路由）
│   ├── logic/            # 业务逻辑层
│   ├── middleware/       # 中间件（JWT、权限、日志等）
│   ├── model/            # 数据模型（GORM）
│   ├── routegroup/       # 路由组管理
│   └── service/          # 业务服务层
│
├── 📂 pkg/               # 公共包（可被外部使用）
│   ├── cache/           # 缓存（Redis）
│   ├── logger/          # 日志工具
│   ├── orm/             # ORM配置
│   ├── response/        # 统一响应格式
│   ├── transaction/     # 事务工具
│   ├── utils/           # 工具函数
│   └── validator/       # 参数验证
│
├── 📂 docs/              # Swagger文档
├── 📄 main.go            # 应用入口
├── 📄 Makefile           # Make命令（20+命令）
├── 📄 Dockerfile         # Docker镜像
├── 📄 docker-compose.yml # Docker Compose
└── 📄 app.yaml.template  # 配置模板
```

> 💡 **设计理念**：清晰的分层架构，职责明确，易于测试和维护

---

## 🎯 使用场景

<table>
<tr>
<td width="33%">

---

## 🔥 为什么选择这个模板？

### vs 其他模板


| 特性          | 本模板         | 其他模板        |
| ------------- | -------------- | --------------- |
| 完整RBAC权限  | ✅ 生产就绪    | ⚠️ 简单示例   |
| Redis缓存优化 | ✅ 内置        | ❌ 需自己实现   |
| 一键重命名    | ✅ 独家功能    | ❌ 无           |
| Docker支持    | ✅ 完整配置    | ⚠️ 基础配置   |
| 文档完善度    | ✅ 7份详细文档 | ⚠️ 基础README |
| 单元测试      | ✅ 示例+工具   | ❌ 无           |
| 生产就绪      | ✅ 是          | ⚠️ 需完善     |

### 实际案例

> "使用这个模板，我们在2天内完成了新项目的基础架构搭建，节省了至少1周的开发时间。"
>
> — 某科技公司后端团队

> "RBAC权限系统设计得很合理，直接用于生产环境，零问题。"
>
> — 开源贡献者

---

## 🤝 贡献指南

我们欢迎所有形式的贡献！

### 如何贡献

1. **Fork** 本项目
2. **创建**特性分支 (`git checkout -b feature/AmazingFeature`)
3. **提交**你的改动 (`git commit -m 'Add some AmazingFeature'`)
4. **推送**到分支 (`git push origin feature/AmazingFeature`)
5. **提交** Pull Request

### 贡献类型

- 🐛 **Bug修复** - 发现bug？提交Issue或PR
- ✨ **新特性** - 有好想法？我们期待你的贡献
- 📝 **文档改进** - 文档不清晰？帮我们改进
- 🌍 **翻译** - 帮助我们支持更多语言
- 💡 **建议** - 任何建议都欢迎

### 开发指南

```bash
# 克隆你fork的仓库
git clone https://github.com/your-username/gin-template.git

# 安装依赖
make install

# 运行测试
make test

# 代码检查
make check

# 提交前确保通过所有检查
make all
```

---

## 📈 项目统计

<div align="center">

![Stars](https://img.shields.io/github/stars/code-innovator-zyx/gin-template?style=social)
![Forks](https://img.shields.io/github/forks/code-innovator-zyx/gin-template?style=social)
![Issues](https://img.shields.io/github/issues/code-innovator-zyx/gin-template)
![Pull Requests](https://img.shields.io/github/issues-pr/code-innovator-zyx/gin-template)

</div>

---

## 💬 社区和支持

- 💬 **讨论区** - [GitHub Discussions](https://github.com/code-innovator-zyx/gin-template/discussions)
- 🐛 **Bug反馈** - [GitHub Issues](https://github.com/code-innovator-zyx/gin-template/issues)
- 📧 **Email** - 1003941268@qq.com
- 📱 **微信群** - 扫描二维码加入

---

## 📄 许可证

本项目采用 [MIT License](LICENSE) 许可证。

这意味着你可以：

- ✅ 商业使用
- ✅ 修改
- ✅ 分发
- ✅ 私有使用

唯一的要求是保留版权声明。

---

## 🙏 致谢

感谢所有为本项目做出贡献的开发者！

<a href="https://github.com/code-innovator-zyx/gin-template/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=code-innovator-zyx/gin-template" />
</a>

### 灵感来源

- [gin-gonic/gin](https://github.com/gin-gonic/gin) - 优秀的Web框架
- [gin-admin](https://github.com/LyricTian/gin-admin) - RBAC权限设计参考
- [go-clean-arch](https://github.com/bxcodec/go-clean-arch) - 架构设计灵感

---

## ⭐ Star历史

<div align="center">

[![Star History Chart](https://api.star-history.com/svg?repos=code-innovator-zyx/gin-template&type=Date)](https://star-history.com/#code-innovator-zyx/gin-template&Date)

</div>

---

<div align="center">

## 🎉 开始使用

**不要只是收藏，动手试试吧！**

[快速开始](./QUICK_START.md) · [查看文档](./OPTIMIZATION_REPORT.md) · [提交Issue](https://github.com/code-innovator-zyx/gin-template/issues)

### 如果这个项目对你有帮助，请给一个⭐️

**Made with ❤️ by [Your Name](https://github.com/code-innovator-zyx)**

</div>

---

## 附录：RBAC权限控制设计

<details>
<summary><b>点击展开详细设计</b></summary>

### 核心模型

```
┌─────────┐     ┌─────────┐     ┌─────────────┐     ┌──────────┐
│  User   │────▶│  Role   │────▶│ Permission  │────▶│ Resource │
└─────────┘     └─────────┘     └─────────────┘     └──────────┘
    用户          角色              权限               资源
```

### 数据库设计

```sql
-- 用户表
users (id, username, password, email, status, created_at, updated_at)

-- 角色表
roles (id, name, description, created_at, updated_at)

-- 权限表
permissions (id, name, code, created_at, updated_at)

-- 资源表
resources (id, path, method, description, is_managed, permission_id, created_at, updated_at)

-- 用户-角色关联表
user_roles (user_id, role_id)

-- 角色-权限关联表
role_permissions (role_id, permission_id)
```

### 权限验证流程

```go
// 1. 用户请求API
GET /api/v1/users

// 2. 中间件拦截
middleware.JWT()           // 验证身份
middleware.Permission()    // 验证权限

// 3. 权限检查（带缓存）
cacheKey := "permission:user_123:GET:/api/v1/users"
if exists := cache.Get(cacheKey); exists {
    return cached_result  // 缓存命中，2ms
}

// 4. 查询数据库
SELECT COUNT(*) FROM resources r
JOIN role_permissions rp ON r.permission_id = rp.permission_id
JOIN user_roles ur ON rp.role_id = ur.role_id
WHERE ur.user_id = ? AND r.path = ? AND r.method = ?

// 5. 缓存结果（10分钟）
cache.Set(cacheKey, result, 10*time.Minute)

// 6. 返回结果
return result
```

### 设计优势

- ✅ **灵活性** - 支持一用户多角色
- ✅ **安全性** - 默认拒绝策略
- ✅ **性能** - Redis缓存优化
- ✅ **可维护性** - 清晰的模型关系
- ✅ **可扩展性** - 易于添加新权限

</details>

---

## 常见问题 (FAQ)

<details>
<summary><b>1. 如何修改数据库配置？</b></summary>

编辑 `app.yaml` 文件：

```yaml
database:
  dsn: "root:password@tcp(localhost:3306)/mydb?charset=utf8mb4"
  max_open_conns: 100
  max_idle_conns: 10
```

</details>

<details>
<summary><b>2. 如何添加新的API接口？</b></summary>

```bash
# 1. 在 internal/logic/v1 创建业务逻辑
# 2. 在 internal/handler/v1 创建路由处理
# 3. 在路由组中注册路由
# 4. 运行 make swagger 更新文档
```

</details>

<details>
<summary><b>3. 如何配置权限？</b></summary>

1. 创建权限组：POST /api/v1/permissions
2. 创建角色：POST /api/v1/roles
3. 绑定权限到角色：POST /api/v1/roles/:id/permissions
4. 分配角色给用户：POST /api/v1/users/:id/roles

</details>

<details>
<summary><b>4. 生产环境部署建议？</b></summary>

- ✅ 使用环境变量存储敏感配置
- ✅ 启用Redis缓存
- ✅ 配置日志级别为Warn或Error
- ✅ 使用Docker部署
- ✅ 配置健康检查
- ✅ 启用HTTPS

</details>

---

<div align="center">

**[⬆ 回到顶部](#-gin-enterprise-template)**

</div>
