package migrates

/*
 * @Author: zouyx
 * @Email: 1003941268@qq.com
 * @Date:   2025 2025/12/5 下午6:27
 * @Package: 模块注册示例
 */

// ============================================
// 示例：如何为新模块添加数据库迁移
// ============================================

/*
添加新模块迁移的步骤：

1. 在 migrates 包中创建新的注册文件，例如 mall.go：

```go
package migrates

import (
	"gin-admin/internal/model/mall"
)

func init() {
	RegisterGroup("mall",
		&mall.Product{},
		&mall.Order{},
		&mall.Category{},
		&mall.Customer{},
	)
}
```

2. 就这样！无需修改任何其他代码

当程序启动时，init() 会自动执行，模型会被自动注册。
然后调用 migrates.Do(svcContext) 就会自动迁移所有已注册的模型。

============================================
高级用法：
============================================

1. 只迁移特定模块：

```go
// 只迁移 rbac 模块
err := migrates.DoGroup(svcContext, "rbac")

// 迁移多个指定模块
err := migrates.DoGroup(svcContext, "rbac", "mall")
```

2. 查看已注册的模块（调试用）：

```go
migrates.ListGroups()
// 输出：
// registered groups: [rbac mall system]
//   - rbac: 4 models
//   - mall: 5 models
//   - system: 2 models
```

3. 手动获取模型列表：

```go
// 获取所有模型
allModels := migrates.GetAllModels()

// 获取特定模块的模型
rbacModels := migrates.GetGroupModels("rbac")

// 获取所有模块名称
groups := migrates.GetAllGroups()
```

============================================
为什么这样设计？
============================================

优势：
✅ 模块解耦：每个模块管理自己的模型注册
✅ 零侵入：添加新模块不需要修改核心迁移代码
✅ 易扩展：支持分组管理、选择性迁移
✅ 自动化：通过 init() 函数自动注册
✅ 可测试：提供了完整的测试覆盖

对比旧方案：
❌ 旧方案：每添加一个 model 都要在 init.go 中手动添加
✅ 新方案：创建新的 xxx.go 文件即可，完全独立

当有 100 个 model 时：
❌ 旧方案：init.go 文件会有 100 行 &xxx.Model{}
✅ 新方案：10 个模块文件，每个文件 10 行，清晰明了

============================================
*/
