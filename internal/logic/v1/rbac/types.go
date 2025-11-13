package rbac

import "gin-template/internal/model/rbac"

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/10/31 下午4:40
* @Package:
 */

// UserRegisterRequest 用户注册请求
type UserRegisterRequest struct {
	Username string `json:"username" binding:"required" example:"johndoe" description:"用户名"`
	Password string `json:"password" binding:"required" example:"password123" description:"密码"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com" description:"邮箱"`
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Account  string `json:"account" binding:"required" example:"johndoe" description:"用户名或者邮箱账号"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// TokenResponse 令牌响应
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int64  `json:"expires_in" example:"3600"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// UserProfile 用户详情
type UserProfile struct {
	User      *rbac.User      `json:"user" description:"用户基础信息"`
	Resources []rbac.Resource `json:"resources" description:"用户可访问的资源列表"`
}

type CreateRoleRequest struct {
	Name        string `json:"name,omitempty" example:"管理员"`
	Description string `json:"description,omitempty" example:"系统管理员"`
	Resources   []uint `json:"resources,omitempty" description:"角色绑定的资源ID列表"`
}

// UpdateRoleRequest 更新角色
type UpdateRoleRequest struct {
	Name        string `json:"name,omitempty" example:"管理员"`
	Description string `json:"description,omitempty" example:"系统管理员"`
	Resources   []uint `json:"resources,omitempty" description:"角色绑定的资源ID列表"`
}
