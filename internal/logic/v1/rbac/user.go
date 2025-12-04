package rbac

import (
	"context"
	"fmt"
	"gin-admin/internal/model/rbac"
	"gin-admin/internal/services"
	types "gin-admin/internal/types/rbac"
	"gin-admin/pkg/consts"
	_interface "gin-admin/pkg/interface"
	"gin-admin/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
	"time"
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
func Register(svcCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req types.UserRegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		if err := svcCtx.UserService.CheckAccountExist(c.Request.Context(), req.Username, req.Email); nil != err {
			response.Forbidden(c, err.Error())
			return
		}
		user := &rbac.User{
			Username: req.Username,
			Password: req.Password,
			Email:    req.Email,
		}
		if err := svcCtx.UserService.Create(c, user); err != nil {
			response.Fail(c, 400, err.Error())
			return
		}
		response.Success(c, user)
	}
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
func Login(svcCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req types.UserLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		user, err := svcCtx.UserService.FindOne(c, _interface.WithScopes(func(db *gorm.DB) *gorm.DB {
			return db.Where("username = ? OR email = ?", req.Account, req.Account)
		}))
		if err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
		if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			response.Fail(c, http.StatusForbidden, "密码错误")
			return
		}
		// 生成JWT令牌对
		tokenPair, err := svcCtx.Jwt.GenerateTokenPair(c, user.ID, user.Username, user.Email)
		if err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
		// 返回令牌对
		tokenResponse := types.TokenResponse{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
			TokenType:    tokenPair.TokenType,
			ExpiresIn:    tokenPair.ExpiresIn,
		}
		// 设置刷新token
		c.SetCookie("X-Refresh-Token",
			tokenPair.RefreshToken,
			int(svcCtx.Config.Jwt.RefreshTokenExpire.Seconds()),
			"/",
			"",
			false,
			true)
		response.Success(c, tokenResponse)
	}
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
func Logout(svcCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("uid")
		sessionId := c.GetString("sessionId")
		if err := svcCtx.CacheService.ClearUserPermissions(c, userID, time.Millisecond*5, func() error {
			return svcCtx.Jwt.RevokeSession(c, sessionId)
		}); err != nil {
			response.Fail(c, 200, err.Error())
			return
		}
		response.Success(c, "登出成功")
	}
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
func GetProfile(svcCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("uid")
		ctx := c.Request.Context()

		// 获取用户基础信息
		user, err := svcCtx.UserService.FindByID(ctx, userID, _interface.WithPreloads("Roles"))
		if err != nil {
			response.Fail(c, 500, "获取用户信息失败: "+err.Error())
			return
		}

		// 获取用户可访问的资源列表
		resources, err := svcCtx.UserService.GetUserPerms(ctx, userID)
		if err != nil {
			response.Fail(c, 500, "获取用户资源失败: "+err.Error())
			return
		}
		user.Password = ""
		// 组装用户完整资料
		profile := types.UserProfile{
			User:        user,
			Permissions: resources,
		}

		response.Success(c, profile)
	}
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
func DeleteUser(svcCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		userID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "无效的用户ID")
			return
		}

		err = svcCtx.UserService.DeleteByID(c.Request.Context(), uint(userID))
		if err != nil {
			response.Fail(c, 500, "删除用户失败: "+err.Error())
			return
		}
		response.Success(c, nil)
	}
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
func CreateUser(svcCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		request := types.UpsertUserRequest{}
		if err := c.ShouldBindJSON(&request); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		if err := svcCtx.UserService.CheckAccountExist(c.Request.Context(), request.Username, request.Email); nil != err {
			response.Fail(c, 500, err.Error())
			return
		}
		roles, err := svcCtx.RoleService.List(c.Request.Context())
		if err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
		// ==== 创建 ====
		user := rbac.User{
			Username: request.Username,
			Password: strings.Split(request.Email, "@")[0],
			Email:    request.Email,
			Gender:   request.Gender,
			Roles:    roles,
			Status:   consts.UserStatusActive,
		}
		if err = svcCtx.UserService.Create(c.Request.Context(), &user); err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
		response.Success(c, "创建成功")
	}
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
func UpdateUser(svcCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		userID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "无效的用户ID")
			return
		}
		request := types.UpsertUserRequest{}
		if err := c.ShouldBindJSON(&request); err != nil {
			response.BadRequest(c, err.Error())
			return
		}

		request.Id = uint(userID)
		exist, err := svcCtx.UserService.Exists(c.Request.Context(), _interface.WithScopes(func(db *gorm.DB) *gorm.DB {
			return db.Where("username = ? AND id <> ?", request.Username, request.Id)
		}))
		if err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
		if exist {
			response.Fail(c, http.StatusConflict, "用户名已存在")
			return
		}
		exist, err = svcCtx.UserService.Exists(c.Request.Context(), _interface.WithScopes(func(db *gorm.DB) *gorm.DB {
			return db.Where("email = ? AND id <> ?", request.Email, request.Id)
		}))
		if err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
		if exist {
			response.Fail(c, http.StatusConflict, "邮箱已存在")
			return
		}
		err = svcCtx.UserService.Transaction(c.Request.Context(), func(ctx context.Context, tx *gorm.DB, txRepo _interface.IRepo[rbac.User]) error {
			err = txRepo.UpdateByID(ctx, request.Id, map[string]interface{}{
				"username": request.Username,
				"email":    request.Email,
				"gender":   request.Gender,
			})
			if err != nil {
				return err
			}
			var roles []rbac.Role
			// 获取所有角色列表
			if len(request.Roles) != 0 {
				roles, err = svcCtx.RoleService.FindByIDs(ctx, request.Roles)
				if err != nil {
					return err
				}
			}
			// 更新用户角色
			if err = tx.Association("Roles").Replace(roles); err != nil {
				return fmt.Errorf("更新用户角色失败: %w", err)
			}
			return nil
		})
		if err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
		response.Success(c, "更新成功")
	}
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
func ListUser(svcCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		request := types.ListUserRequest{}
		err := c.ShouldBindQuery(&request)
		if err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		pr, err := svcCtx.UserService.FindPage(c.Request.Context(), _interface.WithPagination(request.Page, request.PageSize),
			_interface.WithScopes(func(db *gorm.DB) *gorm.DB {
				if request.Username != "" {
					db = db.Where("username LIKE ?", request.Username+"%")
				}
				if request.Email != "" {
					db = db.Where("email = ?", request.Email)
				}
				if request.Status > 0 {
					db = db.Where("status = ?", request.Status)
				}
				if request.Gender > 0 {
					db = db.Where("gender = ?", request.Gender)
				}
				return db.Preload("Roles")
			}))
		if err != nil {
			response.Fail(c, 500, "获取用户列表失败: "+err.Error())
			return
		}
		response.SuccessPage(c, pr.List, pr.Page, pr.PageSize, pr.Total)
	}
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
func UserOptions(svcCtx *services.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
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
			if fn, ok := services.OptionGenerators[services.OptionField(field)]; ok {
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
}
