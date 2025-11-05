# 🚀 多缓存实现使用指南

## 📖 概述

项目支持三种缓存实现，可以通过配置文件自由切换：

| 类型 | 适用场景 | 性能 | 持久化 | 依赖 |
|------|---------|------|--------|------|
| **Redis** | 生产环境、分布式系统 | ⭐⭐⭐⭐⭐ | ✅ | Redis服务器 |
| **LevelDB** | 单机应用、嵌入式 | ⭐⭐⭐⭐ | ✅ | 无（本地文件） |
| **Memory** | 开发测试、临时缓存 | ⭐⭐⭐⭐⭐ | ❌ | 无 |

---

## 🎯 快速开始

### 方式1：使用Redis（推荐生产环境）

```yaml
# app.yaml
cache:
  type: redis
  redis:
    host: localhost
    port: 6379
    password: ""
    db: 0
    pool_size: 10
```

**优点**：
- ✅ 高性能
- ✅ 支持分布式
- ✅ 持久化
- ✅ 功能丰富

**缺点**：
- ⚠️ 需要单独部署Redis服务

### 方式2：使用LevelDB（单机应用）

```yaml
# app.yaml
cache:
  type: leveldb
  leveldb:
    path: ./data/leveldb
```

**优点**：
- ✅ 无需额外服务
- ✅ 持久化
- ✅ 单机性能高
- ✅ 部署简单

**缺点**：
- ⚠️ 不支持分布式
- ⚠️ 并发性能不如Redis

### 方式3：使用Memory（开发测试）

```yaml
# app.yaml
cache:
  type: memory
```

**优点**：
- ✅ 无需配置
- ✅ 性能最高
- ✅ 适合开发测试

**缺点**：
- ❌ 无持久化
- ❌ 进程重启数据丢失
- ❌ 不支持分布式

---

## 📁 架构设计

### 分层架构

```
┌─────────────────────────────────────────┐
│   业务层 (internal/service)              │
│   └── cache_service.go                   │
│       └── 业务缓存逻辑（权限、Token等）   │
└──────────────┬──────────────────────────┘
               │ 使用接口
               ▼
┌─────────────────────────────────────────┐
│   接口层 (pkg/cache)                     │
│   └── interface.go                       │
│       └── Cache 接口定义                 │
└──────────────┬──────────────────────────┘
               │ 实现接口
               ▼
┌─────────────────────────────────────────┐
│   实现层 (pkg/cache)                     │
│   ├── redis_adapter.go    - Redis实现   │
│   ├── leveldb_adapter.go  - LevelDB实现 │
│   ├── memory_adapter.go   - Memory实现  │
│   └── factory.go          - 工厂模式    │
└─────────────────────────────────────────┘
```

### 设计模式

#### 1. **接口模式** - 定义统一的缓存接口
```go
type Cache interface {
    Get(ctx, key, dest) error
    Set(ctx, key, value, ttl) error
    Delete(ctx, keys...) error
    // ...
}
```

#### 2. **适配器模式** - 适配不同的缓存实现
```go
type redisCache struct { ... }    // Redis适配器
type levelDBCache struct { ... }  // LevelDB适配器
type memoryCache struct { ... }   // Memory适配器
```

#### 3. **工厂模式** - 根据配置创建实例
```go
func InitCache(cfg CacheConfig) error {
    switch cfg.Type {
    case "redis":
        Client, err = NewRedisCache(...)
    case "leveldb":
        Client, err = NewLevelDBCache(...)
    case "memory":
        Client = NewMemoryCache()
    }
}
```

---

## 💻 代码示例

### 在业务代码中使用

```go
import "gin-template/internal/service"

// 创建缓存服务
cacheService := service.MustNewCacheService()

// 使用（无需关心底层是Redis还是LevelDB）
hasPermission, err := cacheService.CheckUserPermission(
    ctx, userID, "/api/v1/users", "GET",
)

// 无论底层用什么实现，接口都一样！
```

### 切换缓存实现

只需修改配置文件，无需修改代码：

```yaml
# 开发环境 - 使用内存缓存
cache:
  type: memory

# 测试环境 - 使用LevelDB
cache:
  type: leveldb
  leveldb:
    path: ./data/cache

# 生产环境 - 使用Redis
cache:
  type: redis
  redis:
    host: redis.prod.com
    port: 6379
    password: ${REDIS_PASSWORD}
```

---

## 🔧 高级功能

### 1. 权限缓存（Set数据结构）

所有实现都支持Set操作，用于高效存储用户权限：

```go
// 存储用户的所有权限资源
resources := []rbac.Resource{
    {Method: "GET", Path: "/api/v1/users"},
    {Method: "POST", Path: "/api/v1/posts"},
}

// 设置权限缓存（自动使用Set存储）
cacheService.SetUserPermissions(ctx, userID, resources)

// 检查权限（O(1)复杂度）
hasPermission := cacheService.CheckUserPermission(
    ctx, userID, "/api/v1/users", "GET",
)
```

**内部实现**：
- Redis: 使用`SADD`和`SISMEMBER`
- LevelDB: 模拟Set（`key:member:value`）
- Memory: Map模拟Set

### 2. TTL自动管理

```go
// 自动过期清理（所有实现都支持）
cacheService.SetInstance(ctx, "key", value, 10*time.Minute)

// 10分钟后自动删除
```

**内部实现**：
- Redis: 原生TTL支持
- LevelDB: 后台协程定期清理
- Memory: 后台协程定期清理

### 3. Pipeline批量操作

```go
// 使用Pipeline提高性能（所有实现都支持）
pipe := cache.GetClient().Pipeline()
pipe.Exists(ctx, "key1")
pipe.SIsMember(ctx, "set1", "member1")
pipe.Expire(ctx, "key2", time.Hour)
err := pipe.Exec(ctx)
```

---

## 📊 性能对比

### 读取性能

| 操作 | Redis | LevelDB | Memory |
|------|-------|---------|--------|
| Get单个key | 1-2ms | 0.5-1ms | 0.01ms |
| 权限检查(Set) | 1-2ms | 1-3ms | 0.01ms |
| Pipeline批量 | 2-3ms | 3-5ms | 0.1ms |

### 写入性能

| 操作 | Redis | LevelDB | Memory |
|------|-------|---------|--------|
| Set单个key | 1-2ms | 1-2ms | 0.01ms |
| 批量写入 | 2-4ms | 3-6ms | 0.1ms |

### 并发性能

| 缓存类型 | 并发读 | 并发写 | 说明 |
|----------|--------|--------|------|
| Redis | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 原生支持高并发 |
| LevelDB | ⭐⭐⭐⭐ | ⭐⭐⭐ | 读多写少场景好 |
| Memory | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | RWMutex保护 |

---

## 🎯 选择指南

### 生产环境推荐

#### 分布式系统 → Redis
```yaml
cache:
  type: redis
  redis:
    host: redis.prod.com
    port: 6379
    password: ${REDIS_PASSWORD}
    pool_size: 20
```

#### 单机应用 → LevelDB
```yaml
cache:
  type: leveldb
  leveldb:
    path: /var/lib/myapp/cache
```

### 开发测试推荐

#### 本地开发 → Memory
```yaml
cache:
  type: memory
```

#### 集成测试 → LevelDB
```yaml
cache:
  type: leveldb
  leveldb:
    path: ./tmp/test_cache
```

---

## ⚙️ 配置详解

### Redis完整配置

```yaml
cache:
  type: redis
  redis:
    host: localhost         # Redis服务器地址
    port: 6379             # 端口
    password: ""           # 密码（可选）
    db: 0                  # 数据库编号（0-15）
    pool_size: 10          # 连接池大小
```

**环境变量支持**：
```bash
export APP_CACHE_TYPE=redis
export APP_CACHE_REDIS_HOST=redis.prod.com
export APP_CACHE_REDIS_PASSWORD=secret
```

### LevelDB完整配置

```yaml
cache:
  type: leveldb
  leveldb:
    path: ./data/leveldb   # 数据存储路径
```

**注意**：
- 路径必须有写权限
- 自动创建目录
- 数据持久化到磁盘

### Memory配置

```yaml
cache:
  type: memory
```

无需额外配置！

---

## 🔄 迁移指南

### 从Redis迁移到LevelDB

```bash
# 1. 修改配置
vim app.yaml
# 将 type: redis 改为 type: leveldb

# 2. 添加LevelDB配置
cache:
  type: leveldb
  leveldb:
    path: ./data/leveldb

# 3. 重启应用
make restart
```

### 从LevelDB迁移到Redis

```bash
# 1. 启动Redis
docker run -d -p 6379:6379 redis:alpine

# 2. 修改配置
cache:
  type: redis
  redis:
    host: localhost
    port: 6379

# 3. 重启应用
make restart
```

---

## 🐛 常见问题

### Q1: 切换缓存类型后需要重启吗？

**A**: 是的，缓存类型在应用启动时初始化，需要重启才能生效。

### Q2: 可以同时使用多种缓存吗？

**A**: 当前设计是单例模式，只能选择一种。如果需要多种，可以：

```go
// 创建多个实例
redisCache, _ := cache.NewRedisCache(redisConfig)
leveldbCache, _ := cache.NewLevelDBCache(leveldbConfig)

// 不同场景使用不同缓存
```

### Q3: LevelDB的数据存在哪里？

**A**: 存储在配置的`path`目录下，是持久化的文件数据。

```bash
ls -la ./data/leveldb/
# 可以看到 .ldb, .log 等LevelDB文件
```

### Q4: Memory缓存会占用多少内存？

**A**: 取决于缓存的数据量。建议监控内存使用：

```go
import "runtime"

var m runtime.MemStats
runtime.ReadMemStats(&m)
fmt.Printf("内存使用: %d MB\n", m.Alloc/1024/1024)
```

### Q5: 如何查看当前使用的缓存类型？

**A**: 
```bash
# 方式1：健康检查接口
curl http://localhost:8080/api/v1/health

# 返回示例：
{
  "cache": {
    "status": "ok",
    "type": "redis"  ← 这里显示类型
  }
}

# 方式2：代码中获取
cacheType := cache.GetClient().Type()
```

---

## 📝 配置示例大全

### 示例1：开发环境（Memory）

```yaml
# app.yaml
app:
  name: myapp
  env: dev

cache:
  type: memory
```

### 示例2：测试环境（LevelDB）

```yaml
# app.yaml
app:
  name: myapp
  env: test

cache:
  type: leveldb
  leveldb:
    path: ./tmp/cache
```

### 示例3：生产环境（Redis单机）

```yaml
# app.yaml
app:
  name: myapp
  env: prod

cache:
  type: redis
  redis:
    host: redis.prod.com
    port: 6379
    password: ${REDIS_PASSWORD}  # 从环境变量读取
    db: 0
    pool_size: 20
```

### 示例4：生产环境（Redis + 环境变量）

```yaml
# app.yaml
cache:
  type: redis
  redis:
    host: localhost
    port: 6379
```

```bash
# 通过环境变量覆盖
export APP_CACHE_REDIS_HOST=redis.prod.com
export APP_CACHE_REDIS_PASSWORD=secret_password
export APP_CACHE_REDIS_POOL_SIZE=50
```

---

## 🔍 监控和调试

### 查看缓存状态

```bash
# 健康检查
curl http://localhost:8080/api/v1/health

# 响应示例
{
  "cache": {
    "status": "ok",
    "type": "redis"     # 或 leveldb, memory
  },
  "database": "ok",
  "status": "ok"
}
```

### Redis监控

```bash
# 连接Redis
redis-cli

# 查看所有key
KEYS *

# 查看特定前缀
KEYS permission:*

# 查看Set成员
SMEMBERS permission:1

# 查看内存使用
INFO memory
```

### LevelDB监控

```bash
# 查看数据文件大小
du -sh ./data/leveldb

# 查看文件数量
ls -l ./data/leveldb | wc -l
```

---

## 🚀 性能优化建议

### Redis优化

```yaml
cache:
  type: redis
  redis:
    pool_size: 50        # 增大连接池
    # 使用Redis Cluster
    # hosts: 
    #   - redis1:6379
    #   - redis2:6379
```

### LevelDB优化

```yaml
cache:
  type: leveldb
  leveldb:
    path: /ssd/cache     # 使用SSD存储
```

### Memory优化

```go
// 限制缓存大小（需要自己实现LRU）
// 或使用第三方库：github.com/hashicorp/golang-lru
```

---

## 📊 对比总结

### 功能对比

| 功能 | Redis | LevelDB | Memory |
|------|-------|---------|--------|
| 基础Get/Set | ✅ | ✅ | ✅ |
| Set集合操作 | ✅ 原生 | ✅ 模拟 | ✅ 模拟 |
| TTL过期 | ✅ 原生 | ✅ 后台清理 | ✅ 后台清理 |
| Pipeline | ✅ 原生 | ✅ Batch | ✅ 模拟 |
| 持久化 | ✅ RDB/AOF | ✅ LSM树 | ❌ |
| 分布式 | ✅ Cluster | ❌ | ❌ |
| Pub/Sub | ✅ | ❌ | ❌ |

### 适用场景

| 场景 | 推荐 | 原因 |
|------|------|------|
| 生产环境（分布式） | Redis | 高性能+分布式支持 |
| 生产环境（单机） | LevelDB | 无需额外服务 |
| 开发环境 | Memory | 简单快速 |
| 测试环境 | LevelDB | 可持久化验证 |
| CI/CD | Memory | 快速启动 |
| 嵌入式应用 | LevelDB | 无外部依赖 |

---

## 🎓 最佳实践

### 1. 分环境配置

```bash
# 开发环境
app-dev.yaml:
  cache:
    type: memory

# 测试环境
app-test.yaml:
  cache:
    type: leveldb
    leveldb:
      path: ./data/cache

# 生产环境
app-prod.yaml:
  cache:
    type: redis
    redis:
      host: redis.prod.com
```

### 2. 使用环境变量

```bash
# 生产环境敏感信息通过环境变量
export APP_CACHE_TYPE=redis
export APP_CACHE_REDIS_HOST=redis.prod.com
export APP_CACHE_REDIS_PASSWORD=secret
```

### 3. 健康检查

```go
// 在健康检查中显示缓存类型
health := gin.H{
    "cache": cache.GetClient().Type(),  // redis, leveldb, memory
}
```

### 4. 优雅降级

```go
// 缓存服务内部已实现降级
// 如果缓存不可用，自动查询数据库
hasPermission, err := cacheService.CheckUserPermission(...)
// 无论缓存是否可用，都能正常工作
```

---

## 🔗 相关文档

- [缓存服务API文档](./CACHE_SERVICE_GUIDE.md)
- [Redis官方文档](https://redis.io/documentation)
- [LevelDB文档](https://github.com/google/leveldb/blob/master/doc/index.md)

---

## 🎉 总结

通过接口抽象，你的项目现在支持：

- ✅ **三种缓存实现** - Redis、LevelDB、Memory
- ✅ **配置即切换** - 无需修改代码
- ✅ **统一接口** - 使用体验一致
- ✅ **优雅降级** - 缓存不可用自动降级
- ✅ **易于扩展** - 添加新实现只需实现接口

**这是企业级项目的标准做法！** 🎊

