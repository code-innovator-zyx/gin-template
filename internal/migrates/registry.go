package migrates

import (
	"sync"
)

/*
 * @Author: zouyx
 * @Email: 1003941268@qq.com
 * @Date:   2025 2025/12/5 下午6:27
 * @Package: 自动注册模式的迁移注册表
 */

// ModelRegistry 模型注册表
type ModelRegistry struct {
	mu     sync.RWMutex
	models []interface{}
	groups map[string][]interface{} // 按模块分组
}

var (
	registry = &ModelRegistry{
		groups: make(map[string][]interface{}),
	}
)

// Register 注册单个模型到全局注册表
// 用法：在 model 包的 init() 函数中调用
func Register(models ...interface{}) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.models = append(registry.models, models...)
}

// RegisterGroup 注册一组模型（按模块分组）
// group: 模块名称，如 "rbac", "mall", "system"
func RegisterGroup(group string, models ...interface{}) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	
	registry.models = append(registry.models, models...)
	registry.groups[group] = append(registry.groups[group], models...)
}

// GetAllModels 获取所有已注册的模型
func GetAllModels() []interface{} {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	
	// 返回副本，避免外部修改
	models := make([]interface{}, len(registry.models))
	copy(models, registry.models)
	return models
}

// GetGroupModels 获取指定模块的模型
func GetGroupModels(group string) []interface{} {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	
	models, ok := registry.groups[group]
	if !ok {
		return nil
	}
	
	// 返回副本
	result := make([]interface{}, len(models))
	copy(result, models)
	return result
}

// GetAllGroups 获取所有模块名称
func GetAllGroups() []string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	
	groups := make([]string, 0, len(registry.groups))
	for group := range registry.groups {
		groups = append(groups, group)
	}
	return groups
}

// Reset 重置注册表（主要用于测试）
func Reset() {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	
	registry.models = nil
	registry.groups = make(map[string][]interface{})
}
