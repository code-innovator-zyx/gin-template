package config

import (
	"fmt"
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
	// 选填的配置
	Database *orm.Config `mapstructure:"database" validate:"omitempty"`
	Jwt      *Jwt        `mapstructure:"jwt" validate:"omitempty"`
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
	Secret string `mapstructure:"secret" validate:"required"`
	Expire int    `mapstructure:"expire" validate:"required,min=60"` // 至少 60 秒
}

// Init 初始化配置
func Init() (*AppConfig, error) {
	// 初始化Viper
	viper.SetConfigName("app")     // 配置文件名(无扩展名)
	viper.SetConfigType("yaml")    // 配置文件类型
	viper.AddConfigPath("config/") // 配置文件路径
	viper.AddConfigPath("./")      // 也可以在当前目录查找

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
