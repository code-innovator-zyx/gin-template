package errcode

// 业务错误码定义
// 规则：10000以下为系统级错误，10000以上为业务错误
const (
	// 系统级错误码 (0-9999)
	Success            = 0
	ServerError        = 1000 // 服务器内部错误
	InvalidParams      = 1001 // 请求参数错误
	Unauthorized       = 1002 // 未授权
	Forbidden          = 1003 // 禁止访问
	NotFound           = 1004 // 资源不存在
	TooManyRequests    = 1005 // 请求过于频繁
	ServiceUnavailable = 1006 // 服务不可用

	// 数据库相关错误 (2000-2999)
	DatabaseError      = 2000 // 数据库错误
	RecordNotFound     = 2001 // 记录不存在
	RecordAlreadyExist = 2002 // 记录已存在
	DatabaseConnFailed = 2003 // 数据库连接失败

	// 缓存相关错误 (3000-3999)
	CacheError     = 3000 // 缓存错误
	CacheKeyNotFound = 3001 // 缓存键不存在
	CacheSetFailed = 3002 // 缓存设置失败

	// 认证授权相关错误 (4000-4999)
	TokenExpired        = 4000 // Token过期
	TokenInvalid        = 4001 // Token无效
	TokenMissing        = 4002 // Token缺失
	PermissionDenied    = 4003 // 权限不足
	LoginFailed         = 4004 // 登录失败
	UserNotFound        = 4005 // 用户不存在
	UserAlreadyExist    = 4006 // 用户已存在
	PasswordWrong       = 4007 // 密码错误
	UserDisabled        = 4008 // 用户已被禁用

	// 业务相关错误 (10000+)
	RoleNotFound        = 10000 // 角色不存在
	RoleAlreadyExist    = 10001 // 角色已存在
	PermissionNotFound  = 10002 // 权限不存在
	ResourceNotFound    = 10003 // 资源不存在
)

// Error 错误结构
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

// New 创建新错误
func New(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// 错误信息映射
var messages = map[int]string{
	Success:            "成功",
	ServerError:        "服务器内部错误",
	InvalidParams:      "请求参数错误",
	Unauthorized:       "未授权",
	Forbidden:          "禁止访问",
	NotFound:           "资源不存在",
	TooManyRequests:    "请求过于频繁",
	ServiceUnavailable: "服务不可用",
	
	DatabaseError:      "数据库错误",
	RecordNotFound:     "记录不存在",
	RecordAlreadyExist: "记录已存在",
	DatabaseConnFailed: "数据库连接失败",
	
	CacheError:       "缓存错误",
	CacheKeyNotFound: "缓存键不存在",
	CacheSetFailed:   "缓存设置失败",
	
	TokenExpired:     "Token已过期",
	TokenInvalid:     "Token无效",
	TokenMissing:     "缺少Token",
	PermissionDenied: "权限不足",
	LoginFailed:      "登录失败",
	UserNotFound:     "用户不存在",
	UserAlreadyExist: "用户已存在",
	PasswordWrong:    "密码错误",
	UserDisabled:     "用户已被禁用",
	
	RoleNotFound:       "角色不存在",
	RoleAlreadyExist:   "角色已存在",
	PermissionNotFound: "权限不存在",
	ResourceNotFound:   "资源不存在",
}

// GetMessage 获取错误消息
func GetMessage(code int) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return "未知错误"
}

// NewError 创建标准错误
func NewError(code int) *Error {
	return &Error{
		Code:    code,
		Message: GetMessage(code),
	}
}

// NewCustomError 创建自定义错误消息
func NewCustomError(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

