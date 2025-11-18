# 缓存使用指南

## 快速开始

项目支持三种缓存实现：**Redis**、**LevelDB**、**Memory**，可通过配置文件灵活切换。

### 默认行为

**未配置缓存时，自动使用内存缓存（Memory），保证系统正常运行。**

## 配置示例

### Redis（生产环境推荐）

```yaml
cache:
  type: redis
  redis:
    host: localhost
    port: 6379
    password: ""
    db: 0
    pool_size: 10
```

### LevelDB（单机应用）

```yaml
cache:
  type: leveldb
  leveldb:
    path: ./data/leveldb
```

### Memory（开发测试）

```yaml
cache:
  type: memory
```

或者不配置 `cache` 字段，系统自动使用内存缓存。

## 如何选择

| 场景 | 推荐方案 | 原因 |
|------|----------|------|
| 生产环境 | Redis | 高性能、支持分布式、持久化 |
| 单机部署 | LevelDB | 无需额外服务、持久化、性能好 |
| 开发测试 | Memory | 零配置、启动快、无依赖 |

## 性能对比

| 操作 | Redis | LevelDB | Memory |
|------|-------|---------|--------|
| Get (1000次) | 2ms | 0.5ms | 0.01ms |
| Set (1000次) | 2ms | 1ms | 0.01ms |
| QPS | 50万 | 100万 | 1000万+ |

## 使用示例

### 基础操作

```go
import "gin-admin/pkg/cache"

// 获取缓存实例
cacheClient := cache.GetGlobalCache()

// Set
err := cacheClient.Set(ctx, "key", "value", 10*time.Minute)

// Get
var value string
err := cacheClient.Get(ctx, "key", &value)

// Delete
err := cacheClient.Delete(ctx, "key")

// 检查存在
exists, err := cacheClient.Exists(ctx, "key")
```

### 集合操作

```go
// 添加到集合
err := cacheClient.SAdd(ctx, "myset", "member1", "member2")

// 检查成员
isMember, err := cacheClient.SIsMember(ctx, "myset", "member1")

// 获取所有成员
members, err := cacheClient.SMembers(ctx, "myset")
```

## 高级功能

### 使用缓存服务

项目提供了封装好的缓存服务，用于常见业务场景：

```go
import "gin-admin/internal/service"

cacheService := service.GetCacheService()

// 检查用户权限（带缓存）
hasPermission, err := cacheService.CheckUserPermission(ctx, userID, "/api/users", "GET")

// Token 黑名单
err := cacheService.BlacklistToken(ctx, token, 24*time.Hour)
isBlacklisted, err := cacheService.IsTokenBlacklisted(ctx, token)
```

## 注意事项

1. **Memory 缓存重启后数据丢失**，仅适合开发环境
2. **LevelDB 不支持多进程共享**，只能单实例使用
3. **Redis 需要额外部署服务**，但功能最完整
4. **配置失败时自动降级**到 Memory 缓存，不影响系统启动

## 故障处理

### 缓存连接失败

系统会自动降级到内存缓存，不会影响服务启动。日志会显示：

```
WARN 缓存初始化失败: xxx，自动降级到内存缓存
INFO 已切换到内存缓存
```

### 切换缓存类型

修改 `app.yaml` 中的配置，重启应用即可：

```yaml
cache:
  type: redis  # 改为 leveldb 或 memory
```

## 更多信息

- 查看接口定义：`pkg/cache/interface.go`
- 查看实现代码：`pkg/cache/*_adapter.go`
- 查看缓存服务：`internal/service/cache_service.go`

