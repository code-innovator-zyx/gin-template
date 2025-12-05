package uploader

import (
	"context"
	"fmt"
	_interface "gin-admin/pkg/interface"
	"gin-admin/pkg/utils"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// LocalUploader 本地文件存储上传器
type LocalUploader struct {
	config Config
	port   int
}

// NewLocalUploader 创建本地上传器实例
func NewLocalUploader(cfg Config, port int) *LocalUploader {
	return &LocalUploader{
		config: cfg,
		port:   port,
	}
}

// Upload 上传文件到本地存储
func (u *LocalUploader) Upload(ctx context.Context, fileHeader *multipart.FileHeader) (*_interface.UploadResult, error) {

	// 验证文件扩展名
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != "" && ext[0] == '.' {
		ext = ext[1:] // 移除前导点
	}

	if !u.isAllowedExtension(ext) {
		return nil, fmt.Errorf("不支持的文件格式: %s, 允许的格式: %v", ext, u.config.AllowedExtensions)
	}

	// 验证文件大小
	maxSize := int64(u.config.MaxSizeMB) * 1024 * 1024
	if fileHeader.Size <= 0 {
		return nil, fmt.Errorf("无效图片")
	}
	if fileHeader.Size > maxSize {
		return nil, fmt.Errorf("文件大小超过限制: %d MB", u.config.MaxSizeMB)
	}

	// 生成保存路径 (uploads/images/YYYY-MM-DD/)
	dateStr := time.Now().Format("2006-01-02")
	savePath := filepath.Join(u.config.Local.BaseDir, dateStr)

	// 创建目录
	if err := os.MkdirAll(savePath, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 生成唯一文件名 (UUID.ext)
	filename := fmt.Sprintf("%s.%s", uuid.New().String(), ext)
	fullPath := filepath.Join(savePath, filename)
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	size, err := io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	return &_interface.UploadResult{
		Url:      u.GetURL(fullPath),
		Filename: filename,
		Size:     size,
	}, nil
}

// GetURL 根据相对路径获取完整访问URL
// 实现 Uploader 接口
// 使用本机IP地址动态生成URL
func (u *LocalUploader) GetURL(path string) string {
	if path == "" {
		return ""
	}
	var baseURL string
	if u.config.Local.BaseUrl != nil {
		baseURL = *u.config.Local.BaseUrl
	} else {
		baseURL = fmt.Sprintf(
			"http://%s:%d",
			utils.GetLocalIP(),
			u.port,
		)
	}
	// 获取服务器端口
	return fmt.Sprintf("%s/%s", baseURL, filepath.ToSlash(path))
}

// ParseUrl 从完整URL或本地文件路径中提取相对于 uploads 目录的相对路径
// 支持的输入：
// 1. 完整 URL（http://... 或 https://...）
// 2. 本地绝对路径（/var/www/uploads/...）
// 3. 相对路径（uploads/... 或 YYYY-MM-DD/xxx）
func (u *LocalUploader) ParseUrl(raw string) string {
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return strings.TrimLeft(raw, "/")
	}

	path := filepath.ToSlash(parsed.Path)

	return strings.TrimLeft(path, "/")
}

// isAllowedExtension 检查文件扩展名是否允许
func (u *LocalUploader) isAllowedExtension(ext string) bool {
	if len(u.config.AllowedExtensions) == 0 {
		return true // 如果没有配置限制，则允许所有
	}

	for _, allowed := range u.config.AllowedExtensions {
		if strings.EqualFold(ext, allowed) {
			return true
		}
	}
	return false
}
