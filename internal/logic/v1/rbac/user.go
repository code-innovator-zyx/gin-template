package rbac

import (
	"gin-template/internal/model/rbac"
	"gin-template/internal/service"
	"gin-template/pkg/response"
	"github.com/gin-gonic/gin"
)

// Register godoc
// @Summary 用户注册
// @Description 创建新用户账号
// @Tags RBAC-用户管理
// @Accept json
// @Produce json
// @Param data body UserRegisterRequest true "用户注册信息"
// @Success 200 {object} response.Response{data=rbac.User} "注册成功返回用户信息"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /rbac/user/register [post]
func Register(c *gin.Context) {
	var req UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user := &rbac.User{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	}

	if err := service.UserService.Create(user); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	response.Success(c, user)
}

// Login godoc
// @Summary 用户登录
// @Description 用户登录并获取JWT令牌
// @Tags RBAC-用户管理
// @Accept json
// @Produce json
// @Param data body UserLoginRequest true "用户登录信息"
// @Success 200 {object} response.Response{data=map[string]string} "登录成功返回token"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "用户名或密码错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /rbac/user/login [post]
func Login(c *gin.Context) {
	var req UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	token, err := service.UserService.Login(req.Username, req.Password)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, gin.H{"token": token})
}

// GetProfile godoc
// @Summary 获取用户个人资料
// @Description 获取当前登录用户的个人资料
// @Tags RBAC-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=rbac.User} "成功返回用户信息"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /rbac/user/profile [get]
func GetProfile(c *gin.Context) {
	userID := c.GetUint("userID")
	user, err := service.UserService.GetByID(userID)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	response.Success(c, user)
}