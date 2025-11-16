package rbac

import (
	"gin-template/internal/service/rbac"
	"gin-template/pkg/response"
	"github.com/gin-gonic/gin"
)

/*
* @Author: zouyx
* @Email: zouyx@knowsec.com
* @Date:   2025/11/16 下午4:40
* @Package:
 */

// GetPermissions godoc
// @Summary 获取权限列表
// @Description 获取系统中所有权限的列表
// @Tags RBAC-权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]rbac.Permission} "成功获取权限列表"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /permissions [get]
func GetPermissions(c *gin.Context) {
	permissions, err := rbac.NewPermissionService(c.Request.Context()).FindAll(0)
	if err != nil {
		response.InternalServerError(c, "获取权限列表失败")
		return
	}
	response.Success(c, permissions)
}
