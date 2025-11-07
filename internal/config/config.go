package config

import (
	"fmt"
	"gin-template/pkg/cache"
	"gin-template/pkg/logger"
	"gin-template/pkg/orm"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"time"
)

type AppConfig struct {
	App    App           `mapstructure:"app" validate:"required"`
	Server Server        `mapstructure:"server" validate:"required"`
	Logger logger.Config `mapstructure:"logger" validate:"required"`
	Jwt    *Jwt          `mapstructure:"jwt" validate:"required"`
	RBAC   *RBACConfig   `mapstructure:"rbac" validate:"required"`
	// 选填的配置
	Database *orm.Config        `mapstructure:"database" validate:"omitempty"`
	Cache    *cache.CacheConfig `mapstructure:"cache" validate:"omitempty"`
}

func (a AppConfig) validate() error {
	validate := validator.New()
	return validate.Struct(a)
}

// App 应用基本信息
type App struct {
	Name          string `mapstructure:"name" validate:"required"`
	Version       string `mapstructure:"version" validate:"required"`
	EnableSwagger bool   `mapstructure:"enable_swagger" validate:"omitempty"`
	Env           string `mapstructure:"env" validate:"required,oneof=dev test prod"`
}

// Server HTTP服务器配置
type Server struct {
	Port         int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"required,gt=0"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required,gt=0"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" validate:"required,gt=0"`
}

// jwt 配置
type Jwt struct {
	Secret               string        `mapstructure:"secret" validate:"required"`
	AccessTokenExpire    time.Duration `mapstructure:"access_token_expire" validate:"required,min=60s"`   // Access Token 过期时间（秒），至少 60 秒
	RefreshTokenExpire   time.Duration `mapstructure:"refresh_token_expire" validate:"required,min=600s"` // Refresh Token 过期时间（秒），至少 600 秒
	Issuer               string        `mapstructure:"issuer" validate:"required"`                        // 签发者
	MaxRefreshCount      int           `mapstructure:"max_refresh_count" validate:"omitempty"`            // 单个 Refresh Token 最大刷新次数（0为不限制）
	EnableBlacklist      bool          `mapstructure:"enable_blacklist" validate:"omitempty"`             // 是否启用黑名单
	BlacklistGracePeriod time.Duration `mapstructure:"blacklist_grace_period" validate:"omitempty"`       // 黑名单宽限期（秒）
}

// RBACConfig RBAC权限系统配置
type RBACConfig struct {
	// 是否启用自动初始化
	EnableAutoInit bool `mapstructure:"enable_auto_init" validate:"omitempty"`
	// 默认管理员配置
	AdminUser AdminUserConfig `mapstructure:"admin_user" validate:"omitempty"`
	// 默认角色配置
	AdminRole AdminRoleConfig `mapstructure:"admin_role" validate:"omitempty"`
}

// AdminUserConfig 默认管理员用户配置
type AdminUserConfig struct {
	Username string `mapstructure:"username" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	Email    string `mapstructure:"email" validate:"required,email"`
}

// AdminRoleConfig 默认管理员角色配置
type AdminRoleConfig struct {
	Name        string `mapstructure:"name" validate:"required"`
	Description string `mapstructure:"description" validate:"omitempty"`
}

// Init 初始化配置
func Init() (*AppConfig, error) {
	// 初始化Viper
	viper.SetConfigName("app")     // 配置文件名(无扩展名)
	viper.SetConfigType("yaml")    // 配置文件类型
	viper.AddConfigPath("config/") // 配置文件路径
	viper.AddConfigPath("./")      // 也可以在当前目录查找

	// 支持从环境变量读取配置
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP") // 环境变量前缀

	var config AppConfig
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if err := config.validate(); err != nil {
		return nil, err
	}
	config.Logger.FilePath = fmt.Sprintf("%s/%s.log", config.Logger.FilePath, config.App.Name)
	// todo 增加watch 热加载修改基础信息
	return &config, nil
}
