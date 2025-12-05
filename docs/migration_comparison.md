# 数据库迁移方案对比

## 📊 旧方案 vs 新方案对比

### 场景：项目有 5 个业务模块，共 20 个 Model

---

## 旧方案 ❌

### `internal/migrates/init.go` (混乱、难以维护)

```go
package migrates

import (
	"gin-admin/internal/model/rbac"
	"gin-admin/internal/model/mall"
	"gin-admin/internal/model/system"
	"gin-admin/internal/model/cms"
	"gin-admin/internal/model/analytics"
	"gin-admin/internal/services"
	"github.com/sirupsen/logrus"
)

func Do(svcContext *services.ServiceContext) error {
	if err := svcContext.Db.AutoMigrate(
		// RBAC 模块
		&rbac.User{},
		&rbac.Role{},
		&rbac.Permission{},
		&rbac.Resource{},
		
		// Mall 模块
		&mall.Product{},
		&mall.Order{},
		&mall.Category{},
		&mall.Customer{},
		
		// System 模块
		&system.Config{},
		&system.Log{},
		&system.Notification{},
		
		// CMS 模块
		&cms.Article{},
		&cms.Page{},
		&cms.Media{},
		&cms.Comment{},
		
		// Analytics 模块
		&analytics.Event{},
		&analytics.Report{},
		&analytics.Dashboard{},
		&analytics.Metric{},
		&analytics.Chart{},
	); err != nil {
		return err
	}
	logrus.Info("success migration")
	return nil
}
```

**问题：**
- ❌ 40+ 行代码混在一起
- ❌ 模块之间没有边界
- ❌ 添加新 model 必须修改这个核心文件
- ❌ 容易出错（忘记添加某个 model）
- ❌ 代码审查困难

---

## 新方案 ✅

### 文件结构（清晰、模块化）

```
internal/migrates/
├── registry.go      # 注册表核心逻辑
├── init.go          # 迁移主逻辑（简洁）
├── rbac.go          # RBAC 模块注册
├── mall.go          # Mall 模块注册
├── system.go        # System 模块注册
├── cms.go           # CMS 模块注册
└── analytics.go     # Analytics 模块注册
```

### `internal/migrates/init.go` (简洁)

```go
package migrates

import (
	"fmt"
	"gin-admin/internal/services"
	"github.com/sirupsen/logrus"
)

func Do(svcContext *services.ServiceContext) error {
	models := GetAllModels()  // 自动获取所有已注册的模型

	if len(models) == 0 {
		logrus.Warn("no models registered for migration")
		return nil
	}

	logrus.Infof("migrating %d models...", len(models))

	if err := svcContext.Db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}

	logrus.Info("migration completed successfully")
	return nil
}
```

### `internal/migrates/rbac.go` (模块独立)

```go
package migrates

import "gin-admin/internal/model/rbac"

func init() {
	RegisterGroup("rbac",
		&rbac.User{},
		&rbac.Role{},
		&rbac.Permission{},
		&rbac.Resource{},
	)
}
```

### `internal/migrates/mall.go`

```go
package migrates

import "gin-admin/internal/model/mall"

func init() {
	RegisterGroup("mall",
		&mall.Product{},
		&mall.Order{},
		&mall.Category{},
		&mall.Customer{},
	)
}
```

### `internal/migrates/system.go`

```go
package migrates

import "gin-admin/internal/model/system"

func init() {
	RegisterGroup("system",
		&system.Config{},
		&system.Log{},
		&system.Notification{},
	)
}
```

**优势：**
- ✅ 每个文件只有 10 行左右
- ✅ 模块清晰分离
- ✅ 添加新 model 只修改对应模块文件
- ✅ 核心逻辑 `init.go` 保持稳定
- ✅ 支持按模块迁移

---

## 🎯 实际使用对比

### 添加新 Model 的步骤

#### 旧方案：

1. 打开 `internal/migrates/init.go`
2. 导入新的 model 包
3. 在 `AutoMigrate()` 中添加 `&newpackage.NewModel{}`
4. 保存文件
5. ⚠️ 可能影响其他开发者的工作（修改核心文件）

#### 新方案：

1. 创建或打开 `internal/migrates/yourmodule.go`
2. 在 `RegisterGroup` 中添加 `&newpackage.NewModel{}`
3. 保存文件
4. ✅ 完全不影响其他模块

---

## 🚀 高级功能

### 选择性迁移（新方案独有）

```go
// 只迁移 RBAC 模块（开发阶段很有用）
migrates.DoGroup(svcContext, "rbac")

// 迁移多个指定模块
migrates.DoGroup(svcContext, "rbac", "system")

// 查看所有已注册的模块
migrates.ListGroups()
```

### 调试输出示例

```
INFO migrating 20 models...
INFO migration completed successfully

# 使用 ListGroups()
INFO registered groups: [rbac mall system cms analytics]
INFO   - rbac: 4 models
INFO   - mall: 4 models  
INFO   - system: 3 models
INFO   - cms: 5 models
INFO   - analytics: 4 models
```

---

## 📈 可维护性对比

| 指标 | 旧方案 | 新方案 |
|------|--------|--------|
| **单文件行数** | 50+ 行 | 10-20 行 |
| **模块耦合** | 高（全部耦合在一起） | 低（每个模块独立） |
| **添加新 model** | 修改核心文件 | 修改模块文件 |
| **Git 冲突概率** | 高（团队都改同一文件） | 低（各改各的模块） |
| **代码审查** | 困难（混在一起难以区分） | 简单（按模块审查） |
| **功能扩展** | 困难 | 简单（支持分组等） |
| **测试覆盖** | 无 | 有完整单元测试 |

---

## 🎨 总结

新方案通过**自动注册模式**实现了：

1. **高内聚**：每个模块管理自己的 model 注册
2. **低耦合**：模块之间互不影响
3. **易扩展**：从 10 个 model 扩展到 100 个也不会混乱
4. **团队友好**：减少 Git 冲突，提高协作效率

这是一个**生产级的设计方案**，适用于任何规模的 Go 项目！
