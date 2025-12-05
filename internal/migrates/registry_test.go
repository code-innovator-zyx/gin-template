package migrates

import (
	"gin-admin/internal/model/rbac"
	"testing"
)

/*
 * @Author: zouyx
 * @Email: 1003941268@qq.com
 * @Date:   2025 2025/12/5 下午6:27
 * @Package: 迁移注册测试
 */

func TestModelRegistry(t *testing.T) {
	// 测试前重置注册表
	Reset()

	// 测试注册单个模型
	Register(&rbac.User{})
	models := GetAllModels()
	if len(models) != 1 {
		t.Errorf("expected 1 model, got %d", len(models))
	}

	// 测试重置
	Reset()
	models = GetAllModels()
	if len(models) != 0 {
		t.Errorf("expected 0 models after reset, got %d", len(models))
	}

	// 测试分组注册
	RegisterGroup("rbac",
		&rbac.User{},
		&rbac.Role{},
		&rbac.Permission{},
		&rbac.Resource{},
	)

	// 验证总数
	models = GetAllModels()
	if len(models) != 4 {
		t.Errorf("expected 4 models, got %d", len(models))
	}

	// 验证分组
	rbacModels := GetGroupModels("rbac")
	if len(rbacModels) != 4 {
		t.Errorf("expected 4 rbac models, got %d", len(rbacModels))
	}

	// 验证不存在的分组
	unknownModels := GetGroupModels("unknown")
	if unknownModels != nil {
		t.Errorf("expected nil for unknown group, got %v", unknownModels)
	}

	// 验证分组列表
	groups := GetAllGroups()
	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}
	if groups[0] != "rbac" {
		t.Errorf("expected group name 'rbac', got %s", groups[0])
	}
}

func TestMultipleGroups(t *testing.T) {
	Reset()

	// 注册多个分组
	RegisterGroup("rbac", &rbac.User{}, &rbac.Role{})
	RegisterGroup("system", &rbac.Permission{})

	// 验证总数
	models := GetAllModels()
	if len(models) != 3 {
		t.Errorf("expected 3 models, got %d", len(models))
	}

	// 验证各分组
	if len(GetGroupModels("rbac")) != 2 {
		t.Errorf("expected 2 rbac models")
	}
	if len(GetGroupModels("system")) != 1 {
		t.Errorf("expected 1 system model")
	}

	// 验证分组数量
	groups := GetAllGroups()
	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
}
