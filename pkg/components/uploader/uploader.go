package uploader

import _interface "gin-admin/pkg/interface"

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025 2025/12/5 下午1:17
* @Package:
 */

var uploader _interface.IUploader

// NewUploader 新建一个上传器
func NewUploader(cfg Config, port int) _interface.IUploader {
	ul := NewLocalUploader(cfg, port)

	uploader = ul
	return ul
}
