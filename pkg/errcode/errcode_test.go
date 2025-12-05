package errcode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	code := 1000
	message := "测试错误"

	err := New(code, message)

	assert.NotNil(t, err)
	assert.Equal(t, code, err.Code)
	assert.Equal(t, message, err.Message)
}

func TestError_Error(t *testing.T) {
	err := &Error{
		Code:    1000,
		Message: "服务器错误",
	}

	assert.Equal(t, "服务器错误", err.Error())
}

func TestGetMessage(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected string
	}{
		{"Success", Success, "成功"},
		{"ServerError", ServerError, "服务器内部错误"},
		{"InvalidParams", InvalidParams, "请求参数错误"},
		{"Unauthorized", Unauthorized, "未授权"},
		{"Forbidden", Forbidden, "禁止访问"},
		{"NotFound", NotFound, "资源不存在"},
		{"TooManyRequests", TooManyRequests, "请求过于频繁"},
		{"ServiceUnavailable", ServiceUnavailable, "服务不可用"},
		{"DatabaseError", DatabaseError, "数据库错误"},
		{"RecordNotFound", RecordNotFound, "记录不存在"},
		{"RecordAlreadyExist", RecordAlreadyExist, "记录已存在"},
		{"DatabaseConnFailed", DatabaseConnFailed, "数据库连接失败"},
		{"CacheError", CacheError, "缓存错误"},
		{"CacheKeyNotFound", CacheKeyNotFound, "缓存键不存在"},
		{"CacheSetFailed", CacheSetFailed, "缓存设置失败"},
		{"TokenExpired", TokenExpired, "Token已过期"},
		{"TokenInvalid", TokenInvalid, "Token无效"},
		{"TokenMissing", TokenMissing, "缺少Token"},
		{"PermissionDenied", PermissionDenied, "权限不足"},
		{"LoginFailed", LoginFailed, "登录失败"},
		{"UserNotFound", UserNotFound, "用户不存在"},
		{"UserAlreadyExist", UserAlreadyExist, "用户已存在"},
		{"PasswordWrong", PasswordWrong, "密码错误"},
		{"UserDisabled", UserDisabled, "用户已被禁用"},
		{"RoleNotFound", RoleNotFound, "角色不存在"},
		{"RoleAlreadyExist", RoleAlreadyExist, "角色已存在"},
		{"PermissionNotFound", PermissionNotFound, "权限不存在"},
		{"ResourceNotFound", ResourceNotFound, "资源不存在"},
		{"UnknownCode", 99999, "未知错误"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := GetMessage(tt.code)
			assert.Equal(t, tt.expected, msg)
		})
	}
}

func TestNewError(t *testing.T) {
	tests := []struct {
		name         string
		code         int
		expectedCode int
		expectedMsg  string
	}{
		{"Success", Success, Success, "成功"},
		{"ServerError", ServerError, ServerError, "服务器内部错误"},
		{"UserNotFound", UserNotFound, UserNotFound, "用户不存在"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.code)
			assert.NotNil(t, err)
			assert.Equal(t, tt.expectedCode, err.Code)
			assert.Equal(t, tt.expectedMsg, err.Message)
		})
	}
}

func TestNewCustomError(t *testing.T) {
	code := 1000
	customMsg := "自定义错误消息"

	err := NewCustomError(code, customMsg)

	assert.NotNil(t, err)
	assert.Equal(t, code, err.Code)
	assert.Equal(t, customMsg, err.Message)
	assert.NotEqual(t, GetMessage(code), err.Message)
}

func TestErrorCodes(t *testing.T) {
	// 测试错误码范围
	assert.Equal(t, 0, Success)

	// 系统级错误码 (1000-1999)
	assert.Equal(t, 1000, ServerError)
	assert.Equal(t, 1001, InvalidParams)

	// 数据库相关错误 (2000-2999)
	assert.Equal(t, 2000, DatabaseError)
	assert.Equal(t, 2001, RecordNotFound)

	// 缓存相关错误 (3000-3999)
	assert.Equal(t, 3000, CacheError)
	assert.Equal(t, 3001, CacheKeyNotFound)

	// 认证授权相关错误 (4000-4999)
	assert.Equal(t, 4000, TokenExpired)
	assert.Equal(t, 4001, TokenInvalid)

	// 业务相关错误 (10000+)
	assert.Equal(t, 10000, RoleNotFound)
	assert.Equal(t, 10001, RoleAlreadyExist)
}

func TestErrorAsError(t *testing.T) {
	// Test that Error implements error interface
	var err error = &Error{
		Code:    ServerError,
		Message: "test error",
	}

	assert.Equal(t, "test error", err.Error())
}
