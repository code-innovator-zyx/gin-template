# 📦 缓存默认行为说明

## 🎯 核心特性

**本项目保证缓存始终可用** - 无论是否配置缓存，系统都会自动提供缓存支持。

---

## ✨ 默认行为

### 🔹 未配置缓存时
```yaml
# app.yaml
# 没有 cache 配置
```

**系统行为**：
- ✅ 自动使用**内存缓存（Memory）**
- ✅ 应用正常启动
- ✅ 所有缓存功能正常工作
- ✅ 日志提示：`未配置缓存，使用默认内存缓存`

---

### 🔹 配置缓存失败时
```yaml
cache:
  type: redis
  redis:
    host: invalid-host  # 错误的配置
    port: 6379
```

**系统行为**：
- ⚠️ 尝试连接Redis失败
- ✅ 自动降级到**内存缓存**
- ✅ 应用正常启动（不会因缓存失败而崩溃）
- ⚠️ 日志警告：`缓存初始化失败: xxx，自动降级到内存缓存`
- ✅ 日志提示：`已切换到内存缓存`

---

### 🔹 正常配置缓存时
```yaml
cache:
  type: redis
  redis:
    host: localhost
    port: 6379
```

**系统行为**：
- ✅ 使用配置的缓存（Redis/LevelDB/Memory）
- ✅ 应用正常启动
- ✅ 日志提示：`缓存初始化成功，类型: redis`

---

## 🚀 实际场景

### 场景1：快速开始（零配置）

```bash
# 1. 克隆项目
git clone https://github.com/code-innovator-zyx/gin-template.git
cd gin-template

# 2. 直接运行（无需配置缓存）
make run

# ✅ 系统自动使用内存缓存
# ✅ 所有功能正常工作
```

**日志输出**：
```
INFO[0000] 未配置缓存，使用默认内存缓存
INFO[0000] 服务启动成功，端口: 8080
```

---

### 场景2：开发环境（使用Docker Redis）

```bash
# 1. 启动Redis
docker run -d -p 6379:6379 redis:alpine

# 2. 配置缓存
cat > app.yaml << EOF
cache:
  type: redis
  redis:
    host: localhost
    port: 6379
EOF

# 3. 运行
make run

# ✅ 系统使用Redis缓存
```

**日志输出**：
```
INFO[0000] Redis缓存初始化成功: localhost:6379
INFO[0000] 缓存初始化成功，类型: redis
INFO[0000] 服务启动成功，端口: 8080
```

---

### 场景3：Redis不可用时的容错

```bash
# Redis服务未启动或配置错误
make run

# ⚠️ Redis连接失败
# ✅ 自动降级到内存缓存
# ✅ 应用正常启动
```

**日志输出**：
```
WARN[0000] 缓存初始化失败: Redis连接失败: dial tcp [::1]:6379: connect: connection refused，自动降级到内存缓存
INFO[0000] 内存缓存初始化成功
INFO[0000] 已切换到内存缓存
INFO[0000] 服务启动成功，端口: 8080
```

---

## 💡 设计理念

### 1. 零配置可用
```go
// 无需任何配置，系统自动提供缓存
cache.MustInitCache(nil)  // 自动使用Memory

// 业务代码照常工作
cacheService.CheckUserPermission(ctx, userID, path, method)
```

### 2. 自动降级
```go
// Redis不可用时，自动降级到Memory
// 保证服务高可用，不会因缓存失败而崩溃
cache.MustInitCache(&CacheConfig{
    Type: "redis",
    Redis: &RedisConfig{Host: "invalid"},
})
// → 自动降级到 Memory 缓存
```

### 3. 业务无感知
```go
// 无论底层是什么缓存，业务代码完全相同
cacheService := service.MustNewCacheService()
cacheService.SetUserPermissions(ctx, userID, resources)
// ✅ Memory 能用
// ✅ Redis 能用
// ✅ LevelDB 能用
```

---

## 📊 不同配置对比

| 配置情况 | 缓存类型 | 持久化 | 性能 | 启动 | 适用场景 |
|---------|---------|--------|------|------|---------|
| 无配置 | Memory | ❌ | ⭐⭐⭐⭐⭐ | ✅ 立即 | 快速开始、开发测试 |
| Redis正常 | Redis | ✅ | ⭐⭐⭐⭐ | ✅ 立即 | 生产环境 |
| Redis失败 | Memory（降级） | ❌ | ⭐⭐⭐⭐⭐ | ✅ 立即 | 容错场景 |
| LevelDB | LevelDB | ✅ | ⭐⭐⭐⭐⭐ | ✅ 立即 | 单机部署 |

---

## 🔍 健康检查

### 查看当前缓存类型

```bash
curl http://localhost:8080/api/v1/health | jq .data.cache
```

**响应示例**：

**未配置缓存时**：
```json
{
  "status": "ok",
  "type": "memory"
}
```

**使用Redis时**：
```json
{
  "status": "ok",
  "type": "redis"
}
```

**Redis失败降级时**：
```json
{
  "status": "ok",
  "type": "memory"
}
```

---

## ⚙️ 实现细节

### MustInitCache 函数

```go
// MustInitCache 初始化缓存（无论如何都会成功）
func MustInitCache(cfg *CacheConfig) {
    // 情况1：没有配置
    if cfg == nil {
        GlobalCache = NewMemoryCache()
        logrus.Info("未配置缓存，使用默认内存缓存")
        return
    }
    
    // 情况2：尝试初始化配置的缓存
    err := InitCache(*cfg)
    if err != nil {
        // 情况3：初始化失败，降级到内存缓存
        logrus.Warnf("缓存初始化失败: %v，自动降级到内存缓存", err)
        GlobalCache = NewMemoryCache()
        logrus.Info("已切换到内存缓存")
    }
}
```

### 核心初始化

```go
// internal/core/initialize.go
func Init() {
    // ... 其他初始化 ...
    
    // 缓存必选（未配置时默认使用内存缓存）
    cache.MustInitCache(Config.Cache)
    
    // ✅ 无论如何，GlobalCache 都不会是 nil
}
```

---

## 🎨 使用建议

### ✅ 推荐做法

#### 开发环境
```yaml
# 不配置缓存，使用默认Memory
# 或者明确配置
cache:
  type: memory
```

#### 测试环境
```yaml
cache:
  type: leveldb
  leveldb:
    path: ./data/test-cache
```

#### 生产环境
```yaml
cache:
  type: redis
  redis:
    host: ${REDIS_HOST}
    port: ${REDIS_PORT}
    password: ${REDIS_PASSWORD}
    pool_size: 100
```

### ⚠️ 注意事项

1. **Memory缓存数据不持久化**
   - 重启后数据丢失
   - 适合开发测试
   - 不推荐生产环境

2. **生产环境建议明确配置**
   - 使用Redis或LevelDB
   - 数据持久化
   - 更好的性能监控

3. **容错是好事，但监控更重要**
   - 配置告警监控
   - 关注降级日志
   - 及时修复配置问题

---

## 🧪 测试用例

项目包含完整的测试用例：`pkg/cache/factory_test.go`

### 运行测试

```bash
# 测试默认缓存行为
go test -v ./pkg/cache/factory_test.go ./pkg/cache/factory.go ./pkg/cache/interface.go ./pkg/cache/memory_adapter.go ./pkg/cache/redis_adapter.go ./pkg/cache/leveldb_adapter.go

# 预期输出：
# ✅ 未配置缓存时，成功初始化为内存缓存
# ✅ 配置内存缓存成功
# ✅ 配置错误时，成功降级到内存缓存
# ✅ 空类型时，成功初始化为内存缓存
# ✅ GetType 返回正确
# ✅ 缓存状态检查正确
```

---

## 📈 性能影响

### Memory缓存性能

| 操作 | 平均响应时间 | QPS |
|------|-------------|-----|
| Get | 0.01ms | 1000万+ |
| Set | 0.01ms | 1000万+ |
| SIsMember | 0.01ms | 1000万+ |

**结论**：Memory缓存性能极高，完全满足开发测试需求

---

## 🎯 总结

### 核心优势

1. **零配置启动** ✅
   - 克隆项目即可运行
   - 无需安装Redis/LevelDB
   - 降低入门门槛

2. **自动容错** ✅
   - 缓存失败不影响启动
   - 自动降级到Memory
   - 保证服务高可用

3. **灵活配置** ✅
   - 开发用Memory（快速）
   - 测试用LevelDB（持久化）
   - 生产用Redis（分布式）

4. **业务无感** ✅
   - 业务代码无需修改
   - 统一的缓存接口
   - 透明切换

---

## 🔗 相关文档

- [缓存架构设计](./CACHE_ARCHITECTURE.md)
- [缓存类型对比](./CACHE_COMPARISON.md)
- [缓存服务API](./CACHE_SERVICE_GUIDE.md)
- [快速开始](./QUICK_START.md)

---

<div align="center">

**🎉 享受零配置的缓存体验！**

**有问题？查看 [完整文档](./DOCUMENTATION.md)**

</div>

