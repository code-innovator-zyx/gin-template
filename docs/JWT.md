# JWT 认证指南

## 概述

项目实现了企业级 JWT 双令牌认证系统：

- ✅ **Access Token** - 短期令牌，用于API访问（1小时）
- ✅ **Refresh Token** - 长期令牌，用于刷新（7天）
- ✅ **Token 黑名单** - 支持令牌撤销（登出）
- ✅ **自动刷新** - Token 过期自动续期
- ✅ **会话管理** - 支持多设备登录管理

## 配置

```yaml
jwt:
  secret: "your-secret-key-change-this-in-production"  # 生产环境必须修改
  access_token_expire: 3600      # Access Token 1小时
  refresh_token_expire: 604800   # Refresh Token 7天
  issuer: "gin-admin"
  max_refresh_count: 10          # 单个 Refresh Token 最大刷新次数
  enable_blacklist: true         # 启用黑名单（需要缓存支持）
  blacklist_grace_period: 300    # 黑名单宽限期 5分钟
```

## API 使用

### 登录

```bash
POST /api/v1/user/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password123"
}
```

**响应：**

```json
{
  "code": 200,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 3600
  }
}
```

### 访问受保护的API

```bash
GET /api/v1/user/profile
Authorization: Bearer <access_token>
```

### 刷新 Token

```bash
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 登出

```bash
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## 客户端集成

### JavaScript 示例

```javascript
// 登录
const login = async (username, password) => {
  const response = await fetch('/api/v1/user/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password })
  });
  
  const data = await response.json();
  
  // 保存 Token
  localStorage.setItem('access_token', data.data.access_token);
  localStorage.setItem('refresh_token', data.data.refresh_token);
};

// 请求带 Token
const fetchWithAuth = async (url, options = {}) => {
  const token = localStorage.getItem('access_token');
  
  options.headers = {
    ...options.headers,
    'Authorization': `Bearer ${token}`
  };
  
  let response = await fetch(url, options);
  
  // Token 过期，自动刷新
  if (response.status === 401) {
    const refreshToken = localStorage.getItem('refresh_token');
    
    const refreshResponse = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken })
    });
    
    if (refreshResponse.ok) {
      const data = await refreshResponse.json();
      localStorage.setItem('access_token', data.data.access_token);
      localStorage.setItem('refresh_token', data.data.refresh_token);
      
      // 重试原请求
      options.headers['Authorization'] = `Bearer ${data.data.access_token}`;
      response = await fetch(url, options);
    } else {
      // 刷新失败，跳转登录
      window.location.href = '/login';
    }
  }
  
  return response;
};

// 登出
const logout = async () => {
  const token = localStorage.getItem('access_token');
  const refreshToken = localStorage.getItem('refresh_token');
  
  await fetch('/api/v1/auth/logout', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ refresh_token: refreshToken })
  });
  
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
  window.location.href = '/login';
};
```

## 代码使用

### 生成令牌

```go
import "gin-admin/pkg/utils"

jwtManager := utils.GetJWTManager()

// 生成令牌对
tokenPair, err := jwtManager.GenerateTokenPair(userID, username, email)

// 返回给客户端
response.Success(c, gin.H{
    "access_token":  tokenPair.AccessToken,
    "refresh_token": tokenPair.RefreshToken,
    "token_type":    tokenPair.TokenType,
    "expires_in":    tokenPair.ExpiresIn,
})
```

### 验证令牌

```go
// 在中间件中验证
claims, err := jwtManager.ParseAccessToken(ctx, accessToken)
if err != nil {
    // Token 无效或过期
    return
}

// 使用用户信息
userID := claims.UserID
username := claims.Username
```

### 刷新令牌

```go
// 使用 Refresh Token 获取新的令牌对
newTokenPair, err := jwtManager.RefreshToken(ctx, refreshToken)
```

### 撤销令牌

```go
// 登出时撤销 Token
err := jwtManager.RevokeToken(ctx, accessToken)
err = jwtManager.RevokeToken(ctx, refreshToken)
```

## 安全建议

1. **生产环境必须修改密钥**，使用至少 32 字节的强随机密钥
2. **使用 HTTPS**，避免 Token 在网络中明文传输
3. **Token 存储**：
   - 推荐使用 httpOnly Cookie
   - 或使用 localStorage（注意 XSS 风险）
   - 不要在 URL 中传递 Token
4. **启用黑名单**功能，支持主动撤销令牌
5. **定期轮换密钥**

## 常见问题

### Q: 为什么需要 Refresh Token？

A: Access Token 有效期短（1小时），如果每次过期都要求用户重新登录，体验很差。Refresh Token 可以在 Access Token 过期后自动续期，平衡了安全性和用户体验。

### Q: Token 可以手动续期吗？

A: 可以。在 Access Token 即将过期时，调用刷新接口获取新的 Token。

### Q: 如何实现"记住我"功能？

A: 增加 Refresh Token 的有效期（如 30 天），并在前端保存 Refresh Token。

### Q: 如何踢出其他设备？

A: 使用会话管理功能，调用撤销所有会话的接口。

## 错误处理

| 错误 | 说明 | 处理方式 |
|------|------|----------|
| `ErrTokenExpired` | Token 已过期 | 使用 Refresh Token 刷新 |
| `ErrTokenBlacklisted` | Token 已被撤销 | 重新登录 |
| `ErrRefreshTokenExpired` | Refresh Token 已过期 | 重新登录 |
| `ErrRefreshLimitReached` | 刷新次数已达上限 | 重新登录 |

