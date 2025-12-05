package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			expected: "debug",
		},
		{
			name:     "测试环境映射到 TestMode",
			env:      "test",
			expected: "test",
		},
		{
			name:     "生产环境映射到 ReleaseMode",
			env:      "prod",
			expected: "release",
		},
		{
			name:     "未知环境默认为 DebugMode",
			env:      "unknown",
			expected: "debug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := App{
				Env: tt.env,
			}
			got := app.GetGinMode()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestApp_Fields(t *testing.T) {
	app := App{
		Name:          "test-app",
		Version:       "1.0.0",
		Env:           "dev",
		EnableSwagger: true,
	}

	assert.Equal(t, "test-app", app.Name)
	assert.Equal(t, "1.0.0", app.Version)
	assert.Equal(t, "dev", app.Env)
	assert.True(t, app.EnableSwagger)
	assert.Equal(t, "debug", app.GetGinMode())
}

func TestAdminUserConfig(t *testing.T) {
	adminUser := AdminUserConfig{
		Username: "admin",
		Password: "securepassword123",
		Email:    "admin@example.com",
	}

	assert.Equal(t, "admin", adminUser.Username)
	assert.Equal(t, "admin@example.com", adminUser.Email)
	assert.NotEmpty(t, adminUser.Password)
}

func TestAdminRoleConfig(t *testing.T) {
	adminRole := AdminRoleConfig{
		Name:        "System Administrator",
		Description: "Full system access",
	}

	assert.Equal(t, "System Administrator", adminRole.Name)
	assert.Equal(t, "Full system access", adminRole.Description)
}
