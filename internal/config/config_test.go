package config

import (
	"github.com/gin-gonic/gin"
	"testing"
)

func TestApp_GetGinMode(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected string
	}{
		{
			name:     "开发环境映射到 DebugMode",
			env:      "dev",
			expected: gin.DebugMode,
		},
		{
			name:     "测试环境映射到 TestMode",
			env:      "test",
			expected: gin.TestMode,
		},
		{
			name:     "生产环境映射到 ReleaseMode",
			env:      "prod",
			expected: gin.ReleaseMode,
		},
		{
			name:     "未知环境默认为 DebugMode",
			env:      "unknown",
			expected: gin.DebugMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := App{
				Env: tt.env,
			}
			got := app.GetGinMode()
			if got != tt.expected {
				t.Errorf("GetGinMode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

