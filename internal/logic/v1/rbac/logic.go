package rbac

import (
	"gin-template/internal/model/rbac"
	"gin-template/internal/service"
	"gin-template/pkg/response"
	"net/http"
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
// @Success 200 {object} response.Response{data=[]rbac.Role} "成功获取角色列表"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /roles [get]
func GetRoles(c *gin.Context) {
	roles, err := service.GetRbacService().GetAllRoles(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "获取角色列表失败")
		return
	}
	response.Success(c, roles)
}

// CreateRole godoc
// @Summary 创建角色
// @Description 创建新的系统角色
// @Tags RBAC-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role body rbac.Role true "角色信息"
// @Success 201 {object} response.Response{data=rbac.Role} "成功创建角色"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /roles [post]
func CreateRole(c *gin.Context) {
	var role rbac.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := service.GetRbacService().CreateRole(c.Request.Context(), &role); err != nil {
		response.InternalServerError(c, "创建角色失败")
		return
	}
	c.JSON(http.StatusCreated, response.Response{
		Code:    0,
		Message: "success",
		Data:    role,
	})
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
	role, err := service.GetRbacService().GetRoleByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "角色不存在")
		return
	}
	response.Success(c, role)
}

// UpdateRole godoc
// @Summary 更新角色
// @Description 根据ID更新角色信息
// @Tags RBAC-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "角色ID"
// @Param role body rbac.Role true "角色信息"
// @Success 200 {object} response.Response{data=rbac.Role} "成功更新角色"
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
	role, err := service.GetRbacService().GetRoleByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "角色不存在")
		return
	}
	if err := c.ShouldBindJSON(role); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := service.GetRbacService().UpdateRole(c.Request.Context(), role); err != nil {
		response.InternalServerError(c, "更新角色失败")
		return
	}
	response.Success(c, role)
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
	if err := service.GetRbacService().DeleteRole(c.Request.Context(), uint(id)); err != nil {
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
	permissions, err := service.GetRbacService().GetAllPermissions(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "获取权限列表失败")
		return
	}
	response.Success(c, permissions)
}

// CreatePermission godoc
// @Summary 创建权限
// @Description 创建新的系统权限
// @Tags RBAC-权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param permission body rbac.Permission true "权限信息"
// @Success 201 {object} response.Response{data=rbac.Permission} "成功创建权限"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /permissions [post]
func CreatePermission(c *gin.Context) {
	var permission rbac.Permission
	if err := c.ShouldBindJSON(&permission); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := service.GetRbacService().CreatePermission(c.Request.Context(), &permission); err != nil {
		response.InternalServerError(c, "创建权限失败")
		return
	}
	response.Created(c, permission)
}

// GetResources godoc
// @Summary 获取资源列表
// @Description 获取系统中所有资源的列表
// @Tags RBAC-资源管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]rbac.Resource} "成功获取资源列表"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /resources [get]
func GetResources(c *gin.Context) {
	resources, err := service.GetRbacService().GetAllResources(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "获取资源列表失败")
		return
	}
	response.Success(c, resources)
}

// GetUserRoles godoc
// @Summary 获取用户角色
// @Description 获取指定用户的所有角色
// @Tags RBAC-用户角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=[]rbac.UserRole} "成功获取用户角色"
// @Failure 400 {object} response.Response "无效的用户ID"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /users/{id}/roles [get]
func GetUserRoles(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	userRoles, err := service.GetRbacService().GetUserRoleRelations(c.Request.Context(), uint(userID))
	if err != nil {
		response.InternalServerError(c, "获取用户角色失败")
		return
	}

	response.Success(c, userRoles)
}

// AssignRoleToUser godoc
// @Summary 分配角色给用户
// @Description 为指定用户分配角色
// @Tags RBAC-用户角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userRole body rbac.UserRole true "用户角色信息"
// @Success 201 {object} response.Response{data=rbac.UserRole} "成功分配角色"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /user-roles [post]
func AssignRoleToUser(c *gin.Context) {
	var userRole rbac.UserRole
	if err := c.ShouldBindJSON(&userRole); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := service.GetRbacService().CreateUserRole(c.Request.Context(), &userRole); err != nil {
		response.InternalServerError(c, "分配角色失败")
		return
	}

	response.Created(c, userRole)
}

// RemoveRoleFromUser godoc
// @Summary 从用户移除角色
// @Description 从指定用户移除指定角色
// @Tags RBAC-用户角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path int true "用户ID"
// @Param role_id path int true "角色ID"
// @Success 204 {object} response.Response "成功移除角色"
// @Failure 400 {object} response.Response "无效的用户ID或角色ID"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /users/{user_id}/roles/{role_id} [delete]
func RemoveRoleFromUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	roleID, err := strconv.ParseUint(c.Param("role_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	if err := service.GetRbacService().RemoveRoleFromUser(c.Request.Context(), uint(userID), uint(roleID)); err != nil {
		response.InternalServerError(c, "移除角色失败")
		return
	}

	response.NoContent(c)
}

// AssignPermissionToRole godoc
// @Summary 分配权限给角色
// @Description 为指定角色分配权限
// @Tags RBAC-角色权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param rolePermission body rbac.RolePermission true "角色权限信息"
// @Success 201 {object} response.Response{data=rbac.RolePermission} "成功分配权限"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /role-permissions [post]
func AssignPermissionToRole(c *gin.Context) {
	var rolePermission rbac.RolePermission
	if err := c.ShouldBindJSON(&rolePermission); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := service.GetRbacService().CreateRolePermission(c.Request.Context(), &rolePermission); err != nil {
		response.InternalServerError(c, "分配权限失败")
		return
	}

	response.Created(c, rolePermission)
}

// RemovePermissionFromRole godoc
// @Summary 从角色移除权限
// @Description 从指定角色移除指定权限
// @Tags RBAC-角色权限管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role_id path int true "角色ID"
// @Param permission_id path int true "权限ID"
// @Success 204 {object} response.Response "成功移除权限"
// @Failure 400 {object} response.Response "无效的角色ID或权限ID"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /roles/{role_id}/permissions/{permission_id} [delete]
func RemovePermissionFromRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("role_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	permissionID, err := strconv.ParseUint(c.Param("permission_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的权限ID")
		return
	}

	if err := service.GetRbacService().RemovePermissionFromRole(c.Request.Context(), uint(roleID), uint(permissionID)); err != nil {
		response.InternalServerError(c, "移除权限失败")
		return
	}

	response.NoContent(c)
}
