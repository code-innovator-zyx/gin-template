package rbac

import (
	"fmt"
	"gin-template/internal/model/rbac"
	"gin-template/internal/service"
	types "gin-template/internal/types/rbac"
	"gin-template/pkg/response"
	"gin-template/pkg/transaction"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	result, err := service.GetRbacService().ListRoles(c.Request.Context(), request)
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
	resources, err := service.GetRbacService().GetResources(c.Request.Context(), request.Resources)
	if err != nil {
		response.InternalServerError(c, err.Error())
	}
	role := rbac.Role{
		Name:        request.Name,
		Description: request.Description,
		Resources:   resources,
	}
	if err = service.GetRbacService().CreateRole(c.Request.Context(), &role); err != nil {
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
	role, err := service.GetRbacService().GetRoleByID(c.Request.Context(), uint(id), "Resources")
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

	ctx := c.Request.Context()
	role, err := service.GetRbacService().GetRoleByID(ctx, uint(id))
	if err != nil {
		response.NotFound(c, "角色不存在")
		return
	}

	// 更新角色名称（检查重复）
	if request.Name != "" && request.Name != role.Name {
		exists, err := service.GetRbacService().CheckRoleNameExists(ctx, request.Name, role.ID)
		if err != nil {
			response.InternalServerError(c, "检查角色名称失败")
			return
		}
		if exists {
			response.BadRequest(c, "角色名称已存在")
			return
		}
		role.Name = request.Name
	}

	// 更新角色描述
	if request.Description != "" {
		role.Description = request.Description
	}

	roleID := uint(id)

	// 更新基础信息 + 更新资源绑定 + 清除权限缓存
	err = transaction.WithTransaction(ctx, "UpdateRole",
		// 执行事务
		func(tx *gorm.DB) error {
			// 1. 更新角色基础信息
			if err = service.GetRbacService().UpdateRole(ctx, tx, role); err != nil {
				return fmt.Errorf("更新角色基础信息失败: %w", err)
			}

			// 2. 更新角色资源绑定
			if len(request.Resources) > 0 {
				if err = service.GetRbacService().UpdateRoleResources(ctx, tx, role, request.Resources); err != nil {
					return fmt.Errorf("更新角色资源绑定失败: %w", err)
				}
			}

			// 3. 移除所有相关用户的权限
			// 事务提交后 清除所有拥有该角色的用户的权限缓存
			userIDs, err := service.GetRbacService().GetUsersWithRole(ctx, roleID)
			if err != nil {
				logrus.Errorf("获取角色用户列表失败: %v", err)
				return err
			}
			if len(userIDs) > 0 {
				if err := service.GetCacheService().ClearMultipleUsersPermissions(ctx, userIDs); err != nil {
					logrus.Errorf("批量清除用户权限缓存失败: %v", err)
					return err
				}
				logrus.Infof("已清除 %d 个用户的权限缓存", len(userIDs))
			}
			return nil
		})
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

// AssignResourceToRole godoc
// @Summary 分配资源给角色
// @Description 为指定角色分配资源（细粒度权限控制）
// @Tags RBAC-角色资源管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role_id path int true "角色ID"
// @Param resource_id path int true "资源ID"
// @Success 201 {object} response.Response "成功分配资源"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /roles/{role_id}/resources/{resource_id} [post]
func AssignResourceToRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("role_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	resourceID, err := strconv.ParseUint(c.Param("resource_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的资源ID")
		return
	}

	if err := service.GetRbacService().AssignResourceToRole(c.Request.Context(), uint(roleID), uint(resourceID)); err != nil {
		response.InternalServerError(c, "分配资源失败")
		return
	}

	response.Created(c, nil)
}

// RemoveResourceFromRole godoc
// @Summary 从角色移除资源
// @Description 从指定角色移除指定资源
// @Tags RBAC-角色资源管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role_id path int true "角色ID"
// @Param resource_id path int true "资源ID"
// @Success 204 {object} response.Response "成功移除资源"
// @Failure 400 {object} response.Response "无效的角色ID或资源ID"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /roles/{role_id}/resources/{resource_id} [delete]
func RemoveResourceFromRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("role_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	resourceID, err := strconv.ParseUint(c.Param("resource_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的资源ID")
		return
	}

	if err := service.GetRbacService().RemoveResourceFromRole(c.Request.Context(), uint(roleID), uint(resourceID)); err != nil {
		response.InternalServerError(c, "移除资源失败")
		return
	}

	response.NoContent(c)
}
