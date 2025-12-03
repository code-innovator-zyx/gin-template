package rbac

import (
	"fmt"
	rbac2 "gin-admin/internal/model/rbac"
	"gin-admin/internal/services"
	types "gin-admin/internal/types/rbac"
	_interface "gin-admin/pkg/interface"
	"gin-admin/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

// CreateRole godoc
// @Summary 创建角色
// @Description 创建新的系统角色
// @Tags RBAC-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role body types.UpsertRoleRequest true "角色信息"
// @Success 201 {object} response.Response{data=rbac.Role} "成功创建角色"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /roles [post]
func CreateRole(svcCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request types.UpsertRoleRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		exist, err := svcCtx.RoleService.Exists(c.Request.Context(), _interface.WithConditions(map[string]interface{}{"name": request.Name}))
		if err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
		if exist {
			response.Fail(c, http.StatusConflict, "角色名已存在")
			return
		}
		role := &rbac2.Role{
			Name:        request.Name,
			Description: request.Description,
			Status:      request.Status,
		}
		if err = svcCtx.RoleService.Create(c.Request.Context(), role); err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
		response.Created(c, nil)
	}
}

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
func GetRoles(svcCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		request := types.ListRoleRequest{}
		err := c.ShouldBindQuery(&request)
		if err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		result, err := svcCtx.RoleService.FindPage(c.Request.Context(), _interface.WithPagination(request.Page, request.PageSize))
		if err != nil {
			response.Fail(c, 500, "获取角色列表失败")
			return
		}
		response.SuccessPage(c, result.List, result.Page, result.PageSize, result.Total)
	}
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
func GetRole(svcCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.BadRequest(c, "无效的角色ID")
			return
		}
		role, err := svcCtx.RoleService.FindByID(c.Request.Context(), uint(id), _interface.WithPreloads("Resources"))
		if err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
		response.Success(c, role)
	}
}

// UpdateRole godoc
// @Summary 更新角色
// @Description 根据ID更新角色信息,修改角色资源
// @Tags RBAC-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "角色ID"
// @Param role body types.UpsertRoleRequest true "角色信息"
// @Success 200 {object} response.Response{data=nil} "成功更新角色"
// @Failure 400 {object} response.Response "无效的角色ID或请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "角色不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /roles/{id} [put]
func UpdateRole(svcCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.BadRequest(c, "无效的角色ID")
			return
		}
		var request types.UpsertRoleRequest
		if err = c.ShouldBindJSON(&request); err != nil {
			response.BadRequest(c, err.Error())
			return
		}

		role, err := svcCtx.RoleService.FindByID(c.Request.Context(), uint(id))
		if err != nil {
			logrus.Errorf("failed to find role by id %d, %v", id, err)
			response.Fail(c, 500, err.Error())
			return
		}
		if request.Name != "" && request.Name != role.Name {
			exist, err := svcCtx.RoleService.Exists(c.Request.Context(), _interface.WithScopes(func(db *gorm.DB) *gorm.DB {
				return db.Where("name = ? AND id <> ?", request.Name, id)
			}))
			if err != nil {
				logrus.Errorf("failed to check role exist %s", err)
				response.Fail(c, 500, err.Error())
				return
			}
			if exist {
				response.Fail(c, http.StatusConflict, fmt.Sprintf("角色名 %s 已存在", request.Name))
				return
			}
		}
		err = svcCtx.RoleService.UpdateByID(c.Request.Context(), uint(id), map[string]interface{}{
			"name":        request.Name,
			"description": request.Description,
			"status":      request.Status,
		})
		if err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
		response.Success(c, nil)
	}
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
func DeleteRole(svcCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.BadRequest(c, "无效的角色ID")
			return
		}
		if err = svcCtx.RoleService.DeleteByID(c.Request.Context(), uint(id)); err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
		response.Success(c, nil)
	}
}

// AssignRoleResources godoc
// @Summary 绑定资源权限
// @Description 根据角色ID，为角色绑定资源权限
// @Tags RBAC-角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "角色ID"
// @Success 204 {object} response.Response "成功绑定资源"
// @Failure 400 {object} response.Response "无效的角色ID"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /roles/{id} [delete]
func AssignRoleResources(svcCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.BadRequest(c, "无效的角色ID")
			return
		}
		request := types.AssignResource{}
		if err = c.ShouldBindJSON(&request); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		role, err := svcCtx.RoleService.FindByID(c.Request.Context(), uint(id))
		if err != nil {
			logrus.Errorf("failed to find role by id %d, %v", id, err)
			response.Fail(c, 500, err.Error())
			return
		}
		resources, err := svcCtx.ResourceService.FindByIDs(c.Request.Context(), request.ResourceIds)
		if err != nil {
			logrus.Errorf("failed to find resources by ids %+v, %v", request.ResourceIds, err)
			response.Fail(c, 500, err.Error())
			return
		}
		userIds, err := svcCtx.RoleService.ListRoleUsers(uint(id))
		if err != nil {
			logrus.Errorf("failed to list role users by id %d, %v", id, err)
			response.Fail(c, 500, err.Error())
			return
		}
		err = svcCtx.CacheService.ClearMultipleUsersPermissions(c.Request.Context(), userIds, time.Millisecond*50, func() error {
			tx := svcCtx.RoleService.DB.WithContext(c.Request.Context())
			return tx.Model(&role).Association("Resources").Replace(resources)
		})
		if err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
		response.Success(c, nil)
	}
}
