# 企业级 JWT 鉴权系统设计文档

## 1. 设计目标

- 高性能无状态 Access Token
- 高安全性 Refresh Token（Rotation）
- 可控会话体系（支持踢人、退出所有设备、设备管理）
- 无黑名单设计，性能更优
- 支持分布式

## 2. Token 体系

### Access Token

- 无状态 JWT
- 有效期短（10--30 分钟）
- 不查 Redis
- 内含 userID / username / deviceID / sessionID / tokenType=access

### Refresh Token

- 用于刷新 Access Token
- 有效期长（7--30 天）
- 需要查 Redis（会话体系）
- 存储哈希 refreshTokenHash

### Refresh Token Rotation

- 每次刷新都会生成新的 Refresh Token
- 更新 refresh hash
- 旧 Refresh Token 自动失效

## 3. Session 体系

### Redis Key 结构

jwt:session:<sessionID>
jwt:user:<uid>:sessions → set
### SessionInfo

sessionID
userID
username
deviceID
refreshTokenHash
expiresAt
revoked
### 用途

- 单设备退出
- 全设备退出
- 多端登录管理
- 防止 Refresh Token 被盗用
- 查看当前在线设备

## 4. 登录流程

1. 用户登录
2. 生成 AccessToken & RefreshToken
3. 保存 SessionInfo 到 Redis
4. 返回 TokenPair

## 5. 刷新 Token 流程

1. 校验 refresh token
2. 查询 session
3. 校验 refresh hash
4. Rotation：重新生成 refresh token
5. 更新 session
6. 颁发新的 access token

## 6. 退出登录流程

- RemoveSession → 删除该 session 数据

## 7. Access Token 校验

1. 解析 JWT
2. 校验 TokenType = access
3. 查 Redis 校验 session 是否存在/未撤销
4. 返回 claims

## 8. 安全策略

- Token Rotation
- Session 过期与主动清理
- 防刷策略（限制 refresh 频率）
- HMAC-SHA256
- Refresh Token 哈希存储

## 9. 扩展能力

- 多端登录限制
- 查看已登录设备列表
- 踢人（删除 Session）
- 异常登录提醒
