package services

import (
	"context"
	"gin-admin/internal/model/rbac"
	types "gin-admin/internal/types/rbac"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/15 下午7:43
* @Package:
 */

type OptionField string

const (
	FieldRole OptionField = "role"
	FieldUser OptionField = "user"
)

// TODO 增加缓存
// OptionGenerators 给每个常量定义生成 Option 的方法
var OptionGenerators = map[OptionField]func(ctx context.Context) ([]types.Option, error){
	FieldRole: func(ctx context.Context) ([]types.Option, error) {
		roles := []rbac.Role{}
		tx := SvcContext.Db
		if err := tx.Find(&roles).Error; err != nil {
			return nil, err
		}
		opts := make([]types.Option, len(roles))
		for i, r := range roles {
			opts[i] = types.Option{
				Label: r.Name,
				Value: r.ID,
			}
		}
		return opts, nil
	},
	FieldUser: func(ctx context.Context) ([]types.Option, error) {
		roles := []rbac.User{}
		tx := SvcContext.Db
		if err := tx.Find(&roles).Error; err != nil {
			return nil, err
		}
		opts := make([]types.Option, len(roles))
		for i, r := range roles {
			opts[i] = types.Option{
				Label: r.Username,
				Value: r.ID,
			}
		}
		return opts, nil
	},
}
