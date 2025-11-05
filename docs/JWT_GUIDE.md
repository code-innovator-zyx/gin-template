# 企业级 JWT 认证系统使用指南

## 概述

本项目实现了完整的企业级 JWT 双令牌认证系统，包括：

- ✅ **Access Token + Refresh Token 双令牌机制**
- ✅ **Token 自动刷新**
- ✅ **Token 黑名单（撤销机制）**
- ✅ **会话管理**
- ✅ **刷新次数限制**
- ✅ **设备管理支持**

---

## 核心概念

### Access Token
- **用途**: 用于访问受保护的 API 资源
- **有效期**: 短期（推荐 15 分钟到 1 小时）
- **特点**: 频繁使用，有效期短，安全性高

### Refresh Token
- **用途**: 用于刷新 Access Token
- **有效期**: 长期（推荐 7 天到 30 天）
- **特点**: 不频繁使用，有效期长，用于续期

---

## 配置说明

在 `app.yaml` 中配置 JWT 参数：

```yaml
jwt:
  # JWT 密钥（生产环境必须使用强密钥！）
  secret: "your-secret-key-change-this-in-production"
  
  # Access Token 过期时间（秒）
  access_token_expire: 3600  # 1小时
  
  # Refresh Token 过期时间（秒）
  refresh_token_expire: 604800  # 7天
  
  # Token 签发者
  issuer: "gin-template"
  
  # 单个 Refresh Token 最大刷新次数（0为不限制）
  max_refresh_count: 10
  
  # 是否启用 Token 黑名单（需要缓存支持）
  enable_blacklist: true
  
  # 黑名单宽限期（秒）
  blacklist_grace_period: 300  # 5分钟
```

---

## 基本使用

### 1. 用户登录 - 生成令牌对

```go
package main

import (
    "gin-template/pkg/utils"
    "context"
)

func Login(userID uint, username, email string) {
    // 获取 JWT 管理器
    jwtManager := utils.GetJWTManager()
    
    // 生成令牌对（可选：传入设备ID）
    tokenPair, err := jwtManager.GenerateTokenPair(
        userID,
        username,
        email,
        "device-id-123", // 可选的设备ID
    )
    if err != nil {
        // 处理错误
        return
    }
    
    // 返回给客户端
    response := map[string]interface{}{
        "access_token":  tokenPair.AccessToken,
        "refresh_token": tokenPair.RefreshToken,
        "token_type":    tokenPair.TokenType,    // "Bearer"
        "expires_in":    tokenPair.ExpiresIn,    // 3600
        "expires_at":    tokenPair.ExpiresAt,    // 过期时间点
    }
}
```

### 2. 验证 Access Token

```go
package main

import (
    "gin-template/pkg/utils"
    "context"
)

func ProtectedAPI(ctx context.Context, accessToken string) {
    jwtManager := utils.GetJWTManager()
    
    // 解析并验证 Access Token
    claims, err := jwtManager.ParseAccessToken(ctx, accessToken)
    if err != nil {
        if err == utils.ErrTokenExpired {
            // Token 已过期，提示客户端刷新
            return
        }
        if err == utils.ErrTokenBlacklisted {
            // Token 已被撤销
            return
        }
        // 其他错误
        return
    }
    
    // 使用用户信息
    userID := claims.UserID
    username := claims.Username
    email := claims.Email
    sessionID := claims.SessionID
}
```

### 3. 刷新 Token

```go
package main

import (
    "gin-template/pkg/utils"
    "context"
)

func RefreshToken(ctx context.Context, refreshToken string) {
    jwtManager := utils.GetJWTManager()
    
    // 使用 Refresh Token 获取新的令牌对
    newTokenPair, err := jwtManager.RefreshToken(ctx, refreshToken)
    if err != nil {
        if err == utils.ErrRefreshTokenExpired {
            // Refresh Token 已过期，需要重新登录
            return
        }
        if err == utils.ErrRefreshLimitReached {
            // 刷新次数已达上限
            return
        }
        // 其他错误
        return
    }
    
    // 返回新的令牌对给客户端
    response := map[string]interface{}{
        "access_token":  newTokenPair.AccessToken,
        "refresh_token": newTokenPair.RefreshToken,
        "token_type":    newTokenPair.TokenType,
        "expires_in":    newTokenPair.ExpiresIn,
    }
}
```

### 4. 用户登出 - 撤销 Token

```go
package main

import (
    "gin-template/pkg/utils"
    "context"
)

func Logout(ctx context.Context, accessToken, refreshToken string) {
    jwtManager := utils.GetJWTManager()
    
    // 撤销 Access Token
    _ = jwtManager.RevokeToken(ctx, accessToken)
    
    // 撤销 Refresh Token
    _ = jwtManager.RevokeToken(ctx, refreshToken)
    
    // 或者：撤销用户的所有令牌（登出所有设备）
    // jwtManager.RevokeUserAllTokens(ctx, userID)
}
```

---

## 高级功能

### 会话管理

系统支持追踪用户的所有活跃会话：

```go
package main

import (
    "gin-template/internal/service"
    "context"
)

func ManageSessions(ctx context.Context, userID uint) {
    cacheService := service.MustNewCacheService()
    
    // 获取用户所有会话
    sessions, err := cacheService.GetUserSessions(ctx, userID)
    
    // 撤销特定会话
    err = cacheService.RemoveUserSession(ctx, userID, "session-id")
    
    // 撤销所有会话（登出所有设备）
    err = cacheService.RevokeAllUserSessions(ctx, userID)
}
```

### Token 元数据获取

无需验证即可获取 Token 信息：

```go
package main

import (
    "gin-template/pkg/utils"
)

func GetTokenInfo(tokenString string) {
    jwtManager := utils.GetJWTManager()
    
    // 获取 Token 元数据（不验证签名和有效期）
    metadata, err := jwtManager.GetTokenMetadata(tokenString)
    if err != nil {
        return
    }
    
    userID := metadata.UserID
    username := metadata.Username
    issuedAt := metadata.IssuedAt
    expiresAt := metadata.ExpiresAt
}
```

---

## API 端点示例

### 登录接口

```go
// @Summary 用户登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param data body LoginRequest true "登录信息"
// @Success 200 {object} TokenResponse "登录成功"
// @Router /user/login [post]
func Login(c *gin.Context) {
    // ... 验证用户名密码 ...
    
    jwtManager := utils.GetJWTManager()
    tokenPair, err := jwtManager.GenerateTokenPair(user.ID, user.Username, user.Email)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "生成令牌失败")
        return
    }
    
    response.Success(c, tokenPair)
}
```

### 刷新令牌接口

```go
// @Summary 刷新令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param data body RefreshRequest true "刷新令牌"
// @Success 200 {object} TokenResponse "刷新成功"
// @Router /auth/refresh [post]
func RefreshToken(c *gin.Context) {
    var req RefreshRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "请求参数错误")
        return
    }
    
    jwtManager := utils.GetJWTManager()
    tokenPair, err := jwtManager.RefreshToken(c.Request.Context(), req.RefreshToken)
    if err != nil {
        if err == utils.ErrRefreshTokenExpired {
            response.Error(c, http.StatusUnauthorized, "刷新令牌已过期，请重新登录")
            return
        }
        response.Error(c, http.StatusUnauthorized, "刷新令牌无效")
        return
    }
    
    response.Success(c, tokenPair)
}
```

### 登出接口

```go
// @Summary 用户登出
// @Tags 认证
// @Security ApiKeyAuth
// @Success 200 {object} Response "登出成功"
// @Router /auth/logout [post]
func Logout(c *gin.Context) {
    // 从请求头获取 Access Token
    accessToken := c.GetHeader("Authorization")
    accessToken = strings.TrimPrefix(accessToken, "Bearer ")
    
    // 从请求体获取 Refresh Token（可选）
    var req struct {
        RefreshToken string `json:"refresh_token"`
    }
    c.ShouldBindJSON(&req)
    
    jwtManager := utils.GetJWTManager()
    ctx := c.Request.Context()
    
    // 撤销令牌
    _ = jwtManager.RevokeToken(ctx, accessToken)
    if req.RefreshToken != "" {
        _ = jwtManager.RevokeToken(ctx, req.RefreshToken)
    }
    
    response.Success(c, "登出成功")
}
```

---

## 中间件集成

在中间件中验证 Token：

```go
package middleware

import (
    "gin-template/pkg/utils"
    "gin-template/pkg/response"
    "github.com/gin-gonic/gin"
    "strings"
    "net/http"
)

func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 获取 Authorization 头
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            response.Error(c, http.StatusUnauthorized, "缺少认证令牌")
            c.Abort()
            return
        }
        
        // 解析 Bearer Token
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            response.Error(c, http.StatusUnauthorized, "令牌格式错误")
            c.Abort()
            return
        }
        
        token := parts[1]
        
        // 验证 Token
        jwtManager := utils.GetJWTManager()
        claims, err := jwtManager.ParseAccessToken(c.Request.Context(), token)
        if err != nil {
            if err == utils.ErrTokenExpired {
                response.Error(c, http.StatusUnauthorized, "令牌已过期")
            } else if err == utils.ErrTokenBlacklisted {
                response.Error(c, http.StatusUnauthorized, "令牌已被撤销")
            } else {
                response.Error(c, http.StatusUnauthorized, "无效的令牌")
            }
            c.Abort()
            return
        }
        
        // 将用户信息存入上下文
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("email", claims.Email)
        c.Set("session_id", claims.SessionID)
        
        c.Next()
    }
}
```

---

## 客户端使用指南

### 1. 登录流程

```javascript
// 登录
const response = await fetch('/api/v1/user/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    username: 'user@example.com',
    password: 'password123'
  })
});

const data = await response.json();

// 保存 Token
localStorage.setItem('access_token', data.data.access_token);
localStorage.setItem('refresh_token', data.data.refresh_token);
```

### 2. 请求受保护的 API

```javascript
const accessToken = localStorage.getItem('access_token');

const response = await fetch('/api/v1/protected-resource', {
  method: 'GET',
  headers: {
    'Authorization': `Bearer ${accessToken}`
  }
});
```

### 3. 自动刷新 Token

```javascript
async function apiCall(url, options = {}) {
  let accessToken = localStorage.getItem('access_token');
  
  // 添加 Authorization 头
  options.headers = {
    ...options.headers,
    'Authorization': `Bearer ${accessToken}`
  };
  
  let response = await fetch(url, options);
  
  // 如果 Token 过期，尝试刷新
  if (response.status === 401) {
    const refreshToken = localStorage.getItem('refresh_token');
    
    // 刷新 Token
    const refreshResponse = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken })
    });
    
    if (refreshResponse.ok) {
      const data = await refreshResponse.json();
      
      // 保存新 Token
      localStorage.setItem('access_token', data.data.access_token);
      localStorage.setItem('refresh_token', data.data.refresh_token);
      
      // 重试原请求
      options.headers['Authorization'] = `Bearer ${data.data.access_token}`;
      response = await fetch(url, options);
    } else {
      // 刷新失败，跳转到登录页
      window.location.href = '/login';
      return null;
    }
  }
  
  return response;
}
```

### 4. 登出

```javascript
async function logout() {
  const accessToken = localStorage.getItem('access_token');
  const refreshToken = localStorage.getItem('refresh_token');
  
  await fetch('/api/v1/auth/logout', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      refresh_token: refreshToken
    })
  });
  
  // 清除本地存储
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
  
  // 跳转到登录页
  window.location.href = '/login';
}
```

---

## 错误处理

| 错误 | 说明 | 处理方式 |
|------|------|----------|
| `ErrTokenExpired` | Token 已过期 | 使用 Refresh Token 刷新 |
| `ErrTokenNotValidYet` | Token 尚未生效 | 检查系统时间 |
| `ErrTokenMalformed` | Token 格式错误 | 重新获取 Token |
| `ErrTokenInvalid` | 无效的 Token | 重新登录 |
| `ErrTokenBlacklisted` | Token 已被撤销 | 重新登录 |
| `ErrRefreshTokenExpired` | Refresh Token 已过期 | 重新登录 |
| `ErrRefreshLimitReached` | 刷新次数已达上限 | 重新登录 |
| `ErrInvalidTokenType` | Token 类型错误 | 使用正确的 Token 类型 |

---

## 安全建议

### 1. 密钥管理
- ✅ 使用强随机密钥（至少 32 字节）
- ✅ 生产环境通过环境变量注入密钥
- ✅ 定期轮换密钥

### 2. Token 有效期
- ✅ Access Token: 15分钟 - 1小时
- ✅ Refresh Token: 7天 - 30天
- ✅ 根据业务需求调整

### 3. HTTPS
- ✅ 生产环境必须使用 HTTPS
- ✅ 避免 Token 在网络中明文传输

### 4. 存储
- ✅ 客户端使用 httpOnly Cookie 存储（推荐）
- ✅ 或使用 localStorage（注意 XSS 风险）
- ✅ 不要在 URL 中传递 Token

### 5. 黑名单
- ✅ 启用黑名单功能
- ✅ 配置合适的缓存系统（Redis 推荐）

### 6. 日志和监控
- ✅ 记录 Token 生成和验证日志
- ✅ 监控异常登录行为
- ✅ 设置告警机制

---

## 性能优化

### 1. 缓存策略
- 使用 Redis 作为黑名单存储
- 合理设置 TTL
- 避免频繁查询数据库

### 2. Token 大小
- Claims 中只包含必要信息
- 避免存储大量数据
- 敏感信息不要放入 Token

### 3. 并发处理
- 使用连接池
- 避免锁竞争
- 异步处理非关键操作

---

## 常见问题

### Q: 为什么需要 Refresh Token？
A: Access Token 有效期短，频繁要求用户登录体验不好。Refresh Token 可以在 Access Token 过期后自动续期，平衡了安全性和用户体验。

### Q: Token 黑名单会影响性能吗？
A: 使用 Redis 等高性能缓存，影响很小。可以通过配置 `enable_blacklist` 关闭黑名单功能。

### Q: 如何实现"记住我"功能？
A: 增加 Refresh Token 的有效期（如 30 天），并在前端保存 Refresh Token。

### Q: 如何实现踢出其他设备？
A: 使用会话管理功能，调用 `RevokeAllUserSessions` 撤销用户的所有会话。

### Q: Token 可以手动续期吗？
A: 可以。在 Access Token 即将过期时，调用刷新接口获取新的 Token。

---

## 更新日志

### v1.0.0 (2025-11-05)
- ✅ 实现双令牌机制
- ✅ 支持 Token 刷新
- ✅ 支持 Token 黑名单
- ✅ 支持会话管理
- ✅ 支持刷新次数限制
- ✅ 完整的错误处理

---

## 参考资料

- [JWT 官方文档](https://jwt.io/)
- [RFC 7519 - JSON Web Token](https://tools.ietf.org/html/rfc7519)
- [OAuth 2.0 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [OWASP JWT 安全指南](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)

