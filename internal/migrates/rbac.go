package migrates

import (
	"gin-admin/internal/model/rbac"
)

/*
 * @Author: zouyx
 * @Email: 1003941268@qq.com
 * @Date:   2025 2025/12/5 下午6:27
 * @Package: RBAC 模块模型注册
 */

// init 函数在包被导入时自动执行，自动注册所有 RBAC 相关的模型
func init() {
	RegisterGroup("rbac",
		&rbac.User{},
		&rbac.Role{},
		&rbac.Permission{},
		&rbac.Resource{},
	)
}
