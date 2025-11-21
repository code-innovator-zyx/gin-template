package rbac

import (
	"gin-admin/internal/core"
	"gin-admin/internal/model/rbac"
	"gin-admin/internal/service"
	rbac2 "gin-admin/internal/service/rbac"
	types "gin-admin/internal/types/rbac"
	"gin-admin/pkg/consts"
	"gin-admin/pkg/response"
	"github.com/gin-gonic/gin"
	"strconv"
)

// Register godoc
// @Summary 用户注册
// @Description 创建新用户账号
// @Tags RBAC-用户管理
// @Accept json
// @Produce json
// @Param data body types.UserRegisterRequest true "用户注册信息"
// @Success 200 {object} response.Response{data=rbac.User} "注册成功返回用户信息"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /users/register [post]
func Register(c *gin.Context) {
	var req types.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user := &rbac.User{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	}

	if err := rbac2.NewUserService(c.Request.Context()).Register(user); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	response.Success(c, user)
}

// Login godoc
// @Summary 用户登录
// @Description 用户登录并获取JWT令牌（Access Token + Refresh Token）
// @Tags RBAC-用户管理
// @Accept json
// @Produce json
// @Param data body types.UserLoginRequest true "用户登录信息"
// @Success 200 {object} response.Response{data=types.TokenResponse} "登录成功返回令牌对"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "用户名或密码错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /users/login [post]
func Login(c *gin.Context) {
	var req types.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tokenPair, err := rbac2.NewUserService(c.Request.Context()).Login(req.Account, req.Password)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	// 返回令牌对
	tokenResponse := types.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}
	c.SetCookie("X-Refresh-Token",
		tokenPair.RefreshToken,
		int(core.MustGetConfig().Jwt.RefreshTokenExpire.Seconds()),
		"/",
		"",
		false,
		true)
	response.Success(c, tokenResponse)
}

// Logout godoc
// @Summary 用户登出
// @Description 撤销当前用户的令牌
// @Tags RBAC-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response "登出成功"
// @Failure 401 {object} response.Response "未授权"
// @Router /users/logout [post]
func Logout(c *gin.Context) {
	// 获取 Access Token
	userID := c.GetUint("uid")
	sessionId := c.GetString("sessionId")
	if err := rbac2.NewUserService(c).LoginOut(userID, sessionId); err != nil {
		response.Fail(c, 200, err.Error())
		return
	}
	response.Success(c, "登出成功")
}

// GetProfile godoc
// @Summary 获取用户个人资料
// @Description 获取当前登录用户的完整资料（包括角色、权限和可访问资源）
// @Tags RBAC-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=UserProfile} "成功返回用户完整资料"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /users/profile [get]
func GetProfile(c *gin.Context) {
	userID := c.GetUint("uid")
	ctx := c.Request.Context()

	// 获取用户基础信息
	user, err := rbac2.NewUserService(ctx).FindByID(userID, "Roles")
	if err != nil {
		response.Fail(c, 500, "获取用户信息失败: "+err.Error())
		return
	}

	// 获取用户可访问的资源列表
	resources, err := rbac2.NewPermissionService(ctx).GetUserPerms(userID)
	if err != nil {
		response.Fail(c, 500, "获取用户资源失败: "+err.Error())
		return
	}

	// 组装用户完整资料
	profile := types.UserProfile{
		User:        user,
		Permissions: resources,
	}

	response.Success(c, profile)
}

// DeleteUser godoc
// @Summary 移除用户
// @Description 从系统移除用户
// @Tags RBAC-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response "成功返回"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /users/{id}[delete]
func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	// 获取用户基础信息
	err = rbac2.NewUserService(c.Request.Context()).DeleteByID(uint(userID))
	if err != nil {
		response.Fail(c, 500, "删除用户失败: "+err.Error())
		return
	}
	response.Success(c, nil)
}

// CreateUser godoc
// @Summary 创建用户
// @Description 系统内部管理员创建用户，密码默认就是邮箱号
// @Tags RBAC-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body types.UpsertUserRequest true "创建参数"
// @Success 200 {object} response.PaginatedResponse "成功返回用户列表"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /users [post]
func CreateUser(c *gin.Context) {
	request := types.UpsertUserRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, err.Error())
	}
	// 创建用户
	err := rbac2.NewUserService(c.Request.Context()).Create(request)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, "创建成功")
}

// UpdateUser godoc
// @Summary 更新用户
// @Description 管理员更新用户信息
// @Tags RBAC-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "用户ID"
// @Param data body types.UpsertUserRequest true "更新参数"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "请求格式错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /users/{id} [put]
func UpdateUser(c *gin.Context) {
	// 读取 path ID
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}
	request := types.UpsertUserRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, err.Error())
	}

	request.Id = uint(userID)

	// 更新用户
	err = rbac2.NewUserService(c.Request.Context()).Update(request)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, "更新成功")
}

// ListUser godoc
// @Summary 获取用户列表
// @Description 获取所有用户信息列表（支持分页）
// @Tags RBAC-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request query types.ListUserRequest true "查询参数"
// @Success 200 {object} response.Response "成功返回"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /users [get]
func ListUser(c *gin.Context) {
	// 获取分页参数
	request := types.ListUserRequest{}
	err := c.ShouldBindQuery(&request)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()

	// 获取用户列表
	pageResult, err := rbac2.NewUserService(ctx).List(request)
	if err != nil {
		response.Fail(c, 500, "获取用户列表失败: "+err.Error())
		return
	}

	response.SuccessPage(c, pageResult.List, pageResult.Page, pageResult.PageSize, pageResult.Total)
}

// UserOptions godoc
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
func UserOptions(c *gin.Context) {
	params := types.OptionParams{}
	err := c.ShouldBindQuery(&params)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	options := types.UserOptions{
		SupplementOptions: make(map[string][]types.Option, len(params.IncludeFields)),
	}
	for _, gender := range consts.AllGender() {
		options.Gender = append(options.Gender, types.Option{
			Label: gender.String(),
			Value: gender,
		})
	}
	for _, status := range consts.AllUserStatus() {
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
