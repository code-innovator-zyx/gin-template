package validator

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type TestRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,min=0,max=150"`
}

func TestBindAndValidate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		request     TestRequest
		expectError bool
	}{
		{
			name: "valid request",
			request: TestRequest{
				Name:  "test",
				Email: "test@example.com",
				Age:   25,
			},
			expectError: false,
		},
		{
			name: "missing required field",
			request: TestRequest{
				Email: "test@example.com",
				Age:   25,
			},
			expectError: true,
		},
		{
			name: "invalid email",
			request: TestRequest{
				Name:  "test",
				Email: "invalid-email",
				Age:   25,
			},
			expectError: true,
		},
		{
			name: "age out of range",
			request: TestRequest{
				Name:  "test",
				Email: "test@example.com",
				Age:   200,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试请求
			jsonData, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// 测试验证
			var data TestRequest
			err := BindAndValidate(c, &data)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.request.Name, data.Name)
				assert.Equal(t, tt.request.Email, data.Email)
				assert.Equal(t, tt.request.Age, data.Age)
			}
		})
	}
}

