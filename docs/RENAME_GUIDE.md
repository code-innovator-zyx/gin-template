# 📝 项目重命名指南

## 为什么需要重命名？

当你使用 `gin-admin` 作为模板创建新项目时，你需要将项目名称改为自己的项目名称，这样可以：

- ✅ 避免与模板名称冲突
- ✅ 让代码更具可读性
- ✅ 正确设置 Go 模块名称
- ✅ 符合你的项目命名规范

## 🚀 一键重命名

我们提供了一个简单的命令来自动完成所有重命名工作：

```bash
make rename NEW_NAME=your-project-name
```

### 完整示例

假设你要创建一个名为 `blog-api` 的项目：

```bash
# 1. 克隆模板
git clone https://github.com/your-username/gin-admin.git blog-api
cd blog-api

# 2. 重命名项目
make rename NEW_NAME=blog-api

# 3. 更新依赖
go mod tidy

# 4. 初始化配置
make init-config

# 5. 运行项目
make run
```

## 📋 重命名命令会做什么？

执行 `make rename NEW_NAME=xxx` 时，脚本会自动完成以下操作：

### 1️⃣ 更新 `go.mod`

**修改前：**
```go
module gin-admin

go 1.24.0
```

**修改后：**
```go
module blog-api

go 1.24.0
```

### 2️⃣ 更新所有 Go 文件的导入路径

**修改前：**
```go
import (
    "gin-admin/internal/config"
    "gin-admin/pkg/logger"
    "gin-admin/pkg/response"
)
```

**修改后：**
```go
import (
    "blog-api/internal/config"
    "blog-api/pkg/logger"
    "blog-api/pkg/response"
)
```

### 3️⃣ 更新 `Makefile`

**修改前：**
```makefile
APP_NAME := gin-admin
```

**修改后：**
```makefile
APP_NAME := blog-api
```

### 4️⃣ 更新 `docker-compose.yml`

**修改前：**
```yaml
services:
  app:
    container_name: gin-admin
  mysql:
    container_name: gin-admin-mysql
  redis:
    container_name: gin-admin-redis
```

**修改后：**
```yaml
services:
  app:
    container_name: blog-api
  mysql:
    container_name: blog-api-mysql
  redis:
    container_name: blog-api-redis
```

### 5️⃣ 更新所有 Markdown 文档

所有 `*.md` 文件中的 `gin-admin` 都会被替换为新名称。

## 🎯 命令输出示例

```bash
$ make rename NEW_NAME=blog-api

正在将项目从 'gin-admin' 重命名为 'blog-api'...

步骤 1/5: 更新 go.mod 模块名...
✓ go.mod 已更新

步骤 2/5: 更新所有 Go 文件中的 import 路径...
✓ Go 文件导入路径已更新

步骤 3/5: 更新 Makefile...
✓ Makefile 已更新

步骤 4/5: 更新 docker-compose.yml...
✓ docker-compose.yml 已更新

步骤 5/5: 更新文档...
✓ 文档已更新

==========================================
✅ 重命名完成！
==========================================

项目已从 'gin-admin' 重命名为 'blog-api'

下一步操作：
  1. 运行: go mod tidy
  2. 运行: make init-config (如果还没有 app.yaml)
  3. 运行: make run

提示: 如果使用 Git，建议执行:
  git add .
  git commit -m 'chore: rename project to blog-api'
```

## ⚠️ 注意事项

### 1. 只在第一次使用

重命名命令应该**只在第一次克隆模板后执行一次**。不要在已经开发的项目上反复执行。

### 2. 项目名称规范

建议使用以下命名规范：

✅ **推荐的命名：**
```bash
make rename NEW_NAME=blog-api
make rename NEW_NAME=user-service
make rename NEW_NAME=my-project
make rename NEW_NAME=company/project-name
```

❌ **不推荐的命名：**
```bash
make rename NEW_NAME=Blog-API      # 不要用大写
make rename NEW_NAME=blog_api      # Go 模块名通常用连字符
make rename NEW_NAME="blog api"    # 不要用空格
```

### 3. 执行前备份（可选）

如果你担心出错，可以先备份：

```bash
# 创建备份
cp -r . ../gin-admin-backup

# 然后执行重命名
make rename NEW_NAME=your-project-name

# 如果有问题，可以恢复
# rm -rf *
# cp -r ../gin-admin-backup/* .
```

### 4. 检查重命名结果

重命名后，建议检查以下内容：

```bash
# 检查 go.mod
head -n 1 go.mod

# 检查是否能正常编译
go build

# 检查是否能正常运行
make run
```

## 🔧 手动重命名（不推荐）

如果你想手动重命名（不推荐，容易遗漏），需要修改以下文件：

1. **go.mod** - 模块名称
2. **所有 .go 文件** - import 路径
3. **Makefile** - APP_NAME 变量
4. **docker-compose.yml** - 容器名称
5. **所有 .md 文件** - 文档中的项目名称

## 🆘 常见问题

### Q1: 执行重命名命令后报错？

**A:** 确保：
1. 你在项目根目录下
2. 提供了 NEW_NAME 参数
3. 项目名称不包含特殊字符或空格

### Q2: 重命名后编译失败？

**A:** 执行以下命令：
```bash
go mod tidy
go clean -modcache
go mod download
go build
```

### Q3: 可以重命名多次吗？

**A:** 技术上可以，但不推荐。如果确实需要，建议：
1. 先手动将当前名称改回 `gin-admin`
2. 再执行重命名命令

或者直接手动修改相关文件。

### Q4: 重命名会影响 Git 历史吗？

**A:** 不会。重命名只修改文件内容，不影响 Git 历史。建议重命名后立即提交：

```bash
git add .
git commit -m "chore: rename project to your-project-name"
```

### Q5: Docker 镜像名称也会改变吗？

**A:** 是的。执行以下命令查看新的镜像名称：

```bash
make docker-build
docker images | grep your-project-name
```

## 📚 相关文档

- [快速开始指南](./QUICK_START.md)
- [优化报告](./OPTIMIZATION_REPORT.md)
- [更新日志](./CHANGELOG.md)

## 💡 最佳实践

1. **立即重命名**：克隆模板后，立即执行重命名，然后再开始开发
2. **规范命名**：使用清晰、有意义的项目名称
3. **提交更改**：重命名后立即提交到 Git
4. **更新文档**：根据你的项目需求更新 README 和其他文档

---

**祝你的项目开发顺利！🎉**

