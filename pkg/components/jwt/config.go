package jwt

import "time"

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/19 上午11:50
* @Package:
 */

// Config jwt 配置
type Config struct {
	Secret             string        `mapstructure:"secret" validate:"required"`
	AccessTokenExpire  time.Duration `mapstructure:"access_token_expire" validate:"required,min=30s"`   // Access Token 过期时间（秒），至少 60 秒
	RefreshTokenExpire time.Duration `mapstructure:"refresh_token_expire" validate:"required,min=600s"` // Refresh Token 过期时间（秒），至少 600 秒
	Issuer             string        `mapstructure:"issuer" validate:"required"`                        // 签发者
}
