package v1

import (
	"encoding/json"
	"gin-admin/pkg/response"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试路由
	r := gin.New()
	r.GET("/health", HealthCheck)

	// 创建测试请求
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// 执行请求
	r.ServeHTTP(w, req)

	// 验证状态码
	assert.Equal(t, http.StatusOK, w.Code)

	// 验证响应内容
	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)

	// 验证健康检查数据
	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, data, "status")
	assert.Contains(t, data, "timestamp")
}

