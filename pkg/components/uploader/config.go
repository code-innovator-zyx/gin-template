package uploader

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025 2025/12/5 上午11:52
* @Package:
 */
type LocalConfig struct {
	// BaseDir 上传目录
	BaseDir string `mapstructure:"base_dir" default:"./uploads"`
	// BaseURL 服务器基础URL 如果不配置，使用本地ID，如果是服务器域名访问，需要配置
	BaseUrl *string `mapstructure:"base_url"`
}

// Config 上传器配置
type Config struct {
	// Local 本地存储配置
	Local *LocalConfig `mapstructure:"local" validate:"omitempty"`
	// todo  add other uploader
	// AllowedExtensions 允许的文件扩展名
	AllowedExtensions []string `mapstructure:"allowed_extensions" default:"jpg,jpeg,png,gif,webp"`
	// MaxSizeMB 最大文件大小（MB）
	MaxSizeMB int `mapstructure:"max_size_mb" default:"10"`
}
