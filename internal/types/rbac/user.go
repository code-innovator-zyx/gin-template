package rbac

import (
	"gin-template/internal/model/rbac"
	"gin-template/pkg/consts"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/14 下午3:43
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

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// UserProfile 用户详情
type UserProfile struct {
	User        *rbac.User        `json:"user" description:"用户基础信息"`
	Permissions []rbac.Permission `json:"permissions" description:"用户可访问的资源列表"`
}

// TokenResponse 令牌响应
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int64  `json:"expires_in" example:"3600"`
}

type ListUserRequest struct {
	Username string `form:"username,optional" json:"username" binding:"-" example:"johndoe" description:"用户名"`
	Email    string `form:"email,optional" json:"email" binding:"-" example:"john@example.com"`
	Status   uint8  `form:"status,optional" json:"status" binding:"-" example:"1"`
	Gender   uint8  `form:"gender,optional" json:"gender" binding:"-" example:"1"`
	Page     int    `form:"page,default=1" json:"page" binding:"required" example:"1" default:"1"`
	PageSize int    `form:"pageSize,default=10" json:"pageSize" binding:"required" example:"10" default:"10"`
}
type Option struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

type UpsertUserRequest struct {
	Id       uint
	Username string        `json:"username" binding:"required" example:"johndoe"`
	Email    string        `json:"email" binding:"required" example:"john@example.com"`
	Gender   consts.Gender `json:"gender" binding:"required" example:"1"`
	Roles    []uint        `json:"roles" binding:"required"`
}

type UserOptionParams struct {
	IncludeFields []string `form:"include_fields,optional" json:"include_fields" binding:"-" description:"需要从数据库获取的补充字段"`
}
type UserOptions struct {
	Gender            []Option            `json:"gender"`
	Status            []Option            `json:"status"`
	SupplementOptions map[string][]Option `json:"supplement_options"`
}
