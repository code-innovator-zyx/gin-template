package rbac

import (
	"gin-template/internal/model/rbac"
	"gin-template/internal/service"
	types "gin-template/internal/types/rbac"
	"gin-template/pkg/response"
	"gin-template/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
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
// @Router /user/register [post]
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

	if err := service.GetUserService().Create(c.Request.Context(), user); err != nil {
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
// @Router /user/login [post]
func Login(c *gin.Context) {
	var req types.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tokenPair, err := service.GetUserService().Login(c.Request.Context(), req.Account, req.Password)
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

	response.Success(c, tokenResponse)
}

// RefreshToken godoc
// @Summary 刷新令牌
// @Description 使用 Refresh Token 获取新的 Access Token 和 Refresh Token
// @Tags RBAC-用户管理
// @Accept json
// @Produce json
// @Param data body RefreshTokenRequest true "刷新令牌信息"
// @Success 200 {object} response.Response{data=TokenResponse} "刷新成功返回新令牌对"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "刷新令牌无效或已过期"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /user/refresh [post]
func RefreshToken(c *gin.Context) {
	var req types.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 使用 Refresh Token 获取新的令牌对
	jwtManager := utils.GetJWTManager()
	tokenPair, err := jwtManager.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if err == utils.ErrRefreshTokenExpired || err == utils.ErrTokenExpired {
			response.Unauthorized(c, "刷新令牌已过期，请重新登录")
			return
		}
		if err == utils.ErrTokenBlacklisted {
			response.Unauthorized(c, "令牌已被撤销，请重新登录")
			return
		}
		response.Unauthorized(c, "刷新令牌无效")
		return
	}

	// 返回新的令牌对
	tokenResponse := types.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}

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
// @Router /user/logout [post]
func Logout(c *gin.Context) {
	// 获取 Access Token
	authHeader := c.GetHeader("Authorization")
	accessToken := ""
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			accessToken = parts[1]
		}
	}

	jwtManager := utils.GetJWTManager()
	ctx := c.Request.Context()

	// 撤销 Access Token
	if accessToken != "" {
		_ = jwtManager.RevokeToken(ctx, accessToken)
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
// @Router /user/profile [get]
func GetProfile(c *gin.Context) {
	userID := c.GetUint("userID")
	ctx := c.Request.Context()

	// 获取用户基础信息
	user, err := service.GetUserService().GetByID(ctx, userID)
	if err != nil {
		response.Fail(c, 500, "获取用户信息失败: "+err.Error())
		return
	}

	// 获取用户可访问的资源列表
	resources, err := service.GetRbacService().GetUserPermissionGroups(ctx, userID)
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

// ListUser godoc
// @Summary 获取用户列表
// @Description 获取所有用户信息列表（支持分页）
// @Tags RBAC-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request query types.ListUserRequest true "查询参数"
// @Success 200 {object} response.PaginatedResponse{data=map[string]interface{}} "成功返回用户列表"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /user [get]
func ListUser(c *gin.Context) {
	// 获取分页参数
	request := types.ListUserRequest{}
	err := c.ShouldBindQuery(&request)
	if err != nil {
		response.BadRequest(c, err.Error())
	}

	ctx := c.Request.Context()

	// 获取用户列表
	users, total, err := service.GetUserService().List(ctx, request)
	if err != nil {
		response.Fail(c, 500, "获取用户列表失败: "+err.Error())
		return
	}

	response.SuccessPage(c, users, request.Page, request.PageSize, total)
}
