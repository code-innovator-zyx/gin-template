package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"` // 业务状态码
	Message string      `json:"msg"`  // 提示信息
	Data    interface{} `json:"data"` // 数据
}
type PaginatedResponse struct {
	Code     int    `json:"code"`    // 业务状态码
	Message  string `json:"message"` // 提示信息
	Data     any    `json:"data"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Total    int64  `json:"total"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// SuccessPage 分页响应
func SuccessPage(c *gin.Context, data any, page, pageSize int, total int64) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Code:     200,
		Message:  "success",
		Data:     data,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// FailWithStatus 带状态码的失败响应
func FailWithStatus(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// FailWithData 带数据的失败响应
func FailWithData(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// BadRequest 400错误
func BadRequest(c *gin.Context, message string) {
	FailWithStatus(c, http.StatusBadRequest, 400, message)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, message string) {
	FailWithStatus(c, http.StatusUnauthorized, 401, message)
}

// Forbidden 403错误
func Forbidden(c *gin.Context, message string) {
	FailWithStatus(c, http.StatusForbidden, 403, message)
}

// NotFound 404错误
func NotFound(c *gin.Context, message string) {
	FailWithStatus(c, http.StatusNotFound, 404, message)
}

// InternalServerError 500错误
func InternalServerError(c *gin.Context, message string) {
	FailWithStatus(c, http.StatusInternalServerError, 500, message)
}

// Created 201创建成功响应
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    http.StatusCreated,
		Message: "success",
		Data:    data,
	})
}

// NoContent 204无内容响应
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
