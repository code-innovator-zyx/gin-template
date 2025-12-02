package jwt

import "time"

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/19 上午11:50
* @Package:
 */

// Config JWT配置
type Config struct {
	Secret             string        // 密钥
	Issuer             string        // 发行者
	AccessTokenExpire  time.Duration // 访问令牌过期时间
	RefreshTokenExpire time.Duration // 刷新令牌过期时间
	//MaxRefreshCount      int           // 最大刷新次数限制
	//EnableBlacklist      bool          // 是否启用黑名单
	//BlacklistGracePeriod time.Duration // 黑名单宽限期
	//SessionTimeout       time.Duration // 会话超时时间
}
