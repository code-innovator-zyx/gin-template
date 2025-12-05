package _interface

import (
	"context"
	"mime/multipart"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025 2025/12/5 上午11:44
* @Package: 文件上传器
 */

// UploadResult 上传结果
type UploadResult struct {
	// 访问地址
	Url string `json:"url"`
	// Filename 文件名
	Filename string `json:"filename"`
	// Size 文件大小（字节）
	Size int64 `json:"size"`
}

// IUploader 上传器接口
type IUploader interface {
	// Upload 上传文件
	// fileHeader: 上传的文件
	// Returns: 上传结果和错误
	Upload(ctx context.Context, fileHeader *multipart.FileHeader) (*UploadResult, error)

	// GetURL 根据相对路径获取完整访问URL
	// path: 文件相对路径
	// Returns: 完整访问URL
	GetURL(path string) string

	// ParseUrl 从完整URL中提取相对路径
	// url: 完整的访问URL
	// Returns: 相对路径，如果无法解析则返回原始输入
	ParseUrl(url string) string
}
