package rbac

import (
	"gin-template/internal/service"
	"gin-template/internal/service/rbac"
	types "gin-template/internal/types/rbac"
	"gin-template/pkg/consts"
	"gin-template/pkg/response"
	"github.com/gin-gonic/gin"
	"strconv"
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
func CreateRole(c *gin.Context) {
	var request types.UpsertRoleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := rbac.NewRoleService(c.Request.Context()).Create(request); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Created(c, nil)
}

/*
* @Author: zouyx
* @Email: zouyx@knowsec.com
* @Date:   2025/11/16 下午4:40
* @Package:
 */

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
		response.Fail(c, 500, "获取角色列表失败")
		return
	}
	response.SuccessPage(c, result.List, result.Page, result.PageSize, result.Total)
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
		response.Fail(c, 500, err.Error())
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
// @Param role body types.UpsertRoleRequest true "角色信息"
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
	var request types.UpsertRoleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	request.Id = uint(id)
	err = rbac.NewRoleService(c.Request.Context()).Update(request)
	if err != nil {
		response.Fail(c, 500, err.Error())
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
	if err = rbac.NewRoleService(c.Request.Context()).DeleteByID(uint(id)); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, nil)
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
func AssignRoleResources(c *gin.Context) {
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
	request.Id = uint(id)
	err = rbac.NewRoleService(c.Request.Context()).AssignResource(request)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, nil)
}

// RoleOptions godoc
// @Summary 用户options
// @Description 用户创建修改的option枚举信息
// @Tags RBAC-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request query types.UserOptionParams true "查询参数"
// @Success 200 {object} response.Response{data=types.UserOptions} "成功返回用户列表"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /users/options [get]
func RoleOptions(c *gin.Context) {
	params := types.OptionParams{}
	err := c.ShouldBindQuery(&params)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	options := types.RoleOptions{
		SupplementOptions: make(map[string][]types.Option, len(params.IncludeFields)),
	}
	for _, status := range consts.AllRoleStatus() {
		options.Status = append(options.Status, types.Option{
			Label: status.String(),
			Value: status,
		})
	}
	ctx := c.Request.Context()
	for _, field := range params.IncludeFields {
		if fn, ok := service.OptionGenerators[service.OptionField(field)]; ok {
			fieldOptions, err := fn(ctx)
			if err != nil {
				response.Fail(c, 500, err.Error())
				return
			}
			options.SupplementOptions[field] = fieldOptions
		}
	}
	response.Success(c, options)
}
