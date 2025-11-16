package rbac

import (
	"gin-template/internal/service/rbac"
	types "gin-template/internal/types/rbac"
	"gin-template/pkg/response"
	"github.com/sirupsen/logrus"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetRoles godoc
// @Summary 获取角色列表
// @Description 获取系统中所有角色的列表
// @Tags RBAC-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request query types.ListRoleRequest true "查询参数"
// @Success 200 {object} response.PaginatedResponse{data=[]rbac.Role} "成功获取角色列表"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /roles [get]
func GetRoles(c *gin.Context) {
	// 获取分页参数
	request := types.ListRoleRequest{}
	err := c.ShouldBindQuery(&request)
	if err != nil {
		response.BadRequest(c, err.Error())
	}
	result, err := rbac.NewRoleService(c.Request.Context()).List(request)
	if err != nil {
		response.InternalServerError(c, "获取角色列表失败")
		return
	}
	response.SuccessPage(c, result.List, result.Page, result.PageSize, result.Total)
}

// CreateRole godoc
// @Summary 创建角色
// @Description 创建新的系统角色
// @Tags RBAC-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role body types.CreateRoleRequest true "角色信息"
// @Success 201 {object} response.Response{data=rbac.Role} "成功创建角色"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /roles [post]
func CreateRole(c *gin.Context) {
	var request types.CreateRoleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := rbac.NewRoleService(c.Request.Context()).Create(request); err != nil {
		response.InternalServerError(c, "创建角色失败")
		return
	}
	response.Success(c, nil)
}

// GetRole godoc
// @Summary 获取角色详情
// @Description 根据ID获取角色详细信息
// @Tags RBAC-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "角色ID"
// @Success 200 {object} response.Response{data=rbac.Role} "成功获取角色详情"
// @Failure 400 {object} response.Response "无效的角色ID"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "角色不存在"
// @Router /roles/{id} [get]
func GetRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}
	role, err := rbac.NewRoleService(c.Request.Context()).FindByID(uint(id), "Resources")
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, role)
}

// UpdateRole godoc
// @Summary 更新角色
// @Description 根据ID更新角色信息,修改角色资源
// @Tags RBAC-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "角色ID"
// @Param role body types.UpdateRoleRequest true "角色信息"
// @Success 200 {object} response.Response{data=nil} "成功更新角色"
// @Failure 400 {object} response.Response "无效的角色ID或请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "角色不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /roles/{id} [put]
func UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}
	var request types.UpdateRoleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	request.Id = uint(id)
	err = rbac.NewRoleService(c.Request.Context()).Update(request)
	if err != nil {
		logrus.Errorf("failed to update role: %v", err)
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, nil)
}

// DeleteRole godoc
// @Summary 删除角色
// @Description 根据ID删除角色
// @Tags RBAC-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "角色ID"
// @Success 204 {object} response.Response "成功删除角色"
// @Failure 400 {object} response.Response "无效的角色ID"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /roles/{id} [delete]
func DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}
	if err := rbac.NewResourceService(c.Request.Context()).DeleteByID(uint(id)); err != nil {
		response.InternalServerError(c, "删除角色失败")
		return
	}
	response.NoContent(c)
}

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
