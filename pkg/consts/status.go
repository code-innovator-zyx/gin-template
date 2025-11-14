package consts

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/13 上午10:40
* @Package:
 */

// 角色状态
type RoleStatus uint8

const (
	ROLESTATUS_UNKNOWN RoleStatus = iota
	ROLESTATUS_ACTIVE
	ROLESTATUS_INACTIVE
)

func (status RoleStatus) String() string {
	switch status {
	case ROLESTATUS_ACTIVE:
		return "启用"
	case ROLESTATUS_INACTIVE:
		return "禁用"
	}
	return "invalid"
}

// 用户状态
type UserStatus uint8

const (
	UserStatusUnknown  UserStatus = iota
	UserStatusActive              // 启用：允许登录
	UserStatusDisabled            // 禁用：禁止登录
	UserStatusLocked              // 锁定：异常登录或安全封禁
)

func (s UserStatus) String() string {
	switch s {
	case UserStatusActive:
		return "正常"
	case UserStatusDisabled:
		return "禁用"
	case UserStatusLocked:
		return "锁定"
	default:
		return "未知"
	}
}
