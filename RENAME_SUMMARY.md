# ✅ 一键重命名功能已添加

## 🎉 功能说明

现在你的 `gin-template` 项目支持一键重命名功能！别人clone你的项目后，可以通过一个简单的命令将项目名称改为自己的项目名称。

## 🚀 使用方法

```bash
# 克隆项目
git clone <your-repo-url>
cd gin-template

# 一键重命名（例如改为 blog-api）
make rename NEW_NAME=blog-api

# 更新依赖
go mod tidy

# 运行项目
make run
```

## 📦 功能特性

### 自动更新的内容

✅ **go.mod** - 模块名称
```diff
- module gin-template
+ module blog-api
```

✅ **所有 .go 文件** - import 路径
```diff
- import "gin-template/internal/config"
+ import "blog-api/internal/config"
```

✅ **Makefile** - 应用名称
```diff
- APP_NAME := gin-template
+ APP_NAME := blog-api
```

✅ **docker-compose.yml** - 容器名称
```diff
- container_name: gin-template
+ container_name: blog-api
```

✅ **所有 .md 文档** - 项目名称引用

### 执行过程

```bash
$ make rename NEW_NAME=blog-api

正在将项目从 'gin-template' 重命名为 'blog-api'...

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

项目已从 'gin-template' 重命名为 'blog-api'

下一步操作：
  1. 运行: go mod tidy
  2. 运行: make init-config (如果还没有 app.yaml)
  3. 运行: make run

提示: 如果使用 Git，建议执行:
  git add .
  git commit -m 'chore: rename project to blog-api'
```

## 📚 相关文档

已创建的文档：
- ✅ **RENAME_GUIDE.md** - 详细的重命名指南，包含常见问题
- ✅ **QUICK_START.md** - 已更新，包含重命名步骤
- ✅ **README.md** - 已更新快速开始部分
- ✅ **OPTIMIZATION_REPORT.md** - 已记录此功能
- ✅ **CHANGELOG.md** - 已记录此更新

## ⚡ 实现细节

### Makefile 中的 rename 命令

```makefile
rename: ## 重命名项目 (用法: make rename NEW_NAME=your-project-name)
    # 1. 检查参数
    # 2. 更新 go.mod
    # 3. 更新所有 .go 文件的 import
    # 4. 更新 Makefile
    # 5. 更新 docker-compose.yml
    # 6. 更新所有 .md 文档
    # 7. 显示下一步操作提示
```

### 使用的技术

- `sed` - 文本替换
- `find` - 查找文件
- Makefile 变量和条件判断
- 友好的进度提示

## ✨ 优势

1. **零配置** - 不需要额外安装工具
2. **智能替换** - 只替换代码中的引用，不影响注释
3. **安全可靠** - 创建备份文件（.bak），替换后自动删除
4. **完整性** - 自动更新所有相关文件
5. **用户友好** - 清晰的进度提示和下一步操作说明

## 🎯 使用场景

### 场景1：创建新项目

```bash
git clone https://github.com/xxx/gin-template.git my-new-project
cd my-new-project
make rename NEW_NAME=my-new-project
go mod tidy
make run
```

### 场景2：用作公司内部模板

```bash
git clone https://github.com/company/gin-template.git product-service
cd product-service
make rename NEW_NAME=company/product-service
go mod tidy
make init-config
# 编辑 app.yaml
make run
```

### 场景3：快速原型开发

```bash
git clone https://github.com/xxx/gin-template.git demo-api
cd demo-api
make rename NEW_NAME=demo-api
go mod tidy
make up  # 使用 docker-compose 快速启动
```

## 🔧 命令列表

```bash
# 查看所有命令
make help

# 重命名项目
make rename NEW_NAME=xxx

# 初始化配置
make init-config

# 安装依赖
make install

# 运行测试
make test

# 运行应用
make run
```

## ⚠️ 注意事项

1. **只在第一次使用** - 重命名命令应该只在克隆模板后执行一次
2. **项目名称规范** - 建议使用小写字母和连字符，如 `blog-api`, `user-service`
3. **执行后更新依赖** - 重命名后记得运行 `go mod tidy`
4. **Git 提交** - 重命名后建议立即提交到版本控制

## 📖 更多信息

查看详细文档：
- [重命名指南](./RENAME_GUIDE.md)
- [快速开始](./QUICK_START.md)
- [优化报告](./OPTIMIZATION_REPORT.md)

---

**现在你的项目模板真正做到了"开箱即用"！🎉**

