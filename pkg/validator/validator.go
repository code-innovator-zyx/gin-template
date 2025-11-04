package validator

import (
	"fmt"
	"gin-template/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"strings"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateStruct 验证结构体
func ValidateStruct(data interface{}) error {
	return validate.Struct(data)
}

// BindAndValidate 绑定并验证请求参数
func BindAndValidate(c *gin.Context, obj interface{}) error {
	// 绑定参数
	if err := c.ShouldBindJSON(obj); err != nil {
		return fmt.Errorf("参数绑定失败: %w", err)
	}

	// 验证参数
	if err := validate.Struct(obj); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return fmt.Errorf("参数验证失败: %s", formatValidationErrors(validationErrors))
		}
		return fmt.Errorf("参数验证失败: %w", err)
	}

	return nil
}

// BindAndValidateWithResponse 绑定验证并返回错误响应
func BindAndValidateWithResponse(c *gin.Context, obj interface{}) bool {
	if err := BindAndValidate(c, obj); err != nil {
		response.BadRequest(c, err.Error())
		return false
	}
	return true
}

// formatValidationErrors 格式化验证错误信息
func formatValidationErrors(errs validator.ValidationErrors) string {
	var messages []string
	for _, err := range errs {
		messages = append(messages, formatFieldError(err))
	}
	return strings.Join(messages, "; ")
}

// formatFieldError 格式化单个字段错误
func formatFieldError(err validator.FieldError) string {
	field := err.Field()
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s是必填项", field)
	case "email":
		return fmt.Sprintf("%s必须是有效的邮箱地址", field)
	case "min":
		return fmt.Sprintf("%s最小值为%s", field, err.Param())
	case "max":
		return fmt.Sprintf("%s最大值为%s", field, err.Param())
	case "len":
		return fmt.Sprintf("%s长度必须为%s", field, err.Param())
	case "oneof":
		return fmt.Sprintf("%s必须是以下值之一: %s", field, err.Param())
	default:
		return fmt.Sprintf("%s验证失败: %s", field, err.Tag())
	}
}

